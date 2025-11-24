package pkg

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// PipelineResult represents the result of a pipeline execution
type PipelineResult int

const (
	// PipelineResultContinue continues to the next pipeline or view
	PipelineResultContinue PipelineResult = iota
	// PipelineResultClose closes the connection
	PipelineResultClose
	// PipelineResultChain chains to another pipeline
	PipelineResultChain
	// PipelineResultView executes a view
	PipelineResultView
)

// PipelineFunc is a function that processes data in a pipeline
// It returns a PipelineResult indicating what should happen next
type PipelineFunc func(ctx Context) (PipelineResult, error)

// PipelineConfig holds configuration for a pipeline
type PipelineConfig struct {
	// Name is the pipeline identifier
	Name string

	// Handler is the pipeline function
	Handler PipelineFunc

	// Priority determines execution order (higher priority executes first)
	Priority int

	// Enabled determines if the pipeline is active
	Enabled bool

	// NextPipeline specifies the next pipeline to chain to (if result is PipelineResultChain)
	NextPipeline string

	// ViewHandler specifies the view to execute (if result is PipelineResultView)
	ViewHandler HandlerFunc

	// Async determines if the pipeline should run asynchronously
	Async bool

	// Timeout specifies the maximum execution time for the pipeline
	Timeout int64 // milliseconds
}

// PipelineEngine manages pipeline execution with support for chaining and multiplexing
type PipelineEngine interface {
	// Register adds a pipeline with configuration
	Register(config PipelineConfig) error

	// Unregister removes a pipeline by name
	Unregister(name string) error

	// Enable enables a pipeline by name
	Enable(name string) error

	// Disable disables a pipeline by name
	Disable(name string) error

	// SetPriority changes the priority of a pipeline
	SetPriority(name string, priority int) error

	// Execute runs a specific pipeline by name
	Execute(ctx Context, name string) (PipelineResult, error)

	// ExecuteChain runs a chain of pipelines
	ExecuteChain(ctx Context, names []string) (PipelineResult, error)

	// ExecuteAsync runs a pipeline asynchronously using goroutines
	ExecuteAsync(ctx Context, name string) error

	// ExecuteMultiplex runs multiple pipelines concurrently
	ExecuteMultiplex(ctx Context, names []string) ([]PipelineResult, []error)

	// List returns all registered pipeline configurations
	List() []PipelineConfig

	// Get returns a specific pipeline configuration
	Get(name string) (*PipelineConfig, error)

	// Clear removes all pipelines
	Clear()

	// WaitAll waits for all async pipelines to complete
	WaitAll() error
}

// pipelineEngine implements PipelineEngine
type pipelineEngine struct {
	pipelines map[string]*PipelineConfig
	mu        sync.RWMutex
	wg        sync.WaitGroup
	asyncErrs []error
	asyncMu   sync.Mutex
}

// NewPipelineEngine creates a new pipeline engine
func NewPipelineEngine() PipelineEngine {
	return &pipelineEngine{
		pipelines: make(map[string]*PipelineConfig),
		asyncErrs: make([]error, 0),
	}
}

// Register adds a pipeline with configuration
func (p *pipelineEngine) Register(config PipelineConfig) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if config.Name == "" {
		return fmt.Errorf("pipeline name cannot be empty")
	}

	if config.Handler == nil {
		return fmt.Errorf("pipeline handler cannot be nil")
	}

	// Check if pipeline already exists
	if _, exists := p.pipelines[config.Name]; exists {
		return fmt.Errorf("pipeline %s already registered", config.Name)
	}

	// Store a copy of the config
	configCopy := config
	p.pipelines[config.Name] = &configCopy

	return nil
}

// Unregister removes a pipeline by name
func (p *pipelineEngine) Unregister(name string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.pipelines[name]; !exists {
		return fmt.Errorf("pipeline %s not found", name)
	}

	delete(p.pipelines, name)
	return nil
}

