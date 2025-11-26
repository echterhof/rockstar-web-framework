//go:build ignore

package main

import (
	"fmt"
	"log"
	"net/http"
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

	// Cookie management
	router.GET("/api/cookies/set", setCookieHandler)
	router.GET("/api/cookies/get", getCookieHandler)
	router.GET("/api/cookies/get-all", getAllCookiesHandler)
	router.GET("/api/cookies/delete", deleteCookieHandler)

	// Secure cookie patterns
	router.GET("/api/cookies/secure", setSecureCookieHandler)
	router.GET("/api/cookies/httponly", setHttpOnlyCookieHandler)
	router.GET("/api/cookies/samesite", setSameSiteCookieHandler)

	// Header manipulation
	router.GET("/api/headers/set", setHeaderHandler)
	router.GET("/api/headers/get", getHeaderHandler)
	router.GET("/api/headers/get-all", getAllHeadersHandler)
	router.GET("/api/headers/delete", deleteHeaderHandler)

	// Custom header handling
	router.GET("/api/headers/custom", customHeadersHandler)
	router.GET("/api/headers/content-type", contentTypeHandler)
	router.GET("/api/headers/cache-control", cacheControlHandler)
	router.GET("/api/headers/cors", corsHeadersHandler)

	// Combined cookie and header operations
	router.GET("/api/combined/auth", authenticationHandler)
	router.GET("/api/combined/tracking", trackingHandler)

	// ============================================================================
	// Server Startup
	// ============================================================================
	fmt.Println("🎸 Rockstar Web Framework - Cookie & Header Example")
	fmt.Println("=" + "==========================================================")
	fmt.Println()
	fmt.Println("Server listening on: http://localhost:8080")
	fmt.Println()
	fmt.Println("Try these commands:")
	fmt.Println("  # Cookie management")
	fmt.Println("  curl -v -c cookies.txt http://localhost:8080/api/cookies/set")
	fmt.Println("  curl -v -b cookies.txt http://localhost:8080/api/cookies/get")
	fmt.Println("  curl -v -b cookies.txt http://localhost:8080/api/cookies/get-all")
	fmt.Println("  curl -v -b cookies.txt http://localhost:8080/api/cookies/delete")
	fmt.Println()
	fmt.Println("  # Secure cookie patterns")
	fmt.Println("  curl -v http://localhost:8080/api/cookies/secure")
	fmt.Println("  curl -v http://localhost:8080/api/cookies/httponly")
	fmt.Println("  curl -v http://localhost:8080/api/cookies/samesite")
	fmt.Println()
	fmt.Println("  # Header manipulation")
	fmt.Println("  curl -v http://localhost:8080/api/headers/set")
	fmt.Println("  curl -v -H \"User-Agent: MyApp/1.0\" http://localhost:8080/api/headers/get")
	fmt.Println("  curl -v http://localhost:8080/api/headers/get-all")
	fmt.Println("  curl -v http://localhost:8080/api/headers/delete")
	fmt.Println()
	fmt.Println("  # Custom header handling")
	fmt.Println("  curl -v http://localhost:8080/api/headers/custom")
	fmt.Println("  curl -v http://localhost:8080/api/headers/content-type")
	fmt.Println("  curl -v http://localhost:8080/api/headers/cache-control")
	fmt.Println("  curl -v http://localhost:8080/api/headers/cors")
	fmt.Println()
	fmt.Println("  # Combined operations")
	fmt.Println("  curl -v http://localhost:8080/api/combined/auth")
	fmt.Println("  curl -v http://localhost:8080/api/combined/tracking")
	fmt.Println()
	fmt.Println("=" + "==========================================================")
	fmt.Println()

	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// ============================================================================
// Handler Functions - Cookie Management
// ============================================================================

// setCookieHandler demonstrates setting various types of cookies
func setCookieHandler(ctx pkg.Context) error {
	// Set a simple cookie
	simpleCookie := &pkg.Cookie{
		Name:   "simple_cookie",
		Value:  "simple_value",
		Path:   "/",
		MaxAge: 3600, // 1 hour
	}
	ctx.SetCookie(simpleCookie)

	// Set a cookie with expiration time
	expiringCookie := &pkg.Cookie{
		Name:    "expiring_cookie",
		Value:   "expires_soon",
		Path:    "/",
		Expires: time.Now().Add(30 * time.Minute),
	}
	ctx.SetCookie(expiringCookie)

	// Set a persistent cookie
	persistentCookie := &pkg.Cookie{
		Name:   "persistent_cookie",
		Value:  "long_lived",
		Path:   "/",
		MaxAge: 86400 * 30, // 30 days
	}
	ctx.SetCookie(persistentCookie)

	return ctx.JSON(200, map[string]interface{}{
		"message": "Cookies set successfully",
		"cookies": []string{
			"simple_cookie (1 hour)",
			"expiring_cookie (30 minutes)",
			"persistent_cookie (30 days)",
		},
		"note": "Check Set-Cookie headers with curl -v",
	})
}

// getCookieHandler demonstrates getting a specific cookie
func getCookieHandler(ctx pkg.Context) error {
	// Try to get a cookie
	cookie, err := ctx.GetCookie("simple_cookie")
	if err != nil {
		return ctx.JSON(404, map[string]interface{}{
			"error":   "Cookie not found",
			"message": "The cookie 'simple_cookie' was not sent by the client",
		})
	}

	return ctx.JSON(200, map[string]interface{}{
		"message": "Cookie retrieved successfully",
		"cookie": map[string]interface{}{
			"name":  cookie.Name,
			"value": cookie.Value,
		},
	})
}

// getAllCookiesHandler demonstrates getting all cookies from the request
func getAllCookiesHandler(ctx pkg.Context) error {
	// Get all cookies from the request by parsing the Cookie header
	cookieHeader := ctx.GetHeader("Cookie")
	if cookieHeader == "" {
		return ctx.JSON(200, map[string]interface{}{
			"message": "No cookies found",
			"count":   0,
			"cookies": map[string]string{},
		})
	}

	// Parse cookies manually from the Cookie header
	// Format: "name1=value1; name2=value2"
	cookieMap := make(map[string]string)
	cookies := parseCookieHeader(cookieHeader)
	for name, value := range cookies {
		cookieMap[name] = value
	}

	return ctx.JSON(200, map[string]interface{}{
		"message": "All cookies retrieved",
		"count":   len(cookieMap),
		"cookies": cookieMap,
	})
}

// parseCookieHeader parses the Cookie header into a map
func parseCookieHeader(header string) map[string]string {
	cookies := make(map[string]string)
	pairs := splitCookies(header)
	for _, pair := range pairs {
		name, value := splitCookiePair(pair)
		if name != "" {
			cookies[name] = value
		}
	}
	return cookies
}

// splitCookies splits the cookie header by semicolons
func splitCookies(header string) []string {
	var pairs []string
	current := ""
	for _, char := range header {
		if char == ';' {
			if current != "" {
				pairs = append(pairs, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	if current != "" {
		pairs = append(pairs, current)
	}
	return pairs
}

// splitCookiePair splits a cookie pair into name and value
func splitCookiePair(pair string) (string, string) {
	// Trim spaces
	pair = trimSpace(pair)

	// Find the equals sign
	for i, char := range pair {
		if char == '=' {
			return pair[:i], pair[i+1:]
		}
	}
	return pair, ""
}

// trimSpace removes leading and trailing spaces
func trimSpace(s string) string {
	start := 0
	end := len(s)

	for start < end && s[start] == ' ' {
		start++
	}
	for end > start && s[end-1] == ' ' {
		end--
	}

	return s[start:end]
}

// deleteCookieHandler demonstrates deleting a cookie
func deleteCookieHandler(ctx pkg.Context) error {
	// Delete a cookie by setting it to expire immediately
	deleteCookie := &pkg.Cookie{
		Name:   "simple_cookie",
		Value:  "",
		Path:   "/",
		MaxAge: -1, // Delete immediately
	}
	ctx.SetCookie(deleteCookie)

	return ctx.JSON(200, map[string]interface{}{
		"message": "Cookie deleted successfully",
		"deleted": "simple_cookie",
		"note":    "Cookie will be removed from the client",
	})
}

// ============================================================================
// Handler Functions - Secure Cookie Patterns
// ============================================================================

// setSecureCookieHandler demonstrates setting a secure cookie (HTTPS only)
func setSecureCookieHandler(ctx pkg.Context) error {
	secureCookie := &pkg.Cookie{
		Name:   "secure_cookie",
		Value:  "secure_value",
		Path:   "/",
		MaxAge: 3600,
		Secure: true, // Only sent over HTTPS
	}
	ctx.SetCookie(secureCookie)

	return ctx.JSON(200, map[string]interface{}{
		"message": "Secure cookie set successfully",
		"cookie":  "secure_cookie",
		"note":    "This cookie will only be sent over HTTPS connections",
		"warning": "In development without HTTPS, the browser may not send this cookie",
	})
}

// setHttpOnlyCookieHandler demonstrates setting an HttpOnly cookie
func setHttpOnlyCookieHandler(ctx pkg.Context) error {
	httpOnlyCookie := &pkg.Cookie{
		Name:     "httponly_cookie",
		Value:    "protected_value",
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true, // Not accessible via JavaScript
	}
	ctx.SetCookie(httpOnlyCookie)

	return ctx.JSON(200, map[string]interface{}{
		"message": "HttpOnly cookie set successfully",
		"cookie":  "httponly_cookie",
		"note":    "This cookie cannot be accessed by JavaScript (document.cookie)",
		"security": map[string]string{
			"protection": "XSS attacks",
			"benefit":    "Prevents client-side script access",
		},
	})
}

// setSameSiteCookieHandler demonstrates setting cookies with SameSite attribute
func setSameSiteCookieHandler(ctx pkg.Context) error {
	// SameSite=Strict - Most restrictive
	strictCookie := &pkg.Cookie{
		Name:     "samesite_strict",
		Value:    "strict_value",
		Path:     "/",
		MaxAge:   3600,
		SameSite: http.SameSiteStrictMode,
	}
	ctx.SetCookie(strictCookie)

	// SameSite=Lax - Balanced (default)
	laxCookie := &pkg.Cookie{
		Name:     "samesite_lax",
		Value:    "lax_value",
		Path:     "/",
		MaxAge:   3600,
		SameSite: http.SameSiteLaxMode,
	}
	ctx.SetCookie(laxCookie)

	// SameSite=None - Least restrictive (requires Secure)
	noneCookie := &pkg.Cookie{
		Name:     "samesite_none",
		Value:    "none_value",
		Path:     "/",
		MaxAge:   3600,
		Secure:   true, // Required with SameSite=None
		SameSite: http.SameSiteNoneMode,
	}
	ctx.SetCookie(noneCookie)

	return ctx.JSON(200, map[string]interface{}{
		"message": "SameSite cookies set successfully",
		"cookies": map[string]string{
			"samesite_strict": "Not sent with cross-site requests",
			"samesite_lax":    "Sent with top-level navigation",
			"samesite_none":   "Sent with all requests (requires Secure)",
		},
		"security": "SameSite helps prevent CSRF attacks",
	})
}

// ============================================================================
// Handler Functions - Header Manipulation
// ============================================================================

// setHeaderHandler demonstrates setting response headers
func setHeaderHandler(ctx pkg.Context) error {
	// Set various response headers
	ctx.SetHeader("X-Custom-Header", "CustomValue")
	ctx.SetHeader("X-Request-ID", "req-12345")
	ctx.SetHeader("X-API-Version", "1.0")
	ctx.SetHeader("X-Response-Time", time.Now().Format(time.RFC3339))

	return ctx.JSON(200, map[string]interface{}{
		"message": "Response headers set successfully",
		"headers": map[string]string{
			"X-Custom-Header": "CustomValue",
			"X-Request-ID":    "req-12345",
			"X-API-Version":   "1.0",
			"X-Response-Time": "current timestamp",
		},
		"note": "Check response headers with curl -v",
	})
}

// getHeaderHandler demonstrates getting request headers
func getHeaderHandler(ctx pkg.Context) error {
	// Get specific headers from the request
	userAgent := ctx.GetHeader("User-Agent")
	accept := ctx.GetHeader("Accept")
	contentType := ctx.GetHeader("Content-Type")
	authorization := ctx.GetHeader("Authorization")

	return ctx.JSON(200, map[string]interface{}{
		"message": "Request headers retrieved",
		"headers": map[string]string{
			"User-Agent":    userAgent,
			"Accept":        accept,
			"Content-Type":  contentType,
			"Authorization": authorization,
		},
	})
}

// getAllHeadersHandler demonstrates getting all request headers
func getAllHeadersHandler(ctx pkg.Context) error {
	// Get all headers from the request
	headers := ctx.Headers()

	return ctx.JSON(200, map[string]interface{}{
		"message": "All request headers retrieved",
		"count":   len(headers),
		"headers": headers,
	})
}

// deleteHeaderHandler demonstrates deleting a response header
func deleteHeaderHandler(ctx pkg.Context) error {
	// Set a header first
	ctx.SetHeader("X-Temporary-Header", "temporary_value")

	// Get the response to delete the header
	resp := ctx.Response()
	if resp != nil {
		resp.Header().Del("X-Temporary-Header")
	}

	return ctx.JSON(200, map[string]interface{}{
		"message": "Header deleted successfully",
		"deleted": "X-Temporary-Header",
		"note":    "The header was set and then deleted before sending the response",
	})
}

// ============================================================================
// Handler Functions - Custom Header Handling
// ============================================================================

// customHeadersHandler demonstrates setting multiple custom headers
func customHeadersHandler(ctx pkg.Context) error {
	// Set multiple custom headers for API versioning and tracking
	ctx.SetHeader("X-API-Version", "2.0")
	ctx.SetHeader("X-RateLimit-Limit", "1000")
	ctx.SetHeader("X-RateLimit-Remaining", "999")
	ctx.SetHeader("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(1*time.Hour).Unix()))
	ctx.SetHeader("X-Server-Region", "us-east-1")
	ctx.SetHeader("X-Request-ID", ctx.Request().ID)

	return ctx.JSON(200, map[string]interface{}{
		"message": "Custom headers set for API response",
		"headers": map[string]string{
			"X-API-Version":   "API version information",
			"X-RateLimit-*":   "Rate limiting information",
			"X-Server-Region": "Server location",
			"X-Request-ID":    "Request tracking",
		},
	})
}

// contentTypeHandler demonstrates setting Content-Type headers
func contentTypeHandler(ctx pkg.Context) error {
	// The Content-Type is automatically set by ctx.JSON(), but we can override it
	ctx.SetHeader("Content-Type", "application/json; charset=utf-8")

	return ctx.JSON(200, map[string]interface{}{
		"message":      "Content-Type header demonstration",
		"content_type": "application/json; charset=utf-8",
		"note":         "Content-Type is usually set automatically by response methods",
	})
}

// cacheControlHandler demonstrates setting Cache-Control headers
func cacheControlHandler(ctx pkg.Context) error {
	// Set cache control headers for different scenarios
	ctx.SetHeader("Cache-Control", "public, max-age=3600")
	ctx.SetHeader("ETag", `"abc123"`)
	ctx.SetHeader("Last-Modified", time.Now().Add(-1*time.Hour).Format(http.TimeFormat))

	return ctx.JSON(200, map[string]interface{}{
		"message": "Cache control headers set",
		"caching": map[string]string{
			"Cache-Control": "public, max-age=3600 (1 hour)",
			"ETag":          "Entity tag for cache validation",
			"Last-Modified": "Last modification time",
		},
		"note": "These headers help browsers cache the response",
	})
}

// corsHeadersHandler demonstrates setting CORS headers
func corsHeadersHandler(ctx pkg.Context) error {
	// Set CORS headers for cross-origin requests
	ctx.SetHeader("Access-Control-Allow-Origin", "*")
	ctx.SetHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	ctx.SetHeader("Access-Control-Allow-Headers", "Content-Type, Authorization")
	ctx.SetHeader("Access-Control-Max-Age", "86400")
	ctx.SetHeader("Access-Control-Allow-Credentials", "true")

	return ctx.JSON(200, map[string]interface{}{
		"message": "CORS headers set successfully",
		"cors": map[string]string{
			"Access-Control-Allow-Origin":      "Allowed origins",
			"Access-Control-Allow-Methods":     "Allowed HTTP methods",
			"Access-Control-Allow-Headers":     "Allowed request headers",
			"Access-Control-Max-Age":           "Preflight cache duration",
			"Access-Control-Allow-Credentials": "Allow credentials",
		},
		"note": "These headers enable cross-origin resource sharing",
	})
}

// ============================================================================
// Handler Functions - Combined Cookie and Header Operations
// ============================================================================

// authenticationHandler demonstrates combined cookie and header usage for authentication
func authenticationHandler(ctx pkg.Context) error {
	// Set authentication cookie
	authCookie := &pkg.Cookie{
		Name:     "auth_token",
		Value:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
	ctx.SetCookie(authCookie)

	// Set authentication headers
	ctx.SetHeader("X-Auth-Status", "authenticated")
	ctx.SetHeader("X-User-ID", "user-123")
	ctx.SetHeader("X-Session-ID", "session-456")
	ctx.SetHeader("X-Token-Expires", time.Now().Add(1*time.Hour).Format(time.RFC3339))

	return ctx.JSON(200, map[string]interface{}{
		"message": "Authentication successful",
		"auth": map[string]interface{}{
			"cookie": map[string]string{
				"name":     "auth_token",
				"httponly": "true",
				"secure":   "true",
				"samesite": "Strict",
			},
			"headers": map[string]string{
				"X-Auth-Status":   "authenticated",
				"X-User-ID":       "user-123",
				"X-Session-ID":    "session-456",
				"X-Token-Expires": "1 hour from now",
			},
		},
		"security": "Cookie and headers work together for secure authentication",
	})
}

// trackingHandler demonstrates combined cookie and header usage for tracking
func trackingHandler(ctx pkg.Context) error {
	// Check for existing tracking cookie
	trackingCookie, err := ctx.GetCookie("tracking_id")
	var trackingID string
	if err != nil {
		// Generate new tracking ID
		trackingID = fmt.Sprintf("track-%d", time.Now().UnixNano())
		newCookie := &pkg.Cookie{
			Name:   "tracking_id",
			Value:  trackingID,
			Path:   "/",
			MaxAge: 86400 * 365, // 1 year
		}
		ctx.SetCookie(newCookie)
	} else {
		trackingID = trackingCookie.Value
	}

	// Set tracking headers
	ctx.SetHeader("X-Tracking-ID", trackingID)
	ctx.SetHeader("X-Visit-Count", "1")
	ctx.SetHeader("X-Last-Visit", time.Now().Format(time.RFC3339))
	ctx.SetHeader("X-User-Agent", ctx.GetHeader("User-Agent"))

	return ctx.JSON(200, map[string]interface{}{
		"message":     "Tracking information",
		"tracking_id": trackingID,
		"cookie":      "tracking_id cookie set for 1 year",
		"headers": map[string]string{
			"X-Tracking-ID": "Unique visitor identifier",
			"X-Visit-Count": "Number of visits",
			"X-Last-Visit":  "Last visit timestamp",
		},
		"note": "Tracking uses both cookies for persistence and headers for metadata",
	})
}
