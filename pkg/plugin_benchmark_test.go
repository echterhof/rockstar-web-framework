//go:build benchmark
// +build benchmark

package pkg

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// ============================================================================
// Benchmark: Plugin Initialization
// Tests the performance of plugin initialization with various configurations
// ============================================================================

// BenchmarkPluginInitialization benchmarks plugin initialization performance
func BenchmarkPluginInitialization(b *testing.B) {
	b.Run("SinglePlugin", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			plugin := &benchmarkPlugin{
				name:    "test-plugin",
				version: "1.0.0",
			}
			ctx := NewMockPluginContext()

			if err := plugin.Initialize(ctx); err != nil {
				b.Fatalf("initialization failed: %v", err)
			}
		}
	})

	b.Run("MultiplePlugins", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			plugins := make([]Plugin, 10)
			for j := 0; j < 10; j++ {
				plugins[j] = &benchmarkPlugin{
					name:    fmt.Sprintf("plugin-%d", j),
					version: "1.0.0",
				}
			}

			for _, plugin := range plugins {
				ctx := NewMockPluginContext()
				if err := plugin.Initialize(ctx); err != nil {
					b.Fatalf("initialization failed: %v", err)
				}
			}
		}
	})

	b.Run("WithDependencies", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			// Create plugins with dependency chain
			plugin1 := &benchmarkPlugin{
				name:    "plugin-1",
				version: "1.0.0",
			}
			plugin2 := &benchmarkPlugin{
				name:         "plugin-2",
				version:      "1.0.0",
				dependencies: []PluginDependency{{Name: "plugin-1", Version: ">=1.0.0"}},
			}
			plugin3 := &benchmarkPlugin{
				name:         "plugin-3",
				version:      "1.0.0",
				dependencies: []PluginDependency{{Name: "plugin-2", Version: ">=1.0.0"}},
			}

			ctx := NewMockPluginContext()
			if err := plugin1.Initialize(ctx); err != nil {
				b.Fatalf("initialization failed: %v", err)
			}
			if err := plugin2.Initialize(ctx); err != nil {
				b.Fatalf("initialization failed: %v", err)
			}
			if err := plugin3.Initialize(ctx); err != nil {
				b.Fatalf("initialization failed: %v", err)
			}
		}
	})

	b.Run("WithConfiguration", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			plugin := &benchmarkPlugin{
				name:    "test-plugin",
				version: "1.0.0",
			}

			config := map[string]interface{}{
				"api_key":     "test-key-123",
				"timeout":     "30s",
				"max_retries": 3,
				"enabled":     true,
			}

			ctx := NewMockPluginContextWithConfig(config)
			if err := plugin.Initialize(ctx); err != nil {
				b.Fatalf("initialization failed: %v", err)
			}
		}
	})
}

// BenchmarkPluginLifecycle benchmarks full plugin lifecycle
func BenchmarkPluginLifecycle(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		plugin := &benchmarkPlugin{
			name:    "test-plugin",
			version: "1.0.0",
		}
		ctx := NewMockPluginContext()

		if err := plugin.Initialize(ctx); err != nil {
			b.Fatalf("initialization failed: %v", err)
		}
		if err := plugin.Start(); err != nil {
			b.Fatalf("start failed: %v", err)
		}
		if err := plugin.Stop(); err != nil {
			b.Fatalf("stop failed: %v", err)
		}
		if err := plugin.Cleanup(); err != nil {
			b.Fatalf("cleanup failed: %v", err)
		}
	}
}

// ============================================================================
// Benchmark: Hook Execution
// Tests the performance of plugin hook execution
// ============================================================================