// Enable enables a pipeline by name
func (p *pipelineEngine) Enable(name string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	pl, exists := p.pipelines[name]
	if !exists {
		return fmt.Errorf("pipeline %s not found", name)
	}

	pl.Enabled = true
	return nil
}

// Disable disables a pipeline by name
func (p *pipelineEngine) Disable(name string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	pl, exists := p.pipelines[name]
	if !exists {
		return fmt.Errorf("pipeline %s not found", name)
	}

	pl.Enabled = false
	return nil
}

// SetPriority changes the priority of a pipeline
func (p *pipelineEngine) SetPriority(name string, priority int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	pl, exists := p.pipelines[name]
	if !exists {
		return fmt.Errorf("pipeline %s not found", name)
	}

	pl.Priority = priority
	return nil
}

// Execute runs a specific pipeline by name
func (p *pipelineEngine) Execute(ctx Context, name string) (PipelineResult, error) {
	p.mu.RLock()
	pl, exists := p.pipelines[name]
	p.mu.RUnlock()

	if !exists {
		return PipelineResultContinue, fmt.Errorf("pipeline %s not found", name)
	}

	if !pl.Enabled {
		return PipelineResultContinue, fmt.Errorf("pipeline %s is disabled", name)
	}

	// Execute with timeout if specified
	if pl.Timeout > 0 {
		return p.executeWithTimeout(ctx, pl)
	}

	// Execute the pipeline
	result, err := pl.Handler(ctx)
	if err != nil {
		return result, err
	}

	// Handle the result
	return p.handlePipelineResult(ctx, pl, result)
}

// executeWithTimeout executes a pipeline with a timeout
func (p *pipelineEngine) executeWithTimeout(ctx Context, pl *PipelineConfig) (PipelineResult, error) {
	// Create a context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx.Context(), timeFromMillis(pl.Timeout))
	defer cancel()

	// Create a channel to receive the result
	resultChan := make(chan struct {
		result PipelineResult
		err    error
	}, 1)

	// Execute in a goroutine
	go func() {
		result, err := pl.Handler(ctx)
		resultChan <- struct {
			result PipelineResult
			err    error
		}{result, err}
	}()

	// Wait for result or timeout
	select {
	case res := <-resultChan:
		if res.err != nil {
			return res.result, res.err
		}
		return p.handlePipelineResult(ctx, pl, res.result)
	case <-timeoutCtx.Done():
		return PipelineResultContinue, fmt.Errorf("pipeline %s timed out after %dms", pl.Name, pl.Timeout)
	}
}

// handlePipelineResult handles the result of a pipeline execution
func (p *pipelineEngine) handlePipelineResult(ctx Context, pl *PipelineConfig, result PipelineResult) (PipelineResult, error) {
	switch result {
	case PipelineResultChain:
		// Chain to the next pipeline if specified
		if pl.NextPipeline != "" {
			return p.Execute(ctx, pl.NextPipeline)
		}
		return PipelineResultContinue, nil

	case PipelineResultView:
		// Execute the view handler if specified
		if pl.ViewHandler != nil {
			err := pl.ViewHandler(ctx)
			return PipelineResultContinue, err
		}
		return PipelineResultContinue, nil

	case PipelineResultClose:
		// Close the connection
		return PipelineResultClose, nil

	default:
		// Continue to the next step
		return PipelineResultContinue, nil
	}
}

// ExecuteChain runs a chain of pipelines
func (p *pipelineEngine) ExecuteChain(ctx Context, names []string) (PipelineResult, error) {
	for _, name := range names {
		result, err := p.Execute(ctx, name)
		if err != nil {
			return result, err
		}

		// If the pipeline wants to close or has a specific result, stop the chain
		if result == PipelineResultClose {
			return result, nil
		}
	}

	return PipelineResultContinue, nil
}

