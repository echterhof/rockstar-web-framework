package pkg

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// httpServer implements the Server interface
type httpServer struct {
	// Server state
	addr         string
	listener     net.Listener
	httpServer   *http.Server
	http3Server  *http3.Server
	quicListener *quic.EarlyListener
	router       RouterEngine
	config       ServerConfig
	middleware   []MiddlewareFunc
	errorHandler func(ctx Context, err error) error

	// Protocol flags
	http1Enabled bool
	http2Enabled bool
	quicEnabled  bool

	// Lifecycle management
	running       atomic.Bool
	shutdownHooks []func(ctx context.Context) error
	mu            sync.RWMutex

	// Graceful shutdown
	activeConns  sync.WaitGroup
	shutdownCtx  context.Context
	shutdownFunc context.CancelFunc

	// Managers for context creation
	logger    Logger
	metrics   MetricsCollector
	session   SessionManager
	database  DatabaseManager
	cache     CacheManager
	configMgr ConfigManager
	i18n      I18nManager
	security  SecurityManager
}

// NewServer creates a new HTTP server instance
func NewServer(config ServerConfig) Server {
	ctx, cancel := context.WithCancel(context.Background())

	s := &httpServer{
		config:        config,
		http1Enabled:  config.EnableHTTP1,
		http2Enabled:  config.EnableHTTP2,
		quicEnabled:   config.EnableQUIC,
		shutdownHooks: make([]func(ctx context.Context) error, 0),
		shutdownCtx:   ctx,
		shutdownFunc:  cancel,
	}

	// Set default error handler if not provided
	if s.errorHandler == nil {
		s.errorHandler = defaultErrorHandler
	}

	return s
}

// Listen starts the HTTP server on the specified address
func (s *httpServer) Listen(addr string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running.Load() {
		return errors.New("server is already running")
	}

	s.addr = addr

	// Create listener with platform-specific options
	var listener net.Listener
	var err error

	if s.config.ListenerConfig != nil {
		// Use custom listener configuration
		listenerConfig := *s.config.ListenerConfig
		listenerConfig.Address = addr
		listener, err = CreateListener(listenerConfig)
	} else if s.config.EnablePrefork {
		// Use prefork with default configuration
		listenerConfig := ListenerConfig{
			Network:        "tcp",
			Address:        addr,
			EnablePrefork:  true,
			PreforkWorkers: s.config.PreforkWorkers,
			ReusePort:      true,
			ReuseAddr:      true,
			ReadBuffer:     s.config.ReadBufferSize,
			WriteBuffer:    s.config.WriteBufferSize,
		}
		listener, err = CreateListener(listenerConfig)
	} else {
		// Standard listener
		listenerConfig := ListenerConfig{
			Network:     "tcp",
			Address:     addr,
			ReuseAddr:   true,
			ReadBuffer:  s.config.ReadBufferSize,
			WriteBuffer: s.config.WriteBufferSize,
		}
		listener, err = CreateListener(listenerConfig)
	}

	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}
	s.listener = listener

	// Create HTTP server
	s.httpServer = s.createHTTPServer()

	// Mark as running
	s.running.Store(true)

	// Start serving (blocking)
	var serveErr error
	if s.http2Enabled && !s.config.EnableHTTP1 {
		// HTTP/2 only with h2c (HTTP/2 cleartext)
		h2s := &http2.Server{}
		handler := h2c.NewHandler(s.httpServer.Handler, h2s)
		serveErr = http.Serve(listener, handler)
	} else {
		serveErr = s.httpServer.Serve(listener)
	}

	s.running.Store(false)

	if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
		return fmt.Errorf("server error: %w", serveErr)
	}

	return nil
}

