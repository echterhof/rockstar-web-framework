package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

// BenchmarkRockstarSimpleRoute benchmarks a simple GET route in Rockstar framework
func BenchmarkRockstarSimpleRoute(b *testing.B) {
	config := pkg.ServerConfig{
		EnableHTTP1:  true,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	server := pkg.NewServer(config)
	router := pkg.NewRouter()

	router.GET("/api/hello", func(ctx pkg.Context) error {
		return ctx.String(http.StatusOK, "Hello, World!")
	})

	server.SetRouter(router)

	addr := "127.0.0.1:19301"
	if err := server.Listen(addr); err != nil {
		b.Fatalf("Failed to start server: %v", err)
	}
	defer server.Close()

	time.Sleep(100 * time.Millisecond)

	client := &http.Client{}
	url := "http://" + addr + "/api/hello"

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := client.Get(url)
			if err != nil {
				b.Fatalf("Request failed: %v", err)
			}
			io.ReadAll(resp.Body)
			resp.Body.Close()
		}
	})
}

// BenchmarkRockstarJSONResponse benchmarks JSON response handling
func BenchmarkRockstarJSONResponse(b *testing.B) {
	config := pkg.ServerConfig{
		EnableHTTP1:  true,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	server := pkg.NewServer(config)
	router := pkg.NewRouter()

	type User struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		Email    string `json:"email"`
		IsActive bool   `json:"is_active"`
	}

	router.GET("/api/user", func(ctx pkg.Context) error {
		user := User{
			ID:       1,
			Name:     "John Doe",
			Email:    "john@example.com",
			IsActive: true,
		}
		return ctx.JSON(http.StatusOK, user)
	})

	server.SetRouter(router)

	addr := "127.0.0.1:19302"
	if err := server.Listen(addr); err != nil {
		b.Fatalf("Failed to start server: %v", err)
	}
	defer server.Close()

	time.Sleep(100 * time.Millisecond)

	client := &http.Client{}
	url := "http://" + addr + "/api/user"

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := client.Get(url)
			if err != nil {
				b.Fatalf("Request failed: %v", err)
			}
			io.ReadAll(resp.Body)
			resp.Body.Close()
		}
	})
}

// BenchmarkRockstarRouteParams benchmarks route with parameters
func BenchmarkRockstarRouteParams(b *testing.B) {
	config := pkg.ServerConfig{
		EnableHTTP1:  true,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	server := pkg.NewServer(config)
	router := pkg.NewRouter()

	router.GET("/api/users/:id/posts/:postId", func(ctx pkg.Context) error {
		params := ctx.Params()
		return ctx.JSON(http.StatusOK, map[string]string{
			"user_id": params["id"],
			"post_id": params["postId"],
		})
	})

	server.SetRouter(router)

	addr := "127.0.0.1:19303"
	if err := server.Listen(addr); err != nil {
		b.Fatalf("Failed to start server: %v", err)
	}
	defer server.Close()

	time.Sleep(100 * time.Millisecond)

	client := &http.Client{}
	url := "http://" + addr + "/api/users/123/posts/456"

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := client.Get(url)
			if err != nil {
				b.Fatalf("Request failed: %v", err)
			}
			io.ReadAll(resp.Body)
			resp.Body.Close()
		}
	})
}

// BenchmarkRockstarPOSTRequest benchmarks POST request handling
func BenchmarkRockstarPOSTRequest(b *testing.B) {
	config := pkg.ServerConfig{
		EnableHTTP1:  true,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	server := pkg.NewServer(config)
	router := pkg.NewRouter()

	router.POST("/api/users", func(ctx pkg.Context) error {
		var user map[string]interface{}
		if err := json.Unmarshal(ctx.Body(), &user); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
		}
		user["id"] = 1
		return ctx.JSON(http.StatusCreated, user)
	})

	server.SetRouter(router)

	addr := "127.0.0.1:19304"
	if err := server.Listen(addr); err != nil {
		b.Fatalf("Failed to start server: %v", err)
	}
	defer server.Close()

	time.Sleep(100 * time.Millisecond)

	client := &http.Client{}
	url := "http://" + addr + "/api/users"

	userData := map[string]string{
		"name":  "John Doe",
		"email": "john@example.com",
	}
	jsonData, _ := json.Marshal(userData)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				b.Fatalf("Request failed: %v", err)
			}
			io.ReadAll(resp.Body)
			resp.Body.Close()
		}
	})
}

