package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"strconv"
	"syscall"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

// ============================================================================
// Full Featured Application Example
// ============================================================================
// This example demonstrates all major features of the Rockstar Web Framework:
// - Database integration
// - Session management
// - Authentication and authorization
// - Caching
// - Internationalization
// - WebSocket support
// - Static file serving
// - Multiple API styles (REST, GraphQL)
// - Multi-tenancy
// - Monitoring and profiling
// - Graceful shutdown

func main() {
	// ============================================================================
	// Configuration Setup
	// ============================================================================
	// Create comprehensive framework configuration
	config := pkg.FrameworkConfig{
		// Server configuration with advanced features
		ServerConfig: pkg.ServerConfig{
			ReadTimeout:     15 * time.Second,
			WriteTimeout:    15 * time.Second,
			IdleTimeout:     120 * time.Second,
			MaxHeaderBytes:  2 << 20, // 2 MB
			EnableHTTP1:     true,
			EnableHTTP2:     true,
			EnableMetrics:   true,
			MetricsPath:     "/metrics",
			EnablePprof:     true,
			PprofPath:       "/debug/pprof",
			ShutdownTimeout: 30 * time.Second,
			// Multi-tenancy configuration with host-specific settings
			HostConfigs: map[string]*pkg.HostConfig{
				"api.example.com": {
					Hostname: "api.example.com",
					TenantID: "tenant-1",
					RateLimits: &pkg.RateLimitConfig{
						Enabled:           true,
						RequestsPerSecond: 100,
						BurstSize:         20,
						Storage:           "database",
					},
				},
				"admin.example.com": {
					Hostname: "admin.example.com",
					TenantID: "tenant-admin",
					RateLimits: &pkg.RateLimitConfig{
						Enabled:           true,
						RequestsPerSecond: 50,
						BurstSize:         10,
						Storage:           "database",
					},
				},
			},
		},
		// Database configuration - PostgreSQL for production
		DatabaseConfig: pkg.DatabaseConfig{
			Driver:          "postgres",
			Host:            "localhost",
			Port:            5432,
			Database:        "rockstar_db",
			Username:        "rockstar_user",
			Password:        "rockstar_pass",
			MaxOpenConns:    25,
			MaxIdleConns:    5,
			ConnMaxLifetime: 5 * time.Minute,
		},
		// Cache configuration - in-memory cache with larger size
		CacheConfig: pkg.CacheConfig{
			Type:       "memory",
			MaxSize:    500 * 1024 * 1024, // 500 MB
			DefaultTTL: 10 * time.Minute,
		},
		// Session configuration - database-backed sessions
		SessionConfig: pkg.SessionConfig{
			StorageType:     pkg.SessionStorageDatabase,
			CookieName:      "rockstar_session_id",
			SessionLifetime: 24 * time.Hour,
			CookieSecure:    true, // Enable in production with HTTPS
			CookieHTTPOnly:  true, // Prevent JavaScript access
			CookieSameSite:  "Strict",
			EncryptionKey:   []byte("12345678901234567890123456789012"), // 32 bytes for AES-256
			CleanupInterval: 10 * time.Minute,
		},
		// Configuration files - load from YAML files
		ConfigFiles: []string{
			"config.yaml",
			"config.local.yaml",
		},
		// Internationalization configuration
		I18nConfig: pkg.I18nConfig{
			DefaultLocale:     "en",
			LocalesDir:        "./locales",
			SupportedLocales:  []string{"en", "de", "fr", "es"},
			FallbackToDefault: true,
		},
		// Security configuration
		SecurityConfig: pkg.SecurityConfig{
			XFrameOptions:    "SAMEORIGIN",
			EnableXSSProtect: true,
			EnableCSRF:       true,
			MaxRequestSize:   20 * 1024 * 1024, // 20 MB
			RequestTimeout:   60 * time.Second,
			AllowedOrigins:   []string{"https://example.com"},
			EncryptionKey:    "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			JWTSecret:        "your-secret-jwt-key",
			CSRFTokenExpiry:  24 * time.Hour,
		},
		// Monitoring configuration
		MonitoringConfig: pkg.MonitoringConfig{
			EnableMetrics:        true,
			MetricsPath:          "/metrics",
			EnablePprof:          true,
			PprofPath:            "/debug/pprof",
			EnableSNMP:           true,
			SNMPPort:             161,
			EnableOptimization:   true,
			OptimizationInterval: 10 * time.Second,
		},
		// Proxy configuration - for load balancing and circuit breaking
		ProxyConfig: pkg.ProxyConfig{
			LoadBalancerType:         "round_robin",
			CircuitBreakerEnabled:    true,
			CircuitBreakerThreshold:  5,
			CircuitBreakerTimeout:    30 * time.Second,
			MaxRetries:               3,
			RetryDelay:               1 * time.Second,
			RetryBackoff:             true,
			HealthCheckEnabled:       true,
			HealthCheckInterval:      30 * time.Second,
			HealthCheckTimeout:       5 * time.Second,
			MaxConnectionsPerBackend: 100,
			ConnectionTimeout:        10 * time.Second,
			IdleConnTimeout:          90 * time.Second,
		},
	}

	// ============================================================================
	// Framework Initialization
	// ============================================================================
	app, err := pkg.New(config)
	if err != nil {
		log.Fatalf("Failed to create framework: %v", err)
	}

	// Setup lifecycle hooks
	setupLifecycleHooks(app)

	// Setup global middleware
	setupGlobalMiddleware(app)

	// Setup routes
	setupRoutes(app)

	// Setup graceful shutdown
	setupGracefulShutdown(app)

	// ============================================================================
	// Server Startup
	// ============================================================================
	fmt.Println("🎸 Rockstar Web Framework - Full Featured Application")
	fmt.Println("=" + "==========================================================")
	fmt.Println()
	fmt.Println("Server starting on: http://localhost:8080")
	fmt.Println()
	fmt.Println("Available endpoints:")
	fmt.Println("  http://localhost:8080/              - Home")
	fmt.Println("  http://localhost:8080/health        - Health check")
	fmt.Println("  http://localhost:8080/ready         - Readiness check")
	fmt.Println("  http://localhost:8080/metrics       - Prometheus metrics")
	fmt.Println("  http://localhost:8080/debug/pprof   - Profiling")
	fmt.Println()
	fmt.Println("API endpoints:")
	fmt.Println("  http://localhost:8080/api/v1/status - API status")
	fmt.Println("  http://localhost:8080/api/v1/profile - User profile (auth required)")
	fmt.Println("  http://localhost:8080/ws            - WebSocket endpoint")
	fmt.Println()
	fmt.Println("Admin endpoints (auth required):")
	fmt.Println("  http://localhost:8080/admin/users   - User management")
	fmt.Println()
	fmt.Println("Multi-tenancy (use Host header):")
	fmt.Println("  curl -H \"Host: api.example.com\" http://localhost:8080/")
	fmt.Println("  curl -H \"Host: admin.example.com\" http://localhost:8080/")
	fmt.Println()
	fmt.Println("=" + "==========================================================")
	fmt.Println()

	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// ============================================================================
// Lifecycle Hooks
// ============================================================================

func setupLifecycleHooks(app *pkg.Framework) {
	// Startup hooks - executed in order during server startup
	app.RegisterStartupHook(func(ctx context.Context) error {
		fmt.Println("🚀 Initializing database connections...")
		return nil
	})

	app.RegisterStartupHook(func(ctx context.Context) error {
		fmt.Println("🚀 Loading configuration...")
		return nil
	})

	app.RegisterStartupHook(func(ctx context.Context) error {
		fmt.Println("🚀 Starting background workers...")
		return nil
	})

	// Shutdown hooks - executed in order during graceful shutdown
	app.RegisterShutdownHook(func(ctx context.Context) error {
		fmt.Println("👋 Stopping background workers...")
		return nil
	})

	app.RegisterShutdownHook(func(ctx context.Context) error {
		fmt.Println("👋 Flushing metrics...")
		return nil
	})

	app.RegisterShutdownHook(func(ctx context.Context) error {
		fmt.Println("👋 Closing connections...")
		return nil
	})
}

// ============================================================================
// Global Middleware
// ============================================================================

func setupGlobalMiddleware(app *pkg.Framework) {
	// Request ID middleware - adds unique ID to each request
	app.Use(func(ctx pkg.Context, next pkg.HandlerFunc) error {
		requestID := generateRequestID()
		ctx.SetHeader("X-Request-ID", requestID)
		return next(ctx)
	})

	// Logging middleware - logs all requests
	app.Use(func(ctx pkg.Context, next pkg.HandlerFunc) error {
		start := time.Now()
		logger := ctx.Logger()

		if logger != nil {
			logger.Info("Request started",
				"method", ctx.Request().Method,
				"path", ctx.Request().URL.Path,
				"host", ctx.Request().Host,
			)
		}

		err := next(ctx)

		duration := time.Since(start)
		if logger != nil {
			logger.Info("Request completed",
				"duration", duration,
				"status", ctx.Response().Status(),
			)
		}

		return err
	})

	// Metrics middleware - records request metrics
	app.Use(func(ctx pkg.Context, next pkg.HandlerFunc) error {
		start := time.Now()

		err := next(ctx)

		duration := time.Since(start)
		metrics := ctx.Metrics()
		if metrics != nil {
			statusCode := ctx.Response().Status()
			metrics.RecordRequest(ctx, duration, statusCode)
		}

		return err
	})

	// Recovery middleware - recovers from panics
	app.Use(func(ctx pkg.Context, next pkg.HandlerFunc) error {
		defer func() {
			if r := recover(); r != nil {
				logger := ctx.Logger()
				if logger != nil {
					logger.Error("Panic recovered", "panic", r)
				} else {
					fmt.Printf("Panic recovered: %v\n", r)
				}
				ctx.JSON(500, map[string]interface{}{
					"error": "Internal server error",
				})
			}
		}()
		return next(ctx)
	})
}

// ============================================================================
// Route Setup
// ============================================================================

func setupRoutes(app *pkg.Framework) {
	router := app.Router()

	// ========================================
	// Public Routes
	// ========================================
	router.GET("/", homeHandler)
	router.GET("/health", healthHandler)
	router.GET("/ready", readinessHandler)

	// ========================================
	// Metrics and Profiling Endpoints
	// ========================================
	router.GET("/metrics", metricsHandler)
	router.GET("/debug/pprof", pprofIndexHandler)
	router.GET("/debug/pprof/cmdline", pprofCmdlineHandler)
	router.GET("/debug/pprof/profile", pprofProfileHandler)
	router.GET("/debug/pprof/symbol", pprofSymbolHandler)
	router.GET("/debug/pprof/trace", pprofTraceHandler)
	router.GET("/debug/pprof/heap", pprofHeapHandler)
	router.GET("/debug/pprof/goroutine", pprofGoroutineHandler)
	router.GET("/debug/pprof/threadcreate", pprofThreadCreateHandler)
	router.GET("/debug/pprof/block", pprofBlockHandler)
	router.GET("/debug/pprof/mutex", pprofMutexHandler)

	// ========================================
	// API v1 Routes
	// ========================================
	v1 := router.Group("/api/v1")
	v1.GET("/status", apiStatusHandler)

	// ========================================
	// Authenticated API Routes
	// ========================================
	authAPI := router.Group("/api/v1", authenticationMiddleware)
	authAPI.GET("/profile", profileHandler)
	authAPI.PUT("/profile", updateProfileHandler)
	authAPI.POST("/logout", logoutHandler)

	// ========================================
	// Admin Routes (requires authentication and authorization)
	// ========================================
	admin := router.Group("/admin", authenticationMiddleware, adminAuthorizationMiddleware)
	admin.GET("/users", listUsersHandler)
	admin.POST("/users", createUserHandler)
	admin.GET("/users/:id", getUserHandler)
	admin.PUT("/users/:id", updateUserHandler)
	admin.DELETE("/users/:id", deleteUserHandler)

	// ========================================
	// REST API Example
	// ========================================
	router.Group("/api/rest").GET("/products", restProductsHandler)

	// ========================================
	// WebSocket Example
	// ========================================
	router.WebSocket("/ws", websocketHandler)

	// ========================================
	// Static Files (uncomment to enable)
	// ========================================
	// router.Static("/static", pkg.NewOSFileSystem("./public"))

	// ========================================
	// Multi-Tenancy Routes (host-specific)
	// ========================================
	apiHost := router.Host("api.example.com")
	apiHost.GET("/", apiHomeHandler)

	adminHost := router.Host("admin.example.com")
	adminHost.GET("/", adminHomeHandler)
}

// ============================================================================
// Graceful Shutdown
// ============================================================================

func setupGracefulShutdown(app *pkg.Framework) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n🛑 Shutdown signal received, gracefully shutting down...")

		if err := app.Shutdown(30 * time.Second); err != nil {
			log.Printf("Error during shutdown: %v", err)
		}

		os.Exit(0)
	}()
}

