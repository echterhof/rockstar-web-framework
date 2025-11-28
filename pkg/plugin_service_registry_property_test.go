package pkg

import (
	"fmt"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// mockPluginStatusChecker is a mock implementation of PluginStatusChecker for testing
type mockPluginStatusChecker struct {
	runningPlugins map[string]bool
}

func (m *mockPluginStatusChecker) IsPluginRunning(pluginName string) bool {
	if m.runningPlugins == nil {
		return false
	}
	return m.runningPlugins[pluginName]
}

// mockService is a simple service for testing
type mockService struct {
	name string
	data interface{}
}

// TestProperty_ServiceExportAndImport tests Property 9:
// Service Export and Import
// **Feature: compile-time-plugins, Property 9: Service Export and Import**
// **Validates: Requirements 7.2, 9.1, 9.2**
func TestProperty_ServiceExportAndImport(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("exported services can be imported", prop.ForAll(
		func(pluginName string, serviceName string, serviceData string) bool {
			// Create a fresh service registry
			registry := NewServiceRegistry()

			// Create a status checker that marks the plugin as running
			statusChecker := &mockPluginStatusChecker{
				runningPlugins: map[string]bool{
					pluginName: true,
				},
			}
			registry.SetPluginStatusChecker(statusChecker)

			// Create a service
			service := &mockService{
				name: serviceName,
				data: serviceData,
			}

			// Export the service
			err := registry.Export(pluginName, serviceName, service)
			if err != nil {
				t.Logf("Failed to export service: %v", err)
				return false
			}

			// Import the service
			imported, err := registry.Import(pluginName, serviceName)
			if err != nil {
				t.Logf("Failed to import service: %v", err)
				return false
			}

			// Verify the imported service is the same
			importedService, ok := imported.(*mockService)
			if !ok {
				t.Logf("Imported service has wrong type")
				return false
			}

			if importedService.name != serviceName {
				t.Logf("Service name mismatch: expected %s, got %s", serviceName, importedService.name)
				return false
			}

			if importedService.data != serviceData {
				t.Logf("Service data mismatch: expected %s, got %s", serviceData, importedService.data)
				return false
			}

			return true
		},
		gen.Identifier(),
		gen.Identifier(),
		gen.AlphaString(),
	))

	properties.Property("import fails when plugin is not running", prop.ForAll(
		func(pluginName string, serviceName string) bool {
			// Create a fresh service registry
			registry := NewServiceRegistry()

			// Create a status checker that marks the plugin as NOT running
			statusChecker := &mockPluginStatusChecker{
				runningPlugins: map[string]bool{
					pluginName: false,
				},
			}
			registry.SetPluginStatusChecker(statusChecker)

			// Export the service (this should succeed even if plugin is not running yet)
			service := &mockService{name: serviceName}
			err := registry.Export(pluginName, serviceName, service)
			if err != nil {
				t.Logf("Failed to export service: %v", err)
				return false
			}

			// Try to import the service - this should fail because plugin is not running
			_, err = registry.Import(pluginName, serviceName)
			if err == nil {
				t.Logf("Import should have failed for non-running plugin")
				return false
			}

			// Verify the error message mentions the plugin is not running
			expectedError := fmt.Sprintf("plugin %s is not running", pluginName)
			if err.Error() != expectedError {
				t.Logf("Expected error '%s', got '%s'", expectedError, err.Error())
				return false
			}

			return true
		},
		gen.Identifier(),
		gen.Identifier(),
	))

	properties.Property("multiple services can be exported by same plugin", prop.ForAll(
		func(pluginName string, serviceNames []string) bool {
			// Create a fresh service registry
			registry := NewServiceRegistry()

			// Create a status checker that marks the plugin as running
			statusChecker := &mockPluginStatusChecker{
				runningPlugins: map[string]bool{
					pluginName: true,
				},
			}
			registry.SetPluginStatusChecker(statusChecker)

			// Export all services
			for i, serviceName := range serviceNames {
				service := &mockService{
					name: serviceName,
					data: i,
				}
				err := registry.Export(pluginName, serviceName, service)
				if err != nil {
					t.Logf("Failed to export service %s: %v", serviceName, err)
					return false
				}
			}

			// Verify all services can be imported
			for i, serviceName := range serviceNames {
				imported, err := registry.Import(pluginName, serviceName)
				if err != nil {
					t.Logf("Failed to import service %s: %v", serviceName, err)
					return false
				}

				importedService, ok := imported.(*mockService)
				if !ok {
					t.Logf("Imported service has wrong type")
					return false
				}

				if importedService.data != i {
					t.Logf("Service data mismatch for %s: expected %d, got %v", serviceName, i, importedService.data)
					return false
				}
			}

			// Verify List returns all service names
			listedServices := registry.List(pluginName)
			if len(listedServices) != len(serviceNames) {
				t.Logf("Expected %d services in list, got %d", len(serviceNames), len(listedServices))
				return false
			}

			return true
		},
		gen.Identifier(),
		gen.SliceOfN(5, gen.Identifier()).SuchThat(func(v interface{}) bool {
			// Ensure unique service names
			names := v.([]string)
			seen := make(map[string]bool)
			for _, name := range names {
				if seen[name] {
					return false
				}
				seen[name] = true
			}
			return len(names) > 0
		}),
	))

	properties.Property("cannot export duplicate service names", prop.ForAll(
		func(pluginName string, serviceName string) bool {
			// Create a fresh service registry
			registry := NewServiceRegistry()

			// Export a service
			service1 := &mockService{name: serviceName, data: "first"}
			err := registry.Export(pluginName, serviceName, service1)
			if err != nil {
				t.Logf("Failed to export first service: %v", err)
				return false
			}

			// Try to export another service with the same name
			service2 := &mockService{name: serviceName, data: "second"}
			err = registry.Export(pluginName, serviceName, service2)
			if err == nil {
				t.Logf("Should not be able to export duplicate service name")
				return false
			}

			// Verify the error message
			expectedError := fmt.Sprintf("service %s already exported by plugin %s", serviceName, pluginName)
			if err.Error() != expectedError {
				t.Logf("Expected error '%s', got '%s'", expectedError, err.Error())
				return false
			}

			return true
		},
		gen.Identifier(),
		gen.Identifier(),
	))

	properties.Property("import fails for non-existent service", prop.ForAll(
		func(pluginName string, exportedService string, requestedService string) bool {
			// Skip if the services are the same
			if exportedService == requestedService {
				return true
			}

			// Create a fresh service registry
			registry := NewServiceRegistry()

			// Create a status checker that marks the plugin as running
			statusChecker := &mockPluginStatusChecker{
				runningPlugins: map[string]bool{
					pluginName: true,
				},
			}
			registry.SetPluginStatusChecker(statusChecker)

			// Export one service
			service := &mockService{name: exportedService}
			err := registry.Export(pluginName, exportedService, service)
			if err != nil {
				t.Logf("Failed to export service: %v", err)
				return false
			}

			// Try to import a different service
			_, err = registry.Import(pluginName, requestedService)
			if err == nil {
				t.Logf("Should not be able to import non-existent service")
				return false
			}

			// Verify the error message
			expectedError := fmt.Sprintf("service %s not found in plugin %s", requestedService, pluginName)
			if err.Error() != expectedError {
				t.Logf("Expected error '%s', got '%s'", expectedError, err.Error())
				return false
			}

			return true
		},
		gen.Identifier(),
		gen.Identifier(),
		gen.Identifier(),
	))

	properties.Property("import fails for non-existent plugin", prop.ForAll(
		func(exportingPlugin string, requestedPlugin string, serviceName string) bool {
			// Skip if the plugins are the same
			if exportingPlugin == requestedPlugin {
				return true
			}

			// Create a fresh service registry
			registry := NewServiceRegistry()

			// Create a status checker that marks only the exporting plugin as running
			statusChecker := &mockPluginStatusChecker{
				runningPlugins: map[string]bool{
					exportingPlugin: true,
				},
			}
			registry.SetPluginStatusChecker(statusChecker)

			// Export a service from one plugin
			service := &mockService{name: serviceName}
			err := registry.Export(exportingPlugin, serviceName, service)
			if err != nil {
				t.Logf("Failed to export service: %v", err)
				return false
			}

			// Try to import from a different plugin
			_, err = registry.Import(requestedPlugin, serviceName)
			if err == nil {
				t.Logf("Should not be able to import from non-existent plugin")
				return false
			}

			return true
		},
		gen.Identifier(),
		gen.Identifier(),
		gen.Identifier(),
	))

	properties.Property("unregister removes service", prop.ForAll(
		func(pluginName string, serviceName string) bool {
			// Create a fresh service registry
			registry := NewServiceRegistry()

			// Create a status checker that marks the plugin as running
			statusChecker := &mockPluginStatusChecker{
				runningPlugins: map[string]bool{
					pluginName: true,
				},
			}
			registry.SetPluginStatusChecker(statusChecker)

			// Export a service
			service := &mockService{name: serviceName}
			err := registry.Export(pluginName, serviceName, service)
			if err != nil {
				t.Logf("Failed to export service: %v", err)
				return false
			}

			// Verify the service can be imported
			_, err = registry.Import(pluginName, serviceName)
			if err != nil {
				t.Logf("Failed to import service before unregister: %v", err)
				return false
			}

			// Unregister the service
			err = registry.Unregister(pluginName, serviceName)
			if err != nil {
				t.Logf("Failed to unregister service: %v", err)
				return false
			}

			// Verify the service can no longer be imported
			_, err = registry.Import(pluginName, serviceName)
			if err == nil {
				t.Logf("Should not be able to import unregistered service")
				return false
			}

			return true
		},
		gen.Identifier(),
		gen.Identifier(),
	))

	properties.Property("unregisterAll removes all services from plugin", prop.ForAll(
		func(pluginName string, serviceNames []string) bool {
			// Create a fresh service registry
			registry := NewServiceRegistry()

			// Create a status checker that marks the plugin as running
			statusChecker := &mockPluginStatusChecker{
				runningPlugins: map[string]bool{
					pluginName: true,
				},
			}
			registry.SetPluginStatusChecker(statusChecker)

			// Export all services
			for _, serviceName := range serviceNames {
				service := &mockService{name: serviceName}
				err := registry.Export(pluginName, serviceName, service)
				if err != nil {
					t.Logf("Failed to export service %s: %v", serviceName, err)
					return false
				}
			}

			// Verify all services are listed
			listedServices := registry.List(pluginName)
			if len(listedServices) != len(serviceNames) {
				t.Logf("Expected %d services before unregisterAll, got %d", len(serviceNames), len(listedServices))
				return false
			}

			// Unregister all services
			err := registry.UnregisterAll(pluginName)
			if err != nil {
				t.Logf("Failed to unregister all services: %v", err)
				return false
			}

			// Verify no services are listed
			listedServices = registry.List(pluginName)
			if len(listedServices) != 0 {
				t.Logf("Expected 0 services after unregisterAll, got %d", len(listedServices))
				return false
			}

			// Verify none of the services can be imported
			for _, serviceName := range serviceNames {
				_, err := registry.Import(pluginName, serviceName)
				if err == nil {
					t.Logf("Should not be able to import service %s after unregisterAll", serviceName)
					return false
				}
			}

			return true
		},
		gen.Identifier(),
		gen.SliceOfN(5, gen.Identifier()).SuchThat(func(v interface{}) bool {
			// Ensure unique service names
			names := v.([]string)
			seen := make(map[string]bool)
			for _, name := range names {
				if seen[name] {
					return false
				}
				seen[name] = true
			}
			return len(names) > 0
		}),
	))

	properties.Property("listAll returns all services from all plugins", prop.ForAll(
		func(pluginServices map[string][]string) bool {
			// Create a fresh service registry
			registry := NewServiceRegistry()

			// Create a status checker that marks all plugins as running
			runningPlugins := make(map[string]bool)
			for pluginName := range pluginServices {
				runningPlugins[pluginName] = true
			}
			statusChecker := &mockPluginStatusChecker{
				runningPlugins: runningPlugins,
			}
			registry.SetPluginStatusChecker(statusChecker)

			// Export all services
			for pluginName, serviceNames := range pluginServices {
				for _, serviceName := range serviceNames {
					service := &mockService{name: serviceName}
					err := registry.Export(pluginName, serviceName, service)
					if err != nil {
						t.Logf("Failed to export service %s from plugin %s: %v", serviceName, pluginName, err)
						return false
					}
				}
			}

			// Get all services
			allServices := registry.ListAll()

			// Verify the count matches
			if len(allServices) != len(pluginServices) {
				t.Logf("Expected %d plugins in ListAll, got %d", len(pluginServices), len(allServices))
				return false
			}

			// Verify each plugin's services
			for pluginName, expectedServices := range pluginServices {
				actualServices, exists := allServices[pluginName]
				if !exists {
					t.Logf("Plugin %s not found in ListAll", pluginName)
					return false
				}

				if len(actualServices) != len(expectedServices) {
					t.Logf("Expected %d services for plugin %s, got %d", len(expectedServices), pluginName, len(actualServices))
					return false
				}

				// Create a map for quick lookup
				actualMap := make(map[string]bool)
				for _, serviceName := range actualServices {
					actualMap[serviceName] = true
				}

				// Verify all expected services are present
				for _, serviceName := range expectedServices {
					if !actualMap[serviceName] {
						t.Logf("Service %s not found for plugin %s", serviceName, pluginName)
						return false
					}
				}
			}

			return true
		},
		gen.MapOf(
			gen.Identifier(),
			gen.SliceOfN(3, gen.Identifier()).SuchThat(func(v interface{}) bool {
				// Ensure unique service names
				names := v.([]string)
				seen := make(map[string]bool)
				for _, name := range names {
					if seen[name] {
						return false
					}
					seen[name] = true
				}
				return len(names) > 0
			}),
		).SuchThat(func(v interface{}) bool {
			// Ensure at least one plugin
			m := v.(map[string][]string)
			return len(m) > 0
		}),
	))

	properties.Property("registry is thread-safe for concurrent operations", prop.ForAll(
		func(pluginName string, serviceNames []string) bool {
			// Create a fresh service registry
			registry := NewServiceRegistry()

			// Create a status checker that marks the plugin as running
			statusChecker := &mockPluginStatusChecker{
				runningPlugins: map[string]bool{
					pluginName: true,
				},
			}
			registry.SetPluginStatusChecker(statusChecker)

			// Export services concurrently
			done := make(chan bool, len(serviceNames))
			for i, serviceName := range serviceNames {
				go func(name string, index int) {
					service := &mockService{name: name, data: index}
					_ = registry.Export(pluginName, name, service)
					done <- true
				}(serviceName, i)
			}

			// Wait for all exports to complete
			for i := 0; i < len(serviceNames); i++ {
				<-done
			}

			// Verify all services are listed
			listedServices := registry.List(pluginName)
			if len(listedServices) != len(serviceNames) {
				t.Logf("Expected %d services after concurrent export, got %d", len(serviceNames), len(listedServices))
				return false
			}

			// Import services concurrently
			importDone := make(chan bool, len(serviceNames))
			importErrors := make(chan error, len(serviceNames))
			for _, serviceName := range serviceNames {
				go func(name string) {
					_, err := registry.Import(pluginName, name)
					if err != nil {
						importErrors <- err
					}
					importDone <- true
				}(serviceName)
			}

			// Wait for all imports to complete
			for i := 0; i < len(serviceNames); i++ {
				<-importDone
			}

			// Check if there were any import errors
			close(importErrors)
			for err := range importErrors {
				t.Logf("Import error during concurrent operations: %v", err)
				return false
			}

			return true
		},
		gen.Identifier(),
		gen.SliceOfN(10, gen.Identifier()).SuchThat(func(v interface{}) bool {
			// Ensure unique service names
			names := v.([]string)
			seen := make(map[string]bool)
			for _, name := range names {
				if seen[name] {
					return false
				}
				seen[name] = true
			}
			return len(names) > 0
		}),
	))

	// Run all properties with 100 iterations minimum
	params := gopter.DefaultTestParameters()
	params.MinSuccessfulTests = 100
	properties.TestingRun(t, params)
}
