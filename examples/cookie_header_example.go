package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"net/http"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
	fmt.Println("=== Rockstar Web Framework - Cookie and Header Management Example ===\n")

	// Example 1: Basic Cookie Management
	fmt.Println("1. Basic Cookie Management")
	basicCookieExample()

	// Example 2: Encrypted Cookie Management
	fmt.Println("\n2. Encrypted Cookie Management")
	encryptedCookieExample()

	// Example 3: Header Management
	fmt.Println("\n3. Header Management")
	headerManagementExample()

	// Example 4: Complete HTTP Server with Cookie and Header Management
	fmt.Println("\n4. Starting HTTP Server with Cookie and Header Management...")
	startServer()
}

// basicCookieExample demonstrates basic cookie operations
func basicCookieExample() {
	// Create cookie manager
	config := pkg.DefaultCookieConfig()
	cookieManager, err := pkg.NewCookieManager(config)
	if err != nil {
		log.Fatalf("Failed to create cookie manager: %v", err)
	}

	fmt.Println("✓ Cookie manager created with default configuration")
	fmt.Printf("  - Default Path: %s\n", config.DefaultPath)
	fmt.Printf("  - Default Secure: %v\n", config.DefaultSecure)
	fmt.Printf("  - Default HttpOnly: %v\n", config.DefaultHTTPOnly)
	fmt.Printf("  - Default MaxAge: %d seconds\n", config.DefaultMaxAge)

	// In a real handler, you would use the cookie manager like this:
	fmt.Println("\n  Example usage in handler:")
	fmt.Println("  cookie := &pkg.Cookie{")
	fmt.Println("    Name:  \"user_preference\",")
	fmt.Println("    Value: \"dark_mode\",")
	fmt.Println("  }")
	fmt.Println("  err := cookieManager.SetCookie(ctx, cookie)")

	_ = cookieManager // Suppress unused warning
}

// encryptedCookieExample demonstrates encrypted cookie operations
func encryptedCookieExample() {
	// Generate encryption key (32 bytes for AES-256)
	encryptionKey := make([]byte, 32)
	if _, err := rand.Read(encryptionKey); err != nil {
		log.Fatalf("Failed to generate encryption key: %v", err)
	}

	// Create cookie manager with encryption
	config := &pkg.CookieConfig{
		EncryptionKey:   encryptionKey,
		DefaultPath:     "/",
		DefaultSecure:   true,
		DefaultHTTPOnly: true,
		DefaultMaxAge:   86400, // 24 hours
	}

	cookieManager, err := pkg.NewCookieManager(config)
	if err != nil {
		log.Fatalf("Failed to create cookie manager: %v", err)
	}

	fmt.Println("✓ Cookie manager created with encryption enabled")

	// Demonstrate encryption/decryption
	sensitiveData := "user_session_token_abc123"

	encrypted, err := cookieManager.EncryptValue(sensitiveData)
	if err != nil {
		log.Fatalf("Failed to encrypt value: %v", err)
	}

	fmt.Printf("  - Original value: %s\n", sensitiveData)
	fmt.Printf("  - Encrypted value: %s...\n", encrypted[:40])

	decrypted, err := cookieManager.DecryptValue(encrypted)
	if err != nil {
		log.Fatalf("Failed to decrypt value: %v", err)
	}

	fmt.Printf("  - Decrypted value: %s\n", decrypted)
	fmt.Printf("  - Encryption verified: %v\n", sensitiveData == decrypted)

	fmt.Println("\n  Example usage in handler:")
	fmt.Println("  cookie := &pkg.Cookie{")
	fmt.Println("    Name:  \"session_token\",")
	fmt.Println("    Value: \"sensitive_session_data\",")
	fmt.Println("  }")
	fmt.Println("  err := cookieManager.SetEncryptedCookie(ctx, cookie)")
}

// headerManagementExample demonstrates header operations
func headerManagementExample() {
	// Create header manager
	headerManager := pkg.NewHeaderManager()

	fmt.Println("✓ Header manager created")

	fmt.Println("\n  Common header operations:")
	fmt.Println("  - Get request headers:")
	fmt.Println("    auth := headerManager.GetAuthorization(ctx)")
	fmt.Println("    userAgent := headerManager.GetUserAgent(ctx)")
	fmt.Println("    contentType := headerManager.GetContentType(ctx)")

	fmt.Println("\n  - Set response headers:")
	fmt.Println("    headerManager.SetContentType(ctx, \"application/json\")")
	fmt.Println("    headerManager.SetCacheControl(ctx, \"no-cache\")")
	fmt.Println("    headerManager.SetHeader(ctx, \"X-Custom-Header\", \"value\")")

	fmt.Println("\n  - Set multiple headers:")
	fmt.Println("    headers := map[string]string{")
	fmt.Println("      \"X-Request-ID\": \"req-123\",")
	fmt.Println("      \"X-API-Version\": \"v1\",")
	fmt.Println("    }")
	fmt.Println("    headerManager.SetHeaders(ctx, headers)")

	_ = headerManager // Suppress unused warning
}

