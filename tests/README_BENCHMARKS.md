# Running Benchmarks

The benchmark tests are separated from integration tests using build tags to avoid conflicts and allow independent execution.

## Running All Benchmarks

```bash
# Run all benchmarks with default settings
go test -tags=benchmark -bench=. -benchmem

# Run with shorter benchmark time for quick testing
go test -tags=benchmark -bench=. -benchmem -benchtime=200ms

# Run with specific timeout
go test -tags=benchmark -bench=. -benchmem -timeout=5m
```

## Running Specific Benchmark Groups

```bash
# Run only Rockstar framework benchmarks
go test -tags=benchmark -bench=BenchmarkRockstar -benchmem

# Run only comparison benchmarks
go test -tags=benchmark -bench=BenchmarkComparison -benchmem
```

## Running Individual Benchmarks

```bash
# Run a specific benchmark
go test -tags=benchmark -bench=BenchmarkRockstarSimpleRoute$ -benchmem

# Run with single iteration for testing
go test -tags=benchmark -bench=BenchmarkRockstarSimpleRoute$ -benchtime=1x
```

## Available Benchmarks

### Rockstar Framework Benchmarks
- `BenchmarkRockstarSimpleRoute` - Simple GET request handling
- `BenchmarkRockstarJSONResponse` - JSON serialization
- `BenchmarkRockstarRouteParams` - Route parameter extraction
- `BenchmarkRockstarPOSTRequest` - POST request handling
- `BenchmarkRockstarMiddleware` - Middleware execution overhead
- `BenchmarkRockstarMultipleRoutes` - Routing with many routes
- `BenchmarkRockstarConcurrentRequests` - Concurrent request handling
- `BenchmarkRockstarRESTAPI` - REST operations
- `BenchmarkRockstarAuthentication` - Authentication overhead
- `BenchmarkRockstarMemoryAllocation` - Memory allocation patterns

### Comparison Benchmarks
- `BenchmarkComparison_SimpleRoute` - Compare simple route performance
- `BenchmarkComparison_JSONResponse` - Compare JSON response performance
- `BenchmarkComparison_RouteParams` - Compare route parameter extraction
- `BenchmarkComparison_Middleware` - Compare middleware execution

## Build Tags

The benchmarks use Go build tags to separate them from integration tests:

- Benchmark tests: `//go:build benchmark`
- Integration tests: `//go:build !benchmark`

This allows you to:
- Run benchmarks without running integration tests
- Run integration tests without running benchmarks
- Avoid port conflicts and test interference

## Notes

- Each benchmark starts its own server on a unique port
- Benchmarks include proper server startup checks to avoid connection errors
- The concurrent requests benchmark uses moderate parallelism (10) to avoid overwhelming the server
- All benchmarks properly clean up resources with `defer framework.Shutdown()`
