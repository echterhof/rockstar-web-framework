package pkg

import (
	"net"
	"runtime"
	"testing"
	"time"
)

// TestGetPlatformInfo tests platform information retrieval
func TestGetPlatformInfo(t *testing.T) {
	info := GetPlatformInfo()

	if info.OS == "" {
		t.Error("Platform OS should not be empty")
	}

	if info.Arch == "" {
		t.Error("Platform Arch should not be empty")
	}

	if info.NumCPU <= 0 {
		t.Error("NumCPU should be positive")
	}

	// Verify platform-specific features
	switch runtime.GOOS {
	case "linux", "windows":
		if !info.SupportsPrefork {
			t.Error("Linux and Windows should support prefork")
		}
	default:
		if info.SupportsPrefork {
			t.Errorf("Platform %s should not support prefork", runtime.GOOS)
		}
	}

	// Verify SO_REUSEPORT support
	switch runtime.GOOS {
	case "linux", "darwin", "freebsd", "netbsd", "openbsd", "dragonfly":
		if !info.SupportsReusePort {
			t.Errorf("Platform %s should support SO_REUSEPORT", runtime.GOOS)
		}
	default:
		if info.SupportsReusePort {
			t.Errorf("Platform %s should not support SO_REUSEPORT", runtime.GOOS)
		}
	}
}

// TestCreateListener tests basic listener creation
func TestCreateListener(t *testing.T) {
	config := ListenerConfig{
		Network: "tcp",
		Address: "127.0.0.1:0", // Use port 0 for automatic assignment
	}

	listener, err := CreateListener(config)
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	if listener == nil {
		t.Fatal("Listener should not be nil")
	}

	addr := listener.Addr()
	if addr == nil {
		t.Error("Listener address should not be nil")
	}
}

// TestCreateListenerWithReuseAddr tests listener with SO_REUSEADDR
func TestCreateListenerWithReuseAddr(t *testing.T) {
	config := ListenerConfig{
		Network:   "tcp",
		Address:   "127.0.0.1:0",
		ReuseAddr: true,
	}

	listener, err := CreateListener(config)
	if err != nil {
		t.Fatalf("Failed to create listener with SO_REUSEADDR: %v", err)
	}
	defer listener.Close()

	if listener == nil {
		t.Fatal("Listener should not be nil")
	}
}

// TestCreateListenerWithReusePort tests listener with SO_REUSEPORT
func TestCreateListenerWithReusePort(t *testing.T) {
	// Skip on platforms that don't support SO_REUSEPORT
	if !supportsReusePort() {
		t.Skipf("Platform %s does not support SO_REUSEPORT", runtime.GOOS)
	}

	config := ListenerConfig{
		Network:   "tcp",
		Address:   "127.0.0.1:19090",
		ReusePort: true,
		ReuseAddr: true,
	}

	// Create first listener
	listener1, err := CreateListener(config)
	if err != nil {
		t.Fatalf("Failed to create first listener: %v", err)
	}
	defer listener1.Close()

	// Create second listener on same port (should work with SO_REUSEPORT)
	listener2, err := CreateListener(config)
	if err != nil {
		t.Fatalf("Failed to create second listener with SO_REUSEPORT: %v", err)
	}
	defer listener2.Close()

	// Both listeners should be able to accept connections
	if listener1 == nil || listener2 == nil {
		t.Fatal("Listeners should not be nil")
	}
}

// TestCreateListenerWithBufferSizes tests listener with custom buffer sizes
func TestCreateListenerWithBufferSizes(t *testing.T) {
	config := ListenerConfig{
		Network:     "tcp",
		Address:     "127.0.0.1:0",
		ReadBuffer:  65536,
		WriteBuffer: 65536,
	}

	listener, err := CreateListener(config)
	if err != nil {
		t.Fatalf("Failed to create listener with buffer sizes: %v", err)
	}
	defer listener.Close()

	if listener == nil {
		t.Fatal("Listener should not be nil")
	}
}

// TestCreateListenerInvalidAddress tests listener with invalid address
func TestCreateListenerInvalidAddress(t *testing.T) {
	config := ListenerConfig{
		Network: "tcp",
		Address: "", // Empty address should fail
	}

	_, err := CreateListener(config)
	if err == nil {
		t.Error("Expected error for empty address")
	}
}