// ============================================================================
// Middleware Implementations
// ============================================================================

// authenticationMiddleware checks for valid authentication
func authenticationMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
	token := ctx.GetHeader("Authorization")
	if token == "" {
		return ctx.JSON(401, map[string]interface{}{
			"error": "Authentication required",
		})
	}

	// In production, validate token using security manager:
	// user, err := ctx.Security().AuthenticateJWT(token)
	// if err != nil {
	//     return ctx.JSON(401, map[string]interface{}{"error": "Invalid token"})
	// }

	return next(ctx)
}

// adminAuthorizationMiddleware checks for admin permissions
func adminAuthorizationMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
	user := ctx.User()
	if user == nil {
		return ctx.JSON(401, map[string]interface{}{
			"error": "Authentication required",
		})
	}

	// Check if user has admin role
	if !ctx.IsAuthorized("admin", "access") {
		return ctx.JSON(403, map[string]interface{}{
			"error": "Insufficient permissions",
		})
	}

	return next(ctx)
}

// ============================================================================
// Handler Implementations
// ============================================================================

// homeHandler handles the root endpoint
func homeHandler(ctx pkg.Context) error {
	return ctx.JSON(200, map[string]interface{}{
		"message": "Welcome to Rockstar Web Framework! 🎸",
		"version": "1.0.0",
		"features": []string{
			"Multi-protocol API support",
			"Multi-tenancy",
			"Built-in security",
			"Session management",
			"Internationalization",
			"Monitoring & profiling",
			"Forward proxy",
			"WebSocket support",
			"Database integration",
			"Caching",
		},
	})
}

