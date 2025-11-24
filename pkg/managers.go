package pkg

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

// serverManager implements the ServerManager interface
type serverManager struct {
	servers      map[string]Server      // addr -> Server
	hosts        map[string]*HostConfig // hostname -> HostConfig
	tenants      map[string]*TenantInfo // tenantID -> TenantInfo
	portBindings map[string][]string    // port -> []hostname (for port reuse)
	mu           sync.RWMutex
	running      bool
}

// NewServerManager creates a new ServerManager instance
func NewServerManager() ServerManager {
	return &serverManager{
		servers:      make(map[string]Server),
		hosts:        make(map[string]*HostConfig),
		tenants:      make(map[string]*TenantInfo),
		portBindings: make(map[string][]string),
	}
}

// NewServer creates a new HTTP server instance
func (sm *serverManager) NewServer(config ServerConfig) Server {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	server := NewServer(config)
	return server
}

// StartAll starts all registered servers
func (sm *serverManager) StartAll() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.running {
		return errors.New("servers are already running")
	}

	// Start each server
	var startErrors []error
	for addr, server := range sm.servers {
		if err := server.Listen(addr); err != nil {
			startErrors = append(startErrors, fmt.Errorf("failed to start server on %s: %w", addr, err))
		}
	}

	if len(startErrors) > 0 {
		// If any server failed to start, stop all started servers
		for _, server := range sm.servers {
			if server.IsRunning() {
				server.Close()
			}
		}
		return fmt.Errorf("failed to start servers: %v", startErrors)
	}

	sm.running = true
	return nil
}

// StopAll stops all running servers
func (sm *serverManager) StopAll() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if !sm.running {
		return nil
	}

	var stopErrors []error
	for addr, server := range sm.servers {
		if server.IsRunning() {
			if err := server.Close(); err != nil {
				stopErrors = append(stopErrors, fmt.Errorf("failed to stop server on %s: %w", addr, err))
			}
		}
	}

	sm.running = false

	if len(stopErrors) > 0 {
		return fmt.Errorf("errors stopping servers: %v", stopErrors)
	}

	return nil
}

// GracefulShutdown performs graceful shutdown of all servers
func (sm *serverManager) GracefulShutdown(timeout time.Duration) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if !sm.running {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Create a wait group to shutdown all servers concurrently
	var wg sync.WaitGroup
	errChan := make(chan error, len(sm.servers))

	for addr, server := range sm.servers {
		if server.IsRunning() {
			wg.Add(1)
			go func(addr string, srv Server) {
				defer wg.Done()
				if err := srv.Shutdown(ctx); err != nil {
					errChan <- fmt.Errorf("failed to shutdown server on %s: %w", addr, err)
				}
			}(addr, server)
		}
	}

	// Wait for all shutdowns to complete
	wg.Wait()
	close(errChan)

	sm.running = false

	// Collect any errors
	var shutdownErrors []error
	for err := range errChan {
		shutdownErrors = append(shutdownErrors, err)
	}

	if len(shutdownErrors) > 0 {
		return fmt.Errorf("errors during graceful shutdown: %v", shutdownErrors)
	}

	return nil
}

// RegisterHost registers a host with its configuration
func (sm *serverManager) RegisterHost(hostname string, config HostConfig) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if hostname == "" {
		return errors.New("hostname cannot be empty")
	}

	// Check if host is already registered
	if _, exists := sm.hosts[hostname]; exists {
		return fmt.Errorf("host %s is already registered", hostname)
	}

	// Store host configuration
	sm.hosts[hostname] = &config

	return nil
}

// RegisterTenant registers a tenant with its associated hosts
func (sm *serverManager) RegisterTenant(tenantID string, hosts []string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if tenantID == "" {
		return errors.New("tenantID cannot be empty")
	}

	if len(hosts) == 0 {
		return errors.New("tenant must have at least one host")
	}

	// Check if tenant is already registered
	if _, exists := sm.tenants[tenantID]; exists {
		return fmt.Errorf("tenant %s is already registered", tenantID)
	}

	// Verify all hosts are registered
	for _, hostname := range hosts {
		if _, exists := sm.hosts[hostname]; !exists {
			return fmt.Errorf("host %s is not registered", hostname)
		}

		// Update host config with tenant ID
		sm.hosts[hostname].TenantID = tenantID
	}

	// Store tenant information
	sm.tenants[tenantID] = &TenantInfo{
		TenantID: tenantID,
		Hosts:    hosts,
		Config:   make(map[string]interface{}),
	}

	return nil
}

