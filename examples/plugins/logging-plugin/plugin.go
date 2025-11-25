package main

import (
	"fmt"
	"time"

	"github.com/yourusername/rockstar/pkg"
)

// LoggingPlugin demonstrates request/response logging
type LoggingPlugin struct {
	ctx           pkg.PluginContext
	logRequests   bool
	logResponses  bool
	logHeaders    bool
	requestCount  int64
	responseCount int64
}

// Name returns the plugin name
func (p *LoggingPlugin) Name() string {
	return "logging-plugin"
}

// Version returns the plugin version
func (p *LoggingPlugin) Version() string {
	return "1.0.0"
}

// Description returns the plugin description
func (p *LoggingPlugin) Description() string {
	return "Logging plugin demonstrating request/response logging"
}

// Author returns the plugin author
func (p *LoggingPlugin) Author() string {
	return "Rockstar Framework Team"
}

// Dependencies returns the plugin dependencies
func (p *LoggingPlugin) Dependencies() []pkg.PluginDependency {
	return []pkg.PluginDependency{}
}

// Initialize initializes the plugin with the provided context
func (p *LoggingPlugin) Initialize(ctx pkg.PluginContext) error {
	p.ctx = ctx
	
	// Get configuration
	config := ctx.PluginConfig()
	if logReq, ok := config["log_requests"].(bool); ok {
		p.logRequests = logReq
	} else {
		p.logRequests = true
	}
	
	if logResp, ok := config["log_responses"].(bool); ok {
		p.logResponses = logResp
	} else {
		p.logResponses = true
	}
	
	if logHdr, ok := config["log_headers"].(bool); ok {
		p.logHeaders = logHdr
	} else {
		p.logHeaders = false
	}
	
	// Register pre-request hook for request logging
	if p.logRequests {
		err := ctx.RegisterHook(pkg.HookTypePreRequest, 50, p.logRequest)
		if err != nil {
			return fmt.Errorf("failed to register request logging hook: %w", err)
		}
	}
	
	// Register post-request hook for response logging
	if p.logResponses {
		err := ctx.RegisterHook(pkg.HookTypePostRequest, 50, p.logResponse)
		if err != nil {
			return fmt.Errorf("failed to register response logging hook: %w", err)
		}
	}
	
	// Export logging service
	err := ctx.ExportService("LoggingService", &LoggingService{plugin: p})
	if err != nil {
		return fmt.Errorf("failed to export logging service: %w", err)
	}
	
	if logger := ctx.Logger(); logger != nil {
		logger.Info("Logging plugin initialized")
	}
	
	return nil
}

// Start starts the plugin
func (p *LoggingPlugin) Start() error {
	if logger := p.ctx.Logger(); logger != nil {
		logger.Info("Logging plugin started")
	}
	return nil
}

// Stop stops the plugin
func (p *LoggingPlugin) Stop() error {
	if logger := p.ctx.Logger(); logger != nil {
		logger.Info(fmt.Sprintf("Logging plugin stopped. Logged %d requests and %d responses", 
			p.requestCount, p.responseCount))
	}
	return nil
}

// Cleanup cleans up plugin resources
func (p *LoggingPlugin) Cleanup() error {
	if logger := p.ctx.Logger(); logger != nil {
		logger.Info("Logging plugin cleaned up")
	}
	return nil
}

// ConfigSchema returns the configuration schema
func (p *LoggingPlugin) ConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"log_requests": map[string]interface{}{
			"type":        "bool",
			"default":     true,
			"description": "Whether to log incoming requests",
		},
		"log_responses": map[string]interface{}{
			"type":        "bool",
			"default":     true,
			"description": "Whether to log outgoing responses",
		},
		"log_headers": map[string]interface{}{
			"type":        "bool",
			"default":     false,
			"description": "Whether to log request/response headers",
		},
		"log_body": map[string]interface{}{
			"type":        "bool",
			"default":     false,
			"description": "Whether to log request/response bodies",
		},
	}
}

// OnConfigChange handles configuration changes
func (p *LoggingPlugin) OnConfigChange(config map[string]interface{}) error {
	if logReq, ok := config["log_requests"].(bool); ok {
		p.logRequests = logReq
	}
	
	if logResp, ok := config["log_responses"].(bool); ok {
		p.logResponses = logResp
	}
	
	if logHdr, ok := config["log_headers"].(bool); ok {
		p.logHeaders = logHdr
	}
	
	if logger := p.ctx.Logger(); logger != nil {
		logger.Info("Logging plugin configuration updated")
	}
	
	return nil
}

// logRequest logs incoming requests
func (p *LoggingPlugin) logRequest(ctx pkg.HookContext) error {
	reqCtx := ctx.Context()
	if reqCtx == nil {
		return nil
	}
	
	p.requestCount++
	
	logger := p.ctx.Logger()
	if logger == nil {
		return nil
	}
	
	// Log basic request info
	logger.Info(fmt.Sprintf("[REQUEST] %s %s from %s", 
		reqCtx.Method(), 
		reqCtx.Path(), 
		reqCtx.IP()))
	
	// Log headers if enabled
	if p.logHeaders {
		headers := reqCtx.GetHeaders()
		if len(headers) > 0 {
			logger.Info(fmt.Sprintf("[REQUEST HEADERS] %v", headers))
		}
	}
	
	// Store request start time for duration calculation
	ctx.Set("request_start_time", time.Now())
	
	// Record metrics
	if metrics := p.ctx.Metrics(); metrics != nil {
		metrics.IncrementCounter("logging_plugin.requests_logged", 1)
	}
	
	return nil
}

// logResponse logs outgoing responses
func (p *LoggingPlugin) logResponse(ctx pkg.HookContext) error {
	reqCtx := ctx.Context()
	if reqCtx == nil {
		return nil
	}
	
	p.responseCount++
	
	logger := p.ctx.Logger()
	if logger == nil {
		return nil
	}
	
	// Calculate request duration
	var duration time.Duration
	if startTime, ok := ctx.Get("request_start_time").(time.Time); ok {
		duration = time.Since(startTime)
	}
	
	// Log response info
	logger.Info(fmt.Sprintf("[RESPONSE] %s %s - Status: %d - Duration: %v", 
		reqCtx.Method(), 
		reqCtx.Path(),
		reqCtx.Status(),
		duration))
	
	// Record metrics
	if metrics := p.ctx.Metrics(); metrics != nil {
		metrics.IncrementCounter("logging_plugin.responses_logged", 1)
		metrics.RecordHistogram("logging_plugin.request_duration_ms", float64(duration.Milliseconds()))
	}
	
	return nil
}

// LoggingService is an exported service for other plugins
type LoggingService struct {
	plugin *LoggingPlugin
}

// GetRequestCount returns the number of logged requests
func (s *LoggingService) GetRequestCount() int64 {
	return s.plugin.requestCount
}

// GetResponseCount returns the number of logged responses
func (s *LoggingService) GetResponseCount() int64 {
	return s.plugin.responseCount
}

// NewPlugin creates a new instance of the plugin
func NewPlugin() pkg.Plugin {
	return &LoggingPlugin{}
}
