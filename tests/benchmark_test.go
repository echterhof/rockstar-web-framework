//go:build benchmark
// +build benchmark

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

// BenchmarkRockstarSimpleRoute benchmarks simple GET request handling
func BenchmarkRockstarSimpleRoute(b *testing.B) {
	// Create framework with minimal config
	dbConfig := createTestDatabaseConfig()
	framework, err := pkg.New(pkg.FrameworkConfig{
		DatabaseConfig: dbConfig,
		SessionConfig:  *createTestSessionConfig(),
	})
	if err != nil {
		b.Fatalf("Failed to create framework: %v", err)
	}
	defer framework.Shutdown(2 * time.Second)

	// Register simple route
	framework.Router().GET("/hello", func(ctx pkg.Context) error {
		return ctx.String(200, "Hello, World!")
	})

	// Start server in background
	go func() {
		if err := framework.Listen(":19301"); err != nil {
			b.Logf("Server error: %v", err)
		}
	}()
	time.Sleep(100 * time.Millisecond)

	// Reset timer before benchmark
	b.ResetTimer()
	b.ReportAllocs()

	// Run benchmark
	b.RunParallel(func(pb *testing.PB) {
		client := &http.Client{}
		for pb.Next() {
			resp, err := client.Get("http://localhost:19301/hello")
			if err != nil {
				b.Errorf("Request failed: %v", err)
				continue
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	})
}

// BenchmarkRockstarJSONResponse benchmarks JSON serialization
func BenchmarkRockstarJSONResponse(b *testing.B) {
	// Create framework with minimal config
	dbConfig := createTestDatabaseConfig()
	framework, err := pkg.New(pkg.FrameworkConfig{
		DatabaseConfig: dbConfig,
		SessionConfig:  *createTestSessionConfig(),
	})
	if err != nil {
		b.Fatalf("Failed to create framework: %v", err)
	}
	defer framework.Shutdown(2 * time.Second)

	// Register JSON route
	type Response struct {
		Message string `json:"message"`
		Status  string `json:"status"`
		Code    int    `json:"code"`
	}

	framework.Router().GET("/json", func(ctx pkg.Context) error {
		return ctx.JSON(200, Response{
			Message: "Hello, World!",
			Status:  "success",
			Code:    200,
		})
	})

	// Start server in background
	go func() {
		if err := framework.Listen(":19302"); err != nil {
			b.Logf("Server error: %v", err)
		}
	}()
	time.Sleep(100 * time.Millisecond)

	// Reset timer before benchmark
	b.ResetTimer()
	b.ReportAllocs()

	// Run benchmark
	b.RunParallel(func(pb *testing.PB) {
		client := &http.Client{}
		for pb.Next() {
			resp, err := client.Get("http://localhost:19302/json")
			if err != nil {
				b.Errorf("Request failed: %v", err)
				continue
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	})
}

// BenchmarkRockstarRouteParams benchmarks route parameter extraction
func BenchmarkRockstarRouteParams(b *testing.B) {
	// Create framework with minimal config
	dbConfig := createTestDatabaseConfig()
	framework, err := pkg.New(pkg.FrameworkConfig{
		DatabaseConfig: dbConfig,
		SessionConfig:  *createTestSessionConfig(),
	})
	if err != nil {
		b.Fatalf("Failed to create framework: %v", err)
	}
	defer framework.Shutdown(2 * time.Second)

	// Register route with parameters
	framework.Router().GET("/users/:id/posts/:postId", func(ctx pkg.Context) error {
		params := ctx.Params()
		return ctx.String(200, fmt.Sprintf("User: %s, Post: %s", params["id"], params["postId"]))
	})

	// Start server in background
	go func() {
		if err := framework.Listen(":19303"); err != nil {
			b.Logf("Server error: %v", err)
		}
	}()
	time.Sleep(100 * time.Millisecond)

	// Reset timer before benchmark
	b.ResetTimer()
	b.ReportAllocs()

	// Run benchmark
	b.RunParallel(func(pb *testing.PB) {
		client := &http.Client{}
		for pb.Next() {
			resp, err := client.Get("http://localhost:19303/users/123/posts/456")
			if err != nil {
				b.Errorf("Request failed: %v", err)
				continue
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	})
}

// BenchmarkRockstarPOSTRequest benchmarks POST request handling
func BenchmarkRockstarPOSTRequest(b *testing.B) {
	// Create framework with minimal config
	dbConfig := createTestDatabaseConfig()
	framework, err := pkg.New(pkg.FrameworkConfig{
		DatabaseConfig: dbConfig,
		SessionConfig:  *createTestSessionConfig(),
	})
	if err != nil {
		b.Fatalf("Failed to create framework: %v", err)
	}
	defer framework.Shutdown(2 * time.Second)

	// Register POST route
	type RequestBody struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	framework.Router().POST("/users", func(ctx pkg.Context) error {
		var body RequestBody
		if err := json.Unmarshal(ctx.Body(), &body); err != nil {
			return ctx.String(400, "Invalid request")
		}
		return ctx.JSON(201, map[string]string{
			"id":    "123",
			"name":  body.Name,
			"email": body.Email,
		})
	})

	// Start server in background
	go func() {
		if err := framework.Listen(":19304"); err != nil {
			b.Logf("Server error: %v", err)
		}
	}()
	time.Sleep(100 * time.Millisecond)

	// Prepare request body
	reqBody := RequestBody{Name: "John Doe", Email: "john@example.com"}
	bodyBytes, _ := json.Marshal(reqBody)

	// Reset timer before benchmark
	b.ResetTimer()
	b.ReportAllocs()

	// Run benchmark
	b.RunParallel(func(pb *testing.PB) {
		client := &http.Client{}
		for pb.Next() {
			req, err := http.NewRequest("POST", "http://localhost:19304/users", bytes.NewReader(bodyBytes))
			if err != nil {
				b.Errorf("Request creation failed: %v", err)
				continue
			}
			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			if err != nil {
				b.Errorf("Request failed: %v", err)
				continue
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	})
}

// BenchmarkRockstarMiddleware benchmarks middleware execution overhead
func BenchmarkRockstarMiddleware(b *testing.B) {
	// Create framework with minimal config
	dbConfig := createTestDatabaseConfig()
	framework, err := pkg.New(pkg.FrameworkConfig{
		DatabaseConfig: dbConfig,
		SessionConfig:  *createTestSessionConfig(),
	})
	if err != nil {
		b.Fatalf("Failed to create framework: %v", err)
	}
	defer framework.Shutdown(2 * time.Second)

	// Add multiple middleware
	middleware1 := func(ctx pkg.Context, next pkg.HandlerFunc) error {
		ctx.SetHeader("X-Middleware-1", "true")
		return next(ctx)
	}

	middleware2 := func(ctx pkg.Context, next pkg.HandlerFunc) error {
		ctx.SetHeader("X-Middleware-2", "true")
		return next(ctx)
	}

	middleware3 := func(ctx pkg.Context, next pkg.HandlerFunc) error {
		ctx.SetHeader("X-Middleware-3", "true")
		return next(ctx)
	}

	// Register route with middleware
	framework.Router().GET("/middleware", func(ctx pkg.Context) error {
		return ctx.String(200, "OK")
	}, middleware1, middleware2, middleware3)

	// Start server in background
	go func() {
		if err := framework.Listen(":19305"); err != nil {
			b.Logf("Server error: %v", err)
		}
	}()
	time.Sleep(100 * time.Millisecond)

	// Reset timer before benchmark
	b.ResetTimer()
	b.ReportAllocs()

	// Run benchmark
	b.RunParallel(func(pb *testing.PB) {
		client := &http.Client{}
		for pb.Next() {
			resp, err := client.Get("http://localhost:19305/middleware")
			if err != nil {
				b.Errorf("Request failed: %v", err)
				continue
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	})
}

// BenchmarkRockstarMultipleRoutes benchmarks routing with many routes
func BenchmarkRockstarMultipleRoutes(b *testing.B) {
	// Create framework with minimal config
	dbConfig := createTestDatabaseConfig()
	framework, err := pkg.New(pkg.FrameworkConfig{
		DatabaseConfig: dbConfig,
		SessionConfig:  *createTestSessionConfig(),
	})
	if err != nil {
		b.Fatalf("Failed to create framework: %v", err)
	}
	defer framework.Shutdown(2 * time.Second)

	// Register many routes
	for i := 0; i < 100; i++ {
		path := fmt.Sprintf("/route%d", i)
		framework.Router().GET(path, func(ctx pkg.Context) error {
			return ctx.String(200, "OK")
		})
	}

	// Register target route
	framework.Router().GET("/target", func(ctx pkg.Context) error {
		return ctx.String(200, "Target")
	})

	// Start server in background
	go func() {
		if err := framework.Listen(":19306"); err != nil {
			b.Logf("Server error: %v", err)
		}
	}()
	time.Sleep(100 * time.Millisecond)

	// Reset timer before benchmark
	b.ResetTimer()
	b.ReportAllocs()

	// Run benchmark
	b.RunParallel(func(pb *testing.PB) {
		client := &http.Client{}
		for pb.Next() {
			resp, err := client.Get("http://localhost:19306/target")
			if err != nil {
				b.Errorf("Request failed: %v", err)
				continue
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	})
}

// BenchmarkRockstarConcurrentRequests benchmarks concurrent request handling
func BenchmarkRockstarConcurrentRequests(b *testing.B) {
	// Create framework with minimal config
	dbConfig := createTestDatabaseConfig()
	framework, err := pkg.New(pkg.FrameworkConfig{
		DatabaseConfig: dbConfig,
		SessionConfig:  *createTestSessionConfig(),
	})
	if err != nil {
		b.Fatalf("Failed to create framework: %v", err)
	}
	defer framework.Shutdown(2 * time.Second)

	// Register route
	framework.Router().GET("/concurrent", func(ctx pkg.Context) error {
		return ctx.String(200, "OK")
	})

	// Start server in background
	go func() {
		if err := framework.Listen(":19307"); err != nil {
			b.Logf("Server error: %v", err)
		}
	}()

	// Wait for server to be ready
	time.Sleep(200 * time.Millisecond)
	client := &http.Client{Timeout: 5 * time.Second}
	for i := 0; i < 10; i++ {
		resp, err := client.Get("http://localhost:19307/concurrent")
		if err == nil {
			resp.Body.Close()
			break
		}
		if i == 9 {
			b.Fatalf("Server failed to start: %v", err)
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Reset timer before benchmark
	b.ResetTimer()
	b.ReportAllocs()

	// Run benchmark with moderate parallelism
	b.SetParallelism(10)
	b.RunParallel(func(pb *testing.PB) {
		client := &http.Client{}
		for pb.Next() {
			resp, err := client.Get("http://localhost:19307/concurrent")
			if err != nil {
				b.Errorf("Request failed: %v", err)
				continue
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	})
}

// BenchmarkRockstarRESTAPI benchmarks REST operations
func BenchmarkRockstarRESTAPI(b *testing.B) {
	// Create framework with minimal config
	dbConfig := createTestDatabaseConfig()
	framework, err := pkg.New(pkg.FrameworkConfig{
		DatabaseConfig: dbConfig,
		SessionConfig:  *createTestSessionConfig(),
	})
	if err != nil {
		b.Fatalf("Failed to create framework: %v", err)
	}
	defer framework.Shutdown(2 * time.Second)

	// In-memory data store
	type Item struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	items := make(map[string]Item)
	items["1"] = Item{ID: "1", Name: "Item 1"}
	items["2"] = Item{ID: "2", Name: "Item 2"}

	// Register REST routes
	framework.Router().GET("/api/items", func(ctx pkg.Context) error {
		itemList := make([]Item, 0, len(items))
		for _, item := range items {
			itemList = append(itemList, item)
		}
		return ctx.JSON(200, itemList)
	})

	framework.Router().GET("/api/items/:id", func(ctx pkg.Context) error {
		id := ctx.Params()["id"]
		item, exists := items[id]
		if !exists {
			return ctx.String(404, "Not found")
		}
		return ctx.JSON(200, item)
	})

	framework.Router().POST("/api/items", func(ctx pkg.Context) error {
		var item Item
		if err := json.Unmarshal(ctx.Body(), &item); err != nil {
			return ctx.String(400, "Invalid request")
		}
		items[item.ID] = item
		return ctx.JSON(201, item)
	})

	// Start server in background
	go func() {
		if err := framework.Listen(":19308"); err != nil {
			b.Logf("Server error: %v", err)
		}
	}()
	time.Sleep(100 * time.Millisecond)

	// Reset timer before benchmark
	b.ResetTimer()
	b.ReportAllocs()

	// Run benchmark
	b.RunParallel(func(pb *testing.PB) {
		client := &http.Client{}
		for pb.Next() {
			// Mix of GET and POST requests
			resp, err := client.Get("http://localhost:19308/api/items/1")
			if err != nil {
				b.Errorf("Request failed: %v", err)
				continue
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	})
}

// BenchmarkRockstarAuthentication benchmarks authentication overhead
func BenchmarkRockstarAuthentication(b *testing.B) {
	// Create framework with minimal config
	dbConfig := createTestDatabaseConfig()
	framework, err := pkg.New(pkg.FrameworkConfig{
		DatabaseConfig: dbConfig,
		SessionConfig:  *createTestSessionConfig(),
	})
	if err != nil {
		b.Fatalf("Failed to create framework: %v", err)
	}
	defer framework.Shutdown(2 * time.Second)

	// Create test token
	db := newTestMockDB()
	token := createTestAccessToken("test-token-123", "user-1", "tenant-1", []string{"read", "write"}, 1*time.Hour)
	db.SaveAccessToken(token)

	// Authentication middleware
	authMiddleware := func(ctx pkg.Context, next pkg.HandlerFunc) error {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			return ctx.String(401, "Unauthorized")
		}

		// Validate token
		tokenValue := authHeader[len("Bearer "):]
		_, err := db.ValidateAccessToken(tokenValue)
		if err != nil {
			return ctx.String(401, "Invalid token")
		}

		return next(ctx)
	}

	// Register protected route
	framework.Router().GET("/protected", func(ctx pkg.Context) error {
		return ctx.String(200, "Protected resource")
	}, authMiddleware)

	// Start server in background
	go func() {
		if err := framework.Listen(":19309"); err != nil {
			b.Logf("Server error: %v", err)
		}
	}()
	time.Sleep(100 * time.Millisecond)

	// Reset timer before benchmark
	b.ResetTimer()
	b.ReportAllocs()

	// Run benchmark
	b.RunParallel(func(pb *testing.PB) {
		client := &http.Client{}
		for pb.Next() {
			req, err := http.NewRequest("GET", "http://localhost:19309/protected", nil)
			if err != nil {
				b.Errorf("Request creation failed: %v", err)
				continue
			}
			req.Header.Set("Authorization", "Bearer test-token-123")

			resp, err := client.Do(req)
			if err != nil {
				b.Errorf("Request failed: %v", err)
				continue
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	})
}

// BenchmarkRockstarMemoryAllocation benchmarks memory allocation patterns
func BenchmarkRockstarMemoryAllocation(b *testing.B) {
	// Create framework with minimal config
	dbConfig := createTestDatabaseConfig()
	framework, err := pkg.New(pkg.FrameworkConfig{
		DatabaseConfig: dbConfig,
		SessionConfig:  *createTestSessionConfig(),
	})
	if err != nil {
		b.Fatalf("Failed to create framework: %v", err)
	}
	defer framework.Shutdown(2 * time.Second)

	// Register route
	framework.Router().GET("/memory", func(ctx pkg.Context) error {
		// Simulate typical request processing
		data := make(map[string]interface{})
		data["message"] = "Hello, World!"
		data["timestamp"] = time.Now().Unix()
		data["status"] = "success"
		return ctx.JSON(200, data)
	})

	// Start server in background
	go func() {
		if err := framework.Listen(":19310"); err != nil {
			b.Logf("Server error: %v", err)
		}
	}()
	time.Sleep(100 * time.Millisecond)

	// Reset timer before benchmark
	b.ResetTimer()
	b.ReportAllocs()

	// Run benchmark
	b.RunParallel(func(pb *testing.PB) {
		client := &http.Client{}
		for pb.Next() {
			resp, err := client.Get("http://localhost:19310/memory")
			if err != nil {
				b.Errorf("Request failed: %v", err)
				continue
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	})
}

// Comparison Benchmarks
// These benchmarks provide a structure for comparing Rockstar with other frameworks
// To add comparisons with GoFiber or Gin, implement the corresponding sub-benchmarks

// BenchmarkComparison_SimpleRoute compares simple route performance across frameworks
func BenchmarkComparison_SimpleRoute(b *testing.B) {
	b.Run("Rockstar", func(b *testing.B) {
		// Create framework with minimal config
		dbConfig := createTestDatabaseConfig()
		framework, err := pkg.New(pkg.FrameworkConfig{
			DatabaseConfig: dbConfig,
			SessionConfig:  *createTestSessionConfig(),
		})
		if err != nil {
			b.Fatalf("Failed to create framework: %v", err)
		}
		defer framework.Shutdown(2 * time.Second)

		// Register simple route
		framework.Router().GET("/hello", func(ctx pkg.Context) error {
			return ctx.String(200, "Hello, World!")
		})

		// Start server in background
		go func() {
			if err := framework.Listen(":19311"); err != nil {
				b.Logf("Server error: %v", err)
			}
		}()
		time.Sleep(100 * time.Millisecond)

		// Reset timer before benchmark
		b.ResetTimer()
		b.ReportAllocs()

		// Run benchmark
		b.RunParallel(func(pb *testing.PB) {
			client := &http.Client{}
			for pb.Next() {
				resp, err := client.Get("http://localhost:19311/hello")
				if err != nil {
					b.Errorf("Request failed: %v", err)
					continue
				}
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
			}
		})
	})

	// TODO: Add GoFiber comparison
	// b.Run("GoFiber", func(b *testing.B) {
	//     // Implement GoFiber benchmark here
	//     // app := fiber.New()
	//     // app.Get("/hello", func(c *fiber.Ctx) error {
	//     //     return c.SendString("Hello, World!")
	//     // })
	//     // go app.Listen(":19312")
	//     // ... benchmark code ...
	// })

	// TODO: Add Gin comparison
	// b.Run("Gin", func(b *testing.B) {
	//     // Implement Gin benchmark here
	//     // router := gin.New()
	//     // router.GET("/hello", func(c *gin.Context) {
	//     //     c.String(200, "Hello, World!")
	//     // })
	//     // go router.Run(":19313")
	//     // ... benchmark code ...
	// })
}

// BenchmarkComparison_JSONResponse compares JSON response performance across frameworks
func BenchmarkComparison_JSONResponse(b *testing.B) {
	b.Run("Rockstar", func(b *testing.B) {
		// Create framework with minimal config
		dbConfig := createTestDatabaseConfig()
		framework, err := pkg.New(pkg.FrameworkConfig{
			DatabaseConfig: dbConfig,
			SessionConfig:  *createTestSessionConfig(),
		})
		if err != nil {
			b.Fatalf("Failed to create framework: %v", err)
		}
		defer framework.Shutdown(2 * time.Second)

		// Register JSON route
		type Response struct {
			Message string `json:"message"`
			Status  string `json:"status"`
			Code    int    `json:"code"`
		}

		framework.Router().GET("/json", func(ctx pkg.Context) error {
			return ctx.JSON(200, Response{
				Message: "Hello, World!",
				Status:  "success",
				Code:    200,
			})
		})

		// Start server in background
		go func() {
			if err := framework.Listen(":19314"); err != nil {
				b.Logf("Server error: %v", err)
			}
		}()
		time.Sleep(100 * time.Millisecond)

		// Reset timer before benchmark
		b.ResetTimer()
		b.ReportAllocs()

		// Run benchmark
		b.RunParallel(func(pb *testing.PB) {
			client := &http.Client{}
			for pb.Next() {
				resp, err := client.Get("http://localhost:19314/json")
				if err != nil {
					b.Errorf("Request failed: %v", err)
					continue
				}
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
			}
		})
	})

	// TODO: Add GoFiber comparison
	// b.Run("GoFiber", func(b *testing.B) {
	//     // Implement GoFiber JSON benchmark here
	//     // app := fiber.New()
	//     // app.Get("/json", func(c *fiber.Ctx) error {
	//     //     return c.JSON(Response{...})
	//     // })
	//     // ... benchmark code ...
	// })

	// TODO: Add Gin comparison
	// b.Run("Gin", func(b *testing.B) {
	//     // Implement Gin JSON benchmark here
	//     // router := gin.New()
	//     // router.GET("/json", func(c *gin.Context) {
	//     //     c.JSON(200, Response{...})
	//     // })
	//     // ... benchmark code ...
	// })
}

// BenchmarkComparison_RouteParams compares route parameter extraction across frameworks
func BenchmarkComparison_RouteParams(b *testing.B) {
	b.Run("Rockstar", func(b *testing.B) {
		// Create framework with minimal config
		dbConfig := createTestDatabaseConfig()
		framework, err := pkg.New(pkg.FrameworkConfig{
			DatabaseConfig: dbConfig,
			SessionConfig:  *createTestSessionConfig(),
		})
		if err != nil {
			b.Fatalf("Failed to create framework: %v", err)
		}
		defer framework.Shutdown(2 * time.Second)

		// Register route with parameters
		framework.Router().GET("/users/:id/posts/:postId", func(ctx pkg.Context) error {
			params := ctx.Params()
			return ctx.String(200, fmt.Sprintf("User: %s, Post: %s", params["id"], params["postId"]))
		})

		// Start server in background
		go func() {
			if err := framework.Listen(":19315"); err != nil {
				b.Logf("Server error: %v", err)
			}
		}()
		time.Sleep(100 * time.Millisecond)

		// Reset timer before benchmark
		b.ResetTimer()
		b.ReportAllocs()

		// Run benchmark
		b.RunParallel(func(pb *testing.PB) {
			client := &http.Client{}
			for pb.Next() {
				resp, err := client.Get("http://localhost:19315/users/123/posts/456")
				if err != nil {
					b.Errorf("Request failed: %v", err)
					continue
				}
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
			}
		})
	})

	// TODO: Add GoFiber comparison
	// b.Run("GoFiber", func(b *testing.B) {
	//     // Implement GoFiber route params benchmark here
	//     // app := fiber.New()
	//     // app.Get("/users/:id/posts/:postId", func(c *fiber.Ctx) error {
	//     //     return c.SendString(fmt.Sprintf("User: %s, Post: %s", c.Params("id"), c.Params("postId")))
	//     // })
	//     // ... benchmark code ...
	// })

	// TODO: Add Gin comparison
	// b.Run("Gin", func(b *testing.B) {
	//     // Implement Gin route params benchmark here
	//     // router := gin.New()
	//     // router.GET("/users/:id/posts/:postId", func(c *gin.Context) {
	//     //     c.String(200, fmt.Sprintf("User: %s, Post: %s", c.Param("id"), c.Param("postId")))
	//     // })
	//     // ... benchmark code ...
	// })
}

// BenchmarkComparison_Middleware compares middleware execution overhead across frameworks
func BenchmarkComparison_Middleware(b *testing.B) {
	b.Run("Rockstar", func(b *testing.B) {
		// Create framework with minimal config
		dbConfig := createTestDatabaseConfig()
		framework, err := pkg.New(pkg.FrameworkConfig{
			DatabaseConfig: dbConfig,
			SessionConfig:  *createTestSessionConfig(),
		})
		if err != nil {
			b.Fatalf("Failed to create framework: %v", err)
		}
		defer framework.Shutdown(2 * time.Second)

		// Add multiple middleware
		middleware1 := func(ctx pkg.Context, next pkg.HandlerFunc) error {
			ctx.SetHeader("X-Middleware-1", "true")
			return next(ctx)
		}

		middleware2 := func(ctx pkg.Context, next pkg.HandlerFunc) error {
			ctx.SetHeader("X-Middleware-2", "true")
			return next(ctx)
		}

		middleware3 := func(ctx pkg.Context, next pkg.HandlerFunc) error {
			ctx.SetHeader("X-Middleware-3", "true")
			return next(ctx)
		}

		// Register route with middleware
		framework.Router().GET("/middleware", func(ctx pkg.Context) error {
			return ctx.String(200, "OK")
		}, middleware1, middleware2, middleware3)

		// Start server in background
		go func() {
			if err := framework.Listen(":19316"); err != nil {
				b.Logf("Server error: %v", err)
			}
		}()
		time.Sleep(100 * time.Millisecond)

		// Reset timer before benchmark
		b.ResetTimer()
		b.ReportAllocs()

		// Run benchmark
		b.RunParallel(func(pb *testing.PB) {
			client := &http.Client{}
			for pb.Next() {
				resp, err := client.Get("http://localhost:19316/middleware")
				if err != nil {
					b.Errorf("Request failed: %v", err)
					continue
				}
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
			}
		})
	})

	// TODO: Add GoFiber comparison
	// b.Run("GoFiber", func(b *testing.B) {
	//     // Implement GoFiber middleware benchmark here
	//     // app := fiber.New()
	//     // app.Use(func(c *fiber.Ctx) error {
	//     //     c.Set("X-Middleware-1", "true")
	//     //     return c.Next()
	//     // })
	//     // ... add more middleware ...
	//     // app.Get("/middleware", func(c *fiber.Ctx) error {
	//     //     return c.SendString("OK")
	//     // })
	//     // ... benchmark code ...
	// })

	// TODO: Add Gin comparison
	// b.Run("Gin", func(b *testing.B) {
	//     // Implement Gin middleware benchmark here
	//     // router := gin.New()
	//     // router.Use(func(c *gin.Context) {
	//     //     c.Header("X-Middleware-1", "true")
	//     //     c.Next()
	//     // })
	//     // ... add more middleware ...
	//     // router.GET("/middleware", func(c *gin.Context) {
	//     //     c.String(200, "OK")
	//     // })
	//     // ... benchmark code ...
	// })
}
