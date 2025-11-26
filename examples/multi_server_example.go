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

// This example demonstrates multi-tenancy and host-based routing in the Rockstar Web Framework.
// It shows how to run multiple servers, configure tenant isolation, implement host-based routing,
// and manage tenant-specific configurations.

func main() {
	fmt.Println("🎸 Rockstar Web Framework - Multi-Server & Multi-Tenancy Example\n")

	// Create server manager
	serverManager := pkg.NewServerManager()

	fmt.Println("1. Basic Multi-Server Setup")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	basicMultiServerSetup(serverManager)

	fmt.Println("\n2. Multi-Tenant Configuration")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	multiTenantSetup(serverManager)

	fmt.Println("\n3. Host-Based Routing")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	hostBasedRoutingSetup(serverManager)

	fmt.Println("\n✅ All servers configured and running!")
	fmt.Println("\nServer Endpoints:")
	fmt.Println("  HTTP Server 1:  http://localhost:8080")
	fmt.Println("  HTTP Server 2:  http://localhost:8081")
	fmt.Println("  Admin Server:   http://localhost:9000")
	fmt.Println("\nMulti-Tenant Endpoints (add to /etc/hosts):")
	fmt.Println("  127.0.0.1 tenant1.local")
	fmt.Println("  127.0.0.1 tenant2.local")
	fmt.Println("  127.0.0.1 admin.local")
	fmt.Println("\nPress Ctrl+C to shutdown gracefully...")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\n🛑 Shutting down servers...")

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := serverManager.GracefulShutdown(30 * time.Second); err != nil {
		log.Printf("Shutdown error: %v", err)
	} else {
		fmt.Println("✅ All servers stopped gracefully")
	}

	_ = shutdownCtx
}

// basicMultiServerSetup demonstrates running multiple HTTP servers on different ports
func basicMultiServerSetup(sm pkg.ServerManager) {
	// Server 1 - Main application server
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
		return ctx.JSON(200, map[string]interface{}{
			"server":  "server1",
			"message": "Main Application Server",
			"port":    8080,
		})
	})

	router1.GET("/health", func(ctx pkg.Context) error {
		return ctx.JSON(200, map[string]interface{}{
			"status": "healthy",
			"server": "server1",
			"uptime": time.Since(startTime).String(),
		})
	})

	router1.GET("/api/users", func(ctx pkg.Context) error {
		return ctx.JSON(200, map[string]interface{}{
			"users": []map[string]string{
				{"id": "1", "name": "Alice"},
				{"id": "2", "name": "Bob"},
			},
		})
	})

	server1.SetRouter(router1)

	// Server 2 - API server
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
		return ctx.JSON(200, map[string]interface{}{
			"server":  "server2",
			"message": "API Server",
			"port":    8081,
		})
	})

	router2.GET("/api/data", func(ctx pkg.Context) error {
		return ctx.JSON(200, map[string]interface{}{
			"data": []string{"item1", "item2", "item3"},
		})
	})

	router2.POST("/api/data", func(ctx pkg.Context) error {
		return ctx.JSON(201, map[string]interface{}{
			"message": "Data created",
		})
	})

	server2.SetRouter(router2)

	// Start servers in goroutines
	go func() {
		fmt.Println("  ✓ Starting Server 1 on :8080 (Main Application)")
		if err := server1.Listen(":8080"); err != nil {
			log.Printf("Server 1 error: %v", err)
		}
	}()

	go func() {
		fmt.Println("  ✓ Starting Server 2 on :8081 (API Server)")
		if err := server2.Listen(":8081"); err != nil {
			log.Printf("Server 2 error: %v", err)
		}
	}()

	// Wait for servers to start
	time.Sleep(200 * time.Millisecond)

	fmt.Printf("  Running %d servers\n", sm.GetServerCount())
}

// multiTenantSetup demonstrates multi-tenant configuration with isolation
func multiTenantSetup(sm pkg.ServerManager) {
	// Configure Tenant 1
	tenant1Config := pkg.HostConfig{
		Hostname: "tenant1.local",
		TenantID: "tenant1",
		RateLimits: &pkg.RateLimitConfig{
			Enabled:           true,
			RequestsPerSecond: 100,
			BurstSize:         20,
			Storage:           "memory",
		},
	}

	// Configure Tenant 2
	tenant2Config := pkg.HostConfig{
		Hostname: "tenant2.local",
		TenantID: "tenant2",
		RateLimits: &pkg.RateLimitConfig{
			Enabled:           true,
			RequestsPerSecond: 200,
			BurstSize:         40,
			Storage:           "memory",
		},
	}

	// Configure Admin
	adminConfig := pkg.HostConfig{
		Hostname: "admin.local",
		TenantID: "admin",
		RateLimits: &pkg.RateLimitConfig{
			Enabled:           true,
			RequestsPerSecond: 50,
			BurstSize:         10,
			Storage:           "memory",
		},
	}

	// Register hosts
	if err := sm.RegisterHost("tenant1.local", tenant1Config); err != nil {
		log.Printf("Failed to register tenant1 host: %v", err)
	} else {
		fmt.Println("  ✓ Registered tenant1.local")
	}

	if err := sm.RegisterHost("tenant2.local", tenant2Config); err != nil {
		log.Printf("Failed to register tenant2 host: %v", err)
	} else {
		fmt.Println("  ✓ Registered tenant2.local")
	}

	if err := sm.RegisterHost("admin.local", adminConfig); err != nil {
		log.Printf("Failed to register admin host: %v", err)
	} else {
		fmt.Println("  ✓ Registered admin.local")
	}

	// Register tenants with their hosts
	if err := sm.RegisterTenant("tenant1", []string{"tenant1.local"}); err != nil {
		log.Printf("Failed to register tenant1: %v", err)
	} else {
		fmt.Println("  ✓ Registered tenant1 with 100 req/s limit")
	}

	if err := sm.RegisterTenant("tenant2", []string{"tenant2.local"}); err != nil {
		log.Printf("Failed to register tenant2: %v", err)
	} else {
		fmt.Println("  ✓ Registered tenant2 with 200 req/s limit")
	}

	if err := sm.RegisterTenant("admin", []string{"admin.local"}); err != nil {
		log.Printf("Failed to register admin: %v", err)
	} else {
		fmt.Println("  ✓ Registered admin tenant")
	}

	fmt.Println("\n  Tenant Isolation Features:")
	fmt.Println("    - Separate rate limits per tenant")
	fmt.Println("    - Isolated data access")
	fmt.Println("    - Tenant-specific configuration")
}

