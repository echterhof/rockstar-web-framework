package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

// ============================================================================
// Logging Plugin Example
// ============================================================================
//
// This plugin demonstrates:
// - Request/response logging
// - Logging middleware
// - Event handling
// - Sensitive data masking
// - Async logging with buffering
//
// Requirements: 1.1, 1.2, 1.3, 2.1, 7.2, 7.5
// ============================================================================

// LoggingPlugin implements request/response logging functionality
type LoggingPlugin struct {
	ctx pkg.PluginContext
	
	// Configuration
	logRequests       bool
	logResponses      bool
	logHeaders        bool
	logBody           bool
	logQueryParams    bool
	requestLogLevel   string
	errorLogLevel     string
	outputFormat      string
	includeTimestamp  bool
	includeRequestID  bool
	asyncLogging      bool
	bufferSize        int
	excludedPaths     []string
	excludedMethods   []string
	maskHeaders       []string
	maskQueryParams   []string
	
	// Async logging
	logChannel chan *LogEntry
	stopChan   chan struct{}
}

// LogEntry represents a log entry
type LogEntry struct {
	Timestamp  time.Time              `json:"timestamp,omitempty"`
	RequestID  string                 `json:"request_id,omitempty"`
	Method     string                 `json:"method"`
	Path       string                 `json:"path"`
	Query      map[string]string      `json:"query,omitempty"`
	Headers    map[string]string      `json:"headers,omitempty"`
	Body       string                 `json:"body,omitempty"`
	StatusCode int                    `json:"status_code,omitempty"`
	Duration   time.Duration          `json:"duration,omitempty"`
	Error      string                 `json:"error,omitempty"`
	Level      string                 `json:"level"`
}

// ============================================================================
// Plugin Interface Implementation
// ============================================================================

func (p *LoggingPlugin) Name() string {
	return "logging-plugin"
}

func (p *LoggingPlugin) Version() string {
	return "1.0.0"
}

func (p *LoggingPlugin) Description() string {
	return "Request and response logging plugin with sensitive data masking"
}

func (p *LoggingPlugin) Author() string {
	return "Rockstar Framework Team"
}

func (p *LoggingPlugin) Dependencies() []pkg.PluginDependency {
	return []pkg.PluginDependency{}
}

func (p *LoggingPlugin) Initialize(ctx pkg.PluginContext) error {
	fmt.Printf("[%s] Initializing logging plugin...\n", p.Name())
	
	p.ctx = ctx
	
	// Parse configuration
	if err := p.parseConfig(ctx.PluginConfig()); err != nil {
		return fmt.Errorf("failed to parse configuration: %w", err)
	}
	
	// Initialize async logging if enabled
	if p.asyncLogging {
		p.logChannel = make(chan *LogEntry, p.bufferSize)
		p.stopChan = make(chan struct{})
	}
	
	// Register hooks
	if err := p.registerHooks(); err != nil {
		return fmt.Errorf("failed to register hooks: %w", err)
	}
	
	// Register middleware
	if err := p.registerMiddleware(); err != nil {
		return fmt.Errorf("failed to register middleware: %w", err)
	}
	
	// Subscribe to events
	if err := p.subscribeToEvents(); err != nil {
		return fmt.Errorf("failed to subscribe to events: %w", err)
	}
	
	fmt.Printf("[%s] Initialization complete\n", p.Name())
	return nil
}

func (p *LoggingPlugin) Start() error {
	fmt.Printf("[%s] Starting logging plugin...\n", p.Name())
	
	// Start async logging worker if enabled
	if p.asyncLogging {
		go p.logWorker()
	}
	
	fmt.Printf("[%s] Plugin started\n", p.Name())
	return nil
}

func (p *LoggingPlugin) Stop() error {
	fmt.Printf("[%s] Stopping logging plugin...\n", p.Name())
	
	// Stop async logging worker
	if p.asyncLogging && p.stopChan != nil {
		close(p.stopChan)
		// Drain remaining logs
		time.Sleep(100 * time.Millisecond)
	}
	
	fmt.Printf("[%s] Plugin stopped\n", p.Name())
	return nil
}

func (p *LoggingPlugin) Cleanup() error {
	fmt.Printf("[%s] Cleaning up logging plugin...\n", p.Name())
	
	if p.logChannel != nil {
		close(p.logChannel)
	}
	
	p.ctx = nil
	
	fmt.Printf("[%s] Cleanup complete\n", p.Name())
	return nil
}

