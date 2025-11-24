package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

// TestAuthenticationOAuth2Integration tests OAuth2 authentication end-to-end
func TestAuthenticationOAuth2Integration(t *testing.T) {
	// Setup database - using the mock from auth_test.go
	db := newTestMockDB()
	db.Connect(pkg.DatabaseConfig{Driver: "mock"})

	// Setup auth manager
	authManager := pkg.NewAuthManager(db, "test-secret-key", pkg.OAuth2Config{})

	// Create access token
	token, err := authManager.CreateAccessToken(
		"user123",
		"tenant456",
		[]string{"read", "write"},
		1*time.Hour,
	)
	if err != nil {
		t.Fatalf("Failed to create access token: %v", err)
	}

	// Setup server
	config := pkg.ServerConfig{
		EnableHTTP1:  true,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	server := pkg.NewServer(config)
	router := pkg.NewRouter()

	// Protected route
	router.GET("/api/protected", func(ctx pkg.Context) error {
		// Extract token from header
		authHeader := ctx.Headers()["Authorization"]
		if authHeader == "" {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "missing token"})
		}

		// Authenticate
		user, err := authManager.AuthenticateOAuth2(authHeader)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		}

		return ctx.JSON(http.StatusOK, map[string]interface{}{
			"message": "Access granted",
			"user_id": user.ID,
			"tenant":  user.TenantID,
		})
	})

	server.SetRouter(router)

	addr := "127.0.0.1:19101"
	if err := server.Listen(addr); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Close()

	time.Sleep(100 * time.Millisecond)

	// Test with valid token
	req, _ := http.NewRequest("GET", "http://"+addr+"/api/protected", nil)
	req.Header.Set("Authorization", token.Token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if result["user_id"] != "user123" {
		t.Errorf("Expected user_id 'user123', got '%v'", result["user_id"])
	}

	// Test without token
	req, _ = http.NewRequest("GET", "http://"+addr+"/api/protected", nil)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}

	// Test with invalid token
	req, _ = http.NewRequest("GET", "http://"+addr+"/api/protected", nil)
	req.Header.Set("Authorization", "invalid-token")
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}
}

// TestAuthenticationJWTIntegration tests JWT authentication end-to-end
func TestAuthenticationJWTIntegration(t *testing.T) {
	// Setup
	db := newTestMockDB()
	db.Connect(pkg.DatabaseConfig{Driver: "mock"})
	authManager := pkg.NewAuthManager(db, "test-secret-key", pkg.OAuth2Config{})

	// Create user and generate JWT
	user := &pkg.User{
		ID:       "user123",
		Username: "testuser",
		Email:    "test@example.com",
		Roles:    []string{"admin"},
		TenantID: "tenant456",
	}

	jwtToken, err := authManager.GenerateJWT(user, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate JWT: %v", err)
	}

	// Setup server
	config := pkg.ServerConfig{
		EnableHTTP1:  true,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	server := pkg.NewServer(config)
	router := pkg.NewRouter()

	// JWT protected route
	router.GET("/api/profile", func(ctx pkg.Context) error {
		authHeader := ctx.Headers()["Authorization"]
		if authHeader == "" {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "missing token"})
		}

		user, err := authManager.AuthenticateJWT(authHeader)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		}

		return ctx.JSON(http.StatusOK, map[string]interface{}{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		})
	})

	server.SetRouter(router)

	addr := "127.0.0.1:19102"
	if err := server.Listen(addr); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Close()

	time.Sleep(100 * time.Millisecond)

	// Test with valid JWT
	req, _ := http.NewRequest("GET", "http://"+addr+"/api/profile", nil)
	req.Header.Set("Authorization", jwtToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if result["username"] != "testuser" {
		t.Errorf("Expected username 'testuser', got '%v'", result["username"])
	}
}

