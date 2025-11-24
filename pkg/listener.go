package pkg

import (
	"fmt"
	"net"
	"runtime"
)

// ListenerConfig holds configuration for platform-specific listeners
type ListenerConfig struct {
	// Network type (tcp, tcp4, tcp6, unix)
	Network string

	// Address to listen on
	Address string

	// Prefork configuration
	EnablePrefork  bool
	PreforkWorkers int

	// Socket options
	ReusePort bool
	ReuseAddr bool

	// Buffer sizes
	ReadBuffer  int
	WriteBuffer int

	// Platform-specific options
	PlatformOptions map[string]interface{}
}

// CreateListener creates a platform-specific listener
func CreateListener(config ListenerConfig) (net.Listener, error) {
	// Validate configuration
	if config.Network == "" {
		config.Network = "tcp"
	}

	if config.Address == "" {
		return nil, fmt.Errorf("address is required")
	}

	// Set default prefork workers if enabled
	if config.EnablePrefork && config.PreforkWorkers == 0 {
		config.PreforkWorkers = runtime.NumCPU()
	}

	// Create platform-specific listener
	return createPlatformListener(config)
}

// GetPlatformInfo returns information about the current platform
func GetPlatformInfo() PlatformInfo {
	return PlatformInfo{
		OS:                runtime.GOOS,
		Arch:              runtime.GOARCH,
		NumCPU:            runtime.NumCPU(),
		SupportsPrefork:   supportsPrefork(),
		SupportsReusePort: supportsReusePort(),
	}
}

// PlatformInfo contains platform-specific information
type PlatformInfo struct {
	OS                string
	Arch              string
	NumCPU            int
	SupportsPrefork   bool
	SupportsReusePort bool
}

// supportsPrefork returns whether the current platform supports prefork
func supportsPrefork() bool {
	switch runtime.GOOS {
	case "linux", "windows":
		return true
	default:
		return false
	}
}

// supportsReusePort returns whether the current platform supports SO_REUSEPORT
func supportsReusePort() bool {
	switch runtime.GOOS {
	case "linux", "darwin", "freebsd", "netbsd", "openbsd", "dragonfly":
		return true
	default:
		return false
	}
}