// healthHandler provides basic health check
func healthHandler(ctx pkg.Context) error {
	return ctx.JSON(200, map[string]interface{}{
		"status": "healthy",
		"time":   time.Now().Unix(),
	})
}

// readinessHandler checks if all services are ready
func readinessHandler(ctx pkg.Context) error {
	// Check if all services are ready
	// In production, check actual service health:
	// dbHealthy := ctx.DB().Ping() == nil
	// cacheHealthy := ctx.Cache().Ping() == nil
	dbHealthy := true
	cacheHealthy := true

	if !dbHealthy || !cacheHealthy {
		return ctx.JSON(503, map[string]interface{}{
			"status": "not ready",
			"checks": map[string]bool{
				"database": dbHealthy,
				"cache":    cacheHealthy,
			},
		})
	}

	return ctx.JSON(200, map[string]interface{}{
		"status": "ready",
	})
}

// apiStatusHandler returns API status
func apiStatusHandler(ctx pkg.Context) error {
	return ctx.JSON(200, map[string]interface{}{
		"api_version": "v1",
		"status":      "operational",
	})
}

// profileHandler returns user profile
func profileHandler(ctx pkg.Context) error {
	user := ctx.User()
	return ctx.JSON(200, map[string]interface{}{
		"user": user,
	})
}