// UnregisterHost removes a host registration
func (sm *serverManager) UnregisterHost(hostname string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if hostname == "" {
		return errors.New("hostname cannot be empty")
	}

	// Check if host exists
	hostConfig, exists := sm.hosts[hostname]
	if !exists {
		return fmt.Errorf("host %s is not registered", hostname)
	}

	// If host belongs to a tenant, remove it from tenant's host list
	if hostConfig.TenantID != "" {
		if tenant, ok := sm.tenants[hostConfig.TenantID]; ok {
			// Remove host from tenant's host list
			newHosts := make([]string, 0)
			for _, h := range tenant.Hosts {
				if h != hostname {
					newHosts = append(newHosts, h)
				}
			}
			tenant.Hosts = newHosts

			// If tenant has no more hosts, remove the tenant
			if len(tenant.Hosts) == 0 {
				delete(sm.tenants, hostConfig.TenantID)
			}
		}
	}

	// Remove host
	delete(sm.hosts, hostname)

	return nil
}

// UnregisterTenant removes a tenant registration
func (sm *serverManager) UnregisterTenant(tenantID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if tenantID == "" {
		return errors.New("tenantID cannot be empty")
	}

	// Check if tenant exists
	tenant, exists := sm.tenants[tenantID]
	if !exists {
		return fmt.Errorf("tenant %s is not registered", tenantID)
	}

	// Remove tenant ID from all associated hosts
	for _, hostname := range tenant.Hosts {
		if hostConfig, ok := sm.hosts[hostname]; ok {
			hostConfig.TenantID = ""
		}
	}

	// Remove tenant
	delete(sm.tenants, tenantID)

	return nil
}

// GetServer retrieves a server by its address
func (sm *serverManager) GetServer(addr string) (Server, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	server, exists := sm.servers[addr]
	return server, exists
}

// GetServers returns all registered servers
func (sm *serverManager) GetServers() []Server {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	servers := make([]Server, 0, len(sm.servers))
	for _, server := range sm.servers {
		servers = append(servers, server)
	}

	return servers
}

// GetServerCount returns the number of registered servers
func (sm *serverManager) GetServerCount() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return len(sm.servers)
}

// HealthCheck performs health check on all servers
func (sm *serverManager) HealthCheck() error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var healthErrors []error

	for addr, server := range sm.servers {
		// Check if server is running
		if !server.IsRunning() {
			healthErrors = append(healthErrors, fmt.Errorf("server on %s is not running", addr))
			continue
		}

		// Try to connect to the server
		conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
		if err != nil {
			healthErrors = append(healthErrors, fmt.Errorf("failed to connect to server on %s: %w", addr, err))
			continue
		}
		conn.Close()
	}

	if len(healthErrors) > 0 {
		return fmt.Errorf("health check failed: %v", healthErrors)
	}

	return nil
}

// IsHealthy returns whether all servers are healthy
func (sm *serverManager) IsHealthy() bool {
	return sm.HealthCheck() == nil
}

// AddServer adds a server to the manager
func (sm *serverManager) AddServer(addr string, server Server) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if addr == "" {
		return errors.New("address cannot be empty")
	}

	if server == nil {
		return errors.New("server cannot be nil")
	}

	// Check if server already exists
	if _, exists := sm.servers[addr]; exists {
		return fmt.Errorf("server on %s already exists", addr)
	}

	sm.servers[addr] = server

	// Track port binding for port reuse
	host, port, err := net.SplitHostPort(addr)
	if err == nil {
		portKey := port
		if host != "" {
			portKey = host + ":" + port
		}
		sm.portBindings[portKey] = append(sm.portBindings[portKey], addr)
	}

	return nil
}

// RemoveServer removes a server from the manager
func (sm *serverManager) RemoveServer(addr string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if addr == "" {
		return errors.New("address cannot be empty")
	}

	// Check if server exists
	server, exists := sm.servers[addr]
	if !exists {
		return fmt.Errorf("server on %s does not exist", addr)
	}

	// Stop server if running
	if server.IsRunning() {
		if err := server.Close(); err != nil {
			return fmt.Errorf("failed to stop server on %s: %w", addr, err)
		}
	}

	// Remove from servers map
	delete(sm.servers, addr)

	// Remove from port bindings
	host, port, err := net.SplitHostPort(addr)
	if err == nil {
		portKey := port
		if host != "" {
			portKey = host + ":" + port
		}

		if bindings, ok := sm.portBindings[portKey]; ok {
			newBindings := make([]string, 0)
			for _, binding := range bindings {
				if binding != addr {
					newBindings = append(newBindings, binding)
				}
			}

			if len(newBindings) == 0 {
				delete(sm.portBindings, portKey)
			} else {
				sm.portBindings[portKey] = newBindings
			}
		}
	}

	return nil
}

// GetHostConfig retrieves host configuration by hostname
func (sm *serverManager) GetHostConfig(hostname string) (*HostConfig, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	config, exists := sm.hosts[hostname]
	return config, exists
}

// GetTenantInfo retrieves tenant information by tenant ID
func (sm *serverManager) GetTenantInfo(tenantID string) (*TenantInfo, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	info, exists := sm.tenants[tenantID]
	return info, exists
}

// GetPortBindings returns all addresses bound to a specific port
func (sm *serverManager) GetPortBindings(port string) []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	bindings, exists := sm.portBindings[port]
	if !exists {
		return []string{}
	}

	result := make([]string, len(bindings))
	copy(result, bindings)
	return result
}
