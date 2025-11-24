package main

import (
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
	fmt.Println("Rockstar Web Framework - Forward Proxy Example")
	fmt.Println("==============================================")

	// Create cache manager for proxy caching
	cache := pkg.NewCacheManager(pkg.CacheConfig{})

	// Create proxy configuration
	config := pkg.DefaultProxyConfig()
	config.LoadBalancerType = "round_robin"
	config.CircuitBreakerEnabled = true
	config.CircuitBreakerThreshold = 5
	config.MaxRetries = 3
	config.CacheEnabled = true
	config.CacheTTL = 5 * time.Minute
	config.HealthCheckEnabled = true
	config.HealthCheckInterval = 10 * time.Second

	// Create proxy manager
	proxyManager := pkg.NewProxyManager(config, cache)

	// Add backend servers
	backend1URL, _ := url.Parse("http://backend1.example.com:8080")
	backend1 := &pkg.Backend{
		ID:                  "backend1",
		URL:                 backend1URL,
		Weight:              1,
		IsActive:            true,
		HealthCheckPath:     "/health",
		HealthCheckInterval: 10 * time.Second,
		HealthCheckTimeout:  5 * time.Second,
	}

	backend2URL, _ := url.Parse("http://backend2.example.com:8080")
	backend2 := &pkg.Backend{
		ID:                  "backend2",
		URL:                 backend2URL,
		Weight:              1,
		IsActive:            true,
		HealthCheckPath:     "/health",
		HealthCheckInterval: 10 * time.Second,
		HealthCheckTimeout:  5 * time.Second,
	}

	backend3URL, _ := url.Parse("http://backend3.example.com:8080")
	backend3 := &pkg.Backend{
		ID:                  "backend3",
		URL:                 backend3URL,
		Weight:              2, // Higher weight for weighted load balancing
		IsActive:            true,
		HealthCheckPath:     "/health",
		HealthCheckInterval: 10 * time.Second,
		HealthCheckTimeout:  5 * time.Second,
	}

	// Add backends to proxy manager
	if err := proxyManager.AddBackend(backend1); err != nil {
		log.Fatalf("Failed to add backend1: %v", err)
	}
	fmt.Println("✓ Added backend1")

	if err := proxyManager.AddBackend(backend2); err != nil {
		log.Fatalf("Failed to add backend2: %v", err)
	}
	fmt.Println("✓ Added backend2")

	if err := proxyManager.AddBackend(backend3); err != nil {
		log.Fatalf("Failed to add backend3: %v", err)
	}
	fmt.Println("✓ Added backend3")

	// List all backends
	fmt.Println("\nConfigured Backends:")
	backends := proxyManager.ListBackends()
	for _, backend := range backends {
		fmt.Printf("  - %s: %s (Weight: %d, Active: %v)\n",
			backend.ID, backend.URL.String(), backend.Weight, backend.IsActive)
	}

	// Perform initial health check
	fmt.Println("\nPerforming health check...")
	if err := proxyManager.HealthCheck(); err != nil {
		log.Printf("Health check error: %v", err)
	}

	// Display health status
	healthStatus := proxyManager.GetHealthStatus()
	fmt.Println("\nBackend Health Status:")
	for backendID, health := range healthStatus {
		status := "Healthy"
		if !health.IsHealthy {
			status = "Unhealthy"
		}
		fmt.Printf("  - %s: %s (Last check: %s)\n",
			backendID, status, health.LastCheck.Format(time.RFC3339))
		if health.ErrorMessage != "" {
			fmt.Printf("    Error: %s\n", health.ErrorMessage)
		}
	}

	// Get load balancer info
	lb := proxyManager.GetLoadBalancer()
	fmt.Printf("\nLoad Balancer Type: %s\n", lb.Type())

	// Get circuit breaker info
	cb := proxyManager.GetCircuitBreaker()
	fmt.Println("\nCircuit Breaker States:")
	for _, backend := range backends {
		state := cb.GetState(backend.ID)
		fmt.Printf("  - %s: %s\n", backend.ID, state)
	}

	// Get connection pool stats
	pool := proxyManager.GetConnectionPool()
	poolStats := pool.Stats()
	fmt.Printf("\nConnection Pool Stats:\n")
	fmt.Printf("  Total Connections: %d\n", poolStats.TotalConnections)
	fmt.Printf("  Active Connections: %d\n", poolStats.ActiveConnections)
	fmt.Printf("  Idle Connections: %d\n", poolStats.IdleConnections)

	// Display proxy metrics
	metrics := proxyManager.GetMetrics()
	fmt.Println("\nProxy Metrics:")
	fmt.Printf("  Total Requests: %d\n", metrics.TotalRequests)
	fmt.Printf("  Successful Requests: %d\n", metrics.SuccessfulRequests)
	fmt.Printf("  Failed Requests: %d\n", metrics.FailedRequests)
	fmt.Printf("  Average Response Time: %s\n", metrics.AverageResponseTime)
	fmt.Printf("  Cache Hits: %d\n", metrics.CacheHits)
	fmt.Printf("  Cache Misses: %d\n", metrics.CacheMisses)
	fmt.Printf("  Total Retries: %d\n", metrics.TotalRetries)

	// Display per-backend metrics
	if len(metrics.BackendMetrics) > 0 {
		fmt.Println("\nPer-Backend Metrics:")
		for backendID, backendMetrics := range metrics.BackendMetrics {
			fmt.Printf("  %s:\n", backendID)
			fmt.Printf("    Requests: %d\n", backendMetrics.Requests)
			fmt.Printf("    Successful: %d\n", backendMetrics.SuccessfulRequests)
			fmt.Printf("    Failed: %d\n", backendMetrics.FailedRequests)
			fmt.Printf("    Avg Response Time: %s\n", backendMetrics.AverageResponseTime)
		}
	}

	// Example: Integrate proxy with server
	fmt.Println("\n" + repeat("=", 50))
	fmt.Println("Integration Example:")
	fmt.Println(repeat("=", 50))

	// Create server
	server := pkg.NewServer(pkg.ServerConfig{})

	// Add proxy route
	router := pkg.NewRouter()
	router.GET("/api/*", func(ctx pkg.Context) error {
		// Forward all /api/* requests to backend servers
		request := ctx.Request()

		response, err := proxyManager.Forward(ctx, request)
		if err != nil {
			return ctx.String(502, fmt.Sprintf("Proxy error: %v", err))
		}

		// Write response
		writer := ctx.Response()
		writer.WriteHeader(response.StatusCode)
		writer.Write(response.Body)

		return nil
	})

	server.SetRouter(router)
	fmt.Println("✓ Configured proxy route: GET /api/*")

	// Example: Remove a backend
	fmt.Println("\nRemoving backend2...")
	if err := proxyManager.RemoveBackend("backend2"); err != nil {
		log.Printf("Failed to remove backend: %v", err)
	} else {
		fmt.Println("✓ Removed backend2")
	}

	// List backends after removal
	fmt.Println("\nRemaining Backends:")
	backends = proxyManager.ListBackends()
	for _, backend := range backends {
		fmt.Printf("  - %s: %s\n", backend.ID, backend.URL.String())
	}

	// Example: Deactivate a backend
	fmt.Println("\nDeactivating backend3...")
	backend3.IsActive = false
	fmt.Println("✓ Deactivated backend3")

	// Example: Reset metrics
	fmt.Println("\nResetting metrics...")
	proxyManager.ResetMetrics()
	fmt.Println("✓ Metrics reset")

	metrics = proxyManager.GetMetrics()
	fmt.Printf("Total Requests after reset: %d\n", metrics.TotalRequests)

	fmt.Println("\n" + repeat("=", 50))
	fmt.Println("Proxy Configuration Complete!")
	fmt.Println(repeat("=", 50))

	// Note: In a real application, you would start the server here
	// server.Listen(":8080")
}

// Helper function to repeat strings (Go doesn't have built-in string repeat)
func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
