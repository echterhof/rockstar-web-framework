package pkg

import (
	"testing"
)

// ============================================================================
// Tests for MockPluginContext
// ============================================================================

func TestNewMockPluginContext(t *testing.T) {
	ctx := NewMockPluginContext()

	if ctx == nil {
		t.Fatal("NewMockPluginContext returned nil")
	}

	// Verify all services are initialized
	if ctx.Router() == nil {
		t.Error("Router should not be nil")
	}
	if ctx.Logger() == nil {
		t.Error("Logger should not be nil")
	}
	if ctx.Metrics() == nil {
		t.Error("Metrics should not be nil")
	}
	if ctx.Database() == nil {
		t.Error("Database should not be nil")
	}
	if ctx.Cache() == nil {
		t.Error("Cache should not be nil")
	}
	if ctx.Config() == nil {
		t.Error("Config should not be nil")
	}
	if ctx.FileSystem() == nil {
		t.Error("FileSystem should not be nil")
	}
	if ctx.Network() == nil {
		t.Error("Network should not be nil")
	}
	if ctx.PluginStorage() == nil {
		t.Error("PluginStorage should not be nil")
	}

	// Verify tracking structures are initialized
	if ctx.RegisteredHooks == nil {
		t.Error("RegisteredHooks should not be nil")
	}
	if ctx.PublishedEvents == nil {
		t.Error("PublishedEvents should not be nil")
	}
	if ctx.SubscribedEvents == nil {
		t.Error("SubscribedEvents should not be nil")
	}
	if ctx.ExportedServices == nil {
		t.Error("ExportedServices should not be nil")
	}
	if ctx.ImportedServices == nil {
		t.Error("ImportedServices should not be nil")
	}
	if ctx.RegisteredMiddleware == nil {
		t.Error("RegisteredMiddleware should not be nil")
	}
}

func TestNewMockPluginContextWithPermissions(t *testing.T) {
	permissions := PluginPermissions{
		AllowDatabase:   true,
		AllowCache:      false,
		AllowRouter:     true,
		AllowConfig:     false,
		AllowFileSystem: false,
		AllowNetwork:    false,
		AllowExec:       false,
	}

	ctx := NewMockPluginContextWithPermissions(permissions)

	if ctx == nil {
		t.Fatal("NewMockPluginContextWithPermissions returned nil")
	}

	if ctx.permissions.AllowDatabase != true {
		t.Error("AllowDatabase permission not set correctly")
	}
	if ctx.permissions.AllowCache != false {
		t.Error("AllowCache permission not set correctly")
	}
	if ctx.permissionChecker == nil {
		t.Error("Permission checker should be set")
	}
}

func TestNewMockPluginContextWithConfig(t *testing.T) {
	config := map[string]interface{}{
		"api_key": "test-key",
		"timeout": 30,
	}

	ctx := NewMockPluginContextWithConfig(config)

	if ctx == nil {
		t.Fatal("NewMockPluginContextWithConfig returned nil")
	}

	pluginConfig := ctx.PluginConfig()
	if pluginConfig["api_key"] != "test-key" {
		t.Error("api_key not set correctly in config")
	}
	if pluginConfig["timeout"] != 30 {
		t.Error("timeout not set correctly in config")
	}
}

func TestMockPluginContext_RegisterHook(t *testing.T) {
	ctx := NewMockPluginContext()

	handler := func(ctx HookContext) error {
		return nil
	}

	err := ctx.RegisterHook(HookTypeStartup, 100, handler)
	if err != nil {
		t.Fatalf("RegisterHook failed: %v", err)
	}

	if len(ctx.RegisteredHooks) != 1 {
		t.Errorf("Expected 1 registered hook, got %d", len(ctx.RegisteredHooks))
	}

	hook := ctx.RegisteredHooks[0]
	if hook.HookType != HookTypeStartup {
		t.Errorf("Expected hook type %s, got %s", HookTypeStartup, hook.HookType)
	}
	if hook.Priority != 100 {
		t.Errorf("Expected priority 100, got %d", hook.Priority)
	}
}

