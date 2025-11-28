//go:build ignore

package main

import (
	"fmt"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

// This example demonstrates the plugin monitoring and metrics functionality.
// It shows how to:
// 1. Access plugin metrics
// 2. Monitor plugin health
// 3. Export metrics in Prometheus format

func main() {
	fmt.Println("Plugin Monitoring and Metrics Example")
	fmt.Println("======================================")
	fmt.Println()

	// Create a simple plugin metrics collector
	collector := pkg.NewPluginMetricsCollector()

	// Simulate plugin lifecycle and collect metrics
	fmt.Println("Simulating plugin lifecycle...")
	fmt.Println()

	// Plugin 1: Successful initialization and start
	plugin1Metrics := collector.GetOrCreate("auth-plugin")
	plugin1Metrics.RecordInit(100*time.Millisecond, nil)
	plugin1Metrics.RecordStart(50*time.Millisecond, nil)
	plugin1Metrics.RecordHookExecution(pkg.HookTypePreRequest, 5*time.Millisecond)
	plugin1Metrics.RecordHookExecution(pkg.HookTypePreRequest, 7*time.Millisecond)
	plugin1Metrics.RecordEventPublished()
	plugin1Metrics.RecordServiceCall()

	// Plugin 2: Successful initialization and start with more activity
	plugin2Metrics := collector.GetOrCreate("cache-plugin")
	plugin2Metrics.RecordInit(80*time.Millisecond, nil)
	plugin2Metrics.RecordStart(30*time.Millisecond, nil)
	plugin2Metrics.RecordHookExecution(pkg.HookTypePostRequest, 3*time.Millisecond)
	plugin2Metrics.RecordHookExecution(pkg.HookTypePostRequest, 4*time.Millisecond)
	plugin2Metrics.RecordHookExecution(pkg.HookTypePostRequest, 5*time.Millisecond)
	plugin2Metrics.RecordEventReceived()
	plugin2Metrics.RecordEventReceived()
	plugin2Metrics.RecordServiceCall()
	plugin2Metrics.RecordServiceCall()

	// Plugin 3: Plugin with errors
	plugin3Metrics := collector.GetOrCreate("logging-plugin")
	plugin3Metrics.RecordInit(120*time.Millisecond, nil)
	plugin3Metrics.RecordStart(60*time.Millisecond, nil)
	plugin3Metrics.RecordError(fmt.Errorf("connection timeout"))
	plugin3Metrics.RecordError(fmt.Errorf("write failed"))

	// Display individual plugin metrics
	fmt.Println("Individual Plugin Metrics:")
	fmt.Println("--------------------------")
	for name, metrics := range collector.GetAll() {
		fmt.Printf("\nPlugin: %s\n", name)
		fmt.Printf("  Status: %s\n", metrics.GetStatus())
		fmt.Printf("  Init Duration: %v\n", metrics.InitDuration)
		fmt.Printf("  Start Duration: %v\n", metrics.StartDuration)
		fmt.Printf("  Error Count: %d\n", metrics.GetErrorCount())
		fmt.Printf("  Events Published: %d\n", metrics.EventsPublished)
		fmt.Printf("  Events Received: %d\n", metrics.EventsReceived)
		fmt.Printf("  Service Calls: %d\n", metrics.ServiceCalls)

		// Show hook metrics
		hookMetrics := metrics.GetHookMetrics()
		if len(hookMetrics) > 0 {
			fmt.Println("  Hook Metrics:")
			for hookType, hm := range hookMetrics {
				fmt.Printf("    %s: %d executions, total: %v, avg: %v\n",
					hookType, hm.ExecutionCount, hm.TotalDuration, hm.AverageDuration)
			}
		}

		// Show last error if any
		lastErr, lastErrAt := metrics.GetLastError()
		if lastErr != nil {
			fmt.Printf("  Last Error: %v (at %s)\n", lastErr, lastErrAt.Format("15:04:05"))
		}
	}

	// Export Prometheus metrics
	fmt.Println("\n\nPrometheus Metrics Export:")
	fmt.Println("--------------------------")
	prometheusOutput := collector.ExportPrometheus()
	fmt.Println(prometheusOutput)

	// Demonstrate concurrent metric updates
	fmt.Println("\nDemonstrating concurrent metric updates...")
	done := make(chan bool)

	// Simulate concurrent hook executions
	for i := 0; i < 5; i++ {
		go func(id int) {
			for j := 0; j < 10; j++ {
				plugin1Metrics.RecordHookExecution(pkg.HookTypePreRequest, time.Duration(id+1)*time.Millisecond)
				time.Sleep(10 * time.Millisecond)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		<-done
	}

	fmt.Println("Concurrent updates complete!")
	fmt.Printf("Total PreRequest hook executions for auth-plugin: %d\n",
		plugin1Metrics.HookExecutions[pkg.HookTypePreRequest])

	// Show final metrics summary
	fmt.Println("\n\nFinal Metrics Summary:")
	fmt.Println("----------------------")
	allMetrics := collector.GetAll()
	fmt.Printf("Total plugins monitored: %d\n", len(allMetrics))

	var totalErrors int64
	var totalEvents int64
	var totalServiceCalls int64

	for _, metrics := range allMetrics {
		totalErrors += metrics.GetErrorCount()
		totalEvents += metrics.EventsPublished + metrics.EventsReceived
		totalServiceCalls += metrics.ServiceCalls
	}

	fmt.Printf("Total errors across all plugins: %d\n", totalErrors)
	fmt.Printf("Total events (published + received): %d\n", totalEvents)
	fmt.Printf("Total service calls: %d\n", totalServiceCalls)

	fmt.Println("\n\nExample complete!")
	fmt.Println("\nIn a real application, you would:")
	fmt.Println("  1. Access metrics via PluginManager.GetAllPluginMetrics()")
	fmt.Println("  2. Expose health checks at /_health/plugins")
	fmt.Println("  3. Expose Prometheus metrics at /_metrics/plugins")
	fmt.Println("  4. Use RegisterPluginHealthEndpoints() to set up endpoints")
}