// ExecuteAsync runs a pipeline asynchronously using goroutines
func (p *pipelineEngine) ExecuteAsync(ctx Context, name string) error {
	p.mu.RLock()
	pl, exists := p.pipelines[name]
	p.mu.RUnlock()

	if !exists {
		return fmt.Errorf("pipeline %s not found", name)
	}

	if !pl.Enabled {
		return fmt.Errorf("pipeline %s is disabled", name)
	}

	// Increment wait group
	p.wg.Add(1)

	// Execute in a goroutine
	go func() {
		defer p.wg.Done()

		_, err := p.Execute(ctx, name)
		if err != nil {
			p.asyncMu.Lock()
			p.asyncErrs = append(p.asyncErrs, fmt.Errorf("async pipeline %s failed: %w", name, err))
			p.asyncMu.Unlock()
		}
	}()

	return nil
}

// ExecuteMultiplex runs multiple pipelines concurrently
func (p *pipelineEngine) ExecuteMultiplex(ctx Context, names []string) ([]PipelineResult, []error) {
	results := make([]PipelineResult, len(names))
	errors := make([]error, len(names))

	var wg sync.WaitGroup
	wg.Add(len(names))

	for i, name := range names {
		go func(index int, pipelineName string) {
			defer wg.Done()

			result, err := p.Execute(ctx, pipelineName)
			results[index] = result
			errors[index] = err
		}(i, name)
	}

	wg.Wait()

	return results, errors
}

// List returns all registered pipeline configurations
func (p *pipelineEngine) List() []PipelineConfig {
	p.mu.RLock()
	defer p.mu.RUnlock()

	configs := make([]PipelineConfig, 0, len(p.pipelines))
	for _, pl := range p.pipelines {
		configs = append(configs, *pl)
	}

	return configs
}

// Get returns a specific pipeline configuration
func (p *pipelineEngine) Get(name string) (*PipelineConfig, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	pl, exists := p.pipelines[name]
	if !exists {
		return nil, fmt.Errorf("pipeline %s not found", name)
	}

	// Return a copy
	configCopy := *pl
	return &configCopy, nil
}

// Clear removes all pipelines
func (p *pipelineEngine) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.pipelines = make(map[string]*PipelineConfig)
}

// WaitAll waits for all async pipelines to complete
func (p *pipelineEngine) WaitAll() error {
	p.wg.Wait()

	p.asyncMu.Lock()
	defer p.asyncMu.Unlock()

	if len(p.asyncErrs) > 0 {
		// Return the first error
		return p.asyncErrs[0]
	}

	return nil
}

// Helper function to convert milliseconds to time.Duration
func timeFromMillis(ms int64) time.Duration {
	return time.Duration(ms) * time.Millisecond
}

// PipelineBuilder provides a fluent interface for building pipelines
type PipelineBuilder struct {
	config PipelineConfig
}

// NewPipelineBuilder creates a new pipeline builder
func NewPipelineBuilder(name string) *PipelineBuilder {
	return &PipelineBuilder{
		config: PipelineConfig{
			Name:    name,
			Enabled: true,
		},
	}
}

// WithHandler sets the pipeline handler
func (b *PipelineBuilder) WithHandler(handler PipelineFunc) *PipelineBuilder {
	b.config.Handler = handler
	return b
}

// WithPriority sets the pipeline priority
func (b *PipelineBuilder) WithPriority(priority int) *PipelineBuilder {
	b.config.Priority = priority
	return b
}

// WithNextPipeline sets the next pipeline to chain to
func (b *PipelineBuilder) WithNextPipeline(name string) *PipelineBuilder {
	b.config.NextPipeline = name
	return b
}

// WithViewHandler sets the view handler
func (b *PipelineBuilder) WithViewHandler(handler HandlerFunc) *PipelineBuilder {
	b.config.ViewHandler = handler
	return b
}

// WithAsync sets the pipeline to run asynchronously
func (b *PipelineBuilder) WithAsync(async bool) *PipelineBuilder {
	b.config.Async = async
	return b
}

// WithTimeout sets the pipeline timeout in milliseconds
func (b *PipelineBuilder) WithTimeout(timeoutMs int64) *PipelineBuilder {
	b.config.Timeout = timeoutMs
	return b
}

// Build returns the pipeline configuration
func (b *PipelineBuilder) Build() PipelineConfig {
	return b.config
}