// TestCreateListenerInvalidNetwork tests listener with invalid network
func TestCreateListenerInvalidNetwork(t *testing.T) {
	config := ListenerConfig{
		Network: "invalid",
		Address: "127.0.0.1:0",
	}

	_, err := CreateListener(config)
	if err == nil {
		t.Error("Expected error for invalid network")
	}
}

// TestListenerAcceptConnection tests accepting connections
func TestListenerAcceptConnection(t *testing.T) {
	config := ListenerConfig{
		Network: "tcp",
		Address: "127.0.0.1:0",
	}

	listener, err := CreateListener(config)
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	addr := listener.Addr().String()

	// Accept connection in goroutine
	connChan := make(chan net.Conn, 1)
	errChan := make(chan error, 1)

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			errChan <- err
			return
		}
		connChan <- conn
	}()

	// Connect to listener
	time.Sleep(50 * time.Millisecond)
	clientConn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("Failed to connect to listener: %v", err)
	}
	defer clientConn.Close()

	// Wait for accept
	select {
	case conn := <-connChan:
		if conn == nil {
			t.Error("Accepted connection should not be nil")
		}
		conn.Close()
	case err := <-errChan:
		t.Errorf("Failed to accept connection: %v", err)
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for connection accept")
	}
}

// TestListenerClose tests closing a listener
func TestListenerClose(t *testing.T) {
	config := ListenerConfig{
		Network: "tcp",
		Address: "127.0.0.1:0",
	}

	listener, err := CreateListener(config)
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}

	err = listener.Close()
	if err != nil {
		t.Errorf("Failed to close listener: %v", err)
	}

	// Try to accept after close (should fail)
	_, err = listener.Accept()
	if err == nil {
		t.Error("Accept should fail after close")
	}
}

// TestSupportsPrefork tests prefork support detection
func TestSupportsPrefork(t *testing.T) {
	supported := supportsPrefork()

	switch runtime.GOOS {
	case "linux", "windows":
		if !supported {
			t.Errorf("Platform %s should support prefork", runtime.GOOS)
		}
	default:
		if supported {
			t.Errorf("Platform %s should not support prefork", runtime.GOOS)
		}
	}
}

// TestSupportsReusePort tests SO_REUSEPORT support detection
func TestSupportsReusePort(t *testing.T) {
	supported := supportsReusePort()

	switch runtime.GOOS {
	case "linux", "darwin", "freebsd", "netbsd", "openbsd", "dragonfly":
		if !supported {
			t.Errorf("Platform %s should support SO_REUSEPORT", runtime.GOOS)
		}
	case "windows", "aix":
		if supported {
			t.Errorf("Platform %s should not support SO_REUSEPORT", runtime.GOOS)
		}
	}
}

// TestCreateListenerDefaultNetwork tests listener with default network
func TestCreateListenerDefaultNetwork(t *testing.T) {
	config := ListenerConfig{
		Address: "127.0.0.1:0",
		// Network not specified, should default to "tcp"
	}

	listener, err := CreateListener(config)
	if err != nil {
		t.Fatalf("Failed to create listener with default network: %v", err)
	}
	defer listener.Close()

	if listener == nil {
		t.Fatal("Listener should not be nil")
	}
}

// TestMultipleListenersOnDifferentPorts tests multiple listeners
func TestMultipleListenersOnDifferentPorts(t *testing.T) {
	config1 := ListenerConfig{
		Network: "tcp",
		Address: "127.0.0.1:0",
	}

	config2 := ListenerConfig{
		Network: "tcp",
		Address: "127.0.0.1:0",
	}

	listener1, err := CreateListener(config1)
	if err != nil {
		t.Fatalf("Failed to create first listener: %v", err)
	}
	defer listener1.Close()

	listener2, err := CreateListener(config2)
	if err != nil {
		t.Fatalf("Failed to create second listener: %v", err)
	}
	defer listener2.Close()

	// Both listeners should have different addresses
	addr1 := listener1.Addr().String()
	addr2 := listener2.Addr().String()

	if addr1 == addr2 {
		t.Error("Listeners should have different addresses")
	}
}
