package pkg

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// TestWebSocketConnection tests basic WebSocket connection operations
// Requirements: 12.5
func TestWebSocketConnection(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("Failed to upgrade connection: %v", err)
		}
		defer conn.Close()

		// Echo messages back
		for {
			messageType, data, err := conn.ReadMessage()
			if err != nil {
				break
			}
			if err := conn.WriteMessage(messageType, data); err != nil {
				break
			}
		}
	}))
	defer server.Close()

	// Connect to the test server
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Wrap in our connection type
	wsConn := newWebSocketConnection(conn)
	defer wsConn.Close()

	// Test WriteMessage
	testMessage := []byte("Hello, WebSocket!")
	if err := wsConn.WriteMessage(TextMessage, testMessage); err != nil {
		t.Fatalf("Failed to write message: %v", err)
	}

	// Test ReadMessage
	messageType, data, err := wsConn.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read message: %v", err)
	}

	if messageType != TextMessage {
		t.Errorf("Expected message type %d, got %d", TextMessage, messageType)
	}

	if string(data) != string(testMessage) {
		t.Errorf("Expected message %s, got %s", testMessage, data)
	}

	// Test RemoteAddr and LocalAddr
	if wsConn.RemoteAddr() == "" {
		t.Error("RemoteAddr should not be empty")
	}
	if wsConn.LocalAddr() == "" {
		t.Error("LocalAddr should not be empty")
	}
}

// TestWebSocketConnectionClose tests closing a WebSocket connection
// Requirements: 12.5
func TestWebSocketConnectionClose(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Wait for close
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	wsConn := newWebSocketConnection(conn)

	// Close the connection - ignore error as server may close first
	wsConn.Close()

	// Verify connection is closed
	if !wsConn.closed {
		t.Error("Connection should be marked as closed")
	}

	// Verify operations fail after close
	err = wsConn.WriteMessage(TextMessage, []byte("test"))
	if err == nil {
		t.Error("WriteMessage should fail on closed connection")
	}
}

// TestWebSocketServer tests WebSocket server functionality
// Requirements: 12.5
func TestWebSocketServer(t *testing.T) {
	// Create auth manager with mock database
	db := NewMockDatabaseManager()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	// Create WebSocket server
	wsServer := NewWebSocketServer(authManager)

	// Register a test route
	wsServer.RegisterRoute("/ws/test", func(ctx Context, conn WebSocketConnection) error {
		return nil
	}, nil)

	// Verify route was registered
	route := wsServer.findRoute("", "/ws/test")
	if route == nil {
		t.Fatal("Route should be registered")
	}

	if route.path != "/ws/test" {
		t.Errorf("Expected path /ws/test, got %s", route.path)
	}
}

// TestWebSocketServerHostRouting tests host-specific WebSocket routing
// Requirements: 12.5
func TestWebSocketServerHostRouting(t *testing.T) {
	db := NewMockDatabaseManager()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})
	wsServer := NewWebSocketServer(authManager)

	// Register host-specific route
	wsServer.RegisterHostRoute("example.com", "/ws/chat", func(ctx Context, conn WebSocketConnection) error {
		return nil
	}, nil)

	// Verify host route was registered
	route := wsServer.findRoute("example.com", "/ws/chat")
	if route == nil {
		t.Fatal("Host route should be registered")
	}

	if route.host != "example.com" {
		t.Errorf("Expected host example.com, got %s", route.host)
	}

	// Verify route is not found for different host
	route = wsServer.findRoute("other.com", "/ws/chat")
	if route != nil {
		t.Error("Route should not be found for different host")
	}
}

// TestWebSocketAuthentication tests WebSocket authentication with access tokens
// Requirements: 3.6
func TestWebSocketAuthentication(t *testing.T) {
	db := NewMockDatabaseManager()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	// Create a valid access token
	token, err := authManager.CreateAccessToken("user123", "tenant456", []string{"chat"}, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to create access token: %v", err)
	}

	wsServer := NewWebSocketServer(authManager)

	// Register authenticated route
	wsServer.RegisterAuthenticatedRoute("/ws/secure", func(ctx Context, conn WebSocketConnection) error {
		return nil
	}, nil, []string{"chat"})

	// Create mock context with token in query
	req := &Request{
		Method: "GET",
		Host:   "localhost",
		Query:  map[string]string{"token": token.Token},
		Params: make(map[string]string),
	}

	ctx := &contextImpl{
		request: req,
		query:   req.Query,
		headers: make(map[string]string),
	}

	// Find route
	route := wsServer.findRoute("", "/ws/secure")
	if route == nil {
		t.Fatal("Route should be registered")
	}

	// Test authentication
	err = wsServer.authenticateWebSocket(ctx, route)
	if err != nil {
		t.Errorf("Authentication should succeed: %v", err)
	}

	// Verify user info was set
	if req.UserID != "user123" {
		t.Errorf("Expected UserID user123, got %s", req.UserID)
	}
	if req.TenantID != "tenant456" {
		t.Errorf("Expected TenantID tenant456, got %s", req.TenantID)
	}
}