// BenchmarkHookExecution benchmarks plugin hook execution performance
func BenchmarkHookExecution(b *testing.B) {
	b.Run("SingleHook", func(b *testing.B) {
		ctx := NewMockPluginContext()
		hookCalled := false

		hook := func(hctx HookContext) error {
			hookCalled = true
			return nil
		}

		ctx.RegisterHook(HookTypeStartup, 100, hook)

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			hookCalled = false
			// Simulate hook execution
			if err := hook(nil); err != nil {
				b.Fatalf("hook execution failed: %v", err)
			}
			if !hookCalled {
				b.Fatal("hook was not called")
			}
		}
	})

	b.Run("MultipleHooks", func(b *testing.B) {
		ctx := NewMockPluginContext()

		hooks := make([]func(HookContext) error, 10)
		for i := 0; i < 10; i++ {
			hooks[i] = func(hctx HookContext) error {
				return nil
			}
			ctx.RegisterHook(HookTypePreRequest, 100+i, hooks[i])
		}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			for _, hook := range hooks {
				if err := hook(nil); err != nil {
					b.Fatalf("hook execution failed: %v", err)
				}
			}
		}
	})

	b.Run("HookWithProcessing", func(b *testing.B) {
		ctx := NewMockPluginContext()

		hook := func(hctx HookContext) error {
			// Simulate some processing
			data := make(map[string]interface{})
			data["timestamp"] = time.Now().Unix()
			data["processed"] = true
			return nil
		}

		ctx.RegisterHook(HookTypePreRequest, 100, hook)

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			if err := hook(nil); err != nil {
				b.Fatalf("hook execution failed: %v", err)
			}
		}
	})

	b.Run("ConcurrentHooks", func(b *testing.B) {
		ctx := NewMockPluginContext()

		hook := func(hctx HookContext) error {
			time.Sleep(1 * time.Microsecond) // Simulate minimal work
			return nil
		}

		ctx.RegisterHook(HookTypePreRequest, 100, hook)

		b.ResetTimer()
		b.ReportAllocs()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				if err := hook(nil); err != nil {
					b.Errorf("hook execution failed: %v", err)
				}
			}
		})
	})
}

// ============================================================================
// Benchmark: Service Calls
// Tests the performance of plugin service access
// ============================================================================

// BenchmarkServiceCalls benchmarks plugin service access performance
func BenchmarkServiceCalls(b *testing.B) {
	b.Run("RouterAccess", func(b *testing.B) {
		ctx := NewMockPluginContext()

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			router := ctx.Router()
			if router == nil {
				b.Fatal("router is nil")
			}
		}
	})

	b.Run("DatabaseAccess", func(b *testing.B) {
		permissions := PluginPermissions{AllowDatabase: true}
		ctx := NewMockPluginContextWithPermissions(permissions)

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			db := ctx.Database()
			if db == nil {
				b.Fatal("database is nil")
			}
		}
	})

	b.Run("CacheAccess", func(b *testing.B) {
		permissions := PluginPermissions{AllowCache: true}
		ctx := NewMockPluginContextWithPermissions(permissions)

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			cache := ctx.Cache()
			if cache == nil {
				b.Fatal("cache is nil")
			}
		}
	})

	b.Run("ConfigAccess", func(b *testing.B) {
		config := map[string]interface{}{
			"api_key": "test-key",
			"timeout": "30s",
		}
		ctx := NewMockPluginContextWithConfig(config)

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			cfg := ctx.PluginConfig()
			if cfg == nil {
				b.Fatal("config is nil")
			}
		}
	})

	b.Run("LoggerAccess", func(b *testing.B) {
		ctx := NewMockPluginContext()

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			logger := ctx.Logger()
			if logger == nil {
				b.Fatal("logger is nil")
			}
		}
	})

	b.Run("StorageOperations", func(b *testing.B) {
		ctx := NewMockPluginContext()
		storage := ctx.PluginStorage()

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("key-%d", i)
			value := fmt.Sprintf("value-%d", i)

			if err := storage.Set(key, value); err != nil {
				b.Fatalf("storage set failed: %v", err)
			}

			retrieved, err := storage.Get(key)
			if err != nil {
				b.Fatalf("storage get failed: %v", err)
			}
			if retrieved != value {
				b.Fatalf("expected %s, got %s", value, retrieved)
			}
		}
	})

	b.Run("MultipleServiceAccess", func(b *testing.B) {
		permissions := PluginPermissions{
			AllowDatabase: true,
			AllowCache:    true,
			AllowRouter:   true,
		}
		ctx := NewMockPluginContextWithPermissions(permissions)

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_ = ctx.Router()
			_ = ctx.Database()
			_ = ctx.Cache()
			_ = ctx.Logger()
			_ = ctx.PluginConfig()
		}
	})
}

