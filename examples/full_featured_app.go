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

// This example demonstrates all major features of the Rockstar Web Framework
func main() {
	// Create comprehensive framework configuration
	config := pkg.FrameworkConfig{
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
		CacheConfig: pkg.CacheConfig{
			Type:       "memory",
			MaxSize:    500 * 1024 * 1024, // 500 MB
			DefaultTTL: 10 * time.Minute,
		},
		SessionConfig: pkg.SessionConfig{
			StorageType:     pkg.SessionStorageDatabase,
			CookieName:      "rockstar_session_id",
			SessionLifetime: 24 * time.Hour,
			CookieSecure:    true,
			CookieHTTPOnly:  true,
			CookieSameSite:  "Strict",
			EncryptionKey:   []byte("12345678901234567890123456789012"), // 32 bytes for AES-256
			CleanupInterval: 10 * time.Minute,
		},
		ConfigFiles: []string{
			"config.yaml",
			"config.local.yaml",
		},
		I18nConfig: pkg.I18nConfig{
			DefaultLocale:     "en",
			LocalesDir:        "./locales",
			SupportedLocales:  []string{"en", "de", "fr", "es"},
			FallbackToDefault: true,
		},
		SecurityConfig: pkg.SecurityConfig{
			XFrameOptions:    "SAMEORIGIN",
			EnableXSSProtect: true,
			EnableCSRF:       true,
			MaxRequestSize:   20 * 1024 * 1024, // 20 MB
			RequestTimeout:   60 * time.Second,
			AllowedOrigins:   []string{"https://example.com"},
			EncryptionKey:    "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", // 64 hex chars for 32 bytes
			JWTSecret:        "your-secret-jwt-key",
			CSRFTokenExpiry:  24 * time.Hour,
		},
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

	// Create framework instance
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

	// Start server
	fmt.Println("🎸 Rockstar Web Framework - Full Featured Application")
	fmt.Println("==================================================")
	fmt.Println("Server starting on :8080")
	fmt.Println("")
	fmt.Println("Available endpoints:")
	fmt.Println("  http://localhost:8080/          - Home")
	fmt.Println("  http://localhost:8080/health    - Health check")
	fmt.Println("  http://localhost:8080/metrics   - Metrics")
	fmt.Println("  http://localhost:8080/debug/pprof - Profiling")
	fmt.Println("")
	fmt.Println("Host-specific routes (use Host header or hosts file):")
	fmt.Println("  api.example.com:8080            - API tenant")
	fmt.Println("  admin.example.com:8080          - Admin tenant")
	fmt.Println("")
	fmt.Println("To test host-specific routes:")
	fmt.Println("  curl -H \"Host: api.example.com\" http://localhost:8080/")
	fmt.Println("  curl -H \"Host: admin.example.com\" http://localhost:8080/")
	fmt.Println("==================================================")

	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func setupLifecycleHooks(app *pkg.Framework) {
	// Startup hooks
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

	// Shutdown hooks
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

func setupGlobalMiddleware(app *pkg.Framework) {
	// Request ID middleware
	app.Use(func(ctx pkg.Context, next pkg.HandlerFunc) error {
		requestID := generateRequestID()
		ctx.SetHeader("X-Request-ID", requestID)
		return next(ctx)
	})

	// Logging middleware
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

	// Metrics middleware
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

	// Recovery middleware
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

func setupRoutes(app *pkg.Framework) {
	router := app.Router()

	// Public routes
	router.GET("/", homeHandler)
	router.GET("/health", healthHandler)
	router.GET("/ready", readinessHandler)

	// Metrics and profiling endpoints
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

	// API v1 routes
	v1 := router.Group("/api/v1")
	v1.GET("/status", apiStatusHandler)

	// Authenticated API routes
	authAPI := router.Group("/api/v1", authenticationMiddleware)
	authAPI.GET("/profile", profileHandler)
	authAPI.PUT("/profile", updateProfileHandler)
	authAPI.POST("/logout", logoutHandler)

	// Admin routes (requires authentication and authorization)
	admin := router.Group("/admin", authenticationMiddleware, adminAuthorizationMiddleware)
	admin.GET("/users", listUsersHandler)
	admin.POST("/users", createUserHandler)
	admin.GET("/users/:id", getUserHandler)
	admin.PUT("/users/:id", updateUserHandler)
	admin.DELETE("/users/:id", deleteUserHandler)

	// Multi-protocol API examples

	// REST API
	router.Group("/api/rest").GET("/products", restProductsHandler)

	// GraphQL API (if schema is available)
	// router.GraphQL("/api/graphql", graphqlSchema)

	// WebSocket
	router.WebSocket("/ws", websocketHandler)

	// Static files
	// router.Static("/static", pkg.NewOSFileSystem("./public"))

	// Host-specific routes (multi-tenancy)
	apiHost := router.Host("api.example.com")
	apiHost.GET("/", apiHomeHandler)

	adminHost := router.Host("admin.example.com")
	adminHost.GET("/", adminHomeHandler)
}

func setupGracefulShutdown(app *pkg.Framework) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n🛑 Shutdown signal received, gracefully shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := app.Shutdown(30 * time.Second); err != nil {
			log.Printf("Error during shutdown: %v", err)
		}

		<-ctx.Done()
		os.Exit(0)
	}()
}

// Middleware implementations

func authenticationMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
	token := ctx.GetHeader("Authorization")
	if token == "" {
		ctx.JSON(401, map[string]interface{}{
			"error": "Authentication required",
		})
		return nil
	}

	// Validate token using security manager
	// user, err := ctx.Security().AuthenticateJWT(token)
	// if err != nil {
	//     ctx.JSON(401, map[string]interface{}{"error": "Invalid token"})
	//     return nil
	// }

	return next(ctx)
}

func adminAuthorizationMiddleware(ctx pkg.Context, next pkg.HandlerFunc) error {
	user := ctx.User()
	if user == nil {
		ctx.JSON(401, map[string]interface{}{
			"error": "Authentication required",
		})
		return nil
	}

	// Check if user has admin role
	if !ctx.IsAuthorized("admin", "access") {
		ctx.JSON(403, map[string]interface{}{
			"error": "Insufficient permissions",
		})
		return nil
	}

	return next(ctx)
}

// Handler implementations

func homeHandler(ctx pkg.Context) error {
	ctx.JSON(200, map[string]interface{}{
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
		},
	})
	return nil
}

