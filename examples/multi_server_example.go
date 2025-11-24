package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
	// Create server manager
	sm := pkg.NewServerManager()

	// Example 1: Basic Multi-Server Setup
	basicMultiServerExample(sm)

	// Example 2: Multi-Tenant Configuration
	// multiTenantExample(sm)

	// Example 3: Port Reuse Example
	// portReuseExample(sm)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down servers...")

	// Graceful shutdown with 30 second timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_ = shutdownCtx

	if err := sm.GracefulShutdown(30 * time.Second); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Servers stopped")
}

// basicMultiServerExample demonstrates running multiple servers in one process
func basicMultiServerExample(sm pkg.ServerManager) {
	log.Println("=== Basic Multi-Server Example ===")

	// Configure first server (HTTP on port 8080)
	config1 := pkg.ServerConfig{
		EnableHTTP1:    true,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	server1 := sm.NewServer(config1)
	router1 := pkg.NewRouter()

	router1.GET("/", func(ctx pkg.Context) error {
		return ctx.String(200, "Hello from Server 1 (HTTP)")
	})

	router1.GET("/health", func(ctx pkg.Context) error {
		return ctx.JSON(200, map[string]string{
			"status": "healthy",
			"server": "server1",
		})
	})

	server1.SetRouter(router1)

	// Configure second server (HTTP on port 8081)
	config2 := pkg.ServerConfig{
		EnableHTTP1:    true,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	server2 := sm.NewServer(config2)
	router2 := pkg.NewRouter()

	router2.GET("/", func(ctx pkg.Context) error {
		return ctx.String(200, "Hello from Server 2 (HTTP)")
	})

	router2.GET("/api/data", func(ctx pkg.Context) error {
		return ctx.JSON(200, map[string]interface{}{
			"server": "server2",
			"data":   []string{"item1", "item2", "item3"},
		})
	})

	server2.SetRouter(router2)

	// Start servers
	go func() {
		log.Println("Starting Server 1 on :8080")
		if err := server1.Listen(":8080"); err != nil {
			log.Printf("Server 1 error: %v", err)
		}
	}()

	go func() {
		log.Println("Starting Server 2 on :8081")
		if err := server2.Listen(":8081"); err != nil {
			log.Printf("Server 2 error: %v", err)
		}
	}()

	// Wait a moment for servers to start
	time.Sleep(100 * time.Millisecond)

	log.Printf("Running %d servers", sm.GetServerCount())
	log.Println("Server 1: http://localhost:8080")
	log.Println("Server 2: http://localhost:8081")
}

// multiTenantExample demonstrates multi-tenant configuration
func multiTenantExample(sm pkg.ServerManager) {
	log.Println("=== Multi-Tenant Example ===")

	// Register hosts for Tenant 1
	host1Config := pkg.HostConfig{
		Hostname: "app1.tenant1.local",
		TenantID: "tenant1",
		RateLimits: &pkg.RateLimitConfig{
			Enabled:           true,
			RequestsPerSecond: 100,
			BurstSize:         20,
			Storage:           "memory",
		},
	}

	host2Config := pkg.HostConfig{
		Hostname: "app2.tenant1.local",
		TenantID: "tenant1",
		RateLimits: &pkg.RateLimitConfig{
			Enabled:           true,
			RequestsPerSecond: 50,
			BurstSize:         10,
			Storage:           "memory",
		},
	}

	// Register hosts for Tenant 2
	host3Config := pkg.HostConfig{
		Hostname: "app1.tenant2.local",
		TenantID: "tenant2",
		RateLimits: &pkg.RateLimitConfig{
			Enabled:           true,
			RequestsPerSecond: 200,
			BurstSize:         40,
			Storage:           "memory",
		},
	}

	// Register hosts
	if err := sm.RegisterHost("app1.tenant1.local", host1Config); err != nil {
		log.Fatalf("Failed to register host: %v", err)
	}
	if err := sm.RegisterHost("app2.tenant1.local", host2Config); err != nil {
		log.Fatalf("Failed to register host: %v", err)
	}
	if err := sm.RegisterHost("app1.tenant2.local", host3Config); err != nil {
		log.Fatalf("Failed to register host: %v", err)
	}

	// Register tenants
	if err := sm.RegisterTenant("tenant1", []string{
		"app1.tenant1.local",
		"app2.tenant1.local",
	}); err != nil {
		log.Fatalf("Failed to register tenant1: %v", err)
	}

	if err := sm.RegisterTenant("tenant2", []string{
		"app1.tenant2.local",
	}); err != nil {
		log.Fatalf("Failed to register tenant2: %v", err)
	}

	// Create server with host configurations
	config := pkg.ServerConfig{
		EnableHTTP1: true,
		EnableHTTP2: true,
		HostConfigs: map[string]*pkg.HostConfig{
			"app1.tenant1.local": &host1Config,
			"app2.tenant1.local": &host2Config,
			"app1.tenant2.local": &host3Config,
		},
	}

	server := sm.NewServer(config)
	router := pkg.NewRouter()

	// Routes for app1.tenant1.local
	app1Tenant1Router := router.Host("app1.tenant1.local")
	app1Tenant1Router.GET("/", func(c pkg.Context) error {
		return c.JSON(200, map[string]string{
			"tenant":  "tenant1",
			"app":     "app1",
			"message": "Welcome to Tenant 1 - App 1",
		})
	})

	app1Tenant1Router.GET("/api/data", func(ctx pkg.Context) error {
		return ctx.JSON(200, map[string]interface{}{
			"tenant": "tenant1",
			"app":    "app1",
			"data":   []string{"tenant1-data1", "tenant1-data2"},
		})
	})

	// Routes for app2.tenant1.local
	app2Tenant1Router := router.Host("app2.tenant1.local")
	app2Tenant1Router.GET("/", func(ctx pkg.Context) error {
		return ctx.JSON(200, map[string]string{
			"tenant":  "tenant1",
			"app":     "app2",
			"message": "Welcome to Tenant 1 - App 2",
		})
	})

	// Routes for app1.tenant2.local
	app1Tenant2Router := router.Host("app1.tenant2.local")
	app1Tenant2Router.GET("/", func(ctx pkg.Context) error {
		return ctx.JSON(200, map[string]string{
			"tenant":  "tenant2",
			"app":     "app1",
			"message": "Welcome to Tenant 2 - App 1",
		})
	})

	app1Tenant2Router.GET("/api/users", func(ctx pkg.Context) error {
		return ctx.JSON(200, map[string]interface{}{
			"tenant": "tenant2",
			"users":  []string{"user1", "user2", "user3"},
		})
	})

	server.SetRouter(router)

	// Start server
	go func() {
		log.Println("Starting multi-tenant server on :8080")
		if err := server.Listen(":8080"); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	log.Println("Multi-tenant server running on :8080")
	log.Println("Tenant 1 - App 1: http://app1.tenant1.local:8080")
	log.Println("Tenant 1 - App 2: http://app2.tenant1.local:8080")
	log.Println("Tenant 2 - App 1: http://app1.tenant2.local:8080")
	log.Println("\nNote: Add these entries to your /etc/hosts file:")
	log.Println("127.0.0.1 app1.tenant1.local")
	log.Println("127.0.0.1 app2.tenant1.local")
	log.Println("127.0.0.1 app1.tenant2.local")
}

// portReuseExample demonstrates port reuse with different hosts
func portReuseExample(sm pkg.ServerManager) {
	log.Println("=== Port Reuse Example ===")

	// Register hosts
	site1Config := pkg.HostConfig{
		Hostname: "site1.local",
	}

	site2Config := pkg.HostConfig{
		Hostname: "site2.local",
	}

	if err := sm.RegisterHost("site1.local", site1Config); err != nil {
		log.Fatalf("Failed to register host: %v", err)
	}
	if err := sm.RegisterHost("site2.local", site2Config); err != nil {
		log.Fatalf("Failed to register host: %v", err)
	}

	// Create server with multiple hosts on same port
	config := pkg.ServerConfig{
		EnableHTTP1: true,
		HostConfigs: map[string]*pkg.HostConfig{
			"site1.local": &site1Config,
			"site2.local": &site2Config,
		},
	}

	server := sm.NewServer(config)
	router := pkg.NewRouter()

	// Routes for site1.local
	site1Router := router.Host("site1.local")
	site1Router.GET("/", func(ctx pkg.Context) error {
		return ctx.HTML(200, `
			<html>
				<head><title>Site 1</title></head>
				<body>
					<h1>Welcome to Site 1</h1>
					<p>This is served from site1.local</p>
				</body>
			</html>
		`, nil)
	})

	site1Router.GET("/about", func(ctx pkg.Context) error {
		return ctx.JSON(200, map[string]string{
			"site":        "site1",
			"description": "This is Site 1",
		})
	})

	// Routes for site2.local
	site2Router := router.Host("site2.local")
	site2Router.GET("/", func(ctx pkg.Context) error {
		return ctx.HTML(200, `
			<html>
				<head><title>Site 2</title></head>
				<body>
					<h1>Welcome to Site 2</h1>
					<p>This is served from site2.local</p>
				</body>
			</html>
		`, nil)
	})

	site2Router.GET("/contact", func(ctx pkg.Context) error {
		return ctx.JSON(200, map[string]string{
			"site":  "site2",
			"email": "contact@site2.local",
		})
	})

	server.SetRouter(router)

	// Start server
	go func() {
		log.Println("Starting server with port reuse on :8080")
		if err := server.Listen(":8080"); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	log.Println("Server running on :8080 with multiple hosts")
	log.Println("Site 1: http://site1.local:8080")
	log.Println("Site 2: http://site2.local:8080")
	log.Println("\nNote: Add these entries to your /etc/hosts file:")
	log.Println("127.0.0.1 site1.local")
	log.Println("127.0.0.1 site2.local")
}

// healthMonitoringExample demonstrates health monitoring
func healthMonitoringExample(sm pkg.ServerManager) {
	log.Println("=== Health Monitoring Example ===")

	// Start health monitoring
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if err := sm.HealthCheck(); err != nil {
				log.Printf("❌ Health check failed: %v", err)
			} else {
				log.Println("✅ All servers healthy")
			}

			// Print server status
			servers := sm.GetServers()
			for i, server := range servers {
				status := "stopped"
				if server.IsRunning() {
					status = "running"
				}
				log.Printf("  Server %d: %s - %s - %s",
					i+1,
					server.Addr(),
					server.Protocol(),
					status,
				)
			}
		}
	}()
}

// dynamicServerManagementExample demonstrates adding/removing servers dynamically
func dynamicServerManagementExample(sm pkg.ServerManager) {
	log.Println("=== Dynamic Server Management Example ===")

	// Create and add server dynamically
	config := pkg.ServerConfig{
		EnableHTTP1: true,
	}

	server := sm.NewServer(config)
	router := pkg.NewRouter()

	router.GET("/", func(ctx pkg.Context) error {
		return ctx.String(200, "Dynamic Server")
	})

	server.SetRouter(router)

	// Add to manager
	// This is a simplified example - in production you'd have proper methods
	fmt.Printf("Server manager type: %T\n", sm)

	// Start the server
	go func() {
		log.Println("Starting dynamic server on :9000")
		if err := server.Listen(":9000"); err != nil {
			log.Printf("Dynamic server error: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	log.Println("Dynamic server started on :9000")

	// Simulate running for a while
	time.Sleep(5 * time.Second)

	// Stop the server
	log.Println("Stopping dynamic server...")
	if err := server.Close(); err != nil {
		log.Printf("Error stopping server: %v", err)
	}

	log.Println("Dynamic server stopped")
}
