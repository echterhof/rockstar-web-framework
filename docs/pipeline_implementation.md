# Pipeline System Implementation

## Overview

The Rockstar Web Framework's pipeline system provides a flexible and powerful way to process requests through a series of configurable stages. Pipelines allow you to handle data processing, validation, transformation, and other operations without necessarily requiring a view to be executed.

## Key Features

- **Flexible Execution Flow**: Pipelines can continue to the next step, close the connection, chain to another pipeline, or execute a view
- **Asynchronous Execution**: Support for running pipelines in the background using goroutines
- **Pipeline Multiplexing**: Execute multiple pipelines concurrently for parallel processing
- **Configurable Ordering**: Control execution order through priority settings
- **Timeout Support**: Set maximum execution time for pipelines
- **Chaining**: Automatically chain pipelines together based on results
- **Full Framework Integration**: Access to all framework features including database, cache, sessions, files, and more

## Architecture

### Pipeline Results

Pipelines return a `PipelineResult` that determines what happens next:

```go
const (
    PipelineResultContinue  // Continue to the next pipeline or view
    PipelineResultClose     // Close the connection
    PipelineResultChain     // Chain to another pipeline
    PipelineResultView      // Execute a view
)
```

### Pipeline Configuration

Each pipeline is configured with:

- **Name**: Unique identifier for the pipeline
- **Handler**: The function that processes the request
- **Priority**: Execution order (higher priority executes first)
- **Enabled**: Whether the pipeline is active
- **NextPipeline**: Pipeline to chain to (when result is PipelineResultChain)
- **ViewHandler**: View to execute (when result is PipelineResultView)
- **Async**: Whether to run asynchronously
- **Timeout**: Maximum execution time in milliseconds

## Basic Usage

### Creating a Pipeline

```go
engine := pkg.NewPipelineEngine()

config := pkg.PipelineConfig{
    Name: "data-processor",
    Handler: func(ctx pkg.Context) (pkg.PipelineResult, error) {
        // Process data
        data := ctx.FormValue("data")
        
        // Access database
        if ctx.DB() != nil {
            // Save to database
        }
        
        // Continue to next step
        return pkg.PipelineResultContinue, nil
    },
    Enabled: true,
    Priority: 10,
}

engine.Register(config)
```

### Using the Pipeline Builder

```go
pipeline := pkg.NewPipelineBuilder("validator").
    WithHandler(func(ctx pkg.Context) (pkg.PipelineResult, error) {
        // Validation logic
        return pkg.PipelineResultContinue, nil
    }).
    WithPriority(100).
    WithTimeout(5000).
    Build()

engine.Register(pipeline)
```

### Executing Pipelines

```go
// Execute a single pipeline
result, err := engine.Execute(ctx, "data-processor")

// Execute a chain of pipelines
result, err := engine.ExecuteChain(ctx, []string{"validator", "processor", "logger"})

// Execute asynchronously
err := engine.ExecuteAsync(ctx, "background-task")

// Execute multiple pipelines concurrently
results, errors := engine.ExecuteMultiplex(ctx, []string{"task1", "task2", "task3"})
```

## Advanced Features

### Pipeline Chaining

Pipelines can automatically chain to other pipelines:

```go
firstPipeline := pkg.PipelineConfig{
    Name: "first",
    Handler: func(ctx pkg.Context) (pkg.PipelineResult, error) {
        // Do some work
        return pkg.PipelineResultChain, nil
    },
    NextPipeline: "second",
    Enabled: true,
}

secondPipeline := pkg.PipelineConfig{
    Name: "second",
    Handler: func(ctx pkg.Context) (pkg.PipelineResult, error) {
        // Continue processing
        return pkg.PipelineResultContinue, nil
    },
    Enabled: true,
}

engine.Register(firstPipeline)
engine.Register(secondPipeline)

// Executing "first" will automatically execute "second"
result, err := engine.Execute(ctx, "first")
```

### View Execution

Pipelines can execute views after processing:

```go
pipeline := pkg.PipelineConfig{
    Name: "render-data",
    Handler: func(ctx pkg.Context) (pkg.PipelineResult, error) {
        // Prepare data
        return pkg.PipelineResultView, nil
    },
    ViewHandler: func(ctx pkg.Context) error {
        return ctx.JSON(200, map[string]string{
            "message": "Data processed successfully",
        })
    },
    Enabled: true,
}
```

### Asynchronous Processing

For background tasks that don't need to block the response:

```go
backgroundPipeline := pkg.PipelineConfig{
    Name: "send-email",
    Handler: func(ctx pkg.Context) (pkg.PipelineResult, error) {
        // Send email in background
        time.Sleep(2 * time.Second)
        return pkg.PipelineResultContinue, nil
    },
    Enabled: true,
    Async: true,
}

engine.Register(backgroundPipeline)

// Start async execution
engine.ExecuteAsync(ctx, "send-email")

// Continue processing request immediately

// Later, wait for all async pipelines
engine.WaitAll()
```