func (p *LoggingPlugin) ConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"log_requests": map[string]interface{}{
			"type":        "bool",
			"default":     true,
			"description": "Enable request logging",
		},
		"log_responses": map[string]interface{}{
			"type":        "bool",
			"default":     true,
			"description": "Enable response logging",
		},
		"log_headers": map[string]interface{}{
			"type":        "bool",
			"default":     true,
			"description": "Include headers in logs",
		},
		"log_body": map[string]interface{}{
			"type":        "bool",
			"default":     false,
			"description": "Include request/response body in logs",
		},
		"log_query_params": map[string]interface{}{
			"type":        "bool",
			"default":     true,
			"description": "Include query parameters in logs",
		},
		"output_format": map[string]interface{}{
			"type":        "string",
			"default":     "json",
			"description": "Log output format (json, text, or structured)",
		},
		"async_logging": map[string]interface{}{
			"type":        "bool",
			"default":     true,
			"description": "Enable asynchronous logging",
		},
		"buffer_size": map[string]interface{}{
			"type":        "int",
			"default":     1000,
			"description": "Async logging buffer size",
		},
		"excluded_paths": map[string]interface{}{
			"type":        "array",
			"items":       "string",
			"default":     []string{"/health", "/metrics"},
			"description": "Paths to exclude from logging",
		},
		"mask_headers": map[string]interface{}{
			"type":        "array",
			"items":       "string",
			"default":     []string{"Authorization", "Cookie"},
			"description": "Headers to mask in logs",
		},
	}
}

func (p *LoggingPlugin) OnConfigChange(config map[string]interface{}) error {
	fmt.Printf("[%s] Configuration updated\n", p.Name())
	return p.parseConfig(config)
}

// ============================================================================
// Configuration Parsing
// ============================================================================

func (p *LoggingPlugin) parseConfig(config map[string]interface{}) error {
	// Log Requests
	if val, ok := config["log_requests"].(bool); ok {
		p.logRequests = val
	} else {
		p.logRequests = true
	}
	
	// Log Responses
	if val, ok := config["log_responses"].(bool); ok {
		p.logResponses = val
	} else {
		p.logResponses = true
	}
	
	// Log Headers
	if val, ok := config["log_headers"].(bool); ok {
		p.logHeaders = val
	} else {
		p.logHeaders = true
	}
	
	// Log Body
	if val, ok := config["log_body"].(bool); ok {
		p.logBody = val
	} else {
		p.logBody = false
	}
	
	// Log Query Params
	if val, ok := config["log_query_params"].(bool); ok {
		p.logQueryParams = val
	} else {
		p.logQueryParams = true
	}
	
	// Output Format
	if val, ok := config["output_format"].(string); ok {
		p.outputFormat = val
	} else {
		p.outputFormat = "json"
	}
	
	// Async Logging
	if val, ok := config["async_logging"].(bool); ok {
		p.asyncLogging = val
	} else {
		p.asyncLogging = true
	}
	
	// Buffer Size
	if val, ok := config["buffer_size"].(int); ok {
		p.bufferSize = val
	} else if val, ok := config["buffer_size"].(float64); ok {
		p.bufferSize = int(val)
	} else {
		p.bufferSize = 1000
	}
	
	// Excluded Paths
	if paths, ok := config["excluded_paths"].([]interface{}); ok {
		p.excludedPaths = make([]string, 0, len(paths))
		for _, path := range paths {
			if pathStr, ok := path.(string); ok {
				p.excludedPaths = append(p.excludedPaths, pathStr)
			}
		}
	} else {
		p.excludedPaths = []string{"/health", "/metrics"}
	}
	
	// Mask Headers
	if headers, ok := config["mask_headers"].([]interface{}); ok {
		p.maskHeaders = make([]string, 0, len(headers))
		for _, header := range headers {
			if headerStr, ok := header.(string); ok {
				p.maskHeaders = append(p.maskHeaders, headerStr)
			}
		}
	} else {
		p.maskHeaders = []string{"Authorization", "Cookie"}
	}
	
	return nil
}

// ============================================================================
// Hook Registration
// ============================================================================

