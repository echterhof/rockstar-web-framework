package pkg

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocket message types
const (
	// TextMessage denotes a text data message
	TextMessage = websocket.TextMessage
	// BinaryMessage denotes a binary data message
	BinaryMessage = websocket.BinaryMessage
	// CloseMessage denotes a close control message
	CloseMessage = websocket.CloseMessage
	// PingMessage denotes a ping control message
	PingMessage = websocket.PingMessage
	// PongMessage denotes a pong control message
	PongMessage = websocket.PongMessage
)

// WebSocketUpgrader handles WebSocket upgrade requests
var DefaultUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins by default - should be configured per application
		return true
	},
}

// wsConnection implements the WebSocketConnection interface
type wsConnection struct {
	conn      *websocket.Conn
	mu        sync.Mutex
	closed    bool
	ctx       context.Context
	cancel    context.CancelFunc
	writeChan chan wsMessage
	closeChan chan struct{}
}

// wsMessage represents a message to be sent
type wsMessage struct {
	messageType int
	data        []byte
	err         chan error
}

// newWebSocketConnection creates a new WebSocket connection wrapper
func newWebSocketConnection(conn *websocket.Conn) *wsConnection {
	ctx, cancel := context.WithCancel(context.Background())

	wsc := &wsConnection{
		conn:      conn,
		ctx:       ctx,
		cancel:    cancel,
		writeChan: make(chan wsMessage, 256),
		closeChan: make(chan struct{}),
	}

	// Start write pump
	go wsc.writePump()

	return wsc
}

// ReadMessage reads a message from the WebSocket connection
// Requirements: 12.5
func (wsc *wsConnection) ReadMessage() (messageType int, data []byte, err error) {
	if wsc.closed {
		return 0, nil, errors.New("connection is closed")
	}

	messageType, data, err = wsc.conn.ReadMessage()
	if err != nil {
		wsc.Close()
		return 0, nil, err
	}

	return messageType, data, nil
}

// WriteMessage writes a message to the WebSocket connection
// Requirements: 12.5
func (wsc *wsConnection) WriteMessage(messageType int, data []byte) error {
	if wsc.closed {
		return errors.New("connection is closed")
	}

	msg := wsMessage{
		messageType: messageType,
		data:        data,
		err:         make(chan error, 1),
	}

	select {
	case wsc.writeChan <- msg:
		return <-msg.err
	case <-wsc.closeChan:
		return errors.New("connection is closed")
	case <-wsc.ctx.Done():
		return wsc.ctx.Err()
	}
}

// writePump handles writing messages to the WebSocket connection
func (wsc *wsConnection) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		wsc.conn.Close()
	}()

	for {
		select {
		case msg := <-wsc.writeChan:
			wsc.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			err := wsc.conn.WriteMessage(msg.messageType, msg.data)
			msg.err <- err
			if err != nil {
				return
			}
		case <-ticker.C:
			wsc.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := wsc.conn.WriteMessage(PingMessage, nil); err != nil {
				return
			}
		case <-wsc.closeChan:
			return
		case <-wsc.ctx.Done():
			return
		}
	}
}

// Close closes the WebSocket connection
// Requirements: 12.5
func (wsc *wsConnection) Close() error {
	wsc.mu.Lock()
	defer wsc.mu.Unlock()

	if wsc.closed {
		return nil
	}

	wsc.closed = true
	wsc.cancel()
	close(wsc.closeChan)

	// Send close message
	wsc.conn.WriteControl(CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
		time.Now().Add(time.Second))

	return wsc.conn.Close()
}

// RemoteAddr returns the remote network address
// Requirements: 12.5
func (wsc *wsConnection) RemoteAddr() string {
	return wsc.conn.RemoteAddr().String()
}

// LocalAddr returns the local network address
// Requirements: 12.5
func (wsc *wsConnection) LocalAddr() string {
	return wsc.conn.LocalAddr().String()
}