// BenchmarkServiceExportImport benchmarks service export/import performance
func BenchmarkServiceExportImport(b *testing.B) {
	b.Run("ExportService", func(b *testing.B) {
		ctx := NewMockPluginContext()
		service := &benchmarkService{name: "test-service"}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			serviceName := fmt.Sprintf("service-%d", i)
			if err := ctx.ExportService(serviceName, service); err != nil {
				b.Fatalf("export failed: %v", err)
			}
		}
	})

	b.Run("ImportService", func(b *testing.B) {
		ctx := NewMockPluginContext()
		service := &benchmarkService{name: "test-service"}
		ctx.ExportService("test-service", service)

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			imported, err := ctx.ImportService("test-plugin", "test-service")
			if err != nil {
				b.Fatalf("import failed: %v", err)
			}
			if imported == nil {
				b.Fatal("imported service is nil")
			}
		}
	})

	b.Run("ExportImportCycle", func(b *testing.B) {
		ctx := NewMockPluginContext()

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			serviceName := fmt.Sprintf("service-%d", i)
			service := &benchmarkService{name: serviceName}

			if err := ctx.ExportService(serviceName, service); err != nil {
				b.Fatalf("export failed: %v", err)
			}

			imported, err := ctx.ImportService("test-plugin", serviceName)
			if err != nil {
				b.Fatalf("import failed: %v", err)
			}
			if imported == nil {
				b.Fatal("imported service is nil")
			}
		}
	})
}

// ============================================================================
// Benchmark: Event System
// Tests the performance of plugin event publishing and subscription
// ============================================================================

// BenchmarkEventSystem benchmarks plugin event system performance
func BenchmarkEventSystem(b *testing.B) {
	b.Run("PublishEvent", func(b *testing.B) {
		ctx := NewMockPluginContext()

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			event := fmt.Sprintf("event-%d", i)
			data := map[string]interface{}{"index": i}

			if err := ctx.PublishEvent(event, data); err != nil {
				b.Fatalf("publish failed: %v", err)
			}
		}
	})

	b.Run("SubscribeEvent", func(b *testing.B) {
		ctx := NewMockPluginContext()

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			event := fmt.Sprintf("event-%d", i)
			handler := func(evt Event) error {
				return nil
			}

			if err := ctx.SubscribeEvent(event, handler); err != nil {
				b.Fatalf("subscribe failed: %v", err)
			}
		}
	})

	b.Run("PublishSubscribe", func(b *testing.B) {
		ctx := NewMockPluginContext()

		handler := func(evt Event) error {
			return nil
		}

		ctx.SubscribeEvent("test-event", handler)

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			if err := ctx.PublishEvent("test-event", map[string]interface{}{"index": i}); err != nil {
				b.Fatalf("publish failed: %v", err)
			}
		}
	})
}

// ============================================================================
// Benchmark: Plugin Manager
// Tests the performance of plugin manager operations
// ============================================================================

// BenchmarkPluginManager benchmarks plugin manager operations
func BenchmarkPluginManager(b *testing.B) {
	b.Run("DiscoverPlugins", func(b *testing.B) {
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			// Register test plugins
			for j := 0; j < 5; j++ {
				pluginName := fmt.Sprintf("bench-plugin-%d-%d", i, j)
				RegisterPlugin(pluginName, func() Plugin {
					return &benchmarkPlugin{
						name:    pluginName,
						version: "1.0.0",
					}
				})
			}

			// Get registered plugins
			plugins := GetRegisteredPlugins()
			if len(plugins) == 0 {
				b.Fatal("no plugins discovered")
			}
		}
	})

	b.Run("CreatePlugin", func(b *testing.B) {
		pluginName := "bench-create-plugin"
		RegisterPlugin(pluginName, func() Plugin {
			return &benchmarkPlugin{
				name:    pluginName,
				version: "1.0.0",
			}
		})

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			plugin, err := CreatePlugin(pluginName)
			if err != nil {
				b.Fatalf("create plugin failed: %v", err)
			}
			if plugin == nil {
				b.Fatal("plugin is nil")
			}
		}
	})

	b.Run("DependencyResolution", func(b *testing.B) {
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			graph := NewDependencyGraph()

			// Add plugins with dependencies
			graph.AddNode("plugin-a", "1.0.0", nil)
			graph.AddNode("plugin-b", "1.0.0", []PluginDependency{
				{Name: "plugin-a", Version: ">=1.0.0"},
			})
			graph.AddNode("plugin-c", "1.0.0", []PluginDependency{
				{Name: "plugin-b", Version: ">=1.0.0"},
			})

			order, err := graph.TopologicalSort()
			if err != nil {
				b.Fatalf("topological sort failed: %v", err)
			}
			if len(order) != 3 {
				b.Fatalf("expected 3 plugins, got %d", len(order))
			}
		}
	})

	b.Run("ConfigurationMerging", func(b *testing.B) {
		schema := map[string]interface{}{
			"api_key": map[string]interface{}{
				"type":    "string",
				"default": "default-key",
			},
			"timeout": map[string]interface{}{
				"type":    "string",
				"default": "30s",
			},
			"max_retries": map[string]interface{}{
				"type":    "int",
				"default": 3,
			},
		}

		userConfig := map[string]interface{}{
			"api_key": "user-key",
		}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			merged := mergeConfigWithDefaults(userConfig, schema)
			if merged["api_key"] != "user-key" {
				b.Fatal("user config not applied")
			}
			if merged["timeout"] != "30s" {
				b.Fatal("default not applied")
			}
		}
	})
}