func (p *LoggingPlugin) registerHooks() error {
	// Register pre-request hook
	err := p.ctx.RegisterHook(pkg.HookTypePreRequest, 50, func(hookCtx pkg.HookContext) error {
		if !p.logRequests {
			return nil
		}
		
		reqCtx := hookCtx.Context()
		if reqCtx == nil {
			return nil
		}
		
		req := reqCtx.Request()
		if req == nil || req.URL == nil {
			return nil
		}
		path := req.URL.Path
		if p.isPathExcluded(path) {
			return nil
		}
		
		// Store request start time
		hookCtx.Set("request_start_time", time.Now())
		
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to register pre-request hook: %w", err)
	}
	
	// Register post-request hook
	err = p.ctx.RegisterHook(pkg.HookTypePostRequest, 50, func(hookCtx pkg.HookContext) error {
		if !p.logResponses {
			return nil
		}
		
		reqCtx := hookCtx.Context()
		if reqCtx == nil {
			return nil
		}
		
		req := reqCtx.Request()
		if req == nil || req.URL == nil {
			return nil
		}
		path := req.URL.Path
		if p.isPathExcluded(path) {
			return nil
		}
		
		// Calculate request duration
		var duration time.Duration
		if startTime := hookCtx.Get("request_start_time"); startTime != nil {
			if t, ok := startTime.(time.Time); ok {
				duration = time.Since(t)
			}
		}
		
		// Create log entry
		entry := &LogEntry{
			Method:   req.Method,
			Path:     path,
			Duration: duration,
			Level:    p.requestLogLevel,
		}
		
		if p.includeTimestamp {
			entry.Timestamp = time.Now()
		}
		
		if p.logQueryParams {
			entry.Query = p.getQueryParams(reqCtx)
		}
		
		if p.logHeaders {
			entry.Headers = p.getHeaders(reqCtx)
		}
		
		// Log the entry
		p.log(entry)
		
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to register post-request hook: %w", err)
	}
	
	// Register error hook
	err = p.ctx.RegisterHook(pkg.HookTypeError, 50, func(hookCtx pkg.HookContext) error {
		reqCtx := hookCtx.Context()
		if reqCtx == nil {
			return nil
		}
		
		req := reqCtx.Request()
		if req == nil || req.URL == nil {
			return nil
		}
		
		// Log error
		entry := &LogEntry{
			Method: req.Method,
			Path:   req.URL.Path,
			Level:  p.errorLogLevel,
		}
		
		if p.includeTimestamp {
			entry.Timestamp = time.Now()
		}
		
		p.log(entry)
		
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to register error hook: %w", err)
	}
	
	return nil
}

// ============================================================================
// Middleware Registration
// ============================================================================

func (p *LoggingPlugin) registerMiddleware() error {
	// Register logging middleware
	loggingMiddleware := func(ctx pkg.Context, next pkg.HandlerFunc) error {
		req := ctx.Request()
		if req == nil || req.URL == nil {
			return next(ctx)
		}
		path := req.URL.Path
		
		// Skip excluded paths
		if p.isPathExcluded(path) {
			return next(ctx)
		}
		
		// Record start time
		startTime := time.Now()
		
		// Call next handler
		err := next(ctx)
		
		// Log request/response
		if p.logRequests || p.logResponses {
			entry := &LogEntry{
				Method:   req.Method,
				Path:     path,
				Duration: time.Since(startTime),
				Level:    p.requestLogLevel,
			}
			
			if p.includeTimestamp {
				entry.Timestamp = time.Now()
			}
			
			if p.logQueryParams {
				entry.Query = p.getQueryParams(ctx)
			}
			
			if p.logHeaders {
				entry.Headers = p.getHeaders(ctx)
			}
			
			if err != nil {
				entry.Error = err.Error()
				entry.Level = p.errorLogLevel
			}
			
			p.log(entry)
		}
		
		return err
	}
	
	// Register the middleware
	err := p.ctx.RegisterMiddleware("logging", loggingMiddleware, 50, []string{})
	if err != nil {
		return fmt.Errorf("failed to register logging middleware: %w", err)
	}
	
	fmt.Printf("[%s] Registered logging middleware\n", p.Name())
	return nil
}

// ============================================================================
// Event Handling
// ============================================================================

func (p *LoggingPlugin) subscribeToEvents() error {
	// Subscribe to authentication events for logging
	err := p.ctx.SubscribeEvent("auth.login", func(event pkg.Event) error {
		fmt.Printf("[%s] Event: %s - %v\n", p.Name(), event.Name, event.Data)
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to auth.login: %w", err)
	}
	
	return nil
}

// ============================================================================
// Helper Methods
// ============================================================================

func (p *LoggingPlugin) isPathExcluded(path string) bool {
	for _, excluded := range p.excludedPaths {
		if strings.HasPrefix(path, excluded) {
			return true
		}
	}
	return false
}

func (p *LoggingPlugin) getQueryParams(ctx pkg.Context) map[string]string {
	params := make(map[string]string)
	// In a real implementation, extract query parameters from context
	// and mask sensitive parameters
	return params
}

func (p *LoggingPlugin) getHeaders(ctx pkg.Context) map[string]string {
	headers := make(map[string]string)
	// In a real implementation, extract headers from context
	// and mask sensitive headers
	for _, maskHeader := range p.maskHeaders {
		headers[maskHeader] = "***MASKED***"
	}
	return headers
}

func (p *LoggingPlugin) log(entry *LogEntry) {
	if p.asyncLogging && p.logChannel != nil {
		// Send to async worker
		select {
		case p.logChannel <- entry:
		default:
			// Buffer full, log synchronously
			p.writeLog(entry)
		}
	} else {
		// Log synchronously
		p.writeLog(entry)
	}
}

func (p *LoggingPlugin) writeLog(entry *LogEntry) {
	switch p.outputFormat {
	case "json":
		data, _ := json.Marshal(entry)
		fmt.Printf("[%s] %s\n", p.Name(), string(data))
	case "text":
		fmt.Printf("[%s] %s %s - %v\n", p.Name(), entry.Method, entry.Path, entry.Duration)
	default:
		fmt.Printf("[%s] %+v\n", p.Name(), entry)
	}
}

func (p *LoggingPlugin) logWorker() {
	for {
		select {
		case entry := <-p.logChannel:
			if entry != nil {
				p.writeLog(entry)
			}
		case <-p.stopChan:
			// Drain remaining logs
			for len(p.logChannel) > 0 {
				entry := <-p.logChannel
				p.writeLog(entry)
			}
			return
		}
	}
}

// ============================================================================
// Plugin Entry Point
// ============================================================================

func NewPlugin() pkg.Plugin {
	return &LoggingPlugin{}
}

func main() {
	// Plugin entry point
}
