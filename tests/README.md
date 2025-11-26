# Rockstar Web Framework - Integration Tests

This directory contains comprehensive integration tests and benchmarks for the Rockstar Web Framework. The test suite validates all framework features including multi-protocol support, authentication, authorization, multi-tenancy, HTTP/2 cancellation, and performance characteristics.

## Test Structure

The test suite is organized into the following files:

- `integration_multiprotocol_test.go` - Multi-protocol support tests
- `integration_auth_test.go` - Authentication and authorization tests
- `integration_multitenancy_test.go` - Multi-tenancy isolation tests
- `http2_cancellation_test.go` - HTTP/2 request cancellation tests
- `benchmark_test.go` - Performance benchmarks
- `test_helpers.go` - Test utilities and mock implementations

### Integration Tests

#### 1. Multi-Protocol Tests (`integration_multiprotocol_test.go`)
Tests end-to-end functionality for all supported protocols:
- **HTTP/1.1**: Basic request/response handling, GET/POST operations
- **HTTP/2**: Protocol support and configuration
- **WebSocket**: WebSocket routing and echo functionality
- **REST API**: RESTful resource operations (LIST, GET, CREATE, UPDATE, DELETE)
- **GraphQL**: GraphQL schema and query execution
- **gRPC**: gRPC service registration and method calls
- **SOAP**: SOAP service operations and WSDL generation
- **Concurrent Access**: Multi-protocol concurrent request handling
- **Graceful Shutdown**: Shutdown behavior with active connections

#### 2. Authentication & Authorization Tests (`integration_auth_test.go`)
Tests security features end-to-end:
- **OAuth2 Authentication**: Token-based authentication flow
- **JWT Authentication**: JWT generation and validation
- **Role-Based Authorization**: RBAC with admin/user roles
- **Action-Based Authorization**: Permission-based access control
- **Session Authentication**: Session-based authentication flow
- **Rate Limiting**: API rate limiting enforcement

#### 3. Multi-Tenancy Tests (`integration_multitenancy_test.go`)
Tests tenant isolation and multi-tenancy features:
- **Host-Based Routing**: Routing requests to correct tenant by hostname
- **Data Isolation**: Ensuring tenant data doesn't leak between tenants
- **Session Isolation**: Separate session management per tenant
- **Authentication Isolation**: Tenant-specific authentication tokens
- **Configuration Isolation**: Independent configuration per tenant
- **Rate Limiting Isolation**: Separate rate limits per tenant
- **Workload Metrics**: Tenant-specific performance metrics
- **Concurrent Access**: Concurrent requests from multiple tenants

#### 4. HTTP/2 Cancellation Tests (`http2_cancellation_test.go`)
Tests HTTP/2 request cancellation handling:
- **Stream Cancellation**: HTTP/2 stream cancellation middleware integration
- **Cancellation Detection**: Handler cancellation detection and response
- **Cancellation Utilities**: IsCancellationError utility function validation
- **Client Cancellation**: Client-side request cancellation handling

### Performance Benchmarks (`benchmark_test.go`)

Comprehensive benchmarks for performance testing and framework comparison.

#### Core Benchmarks
- **Simple Route**: Basic GET request performance
- **JSON Response**: JSON serialization performance
- **Route Parameters**: Dynamic route parameter extraction
- **POST Request**: POST request with JSON body
- **Middleware**: Middleware execution overhead
- **Multiple Routes**: Routing performance with many routes
- **Concurrent Requests**: High-concurrency request handling
- **REST API**: REST API operation performance
- **Authentication**: Authentication overhead measurement
- **Memory Allocation**: Memory allocation patterns

#### Comparison Benchmarks
Framework comparison structure (ready for GoFiber and Gin):
- Simple route comparison
- JSON response comparison
- Route parameter comparison
- Middleware comparison

### Test Helpers (`test_helpers.go`)

Utility functions and mock implementations that support test execution:

#### Mock Implementations
- **mockContext**: Mock implementation of Context interface for testing
- **testMockDB**: Mock DatabaseManager with in-memory storage
- **mockVirtualFS**: Mock VirtualFS for testing file serving