### Pipeline Multiplexing

Execute multiple pipelines concurrently for parallel processing:

```go
// Execute multiple pipelines at once
results, errors := engine.ExecuteMultiplex(ctx, []string{
    "validate-user",
    "check-permissions",
    "load-preferences",
})

// Check results
for i, result := range results {
    if errors[i] != nil {
        log.Printf("Pipeline %d failed: %v", i, errors[i])
    }
}
```

### Timeout Handling

Set maximum execution time for pipelines:

```go
pipeline := pkg.PipelineConfig{
    Name: "external-api-call",
    Handler: func(ctx pkg.Context) (pkg.PipelineResult, error) {
        // Call external API
        return pkg.PipelineResultContinue, nil
    },
    Enabled: true,
    Timeout: 3000, // 3 second timeout
}
```

## Integration with Framework Features

### Database Access

```go
pipeline := pkg.PipelineConfig{
    Name: "save-data",
    Handler: func(ctx pkg.Context) (pkg.PipelineResult, error) {
        db := ctx.DB()
        if db == nil {
            return pkg.PipelineResultClose, errors.New("database not available")
        }
        
        // Execute database operations
        _, err := db.Exec("INSERT INTO users (name) VALUES (?)", "John")
        if err != nil {
            return pkg.PipelineResultClose, err
        }
        
        return pkg.PipelineResultContinue, nil
    },
    Enabled: true,
}
```

### Session Management

```go
pipeline := pkg.PipelineConfig{
    Name: "session-handler",
    Handler: func(ctx pkg.Context) (pkg.PipelineResult, error) {
        session := ctx.Session()
        if session == nil {
            return pkg.PipelineResultClose, errors.New("session not available")
        }
        
        // Access session data
        userID, _ := session.Get(ctx.Request().SessionID, "user_id")
        
        return pkg.PipelineResultContinue, nil
    },
    Enabled: true,
}
```

### Cache Operations

```go
pipeline := pkg.PipelineConfig{
    Name: "cache-handler",
    Handler: func(ctx pkg.Context) (pkg.PipelineResult, error) {
        cache := ctx.Cache()
        if cache == nil {
            return pkg.PipelineResultContinue, nil
        }
        
        // Check cache
        data, err := cache.Get("key")
        if err == nil {
            // Use cached data
            return pkg.PipelineResultView, nil
        }
        
        // Cache miss, continue processing
        return pkg.PipelineResultContinue, nil
    },
    ViewHandler: func(ctx pkg.Context) error {
        return ctx.JSON(200, "cached-data")
    },
    Enabled: true,
}
```

### File Operations

```go
pipeline := pkg.PipelineConfig{
    Name: "file-upload-handler",
    Handler: func(ctx pkg.Context) (pkg.PipelineResult, error) {
        files := ctx.Files()
        if files == nil {
            return pkg.PipelineResultClose, errors.New("file manager not available")
        }
        
        // Handle file upload
        err := files.SaveUploadedFile(ctx, "document", "/uploads/doc.pdf")
        if err != nil {
            return pkg.PipelineResultClose, err
        }
        
        return pkg.PipelineResultContinue, nil
    },
    Enabled: true,
}
```

### Form Data Validation

```go
pipeline := pkg.PipelineConfig{
    Name: "form-validator",
    Handler: func(ctx pkg.Context) (pkg.PipelineResult, error) {
        // Validate required fields
        username := ctx.FormValue("username")
        if username == "" {
            return pkg.PipelineResultClose, errors.New("username is required")
        }
        
        email := ctx.FormValue("email")
        if !isValidEmail(email) {
            return pkg.PipelineResultClose, errors.New("invalid email")
        }
        
        // Check uploaded files
        file, err := ctx.FormFile("avatar")
        if err != nil {
            return pkg.PipelineResultClose, errors.New("avatar is required")
        }
        
        if file.Size > 5*1024*1024 {
            return pkg.PipelineResultClose, errors.New("file too large")
        }
        
        return pkg.PipelineResultContinue, nil
    },
    Enabled: true,
}
```

### Logging

```go
pipeline := pkg.PipelineConfig{
    Name: "request-logger",
    Handler: func(ctx pkg.Context) (pkg.PipelineResult, error) {
        logger := ctx.Logger()
        if logger != nil {
            logger.WithContext(ctx).Info("Request processed")
        }
        
        return pkg.PipelineResultContinue, nil
    },
    Enabled: true,
    Priority: 1, // Low priority - log after processing
}
```

### API Integration