func TestMockPluginContext_PublishEvent(t *testing.T) {
	ctx := NewMockPluginContext()

	eventData := map[string]interface{}{
		"user_id": 123,
		"action":  "login",
	}

	err := ctx.PublishEvent("user.login", eventData)
	if err != nil {
		t.Fatalf("PublishEvent failed: %v", err)
	}

	if len(ctx.PublishedEvents) != 1 {
		t.Errorf("Expected 1 published event, got %d", len(ctx.PublishedEvents))
	}

	event := ctx.PublishedEvents[0]
	if event.Event != "user.login" {
		t.Errorf("Expected event 'user.login', got '%s'", event.Event)
	}
}

func TestMockPluginContext_SubscribeEvent(t *testing.T) {
	ctx := NewMockPluginContext()

	handler := func(event Event) error {
		return nil
	}

	err := ctx.SubscribeEvent("user.created", handler)
	if err != nil {
		t.Fatalf("SubscribeEvent failed: %v", err)
	}

	if len(ctx.SubscribedEvents) != 1 {
		t.Errorf("Expected 1 subscribed event, got %d", len(ctx.SubscribedEvents))
	}

	sub := ctx.SubscribedEvents[0]
	if sub.Event != "user.created" {
		t.Errorf("Expected event 'user.created', got '%s'", sub.Event)
	}
}

func TestMockPluginContext_ExportService(t *testing.T) {
	ctx := NewMockPluginContext()

	type MyService struct {
		Name string
	}

	service := &MyService{Name: "test-service"}

	err := ctx.ExportService("my-service", service)
	if err != nil {
		t.Fatalf("ExportService failed: %v", err)
	}

	if len(ctx.ExportedServices) != 1 {
		t.Errorf("Expected 1 exported service, got %d", len(ctx.ExportedServices))
	}

	exported, exists := ctx.ExportedServices["my-service"]
	if !exists {
		t.Error("Service 'my-service' not found in exported services")
	}

	exportedService, ok := exported.(*MyService)
	if !ok {
		t.Error("Exported service is not of type *MyService")
	}
	if exportedService.Name != "test-service" {
		t.Errorf("Expected service name 'test-service', got '%s'", exportedService.Name)
	}
}

func TestMockPluginContext_ImportService(t *testing.T) {
	ctx := NewMockPluginContext()

	// First export a service
	type MyService struct {
		Value int
	}
	service := &MyService{Value: 42}
	ctx.serviceRegistry.Export("other-plugin", "my-service", service)

	// Now import it
	imported, err := ctx.ImportService("other-plugin", "my-service")
	if err != nil {
		t.Fatalf("ImportService failed: %v", err)
	}

	if len(ctx.ImportedServices) != 1 {
		t.Errorf("Expected 1 imported service, got %d", len(ctx.ImportedServices))
	}

	importRecord := ctx.ImportedServices[0]
	if importRecord.PluginName != "other-plugin" {
		t.Errorf("Expected plugin name 'other-plugin', got '%s'", importRecord.PluginName)
	}
	if importRecord.ServiceName != "my-service" {
		t.Errorf("Expected service name 'my-service', got '%s'", importRecord.ServiceName)
	}

	importedService, ok := imported.(*MyService)
	if !ok {
		t.Error("Imported service is not of type *MyService")
	}
	if importedService.Value != 42 {
		t.Errorf("Expected service value 42, got %d", importedService.Value)
	}
}

func TestMockPluginContext_RegisterMiddleware(t *testing.T) {
	ctx := NewMockPluginContext()

	middleware := func(ctx Context, next HandlerFunc) error {
		return next(ctx)
	}

	routes := []string{"/api/*", "/admin/*"}

	err := ctx.RegisterMiddleware("auth-middleware", middleware, 100, routes)
	if err != nil {
		t.Fatalf("RegisterMiddleware failed: %v", err)
	}

	if len(ctx.RegisteredMiddleware) != 1 {
		t.Errorf("Expected 1 registered middleware, got %d", len(ctx.RegisteredMiddleware))
	}

	mw := ctx.RegisteredMiddleware[0]
	if mw.Name != "auth-middleware" {
		t.Errorf("Expected middleware name 'auth-middleware', got '%s'", mw.Name)
	}
	if mw.Priority != 100 {
		t.Errorf("Expected priority 100, got %d", mw.Priority)
	}
	if len(mw.Routes) != 2 {
		t.Errorf("Expected 2 routes, got %d", len(mw.Routes))
	}
}