#### Helper Functions
- **Data Creation**: `createTestUser()`, `createTestTenant()`, `createTestAccessToken()`, `createTestSession()`, `createTestServerConfig()`, `createTestSessionConfig()`
- **Assertions**: `assertEqual()`, `assertNotNil()`, `assertNil()`, `assertNoError()`, `assertError()`, `assertTrue()`, `assertFalse()`
- **HTTP Requests**: `makeGetRequest()`, `makePostRequest()`
- **Concurrency**: `runConcurrent()`, `runConcurrentWithErrors()`
- **Timing**: `measureExecutionTime()`, `waitForCondition()`
- **Cleanup**: `cleanupServer()`, `cleanupServers()`

## Running Tests

### Run All Integration Tests
```bash
go test -v ./tests/
```

### Run Specific Test Suite
```bash
# Multi-protocol tests
go test -v ./tests/ -run TestMultiProtocol

# Authentication tests
go test -v ./tests/ -run TestAuthentication

# Multi-tenancy tests
go test -v ./tests/ -run TestMultiTenancy
```

### Run Benchmarks
```bash
# Run all benchmarks
go test -bench=. ./tests/

# Run specific benchmark
go test -bench=BenchmarkRockstarSimpleRoute ./tests/

# Run benchmarks with memory profiling
go test -bench=. -benchmem ./tests/

# Run benchmarks with CPU profiling
go test -bench=. -cpuprofile=cpu.prof ./tests/

# Run comparison benchmarks
go test -bench=BenchmarkComparison ./tests/
```

### Run with Coverage
```bash
# Run with coverage report
go test -v -cover ./tests/

# Generate coverage profile
go test -v -coverprofile=coverage.out ./tests/

# View coverage in browser
go tool cover -html=coverage.out

# Generate coverage for specific package
go test -v -coverprofile=coverage.out -coverpkg=./pkg/... ./tests/
```

### Run with Profiling
```bash
# CPU profiling
go test -bench=. -cpuprofile=cpu.prof ./tests/
go tool pprof cpu.prof

# Memory profiling
go test -bench=. -memprofile=mem.prof ./tests/
go tool pprof mem.prof

# Block profiling
go test -bench=. -blockprofile=block.prof ./tests/
go tool pprof block.prof

# Trace profiling
go test -trace=trace.out ./tests/
go tool trace trace.out
```

## Test Requirements

### Prerequisites
- **Go Version**: Go 1.24.0 or higher
- **Dependencies**: All framework dependencies installed via `go mod download`
- **Network**: Available ports for test servers (19001-19310)
- **System Resources**: Sufficient memory for concurrent tests (recommended 4GB+)

### Port Allocation
Tests use dedicated port ranges to avoid conflicts:
- **19001-19010**: Multi-protocol tests
- **19101-19110**: Authentication tests
- **19201-19210**: Multi-tenancy tests
- **19301-19310**: Benchmark tests

### Test Data and Mocks
Tests use mock implementations for fast, isolated testing:
- **Database Operations**: In-memory mock database (`testMockDB`)
- **Cache Operations**: In-memory cache implementation
- **Virtual Filesystems**: Mock filesystem (`mockVirtualFS`)
- **Session Storage**: In-memory session storage
- **Context**: Mock context implementation (`mockContext`)

**No external dependencies** (databases, caches, message queues) are required. All tests run in isolation with mock implementations.

## Test Coverage

### Protocol Support Coverage
The test suite validates all supported protocols:
- ✅ **HTTP/1.1**: GET, POST, JSON handling, route parameters
- ✅ **HTTP/2**: Protocol enablement, configuration, stream handling
- ✅ **WebSocket**: Route registration, echo functionality
- ✅ **REST API**: LIST, GET, CREATE operations, RESTful routing
- ✅ **Graceful Shutdown**: Shutdown with active connections
- ✅ **Concurrent Access**: Multiple protocols simultaneously

### Security Features Coverage
Comprehensive authentication and authorization testing:
- ✅ **OAuth2 Authentication**: Token creation, validation, protected routes
- ✅ **JWT Authentication**: Generation, validation, claims extraction
- ✅ **Role-Based Authorization**: Admin/user role access control
- ✅ **Action-Based Authorization**: Read/write permission control
- ✅ **Session Management**: Login flow, session persistence, cookie handling
- ✅ **Rate Limiting**: Limit enforcement, 429 status codes

