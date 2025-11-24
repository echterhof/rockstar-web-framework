package pkg

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestPipelineEngineRegister tests pipeline registration
func TestPipelineEngineRegister(t *testing.T) {
	engine := NewPipelineEngine()

	// Test successful registration
	config := PipelineConfig{
		Name:    "test-pipeline",
		Handler: func(ctx Context) (PipelineResult, error) { return PipelineResultContinue, nil },
		Enabled: true,
	}

	err := engine.Register(config)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test duplicate registration
	err = engine.Register(config)
	if err == nil {
		t.Error("Expected error for duplicate registration")
	}

	// Test registration with empty name
	invalidConfig := PipelineConfig{
		Name:    "",
		Handler: func(ctx Context) (PipelineResult, error) { return PipelineResultContinue, nil },
	}

	err = engine.Register(invalidConfig)
	if err == nil {
		t.Error("Expected error for empty name")
	}

	// Test registration with nil handler
	invalidConfig = PipelineConfig{
		Name:    "invalid",
		Handler: nil,
	}

	err = engine.Register(invalidConfig)
	if err == nil {
		t.Error("Expected error for nil handler")
	}
}

// TestPipelineEngineUnregister tests pipeline unregistration
func TestPipelineEngineUnregister(t *testing.T) {
	engine := NewPipelineEngine()

	config := PipelineConfig{
		Name:    "test-pipeline",
		Handler: func(ctx Context) (PipelineResult, error) { return PipelineResultContinue, nil },
		Enabled: true,
	}

	engine.Register(config)

	// Test successful unregistration
	err := engine.Unregister("test-pipeline")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test unregistering non-existent pipeline
	err = engine.Unregister("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent pipeline")
	}
}

// TestPipelineEngineEnableDisable tests enabling and disabling pipelines
func TestPipelineEngineEnableDisable(t *testing.T) {
	engine := NewPipelineEngine()

	config := PipelineConfig{
		Name:    "test-pipeline",
		Handler: func(ctx Context) (PipelineResult, error) { return PipelineResultContinue, nil },
		Enabled: true,
	}

	engine.Register(config)

	// Test disable
	err := engine.Disable("test-pipeline")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	retrieved, _ := engine.Get("test-pipeline")
	if retrieved.Enabled {
		t.Error("Expected pipeline to be disabled")
	}

	// Test enable
	err = engine.Enable("test-pipeline")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	retrieved, _ = engine.Get("test-pipeline")
	if !retrieved.Enabled {
		t.Error("Expected pipeline to be enabled")
	}

	// Test enable non-existent pipeline
	err = engine.Enable("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent pipeline")
	}
}

// TestPipelineEngineSetPriority tests setting pipeline priority
func TestPipelineEngineSetPriority(t *testing.T) {
	engine := NewPipelineEngine()

	config := PipelineConfig{
		Name:     "test-pipeline",
		Handler:  func(ctx Context) (PipelineResult, error) { return PipelineResultContinue, nil },
		Enabled:  true,
		Priority: 10,
	}

	engine.Register(config)

	// Test setting priority
	err := engine.SetPriority("test-pipeline", 20)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	retrieved, _ := engine.Get("test-pipeline")
	if retrieved.Priority != 20 {
		t.Errorf("Expected priority 20, got %d", retrieved.Priority)
	}

	// Test setting priority for non-existent pipeline
	err = engine.SetPriority("non-existent", 30)
	if err == nil {
		t.Error("Expected error for non-existent pipeline")
	}
}

// TestPipelineEngineExecute tests basic pipeline execution
func TestPipelineEngineExecute(t *testing.T) {
	engine := NewPipelineEngine()
	ctx := createPipelineTestContext()

	executed := false
	config := PipelineConfig{
		Name: "test-pipeline",
		Handler: func(ctx Context) (PipelineResult, error) {
			executed = true
			return PipelineResultContinue, nil
		},
		Enabled: true,
	}

	engine.Register(config)

	// Test successful execution
	result, err := engine.Execute(ctx, "test-pipeline")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result != PipelineResultContinue {
		t.Errorf("Expected PipelineResultContinue, got %v", result)
	}

	if !executed {
		t.Error("Expected pipeline to be executed")
	}

	// Test executing non-existent pipeline
	_, err = engine.Execute(ctx, "non-existent")
	if err == nil {
		t.Error("Expected error for non-existent pipeline")
	}

	// Test executing disabled pipeline
	engine.Disable("test-pipeline")
	_, err = engine.Execute(ctx, "test-pipeline")
	if err == nil {
		t.Error("Expected error for disabled pipeline")
	}
}

