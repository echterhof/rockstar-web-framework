package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

// TestMultiProtocolHTTP1 tests HTTP/1.1 protocol support end-to-end
func TestMultiProtocolHTTP1(t *testing.T) {
	// Setup server
	config := pkg.ServerConfig{
		EnableHTTP1:    true,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	server := pkg.NewServer(config)
	router := pkg.NewRouter()

	// Add test routes
	router.GET("/api/users", func(ctx pkg.Context) error {
		users := []map[string]string{
			{"id": "1", "name": "Alice"},
			{"id": "2", "name": "Bob"},
		}
		return ctx.JSON(http.StatusOK, users)
	})

	router.POST("/api/users", func(ctx pkg.Context) error {
		var user map[string]string
		if err := json.Unmarshal(ctx.Body(), &user); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
		}
		user["id"] = "3"
		return ctx.JSON(http.StatusCreated, user)
	})

	server.SetRouter(router)

	// Start server
	addr := "127.0.0.1:19001"
	if err := server.Listen(addr); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Close()

	time.Sleep(100 * time.Millisecond)

	// Test GET request
	resp, err := http.Get("http://" + addr + "/api/users")
	if err != nil {
		t.Fatalf("GET request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var users []map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}

	// Test POST request
	newUser := map[string]string{"name": "Charlie"}
	jsonData, _ := json.Marshal(newUser)

	resp, err = http.Post("http://"+addr+"/api/users", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("POST request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}

	var createdUser map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&createdUser); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if createdUser["name"] != "Charlie" {
		t.Errorf("Expected name 'Charlie', got '%s'", createdUser["name"])
	}
}

// TestMultiProtocolHTTP2 tests HTTP/2 protocol support
func TestMultiProtocolHTTP2(t *testing.T) {
	config := pkg.ServerConfig{
		EnableHTTP1:    true,
		EnableHTTP2:    true,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	server := pkg.NewServer(config)
	server.EnableHTTP2()

	router := pkg.NewRouter()

	router.GET("/api/test", func(ctx pkg.Context) error {
		return ctx.JSON(http.StatusOK, map[string]string{
			"protocol": "HTTP/2",
			"message":  "Hello from HTTP/2",
		})
	})

	server.SetRouter(router)

	// Verify HTTP/2 is enabled
	protocol := server.Protocol()
	if protocol != "HTTP/1.1, HTTP/2" {
		t.Errorf("Expected 'HTTP/1.1, HTTP/2', got '%s'", protocol)
	}
}

// TestMultiProtocolWebSocket tests WebSocket protocol support
func TestMultiProtocolWebSocket(t *testing.T) {
	config := pkg.ServerConfig{
		EnableHTTP1:  true,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	server := pkg.NewServer(config)
	router := pkg.NewRouter()

	// WebSocket echo handler
	router.WebSocket("/ws/echo", func(ctx pkg.Context, conn pkg.WebSocketConnection) error {
		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				return err
			}

			if err := conn.WriteMessage(messageType, message); err != nil {
				return err
			}
		}
	})

	server.SetRouter(router)

	addr := "127.0.0.1:19002"
	if err := server.Listen(addr); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Close()

	time.Sleep(100 * time.Millisecond)

	// Note: Full WebSocket client test would require gorilla/websocket client
	// This test verifies the route is registered
	route, _, found := router.Match("GET", "/ws/echo", "")
	if !found {
		t.Error("WebSocket route not found")
	}

	if !route.IsWebSocket {
		t.Error("Route should be marked as WebSocket")
	}
}

// TestMultiProtocolREST tests REST API protocol support
func TestMultiProtocolREST(t *testing.T) {
	config := pkg.ServerConfig{
		EnableHTTP1:  true,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	server := pkg.NewServer(config)
	router := pkg.NewRouter()

	// Define RESTful routes manually
	router.GET("/api/posts", func(ctx pkg.Context) error {
		posts := []map[string]interface{}{
			{"id": 1, "title": "First Post"},
			{"id": 2, "title": "Second Post"},
		}
		return ctx.JSON(http.StatusOK, posts)
	})

	router.GET("/api/posts/:id", func(ctx pkg.Context) error {
		id := ctx.Params()["id"]
		post := map[string]interface{}{
			"id":    id,
			"title": "Post " + id,
		}
		return ctx.JSON(http.StatusOK, post)
	})

	router.POST("/api/posts", func(ctx pkg.Context) error {
		var post map[string]interface{}
		if err := json.Unmarshal(ctx.Body(), &post); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
		}
		post["id"] = 3
		return ctx.JSON(http.StatusCreated, post)
	})

	server.SetRouter(router)

	addr := "127.0.0.1:19003"
	if err := server.Listen(addr); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Close()

	time.Sleep(100 * time.Millisecond)

	// Test LIST
	resp, err := http.Get("http://" + addr + "/api/posts")
	if err != nil {
		t.Fatalf("LIST request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Test GET
	resp, err = http.Get("http://" + addr + "/api/posts/1")
	if err != nil {
		t.Fatalf("GET request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Test CREATE
	newPost := map[string]string{"title": "New Post"}
	jsonData, _ := json.Marshal(newPost)

	resp, err = http.Post("http://"+addr+"/api/posts", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("CREATE request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}
}

// TestMultiProtocolGraphQL tests GraphQL protocol support
// Note: GraphQL implementation is pending - test skipped
func TestMultiProtocolGraphQL(t *testing.T) {
	t.Skip("GraphQL implementation pending")
}

// TestMultiProtocolGRPC tests gRPC protocol support
// Note: gRPC implementation is pending - test skipped
func TestMultiProtocolGRPC(t *testing.T) {
	t.Skip("gRPC implementation pending")
}

// TestMultiProtocolSOAP tests SOAP protocol support
// Note: SOAP implementation is pending - test skipped
func TestMultiProtocolSOAP(t *testing.T) {
	t.Skip("SOAP implementation pending")
}

// TestMultiProtocolConcurrent tests concurrent requests across multiple protocols
func TestMultiProtocolConcurrent(t *testing.T) {
	config := pkg.ServerConfig{
		EnableHTTP1:  true,
		EnableHTTP2:  true,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	server := pkg.NewServer(config)
	router := pkg.NewRouter()

	// Add routes for different protocols
	router.GET("/api/http", func(ctx pkg.Context) error {
		return ctx.JSON(http.StatusOK, map[string]string{"protocol": "HTTP"})
	})

	router.GET("/api/rest", func(ctx pkg.Context) error {
		return ctx.JSON(http.StatusOK, []string{"item1", "item2"})
	})

	server.SetRouter(router)

	addr := "127.0.0.1:19006"
	if err := server.Listen(addr); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Close()

	time.Sleep(100 * time.Millisecond)

	// Make concurrent requests
	done := make(chan bool)
	errors := make(chan error, 20)

	for i := 0; i < 10; i++ {
		go func() {
			resp, err := http.Get("http://" + addr + "/api/http")
			if err != nil {
				errors <- err
			} else {
				resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					errors <- fmt.Errorf("expected status 200, got %d", resp.StatusCode)
				}
			}
			done <- true
		}()

		go func() {
			resp, err := http.Get("http://" + addr + "/api/rest")
			if err != nil {
				errors <- err
			} else {
				resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					errors <- fmt.Errorf("expected status 200, got %d", resp.StatusCode)
				}
			}
			done <- true
		}()
	}

	// Wait for all requests
	for i := 0; i < 20; i++ {
		<-done
	}

	close(errors)
	for err := range errors {
		t.Errorf("Concurrent request error: %v", err)
	}
}

// TestMultiProtocolGracefulShutdown tests graceful shutdown with active connections
func TestMultiProtocolGracefulShutdown(t *testing.T) {
	config := pkg.ServerConfig{
		EnableHTTP1:     true,
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    5 * time.Second,
		ShutdownTimeout: 3 * time.Second,
	}

	server := pkg.NewServer(config)
	router := pkg.NewRouter()

	router.GET("/slow", func(ctx pkg.Context) error {
		time.Sleep(500 * time.Millisecond)
		return ctx.String(http.StatusOK, "Done")
	})

	server.SetRouter(router)

	addr := "127.0.0.1:19007"
	if err := server.Listen(addr); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Start a slow request
	done := make(chan bool)
	go func() {
		resp, err := http.Get("http://" + addr + "/slow")
		if err == nil {
			resp.Body.Close()
		}
		done <- true
	}()

	// Give request time to start
	time.Sleep(100 * time.Millisecond)

	// Initiate graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		t.Errorf("Graceful shutdown failed: %v", err)
	}

	// Wait for request to complete
	<-done

	if server.IsRunning() {
		t.Error("Server should not be running after shutdown")
	}
}
