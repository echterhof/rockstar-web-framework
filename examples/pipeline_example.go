package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
	// Create a pipeline engine
	engine := pkg.NewPipelineEngine()

	// Example 1: Basic pipeline that processes data
	basicPipeline := pkg.PipelineConfig{
		Name: "data-processor",
		Handler: func(ctx pkg.Context) (pkg.PipelineResult, error) {
			fmt.Println("Processing data in pipeline...")

			// Access request data
			if ctx.Request() != nil {
				fmt.Printf("Request Method: %s\n", ctx.Request().Method)
			}

			// Access database (if available)
			if ctx.DB() != nil {
				fmt.Println("Database access available")
			}

			// Access session (if available)
			if ctx.Session() != nil {
				fmt.Println("Session access available")
			}

			// Continue to next step
			return pkg.PipelineResultContinue, nil
		},
		Enabled: true,
	}

	engine.Register(basicPipeline)

	// Example 2: Pipeline that chains to another pipeline
	validationPipeline := pkg.PipelineConfig{
		Name: "validator",
		Handler: func(ctx pkg.Context) (pkg.PipelineResult, error) {
			fmt.Println("Validating request...")

			// Perform validation
			if ctx.Request() == nil {
				return pkg.PipelineResultClose, fmt.Errorf("invalid request")
			}

			// Chain to data processor
			return pkg.PipelineResultChain, nil
		},
		NextPipeline: "data-processor",
		Enabled:      true,
	}

	engine.Register(validationPipeline)

	// Example 3: Pipeline that executes a view
	renderPipeline := pkg.PipelineConfig{
		Name: "renderer",
		Handler: func(ctx pkg.Context) (pkg.PipelineResult, error) {
			fmt.Println("Preparing data for rendering...")

			// Process data and prepare for view
			return pkg.PipelineResultView, nil
		},
		ViewHandler: func(ctx pkg.Context) error {
			fmt.Println("Rendering view...")
			return ctx.String(200, "Hello from pipeline view!")
		},
		Enabled: true,
	}

	engine.Register(renderPipeline)

	// Example 4: Pipeline with timeout
	slowPipeline := pkg.PipelineConfig{
		Name: "slow-processor",
		Handler: func(ctx pkg.Context) (pkg.PipelineResult, error) {
			fmt.Println("Starting slow processing...")
			time.Sleep(100 * time.Millisecond)
			fmt.Println("Slow processing complete")
			return pkg.PipelineResultContinue, nil
		},
		Enabled: true,
		Timeout: 5000, // 5 second timeout
	}

	engine.Register(slowPipeline)

	// Example 5: Async pipeline for background processing
	backgroundPipeline := pkg.PipelineConfig{
		Name: "background-task",
		Handler: func(ctx pkg.Context) (pkg.PipelineResult, error) {
			fmt.Println("Running background task...")
			time.Sleep(50 * time.Millisecond)
			fmt.Println("Background task complete")
			return pkg.PipelineResultContinue, nil
		},
		Enabled: true,
		Async:   true,
	}

	engine.Register(backgroundPipeline)

	// Example 6: Pipeline for form data checking
	formValidationPipeline := pkg.PipelineConfig{
		Name: "form-validator",
		Handler: func(ctx pkg.Context) (pkg.PipelineResult, error) {
			fmt.Println("Validating form data...")

			// Check form values
			username := ctx.FormValue("username")
			if username == "" {
				return pkg.PipelineResultClose, fmt.Errorf("username is required")
			}

			// Check uploaded files
			file, err := ctx.FormFile("avatar")
			if err == nil && file != nil {
				fmt.Printf("File uploaded: %s\n", file.Filename)
			}

			return pkg.PipelineResultContinue, nil
		},
		Enabled: true,
	}

	engine.Register(formValidationPipeline)

	// Example 7: Pipeline for API rate limiting
	rateLimitPipeline := pkg.PipelineConfig{
		Name: "rate-limiter",
		Handler: func(ctx pkg.Context) (pkg.PipelineResult, error) {
			fmt.Println("Checking rate limits...")

			// Access cache for rate limiting
			if ctx.Cache() != nil {
				// Check rate limit from cache
				fmt.Println("Rate limit check passed")
			}

			return pkg.PipelineResultContinue, nil
		},
		Enabled:  true,
		Priority: 100, // High priority - execute first
	}

	engine.Register(rateLimitPipeline)

	// Example 8: Pipeline for logging
	loggingPipeline := pkg.PipelineConfig{
		Name: "logger",
		Handler: func(ctx pkg.Context) (pkg.PipelineResult, error) {
			fmt.Println("Logging request...")

			if ctx.Logger() != nil {
				ctx.Logger().Info("Request processed")
			}

			return pkg.PipelineResultContinue, nil
		},
		Enabled:  true,
		Priority: 1, // Low priority - execute last
	}

	engine.Register(loggingPipeline)

	// Example 9: Using the pipeline builder
	builderPipeline := pkg.NewPipelineBuilder("builder-example").
		WithHandler(func(ctx pkg.Context) (pkg.PipelineResult, error) {
			fmt.Println("Pipeline created with builder")
			return pkg.PipelineResultContinue, nil
		}).
		WithPriority(50).
		WithTimeout(3000).
		Build()

	engine.Register(builderPipeline)

	// Demonstrate pipeline execution
	fmt.Println("\n=== Demonstrating Pipeline Execution ===\n")

	// Create a mock context for demonstration
	ctx := createDemoContext()

	// Execute single pipeline
	fmt.Println("1. Executing single pipeline:")
	result, err := engine.Execute(ctx, "data-processor")
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Result: %v\n\n", result)
	}

	// Execute chain of pipelines
	fmt.Println("2. Executing pipeline chain:")
	result, err = engine.ExecuteChain(ctx, []string{"rate-limiter", "form-validator", "data-processor"})
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Result: %v\n\n", result)
	}

	// Execute async pipeline
	fmt.Println("3. Executing async pipeline:")
	err = engine.ExecuteAsync(ctx, "background-task")
	if err != nil {
		log.Printf("Error: %v\n", err)
	}
	fmt.Println("Async pipeline started (running in background)\n")

	// Execute multiple pipelines concurrently
	fmt.Println("4. Executing multiplexed pipelines:")
	results, errors := engine.ExecuteMultiplex(ctx, []string{"data-processor", "slow-processor", "logger"})
	for i, result := range results {
		if errors[i] != nil {
			log.Printf("Pipeline %d error: %v\n", i, errors[i])
		} else {
			fmt.Printf("Pipeline %d result: %v\n", i, result)
		}
	}
	fmt.Println()

	// Execute pipeline with chaining
	fmt.Println("5. Executing pipeline with automatic chaining:")
	result, err = engine.Execute(ctx, "validator")
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Result: %v\n\n", result)
	}

	// Execute pipeline with view
	fmt.Println("6. Executing pipeline with view:")
	result, err = engine.Execute(ctx, "renderer")
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Result: %v\n\n", result)
	}

	// Wait for all async pipelines to complete
	fmt.Println("7. Waiting for async pipelines to complete:")
	err = engine.WaitAll()
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Println("All async pipelines completed\n")
	}

	// List all registered pipelines
	fmt.Println("8. Listing all registered pipelines:")
	pipelines := engine.List()
	for _, p := range pipelines {
		fmt.Printf("  - %s (Priority: %d, Enabled: %v)\n", p.Name, p.Priority, p.Enabled)
	}

	fmt.Println("\n=== Pipeline Examples Complete ===")
}