// TestPipelineEngineExecuteWithError tests pipeline execution with errors
func TestPipelineEngineExecuteWithError(t *testing.T) {
	engine := NewPipelineEngine()
	ctx := createPipelineTestContext()

	expectedErr := errors.New("pipeline error")
	config := PipelineConfig{
		Name: "error-pipeline",
		Handler: func(ctx Context) (PipelineResult, error) {
			return PipelineResultContinue, expectedErr
		},
		Enabled: true,
	}

	engine.Register(config)

	_, err := engine.Execute(ctx, "error-pipeline")
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

// TestPipelineEngineExecuteChain tests pipeline chaining
func TestPipelineEngineExecuteChain(t *testing.T) {
	engine := NewPipelineEngine()
	ctx := createPipelineTestContext()

	var executionOrder []string
	var mu sync.Mutex

	// Register multiple pipelines
	for i := 1; i <= 3; i++ {
		name := string(rune('A' + i - 1))
		config := PipelineConfig{
			Name: name,
			Handler: func(n string) PipelineFunc {
				return func(ctx Context) (PipelineResult, error) {
					mu.Lock()
					executionOrder = append(executionOrder, n)
					mu.Unlock()
					return PipelineResultContinue, nil
				}
			}(name),
			Enabled: true,
		}
		engine.Register(config)
	}

	// Execute chain
	result, err := engine.ExecuteChain(ctx, []string{"A", "B", "C"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result != PipelineResultContinue {
		t.Errorf("Expected PipelineResultContinue, got %v", result)
	}

	// Verify execution order
	if len(executionOrder) != 3 {
		t.Errorf("Expected 3 executions, got %d", len(executionOrder))
	}

	if executionOrder[0] != "A" || executionOrder[1] != "B" || executionOrder[2] != "C" {
		t.Errorf("Expected order [A, B, C], got %v", executionOrder)
	}
}

// TestPipelineEngineExecuteChainWithClose tests chain execution with close result
func TestPipelineEngineExecuteChainWithClose(t *testing.T) {
	engine := NewPipelineEngine()
	ctx := createPipelineTestContext()

	var executionOrder []string
	var mu sync.Mutex

	// Register pipelines where second one closes
	config1 := PipelineConfig{
		Name: "A",
		Handler: func(ctx Context) (PipelineResult, error) {
			mu.Lock()
			executionOrder = append(executionOrder, "A")
			mu.Unlock()
			return PipelineResultContinue, nil
		},
		Enabled: true,
	}

	config2 := PipelineConfig{
		Name: "B",
		Handler: func(ctx Context) (PipelineResult, error) {
			mu.Lock()
			executionOrder = append(executionOrder, "B")
			mu.Unlock()
			return PipelineResultClose, nil
		},
		Enabled: true,
	}

	config3 := PipelineConfig{
		Name: "C",
		Handler: func(ctx Context) (PipelineResult, error) {
			mu.Lock()
			executionOrder = append(executionOrder, "C")
			mu.Unlock()
			return PipelineResultContinue, nil
		},
		Enabled: true,
	}

	engine.Register(config1)
	engine.Register(config2)
	engine.Register(config3)

	// Execute chain
	result, err := engine.ExecuteChain(ctx, []string{"A", "B", "C"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result != PipelineResultClose {
		t.Errorf("Expected PipelineResultClose, got %v", result)
	}

	// Verify C was not executed
	if len(executionOrder) != 2 {
		t.Errorf("Expected 2 executions, got %d", len(executionOrder))
	}

	if executionOrder[0] != "A" || executionOrder[1] != "B" {
		t.Errorf("Expected order [A, B], got %v", executionOrder)
	}
}

// TestPipelineEngineExecuteAsync tests asynchronous pipeline execution
func TestPipelineEngineExecuteAsync(t *testing.T) {
	engine := NewPipelineEngine()
	ctx := createPipelineTestContext()

	var executed int32
	config := PipelineConfig{
		Name: "async-pipeline",
		Handler: func(ctx Context) (PipelineResult, error) {
			time.Sleep(50 * time.Millisecond)
			atomic.AddInt32(&executed, 1)
			return PipelineResultContinue, nil
		},
		Enabled: true,
		Async:   true,
	}

	engine.Register(config)

	// Execute asynchronously
	err := engine.ExecuteAsync(ctx, "async-pipeline")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Should not be executed yet
	if atomic.LoadInt32(&executed) != 0 {
		t.Error("Expected pipeline not to be executed immediately")
	}

	// Wait for completion
	err = engine.WaitAll()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Should be executed now
	if atomic.LoadInt32(&executed) != 1 {
		t.Errorf("Expected pipeline to be executed once, got %d", executed)
	}
}

// TestPipelineEngineExecuteMultiplex tests concurrent pipeline execution
func TestPipelineEngineExecuteMultiplex(t *testing.T) {
	engine := NewPipelineEngine()
	ctx := createPipelineTestContext()

	var executionCount int32

	// Register multiple pipelines
	for i := 1; i <= 5; i++ {
		name := string(rune('A' + i - 1))
		config := PipelineConfig{
			Name: name,
			Handler: func(ctx Context) (PipelineResult, error) {
				time.Sleep(10 * time.Millisecond)
				atomic.AddInt32(&executionCount, 1)
				return PipelineResultContinue, nil
			},
			Enabled: true,
		}
		engine.Register(config)
	}

	// Execute multiplexed
	start := time.Now()
	results, errors := engine.ExecuteMultiplex(ctx, []string{"A", "B", "C", "D", "E"})
	duration := time.Since(start)

	// Verify all executed
	if atomic.LoadInt32(&executionCount) != 5 {
		t.Errorf("Expected 5 executions, got %d", executionCount)
	}

	// Verify results
	if len(results) != 5 {
		t.Errorf("Expected 5 results, got %d", len(results))
	}

	for i, result := range results {
		if result != PipelineResultContinue {
			t.Errorf("Expected PipelineResultContinue for result %d, got %v", i, result)
		}
	}

	// Verify errors
	if len(errors) != 5 {
		t.Errorf("Expected 5 error slots, got %d", len(errors))
	}

	for i, err := range errors {
		if err != nil {
			t.Errorf("Expected no error for result %d, got %v", i, err)
		}
	}

	// Verify concurrent execution (should be faster than sequential)
	// Sequential would take 50ms, concurrent should be around 10-20ms
	if duration > 40*time.Millisecond {
		t.Errorf("Expected concurrent execution to be faster, took %v", duration)
	}
}

// TestPipelineEngineExecuteMultiplexWithErrors tests multiplexed execution with errors
func TestPipelineEngineExecuteMultiplexWithErrors(t *testing.T) {
	engine := NewPipelineEngine()
	ctx := createPipelineTestContext()

	expectedErr := errors.New("pipeline error")

	// Register pipelines where some fail
	config1 := PipelineConfig{
		Name: "success",
		Handler: func(ctx Context) (PipelineResult, error) {
			return PipelineResultContinue, nil
		},
		Enabled: true,
	}

	config2 := PipelineConfig{
		Name: "failure",
		Handler: func(ctx Context) (PipelineResult, error) {
			return PipelineResultContinue, expectedErr
		},
		Enabled: true,
	}

	engine.Register(config1)
	engine.Register(config2)

	// Execute multiplexed
	results, errors := engine.ExecuteMultiplex(ctx, []string{"success", "failure"})

	// Verify results
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// Verify errors
	if len(errors) != 2 {
		t.Errorf("Expected 2 error slots, got %d", len(errors))
	}

	if errors[0] != nil {
		t.Errorf("Expected no error for success pipeline, got %v", errors[0])
	}

	if errors[1] != expectedErr {
		t.Errorf("Expected error %v for failure pipeline, got %v", expectedErr, errors[1])
	}
}

// TestPipelineEngineChainResult tests pipeline chaining via result
func TestPipelineEngineChainResult(t *testing.T) {
	engine := NewPipelineEngine()
	ctx := createPipelineTestContext()

	var executionOrder []string
	var mu sync.Mutex

	// Register first pipeline that chains to second
	config1 := PipelineConfig{
		Name: "first",
		Handler: func(ctx Context) (PipelineResult, error) {
			mu.Lock()
			executionOrder = append(executionOrder, "first")
			mu.Unlock()
			return PipelineResultChain, nil
		},
		NextPipeline: "second",
		Enabled:      true,
	}

	config2 := PipelineConfig{
		Name: "second",
		Handler: func(ctx Context) (PipelineResult, error) {
			mu.Lock()
			executionOrder = append(executionOrder, "second")
			mu.Unlock()
			return PipelineResultContinue, nil
		},
		Enabled: true,
	}

	engine.Register(config1)
	engine.Register(config2)

	// Execute first pipeline
	result, err := engine.Execute(ctx, "first")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result != PipelineResultContinue {
		t.Errorf("Expected PipelineResultContinue, got %v", result)
	}

	// Verify both executed
	if len(executionOrder) != 2 {
		t.Errorf("Expected 2 executions, got %d", len(executionOrder))
	}

	if executionOrder[0] != "first" || executionOrder[1] != "second" {
		t.Errorf("Expected order [first, second], got %v", executionOrder)
	}
}

// TestPipelineEngineViewResult tests pipeline view execution
func TestPipelineEngineViewResult(t *testing.T) {
	engine := NewPipelineEngine()
	ctx := createPipelineTestContext()

	viewExecuted := false
	pipelineExecuted := false

	config := PipelineConfig{
		Name: "with-view",
		Handler: func(ctx Context) (PipelineResult, error) {
			pipelineExecuted = true
			return PipelineResultView, nil
		},
		ViewHandler: func(ctx Context) error {
			viewExecuted = true
			return nil
		},
		Enabled: true,
	}

	engine.Register(config)

	// Execute pipeline
	result, err := engine.Execute(ctx, "with-view")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result != PipelineResultContinue {
		t.Errorf("Expected PipelineResultContinue, got %v", result)
	}

	// Verify both executed
	if !pipelineExecuted {
		t.Error("Expected pipeline to be executed")
	}

	if !viewExecuted {
		t.Error("Expected view to be executed")
	}
}

// TestPipelineEngineList tests listing pipelines
func TestPipelineEngineList(t *testing.T) {
	engine := NewPipelineEngine()

	// Register multiple pipelines
	for i := 1; i <= 3; i++ {
		name := string(rune('A' + i - 1))
		config := PipelineConfig{
			Name:    name,
			Handler: func(ctx Context) (PipelineResult, error) { return PipelineResultContinue, nil },
			Enabled: true,
		}
		engine.Register(config)
	}

	// List pipelines
	configs := engine.List()

	if len(configs) != 3 {
		t.Errorf("Expected 3 pipelines, got %d", len(configs))
	}
}

// TestPipelineEngineClear tests clearing all pipelines
func TestPipelineEngineClear(t *testing.T) {
	engine := NewPipelineEngine()

	// Register pipelines
	config := PipelineConfig{
		Name:    "test",
		Handler: func(ctx Context) (PipelineResult, error) { return PipelineResultContinue, nil },
		Enabled: true,
	}
	engine.Register(config)

	// Clear
	engine.Clear()

	// Verify cleared
	configs := engine.List()
	if len(configs) != 0 {
		t.Errorf("Expected 0 pipelines after clear, got %d", len(configs))
	}
}

// TestPipelineBuilder tests the pipeline builder
func TestPipelineBuilder(t *testing.T) {
	handler := func(ctx Context) (PipelineResult, error) { return PipelineResultContinue, nil }
	viewHandler := func(ctx Context) error { return nil }

	config := NewPipelineBuilder("test").
		WithHandler(handler).
		WithPriority(10).
		WithNextPipeline("next").
		WithViewHandler(viewHandler).
		WithAsync(true).
		WithTimeout(5000).
		Build()

	if config.Name != "test" {
		t.Errorf("Expected name 'test', got '%s'", config.Name)
	}

	if config.Priority != 10 {
		t.Errorf("Expected priority 10, got %d", config.Priority)
	}

	if config.NextPipeline != "next" {
		t.Errorf("Expected next pipeline 'next', got '%s'", config.NextPipeline)
	}

	if !config.Async {
		t.Error("Expected async to be true")
	}

	if config.Timeout != 5000 {
		t.Errorf("Expected timeout 5000, got %d", config.Timeout)
	}

	if config.Handler == nil {
		t.Error("Expected handler to be set")
	}

	if config.ViewHandler == nil {
		t.Error("Expected view handler to be set")
	}
}

// TestPipelineEngineTimeout tests pipeline execution with timeout
func TestPipelineEngineTimeout(t *testing.T) {
	engine := NewPipelineEngine()
	ctx := createPipelineTestContext()

	config := PipelineConfig{
		Name: "slow-pipeline",
		Handler: func(ctx Context) (PipelineResult, error) {
			time.Sleep(200 * time.Millisecond)
			return PipelineResultContinue, nil
		},
		Enabled: true,
		Timeout: 50, // 50ms timeout
	}

	engine.Register(config)

	// Execute with timeout
	_, err := engine.Execute(ctx, "slow-pipeline")
	if err == nil {
		t.Error("Expected timeout error")
	}
}

// Helper function to create a test context for pipeline tests
func createPipelineTestContext() Context {
	req := &Request{
		Method: "GET",
		URL:    &url.URL{Path: "/test"},
		Params: make(map[string]string),
		Query:  make(map[string]string),
		Header: make(http.Header),
	}

	resp := &mockPipelineResponseWriter{
		headers: make(map[string]string),
	}

	return NewContext(req, resp, context.Background())
}

// Mock response writer for pipeline testing
type mockPipelineResponseWriter struct {
	statusCode int
	headers    map[string]string
	body       []byte
}

func (m *mockPipelineResponseWriter) WriteHeader(statusCode int) {
	m.statusCode = statusCode
}

func (m *mockPipelineResponseWriter) Write(data []byte) (int, error) {
	m.body = append(m.body, data...)
	return len(data), nil
}

func (m *mockPipelineResponseWriter) SetHeader(key, value string) {
	if m.headers == nil {
		m.headers = make(map[string]string)
	}
	m.headers[key] = value
}

func (m *mockPipelineResponseWriter) GetHeader(key string) string {
	if m.headers == nil {
		return ""
	}
	return m.headers[key]
}

func (m *mockPipelineResponseWriter) WriteJSON(statusCode int, data interface{}) error {
	m.statusCode = statusCode
	return nil
}

func (m *mockPipelineResponseWriter) WriteXML(statusCode int, data interface{}) error {
	m.statusCode = statusCode
	return nil
}

func (m *mockPipelineResponseWriter) WriteHTML(statusCode int, template string, data interface{}) error {
	m.statusCode = statusCode
	return nil
}

func (m *mockPipelineResponseWriter) WriteString(statusCode int, message string) error {
	m.statusCode = statusCode
	m.body = []byte(message)
	return nil
}

func (m *mockPipelineResponseWriter) SetCookie(cookie *Cookie) error {
	return nil
}

func (m *mockPipelineResponseWriter) Flush() error {
	return nil
}

func (m *mockPipelineResponseWriter) Close() error {
	return nil
}

func (m *mockPipelineResponseWriter) Header() http.Header {
	h := make(http.Header)
	for k, v := range m.headers {
		h.Set(k, v)
	}
	return h
}

func (m *mockPipelineResponseWriter) Status() int {
	return m.statusCode
}

func (m *mockPipelineResponseWriter) Size() int64 {
	return int64(len(m.body))
}

func (m *mockPipelineResponseWriter) Written() bool {
	return len(m.body) > 0
}

func (m *mockPipelineResponseWriter) SetContentType(contentType string) {
	m.SetHeader("Content-Type", contentType)
}

func (m *mockPipelineResponseWriter) WriteStream(statusCode int, contentType string, reader io.Reader) error {
	m.statusCode = statusCode
	return nil
}

func (m *mockPipelineResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, errors.New("hijack not supported")
}

func (m *mockPipelineResponseWriter) Push(target string, opts *http.PushOptions) error {
	return errors.New("push not supported")
}

func (m *mockPipelineResponseWriter) SetTemplateManager(tm TemplateManager) {
	// No-op for mock
}