// TestWebSocketAuthenticationFailure tests WebSocket authentication failures
// Requirements: 3.6
func TestWebSocketAuthenticationFailure(t *testing.T) {
	db := NewMockDatabaseManager()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})
	wsServer := NewWebSocketServer(authManager)

	wsServer.RegisterAuthenticatedRoute("/ws/secure", func(ctx Context, conn WebSocketConnection) error {
		return nil
	}, nil, []string{"admin"})

	route := wsServer.findRoute("", "/ws/secure")

	tests := []struct {
		name        string
		token       string
		expectError bool
		errorCode   string
	}{
		{
			name:        "Missing token",
			token:       "",
			expectError: true,
			errorCode:   ErrCodeAuthenticationFailed,
		},
		{
			name:        "Invalid token",
			token:       "invalid-token",
			expectError: true,
			errorCode:   ErrCodeAuthenticationFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &Request{
				Method: "GET",
				Host:   "localhost",
				Query:  map[string]string{"token": tt.token},
				Params: make(map[string]string),
			}

			ctx := &contextImpl{
				request: req,
				query:   req.Query,
				headers: make(map[string]string),
			}

			err := wsServer.authenticateWebSocket(ctx, route)
			if tt.expectError {
				if err == nil {
					t.Error("Expected authentication to fail")
				}
				if frameworkErr, ok := err.(*FrameworkError); ok {
					if frameworkErr.Code != tt.errorCode {
						t.Errorf("Expected error code %s, got %s", tt.errorCode, frameworkErr.Code)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Expected authentication to succeed: %v", err)
				}
			}
		})
	}
}

// TestWebSocketAuthenticationWithInsufficientScopes tests scope validation
// Requirements: 3.6
func TestWebSocketAuthenticationWithInsufficientScopes(t *testing.T) {
	db := NewMockDatabaseManager()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	// Create token with "chat" scope
	token, err := authManager.CreateAccessToken("user123", "tenant456", []string{"chat"}, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to create access token: %v", err)
	}

	wsServer := NewWebSocketServer(authManager)

	// Register route requiring "admin" scope
	wsServer.RegisterAuthenticatedRoute("/ws/admin", func(ctx Context, conn WebSocketConnection) error {
		return nil
	}, nil, []string{"admin"})

	route := wsServer.findRoute("", "/ws/admin")

	req := &Request{
		Method: "GET",
		Host:   "localhost",
		Query:  map[string]string{"token": token.Token},
		Params: make(map[string]string),
	}

	ctx := &contextImpl{
		request: req,
		query:   req.Query,
		headers: make(map[string]string),
	}

	// Test authentication - should fail due to insufficient scopes
	err = wsServer.authenticateWebSocket(ctx, route)
	if err == nil {
		t.Error("Expected authentication to fail due to insufficient scopes")
	}

	if frameworkErr, ok := err.(*FrameworkError); ok {
		if frameworkErr.Code != ErrCodeAuthorizationFailed {
			t.Errorf("Expected error code %s, got %s", ErrCodeAuthorizationFailed, frameworkErr.Code)
		}
	}
}

// TestWebSocketAuthenticationWithHeader tests token in Authorization header
// Requirements: 3.6
func TestWebSocketAuthenticationWithHeader(t *testing.T) {
	db := NewMockDatabaseManager()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	token, err := authManager.CreateAccessToken("user123", "tenant456", []string{"chat"}, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to create access token: %v", err)
	}

	wsServer := NewWebSocketServer(authManager)
	wsServer.RegisterAuthenticatedRoute("/ws/secure", func(ctx Context, conn WebSocketConnection) error {
		return nil
	}, nil, []string{"chat"})

	route := wsServer.findRoute("", "/ws/secure")

	// Create proper HTTP header
	header := http.Header{}
	header.Set("Authorization", "Bearer "+token.Token)

	req := &Request{
		Method: "GET",
		Host:   "localhost",
		Query:  make(map[string]string),
		Params: make(map[string]string),
		Header: header,
	}

	ctx := &contextImpl{
		request: req,
		query:   req.Query,
		headers: make(map[string]string),
	}

	// Test authentication with Bearer token
	err = wsServer.authenticateWebSocket(ctx, route)
	if err != nil {
		t.Errorf("Authentication should succeed with Bearer token: %v", err)
	}

	if req.UserID != "user123" {
		t.Errorf("Expected UserID user123, got %s", req.UserID)
	}
}