### Multi-Tenancy Coverage
Complete tenant isolation validation:
- ✅ **Host-Based Routing**: Correct tenant routing by hostname
- ✅ **Data Isolation**: No data leakage between tenants
- ✅ **Session Isolation**: Tenant-specific session management
- ✅ **Authentication Isolation**: Tenant-specific tokens and context
- ✅ **Configuration Isolation**: Independent tenant configurations
- ✅ **Rate Limiting Isolation**: Separate quotas per tenant
- ✅ **Workload Metrics Isolation**: Tenant-specific metrics
- ✅ **Concurrent Access**: Multiple tenants under load

### HTTP/2 Cancellation Coverage
Request cancellation handling:
- ✅ **Cancellation Middleware**: Integration and detection
- ✅ **Cancellation Utilities**: Error type detection
- ✅ **Client Cancellation**: Client-side cancellation handling

### Performance Coverage
Comprehensive performance benchmarking:
- ✅ **Request Throughput**: Requests per second measurements
- ✅ **Latency Measurements**: Response time tracking
- ✅ **Concurrent Request Handling**: High-concurrency scenarios
- ✅ **Memory Allocation**: Allocation patterns and efficiency
- ✅ **Middleware Overhead**: Middleware performance impact
- ✅ **Framework Comparison**: Structure for comparing with GoFiber and Gin

## Performance Targets

Based on requirements to match or exceed GoFiber and Gin:

### Target Metrics
- **Simple Route**: > 50,000 req/sec
- **JSON Response**: > 40,000 req/sec
- **Route Parameters**: > 45,000 req/sec
- **Concurrent (100 goroutines)**: > 30,000 req/sec
- **Memory per Request**: < 5KB allocations

### Comparison
Run comparison benchmarks to validate performance against GoFiber and Gin:
```bash
go test -bench=BenchmarkComparison -benchmem ./tests/
```

## Continuous Integration

### CI Pipeline Integration

#### GitHub Actions Example
```yaml
name: Test Suite

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.24'
      
      - name: Install Dependencies
        run: go mod download
      
      - name: Run Integration Tests
        run: go test -v -timeout 30m ./tests/
      
      - name: Run Benchmarks
        run: go test -bench=. -benchmem ./tests/
      
      - name: Generate Coverage
        run: go test -coverprofile=coverage.out ./tests/
      
      - name: Upload Coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
```

#### GitLab CI Example
```yaml
test:
  stage: test
  image: golang:1.24
  script:
    - go mod download
    - go test -v -timeout 30m ./tests/
    - go test -bench=. -benchmem ./tests/
    - go test -coverprofile=coverage.out ./tests/
  artifacts:
    reports:
      coverage_report:
        coverage_format: cobertura
        path: coverage.out
```

### Performance Regression Detection

Monitor performance over time to catch regressions:

```bash
# Create baseline
go test -bench=. -benchmem ./tests/ | tee baseline.txt

# After changes, compare
go test -bench=. -benchmem ./tests/ | tee current.txt

# Use benchcmp or benchstat for comparison
go install golang.org/x/perf/cmd/benchstat@latest
benchstat baseline.txt current.txt
```

Set up automated alerts for performance regressions:
- **Threshold**: Alert if performance degrades by >10%
- **Memory**: Alert if allocations increase by >15%
- **Latency**: Alert if p99 latency increases by >20%

## Troubleshooting

### Common Issues and Solutions

#### Port Already in Use
If tests fail with "address already in use" error:

**Linux/Mac:**
```bash
# Find processes using test ports
lsof -ti:19001-19310 | xargs kill -9

# Or for specific port
lsof -ti:19001 | xargs kill -9
```

**Windows:**
```powershell
# Find processes using test ports
netstat -ano | findstr :19001
taskkill /PID <PID> /F
```

#### Test Timeout Issues
If tests timeout, increase the timeout duration:
```bash
# Increase to 30 minutes
go test -timeout 30m ./tests/

# For specific slow tests
go test -timeout 30m -run TestMultiTenancyConcurrentAccess ./tests/
```