// BenchmarkRockstarMiddleware benchmarks middleware execution
func BenchmarkRockstarMiddleware(b *testing.B) {
	config := pkg.ServerConfig{
		EnableHTTP1:  true,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	server := pkg.NewServer(config)
	router := pkg.NewRouter()

	// Add middleware
	middleware1 := func(ctx pkg.Context, next pkg.HandlerFunc) error {
		ctx.SetHeader("X-Middleware-1", "executed")
		return next(ctx)
	}

	middleware2 := func(ctx pkg.Context, next pkg.HandlerFunc) error {
		ctx.SetHeader("X-Middleware-2", "executed")
		return next(ctx)
	}

	router.GET("/api/test", func(ctx pkg.Context) error {
		return ctx.String(http.StatusOK, "OK")
	}, middleware1, middleware2)

	server.SetRouter(router)

	addr := "127.0.0.1:19305"
	if err := server.Listen(addr); err != nil {
		b.Fatalf("Failed to start server: %v", err)
	}
	defer server.Close()

	time.Sleep(100 * time.Millisecond)

	client := &http.Client{}
	url := "http://" + addr + "/api/test"

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := client.Get(url)
			if err != nil {
				b.Fatalf("Request failed: %v", err)
			}
			io.ReadAll(resp.Body)
			resp.Body.Close()
		}
	})
}

// BenchmarkRockstarMultipleRoutes benchmarks routing with many routes
func BenchmarkRockstarMultipleRoutes(b *testing.B) {
	config := pkg.ServerConfig{
		EnableHTTP1:  true,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	server := pkg.NewServer(config)
	router := pkg.NewRouter()

	// Register 100 routes
	for i := 0; i < 100; i++ {
		path := fmt.Sprintf("/api/route%d", i)
		router.GET(path, func(ctx pkg.Context) error {
			return ctx.String(http.StatusOK, "OK")
		})
	}

	server.SetRouter(router)

	addr := "127.0.0.1:19306"
	if err := server.Listen(addr); err != nil {
		b.Fatalf("Failed to start server: %v", err)
	}
	defer server.Close()

	time.Sleep(100 * time.Millisecond)

	client := &http.Client{}
	url := "http://" + addr + "/api/route50"

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := client.Get(url)
			if err != nil {
				b.Fatalf("Request failed: %v", err)
			}
			io.ReadAll(resp.Body)
			resp.Body.Close()
		}
	})
}

// BenchmarkRockstarConcurrentRequests benchmarks concurrent request handling
func BenchmarkRockstarConcurrentRequests(b *testing.B) {
	config := pkg.ServerConfig{
		EnableHTTP1:  true,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	server := pkg.NewServer(config)
	router := pkg.NewRouter()

	router.GET("/api/concurrent", func(ctx pkg.Context) error {
		// Simulate some work
		time.Sleep(1 * time.Millisecond)
		return ctx.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	server.SetRouter(router)

	addr := "127.0.0.1:19307"
	if err := server.Listen(addr); err != nil {
		b.Fatalf("Failed to start server: %v", err)
	}
	defer server.Close()

	time.Sleep(100 * time.Millisecond)

	client := &http.Client{}
	url := "http://" + addr + "/api/concurrent"

	b.ResetTimer()
	b.SetParallelism(100) // High concurrency
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := client.Get(url)
			if err != nil {
				b.Fatalf("Request failed: %v", err)
			}
			io.ReadAll(resp.Body)
			resp.Body.Close()
		}
	})
}

// BenchmarkRockstarRESTAPI benchmarks REST API operations
func BenchmarkRockstarRESTAPI(b *testing.B) {
	config := pkg.ServerConfig{
		EnableHTTP1:  true,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	server := pkg.NewServer(config)
	router := pkg.NewRouter()

	// Define RESTful routes
	router.GET("/api/posts", func(ctx pkg.Context) error {
		posts := []map[string]interface{}{
			{"id": 1, "title": "Post 1"},
			{"id": 2, "title": "Post 2"},
		}
		return ctx.JSON(http.StatusOK, posts)
	})

	router.GET("/api/posts/:id", func(ctx pkg.Context) error {
		id := ctx.Params()["id"]
		return ctx.JSON(http.StatusOK, map[string]interface{}{
			"id":    id,
			"title": "Post " + id,
		})
	})

	server.SetRouter(router)

	addr := "127.0.0.1:19308"
	if err := server.Listen(addr); err != nil {
		b.Fatalf("Failed to start server: %v", err)
	}
	defer server.Close()

	time.Sleep(100 * time.Millisecond)

	client := &http.Client{}
	url := "http://" + addr + "/api/posts"

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := client.Get(url)
			if err != nil {
				b.Fatalf("Request failed: %v", err)
			}
			io.ReadAll(resp.Body)
			resp.Body.Close()
		}
	})
}

