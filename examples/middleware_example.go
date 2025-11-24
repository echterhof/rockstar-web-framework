package main

import (
	"fmt"
	"log"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
	// Create a new middleware engine
	middlewareEngine := pkg.NewMiddlewareEngine()

	// Register logging middleware (before handler, high priority)
	err := middlewareEngine.Register(pkg.MiddlewareConfig{
		Name:     "logger",
		Position: pkg.MiddlewarePositionBefore,
		Priority: 100,
		Enabled:  true,
		Handler: func(ctx pkg.Context, next pkg.HandlerFunc) error {
			start := time.Now()
			fmt.Printf("[Logger] Request started: %s %s\n",
				ctx.Request().Method,
				ctx.Request().URL.Path)

			err := next(ctx)

			duration := time.Since(start)
			fmt.Printf("[Logger] Request completed in %v\n", duration)
			return err
		},
	})
	if err != nil {
		log.Fatalf("Failed to register logger middleware: %v", err)
	}

	// Register authentication middleware (before handler, medium priority)
	err = middlewareEngine.Register(pkg.MiddlewareConfig{
		Name:     "auth",
		Position: pkg.MiddlewarePositionBefore,
		Priority: 50,
		Enabled:  true,
		Handler: func(ctx pkg.Context, next pkg.HandlerFunc) error {
			fmt.Println("[Auth] Checking authentication...")

			// Simulate authentication check
			authHeader := ctx.GetHeader("Authorization")
			if authHeader == "" {
				fmt.Println("[Auth] No authorization header found")
				return ctx.JSON(401, map[string]string{
					"error": "Unauthorized",
				})
			}

			fmt.Println("[Auth] Authentication successful")
			return next(ctx)
		},
	})
	if err != nil {
		log.Fatalf("Failed to register auth middleware: %v", err)
	}

	// Register response time middleware (after handler, high priority)
	err = middlewareEngine.Register(pkg.MiddlewareConfig{
		Name:     "response-time",
		Position: pkg.MiddlewarePositionAfter,
		Priority: 100,
		Enabled:  true,
		Handler: func(ctx pkg.Context, next pkg.HandlerFunc) error {
			start := time.Now()
			err := next(ctx)
			duration := time.Since(start)

			// Add response time header
			ctx.SetHeader("X-Response-Time", duration.String())
			fmt.Printf("[Response-Time] Added header: %v\n", duration)
			return err
		},
	})
	if err != nil {
		log.Fatalf("Failed to register response-time middleware: %v", err)
	}

	// Register CORS middleware (after handler, medium priority)
	err = middlewareEngine.Register(pkg.MiddlewareConfig{
		Name:     "cors",
		Position: pkg.MiddlewarePositionAfter,
		Priority: 50,
		Enabled:  true,
		Handler: func(ctx pkg.Context, next pkg.HandlerFunc) error {
			err := next(ctx)

			// Add CORS headers
			ctx.SetHeader("Access-Control-Allow-Origin", "*")
			ctx.SetHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
			fmt.Println("[CORS] Added CORS headers")
			return err
		},
	})
	if err != nil {
		log.Fatalf("Failed to register cors middleware: %v", err)
	}

	// List all registered middleware
	fmt.Println("\n=== Registered Middleware ===")
	for _, mw := range middlewareEngine.List() {
		position := "before"
		if mw.Position == pkg.MiddlewarePositionAfter {
			position = "after"
		}
		fmt.Printf("- %s (position: %s, priority: %d, enabled: %v)\n",
			mw.Name, position, mw.Priority, mw.Enabled)
	}

	// Demonstrate dynamic middleware management
	fmt.Println("\n=== Dynamic Middleware Management ===")

	// Disable auth middleware
	fmt.Println("Disabling auth middleware...")
	middlewareEngine.Disable("auth")

	// Change priority of logger
	fmt.Println("Changing logger priority to 200...")
	middlewareEngine.SetPriority("logger", 200)

	// Change position of cors
	fmt.Println("Moving cors to before position...")
	middlewareEngine.SetPosition("cors", pkg.MiddlewarePositionBefore)

	// List middleware again
	fmt.Println("\n=== Updated Middleware ===")
	for _, mw := range middlewareEngine.List() {
		position := "before"
		if mw.Position == pkg.MiddlewarePositionAfter {
			position = "after"
		}
		fmt.Printf("- %s (position: %s, priority: %d, enabled: %v)\n",
			mw.Name, position, mw.Priority, mw.Enabled)
	}

	// Demonstrate middleware execution order
	fmt.Println("\n=== Middleware Execution Order ===")
	fmt.Println("Expected order:")
	fmt.Println("1. Logger (before, priority 200)")
	fmt.Println("2. CORS (before, priority 50)")
	fmt.Println("3. Response-Time (after, priority 100)")
	fmt.Println("4. Handler")
	fmt.Println("5. Response-Time completes")
	fmt.Println("6. CORS completes")
	fmt.Println("7. Logger completes")

	// Example of using middleware with router
	fmt.Println("\n=== Integration with Router ===")

	// Create router
	router := pkg.NewRouter()

	// Create a simple handler
	handler := func(ctx pkg.Context) error {
		fmt.Println("[Handler] Processing request...")
		return ctx.JSON(200, map[string]string{
			"message": "Hello from Rockstar Web Framework!",
		})
	}

	// Register route with per-route middleware
	router.GET("/api/users", handler,
		func(ctx pkg.Context, next pkg.HandlerFunc) error {
			fmt.Println("[Route Middleware] Validating user access...")
			return next(ctx)
		},
	)

	fmt.Println("Route registered with per-route middleware")

	// Demonstrate middleware helpers
	fmt.Println("\n=== Middleware Helpers ===")

	// ChainMiddleware example
	mw1 := func(ctx pkg.Context, next pkg.HandlerFunc) error {
		fmt.Println("[Chain MW1] Before")
		err := next(ctx)
		fmt.Println("[Chain MW1] After")
		return err
	}

	mw2 := func(ctx pkg.Context, next pkg.HandlerFunc) error {
		fmt.Println("[Chain MW2] Before")
		err := next(ctx)
		fmt.Println("[Chain MW2] After")
		return err
	}

	chained := pkg.ChainMiddleware(mw1, mw2)
	fmt.Println("Created chained middleware from mw1 and mw2")
	_ = chained

	// SkipMiddleware example
	skipMw := pkg.SkipMiddleware(
		func(ctx pkg.Context) bool {
			// Skip if path is /health
			return ctx.Request().URL.Path == "/health"
		},
		func(ctx pkg.Context, next pkg.HandlerFunc) error {
			fmt.Println("[Skip MW] This will be skipped for /health")
			return next(ctx)
		},
	)
	fmt.Println("Created conditional skip middleware")
	_ = skipMw

	// RecoverMiddleware example
	recoverMw := pkg.RecoverMiddleware(func(ctx pkg.Context, recovered interface{}) error {
		fmt.Printf("[Recover MW] Recovered from panic: %v\n", recovered)
		return ctx.JSON(500, map[string]string{
			"error": "Internal server error",
		})
	})
	fmt.Println("Created panic recovery middleware")
	_ = recoverMw

	fmt.Println("\n=== Middleware System Demo Complete ===")
	fmt.Println("\nKey Features Demonstrated:")
	fmt.Println("✓ Configurable middleware ordering (not static)")
	fmt.Println("✓ Pre-processing middleware (before handler)")
	fmt.Println("✓ Post-processing middleware (after handler)")
	fmt.Println("✓ Dynamic middleware management (enable/disable/priority)")
	fmt.Println("✓ Middleware helpers (chain, skip, recover)")
	fmt.Println("✓ Integration with router")
}