#### Memory Issues
If tests fail due to memory constraints:
```bash
# Limit concurrent tests
GOMAXPROCS=2 go test ./tests/

# Run tests sequentially
go test -p 1 ./tests/

# Increase available memory (Docker)
docker run --memory=4g golang:1.24 go test ./tests/
```

#### Race Condition Detection
Run tests with race detector:
```bash
go test -race ./tests/

# Note: Race detector increases memory usage and slows tests
```

#### Flaky Tests
If tests fail intermittently:
```bash
# Run tests multiple times
go test -count=10 ./tests/

# Run specific flaky test
go test -count=100 -run TestSpecificFlakyTest ./tests/
```

#### Build Failures
If tests fail to compile:
```bash
# Clean build cache
go clean -cache -testcache

# Verify dependencies
go mod verify
go mod tidy

# Rebuild
go test -v ./tests/
```

#### Coverage Issues
If coverage reports are incomplete:
```bash
# Include all packages
go test -coverprofile=coverage.out -coverpkg=./... ./tests/

# Check coverage for specific package
go test -coverprofile=coverage.out -coverpkg=./pkg/... ./tests/
```

## Contributing

### Adding New Tests

When adding new features to the framework:

1. **Add Integration Tests**: Add tests to the appropriate test file
   - Multi-protocol features → `integration_multiprotocol_test.go`
   - Security features → `integration_auth_test.go`
   - Multi-tenancy features → `integration_multitenancy_test.go`
   - HTTP/2 features → `http2_cancellation_test.go`

2. **Add Benchmarks**: Add benchmarks for performance-critical features to `benchmark_test.go`
   - Follow naming convention: `BenchmarkRockstar<Feature>`
   - Use `b.ReportAllocs()` for memory tracking
   - Add comparison benchmarks if applicable

3. **Update Test Helpers**: Add helper functions to `test_helpers.go` if needed
   - Mock implementations for new interfaces
   - Data creation helpers for new models
   - Assertion helpers for new validation patterns

4. **Update Documentation**: Update this README with:
   - New test coverage areas
   - New helper functions
   - New troubleshooting tips
   - Updated performance targets

5. **Verify Tests Pass**: Ensure all tests pass before submitting PR
   ```bash
   go test -v ./tests/
   go test -race ./tests/
   go test -bench=. ./tests/
   ```

### Code Review Checklist

- [ ] Tests follow existing patterns and conventions
- [ ] Tests are isolated and don't depend on execution order
- [ ] Tests clean up resources (servers, connections)
- [ ] Tests use appropriate timeouts
- [ ] Tests use unique ports (check port allocation)
- [ ] Benchmarks report allocations
- [ ] Documentation is updated
- [ ] All tests pass locally
- [ ] No race conditions detected

## Test Maintenance

### Regular Maintenance Tasks

#### Weekly
- Run full test suite with race detector
- Review test execution times for slowdowns
- Check for flaky tests

#### Monthly
- Update performance baselines
- Review and update test timeouts
- Clean up deprecated tests
- Update mock implementations for interface changes

#### Per Release
- Run benchmarks and compare to previous release
- Update performance targets if needed
- Review test coverage and add tests for gaps
- Update comparison benchmarks for framework updates
- Verify CI pipeline configuration

### Performance Baseline Management

Maintain performance baselines for regression detection:

```bash
# Create baseline for release
go test -bench=. -benchmem ./tests/ | tee baselines/v1.0.0.txt

# Compare current to baseline
go test -bench=. -benchmem ./tests/ > current.txt
benchstat baselines/v1.0.0.txt current.txt

# Archive baseline
git add baselines/v1.0.0.txt
git commit -m "Add performance baseline for v1.0.0"
```

### Updating Mock Implementations

When framework interfaces change:

1. Update mock implementations in `test_helpers.go`
2. Run tests to identify failures
3. Update test assertions if behavior changed
4. Verify all tests pass
5. Update documentation if mock behavior changed

### Deprecating Tests

When removing features:

1. Mark tests as deprecated with comment
2. Update documentation
3. Remove tests in next major version
4. Ensure coverage doesn't drop below targets