// ListenTLS starts the HTTPS server with TLS
func (s *httpServer) ListenTLS(addr, certFile, keyFile string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running.Load() {
		return errors.New("server is already running")
	}

	s.addr = addr

	// Load TLS certificates
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return fmt.Errorf("failed to load TLS certificates: %w", err)
	}

	// Configure TLS
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	// Enable HTTP/2 if configured
	if s.http2Enabled {
		tlsConfig.NextProtos = append(tlsConfig.NextProtos, "h2", "http/1.1")
	}

	if s.config.TLSConfig != nil {
		// Merge with provided TLS config
		if len(s.config.TLSConfig.Certificates) > 0 {
			tlsConfig.Certificates = s.config.TLSConfig.Certificates
		}
		if s.config.TLSConfig.MinVersion > 0 {
			tlsConfig.MinVersion = s.config.TLSConfig.MinVersion
		}
		if len(s.config.TLSConfig.NextProtos) > 0 {
			tlsConfig.NextProtos = s.config.TLSConfig.NextProtos
		}
	}

	// Create base listener with platform-specific options
	var baseListener net.Listener

	if s.config.ListenerConfig != nil {
		// Use custom listener configuration
		listenerConfig := *s.config.ListenerConfig
		listenerConfig.Address = addr
		baseListener, err = CreateListener(listenerConfig)
	} else if s.config.EnablePrefork {
		// Use prefork with default configuration
		listenerConfig := ListenerConfig{
			Network:        "tcp",
			Address:        addr,
			EnablePrefork:  true,
			PreforkWorkers: s.config.PreforkWorkers,
			ReusePort:      true,
			ReuseAddr:      true,
			ReadBuffer:     s.config.ReadBufferSize,
			WriteBuffer:    s.config.WriteBufferSize,
		}
		baseListener, err = CreateListener(listenerConfig)
	} else {
		// Standard listener
		listenerConfig := ListenerConfig{
			Network:     "tcp",
			Address:     addr,
			ReuseAddr:   true,
			ReadBuffer:  s.config.ReadBufferSize,
			WriteBuffer: s.config.WriteBufferSize,
		}
		baseListener, err = CreateListener(listenerConfig)
	}

	if err != nil {
		return fmt.Errorf("failed to create base listener: %w", err)
	}

	// Wrap with TLS
	listener := tls.NewListener(baseListener, tlsConfig)
	s.listener = listener

	// Create HTTP server
	s.httpServer = s.createHTTPServer()
	s.httpServer.TLSConfig = tlsConfig

	// Configure HTTP/2
	if s.http2Enabled {
		if err := http2.ConfigureServer(s.httpServer, &http2.Server{}); err != nil {
			return fmt.Errorf("failed to configure HTTP/2: %w", err)
		}
	}

	// Mark as running
	s.running.Store(true)

	// Start serving (blocking)
	serveErr := s.httpServer.Serve(listener)

	s.running.Store(false)

	if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
		return fmt.Errorf("server error: %w", serveErr)
	}

	return nil
}

// ListenQUIC starts the QUIC server with HTTP/3
func (s *httpServer) ListenQUIC(addr, certFile, keyFile string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running.Load() {
		return errors.New("server is already running")
	}

	if !s.quicEnabled {
		return errors.New("QUIC protocol is not enabled")
	}

	s.addr = addr

	// Load TLS certificates
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return fmt.Errorf("failed to load TLS certificates: %w", err)
	}

	// Configure TLS for QUIC
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
		NextProtos:   []string{"h3"}, // HTTP/3 protocol
	}

	// Merge with provided TLS config if available
	if s.config.TLSConfig != nil {
		if len(s.config.TLSConfig.Certificates) > 0 {
			tlsConfig.Certificates = s.config.TLSConfig.Certificates
		}
		if s.config.TLSConfig.MinVersion > 0 {
			tlsConfig.MinVersion = s.config.TLSConfig.MinVersion
		}
		// Always ensure h3 is in NextProtos for QUIC
		if len(s.config.TLSConfig.NextProtos) > 0 {
			tlsConfig.NextProtos = append([]string{"h3"}, s.config.TLSConfig.NextProtos...)
		}
	}

	// Create QUIC config
	quicConfig := &quic.Config{
		MaxIdleTimeout:  s.config.IdleTimeout,
		KeepAlivePeriod: 30 * time.Second,
	}

	// Create UDP address
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to resolve UDP address: %w", err)
	}

	// Create UDP connection
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return fmt.Errorf("failed to create UDP listener: %w", err)
	}

	// Create QUIC listener
	quicListener, err := quic.ListenEarly(udpConn, tlsConfig, quicConfig)
	if err != nil {
		udpConn.Close()
		return fmt.Errorf("failed to create QUIC listener: %w", err)
	}
	s.quicListener = quicListener

	// Create HTTP/3 server
	s.http3Server = &http3.Server{
		Handler:    s.createHandler(),
		TLSConfig:  tlsConfig,
		QUICConfig: quicConfig,
	}

	// Mark as running
	s.running.Store(true)

	// Start serving (blocking)
	serveErr := s.http3Server.ServeListener(quicListener)

	s.running.Store(false)

	if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) && !errors.Is(serveErr, quic.ErrServerClosed) {
		return fmt.Errorf("QUIC server error: %w", serveErr)
	}

	return nil
}

