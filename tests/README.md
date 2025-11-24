# Rockstar Web Framework - Integration Tests

This directory contains comprehensive integration tests for the Rockstar Web Framework.

## Test Structure

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
- **Virtual Filesystem**: Per-tenant virtual filesystem isolation
- **Configuration Isolation**: Independent configuration per tenant
- **Rate Limiting Isolation**: Separate rate limits per tenant
- **Workload Metrics**: Tenant-specific performance metrics
- **Multiple Servers**: Running multiple servers for different tenants
- **Concurrent Access**: Concurrent requests from multiple tenants

### Performance Benchmarks (`benchmark_test.go`)

Comprehensive benchmarks for performance testing:

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
go test -v -cover ./tests/
go test -v -coverprofile=coverage.out ./tests/
go tool cover -html=coverage.out
```

## Test Requirements

### Prerequisites
- Go 1.24.0 or higher
- All framework dependencies installed (`go mod download`)
- Available ports for test servers (19001-19310)

### Test Data
Tests use mock implementations for:
- Database operations
- Cache operations
- Virtual filesystems
- Session storage

No external dependencies (databases, caches) are required.

## Test Coverage

### Protocol Support
- ✅ HTTP/1.1
- ✅ HTTP/2
- ✅ WebSocket
- ✅ REST API
- ✅ GraphQL
- ✅ gRPC
- ✅ SOAP

### Security Features
- ✅ OAuth2 Authentication
- ✅ JWT Authentication
- ✅ Role-Based Authorization
- ✅ Action-Based Authorization
- ✅ Session Management
- ✅ Rate Limiting

### Multi-Tenancy
- ✅ Host-Based Routing
- ✅ Data Isolation
- ✅ Session Isolation
- ✅ Authentication Isolation
- ✅ Configuration Isolation
- ✅ Resource Isolation

### Performance
- ✅ Request Throughput
- ✅ Latency Measurements
- ✅ Concurrent Request Handling
- ✅ Memory Allocation
- ✅ Middleware Overhead

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

### CI Pipeline
```yaml
# Example GitHub Actions workflow
- name: Run Integration Tests
  run: go test -v ./tests/

- name: Run Benchmarks
  run: go test -bench=. -benchmem ./tests/

- name: Generate Coverage
  run: go test -coverprofile=coverage.out ./tests/
```

## Troubleshooting

### Port Already in Use
If tests fail with "address already in use":
```bash
# Find and kill processes using test ports
lsof -ti:19001-19310 | xargs kill -9
```

### Timeout Issues
Increase test timeout:
```bash
go test -timeout 30m ./tests/
```

### Memory Issues
Run with increased memory:
```bash
GOMAXPROCS=4 go test ./tests/
```

## Contributing

When adding new features:
1. Add integration tests to appropriate test file
2. Add benchmarks for performance-critical features
3. Update this README with new test coverage
4. Ensure all tests pass before submitting PR

## Test Maintenance

### Regular Tasks
- Review and update test timeouts
- Add tests for new features
- Update benchmarks for performance regressions
- Maintain mock implementations
- Update comparison benchmarks when frameworks update

### Performance Regression Detection
Run benchmarks regularly and compare:
```bash
# Baseline
go test -bench=. -benchmem ./tests/ > baseline.txt

# After changes
go test -bench=. -benchmem ./tests/ > current.txt

# Compare
benchcmp baseline.txt current.txt
```