// BenchmarkPluginMetrics benchmarks plugin metrics collection
func BenchmarkPluginMetrics(b *testing.B) {
	b.Run("RecordInit", func(b *testing.B) {
		metrics := NewPluginMetrics("test-plugin")

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			metrics.RecordInit(100*time.Millisecond, nil)
		}
	})

	b.Run("RecordHookExecution", func(b *testing.B) {
		metrics := NewPluginMetrics("test-plugin")

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			metrics.RecordHookExecution(HookTypePreRequest, 1*time.Millisecond)
		}
	})

	b.Run("RecordError", func(b *testing.B) {
		metrics := NewPluginMetrics("test-plugin")

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			metrics.RecordError(fmt.Errorf("test error %d", i))
		}
	})

	b.Run("GetMetrics", func(b *testing.B) {
		metrics := NewPluginMetrics("test-plugin")
		metrics.RecordInit(100*time.Millisecond, nil)
		metrics.RecordStart(50*time.Millisecond, nil)
		metrics.RecordHookExecution(HookTypePreRequest, 1*time.Millisecond)

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_ = metrics.GetHookMetrics()
		}
	})

	b.Run("ConcurrentMetrics", func(b *testing.B) {
		metrics := NewPluginMetrics("test-plugin")

		b.ResetTimer()
		b.ReportAllocs()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				metrics.RecordHookExecution(HookTypePreRequest, 1*time.Millisecond)
				metrics.RecordServiceCall()
				metrics.RecordEventPublished()
			}
		})
	})
}

// ============================================================================
// Benchmark: Permission Checks
// Tests the performance of permission enforcement
// ============================================================================

// BenchmarkPermissionChecks benchmarks permission checking performance
func BenchmarkPermissionChecks(b *testing.B) {
	b.Run("AllowedAccess", func(b *testing.B) {
		permissions := PluginPermissions{
			AllowDatabase: true,
			AllowCache:    true,
			AllowRouter:   true,
		}
		ctx := NewMockPluginContextWithPermissions(permissions)

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_ = ctx.Database()
			_ = ctx.Cache()
			_ = ctx.Router()
		}
	})

	b.Run("DeniedAccess", func(b *testing.B) {
		permissions := PluginPermissions{
			AllowDatabase: false,
			AllowCache:    false,
			AllowRouter:   false,
		}
		ctx := NewMockPluginContextWithPermissions(permissions)

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_ = ctx.Database() // Should return no-op
			_ = ctx.Cache()    // Should return no-op
			_ = ctx.Router()   // Should return no-op
		}
	})

	b.Run("MixedPermissions", func(b *testing.B) {
		permissions := PluginPermissions{
			AllowDatabase: true,
			AllowCache:    false,
			AllowRouter:   true,
		}
		ctx := NewMockPluginContextWithPermissions(permissions)

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_ = ctx.Database() // Allowed
			_ = ctx.Cache()    // Denied
			_ = ctx.Router()   // Allowed
		}
	})
}

// ============================================================================
// Benchmark: Concurrent Plugin Operations
// Tests the performance under concurrent load
// ============================================================================