// Shutdown gracefully shuts down the server
func (s *httpServer) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running.Load() {
		return nil
	}

	// Cancel shutdown context
	if s.shutdownFunc != nil {
		s.shutdownFunc()
	}

	// Execute shutdown hooks
	for _, hook := range s.shutdownHooks {
		if err := hook(ctx); err != nil {
			// Log error but continue shutdown
			fmt.Printf("shutdown hook error: %v\n", err)
		}
	}

	// Shutdown HTTP server
	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown HTTP server: %w", err)
		}
	}

	// Shutdown HTTP/3 server
	if s.http3Server != nil {
		if err := s.http3Server.Close(); err != nil {
			return fmt.Errorf("failed to shutdown HTTP/3 server: %w", err)
		}
	}

	// Close QUIC listener
	if s.quicListener != nil {
		if err := s.quicListener.Close(); err != nil {
			return fmt.Errorf("failed to close QUIC listener: %w", err)
		}
	}

	// Wait for active connections to complete
	done := make(chan struct{})
	go func() {
		s.activeConns.Wait()
		close(done)
	}()

	select {
	case <-done:
		s.running.Store(false)
		return nil
	case <-ctx.Done():
		s.running.Store(false)
		return ctx.Err()
	}
}

// Close immediately closes the server
func (s *httpServer) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running.Load() {
		return nil
	}

	s.shutdownFunc()

	// Close HTTP server
	if s.httpServer != nil {
		if err := s.httpServer.Close(); err != nil {
			return fmt.Errorf("failed to close HTTP server: %w", err)
		}
	}

	// Close HTTP/3 server
	if s.http3Server != nil {
		if err := s.http3Server.Close(); err != nil {
			return fmt.Errorf("failed to close HTTP/3 server: %w", err)
		}
	}

	// Close QUIC listener
	if s.quicListener != nil {
		if err := s.quicListener.Close(); err != nil {
			return fmt.Errorf("failed to close QUIC listener: %w", err)
		}
	}

	s.running.Store(false)
	return nil
}

// EnableHTTP1 enables HTTP/1.1 protocol
func (s *httpServer) EnableHTTP1() Server {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.http1Enabled = true
	return s
}

// EnableHTTP2 enables HTTP/2 protocol
func (s *httpServer) EnableHTTP2() Server {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.http2Enabled = true
	return s
}

// EnableQUIC enables QUIC protocol
func (s *httpServer) EnableQUIC() Server {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.quicEnabled = true
	return s
}

// SetConfig sets the server configuration
func (s *httpServer) SetConfig(config ServerConfig) Server {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config = config
	return s
}

// SetMiddleware sets global middleware
func (s *httpServer) SetMiddleware(middleware ...MiddlewareFunc) Server {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.middleware = middleware
	return s
}

// SetRouter sets the router engine
func (s *httpServer) SetRouter(router RouterEngine) Server {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.router = router
	return s
}

// SetErrorHandler sets the error handler
func (s *httpServer) SetErrorHandler(handler func(ctx Context, err error) error) Server {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.errorHandler = handler
	return s
}

// SetManagers sets the managers for context creation
func (s *httpServer) SetManagers(logger Logger, metrics MetricsCollector, session SessionManager, database DatabaseManager, cache CacheManager, configMgr ConfigManager, i18n I18nManager, security SecurityManager) Server {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logger = logger
	s.metrics = metrics
	s.session = session
	s.database = database
	s.cache = cache
	s.configMgr = configMgr
	s.i18n = i18n
	s.security = security
	return s
}

// Addr returns the server address
func (s *httpServer) Addr() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.addr
}

// IsRunning returns whether the server is running
func (s *httpServer) IsRunning() bool {
	return s.running.Load()
}

// Protocol returns the enabled protocols
func (s *httpServer) Protocol() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	protocols := []string{}
	if s.http1Enabled {
		protocols = append(protocols, "HTTP/1.1")
	}
	if s.http2Enabled {
		protocols = append(protocols, "HTTP/2")
	}
	if s.quicEnabled {
		protocols = append(protocols, "QUIC")
	}

	if len(protocols) == 0 {
		return "none"
	}

	result := protocols[0]
	for i := 1; i < len(protocols); i++ {
		result += ", " + protocols[i]
	}
	return result
}

// RegisterShutdownHook registers a function to be called during shutdown
func (s *httpServer) RegisterShutdownHook(hook func(ctx context.Context) error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.shutdownHooks = append(s.shutdownHooks, hook)
}

// GracefulShutdown performs a graceful shutdown with timeout
func (s *httpServer) GracefulShutdown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return s.Shutdown(ctx)
}