// updateProfileHandler updates user profile
func updateProfileHandler(ctx pkg.Context) error {
	// In production, parse request body and update database
	return ctx.JSON(200, map[string]interface{}{
		"message": "Profile updated successfully",
	})
}

// logoutHandler handles user logout
func logoutHandler(ctx pkg.Context) error {
	// In production, destroy session:
	// session, err := ctx.Session().GetSessionFromCookie(ctx)
	// if err == nil && session != nil {
	//     ctx.Session().Destroy(ctx, session.ID)
	// }

	return ctx.JSON(200, map[string]interface{}{
		"message": "Logged out successfully",
	})
}

// listUsersHandler lists all users (admin only)
func listUsersHandler(ctx pkg.Context) error {
	// In production, fetch from database:
	// users, err := ctx.DB().Query("SELECT * FROM users")
	return ctx.JSON(200, map[string]interface{}{
		"users": []map[string]interface{}{},
		"total": 0,
	})
}

// createUserHandler creates a new user (admin only)
func createUserHandler(ctx pkg.Context) error {
	// In production, parse body and save to database
	return ctx.JSON(201, map[string]interface{}{
		"message": "User created successfully",
	})
}

// getUserHandler gets a specific user (admin only)
func getUserHandler(ctx pkg.Context) error {
	userID := ctx.Params()["id"]
	// In production, fetch from database
	return ctx.JSON(200, map[string]interface{}{
		"id": userID,
	})
}

