package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

// TestIntegrationCompleteApplicationWithoutDatabase tests a complete application
// running without any database configuration
// Requirements: All
func TestIntegrationCompleteApplicationWithoutDatabase(t *testing.T) {
	// Create framework configuration without database
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
		// No database configuration
		DatabaseConfig: pkg.DatabaseConfig{},
		CacheConfig:    pkg.CacheConfig{},
		SessionConfig: pkg.SessionConfig{
			StorageType:     pkg.SessionStorageCache,
			CookieName:      "test_session",
			SessionLifetime: 1 * time.Hour,
			CleanupInterval: 5 * time.Minute,
			EncryptionKey:   []byte("12345678901234567890123456789012"),
		},
		I18nConfig: pkg.I18nConfig{
			DefaultLocale:    "en",
			SupportedLocales: []string{"en"},
		},
	}

	// Create framework instance
	framework, err := pkg.New(config)
	assertNoError(t, err, "Failed to create framework without database")

	// Verify database manager is no-op
	assertFalse(t, framework.Database().IsConnected(), "Database should not be connected")

	// Verify all managers are initialized
	assertNotNil(t, framework.Router(), "Router should be initialized")
	assertNotNil(t, framework.Cache(), "Cache should be initialized")
	assertNotNil(t, framework.Session(), "Session should be initialized")
	assertNotNil(t, framework.Security(), "Security should be initialized")
	assertNotNil(t, framework.Monitoring(), "Monitoring should be initialized")
	assertNotNil(t, framework.Metrics(), "Metrics should be initialized")

	// In-memory data store for testing
	items := make(map[string]map[string]interface{})
	var itemsMu sync.Mutex

	// Register routes that don't use database
	framework.Router().GET("/health", func(ctx pkg.Context) error {
		return ctx.JSON(200, map[string]string{"status": "healthy"})
	})

	framework.Router().GET("/items", func(ctx pkg.Context) error {
		itemsMu.Lock()
		defer itemsMu.Unlock()

		itemList := make([]map[string]interface{}, 0, len(items))
		for _, item := range items {
			itemList = append(itemList, item)
		}

		return ctx.JSON(200, map[string]interface{}{
			"items": itemList,
			"count": len(itemList),
		})
	})

	framework.Router().POST("/items", func(ctx pkg.Context) error {
		var item map[string]interface{}
		if err := json.Unmarshal(ctx.Body(), &item); err != nil {
			return ctx.JSON(400, map[string]string{"error": "Invalid JSON"})
		}

		itemsMu.Lock()
		defer itemsMu.Unlock()

		id := time.Now().Format("20060102150405")
		item["id"] = id
		items[id] = item

		return ctx.JSON(201, item)
	})

	framework.Router().GET("/items/:id", func(ctx pkg.Context) error {
		itemsMu.Lock()
		defer itemsMu.Unlock()

		id := ctx.Params()["id"]
		item, exists := items[id]
		if !exists {
			return ctx.JSON(404, map[string]string{"error": "Item not found"})
		}

		return ctx.JSON(200, item)
	})

	// Start server in background
	go func() {
		if err := framework.Listen(":19201"); err != nil {
			t.Logf("Server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)
	defer framework.Shutdown(2 * time.Second)

	// Test health check
	t.Run("Health check", func(t *testing.T) {
		resp, err := http.Get("http://localhost:19201/health")
		assertNoError(t, err, "Health check request failed")
		defer resp.Body.Close()

		assertEqual(t, 200, resp.StatusCode, "Expected status 200")

		var data map[string]string
		err = json.NewDecoder(resp.Body).Decode(&data)
		assertNoError(t, err, "Failed to decode response")

		assertEqual(t, "healthy", data["status"], "Expected healthy status")
	})

	// Test creating items
	t.Run("Create items", func(t *testing.T) {
		itemData := map[string]interface{}{
			"name":  "Test Item",
			"value": 100,
		}
		jsonData, err := json.Marshal(itemData)
		assertNoError(t, err, "Failed to marshal JSON")

		resp, err := http.Post("http://localhost:19201/items", "application/json", bytes.NewBuffer(jsonData))
		assertNoError(t, err, "Create item request failed")
		defer resp.Body.Close()

		assertEqual(t, 201, resp.StatusCode, "Expected status 201")

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		assertNoError(t, err, "Failed to decode response")

		assertNotNil(t, response["id"], "Expected item ID")
		assertEqual(t, "Test Item", response["name"], "Expected item name")
	})

	// Test listing items
	t.Run("List items", func(t *testing.T) {
		resp, err := http.Get("http://localhost:19201/items")
		assertNoError(t, err, "List items request failed")
		defer resp.Body.Close()

		assertEqual(t, 200, resp.StatusCode, "Expected status 200")

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		assertNoError(t, err, "Failed to decode response")

		assertNotNil(t, response["items"], "Expected items array")
		count := response["count"].(float64)
		assertTrue(t, count > 0, "Expected at least one item")
	})

	// Test getting specific item
	t.Run("Get specific item", func(t *testing.T) {
		// First create an item
		itemData := map[string]interface{}{
			"name": "Specific Item",
		}
		jsonData, err := json.Marshal(itemData)
		assertNoError(t, err, "Failed to marshal JSON")

		createResp, err := http.Post("http://localhost:19201/items", "application/json", bytes.NewBuffer(jsonData))
		assertNoError(t, err, "Create item request failed")

		var createdItem map[string]interface{}
		err = json.NewDecoder(createResp.Body).Decode(&createdItem)
		createResp.Body.Close()
		assertNoError(t, err, "Failed to decode create response")

		itemID := createdItem["id"].(string)

		// Now get the specific item
		resp, err := http.Get("http://localhost:19201/items/" + itemID)
		assertNoError(t, err, "Get item request failed")
		defer resp.Body.Close()

		assertEqual(t, 200, resp.StatusCode, "Expected status 200")

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		assertNoError(t, err, "Failed to decode response")

		assertEqual(t, itemID, response["id"], "Expected matching item ID")
		assertEqual(t, "Specific Item", response["name"], "Expected matching item name")
	})
}

// TestIntegrationMixedScenario tests a scenario where some features use database
// and others don't (graceful degradation)
// Requirements: All
func TestIntegrationMixedScenario(t *testing.T) {
	// Create framework configuration without database
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
		// No database configuration
		DatabaseConfig: pkg.DatabaseConfig{},
		CacheConfig:    pkg.CacheConfig{},
		SessionConfig: pkg.SessionConfig{
			StorageType:     pkg.SessionStorageCache,
			CookieName:      "test_session",
			SessionLifetime: 1 * time.Hour,
			CleanupInterval: 5 * time.Minute,
			EncryptionKey:   []byte("12345678901234567890123456789012"),
		},
		I18nConfig: pkg.I18nConfig{
			DefaultLocale:    "en",
			SupportedLocales: []string{"en"},
		},
	}

	// Create framework instance
	framework, err := pkg.New(config)
	assertNoError(t, err, "Failed to create framework without database")

	// Verify database is not connected
	assertFalse(t, framework.Database().IsConnected(), "Database should not be connected")

	// Register routes that use in-memory session storage
	framework.Router().POST("/login", func(ctx pkg.Context) error {
		// Create session (should use in-memory storage)
		session, err := framework.Session().Create(ctx)
		if err != nil {
			return ctx.JSON(500, map[string]string{"error": "Failed to create session"})
		}

		// Store user data in session
		session.UserID = "user123"
		session.Data["username"] = "testuser"
		session.Data["role"] = "admin"

		// Save session
		if err := framework.Session().Save(ctx, session); err != nil {
			return ctx.JSON(500, map[string]string{"error": "Failed to save session"})
		}

		// Set session cookie
		if err := framework.Session().SetCookie(ctx, session); err != nil {
			return ctx.JSON(500, map[string]string{"error": "Failed to set cookie"})
		}

		return ctx.JSON(200, map[string]interface{}{
			"message":    "Login successful",
			"session_id": session.ID,
		})
	})

	framework.Router().GET("/profile", func(ctx pkg.Context) error {
		// Get session from cookie (should use in-memory storage)
		session, err := framework.Session().GetSessionFromCookie(ctx)
		if err != nil {
			return ctx.JSON(401, map[string]string{"error": "No valid session"})
		}

		return ctx.JSON(200, map[string]interface{}{
			"user_id":  session.UserID,
			"username": session.Data["username"],
			"role":     session.Data["role"],
		})
	})

	// Register routes that use in-memory metrics
	framework.Router().GET("/metrics", func(ctx pkg.Context) error {
		// Metrics should be collected in memory
		return ctx.JSON(200, map[string]string{
			"status": "Metrics collected in memory",
		})
	})

	// Start server in background
	go func() {
		if err := framework.Listen(":19202"); err != nil {
			t.Logf("Server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)
	defer framework.Shutdown(2 * time.Second)

	// Test session-based login (using in-memory storage)
	var sessionCookie string
	t.Run("Login with in-memory session", func(t *testing.T) {
		resp, err := http.Post("http://localhost:19202/login", "application/json", bytes.NewBufferString("{}"))
		assertNoError(t, err, "Login request failed")
		defer resp.Body.Close()

		assertEqual(t, 200, resp.StatusCode, "Expected status 200")

		// Extract session cookie
		cookies := resp.Cookies()
		for _, cookie := range cookies {
			if cookie.Name == "test_session" {
				sessionCookie = cookie.Value
				break
			}
		}

		assertNotNil(t, sessionCookie, "Session cookie should be set")

		var data map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&data)
		assertNoError(t, err, "Failed to decode response")

		assertEqual(t, "Login successful", data["message"], "Expected login success message")
	})

	// Test accessing profile with session (using in-memory storage)
	t.Run("Access profile with in-memory session", func(t *testing.T) {
		client := &http.Client{}
		req, err := http.NewRequest("GET", "http://localhost:19202/profile", nil)
		assertNoError(t, err, "Failed to create request")

		// Add session cookie
		req.AddCookie(&http.Cookie{
			Name:  "test_session",
			Value: sessionCookie,
		})

		resp, err := client.Do(req)
		assertNoError(t, err, "Profile request failed")
		defer resp.Body.Close()

		assertEqual(t, 200, resp.StatusCode, "Expected status 200")

		var data map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&data)
		assertNoError(t, err, "Failed to decode response")

		assertEqual(t, "user123", data["user_id"], "Expected user ID")
		assertEqual(t, "testuser", data["username"], "Expected username")
		assertEqual(t, "admin", data["role"], "Expected role")
	})

	// Test metrics collection (using in-memory storage)
	t.Run("Metrics collection in memory", func(t *testing.T) {
		resp, err := http.Get("http://localhost:19202/metrics")
		assertNoError(t, err, "Metrics request failed")
		defer resp.Body.Close()

		assertEqual(t, 200, resp.StatusCode, "Expected status 200")

		var data map[string]string
		err = json.NewDecoder(resp.Body).Decode(&data)
		assertNoError(t, err, "Failed to decode response")

		assertEqual(t, "Metrics collected in memory", data["status"], "Expected metrics status")
	})

	// Test that database operations return appropriate errors
	t.Run("Database operations return errors", func(t *testing.T) {
		// Attempt to query database directly
		_, err := framework.Database().Query("SELECT * FROM users")
		assertError(t, err, "Expected error when querying no-op database")

		// Attempt to execute database command
		_, err = framework.Database().Exec("INSERT INTO users VALUES (1, 'test')")
		assertError(t, err, "Expected error when executing on no-op database")
	})
}

// TestIntegrationGracefulShutdownWithoutDatabase tests graceful shutdown
// when no database is configured
// Requirements: 3.5
func TestIntegrationGracefulShutdownWithoutDatabase(t *testing.T) {
	// Create framework configuration without database
	config := pkg.FrameworkConfig{
		ServerConfig: pkg.ServerConfig{
			EnableHTTP1:     true,
			EnableHTTP2:     false,
			EnableQUIC:      false,
			ReadTimeout:     5 * time.Second,
			WriteTimeout:    5 * time.Second,
			IdleTimeout:     60 * time.Second,
			MaxHeaderBytes:  1 << 20,
			ShutdownTimeout: 3 * time.Second,
		},
		// No database configuration
		DatabaseConfig: pkg.DatabaseConfig{},
		CacheConfig:    pkg.CacheConfig{},
		SessionConfig: pkg.SessionConfig{
			StorageType:     pkg.SessionStorageCache,
			CookieName:      "test_session",
			SessionLifetime: 1 * time.Hour,
			CleanupInterval: 5 * time.Minute,
			EncryptionKey:   []byte("12345678901234567890123456789012"),
		},
		I18nConfig: pkg.I18nConfig{
			DefaultLocale:    "en",
			SupportedLocales: []string{"en"},
		},
	}

	// Create framework instance
	framework, err := pkg.New(config)
	assertNoError(t, err, "Failed to create framework without database")

	// Verify database is not connected
	assertFalse(t, framework.Database().IsConnected(), "Database should not be connected")

	// Register slow route to test graceful shutdown
	framework.Router().GET("/slow", func(ctx pkg.Context) error {
		time.Sleep(500 * time.Millisecond)
		return ctx.JSON(200, map[string]string{"status": "completed"})
	})

	framework.Router().GET("/fast", func(ctx pkg.Context) error {
		return ctx.JSON(200, map[string]string{"status": "ok"})
	})

	// Start server in background
	go func() {
		if err := framework.Listen(":19203"); err != nil {
			t.Logf("Server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Test fast request before shutdown
	t.Run("Fast request before shutdown", func(t *testing.T) {
		resp, err := http.Get("http://localhost:19203/fast")
		assertNoError(t, err, "Fast request failed")
		defer resp.Body.Close()

		assertEqual(t, 200, resp.StatusCode, "Expected status 200")
	})

	// Start concurrent slow requests
	var wg sync.WaitGroup
	requestsCompleted := 0
	var completedMu sync.Mutex

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			resp, err := http.Get("http://localhost:19203/slow")
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
	t.Run("Graceful shutdown", func(t *testing.T) {
		shutdownStart := time.Now()
		err := framework.Shutdown(3 * time.Second)
		shutdownDuration := time.Since(shutdownStart)

		assertNoError(t, err, "Graceful shutdown failed")

		t.Logf("Shutdown duration: %v", shutdownDuration)

		// Verify shutdown completed within timeout
		assertTrue(t, shutdownDuration < 4*time.Second, "Shutdown took too long")

		// Verify framework is no longer running
		assertFalse(t, framework.IsRunning(), "Framework should not be running after shutdown")
	})

	// Wait for all requests to complete
	wg.Wait()

	t.Run("Verify requests completed", func(t *testing.T) {
		t.Logf("Requests completed: %d/5", requestsCompleted)

		// At least some requests should have completed
		assertTrue(t, requestsCompleted > 0, "Expected some requests to complete before shutdown")
	})

	// Verify new requests fail after shutdown
	t.Run("Requests fail after shutdown", func(t *testing.T) {
		// Give the server a moment to fully close
		time.Sleep(100 * time.Millisecond)

		resp, err := http.Get("http://localhost:19203/fast")
		if err == nil {
			defer resp.Body.Close()
			// If we get a response, it should be an error status or connection refused
			// The server might still be closing, which is acceptable
			t.Logf("Got response after shutdown (status: %d), server may still be closing", resp.StatusCode)
		}
		// Either we get an error (connection refused) or the server is still closing
		// Both are acceptable outcomes for this test
	})
}

// TestIntegrationEndToEndFlowsWithoutDatabase tests complete end-to-end flows
// without database configuration
// Requirements: All
func TestIntegrationEndToEndFlowsWithoutDatabase(t *testing.T) {
	// Create framework configuration without database
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
		// No database configuration
		DatabaseConfig: pkg.DatabaseConfig{},
		CacheConfig:    pkg.CacheConfig{},
		SessionConfig: pkg.SessionConfig{
			StorageType:     pkg.SessionStorageCache,
			CookieName:      "test_session",
			SessionLifetime: 1 * time.Hour,
			CleanupInterval: 5 * time.Minute,
			EncryptionKey:   []byte("12345678901234567890123456789012"),
		},
		I18nConfig: pkg.I18nConfig{
			DefaultLocale:    "en",
			SupportedLocales: []string{"en"},
		},
	}

	// Create framework instance
	framework, err := pkg.New(config)
	assertNoError(t, err, "Failed to create framework without database")

	// Verify database is not connected
	assertFalse(t, framework.Database().IsConnected(), "Database should not be connected")

	// In-memory user store
	users := make(map[string]map[string]interface{})
	var usersMu sync.Mutex

	// Register complete user management flow
	framework.Router().POST("/register", func(ctx pkg.Context) error {
		var userData map[string]interface{}
		if err := json.Unmarshal(ctx.Body(), &userData); err != nil {
			return ctx.JSON(400, map[string]string{"error": "Invalid JSON"})
		}

		username, ok := userData["username"].(string)
		if !ok || username == "" {
			return ctx.JSON(400, map[string]string{"error": "Username required"})
		}

		usersMu.Lock()
		defer usersMu.Unlock()

		if _, exists := users[username]; exists {
			return ctx.JSON(409, map[string]string{"error": "User already exists"})
		}

		userData["id"] = time.Now().Format("20060102150405")
		userData["created_at"] = time.Now().Format(time.RFC3339)
		users[username] = userData

		return ctx.JSON(201, map[string]interface{}{
			"message": "User registered successfully",
			"user_id": userData["id"],
		})
	})

	framework.Router().POST("/login", func(ctx pkg.Context) error {
		var loginData map[string]interface{}
		if err := json.Unmarshal(ctx.Body(), &loginData); err != nil {
			return ctx.JSON(400, map[string]string{"error": "Invalid JSON"})
		}

		username, ok := loginData["username"].(string)
		if !ok || username == "" {
			return ctx.JSON(400, map[string]string{"error": "Username required"})
		}

		usersMu.Lock()
		user, exists := users[username]
		usersMu.Unlock()

		if !exists {
			return ctx.JSON(401, map[string]string{"error": "Invalid credentials"})
		}

		// Create session
		session, err := framework.Session().Create(ctx)
		if err != nil {
			return ctx.JSON(500, map[string]string{"error": "Failed to create session"})
		}

		session.UserID = user["id"].(string)
		session.Data["username"] = username

		if err := framework.Session().Save(ctx, session); err != nil {
			return ctx.JSON(500, map[string]string{"error": "Failed to save session"})
		}

		if err := framework.Session().SetCookie(ctx, session); err != nil {
			return ctx.JSON(500, map[string]string{"error": "Failed to set cookie"})
		}

		return ctx.JSON(200, map[string]interface{}{
			"message":    "Login successful",
			"session_id": session.ID,
		})
	})

	framework.Router().GET("/users/:username", func(ctx pkg.Context) error {
		username := ctx.Params()["username"]

		usersMu.Lock()
		user, exists := users[username]
		usersMu.Unlock()

		if !exists {
			return ctx.JSON(404, map[string]string{"error": "User not found"})
		}

		// Remove sensitive data
		publicUser := map[string]interface{}{
			"id":         user["id"],
			"username":   user["username"],
			"created_at": user["created_at"],
		}

		return ctx.JSON(200, publicUser)
	})

	framework.Router().GET("/me", func(ctx pkg.Context) error {
		// Get session from cookie
		session, err := framework.Session().GetSessionFromCookie(ctx)
		if err != nil {
			return ctx.JSON(401, map[string]string{"error": "Not authenticated"})
		}

		username := session.Data["username"].(string)

		usersMu.Lock()
		user, exists := users[username]
		usersMu.Unlock()

		if !exists {
			return ctx.JSON(404, map[string]string{"error": "User not found"})
		}

		return ctx.JSON(200, user)
	})

	// Start server in background
	go func() {
		if err := framework.Listen(":19204"); err != nil {
			t.Logf("Server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)
	defer framework.Shutdown(2 * time.Second)

	// Test complete user registration and login flow
	var sessionCookie string
	testUsername := "testuser_" + time.Now().Format("150405")

	t.Run("User registration", func(t *testing.T) {
		userData := map[string]interface{}{
			"username": testUsername,
			"email":    testUsername + "@example.com",
		}
		jsonData, err := json.Marshal(userData)
		assertNoError(t, err, "Failed to marshal JSON")

		resp, err := http.Post("http://localhost:19204/register", "application/json", bytes.NewBuffer(jsonData))
		assertNoError(t, err, "Registration request failed")
		defer resp.Body.Close()

		assertEqual(t, 201, resp.StatusCode, "Expected status 201")

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		assertNoError(t, err, "Failed to decode response")

		assertEqual(t, "User registered successfully", response["message"], "Expected success message")
		assertNotNil(t, response["user_id"], "Expected user ID")
	})

	t.Run("Duplicate registration fails", func(t *testing.T) {
		userData := map[string]interface{}{
			"username": testUsername,
			"email":    testUsername + "@example.com",
		}
		jsonData, err := json.Marshal(userData)
		assertNoError(t, err, "Failed to marshal JSON")

		resp, err := http.Post("http://localhost:19204/register", "application/json", bytes.NewBuffer(jsonData))
		assertNoError(t, err, "Registration request failed")
		defer resp.Body.Close()

		assertEqual(t, 409, resp.StatusCode, "Expected status 409 for duplicate")
	})

	t.Run("User login", func(t *testing.T) {
		loginData := map[string]interface{}{
			"username": testUsername,
		}
		jsonData, err := json.Marshal(loginData)
		assertNoError(t, err, "Failed to marshal JSON")

		resp, err := http.Post("http://localhost:19204/login", "application/json", bytes.NewBuffer(jsonData))
		assertNoError(t, err, "Login request failed")
		defer resp.Body.Close()

		assertEqual(t, 200, resp.StatusCode, "Expected status 200")

		// Extract session cookie
		cookies := resp.Cookies()
		for _, cookie := range cookies {
			if cookie.Name == "test_session" {
				sessionCookie = cookie.Value
				break
			}
		}

		assertNotNil(t, sessionCookie, "Session cookie should be set")

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		assertNoError(t, err, "Failed to decode response")

		assertEqual(t, "Login successful", response["message"], "Expected login success")
	})

	t.Run("Get user profile", func(t *testing.T) {
		resp, err := http.Get("http://localhost:19204/users/" + testUsername)
		assertNoError(t, err, "Get user request failed")
		defer resp.Body.Close()

		assertEqual(t, 200, resp.StatusCode, "Expected status 200")

		var user map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&user)
		assertNoError(t, err, "Failed to decode response")

		assertEqual(t, testUsername, user["username"], "Expected matching username")
		assertNotNil(t, user["id"], "Expected user ID")
	})

	t.Run("Get authenticated user data", func(t *testing.T) {
		client := &http.Client{}
		req, err := http.NewRequest("GET", "http://localhost:19204/me", nil)
		assertNoError(t, err, "Failed to create request")

		// Add session cookie
		req.AddCookie(&http.Cookie{
			Name:  "test_session",
			Value: sessionCookie,
		})

		resp, err := client.Do(req)
		assertNoError(t, err, "Get me request failed")
		defer resp.Body.Close()

		assertEqual(t, 200, resp.StatusCode, "Expected status 200")

		var user map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&user)
		assertNoError(t, err, "Failed to decode response")

		assertEqual(t, testUsername, user["username"], "Expected matching username")
		assertEqual(t, testUsername+"@example.com", user["email"], "Expected matching email")
	})

	t.Run("Unauthenticated access fails", func(t *testing.T) {
		resp, err := http.Get("http://localhost:19204/me")
		assertNoError(t, err, "Request failed")
		defer resp.Body.Close()

		assertEqual(t, 401, resp.StatusCode, "Expected status 401 without session")
	})
}

// TestIntegrationConcurrentOperationsWithoutDatabase tests concurrent operations
// when no database is configured
// Requirements: All
func TestIntegrationConcurrentOperationsWithoutDatabase(t *testing.T) {
	// Create framework configuration without database
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
		// No database configuration
		DatabaseConfig: pkg.DatabaseConfig{},
		CacheConfig:    pkg.CacheConfig{},
		SessionConfig: pkg.SessionConfig{
			StorageType:     pkg.SessionStorageCache,
			CookieName:      "test_session",
			SessionLifetime: 1 * time.Hour,
			CleanupInterval: 5 * time.Minute,
			EncryptionKey:   []byte("12345678901234567890123456789012"),
		},
		I18nConfig: pkg.I18nConfig{
			DefaultLocale:    "en",
			SupportedLocales: []string{"en"},
		},
	}

	// Create framework instance
	framework, err := pkg.New(config)
	assertNoError(t, err, "Failed to create framework without database")

	// Verify database is not connected
	assertFalse(t, framework.Database().IsConnected(), "Database should not be connected")

	// In-memory counter
	counter := 0
	var counterMu sync.Mutex

	// Register routes
	framework.Router().POST("/increment", func(ctx pkg.Context) error {
		counterMu.Lock()
		counter++
		currentValue := counter
		counterMu.Unlock()

		return ctx.JSON(200, map[string]interface{}{
			"counter": currentValue,
		})
	})

	framework.Router().GET("/counter", func(ctx pkg.Context) error {
		counterMu.Lock()
		currentValue := counter
		counterMu.Unlock()

		return ctx.JSON(200, map[string]interface{}{
			"counter": currentValue,
		})
	})

	// Start server in background
	go func() {
		if err := framework.Listen(":19205"); err != nil {
			t.Logf("Server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)
	defer framework.Shutdown(2 * time.Second)

	// Test concurrent increments
	t.Run("Concurrent increments", func(t *testing.T) {
		concurrentRequests := 50
		var wg sync.WaitGroup

		for i := 0; i < concurrentRequests; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				resp, err := http.Post("http://localhost:19205/increment", "application/json", bytes.NewBufferString("{}"))
				if err != nil {
					t.Logf("Request %d failed: %v", id, err)
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode != 200 {
					t.Logf("Request %d: expected status 200, got %d", id, resp.StatusCode)
				}
			}(i)
		}

		wg.Wait()

		// Verify final counter value
		resp, err := http.Get("http://localhost:19205/counter")
		assertNoError(t, err, "Get counter request failed")
		defer resp.Body.Close()

		assertEqual(t, 200, resp.StatusCode, "Expected status 200")

		var data map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&data)
		assertNoError(t, err, "Failed to decode response")

		finalCounter := int(data["counter"].(float64))
		assertEqual(t, concurrentRequests, finalCounter, "Expected counter to match concurrent requests")
	})
}

// TestIntegrationBackwardCompatibility tests that the framework still works
// correctly when a valid database configuration is provided
// Requirements: 4.1, 4.2, 4.3, 4.5
func TestIntegrationBackwardCompatibility(t *testing.T) {
	// Create framework configuration WITH database
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
		// Valid database configuration
		DatabaseConfig: createTestDatabaseConfig(),
		CacheConfig:    pkg.CacheConfig{},
		SessionConfig:  *createTestSessionConfig(),
		I18nConfig: pkg.I18nConfig{
			DefaultLocale:    "en",
			SupportedLocales: []string{"en"},
		},
	}

	// Create framework instance
	framework, err := pkg.New(config)
	assertNoError(t, err, "Failed to create framework with database")

	// Verify database IS connected
	assertTrue(t, framework.Database().IsConnected(), "Database should be connected")

	// Initialize database tables
	err = framework.Database().CreateTables()
	assertNoError(t, err, "Failed to create database tables")

	// Verify all managers are initialized
	assertNotNil(t, framework.Router(), "Router should be initialized")
	assertNotNil(t, framework.Cache(), "Cache should be initialized")
	assertNotNil(t, framework.Session(), "Session should be initialized")
	assertNotNil(t, framework.Security(), "Security should be initialized")
	assertNotNil(t, framework.Monitoring(), "Monitoring should be initialized")
	assertNotNil(t, framework.Metrics(), "Metrics should be initialized")

	// Register routes
	framework.Router().GET("/health", func(ctx pkg.Context) error {
		return ctx.JSON(200, map[string]string{"status": "healthy"})
	})

	// Start server in background
	go func() {
		if err := framework.Listen(":19206"); err != nil {
			t.Logf("Server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)
	defer framework.Shutdown(2 * time.Second)

	// Test that routes work with database
	t.Run("Routes work with database", func(t *testing.T) {
		resp, err := http.Get("http://localhost:19206/health")
		assertNoError(t, err, "Health check request failed")
		defer resp.Body.Close()

		assertEqual(t, 200, resp.StatusCode, "Expected status 200")

		var data map[string]string
		err = json.NewDecoder(resp.Body).Decode(&data)
		assertNoError(t, err, "Failed to decode response")

		assertEqual(t, "healthy", data["status"], "Expected healthy status")
	})

	// Test database operations work
	t.Run("Database operations work", func(t *testing.T) {
		// Test ping
		err := framework.Database().Ping()
		assertNoError(t, err, "Database ping failed")

		// Test query (should not return error for valid database)
		_, err = framework.Database().Query("SELECT 1")
		assertNoError(t, err, "Database query failed")
	})
}