// hostBasedRoutingSetup demonstrates host-based routing with different content per host
func hostBasedRoutingSetup(sm pkg.ServerManager) {
	// Create server with host configurations
	config := pkg.ServerConfig{
		EnableHTTP1: true,
		EnableHTTP2: true,
		HostConfigs: map[string]*pkg.HostConfig{
			"tenant1.local": {
				Hostname: "tenant1.local",
				TenantID: "tenant1",
			},
			"tenant2.local": {
				Hostname: "tenant2.local",
				TenantID: "tenant2",
			},
			"admin.local": {
				Hostname: "admin.local",
				TenantID: "admin",
			},
		},
	}

	server := sm.NewServer(config)
	router := pkg.NewRouter()

	// Routes for Tenant 1
	tenant1Router := router.Host("tenant1.local")
	tenant1Router.GET("/", func(ctx pkg.Context) error {
		return ctx.JSON(200, map[string]interface{}{
			"tenant":  "tenant1",
			"message": "Welcome to Tenant 1",
			"features": []string{
				"Feature A",
				"Feature B",
				"Feature C",
			},
		})
	})

	tenant1Router.GET("/api/data", func(ctx pkg.Context) error {
		return ctx.JSON(200, map[string]interface{}{
			"tenant": "tenant1",
			"data":   []string{"tenant1-data1", "tenant1-data2"},
		})
	})

	tenant1Router.GET("/dashboard", func(ctx pkg.Context) error {
		return ctx.HTML(200, `
			<html>
				<head><title>Tenant 1 Dashboard</title></head>
				<body>
					<h1>Tenant 1 Dashboard</h1>
					<p>Welcome to your isolated tenant environment</p>
				</body>
			</html>
		`, nil)
	})

	// Routes for Tenant 2
	tenant2Router := router.Host("tenant2.local")
	tenant2Router.GET("/", func(ctx pkg.Context) error {
		return ctx.JSON(200, map[string]interface{}{
			"tenant":  "tenant2",
			"message": "Welcome to Tenant 2",
			"features": []string{
				"Feature X",
				"Feature Y",
				"Feature Z",
			},
		})
	})

	tenant2Router.GET("/api/users", func(ctx pkg.Context) error {
		return ctx.JSON(200, map[string]interface{}{
			"tenant": "tenant2",
			"users": []map[string]string{
				{"id": "1", "name": "User 1"},
				{"id": "2", "name": "User 2"},
			},
		})
	})

	// Routes for Admin
	adminRouter := router.Host("admin.local")
	adminRouter.GET("/", func(ctx pkg.Context) error {
		return ctx.JSON(200, map[string]interface{}{
			"tenant":  "admin",
			"message": "Admin Portal",
		})
	})

	adminRouter.GET("/admin/tenants", func(ctx pkg.Context) error {
		return ctx.JSON(200, map[string]interface{}{
			"tenants": []map[string]interface{}{
				{
					"id":     "tenant1",
					"name":   "Tenant 1",
					"status": "active",
				},
				{
					"id":     "tenant2",
					"name":   "Tenant 2",
					"status": "active",
				},
			},
		})
	})

	adminRouter.GET("/admin/stats", func(ctx pkg.Context) error {
		return ctx.JSON(200, map[string]interface{}{
			"total_tenants":  2,
			"active_servers": sm.GetServerCount(),
			"uptime":         time.Since(startTime).String(),
		})
	})

	server.SetRouter(router)

	// Start multi-tenant server
	go func() {
		fmt.Println("  ✓ Starting multi-tenant server on :9000")
		if err := server.Listen(":9000"); err != nil {
			log.Printf("Multi-tenant server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(200 * time.Millisecond)

	fmt.Println("\n  Host-Based Routes Configured:")
	fmt.Println("    tenant1.local:9000 → Tenant 1 routes")
	fmt.Println("    tenant2.local:9000 → Tenant 2 routes")
	fmt.Println("    admin.local:9000   → Admin routes")
}

var startTime = time.Now()