// TestAuthorizationRoleBasedIntegration tests role-based authorization end-to-end
func TestAuthorizationRoleBasedIntegration(t *testing.T) {
	// Setup
	db := newTestMockDB()
	db.Connect(pkg.DatabaseConfig{Driver: "mock"})
	authManager := pkg.NewAuthManager(db, "test-secret-key", pkg.OAuth2Config{})

	// Create users with different roles
	adminUser := &pkg.User{
		ID:       "admin123",
		Username: "admin",
		Roles:    []string{"admin"},
		TenantID: "tenant456",
	}

	regularUser := &pkg.User{
		ID:       "user123",
		Username: "user",
		Roles:    []string{"user"},
		TenantID: "tenant456",
	}

	adminToken, _ := authManager.GenerateJWT(adminUser, 1*time.Hour)
	userToken, _ := authManager.GenerateJWT(regularUser, 1*time.Hour)

	// Setup server
	config := pkg.ServerConfig{
		EnableHTTP1:  true,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	server := pkg.NewServer(config)
	router := pkg.NewRouter()

	// Admin-only route
	router.GET("/api/admin/users", func(ctx pkg.Context) error {
		authHeader := ctx.Headers()["Authorization"]
		if authHeader == "" {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "missing token"})
		}

		user, err := authManager.AuthenticateJWT(authHeader)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		}

		// Check authorization
		if err := authManager.AuthorizeRole(user, "admin"); err != nil {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "insufficient permissions"})
		}

		return ctx.JSON(http.StatusOK, []map[string]string{
			{"id": "1", "name": "Alice"},
			{"id": "2", "name": "Bob"},
		})
	})

	server.SetRouter(router)

	addr := "127.0.0.1:19103"
	if err := server.Listen(addr); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Close()

	time.Sleep(100 * time.Millisecond)

	client := &http.Client{}

	// Test with admin token (should succeed)
	req, _ := http.NewRequest("GET", "http://"+addr+"/api/admin/users", nil)
	req.Header.Set("Authorization", adminToken)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for admin, got %d", resp.StatusCode)
	}

	// Test with regular user token (should fail)
	req, _ = http.NewRequest("GET", "http://"+addr+"/api/admin/users", nil)
	req.Header.Set("Authorization", userToken)

	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("Expected status 403 for regular user, got %d", resp.StatusCode)
	}
}

// TestAuthorizationActionBasedIntegration tests action-based authorization end-to-end
func TestAuthorizationActionBasedIntegration(t *testing.T) {
	// Setup
	db := newTestMockDB()
	db.Connect(pkg.DatabaseConfig{Driver: "mock"})
	authManager := pkg.NewAuthManager(db, "test-secret-key", pkg.OAuth2Config{})

	// Create users with different actions
	writeUser := &pkg.User{
		ID:       "writer123",
		Username: "writer",
		Actions:  []string{"read", "write"},
		TenantID: "tenant456",
	}

	readOnlyUser := &pkg.User{
		ID:       "reader123",
		Username: "reader",
		Actions:  []string{"read"},
		TenantID: "tenant456",
	}

	writeToken, _ := authManager.GenerateJWT(writeUser, 1*time.Hour)
	readToken, _ := authManager.GenerateJWT(readOnlyUser, 1*time.Hour)

	// Setup server
	config := pkg.ServerConfig{
		EnableHTTP1:  true,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	server := pkg.NewServer(config)
	router := pkg.NewRouter()

	// Write-protected route
	router.POST("/api/posts", func(ctx pkg.Context) error {
		authHeader := ctx.Headers()["Authorization"]
		if authHeader == "" {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "missing token"})
		}

		user, err := authManager.AuthenticateJWT(authHeader)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		}

		// Check write permission
		if err := authManager.AuthorizeAction(user, "write"); err != nil {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "write permission required"})
		}

		var post map[string]interface{}
		json.Unmarshal(ctx.Body(), &post)
		post["id"] = "1"

		return ctx.JSON(http.StatusCreated, post)
	})

	server.SetRouter(router)

	addr := "127.0.0.1:19104"
	if err := server.Listen(addr); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Close()

	time.Sleep(100 * time.Millisecond)

	client := &http.Client{}
	postData := map[string]string{"title": "New Post"}
	jsonData, _ := json.Marshal(postData)

	// Test with write permission (should succeed)
	req, _ := http.NewRequest("POST", "http://"+addr+"/api/posts", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", writeToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201 for write user, got %d", resp.StatusCode)
	}

	// Test with read-only permission (should fail)
	req, _ = http.NewRequest("POST", "http://"+addr+"/api/posts", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", readToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("Expected status 403 for read-only user, got %d", resp.StatusCode)
	}
}

