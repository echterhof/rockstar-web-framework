//go:build unix || linux || darwin || freebsd || netbsd || openbsd || dragonfly || aix

package pkg

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"syscall"

	"golang.org/x/sys/unix"
)

// createPlatformListener creates a Unix-specific listener
func createPlatformListener(config ListenerConfig) (net.Listener, error) {
	// Check if prefork is enabled
	if config.EnablePrefork {
		return createPreforkListener(config)
	}

	// Create standard listener with socket options
	return createUnixListener(config)
}

// createUnixListener creates a Unix listener with socket options
func createUnixListener(config ListenerConfig) (net.Listener, error) {
	// Parse address
	lc := net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) error {
			var err error
			c.Control(func(fd uintptr) {
				// Set SO_REUSEADDR
				if config.ReuseAddr {
					if e := unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEADDR, 1); e != nil {
						err = fmt.Errorf("failed to set SO_REUSEADDR: %w", e)
						return
					}
				}

				// Set SO_REUSEPORT (Linux, BSD)
				if config.ReusePort && supportsReusePort() {
					if e := unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1); e != nil {
						err = fmt.Errorf("failed to set SO_REUSEPORT: %w", e)
						return
					}
				}

				// Set read buffer size
				if config.ReadBuffer > 0 {
					if e := unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_RCVBUF, config.ReadBuffer); e != nil {
						err = fmt.Errorf("failed to set SO_RCVBUF: %w", e)
						return
					}
				}

				// Set write buffer size
				if config.WriteBuffer > 0 {
					if e := unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_SNDBUF, config.WriteBuffer); e != nil {
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

// createPreforkListener creates a prefork listener for Unix systems
func createPreforkListener(config ListenerConfig) (net.Listener, error) {
	// Check if we're the master process
	if !isChildProcess() {
		// Master process - fork workers
		return forkWorkers(config)
	}

	// Child process - create listener
	return createUnixListener(config)
}

// isChildProcess checks if this is a child process
func isChildProcess() bool {
	return os.Getenv("ROCKSTAR_PREFORK_CHILD") == "1"
}

// forkWorkers forks worker processes
func forkWorkers(config ListenerConfig) (net.Listener, error) {
	// Get the executable path
	executable, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %w", err)
	}

	// Fork worker processes
	for i := 0; i < config.PreforkWorkers; i++ {
		cmd := exec.Command(executable, os.Args[1:]...)
		cmd.Env = append(os.Environ(), "ROCKSTAR_PREFORK_CHILD=1")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Start(); err != nil {
			return nil, fmt.Errorf("failed to fork worker %d: %w", i, err)
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
