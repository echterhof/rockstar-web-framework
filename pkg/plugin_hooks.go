package pkg

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// hookSystemImpl is the default implementation of HookSystem
type hookSystemImpl struct {
	mu      sync.RWMutex
	hooks   map[HookType][]hookEntry
	logger  Logger
	metrics MetricsCollector
}

// hookEntry represents a registered hook with its metadata
type hookEntry struct {
	pluginName string
	priority   int
	handler    HookHandler
	metrics    *hookMetrics
}

// hookMetrics tracks execution metrics for a hook
type hookMetrics struct {
	mu             sync.Mutex
	executionCount int64
	totalDuration  time.Duration
	errorCount     int64
}

// NewHookSystem creates a new hook system
func NewHookSystem(logger Logger, metrics MetricsCollector) HookSystem {
	return &hookSystemImpl{
		hooks:   make(map[HookType][]hookEntry),
		logger:  logger,
		metrics: metrics,
	}
}

// RegisterHook registers a hook with the hook system
func (h *hookSystemImpl) RegisterHook(pluginName string, hookType HookType, priority int, handler HookHandler) error {
	if pluginName == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}
	if handler == nil {
		return fmt.Errorf("hook handler cannot be nil")
	}
	if !isValidHookType(hookType) {
		return fmt.Errorf("invalid hook type: %s", hookType)
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	entry := hookEntry{
		pluginName: pluginName,
		priority:   priority,
		handler:    handler,
		metrics:    &hookMetrics{},
	}

	h.hooks[hookType] = append(h.hooks[hookType], entry)

	// Sort hooks by priority (highest first)
	sort.Slice(h.hooks[hookType], func(i, j int) bool {
		return h.hooks[hookType][i].priority > h.hooks[hookType][j].priority
	})

	if h.logger != nil {
		h.logger.Info(fmt.Sprintf("Registered %s hook for plugin %s with priority %d", hookType, pluginName, priority))
	}

	return nil
}

// UnregisterHook removes a hook from the hook system
func (h *hookSystemImpl) UnregisterHook(pluginName string, hookType HookType) error {
	if pluginName == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}
	if !isValidHookType(hookType) {
		return fmt.Errorf("invalid hook type: %s", hookType)
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	hooks, exists := h.hooks[hookType]
	if !exists {
		return fmt.Errorf("no hooks registered for type %s", hookType)
	}

	filtered := make([]hookEntry, 0, len(hooks))
	found := false
	for _, entry := range hooks {
		if entry.pluginName != pluginName {
			filtered = append(filtered, entry)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("no hook found for plugin %s with type %s", pluginName, hookType)
	}

	h.hooks[hookType] = filtered

	if h.logger != nil {
		h.logger.Info(fmt.Sprintf("Unregistered %s hook for plugin %s", hookType, pluginName))
	}

	return nil
}

// UnregisterAll removes all hooks for a plugin
func (h *hookSystemImpl) UnregisterAll(pluginName string) error {
	if pluginName == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	removedCount := 0
	for hookType, hooks := range h.hooks {
		filtered := make([]hookEntry, 0, len(hooks))
		for _, entry := range hooks {
			if entry.pluginName != pluginName {
				filtered = append(filtered, entry)
			} else {
				removedCount++
			}
		}
		h.hooks[hookType] = filtered
	}

	if h.logger != nil {
		h.logger.Info(fmt.Sprintf("Unregistered %d hooks for plugin %s", removedCount, pluginName))
	}

	return nil
}

// ExecuteHooks executes all hooks of a given type
func (h *hookSystemImpl) ExecuteHooks(hookType HookType, ctx Context) error {
	if !isValidHookType(hookType) {
		return fmt.Errorf("invalid hook type: %s", hookType)
	}

	h.mu.RLock()
	hooks := h.hooks[hookType]
	// Create a copy to avoid holding the lock during execution
	hooksCopy := make([]hookEntry, len(hooks))
	copy(hooksCopy, hooks)
	h.mu.RUnlock()

	if len(hooksCopy) == 0 {
		return nil
	}

	// Create hook context
	hookCtx := newHookContext(hookType, ctx)

	// Execute hooks in priority order
	for _, entry := range hooksCopy {
		if hookCtx.IsSkipped() {
			break
		}

		// Set current plugin name
		hookCtx.setPluginName(entry.pluginName)

		// Execute hook with metrics tracking
		start := time.Now()
		err := h.executeHook(entry, hookCtx)
		duration := time.Since(start)

		// Update metrics
		h.updateMetrics(entry, duration, err)

		// Log errors but continue with remaining hooks
		if err != nil {
			if h.logger != nil {
				h.logger.Error(fmt.Sprintf("Hook %s for plugin %s failed: %v", hookType, entry.pluginName, err))
			}
			// Continue with remaining hooks (error isolation)
		}
	}

	return nil
}

// executeHook executes a single hook with error recovery
func (h *hookSystemImpl) executeHook(entry hookEntry, hookCtx *hookContextImpl) (err error) {
	// Recover from panics in hook handlers
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("hook panic: %v", r)
			if h.logger != nil {
				h.logger.Error(fmt.Sprintf("Hook for plugin %s panicked: %v", entry.pluginName, r))
			}
		}
	}()

	return entry.handler(hookCtx)
}

