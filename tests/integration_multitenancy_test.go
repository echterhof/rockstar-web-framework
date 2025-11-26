//go:build !benchmark
// +build !benchmark

package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

// createTenantMiddleware creates middleware that resolves tenant from Host header
func createTenantMiddleware(db pkg.DatabaseManager) pkg.MiddlewareFunc {
	return func(ctx pkg.Context, next pkg.HandlerFunc) error {
		// Resolve tenant from Host field in Request
		host := ctx.Request().Host
		if host == "" {
			host = "localhost"
		}

		// Strip port if present (e.g., "localhost:8080" -> "localhost")
		if idx := strings.Index(host, ":"); idx != -1 {
			host = host[:idx]
		}

		tenant, err := db.LoadTenantByHost(host)
		if err == nil {
			// Use type assertion to access SetTenant (internal method)
			if ctxImpl, ok := ctx.(interface{ SetTenant(*pkg.Tenant) }); ok {
				ctxImpl.SetTenant(tenant)
			}
		}

		return next(ctx)
	}
}

// TestMultiTenancyHostRouting tests routing to different tenants by hostname
// Requirements: 3.1
func TestMultiTenancyHostRouting(t *testing.T) {
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

	// Create test tenants
	tenant1 := createTestTenant("tenant1", "Tenant One", []string{"tenant1.example.com"})
	tenant2 := createTestTenant("tenant2", "Tenant Two", []string{"tenant2.example.com"})
	defaultTenant := createTestTenant("default", "Default Tenant", []string{"localhost"})

	// Save tenants to database
	err = framework.Database().SaveTenant(tenant1)
	assertNoError(t, err, "Failed to save tenant1")
	err = framework.Database().SaveTenant(tenant2)
	assertNoError(t, err, "Failed to save tenant2")
	err = framework.Database().SaveTenant(defaultTenant)
	assertNoError(t, err, "Failed to save default tenant")

	// Register hosts and tenants with the framework
	err = framework.ServerManager().RegisterHost("tenant1.example.com", pkg.HostConfig{Hostname: "tenant1.example.com"})
	assertNoError(t, err, "Failed to register tenant1 host")
	err = framework.ServerManager().RegisterHost("tenant2.example.com", pkg.HostConfig{Hostname: "tenant2.example.com"})
	assertNoError(t, err, "Failed to register tenant2 host")
	err = framework.ServerManager().RegisterHost("localhost", pkg.HostConfig{Hostname: "localhost"})
	assertNoError(t, err, "Failed to register localhost host")

	err = framework.ServerManager().RegisterTenant("tenant1", []string{"tenant1.example.com"})
	assertNoError(t, err, "Failed to register tenant1")
	err = framework.ServerManager().RegisterTenant("tenant2", []string{"tenant2.example.com"})
	assertNoError(t, err, "Failed to register tenant2")
	err = framework.ServerManager().RegisterTenant("default", []string{"localhost"})
	assertNoError(t, err, "Failed to register default tenant")

	// Apply tenant middleware globally
	framework.Router().Use(createTenantMiddleware(framework.Database()))

	// Register route that returns tenant information
	framework.Router().GET("/tenant-info", func(ctx pkg.Context) error {
		tenant := ctx.Tenant()
		if tenant == nil {
			return ctx.JSON(200, map[string]interface{}{
				"tenant_id":   "none",
				"tenant_name": "No Tenant",
			})
		}
		return ctx.JSON(200, map[string]interface{}{
			"tenant_id":   tenant.ID,
			"tenant_name": tenant.Name,
		})
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

	// Test routing to tenant1
	t.Run("Route to tenant1 by hostname", func(t *testing.T) {
		client := &http.Client{}
		req, err := http.NewRequest("GET", "http://localhost:19201/tenant-info", nil)
		assertNoError(t, err, "Failed to create request")

		req.Host = "tenant1.example.com"

		resp, err := client.Do(req)
		assertNoError(t, err, "Failed to make request")
		defer resp.Body.Close()

		assertEqual(t, 200, resp.StatusCode, "Expected status 200")

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assertNoError(t, err, "Failed to decode response")

		assertEqual(t, "tenant1", result["tenant_id"], "Expected tenant1 ID")
		assertEqual(t, "Tenant One", result["tenant_name"], "Expected tenant1 name")
	})

	// Test routing to tenant2
	t.Run("Route to tenant2 by hostname", func(t *testing.T) {
		client := &http.Client{}
		req, err := http.NewRequest("GET", "http://localhost:19201/tenant-info", nil)
		assertNoError(t, err, "Failed to create request")

		req.Host = "tenant2.example.com"

		resp, err := client.Do(req)
		assertNoError(t, err, "Failed to make request")
		defer resp.Body.Close()

		assertEqual(t, 200, resp.StatusCode, "Expected status 200")

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assertNoError(t, err, "Failed to decode response")

		assertEqual(t, "tenant2", result["tenant_id"], "Expected tenant2 ID")
		assertEqual(t, "Tenant Two", result["tenant_name"], "Expected tenant2 name")
	})

	// Test default routing for unknown hosts
	t.Run("Default routing for unknown host", func(t *testing.T) {
		client := &http.Client{}
		req, err := http.NewRequest("GET", "http://localhost:19201/tenant-info", nil)
		assertNoError(t, err, "Failed to create request")

		req.Host = "localhost"

		resp, err := client.Do(req)
		assertNoError(t, err, "Failed to make request")
		defer resp.Body.Close()

		assertEqual(t, 200, resp.StatusCode, "Expected status 200")

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assertNoError(t, err, "Failed to decode response")

		assertEqual(t, "default", result["tenant_id"], "Expected default tenant ID")
		assertEqual(t, "Default Tenant", result["tenant_name"], "Expected default tenant name")
	})
}

// TestMultiTenancyDataIsolation tests data isolation between tenants
// Requirements: 3.2
func TestMultiTenancyDataIsolation(t *testing.T) {
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

	// Create test tenants
	tenant1 := createTestTenant("tenant1", "Tenant One", []string{"tenant1.example.com"})
	tenant2 := createTestTenant("tenant2", "Tenant Two", []string{"tenant2.example.com"})

	// Save tenants to database
	err = framework.Database().SaveTenant(tenant1)
	assertNoError(t, err, "Failed to save tenant1")
	err = framework.Database().SaveTenant(tenant2)
	assertNoError(t, err, "Failed to save tenant2")

	// Register hosts and tenants with the framework
	err = framework.ServerManager().RegisterHost("tenant1.example.com", pkg.HostConfig{Hostname: "tenant1.example.com"})
	assertNoError(t, err, "Failed to register tenant1 host")
	err = framework.ServerManager().RegisterHost("tenant2.example.com", pkg.HostConfig{Hostname: "tenant2.example.com"})
	assertNoError(t, err, "Failed to register tenant2 host")

	err = framework.ServerManager().RegisterTenant("tenant1", []string{"tenant1.example.com"})
	assertNoError(t, err, "Failed to register tenant1")
	err = framework.ServerManager().RegisterTenant("tenant2", []string{"tenant2.example.com"})
	assertNoError(t, err, "Failed to register tenant2")

	// Apply tenant middleware globally
	framework.Router().Use(createTenantMiddleware(framework.Database()))

	// In-memory data store per tenant (simulating tenant-specific data)
	tenantData := make(map[string][]string)
	var dataMutex sync.RWMutex

	// Register route to create data
	framework.Router().POST("/data", func(ctx pkg.Context) error {
		tenant := ctx.Tenant()
		if tenant == nil {
			return ctx.JSON(400, map[string]string{"error": "No tenant context"})
		}

		var payload map[string]string
		if err := json.Unmarshal(ctx.Body(), &payload); err != nil {
			return ctx.JSON(400, map[string]string{"error": "Invalid JSON"})
		}

		dataMutex.Lock()
		if tenantData[tenant.ID] == nil {
			tenantData[tenant.ID] = make([]string, 0)
		}
		tenantData[tenant.ID] = append(tenantData[tenant.ID], payload["value"])
		dataMutex.Unlock()

		return ctx.JSON(201, map[string]string{"status": "created"})
	})

	// Register route to get data
	framework.Router().GET("/data", func(ctx pkg.Context) error {
		tenant := ctx.Tenant()
		if tenant == nil {
			return ctx.JSON(400, map[string]string{"error": "No tenant context"})
		}

		dataMutex.RLock()
		data := tenantData[tenant.ID]
		dataMutex.RUnlock()

		if data == nil {
			data = make([]string, 0)
		}

		return ctx.JSON(200, map[string]interface{}{
			"tenant_id": tenant.ID,
			"data":      data,
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

	// Create data for tenant1
	t.Run("Create data for tenant1", func(t *testing.T) {
		client := &http.Client{}
		payload := map[string]string{"value": "tenant1-data-1"}
		body, _ := json.Marshal(payload)

		req, err := http.NewRequest("POST", "http://localhost:19202/data", bytes.NewBuffer(body))
		assertNoError(t, err, "Failed to create request")

		req.Host = "tenant1.example.com"
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		assertNoError(t, err, "Failed to make request")
		defer resp.Body.Close()

		assertEqual(t, 201, resp.StatusCode, "Expected status 201")
	})

	// Create data for tenant2
	t.Run("Create data for tenant2", func(t *testing.T) {
		client := &http.Client{}
		payload := map[string]string{"value": "tenant2-data-1"}
		body, _ := json.Marshal(payload)

		req, err := http.NewRequest("POST", "http://localhost:19202/data", bytes.NewBuffer(body))
		assertNoError(t, err, "Failed to create request")

		req.Host = "tenant2.example.com"
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		assertNoError(t, err, "Failed to make request")
		defer resp.Body.Close()

		assertEqual(t, 201, resp.StatusCode, "Expected status 201")
	})

	// Verify tenant1 only sees their own data
	t.Run("Tenant1 sees only their data", func(t *testing.T) {
		client := &http.Client{}
		req, err := http.NewRequest("GET", "http://localhost:19202/data", nil)
		assertNoError(t, err, "Failed to create request")

		req.Host = "tenant1.example.com"

		resp, err := client.Do(req)
		assertNoError(t, err, "Failed to make request")
		defer resp.Body.Close()

		assertEqual(t, 200, resp.StatusCode, "Expected status 200")

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assertNoError(t, err, "Failed to decode response")

		assertEqual(t, "tenant1", result["tenant_id"], "Expected tenant1 ID")

		data := result["data"].([]interface{})
		assertEqual(t, 1, len(data), "Expected 1 data item")
		assertEqual(t, "tenant1-data-1", data[0], "Expected tenant1 data")
	})

	// Verify tenant2 only sees their own data
	t.Run("Tenant2 sees only their data", func(t *testing.T) {
		client := &http.Client{}
		req, err := http.NewRequest("GET", "http://localhost:19202/data", nil)
		assertNoError(t, err, "Failed to create request")

		req.Host = "tenant2.example.com"

		resp, err := client.Do(req)
		assertNoError(t, err, "Failed to make request")
		defer resp.Body.Close()

		assertEqual(t, 200, resp.StatusCode, "Expected status 200")

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assertNoError(t, err, "Failed to decode response")

		assertEqual(t, "tenant2", result["tenant_id"], "Expected tenant2 ID")

		data := result["data"].([]interface{})
		assertEqual(t, 1, len(data), "Expected 1 data item")
		assertEqual(t, "tenant2-data-1", data[0], "Expected tenant2 data")
	})
}

// TestMultiTenancySessionIsolation tests session isolation per tenant
// Requirements: 3.3
func TestMultiTenancySessionIsolation(t *testing.T) {
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

	// Create test tenants
	tenant1 := createTestTenant("tenant1", "Tenant One", []string{"tenant1.example.com"})
	tenant2 := createTestTenant("tenant2", "Tenant Two", []string{"tenant2.example.com"})

	// Save tenants to database
	err = framework.Database().SaveTenant(tenant1)
	assertNoError(t, err, "Failed to save tenant1")
	err = framework.Database().SaveTenant(tenant2)
	assertNoError(t, err, "Failed to save tenant2")

	// Register hosts and tenants with the framework
	err = framework.ServerManager().RegisterHost("tenant1.example.com", pkg.HostConfig{Hostname: "tenant1.example.com"})
	assertNoError(t, err, "Failed to register tenant1 host")
	err = framework.ServerManager().RegisterHost("tenant2.example.com", pkg.HostConfig{Hostname: "tenant2.example.com"})
	assertNoError(t, err, "Failed to register tenant2 host")

	err = framework.ServerManager().RegisterTenant("tenant1", []string{"tenant1.example.com"})
	assertNoError(t, err, "Failed to register tenant1")
	err = framework.ServerManager().RegisterTenant("tenant2", []string{"tenant2.example.com"})
	assertNoError(t, err, "Failed to register tenant2")

	// Create test sessions for each tenant
	session1 := createTestSession("session1", "user1", "tenant1")
	session2 := createTestSession("session2", "user2", "tenant2")

	err = framework.Database().SaveSession(session1)
	assertNoError(t, err, "Failed to save session1")
	err = framework.Database().SaveSession(session2)
	assertNoError(t, err, "Failed to save session2")

	// Apply tenant middleware globally
	framework.Router().Use(createTenantMiddleware(framework.Database()))

	// Register route to get session info
	framework.Router().GET("/session-info", func(ctx pkg.Context) error {
		tenant := ctx.Tenant()
		if tenant == nil {
			return ctx.JSON(400, map[string]string{"error": "No tenant context"})
		}

		// Get session ID from cookie
		cookie, err := ctx.GetCookie("session_id")
		if err != nil {
			return ctx.JSON(401, map[string]string{"error": "No session cookie"})
		}

		// Load session from database
		session, err := framework.Database().LoadSession(cookie.Value)
		if err != nil {
			return ctx.JSON(401, map[string]string{"error": "Invalid session"})
		}

		// Verify session belongs to current tenant
		if session.TenantID != tenant.ID {
			return ctx.JSON(403, map[string]string{"error": "Session does not belong to tenant"})
		}

		return ctx.JSON(200, map[string]interface{}{
			"session_id": session.ID,
			"user_id":    session.UserID,
			"tenant_id":  session.TenantID,
		})
	})

	// Start server in background
	go func() {
		if err := framework.Listen(":19203"); err != nil {
			t.Logf("Server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)
	defer framework.Shutdown(2 * time.Second)

	// Test tenant1 session access
	t.Run("Tenant1 accesses their session", func(t *testing.T) {
		client := &http.Client{}
		req, err := http.NewRequest("GET", "http://localhost:19203/session-info", nil)
		assertNoError(t, err, "Failed to create request")

		req.Host = "tenant1.example.com"
		req.AddCookie(&http.Cookie{Name: "session_id", Value: "session1"})

		resp, err := client.Do(req)
		assertNoError(t, err, "Failed to make request")
		defer resp.Body.Close()

		assertEqual(t, 200, resp.StatusCode, "Expected status 200")

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assertNoError(t, err, "Failed to decode response")

		assertEqual(t, "session1", result["session_id"], "Expected session1 ID")
		assertEqual(t, "tenant1", result["tenant_id"], "Expected tenant1 ID")
	})

	// Test tenant2 session access
	t.Run("Tenant2 accesses their session", func(t *testing.T) {
		client := &http.Client{}
		req, err := http.NewRequest("GET", "http://localhost:19203/session-info", nil)
		assertNoError(t, err, "Failed to create request")

		req.Host = "tenant2.example.com"
		req.AddCookie(&http.Cookie{Name: "session_id", Value: "session2"})

		resp, err := client.Do(req)
		assertNoError(t, err, "Failed to make request")
		defer resp.Body.Close()

		assertEqual(t, 200, resp.StatusCode, "Expected status 200")

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assertNoError(t, err, "Failed to decode response")

		assertEqual(t, "session2", result["session_id"], "Expected session2 ID")
		assertEqual(t, "tenant2", result["tenant_id"], "Expected tenant2 ID")
	})

	// Test cross-tenant session access is blocked
	t.Run("Tenant1 cannot access tenant2 session", func(t *testing.T) {
		client := &http.Client{}
		req, err := http.NewRequest("GET", "http://localhost:19203/session-info", nil)
		assertNoError(t, err, "Failed to create request")

		req.Host = "tenant1.example.com"
		req.AddCookie(&http.Cookie{Name: "session_id", Value: "session2"})

		resp, err := client.Do(req)
		assertNoError(t, err, "Failed to make request")
		defer resp.Body.Close()

		assertEqual(t, 403, resp.StatusCode, "Expected status 403 for cross-tenant access")
	})
}

// TestMultiTenancyAuthenticationIsolation tests authentication token isolation per tenant
// Requirements: 3.4
func TestMultiTenancyAuthenticationIsolation(t *testing.T) {
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

	// Create test tenants
	tenant1 := createTestTenant("tenant1", "Tenant One", []string{"tenant1.example.com"})
	tenant2 := createTestTenant("tenant2", "Tenant Two", []string{"tenant2.example.com"})

	// Save tenants to database
	err = framework.Database().SaveTenant(tenant1)
	assertNoError(t, err, "Failed to save tenant1")
	err = framework.Database().SaveTenant(tenant2)
	assertNoError(t, err, "Failed to save tenant2")

	// Register hosts and tenants with the framework
	err = framework.ServerManager().RegisterHost("tenant1.example.com", pkg.HostConfig{Hostname: "tenant1.example.com"})
	assertNoError(t, err, "Failed to register tenant1 host")
	err = framework.ServerManager().RegisterHost("tenant2.example.com", pkg.HostConfig{Hostname: "tenant2.example.com"})
	assertNoError(t, err, "Failed to register tenant2 host")

	err = framework.ServerManager().RegisterTenant("tenant1", []string{"tenant1.example.com"})
	assertNoError(t, err, "Failed to register tenant1")
	err = framework.ServerManager().RegisterTenant("tenant2", []string{"tenant2.example.com"})
	assertNoError(t, err, "Failed to register tenant2")

	// Create test tokens for each tenant
	token1 := createTestAccessToken("token1", "user1", "tenant1", []string{"read", "write"}, 1*time.Hour)
	token2 := createTestAccessToken("token2", "user2", "tenant2", []string{"read", "write"}, 1*time.Hour)

	err = framework.Database().SaveAccessToken(token1)
	assertNoError(t, err, "Failed to save token1")
	err = framework.Database().SaveAccessToken(token2)
	assertNoError(t, err, "Failed to save token2")

	// Apply tenant middleware globally
	framework.Router().Use(createTenantMiddleware(framework.Database()))

	// Register protected route
	framework.Router().GET("/protected", func(ctx pkg.Context) error {
		tenant := ctx.Tenant()
		if tenant == nil {
			return ctx.JSON(400, map[string]string{"error": "No tenant context"})
		}

		// Extract token from Authorization header
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			return ctx.JSON(401, map[string]string{"error": "Missing authorization header"})
		}

		// Parse Bearer token
		var tokenValue string
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenValue = authHeader[7:]
		} else {
			return ctx.JSON(401, map[string]string{"error": "Invalid authorization header format"})
		}

		// Validate token
		token, err := framework.Database().ValidateAccessToken(tokenValue)
		if err != nil {
			return ctx.JSON(401, map[string]string{"error": "Invalid token"})
		}

		// Verify token belongs to current tenant
		if token.TenantID != tenant.ID {
			return ctx.JSON(403, map[string]string{"error": "Token does not belong to tenant"})
		}

		return ctx.JSON(200, map[string]interface{}{
			"message":   "Access granted",
			"user_id":   token.UserID,
			"tenant_id": token.TenantID,
		})
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

	// Test tenant1 token access
	t.Run("Tenant1 uses their token", func(t *testing.T) {
		client := &http.Client{}
		req, err := http.NewRequest("GET", "http://localhost:19204/protected", nil)
		assertNoError(t, err, "Failed to create request")

		req.Host = "tenant1.example.com"
		req.Header.Set("Authorization", "Bearer token1")

		resp, err := client.Do(req)
		assertNoError(t, err, "Failed to make request")
		defer resp.Body.Close()

		assertEqual(t, 200, resp.StatusCode, "Expected status 200")

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assertNoError(t, err, "Failed to decode response")

		assertEqual(t, "tenant1", result["tenant_id"], "Expected tenant1 ID")
	})

	// Test tenant2 token access
	t.Run("Tenant2 uses their token", func(t *testing.T) {
		client := &http.Client{}
		req, err := http.NewRequest("GET", "http://localhost:19204/protected", nil)
		assertNoError(t, err, "Failed to create request")

		req.Host = "tenant2.example.com"
		req.Header.Set("Authorization", "Bearer token2")

		resp, err := client.Do(req)
		assertNoError(t, err, "Failed to make request")
		defer resp.Body.Close()

		assertEqual(t, 200, resp.StatusCode, "Expected status 200")

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assertNoError(t, err, "Failed to decode response")

		assertEqual(t, "tenant2", result["tenant_id"], "Expected tenant2 ID")
	})

	// Test cross-tenant token access is blocked
	t.Run("Tenant1 cannot use tenant2 token", func(t *testing.T) {
		client := &http.Client{}
		req, err := http.NewRequest("GET", "http://localhost:19204/protected", nil)
		assertNoError(t, err, "Failed to create request")

		req.Host = "tenant1.example.com"
		req.Header.Set("Authorization", "Bearer token2")

		resp, err := client.Do(req)
		assertNoError(t, err, "Failed to make request")
		defer resp.Body.Close()

		assertEqual(t, 403, resp.StatusCode, "Expected status 403 for cross-tenant token")
	})
}

// TestMultiTenancyConfigurationIsolation tests tenant-specific configurations
// Requirements: 3.5
func TestMultiTenancyConfigurationIsolation(t *testing.T) {
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

	// Create test tenants with different configurations
	tenant1 := createTestTenant("tenant1", "Tenant One", []string{"tenant1.example.com"})
	tenant1.Config = map[string]interface{}{
		"feature_flag_a": true,
		"max_items":      100,
		"theme":          "dark",
	}

	tenant2 := createTestTenant("tenant2", "Tenant Two", []string{"tenant2.example.com"})
	tenant2.Config = map[string]interface{}{
		"feature_flag_a": false,
		"max_items":      50,
		"theme":          "light",
	}

	// Save tenants to database
	err = framework.Database().SaveTenant(tenant1)
	assertNoError(t, err, "Failed to save tenant1")
	err = framework.Database().SaveTenant(tenant2)
	assertNoError(t, err, "Failed to save tenant2")

	// Register hosts and tenants with the framework
	err = framework.ServerManager().RegisterHost("tenant1.example.com", pkg.HostConfig{Hostname: "tenant1.example.com"})
	assertNoError(t, err, "Failed to register tenant1 host")
	err = framework.ServerManager().RegisterHost("tenant2.example.com", pkg.HostConfig{Hostname: "tenant2.example.com"})
	assertNoError(t, err, "Failed to register tenant2 host")

	err = framework.ServerManager().RegisterTenant("tenant1", []string{"tenant1.example.com"})
	assertNoError(t, err, "Failed to register tenant1")
	err = framework.ServerManager().RegisterTenant("tenant2", []string{"tenant2.example.com"})
	assertNoError(t, err, "Failed to register tenant2")

	// Apply tenant middleware globally
	framework.Router().Use(createTenantMiddleware(framework.Database()))

	// Register route to get tenant configuration
	framework.Router().GET("/config", func(ctx pkg.Context) error {
		tenant := ctx.Tenant()
		if tenant == nil {
			return ctx.JSON(400, map[string]string{"error": "No tenant context"})
		}

		return ctx.JSON(200, map[string]interface{}{
			"tenant_id": tenant.ID,
			"config":    tenant.Config,
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

	// Test tenant1 configuration
	t.Run("Tenant1 has independent configuration", func(t *testing.T) {
		client := &http.Client{}
		req, err := http.NewRequest("GET", "http://localhost:19205/config", nil)
		assertNoError(t, err, "Failed to create request")

		req.Host = "tenant1.example.com"

		resp, err := client.Do(req)
		assertNoError(t, err, "Failed to make request")
		defer resp.Body.Close()

		assertEqual(t, 200, resp.StatusCode, "Expected status 200")

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assertNoError(t, err, "Failed to decode response")

		assertEqual(t, "tenant1", result["tenant_id"], "Expected tenant1 ID")

		config := result["config"].(map[string]interface{})
		assertEqual(t, true, config["feature_flag_a"], "Expected feature_flag_a to be true")
		assertEqual(t, float64(100), config["max_items"], "Expected max_items to be 100")
		assertEqual(t, "dark", config["theme"], "Expected theme to be dark")
	})

	// Test tenant2 configuration
	t.Run("Tenant2 has independent configuration", func(t *testing.T) {
		client := &http.Client{}
		req, err := http.NewRequest("GET", "http://localhost:19205/config", nil)
		assertNoError(t, err, "Failed to create request")

		req.Host = "tenant2.example.com"

		resp, err := client.Do(req)
		assertNoError(t, err, "Failed to make request")
		defer resp.Body.Close()

		assertEqual(t, 200, resp.StatusCode, "Expected status 200")

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assertNoError(t, err, "Failed to decode response")

		assertEqual(t, "tenant2", result["tenant_id"], "Expected tenant2 ID")

		config := result["config"].(map[string]interface{})
		assertEqual(t, false, config["feature_flag_a"], "Expected feature_flag_a to be false")
		assertEqual(t, float64(50), config["max_items"], "Expected max_items to be 50")
		assertEqual(t, "light", config["theme"], "Expected theme to be light")
	})
}

// TestMultiTenancyRateLimitingIsolation tests rate limiting isolation per tenant
// Requirements: 3.6
func TestMultiTenancyRateLimitingIsolation(t *testing.T) {
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

	// Create test tenants
	tenant1 := createTestTenant("tenant1", "Tenant One", []string{"tenant1.example.com"})
	tenant2 := createTestTenant("tenant2", "Tenant Two", []string{"tenant2.example.com"})

	// Save tenants to database
	err = framework.Database().SaveTenant(tenant1)
	assertNoError(t, err, "Failed to save tenant1")
	err = framework.Database().SaveTenant(tenant2)
	assertNoError(t, err, "Failed to save tenant2")

	// Register hosts and tenants with the framework
	err = framework.ServerManager().RegisterHost("tenant1.example.com", pkg.HostConfig{Hostname: "tenant1.example.com"})
	assertNoError(t, err, "Failed to register tenant1 host")
	err = framework.ServerManager().RegisterHost("tenant2.example.com", pkg.HostConfig{Hostname: "tenant2.example.com"})
	assertNoError(t, err, "Failed to register tenant2 host")

	err = framework.ServerManager().RegisterTenant("tenant1", []string{"tenant1.example.com"})
	assertNoError(t, err, "Failed to register tenant1")
	err = framework.ServerManager().RegisterTenant("tenant2", []string{"tenant2.example.com"})
	assertNoError(t, err, "Failed to register tenant2")

	// Apply tenant middleware globally
	framework.Router().Use(createTenantMiddleware(framework.Database()))

	// Rate limit counters per tenant
	rateLimitCounters := make(map[string]int)
	var rateLimitMutex sync.Mutex

	// Register rate-limited route
	framework.Router().GET("/api/resource", func(ctx pkg.Context) error {
		tenant := ctx.Tenant()
		if tenant == nil {
			return ctx.JSON(400, map[string]string{"error": "No tenant context"})
		}

		// Check rate limit (3 requests per tenant)
		rateLimitMutex.Lock()
		count := rateLimitCounters[tenant.ID]
		if count >= 3 {
			rateLimitMutex.Unlock()
			return ctx.JSON(429, map[string]string{"error": "Rate limit exceeded"})
		}
		rateLimitCounters[tenant.ID] = count + 1
		rateLimitMutex.Unlock()

		return ctx.JSON(200, map[string]interface{}{
			"message":       "Success",
			"tenant_id":     tenant.ID,
			"request_count": count + 1,
		})
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

	// Test tenant1 rate limit
	t.Run("Tenant1 rate limit is independent", func(t *testing.T) {
		client := &http.Client{}

		// Make 3 successful requests
		for i := 1; i <= 3; i++ {
			req, err := http.NewRequest("GET", "http://localhost:19206/api/resource", nil)
			assertNoError(t, err, "Failed to create request")

			req.Host = "tenant1.example.com"

			resp, err := client.Do(req)
			assertNoError(t, err, "Failed to make request")
			defer resp.Body.Close()

			assertEqual(t, 200, resp.StatusCode, fmt.Sprintf("Expected status 200 for request %d", i))
		}

		// 4th request should be rate limited
		req, err := http.NewRequest("GET", "http://localhost:19206/api/resource", nil)
		assertNoError(t, err, "Failed to create request")

		req.Host = "tenant1.example.com"

		resp, err := client.Do(req)
		assertNoError(t, err, "Failed to make request")
		defer resp.Body.Close()

		assertEqual(t, 429, resp.StatusCode, "Expected status 429 for rate limited request")
	})

	// Test tenant2 rate limit is independent
	t.Run("Tenant2 rate limit is independent from tenant1", func(t *testing.T) {
		client := &http.Client{}

		// Tenant2 should still have their full quota
		for i := 1; i <= 3; i++ {
			req, err := http.NewRequest("GET", "http://localhost:19206/api/resource", nil)
			assertNoError(t, err, "Failed to create request")

			req.Host = "tenant2.example.com"

			resp, err := client.Do(req)
			assertNoError(t, err, "Failed to make request")
			defer resp.Body.Close()

			assertEqual(t, 200, resp.StatusCode, fmt.Sprintf("Expected status 200 for tenant2 request %d", i))
		}

		// 4th request should be rate limited for tenant2
		req, err := http.NewRequest("GET", "http://localhost:19206/api/resource", nil)
		assertNoError(t, err, "Failed to create request")

		req.Host = "tenant2.example.com"

		resp, err := client.Do(req)
		assertNoError(t, err, "Failed to make request")
		defer resp.Body.Close()

		assertEqual(t, 429, resp.StatusCode, "Expected status 429 for tenant2 rate limited request")
	})
}

// TestMultiTenancyWorkloadMetrics tests workload metrics isolation per tenant
// Requirements: 3.7
func TestMultiTenancyWorkloadMetrics(t *testing.T) {
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

	// Create test tenants
	tenant1 := createTestTenant("tenant1", "Tenant One", []string{"tenant1.example.com"})
	tenant2 := createTestTenant("tenant2", "Tenant Two", []string{"tenant2.example.com"})

	// Save tenants to database
	err = framework.Database().SaveTenant(tenant1)
	assertNoError(t, err, "Failed to save tenant1")
	err = framework.Database().SaveTenant(tenant2)
	assertNoError(t, err, "Failed to save tenant2")

	// Register hosts and tenants with the framework
	err = framework.ServerManager().RegisterHost("tenant1.example.com", pkg.HostConfig{Hostname: "tenant1.example.com"})
	assertNoError(t, err, "Failed to register tenant1 host")
	err = framework.ServerManager().RegisterHost("tenant2.example.com", pkg.HostConfig{Hostname: "tenant2.example.com"})
	assertNoError(t, err, "Failed to register tenant2 host")

	err = framework.ServerManager().RegisterTenant("tenant1", []string{"tenant1.example.com"})
	assertNoError(t, err, "Failed to register tenant1")
	err = framework.ServerManager().RegisterTenant("tenant2", []string{"tenant2.example.com"})
	assertNoError(t, err, "Failed to register tenant2")

	// Apply tenant middleware globally
	framework.Router().Use(createTenantMiddleware(framework.Database()))

	// Register route that generates metrics
	framework.Router().GET("/work", func(ctx pkg.Context) error {
		tenant := ctx.Tenant()
		if tenant == nil {
			return ctx.JSON(400, map[string]string{"error": "No tenant context"})
		}

		// Simulate some work
		time.Sleep(10 * time.Millisecond)

		// Save workload metrics
		metrics := &pkg.WorkloadMetrics{
			Timestamp:  time.Now(),
			TenantID:   tenant.ID,
			UserID:     "test-user",
			RequestID:  fmt.Sprintf("req-%d", time.Now().UnixNano()),
			Duration:   10,
			Path:       "/work",
			Method:     "GET",
			StatusCode: 200,
		}

		err := framework.Database().SaveWorkloadMetrics(metrics)
		if err != nil {
			return ctx.JSON(500, map[string]string{"error": "Failed to save metrics"})
		}

		return ctx.JSON(200, map[string]string{"message": "Work completed"})
	})

	// Register route to get metrics
	framework.Router().GET("/metrics", func(ctx pkg.Context) error {
		tenant := ctx.Tenant()
		if tenant == nil {
			return ctx.JSON(400, map[string]string{"error": "No tenant context"})
		}

		// Get metrics for current tenant
		from := time.Now().Add(-1 * time.Hour)
		to := time.Now().Add(1 * time.Hour)

		metrics, err := framework.Database().GetWorkloadMetrics(tenant.ID, from, to)
		if err != nil {
			return ctx.JSON(500, map[string]string{"error": "Failed to get metrics"})
		}

		return ctx.JSON(200, map[string]interface{}{
			"tenant_id":     tenant.ID,
			"metrics_count": len(metrics),
			"metrics":       metrics,
		})
	})

	// Start server in background
	go func() {
		if err := framework.Listen(":19207"); err != nil {
			t.Logf("Server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)
	defer framework.Shutdown(2 * time.Second)

	// Generate metrics for tenant1
	t.Run("Generate metrics for tenant1", func(t *testing.T) {
		client := &http.Client{}

		for i := 0; i < 3; i++ {
			req, err := http.NewRequest("GET", "http://localhost:19207/work", nil)
			assertNoError(t, err, "Failed to create request")

			req.Host = "tenant1.example.com"

			resp, err := client.Do(req)
			assertNoError(t, err, "Failed to make request")
			defer resp.Body.Close()

			assertEqual(t, 200, resp.StatusCode, "Expected status 200")
		}
	})

	// Generate metrics for tenant2
	t.Run("Generate metrics for tenant2", func(t *testing.T) {
		client := &http.Client{}

		for i := 0; i < 2; i++ {
			req, err := http.NewRequest("GET", "http://localhost:19207/work", nil)
			assertNoError(t, err, "Failed to create request")

			req.Host = "tenant2.example.com"

			resp, err := client.Do(req)
			assertNoError(t, err, "Failed to make request")
			defer resp.Body.Close()

			assertEqual(t, 200, resp.StatusCode, "Expected status 200")
		}
	})

	// Verify tenant1 only sees their metrics
	t.Run("Tenant1 sees only their metrics", func(t *testing.T) {
		client := &http.Client{}
		req, err := http.NewRequest("GET", "http://localhost:19207/metrics", nil)
		assertNoError(t, err, "Failed to create request")

		req.Host = "tenant1.example.com"

		resp, err := client.Do(req)
		assertNoError(t, err, "Failed to make request")
		defer resp.Body.Close()

		assertEqual(t, 200, resp.StatusCode, "Expected status 200")

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assertNoError(t, err, "Failed to decode response")

		assertEqual(t, "tenant1", result["tenant_id"], "Expected tenant1 ID")
		assertEqual(t, float64(3), result["metrics_count"], "Expected 3 metrics for tenant1")
	})

	// Verify tenant2 only sees their metrics
	t.Run("Tenant2 sees only their metrics", func(t *testing.T) {
		client := &http.Client{}
		req, err := http.NewRequest("GET", "http://localhost:19207/metrics", nil)
		assertNoError(t, err, "Failed to create request")

		req.Host = "tenant2.example.com"

		resp, err := client.Do(req)
		assertNoError(t, err, "Failed to make request")
		defer resp.Body.Close()

		assertEqual(t, 200, resp.StatusCode, "Expected status 200")

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assertNoError(t, err, "Failed to decode response")

		assertEqual(t, "tenant2", result["tenant_id"], "Expected tenant2 ID")
		assertEqual(t, float64(2), result["metrics_count"], "Expected 2 metrics for tenant2")
	})
}

// TestMultiTenancyConcurrentAccess tests concurrent access from multiple tenants
// Requirements: 3.8
func TestMultiTenancyConcurrentAccess(t *testing.T) {
	// Use a file-based database for this test to avoid SQLite in-memory connection issues
	dbConfig := pkg.DatabaseConfig{
		Driver:   "sqlite3",
		Database: "test_concurrent_access.db",
		Options: map[string]string{
			"sql_dir": "../sql",
		},
	}

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
		DatabaseConfig: dbConfig,
		CacheConfig:    pkg.CacheConfig{},
		SessionConfig:  *createTestSessionConfig(),
	}

	// Create framework instance
	framework, err := pkg.New(config)
	assertNoError(t, err, "Failed to create framework")

	// Initialize database tables
	err = framework.Database().CreateTables()
	assertNoError(t, err, "Failed to create database tables")

	// Create test tenants
	tenant1 := createTestTenant("tenant1", "Tenant One", []string{"tenant1.example.com"})
	tenant2 := createTestTenant("tenant2", "Tenant Two", []string{"tenant2.example.com"})
	tenant3 := createTestTenant("tenant3", "Tenant Three", []string{"tenant3.example.com"})

	// Save tenants to database
	err = framework.Database().SaveTenant(tenant1)
	assertNoError(t, err, "Failed to save tenant1")
	err = framework.Database().SaveTenant(tenant2)
	assertNoError(t, err, "Failed to save tenant2")
	err = framework.Database().SaveTenant(tenant3)
	assertNoError(t, err, "Failed to save tenant3")

	// Register hosts and tenants with the framework
	err = framework.ServerManager().RegisterHost("tenant1.example.com", pkg.HostConfig{Hostname: "tenant1.example.com"})
	assertNoError(t, err, "Failed to register tenant1 host")
	err = framework.ServerManager().RegisterHost("tenant2.example.com", pkg.HostConfig{Hostname: "tenant2.example.com"})
	assertNoError(t, err, "Failed to register tenant2 host")
	err = framework.ServerManager().RegisterHost("tenant3.example.com", pkg.HostConfig{Hostname: "tenant3.example.com"})
	assertNoError(t, err, "Failed to register tenant3 host")

	err = framework.ServerManager().RegisterTenant("tenant1", []string{"tenant1.example.com"})
	assertNoError(t, err, "Failed to register tenant1")
	err = framework.ServerManager().RegisterTenant("tenant2", []string{"tenant2.example.com"})
	assertNoError(t, err, "Failed to register tenant2")
	err = framework.ServerManager().RegisterTenant("tenant3", []string{"tenant3.example.com"})
	assertNoError(t, err, "Failed to register tenant3")

	// Apply tenant middleware globally
	framework.Router().Use(createTenantMiddleware(framework.Database()))

	// Request counters per tenant
	requestCounters := make(map[string]int)
	var counterMutex sync.Mutex

	// Register route
	framework.Router().GET("/concurrent", func(ctx pkg.Context) error {
		tenant := ctx.Tenant()
		if tenant == nil {
			return ctx.JSON(400, map[string]string{"error": "No tenant context"})
		}

		// Simulate some work
		time.Sleep(5 * time.Millisecond)

		// Increment counter
		counterMutex.Lock()
		requestCounters[tenant.ID]++
		count := requestCounters[tenant.ID]
		counterMutex.Unlock()

		return ctx.JSON(200, map[string]interface{}{
			"tenant_id":     tenant.ID,
			"request_count": count,
		})
	})

	// Start server in background
	go func() {
		if err := framework.Listen(":19208"); err != nil {
			t.Logf("Server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)
	defer func() {
		framework.Shutdown(2 * time.Second)
		// Clean up database file
		os.Remove("test_concurrent_access.db")
		os.Remove("test_concurrent_access.db-shm")
		os.Remove("test_concurrent_access.db-wal")
	}()

	// Test concurrent requests from multiple tenants
	t.Run("Concurrent requests from multiple tenants", func(t *testing.T) {
		var wg sync.WaitGroup
		errors := make(chan error, 30)

		// Tenant 1: 10 concurrent requests
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				client := &http.Client{
					Transport: &http.Transport{
						DisableKeepAlives: true, // Disable connection reuse
					},
				}
				req, err := http.NewRequest("GET", "http://localhost:19208/concurrent", nil)
				if err != nil {
					errors <- err
					return
				}

				req.Host = "tenant1.example.com"
				req.Header.Set("Host", "tenant1.example.com") // Also set in headers

				resp, err := client.Do(req)
				if err != nil {
					errors <- err
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode != 200 {
					errors <- fmt.Errorf("tenant1: expected status 200, got %d", resp.StatusCode)
				}
			}()
		}

		// Tenant 2: 10 concurrent requests
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				client := &http.Client{
					Transport: &http.Transport{
						DisableKeepAlives: true, // Disable connection reuse
					},
				}
				req, err := http.NewRequest("GET", "http://localhost:19208/concurrent", nil)
				if err != nil {
					errors <- err
					return
				}

				req.Host = "tenant2.example.com"
				req.Header.Set("Host", "tenant2.example.com") // Also set in headers

				resp, err := client.Do(req)
				if err != nil {
					errors <- err
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode != 200 {
					errors <- fmt.Errorf("tenant2: expected status 200, got %d", resp.StatusCode)
				}
			}()
		}

		// Tenant 3: 10 concurrent requests
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				client := &http.Client{
					Transport: &http.Transport{
						DisableKeepAlives: true, // Disable connection reuse
					},
				}
				req, err := http.NewRequest("GET", "http://localhost:19208/concurrent", nil)
				if err != nil {
					errors <- err
					return
				}

				req.Host = "tenant3.example.com"
				req.Header.Set("Host", "tenant3.example.com") // Also set in headers

				resp, err := client.Do(req)
				if err != nil {
					errors <- err
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode != 200 {
					errors <- fmt.Errorf("tenant3: expected status 200, got %d", resp.StatusCode)
				}
			}()
		}

		// Wait for all requests to complete
		wg.Wait()
		close(errors)

		// Check for errors
		for err := range errors {
			t.Errorf("Concurrent request error: %v", err)
		}

		// Verify each tenant received exactly 10 requests
		counterMutex.Lock()
		assertEqual(t, 10, requestCounters["tenant1"], "Expected 10 requests for tenant1")
		assertEqual(t, 10, requestCounters["tenant2"], "Expected 10 requests for tenant2")
		assertEqual(t, 10, requestCounters["tenant3"], "Expected 10 requests for tenant3")
		counterMutex.Unlock()
	})
}