func healthHandler(ctx pkg.Context) error {
	ctx.JSON(200, map[string]interface{}{
		"status": "healthy",
		"time":   time.Now().Unix(),
	})
	return nil
}

func readinessHandler(ctx pkg.Context) error {
	// Check if all services are ready
	dbHealthy := true    // ctx.DB().Ping() == nil
	cacheHealthy := true // ctx.Cache().Ping() == nil

	if !dbHealthy || !cacheHealthy {
		ctx.JSON(503, map[string]interface{}{
			"status": "not ready",
			"checks": map[string]bool{
				"database": dbHealthy,
				"cache":    cacheHealthy,
			},
		})
		return nil
	}

	ctx.JSON(200, map[string]interface{}{
		"status": "ready",
	})
	return nil
}

func apiStatusHandler(ctx pkg.Context) error {
	ctx.JSON(200, map[string]interface{}{
		"api_version": "v1",
		"status":      "operational",
	})
	return nil
}

func profileHandler(ctx pkg.Context) error {
	user := ctx.User()
	ctx.JSON(200, map[string]interface{}{
		"user": user,
	})
	return nil
}

func updateProfileHandler(ctx pkg.Context) error {
	// Update user profile
	ctx.JSON(200, map[string]interface{}{
		"message": "Profile updated successfully",
	})
	return nil
}

func logoutHandler(ctx pkg.Context) error {
	// Session is managed by the framework
	// In a real implementation, you would get the session ID from the cookie
	// and destroy it using the session manager

	ctx.JSON(200, map[string]interface{}{
		"message": "Logged out successfully",
	})
	return nil
}