// WebSocketServer manages WebSocket connections and routing
type WebSocketServer struct {
	upgrader    websocket.Upgrader
	authManager *AuthManager
	routes      map[string]*wsRoute
	hostRoutes  map[string]map[string]*wsRoute
	mu          sync.RWMutex
}

// wsRoute represents a WebSocket route configuration
type wsRoute struct {
	path          string
	handler       WebSocketHandler
	middleware    []MiddlewareFunc
	requireAuth   bool
	allowedScopes []string
	host          string
}

// NewWebSocketServer creates a new WebSocket server
func NewWebSocketServer(authManager *AuthManager) *WebSocketServer {
	return &WebSocketServer{
		upgrader:    DefaultUpgrader,
		authManager: authManager,
		routes:      make(map[string]*wsRoute),
		hostRoutes:  make(map[string]map[string]*wsRoute),
	}
}

// SetUpgrader sets a custom WebSocket upgrader
func (wss *WebSocketServer) SetUpgrader(upgrader websocket.Upgrader) {
	wss.mu.Lock()
	defer wss.mu.Unlock()
	wss.upgrader = upgrader
}

// RegisterRoute registers a WebSocket route
// Requirements: 12.5
func (wss *WebSocketServer) RegisterRoute(path string, handler WebSocketHandler, middleware []MiddlewareFunc) {
	wss.mu.Lock()
	defer wss.mu.Unlock()

	wss.routes[path] = &wsRoute{
		path:       path,
		handler:    handler,
		middleware: middleware,
	}
}

// RegisterHostRoute registers a WebSocket route for a specific host
// Requirements: 12.5
func (wss *WebSocketServer) RegisterHostRoute(host, path string, handler WebSocketHandler, middleware []MiddlewareFunc) {
	wss.mu.Lock()
	defer wss.mu.Unlock()

	if wss.hostRoutes[host] == nil {
		wss.hostRoutes[host] = make(map[string]*wsRoute)
	}

	wss.hostRoutes[host][path] = &wsRoute{
		path:       path,
		handler:    handler,
		middleware: middleware,
		host:       host,
	}
}

// RegisterAuthenticatedRoute registers a WebSocket route that requires authentication
// Requirements: 3.6
func (wss *WebSocketServer) RegisterAuthenticatedRoute(path string, handler WebSocketHandler,
	middleware []MiddlewareFunc, allowedScopes []string) {
	wss.mu.Lock()
	defer wss.mu.Unlock()

	wss.routes[path] = &wsRoute{
		path:          path,
		handler:       handler,
		middleware:    middleware,
		requireAuth:   true,
		allowedScopes: allowedScopes,
	}
}

// RegisterAuthenticatedHostRoute registers an authenticated WebSocket route for a specific host
// Requirements: 3.6, 12.5
func (wss *WebSocketServer) RegisterAuthenticatedHostRoute(host, path string, handler WebSocketHandler,
	middleware []MiddlewareFunc, allowedScopes []string) {
	wss.mu.Lock()
	defer wss.mu.Unlock()

	if wss.hostRoutes[host] == nil {
		wss.hostRoutes[host] = make(map[string]*wsRoute)
	}

	wss.hostRoutes[host][path] = &wsRoute{
		path:          path,
		handler:       handler,
		middleware:    middleware,
		requireAuth:   true,
		allowedScopes: allowedScopes,
		host:          host,
	}
}

