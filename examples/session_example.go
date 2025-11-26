package main

import (
	"fmt"
	"log"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
	// ============================================================================
	// Configuration Setup
	// ============================================================================
	config := pkg.FrameworkConfig{
		ServerConfig: pkg.ServerConfig{
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			EnableHTTP1:  true,
			EnableHTTP2:  true,
		},
		DatabaseConfig: pkg.DatabaseConfig{
			Driver:   "sqlite3",
			Database: "session_example.db",
		},
		CacheConfig: pkg.CacheConfig{
			Type:       "memory",
			MaxSize:    10 * 1024 * 1024, // 10 MB
			DefaultTTL: 5 * time.Minute,
		},
		// Session configuration with different storage options
		SessionConfig: pkg.SessionConfig{
			StorageType:     pkg.SessionStorageDatabase, // Can be: database, cache, filesystem
			CookieName:      "rockstar_session",
			CookiePath:      "/",
			CookieDomain:    "",
			CookieSecure:    false, // Set to true in production with HTTPS
			CookieHTTPOnly:  true,
			CookieSameSite:  "Lax",
			SessionLifetime: 24 * time.Hour,
			EncryptionKey:   []byte("12345678901234567890123456789012"), // 32 bytes for AES-256
			FilesystemPath:  "./sessions",
			CleanupInterval: 15 * time.Minute,
		},
	}

	// ============================================================================
	// Framework Initialization
	// ============================================================================
	app, err := pkg.New(config)
	if err != nil {
		log.Fatalf("Failed to create framework: %v", err)
	}

	// ============================================================================
	// Route Registration
	// ============================================================================
	router := app.Router()

	// Session creation and management
	router.GET("/api/session/create", createSessionHandler)
	router.GET("/api/session/load", loadSessionHandler)
	router.GET("/api/session/destroy", destroySessionHandler)

	// Session data operations
	router.GET("/api/session/set", setSessionDataHandler)
	router.GET("/api/session/get", getSessionDataHandler)
	router.GET("/api/session/delete", deleteSessionDataHandler)
	router.GET("/api/session/clear", clearSessionDataHandler)

	// Session lifecycle
	router.GET("/api/session/refresh", refreshSessionHandler)
	router.GET("/api/session/validate", validateSessionHandler)

	// Encrypted session cookies
	router.GET("/api/session/cookie/set", setSessionCookieHandler)
	router.GET("/api/session/cookie/get", getSessionCookieHandler)

	// Session information
	router.GET("/api/session/info", sessionInfoHandler)

	// ============================================================================
	// Server Startup
	// ============================================================================
	fmt.Println("🎸 Rockstar Web Framework - Session Example")
	fmt.Println("=" + "==========================================================")
	fmt.Println()
	fmt.Println("Server listening on: http://localhost:8080")
	fmt.Println()
	fmt.Println("Try these commands:")
	fmt.Println("  # Session creation and management")
	fmt.Println("  curl -c cookies.txt http://localhost:8080/api/session/create")
	fmt.Println("  curl -b cookies.txt http://localhost:8080/api/session/load")
	fmt.Println("  curl -b cookies.txt http://localhost:8080/api/session/destroy")
	fmt.Println()
	fmt.Println("  # Session data operations")
	fmt.Println("  curl -b cookies.txt http://localhost:8080/api/session/set")
	fmt.Println("  curl -b cookies.txt http://localhost:8080/api/session/get")
	fmt.Println("  curl -b cookies.txt http://localhost:8080/api/session/delete")
	fmt.Println("  curl -b cookies.txt http://localhost:8080/api/session/clear")
	fmt.Println()
	fmt.Println("  # Session lifecycle")
	fmt.Println("  curl -b cookies.txt http://localhost:8080/api/session/refresh")
	fmt.Println("  curl -b cookies.txt http://localhost:8080/api/session/validate")
	fmt.Println()
	fmt.Println("  # Encrypted session cookies")
	fmt.Println("  curl -c cookies.txt http://localhost:8080/api/session/cookie/set")
	fmt.Println("  curl -b cookies.txt http://localhost:8080/api/session/cookie/get")
	fmt.Println()
	fmt.Println("  # Session information")
	fmt.Println("  curl -b cookies.txt http://localhost:8080/api/session/info")
	fmt.Println()
	fmt.Println("=" + "==========================================================")
	fmt.Println()

	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// ============================================================================
// Handler Functions - Session Creation and Management
// ============================================================================

// createSessionHandler demonstrates creating a new session
func createSessionHandler(ctx pkg.Context) error {
	sessionMgr := ctx.Session()

	// Create a new session
	session, err := sessionMgr.Create(ctx)
	if err != nil {
		return ctx.JSON(500, map[string]interface{}{
			"error": "Failed to create session",
		})
	}

	// Store some initial data in the session
	session.Data["user_id"] = 123
	session.Data["username"] = "john_doe"
	session.Data["login_time"] = time.Now().Format(time.RFC3339)

	// Save the session
	err = sessionMgr.Save(ctx, session)
	if err != nil {
		return ctx.JSON(500, map[string]interface{}{
			"error": "Failed to save session",
		})
	}

	// Set session cookie
	err = sessionMgr.SetCookie(ctx, session)
	if err != nil {
		return ctx.JSON(500, map[string]interface{}{
			"error": "Failed to set session cookie",
		})
	}

	return ctx.JSON(200, map[string]interface{}{
		"message":    "Session created successfully",
		"session_id": session.ID,
		"expires_at": session.ExpiresAt.Format(time.RFC3339),
		"data":       session.Data,
		"note":       "Session cookie has been set. Use -c cookies.txt to save it.",
	})
}

// loadSessionHandler demonstrates loading an existing session
func loadSessionHandler(ctx pkg.Context) error {
	sessionMgr := ctx.Session()

	// Get session from cookie
	session, err := sessionMgr.GetSessionFromCookie(ctx)
	if err != nil {
		return ctx.JSON(404, map[string]interface{}{
			"error":   "Session not found",
			"message": "No valid session cookie found. Create a session first.",
		})
	}

	return ctx.JSON(200, map[string]interface{}{
		"message":    "Session loaded successfully",
		"session_id": session.ID,
		"created_at": session.CreatedAt.Format(time.RFC3339),
		"updated_at": session.UpdatedAt.Format(time.RFC3339),
		"expires_at": session.ExpiresAt.Format(time.RFC3339),
		"data":       session.Data,
	})
}

// destroySessionHandler demonstrates destroying a session
func destroySessionHandler(ctx pkg.Context) error {
	sessionMgr := ctx.Session()

	// Get session from cookie
	session, err := sessionMgr.GetSessionFromCookie(ctx)
	if err != nil {
		return ctx.JSON(404, map[string]interface{}{
			"error": "Session not found",
		})
	}

	// Destroy the session
	err = sessionMgr.Destroy(ctx, session.ID)
	if err != nil {
		return ctx.JSON(500, map[string]interface{}{
			"error": "Failed to destroy session",
		})
	}

	return ctx.JSON(200, map[string]interface{}{
		"message": "Session destroyed successfully",
		"note":    "Session has been removed from storage",
	})
}

// ============================================================================
// Handler Functions - Session Data Operations
// ============================================================================

// setSessionDataHandler demonstrates setting data in a session
func setSessionDataHandler(ctx pkg.Context) error {
	sessionMgr := ctx.Session()

	// Get session from cookie
	session, err := sessionMgr.GetSessionFromCookie(ctx)
	if err != nil {
		return ctx.JSON(404, map[string]interface{}{
			"error": "Session not found. Create a session first.",
		})
	}

	// Set various types of data
	err = sessionMgr.Set(session.ID, "preferences", map[string]interface{}{
		"theme":    "dark",
		"language": "en",
		"timezone": "UTC",
	})
	if err != nil {
		return ctx.JSON(500, map[string]interface{}{
			"error": "Failed to set session data",
		})
	}

	err = sessionMgr.Set(session.ID, "cart_items", []string{"item1", "item2", "item3"})
	if err != nil {
		return ctx.JSON(500, map[string]interface{}{
			"error": "Failed to set cart items",
		})
	}

	err = sessionMgr.Set(session.ID, "last_activity", time.Now().Format(time.RFC3339))
	if err != nil {
		return ctx.JSON(500, map[string]interface{}{
			"error": "Failed to set last activity",
		})
	}

	return ctx.JSON(200, map[string]interface{}{
		"message": "Session data set successfully",
		"keys":    []string{"preferences", "cart_items", "last_activity"},
	})
}

// getSessionDataHandler demonstrates getting data from a session
func getSessionDataHandler(ctx pkg.Context) error {
	sessionMgr := ctx.Session()

	// Get session from cookie
	session, err := sessionMgr.GetSessionFromCookie(ctx)
	if err != nil {
		return ctx.JSON(404, map[string]interface{}{
			"error": "Session not found",
		})
	}

	// Get specific values
	preferences, err := sessionMgr.Get(session.ID, "preferences")
	if err != nil {
		preferences = "not set"
	}

	cartItems, err := sessionMgr.Get(session.ID, "cart_items")
	if err != nil {
		cartItems = "not set"
	}

	lastActivity, err := sessionMgr.Get(session.ID, "last_activity")
	if err != nil {
		lastActivity = "not set"
	}

	return ctx.JSON(200, map[string]interface{}{
		"message":       "Session data retrieved successfully",
		"preferences":   preferences,
		"cart_items":    cartItems,
		"last_activity": lastActivity,
		"all_data":      session.Data,
	})
}

// deleteSessionDataHandler demonstrates deleting data from a session
func deleteSessionDataHandler(ctx pkg.Context) error {
	sessionMgr := ctx.Session()

	// Get session from cookie
	session, err := sessionMgr.GetSessionFromCookie(ctx)
	if err != nil {
		return ctx.JSON(404, map[string]interface{}{
			"error": "Session not found",
		})
	}

	// Delete a specific key
	err = sessionMgr.Delete(session.ID, "cart_items")
	if err != nil {
		return ctx.JSON(500, map[string]interface{}{
			"error": "Failed to delete session data",
		})
	}

	return ctx.JSON(200, map[string]interface{}{
		"message": "Session data deleted successfully",
		"deleted": "cart_items",
	})
}

// clearSessionDataHandler demonstrates clearing all data from a session
func clearSessionDataHandler(ctx pkg.Context) error {
	sessionMgr := ctx.Session()

	// Get session from cookie
	session, err := sessionMgr.GetSessionFromCookie(ctx)
	if err != nil {
		return ctx.JSON(404, map[string]interface{}{
			"error": "Session not found",
		})
	}

	// Clear all session data
	err = sessionMgr.Clear(session.ID)
	if err != nil {
		return ctx.JSON(500, map[string]interface{}{
			"error": "Failed to clear session data",
		})
	}

	return ctx.JSON(200, map[string]interface{}{
		"message": "Session data cleared successfully",
		"note":    "All data has been removed from the session",
	})
}

// ============================================================================
// Handler Functions - Session Lifecycle
// ============================================================================

// refreshSessionHandler demonstrates refreshing a session's expiration
func refreshSessionHandler(ctx pkg.Context) error {
	sessionMgr := ctx.Session()

	// Get session from cookie
	session, err := sessionMgr.GetSessionFromCookie(ctx)
	if err != nil {
		return ctx.JSON(404, map[string]interface{}{
			"error": "Session not found",
		})
	}

	oldExpiry := session.ExpiresAt

	// Refresh the session (extends expiration)
	err = sessionMgr.Refresh(ctx, session.ID)
	if err != nil {
		return ctx.JSON(500, map[string]interface{}{
			"error": "Failed to refresh session",
		})
	}

	// Load the refreshed session
	refreshedSession, _ := sessionMgr.Load(ctx, session.ID)

	return ctx.JSON(200, map[string]interface{}{
		"message":        "Session refreshed successfully",
		"old_expires_at": oldExpiry.Format(time.RFC3339),
		"new_expires_at": refreshedSession.ExpiresAt.Format(time.RFC3339),
		"extended_by":    refreshedSession.ExpiresAt.Sub(oldExpiry).String(),
	})
}

// validateSessionHandler demonstrates validating a session
func validateSessionHandler(ctx pkg.Context) error {
	sessionMgr := ctx.Session()

	// Get session from cookie
	session, err := sessionMgr.GetSessionFromCookie(ctx)
	if err != nil {
		return ctx.JSON(404, map[string]interface{}{
			"error":   "Session not found or invalid",
			"valid":   false,
			"expired": true,
		})
	}

	// Check if session is valid
	isValid := sessionMgr.IsValid(session.ID)
	isExpired := sessionMgr.IsExpired(session.ID)

	return ctx.JSON(200, map[string]interface{}{
		"message":    "Session validation complete",
		"session_id": session.ID,
		"valid":      isValid,
		"expired":    isExpired,
		"expires_at": session.ExpiresAt.Format(time.RFC3339),
	})
}

// ============================================================================
// Handler Functions - Encrypted Session Cookies
// ============================================================================

// setSessionCookieHandler demonstrates setting an encrypted session cookie
func setSessionCookieHandler(ctx pkg.Context) error {
	sessionMgr := ctx.Session()

	// Create a new session
	session, err := sessionMgr.Create(ctx)
	if err != nil {
		return ctx.JSON(500, map[string]interface{}{
			"error": "Failed to create session",
		})
	}

	// Add some data
	session.Data["secure_data"] = "This is encrypted in the cookie"
	session.Data["timestamp"] = time.Now().Format(time.RFC3339)

	// Save the session
	err = sessionMgr.Save(ctx, session)
	if err != nil {
		return ctx.JSON(500, map[string]interface{}{
			"error": "Failed to save session",
		})
	}

	// Set encrypted session cookie
	err = sessionMgr.SetCookie(ctx, session)
	if err != nil {
		return ctx.JSON(500, map[string]interface{}{
			"error": "Failed to set encrypted session cookie",
		})
	}

	return ctx.JSON(200, map[string]interface{}{
		"message":    "Encrypted session cookie set successfully",
		"session_id": session.ID,
		"note":       "The session ID is encrypted in the cookie using AES-256",
		"security":   "Cookie is HttpOnly and uses encryption for security",
	})
}

// getSessionCookieHandler demonstrates getting data from an encrypted session cookie
func getSessionCookieHandler(ctx pkg.Context) error {
	sessionMgr := ctx.Session()

	// Get session from encrypted cookie
	session, err := sessionMgr.GetSessionFromCookie(ctx)
	if err != nil {
		return ctx.JSON(404, map[string]interface{}{
			"error":   "Session not found",
			"message": "No valid encrypted session cookie found",
		})
	}

	return ctx.JSON(200, map[string]interface{}{
		"message":    "Session retrieved from encrypted cookie",
		"session_id": session.ID,
		"data":       session.Data,
		"note":       "The session ID was decrypted from the cookie",
	})
}

// ============================================================================
// Handler Functions - Session Information
// ============================================================================

// sessionInfoHandler demonstrates getting comprehensive session information
func sessionInfoHandler(ctx pkg.Context) error {
	sessionMgr := ctx.Session()

	// Get session from cookie
	session, err := sessionMgr.GetSessionFromCookie(ctx)
	if err != nil {
		return ctx.JSON(404, map[string]interface{}{
			"error": "Session not found",
		})
	}

	// Calculate session age and remaining time
	age := time.Since(session.CreatedAt)
	remaining := time.Until(session.ExpiresAt)

	return ctx.JSON(200, map[string]interface{}{
		"message": "Session information",
		"session": map[string]interface{}{
			"id":         session.ID,
			"user_id":    session.UserID,
			"tenant_id":  session.TenantID,
			"ip_address": session.IPAddress,
			"user_agent": session.UserAgent,
			"created_at": session.CreatedAt.Format(time.RFC3339),
			"updated_at": session.UpdatedAt.Format(time.RFC3339),
			"expires_at": session.ExpiresAt.Format(time.RFC3339),
			"age":        age.String(),
			"remaining":  remaining.String(),
			"data_count": len(session.Data),
		},
		"data": session.Data,
		"storage": map[string]interface{}{
			"type":      "database", // From config
			"encrypted": true,
		},
	})
}