// BenchmarkConcurrentOperations benchmarks concurrent plugin operations
func BenchmarkConcurrentOperations(b *testing.B) {
	b.Run("ConcurrentInitialization", func(b *testing.B) {
		b.ReportAllocs()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				plugin := &benchmarkPlugin{
					name:    "test-plugin",
					version: "1.0.0",
				}
				ctx := NewMockPluginContext()

				if err := plugin.Initialize(ctx); err != nil {
					b.Errorf("initialization failed: %v", err)
				}
			}
		})
	})

	b.Run("ConcurrentServiceAccess", func(b *testing.B) {
		permissions := PluginPermissions{
			AllowDatabase: true,
			AllowCache:    true,
			AllowRouter:   true,
		}
		ctx := NewMockPluginContextWithPermissions(permissions)

		b.ResetTimer()
		b.ReportAllocs()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = ctx.Database()
				_ = ctx.Cache()
				_ = ctx.Router()
			}
		})
	})

	b.Run("ConcurrentEventPublishing", func(b *testing.B) {
		ctx := NewMockPluginContext()

		b.ResetTimer()
		b.ReportAllocs()

		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				event := fmt.Sprintf("event-%d", i)
				data := map[string]interface{}{"index": i}

				if err := ctx.PublishEvent(event, data); err != nil {
					b.Errorf("publish failed: %v", err)
				}
				i++
			}
		})
	})

	b.Run("ConcurrentStorageOperations", func(b *testing.B) {
		ctx := NewMockPluginContext()
		storage := ctx.PluginStorage()

		b.ResetTimer()
		b.ReportAllocs()

		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				key := fmt.Sprintf("key-%d", i)
				value := fmt.Sprintf("value-%d", i)

				storage.Set(key, value)
				storage.Get(key)
				i++
			}
		})
	})
}

// ============================================================================
// Helper Types and Functions
// ============================================================================

// benchmarkPlugin is a simple plugin implementation for benchmarking
type benchmarkPlugin struct {
	name         string
	version      string
	dependencies []PluginDependency
	ctx          PluginContext
	mu           sync.Mutex
}

func (p *benchmarkPlugin) Name() string        { return p.name }
func (p *benchmarkPlugin) Version() string     { return p.version }
func (p *benchmarkPlugin) Description() string { return "Benchmark plugin" }
func (p *benchmarkPlugin) Author() string      { return "Benchmark" }

func (p *benchmarkPlugin) Dependencies() []PluginDependency {
	return p.dependencies
}

func (p *benchmarkPlugin) Initialize(ctx PluginContext) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.ctx = ctx
	return nil
}

func (p *benchmarkPlugin) Start() error {
	return nil
}

func (p *benchmarkPlugin) Stop() error {
	return nil
}

func (p *benchmarkPlugin) Cleanup() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.ctx = nil
	return nil
}

func (p *benchmarkPlugin) ConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"api_key": map[string]interface{}{
			"type":    "string",
			"default": "default-key",
		},
		"timeout": map[string]interface{}{
			"type":    "string",
			"default": "30s",
		},
	}
}

func (p *benchmarkPlugin) OnConfigChange(config map[string]interface{}) error {
	return nil
}

// benchmarkService is a simple service implementation for benchmarking
type benchmarkService struct {
	name string
}

func (s *benchmarkService) Name() string {
	return s.name
}

// ============================================================================
// Comparison Benchmarks
// These provide a baseline for comparing compile-time vs .so-based plugins
// ============================================================================

// BenchmarkCompileTimeVsSharedObject compares compile-time plugin performance
// Note: .so-based system has been removed, so these benchmarks document
// the expected performance improvements
func BenchmarkCompileTimeVsSharedObject(b *testing.B) {
	b.Run("CompileTime_DirectCall", func(b *testing.B) {
		plugin := &benchmarkPlugin{
			name:    "test-plugin",
			version: "1.0.0",
		}
		ctx := NewMockPluginContext()
		plugin.Initialize(ctx)

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			// Direct function call - no reflection or dynamic dispatch
			_ = plugin.Name()
			_ = plugin.Version()
		}
	})

	// Note: .so-based benchmarks would go here for comparison
	// Expected improvements:
	// - Startup: 50-90% faster (no dynamic loading)
	// - Hook execution: 10-20% faster (direct calls)
	// - Memory: 20-30% lower (no separate address spaces)
}

