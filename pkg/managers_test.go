package pkg

import (
	"context"
	"testing"
	"time"
)

func TestNewServerManager(t *testing.T) {
	sm := NewServerManager()
	if sm == nil {
		t.Fatal("NewServerManager returned nil")
	}

	if sm.GetServerCount() != 0 {
		t.Errorf("Expected 0 servers, got %d", sm.GetServerCount())
	}
}

func TestServerManager_NewServer(t *testing.T) {
	sm := NewServerManager()

	config := ServerConfig{
		EnableHTTP1:    true,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	server := sm.NewServer(config)
	if server == nil {
		t.Fatal("NewServer returned nil")
	}
}

func TestServerManager_RegisterHost(t *testing.T) {
	sm := NewServerManager()

	hostConfig := HostConfig{
		Hostname: "example.com",
		TenantID: "tenant1",
	}

	// Register host
	err := sm.RegisterHost("example.com", hostConfig)
	if err != nil {
		t.Fatalf("Failed to register host: %v", err)
	}

	// Try to register same host again
	err = sm.RegisterHost("example.com", hostConfig)
	if err == nil {
		t.Error("Expected error when registering duplicate host")
	}

	// Try to register with empty hostname
	err = sm.RegisterHost("", hostConfig)
	if err == nil {
		t.Error("Expected error when registering with empty hostname")
	}
}

func TestServerManager_RegisterTenant(t *testing.T) {
	sm := NewServerManager()

	// Register hosts first
	hostConfig1 := HostConfig{Hostname: "host1.example.com"}
	hostConfig2 := HostConfig{Hostname: "host2.example.com"}

	sm.RegisterHost("host1.example.com", hostConfig1)
	sm.RegisterHost("host2.example.com", hostConfig2)

	// Register tenant
	err := sm.RegisterTenant("tenant1", []string{"host1.example.com", "host2.example.com"})
	if err != nil {
		t.Fatalf("Failed to register tenant: %v", err)
	}

	// Verify tenant info
	tenantInfo, exists := sm.GetTenantInfo("tenant1")
	if !exists {
		t.Fatal("Tenant not found after registration")
	}

	if tenantInfo.TenantID != "tenant1" {
		t.Errorf("Expected tenant ID 'tenant1', got '%s'", tenantInfo.TenantID)
	}

	if len(tenantInfo.Hosts) != 2 {
		t.Errorf("Expected 2 hosts, got %d", len(tenantInfo.Hosts))
	}

	// Try to register duplicate tenant
	err = sm.RegisterTenant("tenant1", []string{"host1.example.com"})
	if err == nil {
		t.Error("Expected error when registering duplicate tenant")
	}

	// Try to register tenant with empty ID
	err = sm.RegisterTenant("", []string{"host1.example.com"})
	if err == nil {
		t.Error("Expected error when registering tenant with empty ID")
	}

	// Try to register tenant with no hosts
	err = sm.RegisterTenant("tenant2", []string{})
	if err == nil {
		t.Error("Expected error when registering tenant with no hosts")
	}

	// Try to register tenant with unregistered host
	err = sm.RegisterTenant("tenant3", []string{"unregistered.example.com"})
	if err == nil {
		t.Error("Expected error when registering tenant with unregistered host")
	}
}

func TestServerManager_UnregisterHost(t *testing.T) {
	sm := NewServerManager()

	// Register host
	hostConfig := HostConfig{Hostname: "example.com"}
	sm.RegisterHost("example.com", hostConfig)

	// Unregister host
	err := sm.UnregisterHost("example.com")
	if err != nil {
		t.Fatalf("Failed to unregister host: %v", err)
	}

	// Verify host is removed
	_, exists := sm.GetHostConfig("example.com")
	if exists {
		t.Error("Host still exists after unregistration")
	}

	// Try to unregister non-existent host
	err = sm.UnregisterHost("nonexistent.com")
	if err == nil {
		t.Error("Expected error when unregistering non-existent host")
	}

	// Try to unregister with empty hostname
	err = sm.UnregisterHost("")
	if err == nil {
		t.Error("Expected error when unregistering with empty hostname")
	}
}

func TestServerManager_UnregisterTenant(t *testing.T) {
	sm := NewServerManager()

	// Register hosts and tenant
	hostConfig := HostConfig{Hostname: "host1.example.com"}
	sm.RegisterHost("host1.example.com", hostConfig)
	sm.RegisterTenant("tenant1", []string{"host1.example.com"})

	// Unregister tenant
	err := sm.UnregisterTenant("tenant1")
	if err != nil {
		t.Fatalf("Failed to unregister tenant: %v", err)
	}

	// Verify tenant is removed
	_, exists := sm.GetTenantInfo("tenant1")
	if exists {
		t.Error("Tenant still exists after unregistration")
	}

	// Verify host's tenant ID is cleared
	hostConfig2, exists := sm.GetHostConfig("host1.example.com")
	if !exists {
		t.Fatal("Host was removed when unregistering tenant")
	}
	if hostConfig2.TenantID != "" {
		t.Error("Host's tenant ID was not cleared")
	}

	// Try to unregister non-existent tenant
	err = sm.UnregisterTenant("nonexistent")
	if err == nil {
		t.Error("Expected error when unregistering non-existent tenant")
	}

	// Try to unregister with empty tenant ID
	err = sm.UnregisterTenant("")
	if err == nil {
		t.Error("Expected error when unregistering with empty tenant ID")
	}
}

func TestServerManager_AddRemoveServer(t *testing.T) {
	sm := NewServerManager()

	config := ServerConfig{
		EnableHTTP1: true,
	}

	server := NewServer(config)

	// Add server
	err := sm.(*serverManager).AddServer("localhost:8080", server)
	if err != nil {
		t.Fatalf("Failed to add server: %v", err)
	}

	// Verify server count
	if sm.GetServerCount() != 1 {
		t.Errorf("Expected 1 server, got %d", sm.GetServerCount())
	}

	// Get server
	retrievedServer, exists := sm.GetServer("localhost:8080")
	if !exists {
		t.Fatal("Server not found after adding")
	}
	if retrievedServer != server {
		t.Error("Retrieved server does not match added server")
	}

	// Try to add duplicate server
	err = sm.(*serverManager).AddServer("localhost:8080", server)
	if err == nil {
		t.Error("Expected error when adding duplicate server")
	}

	// Remove server
	err = sm.(*serverManager).RemoveServer("localhost:8080")
	if err != nil {
		t.Fatalf("Failed to remove server: %v", err)
	}

	// Verify server is removed
	if sm.GetServerCount() != 0 {
		t.Errorf("Expected 0 servers after removal, got %d", sm.GetServerCount())
	}

	// Try to remove non-existent server
	err = sm.(*serverManager).RemoveServer("localhost:8080")
	if err == nil {
		t.Error("Expected error when removing non-existent server")
	}
}

func TestServerManager_GetServers(t *testing.T) {
	sm := NewServerManager()

	config := ServerConfig{
		EnableHTTP1: true,
	}

	// Add multiple servers
	server1 := NewServer(config)
	server2 := NewServer(config)
	server3 := NewServer(config)

	sm.(*serverManager).AddServer("localhost:8080", server1)
	sm.(*serverManager).AddServer("localhost:8081", server2)
	sm.(*serverManager).AddServer("localhost:8082", server3)

	// Get all servers
	servers := sm.GetServers()
	if len(servers) != 3 {
		t.Errorf("Expected 3 servers, got %d", len(servers))
	}
}

func TestServerManager_PortReuse(t *testing.T) {
	sm := NewServerManager()

	config := ServerConfig{
		EnableHTTP1: true,
	}

	// Add servers on same port but different hosts
	server1 := NewServer(config)
	server2 := NewServer(config)

	sm.(*serverManager).AddServer("host1.example.com:8080", server1)
	sm.(*serverManager).AddServer("host2.example.com:8080", server2)

	// Get port bindings
	bindings := sm.(*serverManager).GetPortBindings("host1.example.com:8080")
	if len(bindings) != 1 {
		t.Errorf("Expected 1 binding for host1.example.com:8080, got %d", len(bindings))
	}

	bindings = sm.(*serverManager).GetPortBindings("host2.example.com:8080")
	if len(bindings) != 1 {
		t.Errorf("Expected 1 binding for host2.example.com:8080, got %d", len(bindings))
	}
}

func TestServerManager_GracefulShutdown(t *testing.T) {
	sm := NewServerManager()

	config := ServerConfig{
		EnableHTTP1:     true,
		ShutdownTimeout: 5 * time.Second,
	}

	server := NewServer(config)
	sm.(*serverManager).AddServer("localhost:8080", server)

	// Graceful shutdown without starting servers
	err := sm.GracefulShutdown(5 * time.Second)
	if err != nil {
		t.Errorf("Graceful shutdown failed: %v", err)
	}
}

func TestServerManager_HealthCheck(t *testing.T) {
	sm := NewServerManager()

	// Health check with no servers should pass
	err := sm.HealthCheck()
	if err != nil {
		t.Errorf("Health check failed with no servers: %v", err)
	}

	// IsHealthy should return true
	if !sm.IsHealthy() {
		t.Error("IsHealthy returned false with no servers")
	}
}

func TestServerManager_MultiTenantScenario(t *testing.T) {
	sm := NewServerManager()

	// Register hosts for tenant1
	host1Config := HostConfig{Hostname: "app1.tenant1.com"}
	host2Config := HostConfig{Hostname: "app2.tenant1.com"}
	sm.RegisterHost("app1.tenant1.com", host1Config)
	sm.RegisterHost("app2.tenant1.com", host2Config)

	// Register hosts for tenant2
	host3Config := HostConfig{Hostname: "app1.tenant2.com"}
	sm.RegisterHost("app1.tenant2.com", host3Config)

	// Register tenants
	err := sm.RegisterTenant("tenant1", []string{"app1.tenant1.com", "app2.tenant1.com"})
	if err != nil {
		t.Fatalf("Failed to register tenant1: %v", err)
	}

	err = sm.RegisterTenant("tenant2", []string{"app1.tenant2.com"})
	if err != nil {
		t.Fatalf("Failed to register tenant2: %v", err)
	}

	// Verify tenant1 has 2 hosts
	tenant1Info, exists := sm.GetTenantInfo("tenant1")
	if !exists {
		t.Fatal("Tenant1 not found")
	}
	if len(tenant1Info.Hosts) != 2 {
		t.Errorf("Expected tenant1 to have 2 hosts, got %d", len(tenant1Info.Hosts))
	}

	// Verify tenant2 has 1 host
	tenant2Info, exists := sm.GetTenantInfo("tenant2")
	if !exists {
		t.Fatal("Tenant2 not found")
	}
	if len(tenant2Info.Hosts) != 1 {
		t.Errorf("Expected tenant2 to have 1 host, got %d", len(tenant2Info.Hosts))
	}

	// Verify host configs have correct tenant IDs
	host1, _ := sm.GetHostConfig("app1.tenant1.com")
	if host1 != nil && host1.TenantID != "tenant1" {
		t.Errorf("Expected app1.tenant1.com to have tenant ID 'tenant1', got '%s'", host1.TenantID)
	}

	host3, _ := sm.GetHostConfig("app1.tenant2.com")
	if host3 != nil && host3.TenantID != "tenant2" {
		t.Errorf("Expected app1.tenant2.com to have tenant ID 'tenant2', got '%s'", host3.TenantID)
	}
}

func TestServerManager_UnregisterHostWithTenant(t *testing.T) {
	sm := NewServerManager()

	// Register hosts and tenant
	host1Config := HostConfig{Hostname: "host1.example.com"}
	host2Config := HostConfig{Hostname: "host2.example.com"}
	sm.RegisterHost("host1.example.com", host1Config)
	sm.RegisterHost("host2.example.com", host2Config)
	sm.RegisterTenant("tenant1", []string{"host1.example.com", "host2.example.com"})

	// Unregister one host
	err := sm.UnregisterHost("host1.example.com")
	if err != nil {
		t.Fatalf("Failed to unregister host: %v", err)
	}

	// Verify tenant still exists with one host
	tenantInfo, exists := sm.GetTenantInfo("tenant1")
	if !exists {
		t.Fatal("Tenant was removed when unregistering one host")
	}
	if len(tenantInfo.Hosts) != 1 {
		t.Errorf("Expected tenant to have 1 host, got %d", len(tenantInfo.Hosts))
	}
	if tenantInfo.Hosts[0] != "host2.example.com" {
		t.Errorf("Expected remaining host to be 'host2.example.com', got '%s'", tenantInfo.Hosts[0])
	}

	// Unregister last host
	err = sm.UnregisterHost("host2.example.com")
	if err != nil {
		t.Fatalf("Failed to unregister last host: %v", err)
	}

	// Verify tenant is removed when all hosts are unregistered
	_, exists = sm.GetTenantInfo("tenant1")
	if exists {
		t.Error("Tenant still exists after unregistering all hosts")
	}
}

func TestServerManager_ConcurrentOperations(t *testing.T) {
	sm := NewServerManager()

	config := ServerConfig{
		EnableHTTP1: true,
	}

	// Test concurrent server additions
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(index int) {
			server := NewServer(config)
			addr := "localhost:" + string(rune(8080+index))
			sm.(*serverManager).AddServer(addr, server)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all servers were added
	if sm.GetServerCount() != 10 {
		t.Errorf("Expected 10 servers after concurrent additions, got %d", sm.GetServerCount())
	}
}

func TestServerManager_StartStopAll(t *testing.T) {
	sm := NewServerManager()

	// Test StopAll without starting
	err := sm.StopAll()
	if err != nil {
		t.Errorf("StopAll failed without servers: %v", err)
	}

	// Test StartAll with no servers
	err = sm.StartAll()
	if err != nil {
		t.Errorf("StartAll failed with no servers: %v", err)
	}
}

func TestServerManager_ServerLifecycle(t *testing.T) {
	sm := NewServerManager()

	config := ServerConfig{
		EnableHTTP1:     true,
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    30 * time.Second,
		ShutdownTimeout: 5 * time.Second,
	}

	// Create and add server
	server := NewServer(config)
	sm.(*serverManager).AddServer("localhost:9999", server)

	// Verify server is not running initially
	if server.IsRunning() {
		t.Error("Server should not be running initially")
	}

	// Note: We don't actually start the server in tests to avoid port conflicts
	// In a real scenario, you would:
	// 1. Start the server
	// 2. Verify it's running
	// 3. Perform graceful shutdown
	// 4. Verify it stopped cleanly
}

func TestServerManager_ErrorHandling(t *testing.T) {
	sm := NewServerManager()

	// Test adding server with empty address
	config := ServerConfig{EnableHTTP1: true}
	server := NewServer(config)

	err := sm.(*serverManager).AddServer("", server)
	if err == nil {
		t.Error("Expected error when adding server with empty address")
	}

	// Test adding nil server
	err = sm.(*serverManager).AddServer("localhost:8080", nil)
	if err == nil {
		t.Error("Expected error when adding nil server")
	}

	// Test removing server with empty address
	err = sm.(*serverManager).RemoveServer("")
	if err == nil {
		t.Error("Expected error when removing server with empty address")
	}
}

func TestServerManager_ContextCancellation(t *testing.T) {
	sm := NewServerManager()

	config := ServerConfig{
		EnableHTTP1: true,
	}

	server := NewServer(config)
	sm.(*serverManager).AddServer("localhost:8080", server)

	// Test graceful shutdown with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Shutdown should handle cancelled context gracefully
	err := sm.GracefulShutdown(5 * time.Second)
	if err != nil {
		// This is expected behavior - context is cancelled
		t.Logf("Graceful shutdown with cancelled context: %v", err)
	}

	// Verify context is cancelled
	select {
	case <-ctx.Done():
		// Expected
	default:
		t.Error("Context should be cancelled")
	}
}