// BenchmarkRockstarAuthentication benchmarks authentication overhead
func BenchmarkRockstarAuthentication(b *testing.B) {
	db := newTestMockDB()
	db.Connect(pkg.DatabaseConfig{Driver: "mock"})
	authManager := pkg.NewAuthManager(db, "test-secret", pkg.OAuth2Config{})

	// Create token
	token, _ := authManager.CreateAccessToken("user123", "tenant456", []string{"read"}, 1*time.Hour)

	config := pkg.ServerConfig{
		EnableHTTP1:  true,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	server := pkg.NewServer(config)
	router := pkg.NewRouter()

	router.GET("/api/protected", func(ctx pkg.Context) error {
		authHeader := ctx.Headers()["Authorization"]
		if authHeader == "" {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "missing token"})
		}

		user, err := authManager.AuthenticateOAuth2(authHeader)
		if err != nil {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		}

		return ctx.JSON(http.StatusOK, map[string]string{"user_id": user.ID})
	})

	server.SetRouter(router)

	addr := "127.0.0.1:19309"
	if err := server.Listen(addr); err != nil {
		b.Fatalf("Failed to start server: %v", err)
	}
	defer server.Close()

	time.Sleep(100 * time.Millisecond)

	client := &http.Client{}
	url := "http://" + addr + "/api/protected"

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req, _ := http.NewRequest("GET", url, nil)
			req.Header.Set("Authorization", token.Token)

			resp, err := client.Do(req)
			if err != nil {
				b.Fatalf("Request failed: %v", err)
			}
			io.ReadAll(resp.Body)
			resp.Body.Close()
		}
	})
}

// BenchmarkRockstarMemoryAllocation benchmarks memory allocation patterns
func BenchmarkRockstarMemoryAllocation(b *testing.B) {
	config := pkg.ServerConfig{
		EnableHTTP1:  true,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	server := pkg.NewServer(config)
	router := pkg.NewRouter()

	router.GET("/api/data", func(ctx pkg.Context) error {
		// Create some data structures
		data := make(map[string]interface{})
		data["items"] = []int{1, 2, 3, 4, 5}
		data["message"] = "Hello, World!"
		data["nested"] = map[string]string{
			"key1": "value1",
			"key2": "value2",
		}
		return ctx.JSON(http.StatusOK, data)
	})

	server.SetRouter(router)

	addr := "127.0.0.1:19310"
	if err := server.Listen(addr); err != nil {
		b.Fatalf("Failed to start server: %v", err)
	}
	defer server.Close()

	time.Sleep(100 * time.Millisecond)

	client := &http.Client{}
	url := "http://" + addr + "/api/data"

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := client.Get(url)
			if err != nil {
				b.Fatalf("Request failed: %v", err)
			}
			io.ReadAll(resp.Body)
			resp.Body.Close()
		}
	})
}

// Comparison benchmarks (conceptual - would need actual GoFiber/Gin implementations)

// BenchmarkComparison_SimpleRoute compares simple route performance
func BenchmarkComparison_SimpleRoute(b *testing.B) {
	b.Run("Rockstar", func(b *testing.B) {
		BenchmarkRockstarSimpleRoute(b)
	})

	// Note: To add GoFiber and Gin benchmarks, you would need to:
	// 1. Import those frameworks
	// 2. Set up equivalent servers
	// 3. Run the same benchmark tests
	//
	// b.Run("GoFiber", func(b *testing.B) {
	//     BenchmarkGoFiberSimpleRoute(b)
	// })
	//
	// b.Run("Gin", func(b *testing.B) {
	//     BenchmarkGinSimpleRoute(b)
	// })
}

// BenchmarkComparison_JSONResponse compares JSON response performance
func BenchmarkComparison_JSONResponse(b *testing.B) {
	b.Run("Rockstar", func(b *testing.B) {
		BenchmarkRockstarJSONResponse(b)
	})

	// Add GoFiber and Gin comparisons here
}

// BenchmarkComparison_RouteParams compares route parameter extraction
func BenchmarkComparison_RouteParams(b *testing.B) {
	b.Run("Rockstar", func(b *testing.B) {
		BenchmarkRockstarRouteParams(b)
	})

	// Add GoFiber and Gin comparisons here
}

// BenchmarkComparison_Middleware compares middleware execution
func BenchmarkComparison_Middleware(b *testing.B) {
	b.Run("Rockstar", func(b *testing.B) {
		BenchmarkRockstarMiddleware(b)
	})

	// Add GoFiber and Gin comparisons here
}

// Performance metrics helper
type PerformanceMetrics struct {
	RequestsPerSecond float64
	AvgLatency        time.Duration
	P95Latency        time.Duration
	P99Latency        time.Duration
	MemoryUsage       uint64
	AllocsPerOp       uint64
}

// MeasurePerformance measures detailed performance metrics
func MeasurePerformance(b *testing.B, url string) PerformanceMetrics {
	client := &http.Client{}
	latencies := make([]time.Duration, b.N)

	start := time.Now()

	for i := 0; i < b.N; i++ {
		reqStart := time.Now()
		resp, err := client.Get(url)
		if err != nil {
			b.Fatalf("Request failed: %v", err)
		}
		io.ReadAll(resp.Body)
		resp.Body.Close()
		latencies[i] = time.Since(reqStart)
	}

	duration := time.Since(start)

	// Calculate metrics
	metrics := PerformanceMetrics{
		RequestsPerSecond: float64(b.N) / duration.Seconds(),
	}

	// Calculate average latency
	var totalLatency time.Duration
	for _, lat := range latencies {
		totalLatency += lat
	}
	metrics.AvgLatency = totalLatency / time.Duration(b.N)

	return metrics
}