```go
pipeline := pkg.PipelineConfig{
    Name: "api-handler",
    Handler: func(ctx pkg.Context) (pkg.PipelineResult, error) {
        // Handle API request
        data := map[string]interface{}{
            "status": "success",
            "data": "processed",
        }
        
        return pkg.PipelineResultView, nil
    },
    ViewHandler: func(ctx pkg.Context) error {
        return ctx.JSON(200, map[string]string{
            "status": "success",
        })
    },
    Enabled: true,
}
```

## Pipeline Management

### Enable/Disable Pipelines

```go
// Disable a pipeline
engine.Disable("pipeline-name")

// Enable a pipeline
engine.Enable("pipeline-name")
```

### Change Priority

```go
// Set higher priority to execute first
engine.SetPriority("critical-pipeline", 100)

// Set lower priority to execute last
engine.SetPriority("logging-pipeline", 1)
```

### List Pipelines

```go
pipelines := engine.List()
for _, p := range pipelines {
    fmt.Printf("Pipeline: %s, Priority: %d, Enabled: %v\n", 
        p.Name, p.Priority, p.Enabled)
}
```

### Get Pipeline Configuration

```go
config, err := engine.Get("pipeline-name")
if err == nil {
    fmt.Printf("Priority: %d\n", config.Priority)
}
```

### Clear All Pipelines

```go
engine.Clear()
```

## Best Practices

1. **Use Descriptive Names**: Give pipelines clear, descriptive names that indicate their purpose
2. **Set Appropriate Priorities**: Use priorities to control execution order (100 for critical, 50 for normal, 1 for logging)
3. **Handle Errors Gracefully**: Always return appropriate errors and results
4. **Use Async for Background Tasks**: Don't block the response for non-critical operations
5. **Leverage Multiplexing**: Use concurrent execution for independent operations
6. **Set Timeouts**: Protect against long-running operations with timeouts
7. **Chain Logically**: Use chaining for sequential operations that depend on each other
8. **Close Connections Appropriately**: Use PipelineResultClose when the request should not continue

## Common Use Cases

### Request Validation Pipeline

```go
validationPipeline := pkg.NewPipelineBuilder("request-validator").
    WithHandler(func(ctx pkg.Context) (pkg.PipelineResult, error) {
        // Validate request size
        if len(ctx.Body()) > 10*1024*1024 {
            return pkg.PipelineResultClose, errors.New("request too large")
        }
        
        // Validate content type
        contentType := ctx.GetHeader("Content-Type")
        if contentType != "application/json" {
            return pkg.PipelineResultClose, errors.New("invalid content type")
        }
        
        return pkg.PipelineResultContinue, nil
    }).
    WithPriority(100).
    Build()
```

### Rate Limiting Pipeline

```go
rateLimitPipeline := pkg.NewPipelineBuilder("rate-limiter").
    WithHandler(func(ctx pkg.Context) (pkg.PipelineResult, error) {
        cache := ctx.Cache()
        userID := ctx.User().ID
        
        key := fmt.Sprintf("rate_limit:%s", userID)
        count, _ := cache.Increment(key, 1)
        
        if count == 1 {
            cache.Expire(key, time.Minute)
        }
        
        if count > 100 {
            return pkg.PipelineResultClose, errors.New("rate limit exceeded")
        }
        
        return pkg.PipelineResultContinue, nil
    }).
    WithPriority(90).
    Build()
```

### Data Transformation Pipeline

```go
transformPipeline := pkg.NewPipelineBuilder("data-transformer").
    WithHandler(func(ctx pkg.Context) (pkg.PipelineResult, error) {
        // Transform request data
        rawData := ctx.Body()
        
        // Process and transform
        transformedData := transform(rawData)
        
        // Store in context for next pipeline
        // (implementation depends on context extensions)
        
        return pkg.PipelineResultContinue, nil
    }).
    WithPriority(50).
    Build()
```

## Requirements Validation

This implementation satisfies the following requirements:

- **Requirement 7.4**: Pipelines can process data without requiring a view
- **Requirement 7.5**: Pipelines support connection closure, chaining, and view execution
- **Requirement 7.6**: Pipeline multiplexing is supported via goroutines
- **Requirement 7.7**: Pipelines integrate with API, cookies, sessions, redirect, database access, file access, form data checking, logging, and cache access

## Testing

The pipeline system includes comprehensive unit tests covering:

- Pipeline registration and management
- Execution with different results
- Chaining and view execution
- Asynchronous execution
- Multiplexed execution
- Timeout handling
- Error handling
- Priority ordering

Run tests with:

```bash
go test -v -run TestPipeline ./pkg
```

## Example Application

See `examples/pipeline_example.go` for a complete demonstration of pipeline features including:

- Basic pipeline execution
- Pipeline chaining
- View execution
- Asynchronous processing
- Multiplexed execution
- Timeout handling
- Integration with framework features

Run the example:

```bash
go run examples/pipeline_example.go
```