func listUsersHandler(ctx pkg.Context) error {
	// Fetch users from database
	ctx.JSON(200, map[string]interface{}{
		"users": []map[string]interface{}{},
		"total": 0,
	})
	return nil
}

func createUserHandler(ctx pkg.Context) error {
	ctx.JSON(201, map[string]interface{}{
		"message": "User created successfully",
	})
	return nil
}

func getUserHandler(ctx pkg.Context) error {
	userID := ctx.Params()["id"]
	ctx.JSON(200, map[string]interface{}{
		"id": userID,
	})
	return nil
}

func updateUserHandler(ctx pkg.Context) error {
	userID := ctx.Params()["id"]
	ctx.JSON(200, map[string]interface{}{
		"message": fmt.Sprintf("User %s updated successfully", userID),
	})
	return nil
}

func deleteUserHandler(ctx pkg.Context) error {
	userID := ctx.Params()["id"]
	ctx.JSON(200, map[string]interface{}{
		"message": fmt.Sprintf("User %s deleted successfully", userID),
	})
	return nil
}

func restProductsHandler(ctx pkg.Context) error {
	ctx.JSON(200, map[string]interface{}{
		"products": []map[string]interface{}{},
	})
	return nil
}

func websocketHandler(ctx pkg.Context, conn pkg.WebSocketConnection) error {
	for {
		messageType, data, err := conn.ReadMessage()
		if err != nil {
			return err
		}

		// Echo message back
		if err := conn.WriteMessage(messageType, data); err != nil {
			return err
		}
	}
}

func apiHomeHandler(ctx pkg.Context) error {
	ctx.JSON(200, map[string]interface{}{
		"message": "API Host - api.example.com",
		"tenant":  ctx.Tenant(),
	})
	return nil
}

func adminHomeHandler(ctx pkg.Context) error {
	ctx.JSON(200, map[string]interface{}{
		"message": "Admin Host - admin.example.com",
		"tenant":  ctx.Tenant(),
	})
	return nil
}

// Utility functions

func generateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// Metrics and profiling handlers

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

func pprofIndexHandler(ctx pkg.Context) error {
	ctx.Response().SetHeader("Content-Type", "text/html; charset=utf-8")
	ctx.Response().WriteHeader(200)
	ctx.Response().Write([]byte(`<html>
<head><title>pprof</title></head>
<body>
<h1>/debug/pprof/</h1>
<p>Profiles:</p>
<ul>
<li><a href="/debug/pprof/heap">heap</a></li>
<li><a href="/debug/pprof/goroutine">goroutine</a></li>
<li><a href="/debug/pprof/threadcreate">threadcreate</a></li>
<li><a href="/debug/pprof/block">block</a></li>
<li><a href="/debug/pprof/mutex">mutex</a></li>
<li><a href="/debug/pprof/profile">profile (30s CPU profile)</a></li>
<li><a href="/debug/pprof/trace?seconds=5">trace (5s trace)</a></li>
</ul>
</body>
</html>`))
	return nil
}

func pprofCmdlineHandler(ctx pkg.Context) error {
	ctx.Response().SetHeader("Content-Type", "text/plain; charset=utf-8")
	ctx.Response().WriteHeader(200)
	cmdline := fmt.Sprintf("%s\n", os.Args[0])
	ctx.Response().Write([]byte(cmdline))
	return nil
}

func pprofProfileHandler(ctx pkg.Context) error {
	// CPU profiling for 30 seconds
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

func pprofSymbolHandler(ctx pkg.Context) error {
	ctx.Response().SetHeader("Content-Type", "text/plain; charset=utf-8")
	ctx.Response().WriteHeader(200)
	ctx.Response().Write([]byte("num_symbols: 1\n"))
	return nil
}

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

	// Start tracing
	if err := trace.Start(ctx.Response()); err != nil {
		ctx.Response().Write([]byte(fmt.Sprintf("Could not start trace: %v", err)))
		return nil
	}

	// Trace for the specified duration
	time.Sleep(time.Duration(seconds) * time.Second)
	trace.Stop()

	return nil
}

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