// createHTTPServer creates the underlying http.Server
func (s *httpServer) createHTTPServer() *http.Server {
	return &http.Server{
		Handler:        s.createHandler(),
		ReadTimeout:    s.config.ReadTimeout,
		WriteTimeout:   s.config.WriteTimeout,
		IdleTimeout:    s.config.IdleTimeout,
		MaxHeaderBytes: s.config.MaxHeaderBytes,
		ConnContext:    s.connContext,
	}
}

// createHandler creates the HTTP handler
func (s *httpServer) createHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Track active connection
		s.activeConns.Add(1)
		defer s.activeConns.Done()

		// Check if server is shutting down
		select {
		case <-s.shutdownCtx.Done():
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		default:
		}

		// Parse request
		req, err := s.parseRequest(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Create response writer
		respWriter := newResponseWriter(w)

		// Create context
		ctx := s.createContext(req, respWriter, r)

		// Execute middleware and handler
		if err := s.executeHandler(ctx); err != nil {
			if s.errorHandler != nil {
				if handlerErr := s.errorHandler(ctx, err); handlerErr != nil {
					http.Error(w, handlerErr.Error(), http.StatusInternalServerError)
				}
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
	})
}

// parseRequest parses the HTTP request into framework Request
func (s *httpServer) parseRequest(r *http.Request) (*Request, error) {
	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}
	defer r.Body.Close()

	// Detect protocol
	protocol := "HTTP/1.1"
	if r.ProtoMajor == 2 {
		protocol = "HTTP/2"
	}

	req := &Request{
		Method:     r.Method,
		URL:        r.URL,
		Proto:      r.Proto,
		Header:     r.Header,
		Host:       r.Host,
		RemoteAddr: r.RemoteAddr,
		RequestURI: r.RequestURI,
		StartTime:  time.Now(),
		RawBody:    body,
		Protocol:   protocol,
		Query:      make(map[string]string),
		Params:     make(map[string]string),
		Form:       make(map[string]string),
		Files:      make(map[string]*FormFile),
	}

	// Parse query parameters
	for key, values := range r.URL.Query() {
		if len(values) > 0 {
			req.Query[key] = values[0]
		}
	}

	// Generate request ID
	req.ID = generateRequestID()

	return req, nil
}

// createContext creates a framework Context from the request
func (s *httpServer) createContext(req *Request, respWriter ResponseWriter, httpReq *http.Request) Context {
	return &contextImpl{
		request:  req,
		response: respWriter,
		httpReq:  httpReq,
		params:   req.Params,
		query:    req.Query,
		headers:  make(map[string]string),
		ctx:      httpReq.Context(),
		logger:   s.logger,
		metrics:  s.metrics,
		session:  s.session,
		db:       s.database,
		cache:    s.cache,
		config:   s.configMgr,
		i18n:     s.i18n,
	}
}

// executeHandler executes middleware chain and handler
func (s *httpServer) executeHandler(ctx Context) error {
	// If no router is set, return error
	if s.router == nil {
		return errors.New("no router configured")
	}

	// Match route
	req := ctx.Request()
	route, params, found := s.router.Match(req.Method, req.URL.Path, req.Host)
	if !found {
		ctx.Response().WriteHeader(http.StatusNotFound)
		return nil
	}

	// Update context params
	for k, v := range params {
		req.Params[k] = v
	}

	// Build middleware chain
	handler := route.Handler

	// Apply route-specific middleware (in reverse order)
	for i := len(route.Middleware) - 1; i >= 0; i-- {
		mw := route.Middleware[i]
		next := handler
		handler = func(ctx Context) error {
			return mw(ctx, next)
		}
	}

	// Apply global middleware (in reverse order)
	for i := len(s.middleware) - 1; i >= 0; i-- {
		mw := s.middleware[i]
		next := handler
		handler = func(ctx Context) error {
			return mw(ctx, next)
		}
	}

	// Execute handler chain
	return handler(ctx)
}

// connContext tracks connection context for graceful shutdown
func (s *httpServer) connContext(ctx context.Context, c net.Conn) context.Context {
	return ctx
}

// defaultErrorHandler is the default error handler
func defaultErrorHandler(ctx Context, err error) error {
	ctx.Response().WriteHeader(http.StatusInternalServerError)
	_, writeErr := ctx.Response().Write([]byte(err.Error()))
	if writeErr != nil {
		return writeErr
	}
	return nil
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	// Simple implementation - in production use UUID or similar
	return fmt.Sprintf("req-%d", time.Now().UnixNano())
}