// startServer starts an HTTP server demonstrating cookie and header management
func startServer() {
	// Generate encryption key
	encryptionKey := make([]byte, 32)
	if _, err := rand.Read(encryptionKey); err != nil {
		log.Fatalf("Failed to generate encryption key: %v", err)
	}

	// Create managers
	cookieConfig := &pkg.CookieConfig{
		EncryptionKey:   encryptionKey,
		DefaultPath:     "/",
		DefaultSecure:   false, // Set to false for local testing
		DefaultHTTPOnly: true,
		DefaultMaxAge:   3600,
	}

	cookieManager, err := pkg.NewCookieManager(cookieConfig)
	if err != nil {
		log.Fatalf("Failed to create cookie manager: %v", err)
	}

	headerManager := pkg.NewHeaderManager()

	// Handler: Set cookie and headers
	http.HandleFunc("/set", func(w http.ResponseWriter, r *http.Request) {
		// Create context
		request := &pkg.Request{
			Method: r.Method,
			URL:    r.URL,
			Header: r.Header,
		}

		resp := pkg.NewResponseWriter(w)
		ctx := pkg.NewContext(request, resp, r.Context())

		// Set a regular cookie
		cookie := &pkg.Cookie{
			Name:  "user_preference",
			Value: "dark_mode",
		}

		if err := cookieManager.SetCookie(ctx, cookie); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Set an encrypted cookie
		encryptedCookie := &pkg.Cookie{
			Name:  "session_token",
			Value: "sensitive_session_data_12345",
		}

		if err := cookieManager.SetEncryptedCookie(ctx, encryptedCookie); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Set response headers
		headerManager.SetContentType(ctx, "application/json")
		headerManager.SetHeader(ctx, "X-Request-ID", "req-123")
		headerManager.SetCacheControl(ctx, "no-cache")

		// Set multiple headers
		headers := map[string]string{
			"X-API-Version":   "v1",
			"X-Server":        "Rockstar",
			"X-Response-Time": "10ms",
		}
		headerManager.SetHeaders(ctx, headers)

		fmt.Fprintf(w, `{"message": "Cookies and headers set successfully"}`)
	})

	// Handler: Get cookies and headers
	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		// Create context
		request := &pkg.Request{
			Method: r.Method,
			URL:    r.URL,
			Header: r.Header,
		}

		resp := pkg.NewResponseWriter(w)
		ctx := pkg.NewContext(request, resp, r.Context())

		// Get regular cookie
		userPref, err := cookieManager.GetCookie(ctx, "user_preference")
		if err != nil {
			fmt.Fprintf(w, "Regular cookie not found: %v\n", err)
		} else {
			fmt.Fprintf(w, "User Preference: %s\n", userPref.Value)
		}

		// Get encrypted cookie
		sessionToken, err := cookieManager.GetEncryptedCookie(ctx, "session_token")
		if err != nil {
			fmt.Fprintf(w, "Encrypted cookie not found: %v\n", err)
		} else {
			fmt.Fprintf(w, "Session Token (decrypted): %s\n", sessionToken.Value)
		}

		// Get all cookies
		allCookies, err := cookieManager.GetAllCookies(ctx)
		if err != nil {
			fmt.Fprintf(w, "Failed to get all cookies: %v\n", err)
		} else {
			fmt.Fprintf(w, "\nAll Cookies (%d):\n", len(allCookies))
			for _, c := range allCookies {
				fmt.Fprintf(w, "  - %s: %s\n", c.Name, c.Value)
			}
		}

		// Get request headers
		fmt.Fprintf(w, "\nRequest Headers:\n")
		fmt.Fprintf(w, "  - User-Agent: %s\n", headerManager.GetUserAgent(ctx))
		fmt.Fprintf(w, "  - Content-Type: %s\n", headerManager.GetContentType(ctx))

		allHeaders := headerManager.GetAllHeaders(ctx)
		fmt.Fprintf(w, "\nAll Headers (%d):\n", len(allHeaders))
		for key, value := range allHeaders {
			fmt.Fprintf(w, "  - %s: %s\n", key, value)
		}

		// Set response headers
		headerManager.SetContentType(ctx, "text/plain")
		headerManager.SetHeader(ctx, "X-Response-ID", "resp-456")
	})

	// Handler: Delete cookie
	http.HandleFunc("/delete", func(w http.ResponseWriter, r *http.Request) {
		// Create context
		request := &pkg.Request{
			Method: r.Method,
			URL:    r.URL,
			Header: r.Header,
		}

		resp := pkg.NewResponseWriter(w)
		ctx := pkg.NewContext(request, resp, r.Context())

		// Delete cookie
		cookieName := r.URL.Query().Get("name")
		if cookieName == "" {
			cookieName = "user_preference"
		}

		if err := cookieManager.DeleteCookie(ctx, cookieName); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		headerManager.SetContentType(ctx, "application/json")
		fmt.Fprintf(w, `{"message": "Cookie '%s' deleted successfully"}`, cookieName)
	})

	// Start server
	fmt.Println("\n✓ Server started on http://localhost:8080")
	fmt.Println("\nAvailable endpoints:")
	fmt.Println("  - GET  /set    - Set cookies and headers")
	fmt.Println("  - GET  /get    - Get cookies and headers")
	fmt.Println("  - GET  /delete - Delete a cookie (use ?name=cookie_name)")
	fmt.Println("\nPress Ctrl+C to stop the server")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
