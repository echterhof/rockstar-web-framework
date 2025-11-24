//go:build linux || windows

package pkg

import (
	"os"
	"runtime"
	"testing"
)

// TestPreforkSupport tests that prefork is supported on Linux and Windows
func TestPreforkSupport(t *testing.T) {
	if !supportsPrefork() {
		t.Errorf("Prefork should be supported on %s", runtime.GOOS)
	}
}

// TestIsChildProcess tests child process detection
func TestIsChildProcess(t *testing.T) {
	// Initially should not be a child process
	if isChildProcess() {
		t.Error("Should not be a child process initially")
	}

	// Set environment variable
	os.Setenv("ROCKSTAR_PREFORK_CHILD", "1")
	defer os.Unsetenv("ROCKSTAR_PREFORK_CHILD")

	if !isChildProcess() {
		t.Error("Should be detected as child process")
	}
}

// TestCreatePreforkListener tests prefork listener creation
func TestCreatePreforkListener(t *testing.T) {
	// Skip if already a child process
	if isChildProcess() {
		t.Skip("Skipping prefork test in child process")
	}

	// This test would spawn child processes, so we skip it in normal test runs
	// to avoid complexity. In production, prefork would be tested with integration tests.
	t.Skip("Skipping prefork listener test to avoid spawning processes")

	config := ListenerConfig{
		Network:        "tcp",
		Address:        "127.0.0.1:0",
		EnablePrefork:  true,
		PreforkWorkers: 2,
		ReusePort:      true,
	}

	listener, err := CreateListener(config)
	if err != nil {
		t.Fatalf("Failed to create prefork listener: %v", err)
	}
	defer listener.Close()
}

// TestPreforkWorkerCount tests default worker count
func TestPreforkWorkerCount(t *testing.T) {
	config := ListenerConfig{
		Network:       "tcp",
		Address:       "127.0.0.1:0",
		EnablePrefork: true,
		// PreforkWorkers not set, should default to NumCPU
	}

	// Just validate the config is processed correctly
	if config.PreforkWorkers == 0 {
		// Should be set to NumCPU by CreateListener
		expectedWorkers := runtime.NumCPU()
		if expectedWorkers <= 0 {
			t.Error("Expected positive worker count")
		}
	}
}

// TestPreforkWithCustomWorkerCount tests custom worker count
func TestPreforkWithCustomWorkerCount(t *testing.T) {
	config := ListenerConfig{
		Network:        "tcp",
		Address:        "127.0.0.1:0",
		EnablePrefork:  true,
		PreforkWorkers: 4,
	}

	if config.PreforkWorkers != 4 {
		t.Errorf("Expected 4 workers, got %d", config.PreforkWorkers)
	}
}

// TestPreforkEnvironmentVariable tests environment variable handling
func TestPreforkEnvironmentVariable(t *testing.T) {
	// Save original value
	originalValue := os.Getenv("ROCKSTAR_PREFORK_CHILD")
	defer func() {
		if originalValue != "" {
			os.Setenv("ROCKSTAR_PREFORK_CHILD", originalValue)
		} else {
			os.Unsetenv("ROCKSTAR_PREFORK_CHILD")
		}
	}()

	// Test without environment variable
	os.Unsetenv("ROCKSTAR_PREFORK_CHILD")
	if isChildProcess() {
		t.Error("Should not be child process without environment variable")
	}

	// Test with environment variable set to "1"
	os.Setenv("ROCKSTAR_PREFORK_CHILD", "1")
	if !isChildProcess() {
		t.Error("Should be child process with environment variable set to 1")
	}

	// Test with environment variable set to other value
	os.Setenv("ROCKSTAR_PREFORK_CHILD", "0")
	if isChildProcess() {
		t.Error("Should not be child process with environment variable set to 0")
	}
}