// createDemoContext creates a mock context for demonstration
func createDemoContext() pkg.Context {
	// In a real application, this would be created from an actual HTTP request
	// For demonstration, we create a minimal context
	req := &pkg.Request{
		Method: "POST",
		Params: make(map[string]string),
		Query:  make(map[string]string),
		Form: map[string]string{
			"username": "demo-user",
		},
	}

	resp := &demoResponseWriter{}

	return pkg.NewContext(req, resp, nil)
}

// demoResponseWriter is a minimal response writer for demonstration
type demoResponseWriter struct {
	statusCode int
	body       []byte
	headers    map[string][]string
}

func (d *demoResponseWriter) Header() http.Header {
	if d.headers == nil {
		d.headers = make(map[string][]string)
	}
	return http.Header(d.headers)
}

func (d *demoResponseWriter) WriteHeader(statusCode int) {
	d.statusCode = statusCode
}

func (d *demoResponseWriter) Write(data []byte) (int, error) {
	d.body = append(d.body, data...)
	return len(data), nil
}

func (d *demoResponseWriter) SetHeader(key, value string) {
	if d.headers == nil {
		d.headers = make(map[string][]string)
	}
	d.headers[key] = []string{value}
}

func (d *demoResponseWriter) GetHeader(key string) string {
	if d.headers == nil {
		return ""
	}
	if vals, ok := d.headers[key]; ok && len(vals) > 0 {
		return vals[0]
	}
	return ""
}

func (d *demoResponseWriter) WriteJSON(statusCode int, data interface{}) error {
	d.statusCode = statusCode
	return nil
}

func (d *demoResponseWriter) WriteXML(statusCode int, data interface{}) error {
	d.statusCode = statusCode
	return nil
}

func (d *demoResponseWriter) WriteHTML(statusCode int, template string, data interface{}) error {
	d.statusCode = statusCode
	return nil
}

func (d *demoResponseWriter) WriteString(statusCode int, message string) error {
	d.statusCode = statusCode
	d.body = []byte(message)
	fmt.Printf("Response: %s\n", message)
	return nil
}

func (d *demoResponseWriter) SetCookie(cookie *pkg.Cookie) error {
	return nil
}

func (d *demoResponseWriter) Flush() error {
	return nil
}

func (d *demoResponseWriter) Close() error {
	return nil
}

func (d *demoResponseWriter) SetContentType(contentType string) {
	d.SetHeader("Content-Type", contentType)
}

func (d *demoResponseWriter) SetTemplateManager(tm pkg.TemplateManager) {
	// Not needed for demo
}

func (d *demoResponseWriter) Size() int64 {
	return int64(len(d.body))
}

func (d *demoResponseWriter) Status() int {
	return d.statusCode
}

func (d *demoResponseWriter) Written() bool {
	return len(d.body) > 0
}

func (d *demoResponseWriter) WriteStream(statusCode int, contentType string, reader io.Reader) error {
	d.statusCode = statusCode
	return nil
}
