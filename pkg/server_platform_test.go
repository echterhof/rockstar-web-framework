package pkg

import (
	"io"
	"net/http"
	"runtime"
	"testing"
	"time"
)

// TestServerWithPlatformListener tests server with platform-specific listener
func TestServerWithPlatformListener(t *testing.T) {
	config := ServerConfig{
		EnableHTTP1: true,
		ListenerConfig: &ListenerConfig{
			Network:   "tcp",
			ReuseAddr: true,
		},
	}

	server := NewServer(config)
	router := NewRouter()

	router.GET("/platform", func(ctx Context) error {
		info := GetPlatformInfo()
		return ctx.JSON(http.StatusOK, info)
	})

	server.SetRouter(router)

	addr := "127.0.0.1:19091"
	err := server.Listen(addr)
	if err != nil {
		t.Fatalf("Failed to start server with platform listener: %v", err)
	}
	defer server.Close()

	time.Sleep(100 * time.Millisecond)

	// Make request
	resp, err := http.Get("http://" + addr + "/platform")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// TestServerWithBufferSizes tests server with custom buffer sizes
func TestServerWithBufferSizes(t *testing.T) {
	config := ServerConfig{
		EnableHTTP1:     true,
		ReadBufferSize:  65536,
		WriteBufferSize: 65536,
	}

	server := NewServer(config)
	router := NewRouter()

	router.GET("/buffer", func(ctx Context) error {
		return ctx.String(http.StatusOK, "Buffer test")
	})

	server.SetRouter(router)

	addr := "127.0.0.1:19092"
	err := server.Listen(addr)
	if err != nil {
		t.Fatalf("Failed to start server with buffer sizes: %v", err)
	}
	defer server.Close()

	time.Sleep(100 * time.Millisecond)

	// Make request
	resp, err := http.Get("http://" + addr + "/buffer")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "Buffer test" {
		t.Errorf("Expected 'Buffer test', got '%s'", string(body))
	}
}

// TestServerWithReusePort tests server with SO_REUSEPORT
func TestServerWithReusePort(t *testing.T) {
	// Skip on platforms that don't support SO_REUSEPORT
	if !supportsReusePort() {
		t.Skipf("Platform %s does not support SO_REUSEPORT", runtime.GOOS)
	}

	config := ServerConfig{
		EnableHTTP1: true,
		ListenerConfig: &ListenerConfig{
			Network:   "tcp",
			ReusePort: true,
			ReuseAddr: true,
		},
	}

	server1 := NewServer(config)
	router1 := NewRouter()
	router1.GET("/test", func(ctx Context) error {
		return ctx.String(http.StatusOK, "Server 1")
	})
	server1.SetRouter(router1)

	server2 := NewServer(config)
	router2 := NewRouter()
	router2.GET("/test", func(ctx Context) error {
		return ctx.String(http.StatusOK, "Server 2")
	})
	server2.SetRouter(router2)

	// Both servers should be able to listen on the same port with SO_REUSEPORT
	addr := "127.0.0.1:19093"

	err := server1.Listen(addr)
	if err != nil {
		t.Fatalf("Failed to start first server: %v", err)
	}
	defer server1.Close()

	time.Sleep(100 * time.Millisecond)

	err = server2.Listen(addr)
	if err != nil {
		t.Fatalf("Failed to start second server with SO_REUSEPORT: %v", err)
	}
	defer server2.Close()

	time.Sleep(100 * time.Millisecond)

	// Make requests - they should be load balanced between servers
	for i := 0; i < 5; i++ {
		resp, err := http.Get("http://" + addr + "/test")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	}
}

// TestServerPlatformInfo tests retrieving platform information
func TestServerPlatformInfo(t *testing.T) {
	info := GetPlatformInfo()

	if info.OS != runtime.GOOS {
		t.Errorf("Expected OS %s, got %s", runtime.GOOS, info.OS)
	}

	if info.Arch != runtime.GOARCH {
		t.Errorf("Expected Arch %s, got %s", runtime.GOARCH, info.Arch)
	}

	if info.NumCPU != runtime.NumCPU() {
		t.Errorf("Expected NumCPU %d, got %d", runtime.NumCPU(), info.NumCPU)
	}
}

// TestServerWithPreforkDisabled tests server without prefork
func TestServerWithPreforkDisabled(t *testing.T) {
	config := ServerConfig{
		EnableHTTP1:   true,
		EnablePrefork: false,
	}

	server := NewServer(config)
	router := NewRouter()

	router.GET("/noprefork", func(ctx Context) error {
		return ctx.String(http.StatusOK, "No prefork")
	})

	server.SetRouter(router)

	addr := "127.0.0.1:19094"
	err := server.Listen(addr)
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Close()

	time.Sleep(100 * time.Millisecond)

	resp, err := http.Get("http://" + addr + "/noprefork")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "No prefork" {
		t.Errorf("Expected 'No prefork', got '%s'", string(body))
	}
}

// TestServerMultiplePlatformListeners tests multiple servers with platform listeners
func TestServerMultiplePlatformListeners(t *testing.T) {
	config1 := ServerConfig{
		EnableHTTP1: true,
		ListenerConfig: &ListenerConfig{
			Network:   "tcp",
			ReuseAddr: true,
		},
	}

	config2 := ServerConfig{
		EnableHTTP1: true,
		ListenerConfig: &ListenerConfig{
			Network:   "tcp",
			ReuseAddr: true,
		},
	}

	server1 := NewServer(config1)
	router1 := NewRouter()
	router1.GET("/server1", func(ctx Context) error {
		return ctx.String(http.StatusOK, "Server 1")
	})
	server1.SetRouter(router1)

	server2 := NewServer(config2)
	router2 := NewRouter()
	router2.GET("/server2", func(ctx Context) error {
		return ctx.String(http.StatusOK, "Server 2")
	})
	server2.SetRouter(router2)

	// Start servers on different ports
	addr1 := "127.0.0.1:19095"
	addr2 := "127.0.0.1:19096"

	err := server1.Listen(addr1)
	if err != nil {
		t.Fatalf("Failed to start server 1: %v", err)
	}
	defer server1.Close()

	err = server2.Listen(addr2)
	if err != nil {
		t.Fatalf("Failed to start server 2: %v", err)
	}
	defer server2.Close()

	time.Sleep(100 * time.Millisecond)

	// Test server 1
	resp1, err := http.Get("http://" + addr1 + "/server1")
	if err != nil {
		t.Fatalf("Failed to make request to server 1: %v", err)
	}
	defer resp1.Body.Close()

	body1, _ := io.ReadAll(resp1.Body)
	if string(body1) != "Server 1" {
		t.Errorf("Expected 'Server 1', got '%s'", string(body1))
	}

	// Test server 2
	resp2, err := http.Get("http://" + addr2 + "/server2")
	if err != nil {
		t.Fatalf("Failed to make request to server 2: %v", err)
	}
	defer resp2.Body.Close()

	body2, _ := io.ReadAll(resp2.Body)
	if string(body2) != "Server 2" {
		t.Errorf("Expected 'Server 2', got '%s'", string(body2))
	}
}

// TestServerPlatformCompatibility tests platform compatibility
func TestServerPlatformCompatibility(t *testing.T) {
	// Test that server works on current platform
	config := ServerConfig{
		EnableHTTP1: true,
	}

	server := NewServer(config)
	router := NewRouter()

	router.GET("/compat", func(ctx Context) error {
		return ctx.String(http.StatusOK, runtime.GOOS)
	})

	server.SetRouter(router)

	addr := "127.0.0.1:19097"
	err := server.Listen(addr)
	if err != nil {
		t.Fatalf("Failed to start server on %s: %v", runtime.GOOS, err)
	}
	defer server.Close()

	time.Sleep(100 * time.Millisecond)

	resp, err := http.Get("http://" + addr + "/compat")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if string(body) != runtime.GOOS {
		t.Errorf("Expected '%s', got '%s'", runtime.GOOS, string(body))
	}
}