// TestSessionAuthenticationIntegration tests session-based authentication
func TestSessionAuthenticationIntegration(t *testing.T) {
	// Setup
	db := newTestMockDB()
	db.Connect(pkg.DatabaseConfig{Driver: "mock"})

	sessionConfig := pkg.DefaultSessionConfig()
	sessionConfig.EncryptionKey = make([]byte, 32)
	sessionConfig.StorageType = pkg.SessionStorageDatabase

	sessionManager, err := pkg.NewSessionManager(sessionConfig, db, nil)
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}

	// Setup server
	config := pkg.ServerConfig{
		EnableHTTP1:  true,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	server := pkg.NewServer(config)
	router := pkg.NewRouter()

	// Login route
	router.POST("/api/login", func(ctx pkg.Context) error {
		var credentials map[string]string
		json.Unmarshal(ctx.Body(), &credentials)

		// Simulate authentication
		if credentials["username"] == "testuser" && credentials["password"] == "password" {
			// Create session
			session, err := sessionManager.Create(ctx)
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "session creation failed"})
			}

			session.UserID = "user123"
			session.Data["username"] = "testuser"
			sessionManager.Save(ctx, session)

			// Set cookie
			sessionManager.SetCookie(ctx, session)

			return ctx.JSON(http.StatusOK, map[string]string{
				"message":    "login successful",
				"session_id": session.ID,
			})
		}

		return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
	})

	// Protected route
	router.GET("/api/dashboard", func(ctx pkg.Context) error {
		session, err := sessionManager.GetSessionFromCookie(ctx)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "not authenticated"})
		}

		if session.UserID == "" {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "not authenticated"})
		}

		return ctx.JSON(http.StatusOK, map[string]interface{}{
			"message":  "welcome to dashboard",
			"username": session.Data["username"],
		})
	})

	server.SetRouter(router)

	addr := "127.0.0.1:19105"
	if err := server.Listen(addr); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Close()

	time.Sleep(100 * time.Millisecond)

	// Test login
	credentials := map[string]string{
		"username": "testuser",
		"password": "password",
	}
	jsonData, _ := json.Marshal(credentials)

	resp, err := http.Post("http://"+addr+"/api/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Login request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Extract session cookie
	cookies := resp.Cookies()
	if len(cookies) == 0 {
		t.Fatal("Expected session cookie")
	}

	// Test protected route with session
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "http://"+addr+"/api/dashboard", nil)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Dashboard request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 with valid session, got %d", resp.StatusCode)
	}
}

// TestRateLimitingIntegration tests API rate limiting
func TestRateLimitingIntegration(t *testing.T) {
	// Setup
	db := newTestMockDB()
	db.Connect(pkg.DatabaseConfig{Driver: "mock"})

	securityConfig := pkg.SecurityConfig{
		EncryptionKey: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
	}
	securityManager, err := pkg.NewSecurityManager(db, securityConfig)
	if err != nil {
		t.Fatalf("Failed to create security manager: %v", err)
	}

	// Setup server
	config := pkg.ServerConfig{
		EnableHTTP1:  true,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	server := pkg.NewServer(config)
	router := pkg.NewRouter()

	// Rate-limited route
	router.GET("/api/limited", func(ctx pkg.Context) error {
		// Check rate limit
		if err := securityManager.CheckRateLimit(ctx, "api_endpoint"); err != nil {
			return ctx.JSON(http.StatusTooManyRequests, map[string]string{"error": "rate limit exceeded"})
		}

		return ctx.JSON(http.StatusOK, map[string]string{"message": "success"})
	})

	server.SetRouter(router)

	addr := "127.0.0.1:19106"
	if err := server.Listen(addr); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Close()

	time.Sleep(100 * time.Millisecond)

	// Make requests up to limit
	for i := 0; i < 5; i++ {
		resp, err := http.Get("http://" + addr + "/api/limited")
		if err != nil {
			t.Fatalf("Request %d failed: %v", i+1, err)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Request %d: expected status 200, got %d", i+1, resp.StatusCode)
		}
	}

	// Next request should be rate limited
	resp, err := http.Get("http://" + addr + "/api/limited")
	if err != nil {
		t.Fatalf("Rate limit test request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("Expected status 429 for rate limited request, got %d", resp.StatusCode)
	}
}