// updateUserHandler updates a user (admin only)
func updateUserHandler(ctx pkg.Context) error {
	userID := ctx.Params()["id"]
	// In production, parse body and update database
	return ctx.JSON(200, map[string]interface{}{
		"message": fmt.Sprintf("User %s updated successfully", userID),
	})
}

// deleteUserHandler deletes a user (admin only)
func deleteUserHandler(ctx pkg.Context) error {
	userID := ctx.Params()["id"]
	// In production, delete from database
	return ctx.JSON(200, map[string]interface{}{
		"message": fmt.Sprintf("User %s deleted successfully", userID),
	})
}

// restProductsHandler demonstrates REST API
func restProductsHandler(ctx pkg.Context) error {
	// In production, fetch from database with pagination
	return ctx.JSON(200, map[string]interface{}{
		"products": []map[string]interface{}{},
	})
}

// websocketHandler handles WebSocket connections
func websocketHandler(ctx pkg.Context, conn pkg.WebSocketConnection) error {
	fmt.Println("WebSocket connection established")

	// Echo server - reads messages and sends them back
	for {
		messageType, data, err := conn.ReadMessage()
		if err != nil {
			fmt.Printf("WebSocket read error: %v\n", err)
			return err
		}

		fmt.Printf("Received WebSocket message: %s\n", string(data))

		// Echo message back
		if err := conn.WriteMessage(messageType, data); err != nil {
			fmt.Printf("WebSocket write error: %v\n", err)
			return err
		}
	}
}

// apiHomeHandler handles API tenant home
func apiHomeHandler(ctx pkg.Context) error {
	return ctx.JSON(200, map[string]interface{}{
		"message": "API Host - api.example.com",
		"tenant":  ctx.Tenant(),
	})
}

// adminHomeHandler handles admin tenant home
func adminHomeHandler(ctx pkg.Context) error {
	return ctx.JSON(200, map[string]interface{}{
		"message": "Admin Host - admin.example.com",
		"tenant":  ctx.Tenant(),
	})
}

// ============================================================================
// Utility Functions
// ============================================================================

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// ============================================================================
// Metrics and Profiling Handlers
// ============================================================================

