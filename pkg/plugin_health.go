package pkg

import (
	"net/http"
)

// PluginHealthResponse represents the health check response
type PluginHealthResponse struct {
	Status  string                      `json:"status"`
	Plugins map[string]PluginHealthInfo `json:"plugins"`
}

// PluginHealthInfo contains health information for a single plugin
type PluginHealthInfo struct {
	Name            string `json:"name"`
	Status          string `json:"status"`
	Healthy         bool   `json:"healthy"`
	ErrorCount      int64  `json:"error_count"`
	LastError       string `json:"last_error,omitempty"`
	LastErrorAt     string `json:"last_error_at,omitempty"`
	InitDuration    string `json:"init_duration"`
	StartDuration   string `json:"start_duration"`
	HookExecutions  int64  `json:"hook_executions"`
	EventsPublished int64  `json:"events_published"`
	EventsReceived  int64  `json:"events_received"`
	ServiceCalls    int64  `json:"service_calls"`
}

// PluginHealthHandler creates a handler function for plugin health checks
func PluginHealthHandler(pm PluginManager) func(ctx Context) error {
	return func(ctx Context) error {
		// Get all plugin health information
		allHealth := pm.GetAllHealth()
		allMetrics := pm.GetAllPluginMetrics()

		// Build response
		response := PluginHealthResponse{
			Status:  "ok",
			Plugins: make(map[string]PluginHealthInfo),
		}

		overallHealthy := true

		for name, health := range allHealth {
			metrics := allMetrics[name]
			if metrics == nil {
				continue
			}

			// Determine if plugin is healthy
			healthy := health.Status == PluginStatusRunning && health.ErrorCount < 10

			if !healthy {
				overallHealthy = false
			}

			// Calculate total hook executions
			var totalHookExecutions int64
			for _, count := range metrics.HookExecutions {
				totalHookExecutions += count
			}

			info := PluginHealthInfo{
				Name:            name,
				Status:          string(health.Status),
				Healthy:         healthy,
				ErrorCount:      health.ErrorCount,
				InitDuration:    metrics.InitDuration.String(),
				StartDuration:   metrics.StartDuration.String(),
				HookExecutions:  totalHookExecutions,
				EventsPublished: metrics.EventsPublished,
				EventsReceived:  metrics.EventsReceived,
				ServiceCalls:    metrics.ServiceCalls,
			}

			if health.LastError != nil {
				info.LastError = health.LastError.Error()
				info.LastErrorAt = health.LastErrorAt.Format("2006-01-02T15:04:05Z07:00")
			}

			response.Plugins[name] = info
		}

		// Set overall status
		if !overallHealthy {
			response.Status = "degraded"
		}

		// Set status code based on overall health
		statusCode := http.StatusOK
		if !overallHealthy {
			statusCode = http.StatusServiceUnavailable
		}

		// Send JSON response
		return ctx.Response().WriteJSON(statusCode, response)
	}
}

// PluginMetricsHandler creates a handler function for Prometheus metrics
func PluginMetricsHandler(pm PluginManager) func(ctx Context) error {
	return func(ctx Context) error {
		// Export metrics in Prometheus format
		metrics := pm.ExportPrometheusMetrics()

		// Set content type for Prometheus
		ctx.Response().SetContentType("text/plain; version=0.0.4")

		// Write metrics
		return ctx.Response().WriteString(http.StatusOK, metrics)
	}
}

// RegisterPluginHealthEndpoints registers health check and metrics endpoints
func RegisterPluginHealthEndpoints(router RouterEngine, pm PluginManager) {
	// Register health check endpoint
	router.GET("/_health/plugins", PluginHealthHandler(pm))

	// Register metrics endpoint
	router.GET("/_metrics/plugins", PluginMetricsHandler(pm))
}