func TestMockPluginContext_SetCustomServices(t *testing.T) {
	ctx := NewMockPluginContext()

	// Create custom mocks
	customRouter := &MockRouter{}
	customLogger := &MockLogger{}
	customDatabase := &MockDatabase{}

	// Set custom services
	ctx.SetRouter(customRouter)
	ctx.SetLogger(customLogger)
	ctx.SetDatabase(customDatabase)

	// Verify they were set
	if ctx.Router() != customRouter {
		t.Error("Custom router not set correctly")
	}
	if ctx.Logger() != customLogger {
		t.Error("Custom logger not set correctly")
	}
	if ctx.Database() != customDatabase {
		t.Error("Custom database not set correctly")
	}
}

// ============================================================================
// Tests for MockPluginStorage
// ============================================================================

func TestMockPluginStorage_SetAndGet(t *testing.T) {
	storage := NewMockPluginStorage()

	err := storage.Set("key1", "value1")
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	value, err := storage.Get("key1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if value != "value1" {
		t.Errorf("Expected 'value1', got '%v'", value)
	}
}

func TestMockPluginStorage_GetNonExistent(t *testing.T) {
	storage := NewMockPluginStorage()

	_, err := storage.Get("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent key, got nil")
	}
}

func TestMockPluginStorage_Delete(t *testing.T) {
	storage := NewMockPluginStorage()

	storage.Set("key1", "value1")

	err := storage.Delete("key1")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = storage.Get("key1")
	if err == nil {
		t.Error("Expected error after delete, got nil")
	}
}

func TestMockPluginStorage_List(t *testing.T) {
	storage := NewMockPluginStorage()

	storage.Set("key1", "value1")
	storage.Set("key2", "value2")
	storage.Set("key3", "value3")

	keys, err := storage.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(keys) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(keys))
	}
}

func TestMockPluginStorage_Clear(t *testing.T) {
	storage := NewMockPluginStorage()

	storage.Set("key1", "value1")
	storage.Set("key2", "value2")

	err := storage.Clear()
	if err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	keys, _ := storage.List()
	if len(keys) != 0 {
		t.Errorf("Expected 0 keys after clear, got %d", len(keys))
	}
}

// ============================================================================
// Tests for MockHookSystem
// ============================================================================

func TestMockHookSystem_RegisterAndList(t *testing.T) {
	hookSystem := NewMockHookSystem()

	handler := func(ctx HookContext) error {
		return nil
	}

	err := hookSystem.RegisterHook("test-plugin", HookTypeStartup, 100, handler)
	if err != nil {
		t.Fatalf("RegisterHook failed: %v", err)
	}

	hooks := hookSystem.ListHooks(HookTypeStartup)
	if len(hooks) != 1 {
		t.Errorf("Expected 1 hook, got %d", len(hooks))
	}
}

func TestMockHookSystem_UnregisterHook(t *testing.T) {
	hookSystem := NewMockHookSystem()

	handler := func(ctx HookContext) error {
		return nil
	}

	hookSystem.RegisterHook("test-plugin", HookTypeStartup, 100, handler)

	err := hookSystem.UnregisterHook("test-plugin", HookTypeStartup)
	if err != nil {
		t.Fatalf("UnregisterHook failed: %v", err)
	}

	hooks := hookSystem.ListHooks(HookTypeStartup)
	if len(hooks) != 0 {
		t.Errorf("Expected 0 hooks after unregister, got %d", len(hooks))
	}
}

func TestMockHookSystem_UnregisterAll(t *testing.T) {
	hookSystem := NewMockHookSystem()

	handler := func(ctx HookContext) error {
		return nil
	}

	hookSystem.RegisterHook("test-plugin", HookTypeStartup, 100, handler)
	hookSystem.RegisterHook("test-plugin", HookTypeShutdown, 100, handler)

	err := hookSystem.UnregisterAll("test-plugin")
	if err != nil {
		t.Fatalf("UnregisterAll failed: %v", err)
	}

	startupHooks := hookSystem.ListHooks(HookTypeStartup)
	shutdownHooks := hookSystem.ListHooks(HookTypeShutdown)

	if len(startupHooks) != 0 {
		t.Errorf("Expected 0 startup hooks, got %d", len(startupHooks))
	}
	if len(shutdownHooks) != 0 {
		t.Errorf("Expected 0 shutdown hooks, got %d", len(shutdownHooks))
	}
}

