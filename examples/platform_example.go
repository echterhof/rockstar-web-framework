package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
	// Get platform information
	info := pkg.GetPlatformInfo()

	fmt.Println("=== Rockstar Web Framework - Platform Support Example ===")
	fmt.Printf("Platform: %s/%s\n", info.OS, info.Arch)
	fmt.Printf("CPU Cores: %d\n", info.NumCPU)
	fmt.Printf("Supports Prefork: %v\n", info.SupportsPrefork)
	fmt.Printf("Supports SO_REUSEPORT: %v\n", info.SupportsReusePort)
	fmt.Println()

	// Create server configuration with platform-specific optimizations
	config := pkg.ServerConfig{
		EnableHTTP1:     true,
		ReadTimeout:     10 * time.Second,
		WriteTimeout:    10 * time.Second,
		IdleTimeout:     60 * time.Second,
		MaxHeaderBytes:  1 << 20,
		ReadBufferSize:  65536,
		WriteBufferSize: 65536,
	}

	// Enable prefork on supported platforms
	if info.SupportsPrefork {
		fmt.Println("Enabling prefork mode...")
		config.EnablePrefork = true
		config.PreforkWorkers = info.NumCPU
		fmt.Printf("Worker processes: %d\n", config.PreforkWorkers)
	}

	// Configure listener with platform-specific options
	listenerConfig := &pkg.ListenerConfig{
		Network:     "tcp",
		ReuseAddr:   true,
		ReadBuffer:  65536,
		WriteBuffer: 65536,
	}

	// Enable SO_REUSEPORT on supported platforms
	if info.SupportsReusePort {
		fmt.Println("Enabling SO_REUSEPORT...")
		listenerConfig.ReusePort = true
	}

	config.ListenerConfig = listenerConfig

	// Create server
	server := pkg.NewServer(config)

	// Create router
	router := pkg.NewRouter()

	// Add routes
	router.GET("/", handleHome)
	router.GET("/platform", handlePlatform)
	router.GET("/health", handleHealth)
	router.GET("/worker", handleWorker)

	server.SetRouter(router)

	// Start server
	addr := ":8080"
	fmt.Printf("\nStarting server on %s...\n", addr)
	fmt.Println("Press Ctrl+C to stop")
	fmt.Println()

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		fmt.Println("\nShutting down gracefully...")
		if err := server.GracefulShutdown(10 * time.Second); err != nil {
			log.Printf("Error during shutdown: %v\n", err)
		}
		os.Exit(0)
	}()

	// Start listening
	if err := server.Listen(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	// Keep main goroutine alive
	select {}
}

// handleHome handles the home page
func handleHome(ctx pkg.Context) error {
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>Rockstar Framework - Platform Support</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .info { background: #f0f0f0; padding: 20px; border-radius: 5px; }
        .feature { margin: 10px 0; }
        .supported { color: green; }
        .not-supported { color: red; }
    </style>
</head>
<body>
    <h1>Rockstar Web Framework</h1>
    <h2>Cross-Platform Support Demo</h2>
    
    <div class="info">
        <h3>Platform Information</h3>
        <div class="feature">Operating System: <strong>` + runtime.GOOS + `</strong></div>
        <div class="feature">Architecture: <strong>` + runtime.GOARCH + `</strong></div>
        <div class="feature">CPU Cores: <strong>` + fmt.Sprintf("%d", runtime.NumCPU()) + `</strong></div>
    </div>
    
    <h3>Available Endpoints</h3>
    <ul>
        <li><a href="/platform">/platform</a> - Platform information (JSON)</li>
        <li><a href="/health">/health</a> - Health check</li>
        <li><a href="/worker">/worker</a> - Worker process info</li>
    </ul>
</body>
</html>
`
	return ctx.HTML(http.StatusOK, html, nil)
}

// handlePlatform returns platform information as JSON
func handlePlatform(ctx pkg.Context) error {
	info := pkg.GetPlatformInfo()

	response := map[string]interface{}{
		"os":                 info.OS,
		"arch":               info.Arch,
		"num_cpu":            info.NumCPU,
		"supports_prefork":   info.SupportsPrefork,
		"supports_reuseport": info.SupportsReusePort,
		"go_version":         runtime.Version(),
		"num_goroutines":     runtime.NumGoroutine(),
	}

	return ctx.JSON(http.StatusOK, response)
}

// handleHealth returns health status
func handleHealth(ctx pkg.Context) error {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"uptime":    time.Since(startTime).Seconds(),
	}

	return ctx.JSON(http.StatusOK, health)
}

// handleWorker returns worker process information
func handleWorker(ctx pkg.Context) error {
	workerInfo := map[string]interface{}{
		"pid":            os.Getpid(),
		"ppid":           os.Getppid(),
		"is_child":       os.Getenv("ROCKSTAR_PREFORK_CHILD") == "1",
		"num_goroutines": runtime.NumGoroutine(),
	}

	return ctx.JSON(http.StatusOK, workerInfo)
}

var startTime = time.Now()