// metricsHandler exposes Prometheus-compatible metrics
func metricsHandler(ctx pkg.Context) error {
	metrics := ctx.Metrics()

	ctx.Response().SetHeader("Content-Type", "text/plain; version=0.0.4")
	ctx.Response().WriteHeader(200)

	if metrics != nil {
		// Try to get metrics in Prometheus format
		metricsData, err := metrics.ExportPrometheus()
		if err == nil && len(metricsData) > 0 {
			ctx.Response().Write(metricsData)
			return nil
		}
	}

	// Fallback: show basic runtime metrics
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	output := fmt.Sprintf(`# HELP go_goroutines Number of goroutines that currently exist.
# TYPE go_goroutines gauge
go_goroutines %d

# HELP go_memstats_alloc_bytes Number of bytes allocated and still in use.
# TYPE go_memstats_alloc_bytes gauge
go_memstats_alloc_bytes %d

# HELP go_memstats_sys_bytes Number of bytes obtained from system.
# TYPE go_memstats_sys_bytes gauge
go_memstats_sys_bytes %d

# HELP go_memstats_heap_alloc_bytes Number of heap bytes allocated and still in use.
# TYPE go_memstats_heap_alloc_bytes gauge
go_memstats_heap_alloc_bytes %d

# HELP go_memstats_heap_sys_bytes Number of heap bytes obtained from system.
# TYPE go_memstats_heap_sys_bytes gauge
go_memstats_heap_sys_bytes %d

# HELP go_memstats_gc_sys_bytes Number of bytes used for garbage collection system metadata.
# TYPE go_memstats_gc_sys_bytes gauge
go_memstats_gc_sys_bytes %d
`,
		runtime.NumGoroutine(),
		m.Alloc,
		m.Sys,
		m.HeapAlloc,
		m.HeapSys,
		m.GCSys,
	)

	ctx.Response().Write([]byte(output))
	return nil
}

// pprofIndexHandler shows pprof index page
func pprofIndexHandler(ctx pkg.Context) error {
	ctx.Response().SetHeader("Content-Type", "text/html; charset=utf-8")
	ctx.Response().WriteHeader(200)
	ctx.Response().Write([]byte(`<html>
<head><title>pprof</title></head>
<body>
<h1>/debug/pprof/</h1>
<p>Profiles:</p>
<ul>
<li><a href="/debug/pprof/heap">heap</a> - Memory allocation profile</li>
<li><a href="/debug/pprof/goroutine">goroutine</a> - Goroutine stack traces</li>
<li><a href="/debug/pprof/threadcreate">threadcreate</a> - Thread creation profile</li>
<li><a href="/debug/pprof/block">block</a> - Blocking profile</li>
<li><a href="/debug/pprof/mutex">mutex</a> - Mutex contention profile</li>
<li><a href="/debug/pprof/profile">profile</a> - 30s CPU profile</li>
<li><a href="/debug/pprof/trace?seconds=5">trace</a> - 5s execution trace</li>
</ul>
</body>
</html>`))
	return nil
}

// pprofCmdlineHandler shows command line
func pprofCmdlineHandler(ctx pkg.Context) error {
	ctx.Response().SetHeader("Content-Type", "text/plain; charset=utf-8")
	ctx.Response().WriteHeader(200)
	cmdline := fmt.Sprintf("%s\n", os.Args[0])
	ctx.Response().Write([]byte(cmdline))
	return nil
}

// pprofProfileHandler generates CPU profile
func pprofProfileHandler(ctx pkg.Context) error {
	ctx.Response().SetHeader("Content-Type", "application/octet-stream")
	ctx.Response().SetHeader("Content-Disposition", "attachment; filename=profile")
	ctx.Response().WriteHeader(200)

	if err := pprof.StartCPUProfile(ctx.Response()); err != nil {
		ctx.Response().Write([]byte(fmt.Sprintf("Could not start CPU profile: %v", err)))
		return nil
	}
	time.Sleep(30 * time.Second)
	pprof.StopCPUProfile()
	return nil
}

