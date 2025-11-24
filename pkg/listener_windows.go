//go:build windows

package pkg

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"syscall"
)

// createPlatformListener creates a Windows-specific listener
func createPlatformListener(config ListenerConfig) (net.Listener, error) {
	// Check if prefork is enabled
	if config.EnablePrefork {
		return createPreforkListener(config)
	}

	// Create standard listener with socket options
	return createWindowsListener(config)
}

// createWindowsListener creates a Windows listener with socket options
func createWindowsListener(config ListenerConfig) (net.Listener, error) {
	// Parse address
	lc := net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) error {
			var err error
			c.Control(func(fd uintptr) {
				// Set SO_REUSEADDR on Windows
				if config.ReuseAddr {
					if e := syscall.SetsockoptInt(syscall.Handle(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); e != nil {
						err = fmt.Errorf("failed to set SO_REUSEADDR: %w", e)
						return
					}
				}

				// Windows doesn't support SO_REUSEPORT, but SO_REUSEADDR provides similar functionality

				// Set read buffer size
				if config.ReadBuffer > 0 {
					if e := syscall.SetsockoptInt(syscall.Handle(fd), syscall.SOL_SOCKET, syscall.SO_RCVBUF, config.ReadBuffer); e != nil {
						err = fmt.Errorf("failed to set SO_RCVBUF: %w", e)
						return
					}
				}

				// Set write buffer size
				if config.WriteBuffer > 0 {
					if e := syscall.SetsockoptInt(syscall.Handle(fd), syscall.SOL_SOCKET, syscall.SO_SNDBUF, config.WriteBuffer); e != nil {
						err = fmt.Errorf("failed to set SO_SNDBUF: %w", e)
						return
					}
				}
			})
			return err
		},
	}

	return lc.Listen(context.Background(), config.Network, config.Address)
}

// createPreforkListener creates a prefork listener for Windows
func createPreforkListener(config ListenerConfig) (net.Listener, error) {
	// Check if we're the master process
	if !isChildProcess() {
		// Master process - spawn workers
		return spawnWorkers(config)
	}

	// Child process - create listener
	return createWindowsListener(config)
}

// isChildProcess checks if this is a child process
func isChildProcess() bool {
	return os.Getenv("ROCKSTAR_PREFORK_CHILD") == "1"
}

// spawnWorkers spawns worker processes on Windows
func spawnWorkers(config ListenerConfig) (net.Listener, error) {
	// Get the executable path
	executable, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %w", err)
	}

	// Spawn worker processes
	for i := 0; i < config.PreforkWorkers; i++ {
		cmd := exec.Command(executable, os.Args[1:]...)
		cmd.Env = append(os.Environ(), "ROCKSTAR_PREFORK_CHILD=1")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		// Windows-specific: Create new process group
		cmd.SysProcAttr = &syscall.SysProcAttr{
			CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
		}

		if err := cmd.Start(); err != nil {
			return nil, fmt.Errorf("failed to spawn worker %d: %w", i, err)
		}

		// Don't wait for child processes
		go cmd.Wait()
	}

	// Master process doesn't listen
	// Return a dummy listener that blocks forever
	return &dummyListener{}, nil
}

// dummyListener is a listener that never accepts connections (for master process)
type dummyListener struct {
	closed chan struct{}
}

func (l *dummyListener) Accept() (net.Conn, error) {
	if l.closed == nil {
		l.closed = make(chan struct{})
	}
	<-l.closed
	return nil, net.ErrClosed
}

func (l *dummyListener) Close() error {
	if l.closed != nil {
		close(l.closed)
	}
	return nil
}

func (l *dummyListener) Addr() net.Addr {
	return &net.TCPAddr{}
}
