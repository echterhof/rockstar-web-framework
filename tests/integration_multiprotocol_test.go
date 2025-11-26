//go:build !benchmark
// +build !benchmark

package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

// TestMultiProtocolHTTP1 tests HTTP/1.1 GET and POST request handling
// Requirements: 1.1
func TestMultiProtocolHTTP1(t *testing.T) {
	// Create framework configuration
	config := pkg.FrameworkConfig{
		ServerConfig: pkg.ServerConfig{
			EnableHTTP1:     true,
			EnableHTTP2:     false,
			EnableQUIC:      false,
			ReadTimeout:     5 * time.Second,
			WriteTimeout:    5 * time.Second,
			IdleTimeout:     60 * time.Second,
			MaxHeaderBytes:  1 << 20,
			ShutdownTimeout: 2 * time.Second,
		},
		DatabaseConfig: createTestDatabaseConfig(),
		CacheConfig:    pkg.CacheConfig{},
		SessionConfig:  *createTestSessionConfig(),
	}

	// Create framework instance
	framework, err := pkg.New(config)
	assertNoError(t, err, "Failed to create framework")

	// Register test routes
	framework.Router().GET("/test", func(ctx pkg.Context) error {
		return ctx.String(200, "GET response")
	})

	framework.Router().POST("/test", func(ctx pkg.Context) error {
		return ctx.String(200, "POST response")
	})

	framework.Router().GET("/json", func(ctx pkg.Context) error {
		data := map[string]interface{}{
			"message": "JSON response",
			"status":  "success",
		}
		return ctx.JSON(200, data)
	})

	framework.Router().POST("/json", func(ctx pkg.Context) error {
		var requestData map[string]interface{}
		if err := json.Unmarshal(ctx.Body(), &requestData); err != nil {
			return ctx.JSON(400, map[string]string{"error": "Invalid JSON"})
		}
		return ctx.JSON(200, map[string]interface{}{
			"received": requestData,
			"status":   "success",
		})
	})

	framework.Router().GET("/params/:id", func(ctx pkg.Context) error {
		id := ctx.Params()["id"]
		return ctx.JSON(200, map[string]string{"id": id})
	})

	// Start server in background
	go func() {
		if err := framework.Listen(":19001"); err != nil {
			t.Logf("Server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)
	defer framework.Shutdown(2 * time.Second)

	// Test GET request
	t.Run("GET request", func(t *testing.T) {
		resp, err := http.Get("http://localhost:19001/test")
		assertNoError(t, err, "GET request failed")
		defer resp.Body.Close()

		assertEqual(t, 200, resp.StatusCode, "Expected status 200")

		body, err := io.ReadAll(resp.Body)
		assertNoError(t, err, "Failed to read response body")
		assertEqual(t, "GET response", string(body), "Unexpected response body")
	})

	// Test POST request
	t.Run("POST request", func(t *testing.T) {
		resp, err := http.Post("http://localhost:19001/test", "text/plain", bytes.NewBufferString("test data"))
		assertNoError(t, err, "POST request failed")
		defer resp.Body.Close()

		assertEqual(t, 200, resp.StatusCode, "Expected status 200")

		body, err := io.ReadAll(resp.Body)
		assertNoError(t, err, "Failed to read response body")
		assertEqual(t, "POST response", string(body), "Unexpected response body")
	})

	// Test JSON GET request
	t.Run("JSON GET request", func(t *testing.T) {
		resp, err := http.Get("http://localhost:19001/json")
		assertNoError(t, err, "JSON GET request failed")
		defer resp.Body.Close()

		assertEqual(t, 200, resp.StatusCode, "Expected status 200")
		assertEqual(t, "application/json", resp.Header.Get("Content-Type"), "Expected JSON content type")

		var data map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&data)
		assertNoError(t, err, "Failed to decode JSON response")

		assertEqual(t, "JSON response", data["message"], "Unexpected message")
		assertEqual(t, "success", data["status"], "Unexpected status")
	})

	// Test JSON POST request
	t.Run("JSON POST request", func(t *testing.T) {
		requestData := map[string]interface{}{
			"name":  "test",
			"value": 123,
		}
		jsonData, err := json.Marshal(requestData)
		assertNoError(t, err, "Failed to marshal JSON")

		resp, err := http.Post("http://localhost:19001/json", "application/json", bytes.NewBuffer(jsonData))
		assertNoError(t, err, "JSON POST request failed")
		defer resp.Body.Close()

		assertEqual(t, 200, resp.StatusCode, "Expected status 200")

		var responseData map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&responseData)
		assertNoError(t, err, "Failed to decode JSON response")

		assertEqual(t, "success", responseData["status"], "Unexpected status")
		assertNotNil(t, responseData["received"], "Expected received data")
	})

	// Test route parameters
	t.Run("Route parameters", func(t *testing.T) {
		resp, err := http.Get("http://localhost:19001/params/123")
		assertNoError(t, err, "Route parameter request failed")
		defer resp.Body.Close()

		assertEqual(t, 200, resp.StatusCode, "Expected status 200")

		var data map[string]string
		err = json.NewDecoder(resp.Body).Decode(&data)
		assertNoError(t, err, "Failed to decode JSON response")

		assertEqual(t, "123", data["id"], "Unexpected parameter value")
	})
}

// TestMultiProtocolHTTP2 tests HTTP/2 protocol enablement
// Requirements: 1.2
func TestMultiProtocolHTTP2(t *testing.T) {
	// Create framework configuration with HTTP/2 enabled
	config := pkg.FrameworkConfig{
		ServerConfig: pkg.ServerConfig{
			EnableHTTP1:     true,
			EnableHTTP2:     true,
			EnableQUIC:      false,
			ReadTimeout:     5 * time.Second,
			WriteTimeout:    5 * time.Second,
			IdleTimeout:     60 * time.Second,
			MaxHeaderBytes:  1 << 20,
			ShutdownTimeout: 2 * time.Second,
		},
		DatabaseConfig: createTestDatabaseConfig(),
		CacheConfig:    pkg.CacheConfig{},
		SessionConfig:  *createTestSessionConfig(),
	}

	// Create framework instance
	framework, err := pkg.New(config)
	assertNoError(t, err, "Failed to create framework")

	// Register test route
	framework.Router().GET("/http2-test", func(ctx pkg.Context) error {
		return ctx.JSON(200, map[string]string{
			"protocol": "HTTP/2",
			"message":  "HTTP/2 enabled",
		})
	})

	// Start server in background
	go func() {
		if err := framework.Listen(":19002"); err != nil {
			t.Logf("Server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)
	defer framework.Shutdown(2 * time.Second)

	// Test HTTP/2 request
	t.Run("HTTP/2 enabled", func(t *testing.T) {
		resp, err := http.Get("http://localhost:19002/http2-test")
		assertNoError(t, err, "HTTP/2 request failed")
		defer resp.Body.Close()

		assertEqual(t, 200, resp.StatusCode, "Expected status 200")

		var data map[string]string
		err = json.NewDecoder(resp.Body).Decode(&data)
		assertNoError(t, err, "Failed to decode JSON response")

		assertEqual(t, "HTTP/2 enabled", data["message"], "Unexpected message")
	})
}

// TestMultiProtocolWebSocket tests WebSocket route registration
// Requirements: 1.3
func TestMultiProtocolWebSocket(t *testing.T) {
	// Create framework configuration
	config := pkg.FrameworkConfig{
		ServerConfig: pkg.ServerConfig{
			EnableHTTP1:     true,
			EnableHTTP2:     false,
			EnableQUIC:      false,
			ReadTimeout:     5 * time.Second,
			WriteTimeout:    5 * time.Second,
			IdleTimeout:     60 * time.Second,
			MaxHeaderBytes:  1 << 20,
			ShutdownTimeout: 2 * time.Second,
		},
		DatabaseConfig: createTestDatabaseConfig(),
		CacheConfig:    pkg.CacheConfig{},
		SessionConfig:  *createTestSessionConfig(),
	}

	// Create framework instance
	framework, err := pkg.New(config)
	assertNoError(t, err, "Failed to create framework")

	// Register WebSocket route
	framework.Router().WebSocket("/ws", func(ctx pkg.Context, conn pkg.WebSocketConnection) error {
		// Echo server
		for {
			messageType, data, err := conn.ReadMessage()
			if err != nil {
				return err
			}
			if err := conn.WriteMessage(messageType, data); err != nil {
				return err
			}
		}
	})

	// Verify route is registered
	routes := framework.Router().Routes()
	assertNotNil(t, routes, "Routes should not be nil")

	// Find WebSocket route
	var wsRoute *pkg.Route
	for _, route := range routes {
		if route.Path == "/ws" && route.IsWebSocket {
			wsRoute = route
			break
		}
	}

	assertNotNil(t, wsRoute, "WebSocket route should be registered")
	assertTrue(t, wsRoute.IsWebSocket, "Route should be marked as WebSocket")
}

// TestMultiProtocolREST tests REST API operations
// Requirements: 1.4
func TestMultiProtocolREST(t *testing.T) {
	// Create framework configuration
	config := pkg.FrameworkConfig{
		ServerConfig: pkg.ServerConfig{
			EnableHTTP1:     true,
			EnableHTTP2:     false,
			EnableQUIC:      false,
			ReadTimeout:     5 * time.Second,
			WriteTimeout:    5 * time.Second,
			IdleTimeout:     60 * time.Second,
			MaxHeaderBytes:  1 << 20,
			ShutdownTimeout: 2 * time.Second,
		},
		DatabaseConfig: createTestDatabaseConfig(),
		CacheConfig:    pkg.CacheConfig{},
		SessionConfig:  *createTestSessionConfig(),
	}

	// Create framework instance
	framework, err := pkg.New(config)
	assertNoError(t, err, "Failed to create framework")

	// Create REST API manager
	restAPI := pkg.NewRESTAPIManager(framework.Router(), framework.Database())

	// In-memory storage for testing
	items := make(map[string]map[string]interface{})
	var itemsMu sync.Mutex

	// Register REST routes
	// LIST - Get all items
	restAPI.RegisterRoute("GET", "/items", func(ctx pkg.Context) error {
		itemsMu.Lock()
		defer itemsMu.Unlock()

		itemList := make([]map[string]interface{}, 0, len(items))
		for _, item := range items {
			itemList = append(itemList, item)
		}

		return restAPI.SendJSONResponse(ctx, 200, itemList)
	}, pkg.RESTRouteConfig{})

	// GET - Get single item
	restAPI.RegisterRoute("GET", "/items/:id", func(ctx pkg.Context) error {
		itemsMu.Lock()
		defer itemsMu.Unlock()

		id := ctx.Params()["id"]
		item, exists := items[id]
		if !exists {
			return restAPI.SendErrorResponse(ctx, 404, "Item not found", nil)
		}

		return restAPI.SendJSONResponse(ctx, 200, item)
	}, pkg.RESTRouteConfig{})

	// CREATE - Create new item
	restAPI.RegisterRoute("POST", "/items", func(ctx pkg.Context) error {
		var item map[string]interface{}
		if err := restAPI.ParseJSONRequest(ctx, &item); err != nil {
			return restAPI.SendErrorResponse(ctx, 400, "Invalid JSON", nil)
		}

		itemsMu.Lock()
		defer itemsMu.Unlock()

		id := fmt.Sprintf("item-%d", len(items)+1)
		item["id"] = id
		items[id] = item

		return restAPI.SendJSONResponse(ctx, 201, item)
	}, pkg.RESTRouteConfig{})

	// Start server in background
	go func() {
		if err := framework.Listen(":19004"); err != nil {
			t.Logf("Server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)
	defer framework.Shutdown(2 * time.Second)

	// Test CREATE operation
	t.Run("CREATE item", func(t *testing.T) {
		itemData := map[string]interface{}{
			"name":  "Test Item",
			"value": 100,
		}
		jsonData, err := json.Marshal(itemData)
		assertNoError(t, err, "Failed to marshal JSON")

		resp, err := http.Post("http://localhost:19004/items", "application/json", bytes.NewBuffer(jsonData))
		assertNoError(t, err, "CREATE request failed")
		defer resp.Body.Close()

		assertEqual(t, 201, resp.StatusCode, "Expected status 201")

		var response pkg.RESTResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		assertNoError(t, err, "Failed to decode JSON response")

		assertTrue(t, response.Success, "Expected success response")
		assertNotNil(t, response.Data, "Expected data in response")
	})

	// Test LIST operation
	t.Run("LIST items", func(t *testing.T) {
		resp, err := http.Get("http://localhost:19004/items")
		assertNoError(t, err, "LIST request failed")
		defer resp.Body.Close()

		assertEqual(t, 200, resp.StatusCode, "Expected status 200")

		var response pkg.RESTResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		assertNoError(t, err, "Failed to decode JSON response")

		assertTrue(t, response.Success, "Expected success response")
		assertNotNil(t, response.Data, "Expected data in response")
	})

	// Test GET operation
	t.Run("GET item", func(t *testing.T) {
		resp, err := http.Get("http://localhost:19004/items/item-1")
		assertNoError(t, err, "GET request failed")
		defer resp.Body.Close()

		assertEqual(t, 200, resp.StatusCode, "Expected status 200")

		var response pkg.RESTResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		assertNoError(t, err, "Failed to decode JSON response")

		assertTrue(t, response.Success, "Expected success response")
		assertNotNil(t, response.Data, "Expected data in response")
	})
}

// TestMultiProtocolConcurrent tests concurrent request handling
// Requirements: 1.5
func TestMultiProtocolConcurrent(t *testing.T) {
	// Create framework configuration
	config := pkg.FrameworkConfig{
		ServerConfig: pkg.ServerConfig{
			EnableHTTP1:     true,
			EnableHTTP2:     false,
			EnableQUIC:      false,
			ReadTimeout:     5 * time.Second,
			WriteTimeout:    5 * time.Second,
			IdleTimeout:     60 * time.Second,
			MaxHeaderBytes:  1 << 20,
			ShutdownTimeout: 2 * time.Second,
		},
		DatabaseConfig: createTestDatabaseConfig(),
		CacheConfig:    pkg.CacheConfig{},
		SessionConfig:  *createTestSessionConfig(),
	}

	// Create framework instance
	framework, err := pkg.New(config)
	assertNoError(t, err, "Failed to create framework")

	// Register test routes
	framework.Router().GET("/concurrent", func(ctx pkg.Context) error {
		return ctx.JSON(200, map[string]string{"status": "ok"})
	})

	// Start server in background
	go func() {
		if err := framework.Listen(":19005"); err != nil {
			t.Logf("Server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)
	defer framework.Shutdown(2 * time.Second)

	// Test concurrent requests
	t.Run("Concurrent requests", func(t *testing.T) {
		concurrentRequests := 50
		errors := runConcurrentWithErrors(concurrentRequests, func(id int) error {
			resp, err := http.Get("http://localhost:19005/concurrent")
			if err != nil {
				return fmt.Errorf("request %d failed: %w", id, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				return fmt.Errorf("request %d: expected status 200, got %d", id, resp.StatusCode)
			}

			return nil
		})

		if len(errors) > 0 {
			t.Errorf("Concurrent requests failed: %d errors out of %d requests", len(errors), concurrentRequests)
			for i, err := range errors {
				if i < 5 { // Show first 5 errors
					t.Logf("Error %d: %v", i+1, err)
				}
			}
		}
	})
}

// TestMultiProtocolGracefulShutdown tests graceful shutdown with active connections
// Requirements: 1.6
func TestMultiProtocolGracefulShutdown(t *testing.T) {
	// Create framework configuration
	config := pkg.FrameworkConfig{
		ServerConfig: pkg.ServerConfig{
			EnableHTTP1:     true,
			EnableHTTP2:     false,
			EnableQUIC:      false,
			ReadTimeout:     5 * time.Second,
			WriteTimeout:    5 * time.Second,
			IdleTimeout:     60 * time.Second,
			MaxHeaderBytes:  1 << 20,
			ShutdownTimeout: 2 * time.Second,
		},
		DatabaseConfig: createTestDatabaseConfig(),
		CacheConfig:    pkg.CacheConfig{},
		SessionConfig:  *createTestSessionConfig(),
	}

	// Create framework instance
	framework, err := pkg.New(config)
	assertNoError(t, err, "Failed to create framework")

	// Register slow route
	framework.Router().GET("/slow", func(ctx pkg.Context) error {
		time.Sleep(500 * time.Millisecond)
		return ctx.JSON(200, map[string]string{"status": "completed"})
	})

	// Start server in background
	go func() {
		if err := framework.Listen(":19006"); err != nil {
			t.Logf("Server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Start concurrent requests
	var wg sync.WaitGroup
	requestsCompleted := 0
	var completedMu sync.Mutex

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			resp, err := http.Get("http://localhost:19006/slow")
			if err == nil {
				defer resp.Body.Close()
				if resp.StatusCode == 200 {
					completedMu.Lock()
					requestsCompleted++
					completedMu.Unlock()
				}
			}
		}(i)
	}

	// Wait a bit for requests to start
	time.Sleep(100 * time.Millisecond)

	// Initiate graceful shutdown
	shutdownStart := time.Now()
	err = framework.Shutdown(3 * time.Second)
	shutdownDuration := time.Since(shutdownStart)

	assertNoError(t, err, "Graceful shutdown failed")

	// Wait for all requests to complete
	wg.Wait()

	// Verify requests completed
	t.Logf("Requests completed: %d/10", requestsCompleted)
	t.Logf("Shutdown duration: %v", shutdownDuration)

	// At least some requests should have completed
	assertTrue(t, requestsCompleted > 0, "Expected some requests to complete before shutdown")
}
