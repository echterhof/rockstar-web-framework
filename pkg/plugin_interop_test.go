package pkg

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// **Feature: plugin-system, Property 41: Inter-plugin call mediation**
// **Validates: Requirements 10.3**
// For any plugin calling another plugin's exported function, the call should be routed through the Plugin Registry
func TestProperty_InterPluginCallMediation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("service export and import routes through service registry",
		prop.ForAll(
			func(exporterName, importerName, serviceName string) bool {
				// Skip if names are empty or the same
				if exporterName == "" || importerName == "" || serviceName == "" {
					return true
				}
				if exporterName == importerName {
					return true
				}

				// Create a shared service registry
				serviceRegistry := NewServiceRegistry()

				// Create context for exporter plugin
				exporterCtx := NewPluginContext(
					exporterName,
					&MockRouter{},
					&MockLogger{},
					&MockMetrics{},
					&MockDatabase{},
					&MockCache{},
					&MockConfig{},
					map[string]interface{}{},
					PluginPermissions{},
					nil, // hookSystem
					nil, // eventBus
					serviceRegistry,
					NewMiddlewareRegistry(),
					nil, // permissionChecker
				)

				// Create context for importer plugin
				importerCtx := NewPluginContext(
					importerName,
					&MockRouter{},
					&MockLogger{},
					&MockMetrics{},
					&MockDatabase{},
					&MockCache{},
					&MockConfig{},
					map[string]interface{}{},
					PluginPermissions{},
					nil, // hookSystem
					nil, // eventBus
					serviceRegistry,
					NewMiddlewareRegistry(),
					nil, // permissionChecker
				)

				// Exporter exports a service
				testService := &TestService{Value: "test-value"}
				err := exporterCtx.ExportService(serviceName, testService)
				if err != nil {
					return false
				}

				// Importer imports the service
				importedService, err := importerCtx.ImportService(exporterName, serviceName)
				if err != nil {
					return false
				}

				// Verify the imported service is the same as exported
				if importedService == nil {
					return false
				}

				// Type assert to verify it's the correct service
				typedService, ok := importedService.(*TestService)
				if !ok {
					return false
				}

				if typedService.Value != testService.Value {
					return false
				}

				return true
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	properties.Property("importing non-existent service returns error",
		prop.ForAll(
			func(importerName, exporterName, serviceName string) bool {
				// Skip if names are empty
				if importerName == "" || exporterName == "" || serviceName == "" {
					return true
				}

				// Create a shared service registry
				serviceRegistry := NewServiceRegistry()

				// Create context for importer plugin
				importerCtx := NewPluginContext(
					importerName,
					&MockRouter{},
					&MockLogger{},
					&MockMetrics{},
					&MockDatabase{},
					&MockCache{},
					&MockConfig{},
					map[string]interface{}{},
					PluginPermissions{},
					nil, // hookSystem
					nil, // eventBus
					serviceRegistry,
					NewMiddlewareRegistry(),
					nil, // permissionChecker
				)

				// Try to import a service that doesn't exist
				_, err := importerCtx.ImportService(exporterName, serviceName)

				// Should return an error
				if err == nil {
					return false
				}

				return true
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	properties.Property("multiple plugins can export different services",
		prop.ForAll(
			func(plugin1Name, plugin2Name, service1Name, service2Name string) bool {
				// Skip if names are empty or duplicate
				if plugin1Name == "" || plugin2Name == "" || service1Name == "" || service2Name == "" {
					return true
				}
				if plugin1Name == plugin2Name {
					return true
				}

				// Create a shared service registry
				serviceRegistry := NewServiceRegistry()

				// Create context for plugin 1
				plugin1Ctx := NewPluginContext(
					plugin1Name,
					&MockRouter{},
					&MockLogger{},
					&MockMetrics{},
					&MockDatabase{},
					&MockCache{},
					&MockConfig{},
					map[string]interface{}{},
					PluginPermissions{},
					nil, // hookSystem
					nil, // eventBus
					serviceRegistry,
					NewMiddlewareRegistry(),
					nil, // permissionChecker
				)

				// Create context for plugin 2
				plugin2Ctx := NewPluginContext(
					plugin2Name,
					&MockRouter{},
					&MockLogger{},
					&MockMetrics{},
					&MockDatabase{},
					&MockCache{},
					&MockConfig{},
					map[string]interface{}{},
					PluginPermissions{},
					nil, // hookSystem
					nil, // eventBus
					serviceRegistry,
					NewMiddlewareRegistry(),
					nil, // permissionChecker
				)

				// Plugin 1 exports a service
				service1 := &TestService{Value: "service1"}
				err := plugin1Ctx.ExportService(service1Name, service1)
				if err != nil {
					return false
				}

				// Plugin 2 exports a service
				service2 := &TestService{Value: "service2"}
				err = plugin2Ctx.ExportService(service2Name, service2)
				if err != nil {
					return false
				}

				// Plugin 2 can import plugin 1's service
				imported1, err := plugin2Ctx.ImportService(plugin1Name, service1Name)
				if err != nil {
					return false
				}

				// Plugin 1 can import plugin 2's service
				imported2, err := plugin1Ctx.ImportService(plugin2Name, service2Name)
				if err != nil {
					return false
				}

				// Verify both imports are correct
				typed1, ok := imported1.(*TestService)
				if !ok || typed1.Value != "service1" {
					return false
				}

				typed2, ok := imported2.(*TestService)
				if !ok || typed2.Value != "service2" {
					return false
				}

				return true
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	properties.Property("exporting duplicate service name returns error",
		prop.ForAll(
			func(pluginName, serviceName string) bool {
				// Skip if names are empty
				if pluginName == "" || serviceName == "" {
					return true
				}

				// Create a shared service registry
				serviceRegistry := NewServiceRegistry()

				// Create context for plugin
				pluginCtx := NewPluginContext(
					pluginName,
					&MockRouter{},
					&MockLogger{},
					&MockMetrics{},
					&MockDatabase{},
					&MockCache{},
					&MockConfig{},
					map[string]interface{}{},
					PluginPermissions{},
					nil, // hookSystem
					nil, // eventBus
					serviceRegistry,
					NewMiddlewareRegistry(),
					nil, // permissionChecker
				)

				// Export a service
				service1 := &TestService{Value: "service1"}
				err := pluginCtx.ExportService(serviceName, service1)
				if err != nil {
					return false
				}

				// Try to export another service with the same name
				service2 := &TestService{Value: "service2"}
				err = pluginCtx.ExportService(serviceName, service2)

				// Should return an error
				if err == nil {
					return false
				}

				return true
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	properties.Property("service registry isolates services by plugin name",
		prop.ForAll(
			func(plugin1Name, plugin2Name, serviceName string) bool {
				// Skip if names are empty or the same
				if plugin1Name == "" || plugin2Name == "" || serviceName == "" {
					return true
				}
				if plugin1Name == plugin2Name {
					return true
				}

				// Create a shared service registry
				serviceRegistry := NewServiceRegistry()

				// Create context for plugin 1
				plugin1Ctx := NewPluginContext(
					plugin1Name,
					&MockRouter{},
					&MockLogger{},
					&MockMetrics{},
					&MockDatabase{},
					&MockCache{},
					&MockConfig{},
					map[string]interface{}{},
					PluginPermissions{},
					nil, // hookSystem
					nil, // eventBus
					serviceRegistry,
					NewMiddlewareRegistry(),
					nil, // permissionChecker
				)

				// Create context for plugin 2
				plugin2Ctx := NewPluginContext(
					plugin2Name,
					&MockRouter{},
					&MockLogger{},
					&MockMetrics{},
					&MockDatabase{},
					&MockCache{},
					&MockConfig{},
					map[string]interface{}{},
					PluginPermissions{},
					nil, // hookSystem
					nil, // eventBus
					serviceRegistry,
					NewMiddlewareRegistry(),
					nil, // permissionChecker
				)

				// Both plugins export services with the same name
				service1 := &TestService{Value: "plugin1-service"}
				err := plugin1Ctx.ExportService(serviceName, service1)
				if err != nil {
					return false
				}

				service2 := &TestService{Value: "plugin2-service"}
				err = plugin2Ctx.ExportService(serviceName, service2)
				if err != nil {
					return false
				}

				// Import from plugin 1 should get plugin 1's service
				imported1, err := plugin2Ctx.ImportService(plugin1Name, serviceName)
				if err != nil {
					return false
				}

				typed1, ok := imported1.(*TestService)
				if !ok || typed1.Value != "plugin1-service" {
					return false
				}

				// Import from plugin 2 should get plugin 2's service
				imported2, err := plugin1Ctx.ImportService(plugin2Name, serviceName)
				if err != nil {
					return false
				}

				typed2, ok := imported2.(*TestService)
				if !ok || typed2.Value != "plugin2-service" {
					return false
				}

				return true
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	properties.TestingRun(t)
}

// **Feature: plugin-system, Property 42: Service discoverability**
// **Validates: Requirements 10.4**
// For any plugin exporting a service, other plugins should be able to discover and import that service by name
func TestProperty_ServiceDiscoverability(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("exported services are discoverable by other plugins",
		prop.ForAll(
			func(exporterName, importerName, serviceName string) bool {
				// Skip if names are empty or the same
				if exporterName == "" || importerName == "" || serviceName == "" {
					return true
				}
				if exporterName == importerName {
					return true
				}

				// Create a shared service registry
				serviceRegistry := NewServiceRegistry()

				// Create context for exporter plugin
				exporterCtx := NewPluginContext(
					exporterName,
					&MockRouter{},
					&MockLogger{},
					&MockMetrics{},
					&MockDatabase{},
					&MockCache{},
					&MockConfig{},
					map[string]interface{}{},
					PluginPermissions{},
					nil, // hookSystem
					nil, // eventBus
					serviceRegistry,
					NewMiddlewareRegistry(),
					nil, // permissionChecker
				)

				// Create context for importer plugin
				importerCtx := NewPluginContext(
					importerName,
					&MockRouter{},
					&MockLogger{},
					&MockMetrics{},
					&MockDatabase{},
					&MockCache{},
					&MockConfig{},
					map[string]interface{}{},
					PluginPermissions{},
					nil, // hookSystem
					nil, // eventBus
					serviceRegistry,
					NewMiddlewareRegistry(),
					nil, // permissionChecker
				)

				// Exporter exports a service
				testService := &TestService{Value: "discoverable-service"}
				err := exporterCtx.ExportService(serviceName, testService)
				if err != nil {
					return false
				}

				// List services from the exporter
				services := serviceRegistry.List(exporterName)

				// The service should be in the list
				found := false
				for _, name := range services {
					if name == serviceName {
						found = true
						break
					}
				}

				if !found {
					return false
				}

				// Importer should be able to import the discovered service
				importedService, err := importerCtx.ImportService(exporterName, serviceName)
				if err != nil {
					return false
				}

				if importedService == nil {
					return false
				}

				return true
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	properties.Property("multiple exported services are all discoverable",
		prop.ForAll(
			func(exporterName string, serviceNames []string) bool {
				// Skip if names are empty
				if exporterName == "" || len(serviceNames) == 0 {
					return true
				}

				// Filter out empty service names and duplicates
				uniqueServices := make(map[string]bool)
				for _, name := range serviceNames {
					if name != "" {
						uniqueServices[name] = true
					}
				}

				if len(uniqueServices) == 0 {
					return true
				}

				// Create a shared service registry
				serviceRegistry := NewServiceRegistry()

				// Create context for exporter plugin
				exporterCtx := NewPluginContext(
					exporterName,
					&MockRouter{},
					&MockLogger{},
					&MockMetrics{},
					&MockDatabase{},
					&MockCache{},
					&MockConfig{},
					map[string]interface{}{},
					PluginPermissions{},
					nil, // hookSystem
					nil, // eventBus
					serviceRegistry,
					NewMiddlewareRegistry(),
					nil, // permissionChecker
				)

				// Export all services
				for serviceName := range uniqueServices {
					testService := &TestService{Value: serviceName}
					err := exporterCtx.ExportService(serviceName, testService)
					if err != nil {
						return false
					}
				}

				// List all services
				listedServices := serviceRegistry.List(exporterName)

				// Verify all exported services are in the list
				if len(listedServices) != len(uniqueServices) {
					return false
				}

				for _, serviceName := range listedServices {
					if !uniqueServices[serviceName] {
						return false
					}
				}

				return true
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.SliceOf(gen.AlphaString()),
		),
	)

	properties.Property("service list is empty for plugin with no exports",
		prop.ForAll(
			func(pluginName string) bool {
				// Skip if name is empty
				if pluginName == "" {
					return true
				}

				// Create a shared service registry
				serviceRegistry := NewServiceRegistry()

				// List services for a plugin that hasn't exported anything
				services := serviceRegistry.List(pluginName)

				// Should return an empty list
				if len(services) != 0 {
					return false
				}

				return true
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	properties.Property("unregistering a service removes it from discovery",
		prop.ForAll(
			func(pluginName, serviceName string) bool {
				// Skip if names are empty
				if pluginName == "" || serviceName == "" {
					return true
				}

				// Create a shared service registry
				serviceRegistry := NewServiceRegistry()

				// Create context for plugin
				pluginCtx := NewPluginContext(
					pluginName,
					&MockRouter{},
					&MockLogger{},
					&MockMetrics{},
					&MockDatabase{},
					&MockCache{},
					&MockConfig{},
					map[string]interface{}{},
					PluginPermissions{},
					nil, // hookSystem
					nil, // eventBus
					serviceRegistry,
					NewMiddlewareRegistry(),
					nil, // permissionChecker
				)

				// Export a service
				testService := &TestService{Value: "test"}
				err := pluginCtx.ExportService(serviceName, testService)
				if err != nil {
					return false
				}

				// Verify it's discoverable
				services := serviceRegistry.List(pluginName)
				if len(services) != 1 {
					return false
				}

				// Unregister the service
				err = serviceRegistry.Unregister(pluginName, serviceName)
				if err != nil {
					return false
				}

				// Verify it's no longer discoverable
				services = serviceRegistry.List(pluginName)
				if len(services) != 0 {
					return false
				}

				return true
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	properties.Property("service discovery is isolated by plugin",
		prop.ForAll(
			func(plugin1Name, plugin2Name, service1Name, service2Name string) bool {
				// Skip if names are empty or duplicate
				if plugin1Name == "" || plugin2Name == "" || service1Name == "" || service2Name == "" {
					return true
				}
				if plugin1Name == plugin2Name {
					return true
				}

				// Create a shared service registry
				serviceRegistry := NewServiceRegistry()

				// Create context for plugin 1
				plugin1Ctx := NewPluginContext(
					plugin1Name,
					&MockRouter{},
					&MockLogger{},
					&MockMetrics{},
					&MockDatabase{},
					&MockCache{},
					&MockConfig{},
					map[string]interface{}{},
					PluginPermissions{},
					nil, // hookSystem
					nil, // eventBus
					serviceRegistry,
					NewMiddlewareRegistry(),
					nil, // permissionChecker
				)

				// Create context for plugin 2
				plugin2Ctx := NewPluginContext(
					plugin2Name,
					&MockRouter{},
					&MockLogger{},
					&MockMetrics{},
					&MockDatabase{},
					&MockCache{},
					&MockConfig{},
					map[string]interface{}{},
					PluginPermissions{},
					nil, // hookSystem
					nil, // eventBus
					serviceRegistry,
					NewMiddlewareRegistry(),
					nil, // permissionChecker
				)

				// Plugin 1 exports a service
				service1 := &TestService{Value: "service1"}
				err := plugin1Ctx.ExportService(service1Name, service1)
				if err != nil {
					return false
				}

				// Plugin 2 exports a service
				service2 := &TestService{Value: "service2"}
				err = plugin2Ctx.ExportService(service2Name, service2)
				if err != nil {
					return false
				}

				// List services for plugin 1
				services1 := serviceRegistry.List(plugin1Name)

				// Should only contain plugin 1's service
				if len(services1) != 1 {
					return false
				}
				if services1[0] != service1Name {
					return false
				}

				// List services for plugin 2
				services2 := serviceRegistry.List(plugin2Name)

				// Should only contain plugin 2's service
				if len(services2) != 1 {
					return false
				}
				if services2[0] != service2Name {
					return false
				}

				return true
			},
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
			gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		),
	)

	properties.TestingRun(t)
}

// TestService is a simple test service for inter-plugin communication testing
type TestService struct {
	Value string
}

// TestServiceMethod is a method on the test service
func (s *TestService) TestServiceMethod() string {
	return s.Value
}