// HandleUpgrade handles WebSocket upgrade requests
// Requirements: 12.5, 3.6
func (wss *WebSocketServer) HandleUpgrade(ctx Context) error {
	req := ctx.Request()

	// Find matching route
	route := wss.findRoute(req.Host, req.URL.Path)
	if route == nil {
		return &FrameworkError{
			Code:       ErrCodeNotFound,
			Message:    "WebSocket route not found",
			StatusCode: http.StatusNotFound,
		}
	}

	// Authenticate if required
	if route.requireAuth {
		if err := wss.authenticateWebSocket(ctx, route); err != nil {
			return err
		}
	}

	// Get underlying HTTP response writer and request
	respWriter := ctx.Response().(*responseWriter)
	w := respWriter.ResponseWriter
	httpReq := ctx.(*contextImpl).httpReq

	// Upgrade connection
	conn, err := wss.upgrader.Upgrade(w, httpReq, nil)
	if err != nil {
		return &FrameworkError{
			Code:       ErrCodeWebSocketUpgradeFailed,
			Message:    "Failed to upgrade WebSocket connection",
			StatusCode: http.StatusBadRequest,
			Cause:      err,
		}
	}

	// Create WebSocket connection wrapper
	wsConn := newWebSocketConnection(conn)
	defer wsConn.Close()

	// Execute middleware chain
	handler := route.handler
	for i := len(route.middleware) - 1; i >= 0; i-- {
		mw := route.middleware[i]
		nextHandler := handler
		handler = func(ctx Context, conn WebSocketConnection) error {
			// Convert middleware to WebSocket context
			return mw(ctx, func(ctx Context) error {
				return nextHandler(ctx, conn)
			})
		}
	}

	// Execute handler
	return handler(ctx, wsConn)
}

// findRoute finds a matching WebSocket route
func (wss *WebSocketServer) findRoute(host, path string) *wsRoute {
	wss.mu.RLock()
	defer wss.mu.RUnlock()

	// Check host-specific routes first
	if hostRoutes, exists := wss.hostRoutes[host]; exists {
		if route, found := hostRoutes[path]; found {
			return route
		}
	}

	// Check global routes
	if route, found := wss.routes[path]; found {
		return route
	}

	return nil
}

// authenticateWebSocket authenticates a WebSocket connection using access token
// Requirements: 3.6
func (wss *WebSocketServer) authenticateWebSocket(ctx Context, route *wsRoute) error {
	if wss.authManager == nil {
		return &FrameworkError{
			Code:       ErrCodeAuthenticationFailed,
			Message:    "Authentication manager not configured",
			StatusCode: http.StatusInternalServerError,
		}
	}

	// Extract token from query parameter or header
	token := ctx.Query()["token"]
	if token == "" {
		token = ctx.GetHeader("Authorization")
		if token != "" && len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}
	}

	if token == "" {
		return &FrameworkError{
			Code:       ErrCodeAuthenticationFailed,
			Message:    "Access token required for WebSocket connection",
			StatusCode: http.StatusUnauthorized,
			I18nKey:    "error.websocket.token_required",
		}
	}

	// Validate access token
	accessToken, err := wss.authManager.AuthenticateAccessToken(token)
	if err != nil {
		return &FrameworkError{
			Code:       ErrCodeAuthenticationFailed,
			Message:    "Invalid access token",
			StatusCode: http.StatusUnauthorized,
			Cause:      err,
			I18nKey:    "error.websocket.invalid_token",
		}
	}

	// Check scopes if specified
	if len(route.allowedScopes) > 0 {
		hasScope := false
		for _, scope := range accessToken.Scopes {
			for _, allowedScope := range route.allowedScopes {
				if scope == allowedScope {
					hasScope = true
					break
				}
			}
			if hasScope {
				break
			}
		}

		if !hasScope {
			return &FrameworkError{
				Code:       ErrCodeAuthorizationFailed,
				Message:    fmt.Sprintf("Insufficient scopes for WebSocket connection. Required: %v", route.allowedScopes),
				StatusCode: http.StatusForbidden,
				I18nKey:    "error.websocket.insufficient_scopes",
				Details: map[string]interface{}{
					"required_scopes": route.allowedScopes,
					"user_scopes":     accessToken.Scopes,
				},
			}
		}
	}

	// Store user information in context
	ctx.Request().UserID = accessToken.UserID
	ctx.Request().TenantID = accessToken.TenantID
	ctx.Request().AccessToken = token

	return nil
}

// GetConnections returns the number of active WebSocket connections
func (wss *WebSocketServer) GetConnections() int {
	// This would be implemented with connection tracking
	// For now, return 0 as placeholder
	return 0
}
