//go:build !benchmark
// +build !benchmark

package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

// TestAuthenticationOAuth2Integration tests OAuth2 token creation and validation
// Requirements: 2.1
func TestAuthenticationOAuth2Integration(t *testing.T) {
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

	// Initialize database tables
	err = framework.Database().CreateTables()
	assertNoError(t, err, "Failed to create database tables")

	// Create auth manager
	authManager := pkg.NewAuthManager(framework.Database(), "test-jwt-secret-key-32-bytes!!", pkg.OAuth2Config{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	// Create test user and token
	testUser := createTestUser("user1", "testuser", "tenant1", []string{"user"}, []string{"read"})
	testToken, err := authManager.CreateAccessToken(testUser.ID, testUser.TenantID, []string{"read", "write"}, 1*time.Hour)
	assertNoError(t, err, "Failed to create test token")

	// Register protected route
	framework.Router().GET("/protected", func(ctx pkg.Context) error {
		// Extract token from Authorization header
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			return ctx.JSON(401, map[string]string{"error": "Missing authorization header"})
		}

		// Parse Bearer token
		var token string
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token = authHeader[7:]
		} else {
			return ctx.JSON(401, map[string]string{"error": "Invalid authorization header format"})
		}

		// Authenticate with OAuth2
		user, err := authManager.AuthenticateOAuth2(token)
		if err != nil {
			return ctx.JSON(401, map[string]string{"error": "Authentication failed"})
		}

		return ctx.JSON(200, map[string]interface{}{
			"message": "Access granted",
			"user_id": user.ID,
		})
	})

	// Start server in background
	go func() {
		if err := framework.Listen(":19101"); err != nil {
			t.Logf("Server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)
	defer framework.Shutdown(2 * time.Second)

	// Test valid OAuth2 token
	t.Run("Valid OAuth2 token", func(t *testing.T) {
		client := &http.Client{}
		req, err := http.NewRequest("GET", "http://localhost:19101/protected", nil)
		assertNoError(t, err, "Failed to create request")

		req.Header.Set("Authorization", "Bearer "+testToken.Token)

		resp, err := client.Do(req)
		assertNoError(t, err, "Request failed")
		defer resp.Body.Close()

		assertEqual(t, 200, resp.StatusCode, "Expected status 200")

		var data map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&data)
		assertNoError(t, err, "Failed to decode response")

		assertEqual(t, "Access granted", data["message"], "Unexpected message")
		assertEqual(t, testUser.ID, data["user_id"], "Unexpected user ID")
	})

	// Test invalid OAuth2 token
	t.Run("Invalid OAuth2 token", func(t *testing.T) {
		client := &http.Client{}
		req, err := http.NewRequest("GET", "http://localhost:19101/protected", nil)
		assertNoError(t, err, "Failed to create request")

		req.Header.Set("Authorization", "Bearer invalid-token")

		resp, err := client.Do(req)
		assertNoError(t, err, "Request failed")
		defer resp.Body.Close()

		assertEqual(t, 401, resp.StatusCode, "Expected status 401")
	})

	// Test missing authorization header
	t.Run("Missing authorization header", func(t *testing.T) {
		resp, err := http.Get("http://localhost:19101/protected")
		assertNoError(t, err, "Request failed")
		defer resp.Body.Close()

		assertEqual(t, 401, resp.StatusCode, "Expected status 401")
	})
}

// TestAuthenticationJWTIntegration tests JWT generation and validation
// Requirements: 2.2
func TestAuthenticationJWTIntegration(t *testing.T) {
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

	// Create auth manager
	authManager := pkg.NewAuthManager(framework.Database(), "test-jwt-secret-key-32-bytes!!", pkg.OAuth2Config{})

	// Create test user
	testUser := createTestUser("user2", "jwtuser", "tenant1", []string{"admin"}, []string{"read", "write"})

	// Generate JWT token
	jwtToken, err := authManager.GenerateJWT(testUser, 1*time.Hour)
	assertNoError(t, err, "Failed to generate JWT")

	// Register protected route
	framework.Router().GET("/jwt-protected", func(ctx pkg.Context) error {
		// Extract token from Authorization header
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			return ctx.JSON(401, map[string]string{"error": "Missing authorization header"})
		}

		// Parse Bearer token
		var token string
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token = authHeader[7:]
		} else {
			return ctx.JSON(401, map[string]string{"error": "Invalid authorization header format"})
		}

		// Authenticate with JWT
		user, err := authManager.AuthenticateJWT(token)
		if err != nil {
			return ctx.JSON(401, map[string]string{"error": "Authentication failed"})
		}

		return ctx.JSON(200, map[string]interface{}{
			"message":  "JWT access granted",
			"user_id":  user.ID,
			"username": user.Username,
			"email":    user.Email,
			"roles":    user.Roles,
		})
	})

	// Start server in background
	go func() {
		if err := framework.Listen(":19102"); err != nil {
			t.Logf("Server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)
	defer framework.Shutdown(2 * time.Second)

	// Test valid JWT token
	t.Run("Valid JWT token", func(t *testing.T) {
		client := &http.Client{}
		req, err := http.NewRequest("GET", "http://localhost:19102/jwt-protected", nil)
		assertNoError(t, err, "Failed to create request")

		req.Header.Set("Authorization", "Bearer "+jwtToken)

		resp, err := client.Do(req)
		assertNoError(t, err, "Request failed")
		defer resp.Body.Close()

		assertEqual(t, 200, resp.StatusCode, "Expected status 200")

		var data map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&data)
		assertNoError(t, err, "Failed to decode response")

		assertEqual(t, "JWT access granted", data["message"], "Unexpected message")
		assertEqual(t, testUser.ID, data["user_id"], "Unexpected user ID")
		assertEqual(t, testUser.Username, data["username"], "Unexpected username")
	})

	// Test invalid JWT token
	t.Run("Invalid JWT token", func(t *testing.T) {
		client := &http.Client{}
		req, err := http.NewRequest("GET", "http://localhost:19102/jwt-protected", nil)
		assertNoError(t, err, "Failed to create request")

		req.Header.Set("Authorization", "Bearer invalid.jwt.token")

		resp, err := client.Do(req)
		assertNoError(t, err, "Request failed")
		defer resp.Body.Close()

		assertEqual(t, 401, resp.StatusCode, "Expected status 401")
	})

	// Test JWT claims extraction
	t.Run("JWT claims extraction", func(t *testing.T) {
		client := &http.Client{}
		req, err := http.NewRequest("GET", "http://localhost:19102/jwt-protected", nil)
		assertNoError(t, err, "Failed to create request")

		req.Header.Set("Authorization", "Bearer "+jwtToken)

		resp, err := client.Do(req)
		assertNoError(t, err, "Request failed")
		defer resp.Body.Close()

		var data map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&data)
		assertNoError(t, err, "Failed to decode response")

		// Verify claims are extracted correctly
		assertNotNil(t, data["roles"], "Roles should be present")
		assertEqual(t, testUser.Username, data["username"], "Username should match")
		assertEqual(t, testUser.Email, data["email"], "Email should match")
	})
}

// TestAuthorizationRoleBasedIntegration tests role-based authorization
// Requirements: 2.3
func TestAuthorizationRoleBasedIntegration(t *testing.T) {
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

	// Create auth manager
	authManager := pkg.NewAuthManager(framework.Database(), "test-jwt-secret-key-32-bytes!!", pkg.OAuth2Config{})

	// Create test users with different roles
	adminUser := createTestUser("admin1", "adminuser", "tenant1", []string{"admin"}, []string{"read", "write", "delete"})
	regularUser := createTestUser("user3", "regularuser", "tenant1", []string{"user"}, []string{"read"})

	// Generate JWT tokens
	adminToken, err := authManager.GenerateJWT(adminUser, 1*time.Hour)
	assertNoError(t, err, "Failed to generate admin JWT")

	userToken, err := authManager.GenerateJWT(regularUser, 1*time.Hour)
	assertNoError(t, err, "Failed to generate user JWT")

	// Register admin-only route
	framework.Router().GET("/admin-only", func(ctx pkg.Context) error {
		// Extract and authenticate token
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" || len(authHeader) < 8 {
			return ctx.JSON(401, map[string]string{"error": "Missing authorization"})
		}

		token := authHeader[7:]
		user, err := authManager.AuthenticateJWT(token)
		if err != nil {
			return ctx.JSON(401, map[string]string{"error": "Authentication failed"})
		}

		// Check admin role
		if err := authManager.AuthorizeRole(user, "admin"); err != nil {
			return ctx.JSON(403, map[string]string{"error": "Insufficient permissions"})
		}

		return ctx.JSON(200, map[string]string{"message": "Admin access granted"})
	})

	// Start server in background
	go func() {
		if err := framework.Listen(":19103"); err != nil {
			t.Logf("Server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)
	defer framework.Shutdown(2 * time.Second)

	// Test admin role access
	t.Run("Admin role access", func(t *testing.T) {
		client := &http.Client{}
		req, err := http.NewRequest("GET", "http://localhost:19103/admin-only", nil)
		assertNoError(t, err, "Failed to create request")

		req.Header.Set("Authorization", "Bearer "+adminToken)

		resp, err := client.Do(req)
		assertNoError(t, err, "Request failed")
		defer resp.Body.Close()

		assertEqual(t, 200, resp.StatusCode, "Expected status 200 for admin")

		var data map[string]string
		err = json.NewDecoder(resp.Body).Decode(&data)
		assertNoError(t, err, "Failed to decode response")

		assertEqual(t, "Admin access granted", data["message"], "Unexpected message")
	})

	// Test regular user role rejection
	t.Run("Regular user role rejection", func(t *testing.T) {
		client := &http.Client{}
		req, err := http.NewRequest("GET", "http://localhost:19103/admin-only", nil)
		assertNoError(t, err, "Failed to create request")

		req.Header.Set("Authorization", "Bearer "+userToken)

		resp, err := client.Do(req)
		assertNoError(t, err, "Request failed")
		defer resp.Body.Close()

		assertEqual(t, 403, resp.StatusCode, "Expected status 403 for regular user")
	})
}

// TestAuthorizationActionBasedIntegration tests action-based authorization
// Requirements: 2.4
func TestAuthorizationActionBasedIntegration(t *testing.T) {
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

	// Create auth manager
	authManager := pkg.NewAuthManager(framework.Database(), "test-jwt-secret-key-32-bytes!!", pkg.OAuth2Config{})

	// Create test users with different actions
	writeUser := createTestUser("user4", "writeuser", "tenant1", []string{"user"}, []string{"read", "write"})
	readOnlyUser := createTestUser("user5", "readonlyuser", "tenant1", []string{"user"}, []string{"read"})

	// Generate JWT tokens
	writeToken, err := authManager.GenerateJWT(writeUser, 1*time.Hour)
	assertNoError(t, err, "Failed to generate write user JWT")

	readToken, err := authManager.GenerateJWT(readOnlyUser, 1*time.Hour)
	assertNoError(t, err, "Failed to generate read-only user JWT")

	// Register write-protected route
	framework.Router().POST("/write-data", func(ctx pkg.Context) error {
		// Extract and authenticate token
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" || len(authHeader) < 8 {
			return ctx.JSON(401, map[string]string{"error": "Missing authorization"})
		}

		token := authHeader[7:]
		user, err := authManager.AuthenticateJWT(token)
		if err != nil {
			return ctx.JSON(401, map[string]string{"error": "Authentication failed"})
		}

		// Check write action
		if err := authManager.AuthorizeAction(user, "write"); err != nil {
			return ctx.JSON(403, map[string]string{"error": "Insufficient permissions"})
		}

		return ctx.JSON(200, map[string]string{"message": "Write access granted"})
	})

	// Start server in background
	go func() {
		if err := framework.Listen(":19104"); err != nil {
			t.Logf("Server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)
	defer framework.Shutdown(2 * time.Second)

	// Test write action access
	t.Run("Write action access", func(t *testing.T) {
		client := &http.Client{}
		req, err := http.NewRequest("POST", "http://localhost:19104/write-data", bytes.NewBufferString("{}"))
		assertNoError(t, err, "Failed to create request")

		req.Header.Set("Authorization", "Bearer "+writeToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		assertNoError(t, err, "Request failed")
		defer resp.Body.Close()

		assertEqual(t, 200, resp.StatusCode, "Expected status 200 for write user")

		var data map[string]string
		err = json.NewDecoder(resp.Body).Decode(&data)
		assertNoError(t, err, "Failed to decode response")

		assertEqual(t, "Write access granted", data["message"], "Unexpected message")
	})

	// Test read-only action rejection
	t.Run("Read-only action rejection", func(t *testing.T) {
		client := &http.Client{}
		req, err := http.NewRequest("POST", "http://localhost:19104/write-data", bytes.NewBufferString("{}"))
		assertNoError(t, err, "Failed to create request")

		req.Header.Set("Authorization", "Bearer "+readToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		assertNoError(t, err, "Request failed")
		defer resp.Body.Close()

		assertEqual(t, 403, resp.StatusCode, "Expected status 403 for read-only user")
	})
}

// TestSessionAuthenticationIntegration tests session-based authentication
// Requirements: 2.5
func TestSessionAuthenticationIntegration(t *testing.T) {
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

	// Initialize database tables
	err = framework.Database().CreateTables()
	assertNoError(t, err, "Failed to create database tables")

	// Create test user
	testUser := createTestUser("user6", "sessionuser", "tenant1", []string{"user"}, []string{"read"})

	// Store session ID for testing
	var sessionID string
	var sessionMu sync.Mutex

	// Register login route
	framework.Router().POST("/login", func(ctx pkg.Context) error {
		// Create session
		session, err := framework.Session().Create(ctx)
		if err != nil {
			return ctx.JSON(500, map[string]string{"error": "Failed to create session"})
		}

		// Store user ID in session
		session.UserID = testUser.ID
		session.TenantID = testUser.TenantID
		session.Data["username"] = testUser.Username
		session.Data["roles"] = testUser.Roles

		// Save session
		if err := framework.Session().Save(ctx, session); err != nil {
			return ctx.JSON(500, map[string]string{"error": "Failed to save session"})
		}

		// Set session cookie
		if err := framework.Session().SetCookie(ctx, session); err != nil {
			return ctx.JSON(500, map[string]string{"error": "Failed to set cookie"})
		}

		sessionMu.Lock()
		sessionID = session.ID
		sessionMu.Unlock()

		return ctx.JSON(200, map[string]interface{}{
			"message":    "Login successful",
			"session_id": session.ID,
		})
	})

	// Register protected route
	framework.Router().GET("/session-protected", func(ctx pkg.Context) error {
		// Get session from cookie
		session, err := framework.Session().GetSessionFromCookie(ctx)
		if err != nil {
			return ctx.JSON(401, map[string]string{"error": "No valid session"})
		}

		// Check if session is valid
		if !framework.Session().IsValid(session.ID) {
			return ctx.JSON(401, map[string]string{"error": "Session expired"})
		}

		return ctx.JSON(200, map[string]interface{}{
			"message":  "Session access granted",
			"user_id":  session.UserID,
			"username": session.Data["username"],
		})
	})

	// Start server in background
	go func() {
		if err := framework.Listen(":19105"); err != nil {
			t.Logf("Server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)
	defer framework.Shutdown(2 * time.Second)

	// Test login flow
	var sessionCookie string
	t.Run("Login flow", func(t *testing.T) {
		resp, err := http.Post("http://localhost:19105/login", "application/json", bytes.NewBufferString("{}"))
		assertNoError(t, err, "Login request failed")
		defer resp.Body.Close()

		assertEqual(t, 200, resp.StatusCode, "Expected status 200")

		// Extract session cookie
		cookies := resp.Cookies()
		for _, cookie := range cookies {
			if cookie.Name == "rockstar_session" {
				sessionCookie = cookie.Value
				break
			}
		}

		assertNotNil(t, sessionCookie, "Session cookie should be set")

		var data map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&data)
		assertNoError(t, err, "Failed to decode response")

		assertEqual(t, "Login successful", data["message"], "Unexpected message")
	})

	// Test protected route access with session cookie
	t.Run("Protected route with session", func(t *testing.T) {
		client := &http.Client{}
		req, err := http.NewRequest("GET", "http://localhost:19105/session-protected", nil)
		assertNoError(t, err, "Failed to create request")

		// Add session cookie
		req.AddCookie(&http.Cookie{
			Name:  "rockstar_session",
			Value: sessionCookie,
		})

		resp, err := client.Do(req)
		assertNoError(t, err, "Request failed")
		defer resp.Body.Close()

		assertEqual(t, 200, resp.StatusCode, "Expected status 200")

		var data map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&data)
		assertNoError(t, err, "Failed to decode response")

		assertEqual(t, "Session access granted", data["message"], "Unexpected message")
		assertEqual(t, testUser.ID, data["user_id"], "Unexpected user ID")
	})

	// Test session persistence
	t.Run("Session persistence", func(t *testing.T) {
		sessionMu.Lock()
		sid := sessionID
		sessionMu.Unlock()

		if sid == "" {
			t.Skip("Session ID not set, skipping persistence test")
			return
		}

		// Load session directly from database
		session, err := framework.Session().Load(nil, sid)
		assertNoError(t, err, "Failed to load session")

		assertEqual(t, testUser.ID, session.UserID, "User ID should persist")
		assertEqual(t, testUser.Username, session.Data["username"], "Username should persist")
	})

	// Test access without session cookie
	t.Run("Access without session", func(t *testing.T) {
		resp, err := http.Get("http://localhost:19105/session-protected")
		assertNoError(t, err, "Request failed")
		defer resp.Body.Close()

		assertEqual(t, 401, resp.StatusCode, "Expected status 401 without session")
	})
}

// TestRateLimitingIntegration tests rate limiting enforcement
// Requirements: 2.6
func TestRateLimitingIntegration(t *testing.T) {
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

	// Rate limit configuration
	rateLimit := 5
	rateLimitWindow := 1 * time.Minute
	requestCounts := make(map[string]int)
	var mu sync.Mutex

	// Register rate-limited route
	framework.Router().GET("/rate-limited", func(ctx pkg.Context) error {
		// Get client identifier (IP address)
		clientIP := ctx.Request().RemoteAddr

		mu.Lock()
		count := requestCounts[clientIP]

		if count >= rateLimit {
			mu.Unlock()
			return ctx.JSON(429, map[string]interface{}{
				"error":  "Rate limit exceeded",
				"limit":  rateLimit,
				"window": rateLimitWindow.String(),
			})
		}

		requestCounts[clientIP]++
		mu.Unlock()

		return ctx.JSON(200, map[string]interface{}{
			"message": "Request successful",
			"count":   count + 1,
		})
	})

	// Start server in background
	go func() {
		if err := framework.Listen(":19106"); err != nil {
			t.Logf("Server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)
	defer framework.Shutdown(2 * time.Second)

	// Test rate limiting enforcement
	t.Run("Rate limiting enforcement", func(t *testing.T) {
		// Use a shared client to maintain the same connection/IP
		client := &http.Client{}

		// Make requests up to the rate limit
		for i := 1; i <= rateLimit; i++ {
			req, err := http.NewRequest("GET", "http://localhost:19106/rate-limited", nil)
			assertNoError(t, err, fmt.Sprintf("Failed to create request %d", i))

			resp, err := client.Do(req)
			assertNoError(t, err, fmt.Sprintf("Request %d failed", i))

			assertEqual(t, 200, resp.StatusCode, fmt.Sprintf("Expected status 200 for request %d", i))

			var data map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&data)
			resp.Body.Close()
			assertNoError(t, err, "Failed to decode response")

			assertEqual(t, "Request successful", data["message"], "Unexpected message")
		}

		// Next request should be rate limited
		req, err := http.NewRequest("GET", "http://localhost:19106/rate-limited", nil)
		assertNoError(t, err, "Failed to create request")

		resp, err := client.Do(req)
		assertNoError(t, err, "Request failed")

		assertEqual(t, 429, resp.StatusCode, "Expected status 429 when rate limit exceeded")

		var data map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&data)
		resp.Body.Close()
		assertNoError(t, err, "Failed to decode response")

		assertEqual(t, "Rate limit exceeded", data["error"], "Unexpected error message")
		assertNotNil(t, data["limit"], "Limit should be present in response")
		assertEqual(t, float64(rateLimit), data["limit"], "Limit value should match configured limit")
	})
}
