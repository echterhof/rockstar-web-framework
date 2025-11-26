//go:build ignore

package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

var startTime = time.Now()

func main() {
	// Get platform information
	info := pkg.GetPlatformInfo()

	// Configuration with platform-specific optimizations
	config := pkg.FrameworkConfig{
		ServerConfig: pkg.ServerConfig{
			ReadTimeout:     10 * time.Second,
			WriteTimeout:    10 * time.Second,
			IdleTimeout:     60 * time.Second,
			EnableHTTP1:     true,
			EnableHTTP2:     true,
			MaxHeaderBytes:  1 << 20,
			ReadBufferSize:  65536,
			WriteBufferSize: 65536,
		},
	}

	// Enable platform-specific optimizations
	if info.SupportsPrefork {
		// Enable prefork on supported platforms (Linux, Windows)
		config.ServerConfig.EnablePrefork = true
		config.ServerConfig.PreforkWorkers = info.NumCPU
	}

	// Configure listener with platform-specific options
	listenerConfig := &pkg.ListenerConfig{
		Network:     "tcp",
		ReuseAddr:   true,
		ReadBuffer:  65536,
		WriteBuffer: 65536,
	}

	// Enable SO_REUSEPORT on supported platforms (Linux, BSD, macOS)
	if info.SupportsReusePort {
		listenerConfig.ReusePort = true
	}

	config.ServerConfig.ListenerConfig = listenerConfig

	// Create framework instance
	app, err := pkg.New(config)
	if err != nil {
		log.Fatalf("Failed to create framework: %v", err)
	}

	// Get router
	router := app.Router()

	// Route 1: Platform information
	router.GET("/platform", func(ctx pkg.Context) error {
		platformInfo := map[string]interface{}{
			"os":                 info.OS,
			"arch":               info.Arch,
			"num_cpu":            info.NumCPU,
			"supports_prefork":   info.SupportsPrefork,
			"supports_reuseport": info.SupportsReusePort,
			"go_version":         runtime.Version(),
			"compiler":           runtime.Compiler,
		}

		return ctx.JSON(200, platformInfo)
	})

	// Route 2: Runtime statistics
	router.GET("/runtime", func(ctx pkg.Context) error {
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)

		runtimeInfo := map[string]interface{}{
			"num_goroutines": runtime.NumGoroutine(),
			"num_cgo_calls":  runtime.NumCgoCall(),
			"memory": map[string]interface{}{
				"alloc_bytes":       memStats.Alloc,
				"total_alloc_bytes": memStats.TotalAlloc,
				"sys_bytes":         memStats.Sys,
				"num_gc":            memStats.NumGC,
			},
		}

		return ctx.JSON(200, runtimeInfo)
	})

	// Route 3: Process information
	router.GET("/process", func(ctx pkg.Context) error {
		processInfo := map[string]interface{}{
			"pid":         os.Getpid(),
			"ppid":        os.Getppid(),
			"is_child":    os.Getenv("ROCKSTAR_PREFORK_CHILD") == "1",
			"executable":  getExecutablePath(),
			"working_dir": getWorkingDir(),
		}

		return ctx.JSON(200, processInfo)
	})

	// Route 4: Environment variables
	router.GET("/environment", func(ctx pkg.Context) error {
		// Get relevant environment variables
		envVars := map[string]string{
			"GOOS":                   runtime.GOOS,
			"GOARCH":                 runtime.GOARCH,
			"GOMAXPROCS":             fmt.Sprintf("%d", runtime.GOMAXPROCS(0)),
			"ROCKSTAR_PREFORK_CHILD": os.Getenv("ROCKSTAR_PREFORK_CHILD"),
		}

		return ctx.JSON(200, map[string]interface{}{
			"environment": envVars,
		})
	})

	// Route 5: System capabilities
	router.GET("/capabilities", func(ctx pkg.Context) error {
		capabilities := map[string]interface{}{
			"prefork": map[string]interface{}{
				"supported": info.SupportsPrefork,
				"enabled":   config.ServerConfig.EnablePrefork,
				"workers":   config.ServerConfig.PreforkWorkers,
			},
			"reuseport": map[string]interface{}{
				"supported": info.SupportsReusePort,
				"enabled":   listenerConfig.ReusePort,
			},
			"http2": map[string]interface{}{
				"enabled": config.ServerConfig.EnableHTTP2,
			},
			"buffers": map[string]interface{}{
				"read_buffer":  listenerConfig.ReadBuffer,
				"write_buffer": listenerConfig.WriteBuffer,
			},
		}

		return ctx.JSON(200, capabilities)
	})

	// Route 6: Health check with platform info
	router.GET("/health", func(ctx pkg.Context) error {
		health := map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
			"uptime":    time.Since(startTime).Seconds(),
			"platform":  fmt.Sprintf("%s/%s", info.OS, info.Arch),
			"pid":       os.Getpid(),
		}

		return ctx.JSON(200, health)
	})

	// Route 7: OS-specific features
	router.GET("/os-features", func(ctx pkg.Context) error {
		features := getOSSpecificFeatures()
		return ctx.JSON(200, features)
	})

	// Route 8: Architecture-specific features
	router.GET("/arch-features", func(ctx pkg.Context) error {
		features := getArchSpecificFeatures()
		return ctx.JSON(200, features)
	})

	// Route 9: Performance test
	router.GET("/performance", func(ctx pkg.Context) error {
		// Measure response time
		start := time.Now()

		// Simulate some work
		sum := 0
		for i := 0; i < 1000000; i++ {
			sum += i
		}

		elapsed := time.Since(start)

		return ctx.JSON(200, map[string]interface{}{
			"elapsed_ms":     elapsed.Milliseconds(),
			"elapsed_ns":     elapsed.Nanoseconds(),
			"platform":       fmt.Sprintf("%s/%s", info.OS, info.Arch),
			"num_goroutines": runtime.NumGoroutine(),
		})
	})

	// Route 10: Home page with platform information
	router.GET("/", func(ctx pkg.Context) error {
		html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>Rockstar Framework - Platform Support</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background: #f5f5f5; }
        .container { max-width: 800px; margin: 0 auto; background: white; padding: 30px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { color: #333; border-bottom: 3px solid #4CAF50; padding-bottom: 10px; }
        .info-box { background: #f9f9f9; padding: 15px; margin: 15px 0; border-radius: 5px; border-left: 4px solid #4CAF50; }
        .feature { margin: 10px 0; padding: 8px; }
        .supported { color: #4CAF50; font-weight: bold; }
        .not-supported { color: #f44336; }
        .endpoints { background: #e3f2fd; padding: 15px; border-radius: 5px; margin-top: 20px; }
        .endpoints ul { list-style: none; padding: 0; }
        .endpoints li { margin: 8px 0; }
        .endpoints a { color: #1976D2; text-decoration: none; }
        .endpoints a:hover { text-decoration: underline; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🎸 Rockstar Web Framework</h1>
        <h2>Cross-Platform Support Demo</h2>
        
        <div class="info-box">
            <h3>Platform Information</h3>
            <div class="feature">Operating System: <strong>%s</strong></div>
            <div class="feature">Architecture: <strong>%s</strong></div>
            <div class="feature">CPU Cores: <strong>%d</strong></div>
            <div class="feature">Go Version: <strong>%s</strong></div>
            <div class="feature">Process ID: <strong>%d</strong></div>
        </div>
        
        <div class="info-box">
            <h3>Platform Features</h3>
            <div class="feature">Prefork Support: <span class="%s">%s</span></div>
            <div class="feature">SO_REUSEPORT Support: <span class="%s">%s</span></div>
            <div class="feature">HTTP/2 Support: <span class="supported">Enabled</span></div>
        </div>
        
        <div class="endpoints">
            <h3>Available Endpoints</h3>
            <ul>
                <li><a href="/platform">/platform</a> - Platform information (JSON)</li>
                <li><a href="/runtime">/runtime</a> - Runtime statistics</li>
                <li><a href="/process">/process</a> - Process information</li>
                <li><a href="/environment">/environment</a> - Environment variables</li>
                <li><a href="/capabilities">/capabilities</a> - System capabilities</li>
                <li><a href="/health">/health</a> - Health check</li>
                <li><a href="/os-features">/os-features</a> - OS-specific features</li>
                <li><a href="/arch-features">/arch-features</a> - Architecture features</li>
                <li><a href="/performance">/performance</a> - Performance test</li>
            </ul>
        </div>
    </div>
</body>
</html>
`,
			info.OS,
			info.Arch,
			info.NumCPU,
			runtime.Version(),
			os.Getpid(),
			getPreforkClass(info.SupportsPrefork),
			getPreforkText(info.SupportsPrefork),
			getReusePortClass(info.SupportsReusePort),
			getReusePortText(info.SupportsReusePort),
		)

		return ctx.HTML(200, html, nil)
	})

	// Startup message
	fmt.Printf("🎸 Rockstar Web Framework - Platform Example\n")
	fmt.Printf("==============================================\n\n")
	fmt.Printf("Platform Information:\n")
	fmt.Printf("  OS:                  %s\n", info.OS)
	fmt.Printf("  Architecture:        %s\n", info.Arch)
	fmt.Printf("  CPU Cores:           %d\n", info.NumCPU)
	fmt.Printf("  Go Version:          %s\n", runtime.Version())
	fmt.Printf("  Compiler:            %s\n", runtime.Compiler)
	fmt.Printf("\nPlatform Features:\n")
	fmt.Printf("  Prefork Support:     %v\n", info.SupportsPrefork)
	if info.SupportsPrefork {
		fmt.Printf("  Prefork Enabled:     %v\n", config.ServerConfig.EnablePrefork)
		fmt.Printf("  Worker Processes:    %d\n", config.ServerConfig.PreforkWorkers)
	}
	fmt.Printf("  SO_REUSEPORT:        %v\n", info.SupportsReusePort)
	if info.SupportsReusePort {
		fmt.Printf("  REUSEPORT Enabled:   %v\n", listenerConfig.ReusePort)
	}
	fmt.Printf("\nProcess Information:\n")
	fmt.Printf("  PID:                 %d\n", os.Getpid())
	fmt.Printf("  PPID:                %d\n", os.Getppid())
	fmt.Printf("  Is Child Process:    %v\n", os.Getenv("ROCKSTAR_PREFORK_CHILD") == "1")
	fmt.Printf("\nListening on :8080\n\n")
	fmt.Printf("Examples:\n")
	fmt.Printf("  curl http://localhost:8080/\n")
	fmt.Printf("  curl http://localhost:8080/platform\n")
	fmt.Printf("  curl http://localhost:8080/runtime\n")
	fmt.Printf("  curl http://localhost:8080/capabilities\n\n")

	// Start server
	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// Helper functions

func getExecutablePath() string {
	exe, err := os.Executable()
	if err != nil {
		return "unknown"
	}
	return exe
}

func getWorkingDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return dir
}

func getPreforkClass(supported bool) string {
	if supported {
		return "supported"
	}
	return "not-supported"
}

func getPreforkText(supported bool) string {
	if supported {
		return "Supported ✓"
	}
	return "Not Supported ✗"
}

func getReusePortClass(supported bool) string {
	if supported {
		return "supported"
	}
	return "not-supported"
}

func getReusePortText(supported bool) string {
	if supported {
		return "Supported ✓"
	}
	return "Not Supported ✗"
}

func getOSSpecificFeatures() map[string]interface{} {
	features := map[string]interface{}{
		"os": runtime.GOOS,
	}

	switch runtime.GOOS {
	case "linux":
		features["features"] = []string{
			"SO_REUSEPORT support",
			"Prefork support",
			"epoll for efficient I/O",
			"TCP_FASTOPEN support",
			"SO_KEEPALIVE support",
		}
	case "darwin":
		features["features"] = []string{
			"SO_REUSEPORT support",
			"kqueue for efficient I/O",
			"SO_KEEPALIVE support",
		}
	case "windows":
		features["features"] = []string{
			"Prefork support",
			"IOCP for efficient I/O",
			"SO_KEEPALIVE support",
		}
	case "freebsd", "netbsd", "openbsd", "dragonfly":
		features["features"] = []string{
			"SO_REUSEPORT support",
			"kqueue for efficient I/O",
			"SO_KEEPALIVE support",
		}
	default:
		features["features"] = []string{
			"Standard POSIX features",
		}
	}

	return features
}

func getArchSpecificFeatures() map[string]interface{} {
	features := map[string]interface{}{
		"arch": runtime.GOARCH,
	}

	switch runtime.GOARCH {
	case "amd64":
		features["features"] = []string{
			"64-bit architecture",
			"SSE/SSE2 support",
			"AVX support (if available)",
			"Large memory addressing",
		}
	case "arm64":
		features["features"] = []string{
			"64-bit ARM architecture",
			"NEON SIMD support",
			"Energy efficient",
		}
	case "386":
		features["features"] = []string{
			"32-bit architecture",
			"Legacy x86 support",
		}
	case "arm":
		features["features"] = []string{
			"32-bit ARM architecture",
			"Embedded systems support",
		}
	default:
		features["features"] = []string{
			fmt.Sprintf("%s architecture", runtime.GOARCH),
		}
	}

	return features
}
