# Contributing to Rockstar Web Framework

Thank you for your interest in contributing to the Rockstar Web Framework! This document provides guidelines and instructions for contributing.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Project Structure](#project-structure)
- [Coding Standards](#coding-standards)
- [Testing Guidelines](#testing-guidelines)
- [Documentation](#documentation)
- [Submitting Changes](#submitting-changes)
- [Review Process](#review-process)

## Code of Conduct

We are committed to providing a welcoming and inclusive environment. Please be respectful and professional in all interactions.

## Getting Started

1. Fork the repository on GitHub
2. Clone your fork locally
3. Create a feature branch from `main`
4. Make your changes
5. Test your changes thoroughly
6. Submit a pull request

## Development Setup

### Prerequisites

- Go 1.25 or higher
- Git
- Make (optional, for convenience commands)

### Setup Steps

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/rockstar-web-framework.git
cd rockstar-web-framework

# Add upstream remote
git remote add upstream https://github.com/echterhof/rockstar-web-framework.git

# Install dependencies
go mod download

# Run tests to verify setup
make test
```

## Project Structure

```
rockstar-web-framework/
├── cmd/                   # Main applications and CLI tools
├── pkg/                   # Framework library code (single package)
├── examples/              # Example applications
├── tests/                 # Integration and benchmark tests
├── docs/                  # Documentation
├── sql/                   # Database-specific SQL queries
├── locales/               # i18n locale files
├── plugins/               # Runtime plugin directory
└── scripts/               # Build and utility scripts
```

### Key Principles

- All framework code is in the `pkg/` package (flat structure, no sub-packages)
- Examples are standalone programs in `examples/`
- SQL queries are externalized in `sql/{driver}/` directories
- Documentation is comprehensive and in `docs/`

## Coding Standards

### Go Style

Follow standard Go conventions:
- Use `gofmt` for formatting
- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use meaningful variable and function names
- Add comments for exported types and functions

### Naming Conventions

**Interfaces:**
- Use descriptive names: `Context`, `RouterEngine`, `DatabaseManager`
- Manager pattern for services: `{Component}Manager`

**Implementations:**
- Use `{name}Impl` suffix or `default{Name}` prefix
- Example: `routerImpl`, `defaultDatabase`

**Files:**
- Interfaces: `{component}.go`
- Implementations: `{component}_impl.go`
- Tests: `{component}_test.go`
- Platform-specific: `{component}_{os}.go`

### Code Organization

```go
// 1. Package declaration
package pkg

// 2. Imports (grouped: stdlib, external, internal)
import (
    "context"
    "net/http"
    
    "github.com/gorilla/websocket"
    
    "github.com/echterhof/rockstar-web-framework/pkg"
)

// 3. Constants
const (
    DefaultTimeout = 30 * time.Second
)

// 4. Types (interfaces first, then structs)
type Manager interface {
    // Methods
}

type managerImpl struct {
    // Fields
}

// 5. Constructor functions
func NewManager() Manager {
    return &managerImpl{}
}

// 6. Methods
func (m *managerImpl) Method() error {
    // Implementation
}

// 7. Helper functions
func helperFunction() {
    // Implementation
}
```

### Error Handling

```go
// Return errors, don't panic
func DoSomething() error {
    if err := validate(); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }
    return nil
}

// Use error wrapping for context
return fmt.Errorf("failed to connect to database: %w", err)

// Check errors immediately
result, err := operation()
if err != nil {
    return err
}
```

### Context Usage

```go
// Always accept context as first parameter
func ProcessRequest(ctx context.Context, data []byte) error {
    // Use context for cancellation and timeouts
    select {
    case <-ctx.Done():
        return ctx.Err()
    case result := <-process(data):
        return result
    }
}
```

## Testing Guidelines

### Unit Tests

- Place tests in `pkg/*_test.go`
- Test file should match source file: `router.go` → `router_test.go`
- Use table-driven tests for multiple scenarios
- Mock external dependencies

```go
func TestRouterAddRoute(t *testing.T) {
    tests := []struct {
        name    string
        method  string
        path    string
        wantErr bool
    }{
        {"valid GET", "GET", "/users", false},
        {"valid POST", "POST", "/users", false},
        {"invalid method", "INVALID", "/users", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            router := NewRouter()
            err := router.AddRoute(tt.method, tt.path, handler)
            if (err != nil) != tt.wantErr {
                t.Errorf("AddRoute() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Integration Tests

- Place tests in `tests/*_test.go`
- Test complete workflows
- Use real dependencies when possible
- Clean up resources in `defer` or `t.Cleanup()`

```go
func TestFullWorkflow(t *testing.T) {
    app, err := pkg.New(testConfig)
    if err != nil {
        t.Fatal(err)
    }
    defer app.Shutdown(context.Background())
    
    // Test workflow
}
```

### Benchmarks

- Place benchmarks in `tests/benchmark_test.go`
- Use `b.ResetTimer()` before measured code
- Use `b.ReportAllocs()` to track allocations

```go
func BenchmarkRouterLookup(b *testing.B) {
    router := setupRouter()
    b.ResetTimer()
    b.ReportAllocs()
    
    for i := 0; i < b.N; i++ {
        router.Lookup("GET", "/users/123")
    }
}
```

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific package
go test ./pkg/...

# Run specific test
go test -run TestRouterAddRoute ./pkg/

# Run benchmarks
go test -bench=. ./tests/
```

## Documentation

### Code Documentation

- Document all exported types, functions, and methods
- Use complete sentences
- Include examples for complex functionality

```go
// Context provides access to request/response data and framework services.
// It is passed to all handler functions and middleware.
//
// Example:
//
//	func handler(ctx pkg.Context) error {
//	    user := ctx.Params().Get("user")
//	    return ctx.JSON(200, map[string]string{"user": user})
//	}
type Context interface {
    // Methods...
}
```

### Documentation Files

- Place all documentation in `docs/` directory
- Use Markdown format
- Include code examples with correct imports
- Keep examples runnable and complete

### Documentation Standards

- **Primary docs**: UPPERCASE names (e.g., `API_REFERENCE.md`)
- **Implementation guides**: lowercase names (e.g., `rest_api_implementation.md`)
- Use relative links: `[Link](docs/FILE.md)`
- Include table of contents for long documents
- Update `DOCUMENTATION_INDEX.md` when adding new docs

## Submitting Changes

### Branch Naming

- Feature: `feature/description`
- Bug fix: `fix/description`
- Documentation: `docs/description`
- Refactor: `refactor/description`

### Commit Messages

Follow conventional commits format:

```
type(scope): subject

body

footer
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting)
- `refactor`: Code refactoring
- `test`: Test changes
- `chore`: Build/tooling changes

Example:
```
feat(router): add support for route groups

Add Router.Group() method to create route groups with shared
middleware and path prefixes.

Closes #123
```

### Pull Request Process

1. Update documentation if needed
2. Add tests for new functionality
3. Ensure all tests pass: `make check`
4. Update CHANGELOG.md if applicable
5. Create pull request with clear description
6. Link related issues
7. Request review from maintainers

### Pull Request Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] All tests pass

## Documentation
- [ ] Code comments updated
- [ ] Documentation files updated
- [ ] Examples added/updated

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Tests added and passing
- [ ] Documentation updated
```

## Review Process

### What Reviewers Look For

- Code quality and style
- Test coverage
- Documentation completeness
- Performance implications
- Breaking changes
- Security considerations

### Addressing Feedback

- Respond to all comments
- Make requested changes
- Push updates to same branch
- Mark conversations as resolved when addressed

### Approval and Merge

- Requires approval from at least one maintainer
- All CI checks must pass
- No unresolved conversations
- Maintainers will merge when ready

## Additional Resources

- [Go Documentation](https://golang.org/doc/)
- [Effective Go](https://golang.org/doc/effective_go.html)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Project Documentation](DOCUMENTATION_INDEX.md)

## Questions?

- Open an issue for bugs or feature requests
- Start a discussion for questions or ideas
- Check existing documentation first

Thank you for contributing to Rockstar Web Framework! 🎸