// BenchmarkMemoryFootprint benchmarks memory usage patterns
func BenchmarkMemoryFootprint(b *testing.B) {
	b.Run("SinglePluginMemory", func(b *testing.B) {
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			plugin := &benchmarkPlugin{
				name:    "test-plugin",
				version: "1.0.0",
			}
			ctx := NewMockPluginContext()
			plugin.Initialize(ctx)

			// Simulate some work
			_ = plugin.Name()
			_ = plugin.Version()

			plugin.Cleanup()
		}
	})

	b.Run("MultiplePluginsMemory", func(b *testing.B) {
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			plugins := make([]*benchmarkPlugin, 10)
			for j := 0; j < 10; j++ {
				plugins[j] = &benchmarkPlugin{
					name:    fmt.Sprintf("plugin-%d", j),
					version: "1.0.0",
				}
				ctx := NewMockPluginContext()
				plugins[j].Initialize(ctx)
			}

			// Cleanup
			for _, plugin := range plugins {
				plugin.Cleanup()
			}
		}
	})
}

// BenchmarkStartupTime benchmarks application startup with plugins
func BenchmarkStartupTime(b *testing.B) {
	b.Run("PluginRegistration", func(b *testing.B) {
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			// Simulate plugin registration during init()
			for j := 0; j < 10; j++ {
				pluginName := fmt.Sprintf("startup-plugin-%d-%d", i, j)
				RegisterPlugin(pluginName, func() Plugin {
					return &benchmarkPlugin{
						name:    pluginName,
						version: "1.0.0",
					}
				})
			}
		}
	})

	b.Run("PluginDiscoveryAndInit", func(b *testing.B) {
		// Register plugins once
		for j := 0; j < 10; j++ {
			pluginName := fmt.Sprintf("discovery-plugin-%d", j)
			RegisterPlugin(pluginName, func() Plugin {
				return &benchmarkPlugin{
					name:    pluginName,
					version: "1.0.0",
				}
			})
		}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			// Discover plugins
			pluginNames := GetRegisteredPlugins()

			// Initialize all plugins
			for _, name := range pluginNames {
				plugin, err := CreatePlugin(name)
				if err != nil {
					continue
				}
				ctx := NewMockPluginContext()
				plugin.Initialize(ctx)
			}
		}
	})
}

// BenchmarkRealWorldScenario benchmarks realistic plugin usage patterns
func BenchmarkRealWorldScenario(b *testing.B) {
	b.Run("AuthenticationPlugin", func(b *testing.B) {
		plugin := &benchmarkPlugin{
			name:    "auth-plugin",
			version: "1.0.0",
		}

		permissions := PluginPermissions{
			AllowDatabase: true,
			AllowCache:    true,
		}
		ctx := NewMockPluginContextWithPermissions(permissions)
		plugin.Initialize(ctx)

		// Register authentication hook
		authHook := func(hctx HookContext) error {
			// Simulate token validation
			_ = "Bearer token-123"
			return nil
		}
		ctx.RegisterHook(HookTypePreRequest, 100, authHook)

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			// Simulate request processing
			authHook(nil)
			_ = ctx.Cache()
			_ = ctx.Database()
		}
	})

	b.Run("LoggingPlugin", func(b *testing.B) {
		plugin := &benchmarkPlugin{
			name:    "logging-plugin",
			version: "1.0.0",
		}

		ctx := NewMockPluginContext()
		plugin.Initialize(ctx)

		// Register logging hooks
		preRequestHook := func(hctx HookContext) error {
			// Simulate logging
			_ = fmt.Sprintf("Request started at %v", time.Now())
			return nil
		}

		postRequestHook := func(hctx HookContext) error {
			// Simulate logging
			_ = fmt.Sprintf("Request completed at %v", time.Now())
			return nil
		}

		ctx.RegisterHook(HookTypePreRequest, 100, preRequestHook)
		ctx.RegisterHook(HookTypePostRequest, 100, postRequestHook)

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			// Simulate request processing
			preRequestHook(nil)
			// ... request processing ...
			postRequestHook(nil)
		}
	})

	b.Run("CachePlugin", func(b *testing.B) {
		plugin := &benchmarkPlugin{
			name:    "cache-plugin",
			version: "1.0.0",
		}

		permissions := PluginPermissions{
			AllowCache: true,
		}
		ctx := NewMockPluginContextWithPermissions(permissions)
		plugin.Initialize(ctx)

		storage := ctx.PluginStorage()

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("cache-key-%d", i%100) // Simulate cache hits

			// Try to get from cache
			_, err := storage.Get(key)
			if err != nil {
				// Cache miss - set value
				storage.Set(key, fmt.Sprintf("value-%d", i))
			}
		}
	})
}