// TestWebSocketRouterIntegration tests WebSocket integration with router
// Requirements: 12.5
func TestWebSocketRouterIntegration(t *testing.T) {
	router := NewRouter()

	router.WebSocket("/ws/test", func(ctx Context, conn WebSocketConnection) error {
		return nil
	})

	// Verify route was registered
	routes := router.Routes()
	found := false
	for _, route := range routes {
		if route.Path == "/ws/test" && route.IsWebSocket {
			found = true
			if route.WebSocketHandler == nil {
				t.Error("WebSocket handler should be set")
			}
			break
		}
	}

	if !found {
		t.Error("WebSocket route should be registered")
	}
}

// TestWebSocketHostRouterIntegration tests host-specific WebSocket routing
// Requirements: 12.5
func TestWebSocketHostRouterIntegration(t *testing.T) {
	router := NewRouter()

	// Register host-specific WebSocket route
	hostRouter := router.Host("example.com")
	hostRouter.WebSocket("/ws/chat", func(ctx Context, conn WebSocketConnection) error {
		return nil
	})

	// Verify route was registered for host
	routes := router.Routes()
	found := false
	for _, route := range routes {
		if route.Host == "example.com" && route.Path == "/ws/chat" && route.IsWebSocket {
			found = true
			break
		}
	}

	if !found {
		t.Error("Host-specific WebSocket route should be registered")
	}
}

// TestWebSocketConnectionWriteAfterClose tests writing after connection close
// Requirements: 12.5
func TestWebSocketConnectionWriteAfterClose(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Wait briefly then close
		time.Sleep(100 * time.Millisecond)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	wsConn := newWebSocketConnection(conn)
	wsConn.Close()

	// Try to write after close
	err = wsConn.WriteMessage(TextMessage, []byte("test"))
	if err == nil {
		t.Error("WriteMessage should fail after close")
	}

	// Try to read after close
	_, _, err = wsConn.ReadMessage()
	if err == nil {
		t.Error("ReadMessage should fail after close")
	}
}

// TestWebSocketServerSetUpgrader tests custom upgrader configuration
// Requirements: 12.5
func TestWebSocketServerSetUpgrader(t *testing.T) {
	db := NewMockDatabaseManager()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})
	wsServer := NewWebSocketServer(authManager)

	customUpgrader := websocket.Upgrader{
		ReadBufferSize:  2048,
		WriteBufferSize: 2048,
		CheckOrigin: func(r *http.Request) bool {
			return r.Host == "trusted.com"
		},
	}

	wsServer.SetUpgrader(customUpgrader)

	if wsServer.upgrader.ReadBufferSize != 2048 {
		t.Errorf("Expected ReadBufferSize 2048, got %d", wsServer.upgrader.ReadBufferSize)
	}
	if wsServer.upgrader.WriteBufferSize != 2048 {
		t.Errorf("Expected WriteBufferSize 2048, got %d", wsServer.upgrader.WriteBufferSize)
	}
}

// TestWebSocketMultipleMessages tests sending and receiving multiple messages
// Requirements: 12.5
func TestWebSocketMultipleMessages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Echo messages
		for i := 0; i < 5; i++ {
			messageType, data, err := conn.ReadMessage()
			if err != nil {
				break
			}
			if err := conn.WriteMessage(messageType, data); err != nil {
				break
			}
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	wsConn := newWebSocketConnection(conn)
	defer wsConn.Close()

	// Send and receive multiple messages
	for i := 0; i < 5; i++ {
		message := []byte("Message " + string(rune('0'+i)))

		if err := wsConn.WriteMessage(TextMessage, message); err != nil {
			t.Fatalf("Failed to write message %d: %v", i, err)
		}

		_, data, err := wsConn.ReadMessage()
		if err != nil {
			t.Fatalf("Failed to read message %d: %v", i, err)
		}

		if string(data) != string(message) {
			t.Errorf("Message %d mismatch: expected %s, got %s", i, message, data)
		}
	}
}

// TestWebSocketContextCancellation tests connection behavior with context cancellation
// Requirements: 12.5
func TestWebSocketContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Keep connection open
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	wsConn := newWebSocketConnection(conn)

	// Cancel context
	wsConn.cancel()

	// Wait a bit for cancellation to propagate
	time.Sleep(100 * time.Millisecond)

	// Verify context is done
	select {
	case <-wsConn.ctx.Done():
		// Expected
	default:
		t.Error("Context should be cancelled")
	}

	wsConn.Close()
}
