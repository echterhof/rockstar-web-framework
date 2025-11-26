//go:build ignore

package main

import (
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

// This example demonstrates forward proxy and load balancing in the Rockstar Web Framework.
// It shows how to configure backend servers, implement load balancing strategies,
// use circuit breakers for fault tolerance, and monitor proxy performance.

func main() {
	fmt.Println("🎸 Rockstar Web Framework - Forward Proxy Example\n")

	// Configure framework
	config := pkg.FrameworkConfig{
		ServerConfig: pkg.ServerConfig{
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			EnableHTTP1:  true,
		},
		// Configure database (required by framework)
		DatabaseConfig: pkg.DatabaseConfig{
			Driver:   "sqlite",
			Database: "proxy_example.db",
		},
		// Configure proxy with load balancing and circuit breaker
		ProxyConfig: pkg.ProxyConfig{
			LoadBalancerType:        "round_robin",
			CircuitBreakerEnabled:   true,
			CircuitBreakerThreshold: 5,
			MaxRetries:              3,
			CacheEnabled:            true,
			CacheTTL:                5 * time.Minute,
			HealthCheckEnabled:      true,
			HealthCheckInterval:     10 * time.Second,
		},
	}

	// Create framework instance
	app, err := pkg.New(config)
	if err != nil {
		log.Fatalf("Failed to create framework: %v", err)
	}

	// Get proxy manager
	proxy := app.Proxy()

	fmt.Println("1. Configuring Backend Servers")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	configureBackends(proxy)

	fmt.Println("\n2. Load Balancing Configuration")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	configureLoadBalancing(proxy)

	fmt.Println("\n3. Circuit Breaker Configuration")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	configureCircuitBreaker(proxy)

	fmt.Println("\n4. Health Check Status")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	checkBackendHealth(proxy)

	fmt.Println("\n5. Connection Pool Statistics")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	showConnectionPoolStats(proxy)

	fmt.Println("\n6. Proxy Metrics")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	showProxyMetrics(proxy)

	// Set up HTTP routes with proxy
	router := app.Router()

	// Proxy all /api/* requests to backend servers
	router.GET("/api/*", func(ctx pkg.Context) error {
		request := ctx.Request()

		// Forward request through proxy
		response, err := proxy.Forward(ctx, request)
		if err != nil {
			return ctx.JSON(502, map[string]interface{}{
				"error": fmt.Sprintf("Proxy error: %v", err),
			})
		}

		// Return proxied response
		writer := ctx.Response()
		writer.WriteHeader(response.StatusCode)
		writer.Write(response.Body)

		return nil
	})

	// Proxy status endpoint
	router.GET("/proxy/status", func(ctx pkg.Context) error {
		backends := proxy.ListBackends()
		healthStatus := proxy.GetHealthStatus()
		metrics := proxy.GetMetrics()

		status := map[string]interface{}{
			"backends":      backends,
			"health":        healthStatus,
			"metrics":       metrics,
			"load_balancer": proxy.GetLoadBalancer().Type(),
		}

		return ctx.JSON(200, status)
	})

	// Backend management endpoints
	router.POST("/proxy/backends", func(ctx pkg.Context) error {
		// Add new backend (in production, parse from request body)
		return ctx.JSON(200, map[string]interface{}{
			"message": "Backend management endpoint",
		})
	})

	router.DELETE("/proxy/backends/:id", func(ctx pkg.Context) error {
		backendID := ctx.Params()["id"]

		if err := proxy.RemoveBackend(backendID); err != nil {
			return ctx.JSON(400, map[string]interface{}{
				"error": err.Error(),
			})
		}

		return ctx.JSON(200, map[string]interface{}{
			"message": fmt.Sprintf("Backend %s removed", backendID),
		})
	})

	fmt.Println("\n✅ Proxy configuration completed!")
	fmt.Println("\nStarting server on :8080...")
	fmt.Println("Try: curl http://localhost:8080/api/data")
	fmt.Println("Try: curl http://localhost:8080/proxy/status")
	fmt.Println()

	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// configureBackends sets up backend servers for the proxy
func configureBackends(proxy pkg.ProxyManager) {
	// Backend 1 - Primary server
	backend1URL, _ := url.Parse("http://backend1.example.com:8080")
	backend1 := &pkg.Backend{
		ID:                  "backend1",
		URL:                 backend1URL,
		Weight:              1,
		IsActive:            true,
		HealthCheckPath:     "/health",
		HealthCheckInterval: 10 * time.Second,
		HealthCheckTimeout:  5 * time.Second,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	if err := proxy.AddBackend(backend1); err != nil {
		log.Printf("Failed to add backend1: %v", err)
	} else {
		fmt.Printf("  ✓ Added backend1: %s\n", backend1.URL.String())
	}

	// Backend 2 - Secondary server
	backend2URL, _ := url.Parse("http://backend2.example.com:8080")
	backend2 := &pkg.Backend{
		ID:                  "backend2",
		URL:                 backend2URL,
		Weight:              1,
		IsActive:            true,
		HealthCheckPath:     "/health",
		HealthCheckInterval: 10 * time.Second,
		HealthCheckTimeout:  5 * time.Second,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	if err := proxy.AddBackend(backend2); err != nil {
		log.Printf("Failed to add backend2: %v", err)
	} else {
		fmt.Printf("  ✓ Added backend2: %s\n", backend2.URL.String())
	}

	// Backend 3 - High-capacity server (higher weight)
	backend3URL, _ := url.Parse("http://backend3.example.com:8080")
	backend3 := &pkg.Backend{
		ID:                  "backend3",
		URL:                 backend3URL,
		Weight:              2, // Higher weight for weighted load balancing
		IsActive:            true,
		HealthCheckPath:     "/health",
		HealthCheckInterval: 10 * time.Second,
		HealthCheckTimeout:  5 * time.Second,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	if err := proxy.AddBackend(backend3); err != nil {
		log.Printf("Failed to add backend3: %v", err)
	} else {
		fmt.Printf("  ✓ Added backend3: %s (weight: %d)\n", backend3.URL.String(), backend3.Weight)
	}

	// List all configured backends
	backends := proxy.ListBackends()
	fmt.Printf("\n  Total backends configured: %d\n", len(backends))
}

// configureLoadBalancing demonstrates load balancing configuration
func configureLoadBalancing(proxy pkg.ProxyManager) {
	lb := proxy.GetLoadBalancer()
	fmt.Printf("  Load Balancer Type: %s\n", lb.Type())

	fmt.Println("\n  Load balancing strategies available:")
	fmt.Println("    - round_robin: Distributes requests evenly")
	fmt.Println("    - weighted_round_robin: Distributes based on backend weights")
	fmt.Println("    - least_connections: Routes to backend with fewest active connections")

	fmt.Println("\n  Current distribution pattern:")
	backends := proxy.ListBackends()
	for i, backend := range backends {
		fmt.Printf("    Request %d → %s (weight: %d)\n", i+1, backend.ID, backend.Weight)
	}
}

// configureCircuitBreaker demonstrates circuit breaker configuration
func configureCircuitBreaker(proxy pkg.ProxyManager) {
	cb := proxy.GetCircuitBreaker()

	fmt.Println("  Circuit Breaker Pattern:")
	fmt.Println("    - Protects against cascading failures")
	fmt.Println("    - Opens circuit after threshold failures")
	fmt.Println("    - Automatically retries after timeout")

	fmt.Println("\n  Backend Circuit States:")
	backends := proxy.ListBackends()
	for _, backend := range backends {
		state := cb.GetState(backend.ID)
		stateIcon := getStateIcon(state)
		fmt.Printf("    %s %s: %s\n", stateIcon, backend.ID, state)
	}
}

// checkBackendHealth performs and displays health check results
func checkBackendHealth(proxy pkg.ProxyManager) {
	fmt.Println("  Performing health checks...")

	if err := proxy.HealthCheck(); err != nil {
		log.Printf("  Health check error: %v", err)
	}

	healthStatus := proxy.GetHealthStatus()

	fmt.Println("\n  Backend Health Status:")
	for backendID, health := range healthStatus {
		statusIcon := "✓"
		statusText := "Healthy"
		if !health.IsHealthy {
			statusIcon = "✗"
			statusText = "Unhealthy"
		}

		fmt.Printf("    %s %s: %s\n", statusIcon, backendID, statusText)
		fmt.Printf("      Last Check: %s\n", health.LastCheck.Format("15:04:05"))
		fmt.Printf("      Response Time: %v\n", health.ResponseTime)
		fmt.Printf("      Consecutive Fails: %d\n", health.ConsecutiveFails)

		if health.ErrorMessage != "" {
			fmt.Printf("      Error: %s\n", health.ErrorMessage)
		}
	}
}

// showConnectionPoolStats displays connection pool statistics
func showConnectionPoolStats(proxy pkg.ProxyManager) {
	pool := proxy.GetConnectionPool()
	stats := pool.Stats()

	fmt.Printf("  Total Connections: %d\n", stats.TotalConnections)
	fmt.Printf("  Active Connections: %d\n", stats.ActiveConnections)
	fmt.Printf("  Idle Connections: %d\n", stats.IdleConnections)

	if len(stats.PerBackend) > 0 {
		fmt.Println("\n  Per-Backend Connections:")
		for backendID, count := range stats.PerBackend {
			fmt.Printf("    %s: %d connections\n", backendID, count)
		}
	}
}

// showProxyMetrics displays proxy performance metrics
func showProxyMetrics(proxy pkg.ProxyManager) {
	metrics := proxy.GetMetrics()

	fmt.Printf("  Total Requests: %d\n", metrics.TotalRequests)
	fmt.Printf("  Successful: %d (%.1f%%)\n",
		metrics.SuccessfulRequests,
		calculatePercentage(metrics.SuccessfulRequests, metrics.TotalRequests))
	fmt.Printf("  Failed: %d (%.1f%%)\n",
		metrics.FailedRequests,
		calculatePercentage(metrics.FailedRequests, metrics.TotalRequests))
	fmt.Printf("  Average Response Time: %v\n", metrics.AverageResponseTime)

	fmt.Println("\n  Cache Performance:")
	fmt.Printf("    Hits: %d\n", metrics.CacheHits)
	fmt.Printf("    Misses: %d\n", metrics.CacheMisses)
	if metrics.CacheHits+metrics.CacheMisses > 0 {
		hitRate := float64(metrics.CacheHits) / float64(metrics.CacheHits+metrics.CacheMisses) * 100
		fmt.Printf("    Hit Rate: %.1f%%\n", hitRate)
	}

	fmt.Printf("\n  Total Retries: %d\n", metrics.TotalRetries)

	if len(metrics.BackendMetrics) > 0 {
		fmt.Println("\n  Per-Backend Metrics:")
		for backendID, backendMetrics := range metrics.BackendMetrics {
			fmt.Printf("    %s:\n", backendID)
			fmt.Printf("      Requests: %d\n", backendMetrics.Requests)
			fmt.Printf("      Success Rate: %.1f%%\n",
				calculatePercentage(backendMetrics.SuccessfulRequests, backendMetrics.Requests))
			fmt.Printf("      Avg Response Time: %v\n", backendMetrics.AverageResponseTime)
		}
	}
}

// Helper functions

func getStateIcon(state pkg.CircuitState) string {
	switch state {
	case pkg.CircuitStateClosed:
		return "✓"
	case pkg.CircuitStateOpen:
		return "✗"
	case pkg.CircuitStateHalfOpen:
		return "⚠"
	default:
		return "?"
	}
}

func calculatePercentage(part, total int64) float64 {
	if total == 0 {
		return 0
	}
	return float64(part) / float64(total) * 100
}

func repeat(s string, count int) string {
	return strings.Repeat(s, count)
}