// pprofSymbolHandler handles symbol lookups
func pprofSymbolHandler(ctx pkg.Context) error {
	ctx.Response().SetHeader("Content-Type", "text/plain; charset=utf-8")
	ctx.Response().WriteHeader(200)
	ctx.Response().Write([]byte("num_symbols: 1\n"))
	return nil
}

// pprofTraceHandler generates execution trace
func pprofTraceHandler(ctx pkg.Context) error {
	// Get trace duration from query parameter (default 5 seconds)
	seconds := 5
	if secStr := ctx.Query()["seconds"]; secStr != "" {
		if sec, err := strconv.Atoi(secStr); err == nil && sec > 0 && sec <= 60 {
			seconds = sec
		}
	}

	ctx.Response().SetHeader("Content-Type", "application/octet-stream")
	ctx.Response().SetHeader("Content-Disposition", "attachment; filename=trace")
	ctx.Response().WriteHeader(200)

	if err := trace.Start(ctx.Response()); err != nil {
		ctx.Response().Write([]byte(fmt.Sprintf("Could not start trace: %v", err)))
		return nil
	}

	time.Sleep(time.Duration(seconds) * time.Second)
	trace.Stop()

	return nil
}

// pprofHeapHandler generates heap profile
func pprofHeapHandler(ctx pkg.Context) error {
	ctx.Response().SetHeader("Content-Type", "application/octet-stream")
	ctx.Response().SetHeader("Content-Disposition", "attachment; filename=heap")
	ctx.Response().WriteHeader(200)

	runtime.GC() // Get up-to-date statistics
	if err := pprof.WriteHeapProfile(ctx.Response()); err != nil {
		ctx.Response().Write([]byte(fmt.Sprintf("Could not write heap profile: %v", err)))
	}
	return nil
}

// pprofGoroutineHandler generates goroutine profile
func pprofGoroutineHandler(ctx pkg.Context) error {
	ctx.Response().SetHeader("Content-Type", "text/plain; charset=utf-8")

	profile := pprof.Lookup("goroutine")
	if profile == nil {
		ctx.Response().WriteHeader(404)
		ctx.Response().Write([]byte("goroutine profile not found"))
		return nil
	}

	ctx.Response().WriteHeader(200)
	profile.WriteTo(ctx.Response(), 1)
	return nil
}

// pprofThreadCreateHandler generates thread creation profile
func pprofThreadCreateHandler(ctx pkg.Context) error {
	ctx.Response().SetHeader("Content-Type", "application/octet-stream")
	ctx.Response().SetHeader("Content-Disposition", "attachment; filename=threadcreate")

	profile := pprof.Lookup("threadcreate")
	if profile == nil {
		ctx.Response().WriteHeader(404)
		ctx.Response().Write([]byte("threadcreate profile not found"))
		return nil
	}

	ctx.Response().WriteHeader(200)
	profile.WriteTo(ctx.Response(), 0)
	return nil
}

// pprofBlockHandler generates blocking profile
func pprofBlockHandler(ctx pkg.Context) error {
	ctx.Response().SetHeader("Content-Type", "application/octet-stream")
	ctx.Response().SetHeader("Content-Disposition", "attachment; filename=block")

	profile := pprof.Lookup("block")
	if profile == nil {
		ctx.Response().WriteHeader(404)
		ctx.Response().Write([]byte("block profile not found"))
		return nil
	}

	ctx.Response().WriteHeader(200)
	profile.WriteTo(ctx.Response(), 0)
	return nil
}

// pprofMutexHandler generates mutex contention profile
func pprofMutexHandler(ctx pkg.Context) error {
	ctx.Response().SetHeader("Content-Type", "application/octet-stream")
	ctx.Response().SetHeader("Content-Disposition", "attachment; filename=mutex")

	profile := pprof.Lookup("mutex")
	if profile == nil {
		ctx.Response().WriteHeader(404)
		ctx.Response().Write([]byte("mutex profile not found"))
		return nil
	}

	ctx.Response().WriteHeader(200)
	profile.WriteTo(ctx.Response(), 0)
	return nil
}