// ============================================================================
// Tests for MockEventBus
// ============================================================================

func TestMockEventBus_PublishAndSubscribe(t *testing.T) {
	eventBus := NewMockEventBus()

	received := false
	handler := func(event Event) error {
		received = true
		return nil
	}

	err := eventBus.Subscribe("test-plugin", "test.event", handler)
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	err = eventBus.Publish("test.event", "test-data")
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	if !received {
		t.Error("Event handler was not called")
	}
}

func TestMockEventBus_Unsubscribe(t *testing.T) {
	eventBus := NewMockEventBus()

	received := false
	handler := func(event Event) error {
		received = true
		return nil
	}

	eventBus.Subscribe("test-plugin", "test.event", handler)

	err := eventBus.Unsubscribe("test-plugin", "test.event")
	if err != nil {
		t.Fatalf("Unsubscribe failed: %v", err)
	}

	eventBus.Publish("test.event", "test-data")

	if received {
		t.Error("Event handler should not have been called after unsubscribe")
	}
}

// ============================================================================
// Tests for MockPermissionChecker
// ============================================================================

func TestMockPermissionChecker_AllowedPermission(t *testing.T) {
	permissions := PluginPermissions{
		AllowDatabase: true,
		AllowCache:    true,
	}

	checker := NewMockPermissionChecker(permissions)

	err := checker.CheckPermission("test-plugin", "database")
	if err != nil {
		t.Errorf("Expected no error for allowed permission, got: %v", err)
	}

	err = checker.CheckPermission("test-plugin", "cache")
	if err != nil {
		t.Errorf("Expected no error for allowed permission, got: %v", err)
	}
}

func TestMockPermissionChecker_DeniedPermission(t *testing.T) {
	permissions := PluginPermissions{
		AllowDatabase: false,
		AllowCache:    false,
	}

	checker := NewMockPermissionChecker(permissions)

	err := checker.CheckPermission("test-plugin", "database")
	if err == nil {
		t.Error("Expected error for denied permission, got nil")
	}

	err = checker.CheckPermission("test-plugin", "cache")
	if err == nil {
		t.Error("Expected error for denied permission, got nil")
	}
}

// ============================================================================
// Tests for Test Helper Functions
// ============================================================================

func TestAssertPluginInitialized(t *testing.T) {
	plugin := &TestPlugin{
		name:    "test-plugin",
		version: "1.0.0",
	}
	ctx := NewMockPluginContext()

	// Should not fail
	AssertPluginInitialized(t, plugin, ctx)
}

func TestAssertPluginLifecycle(t *testing.T) {
	plugin := &TestPlugin{
		name:    "test-plugin",
		version: "1.0.0",
	}
	ctx := NewMockPluginContext()

	// Should not fail
	AssertPluginLifecycle(t, plugin, ctx)
}

func TestAssertHookRegistered(t *testing.T) {
	ctx := NewMockPluginContext()

	handler := func(ctx HookContext) error {
		return nil
	}

	ctx.RegisterHook(HookTypeStartup, 100, handler)

	// Should not fail
	AssertHookRegistered(t, ctx, HookTypeStartup)
}

func TestAssertEventPublished(t *testing.T) {
	ctx := NewMockPluginContext()

	ctx.PublishEvent("test.event", "data")

	// Should not fail
	AssertEventPublished(t, ctx, "test.event")
}

func TestAssertServiceExported(t *testing.T) {
	ctx := NewMockPluginContext()

	ctx.ExportService("my-service", "service-data")

	// Should not fail
	AssertServiceExported(t, ctx, "my-service")
}

func TestAssertMiddlewareRegistered(t *testing.T) {
	ctx := NewMockPluginContext()

	middleware := func(ctx Context, next HandlerFunc) error {
		return next(ctx)
	}

	ctx.RegisterMiddleware("my-middleware", middleware, 100, []string{})

	// Should not fail
	AssertMiddlewareRegistered(t, ctx, "my-middleware")
}