// updateMetrics updates execution metrics for a hook
func (h *hookSystemImpl) updateMetrics(entry hookEntry, duration time.Duration, err error) {
	entry.metrics.mu.Lock()
	defer entry.metrics.mu.Unlock()

	entry.metrics.executionCount++
	entry.metrics.totalDuration += duration

	if err != nil {
		entry.metrics.errorCount++
	}

	// Report to framework metrics collector if available
	if h.metrics != nil {
		h.metrics.RecordHistogram(
			fmt.Sprintf("plugin.hook.%s.duration", entry.pluginName),
			float64(duration.Milliseconds()),
			map[string]string{
				"plugin": entry.pluginName,
			},
		)

		if err != nil {
			h.metrics.IncrementCounter(
				fmt.Sprintf("plugin.hook.%s.errors", entry.pluginName),
				map[string]string{
					"plugin": entry.pluginName,
				},
			)
		}
	}
}

// ListHooks returns all registered hooks for a given type
func (h *hookSystemImpl) ListHooks(hookType HookType) []HookRegistration {
	if !isValidHookType(hookType) {
		return []HookRegistration{}
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	hooks, exists := h.hooks[hookType]
	if !exists {
		return []HookRegistration{}
	}

	registrations := make([]HookRegistration, len(hooks))
	for i, entry := range hooks {
		registrations[i] = HookRegistration{
			PluginName: entry.pluginName,
			HookType:   hookType,
			Priority:   entry.priority,
			Handler:    entry.handler,
		}
	}

	return registrations
}

// GetHookMetrics returns metrics for a specific plugin's hooks
func (h *hookSystemImpl) GetHookMetrics(pluginName string) map[string]HookMetrics {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make(map[string]HookMetrics)

	for hookType, hooks := range h.hooks {
		for _, entry := range hooks {
			if entry.pluginName == pluginName {
				entry.metrics.mu.Lock()
				avgDuration := time.Duration(0)
				if entry.metrics.executionCount > 0 {
					avgDuration = entry.metrics.totalDuration / time.Duration(entry.metrics.executionCount)
				}
				result[string(hookType)] = HookMetrics{
					ExecutionCount:  entry.metrics.executionCount,
					TotalDuration:   entry.metrics.totalDuration,
					AverageDuration: avgDuration,
					ErrorCount:      entry.metrics.errorCount,
				}
				entry.metrics.mu.Unlock()
			}
		}
	}

	return result
}

// isValidHookType checks if a hook type is valid
func isValidHookType(hookType HookType) bool {
	switch hookType {
	case HookTypeStartup, HookTypeShutdown, HookTypePreRequest,
		HookTypePostRequest, HookTypePreResponse, HookTypePostResponse, HookTypeError:
		return true
	default:
		return false
	}
}

// hookContextImpl implements HookContext
type hookContextImpl struct {
	hookType   HookType
	ctx        Context
	pluginName string
	data       map[string]interface{}
	skipped    bool
	mu         sync.RWMutex
}

// newHookContext creates a new hook context
func newHookContext(hookType HookType, ctx Context) *hookContextImpl {
	return &hookContextImpl{
		hookType: hookType,
		ctx:      ctx,
		data:     make(map[string]interface{}),
		skipped:  false,
	}
}

// Context returns the request context (nil for non-request hooks)
func (h *hookContextImpl) Context() Context {
	return h.ctx
}

// HookType returns the type of hook being executed
func (h *hookContextImpl) HookType() HookType {
	return h.hookType
}

// PluginName returns the name of the plugin executing the hook
func (h *hookContextImpl) PluginName() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.pluginName
}

// setPluginName sets the plugin name (internal use only)
func (h *hookContextImpl) setPluginName(name string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.pluginName = name
}

// Set stores a value in the hook context for passing data between hooks
func (h *hookContextImpl) Set(key string, value interface{}) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.data[key] = value
}

// Get retrieves a value from the hook context
func (h *hookContextImpl) Get(key string) interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.data[key]
}

// Skip marks the hook chain as skipped, preventing remaining hooks from executing
func (h *hookContextImpl) Skip() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.skipped = true
}

// IsSkipped returns whether the hook chain has been skipped
func (h *hookContextImpl) IsSkipped() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.skipped
}
