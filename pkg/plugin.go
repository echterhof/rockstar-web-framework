package pkg

import (
	"time"
)

// Plugin is the core interface that all plugins must implement
type Plugin interface {
	// Metadata
	Name() string
	Version() string
	Description() string
	Author() string

	// Dependencies
	Dependencies() []PluginDependency

	// Lifecycle
	Initialize(ctx PluginContext) error
	Start() error
	Stop() error
	Cleanup() error

	// Configuration
	ConfigSchema() map[string]interface{}
	OnConfigChange(config map[string]interface{}) error
}

// PluginDependency represents a dependency on another plugin or framework version
type PluginDependency struct {
	Name             string
	Version          string // Semantic version constraint (e.g., ">=1.0.0,<2.0.0")
	Optional         bool
	FrameworkVersion string // Framework version requirement
}

// PluginStatus represents the current state of a plugin
type PluginStatus string

const (
	PluginStatusUnloaded    PluginStatus = "unloaded"
	PluginStatusLoading     PluginStatus = "loading"
	PluginStatusInitialized PluginStatus = "initialized"
	PluginStatusRunning     PluginStatus = "running"
	PluginStatusStopped     PluginStatus = "stopped"
	PluginStatusError       PluginStatus = "error"
)

// PluginInfo contains metadata about a loaded plugin
type PluginInfo struct {
	Name        string
	Version     string
	Description string
	Author      string
	Loaded      bool
	Enabled     bool
	LoadTime    time.Time
	Status      PluginStatus
}

// PluginHealth tracks the health and performance of a plugin
type PluginHealth struct {
	Status      PluginStatus
	ErrorCount  int64
	LastError   error
	LastErrorAt time.Time
	HookMetrics map[string]HookMetrics
}

// HookMetrics tracks execution metrics for plugin hooks
type HookMetrics struct {
	ExecutionCount  int64
	TotalDuration   time.Duration
	AverageDuration time.Duration
	ErrorCount      int64
}

// HookType defines the type of lifecycle hook
type HookType string

const (
	HookTypeStartup      HookType = "startup"
	HookTypeShutdown     HookType = "shutdown"
	HookTypePreRequest   HookType = "pre_request"
	HookTypePostRequest  HookType = "post_request"
	HookTypePreResponse  HookType = "pre_response"
	HookTypePostResponse HookType = "post_response"
	HookTypeError        HookType = "error"
)

// HookHandler is a function that handles a hook event
type HookHandler func(ctx HookContext) error

// HookContext provides context for hook execution
type HookContext interface {
	// Request context (nil for non-request hooks)
	Context() Context

	// Hook metadata
	HookType() HookType
	PluginName() string

	// Data passing between hooks
	Set(key string, value interface{})
	Get(key string) interface{}

	// Control flow
	Skip() // Skip remaining hooks
	IsSkipped() bool
}

// PluginContext provides isolated access to framework services
type PluginContext interface {
	// Core framework access
	Router() RouterEngine
	Logger() Logger
	Metrics() MetricsCollector

	// Data access (permission-controlled)
	Database() DatabaseManager
	Cache() CacheManager
	Config() ConfigManager
	FileSystem() FileManager
	Network() NetworkClient

	// Plugin-specific
	PluginConfig() map[string]interface{}
	PluginStorage() PluginStorage

	// Hook registration
	RegisterHook(hookType HookType, priority int, handler HookHandler) error

	// Event system
	PublishEvent(event string, data interface{}) error
	SubscribeEvent(event string, handler EventHandler) error

	// Service export/import
	ExportService(name string, service interface{}) error
	ImportService(pluginName, serviceName string) (interface{}, error)

	// Middleware registration
	RegisterMiddleware(name string, handler MiddlewareFunc, priority int, routes []string) error
	UnregisterMiddleware(name string) error
}

// PluginStorage provides isolated key-value storage for plugins
type PluginStorage interface {
	Get(key string) (interface{}, error)
	Set(key string, value interface{}) error
	Delete(key string) error
	List() ([]string, error)
	Clear() error
}

// EventBus manages inter-plugin communication through events
type EventBus interface {
	Publish(event string, data interface{}) error
	Subscribe(pluginName, event string, handler EventHandler) error
	UnregisterAll(pluginName string) error
	Unsubscribe(pluginName, event string) error
	ListSubscriptions(event string) []string
}

// EventHandler is a function that handles an event
type EventHandler func(event Event) error

// Event represents an event in the system
type Event struct {
	Name      string
	Data      interface{}
	Source    string // Plugin name
	Timestamp time.Time
}

// PluginPermissions defines what operations a plugin is allowed to perform
type PluginPermissions struct {
	// AllowDatabase grants permission to access the database.
	// Default: false
	AllowDatabase bool

	// AllowCache grants permission to access the cache.
	// Default: false
	AllowCache bool

	// AllowConfig grants permission to access configuration.
	// Default: false
	AllowConfig bool

	// AllowRouter grants permission to modify routing.
	// Default: false
	AllowRouter bool

	// AllowFileSystem grants permission to access the file system.
	// Default: false
	AllowFileSystem bool

	// AllowNetwork grants permission to make network requests.
	// Default: false
	AllowNetwork bool

	// AllowExec grants permission to execute external commands.
	// Default: false
	AllowExec bool

	// CustomPermissions provides additional custom permission flags.
	// Default: empty map
	CustomPermissions map[string]bool
}

// PermissionChecker verifies and manages plugin permissions
type PermissionChecker interface {
	CheckPermission(pluginName string, permission string) error
	GrantPermission(pluginName string, permission string) error
	RevokePermission(pluginName string, permission string) error
	GetPermissions(pluginName string) PluginPermissions
}

// NetworkClient provides network access for plugins
type NetworkClient interface {
	// HTTP methods
	Get(url string, headers map[string]string) ([]byte, error)
	Post(url string, body []byte, headers map[string]string) ([]byte, error)
	Put(url string, body []byte, headers map[string]string) ([]byte, error)
	Delete(url string, headers map[string]string) ([]byte, error)

	// Generic request
	Do(method, url string, body []byte, headers map[string]string) ([]byte, int, error)
}
