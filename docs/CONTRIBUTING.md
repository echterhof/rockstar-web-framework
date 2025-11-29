# Contributing to Rockstar Web Framework

Thank you for your interest in contributing to the Rockstar Web Framework! We welcome contributions from the community and are grateful for your support.

This document provides guidelines for contributing to the project. Please read it carefully before submitting your contribution.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [How to Contribute](#how-to-contribute)
- [Development Setup](#development-setup)
- [Code Style Guidelines](#code-style-guidelines)
- [Testing Requirements](#testing-requirements)
- [Commit Message Guidelines](#commit-message-guidelines)
- [Pull Request Process](#pull-request-process)
- [Bug Reports](#bug-reports)
- [Feature Requests](#feature-requests)
- [Documentation](#documentation)
- [Community](#community)

## Code of Conduct

We are committed to providing a welcoming and inclusive environment for all contributors. Please be respectful and professional in all interactions.

### Our Standards

- **Be Respectful**: Treat everyone with respect and consideration
- **Be Collaborative**: Work together constructively
- **Be Professional**: Maintain professionalism in all communications
- **Be Inclusive**: Welcome diverse perspectives and experiences
- **Be Constructive**: Provide helpful feedback and suggestions

### Unacceptable Behavior

- Harassment, discrimination, or offensive comments
- Personal attacks or trolling
- Publishing others' private information
- Any conduct that would be inappropriate in a professional setting

## Getting Started

### Prerequisites

- **Go 1.25+**: Required for development
- **Git**: For version control
- **Make**: For build automation (optional but recommended)
- **Text Editor/IDE**: VS Code, GoLand, or your preferred editor

### Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/rockstar-web-framework.git
   cd rockstar-web-framework
   ```
3. Add the upstream repository:
   ```bash
   git remote add upstream https://github.com/echterhof/rockstar-web-framework.git
   ```

### Build the Project

```bash
# Download dependencies
make deps

# Build the framework
make build

# Run tests
make test
```

## How to Contribute

### Types of Contributions

We welcome various types of contributions:

- **Bug Fixes**: Fix issues and improve stability
- **New Features**: Add new functionality to the framework
- **Documentation**: Improve or add documentation
- **Examples**: Create example applications
- **Tests**: Add or improve test coverage
- **Performance**: Optimize performance
- **Plugins**: Create new plugins for the framework

### Contribution Workflow

1. **Check Existing Issues**: Look for existing issues or create a new one
2. **Discuss First**: For major changes, discuss your approach in an issue first
3. **Create a Branch**: Create a feature branch from `main`
4. **Make Changes**: Implement your changes following our guidelines
5. **Test Thoroughly**: Ensure all tests pass and add new tests
6. **Submit PR**: Submit a pull request with a clear description

## Development Setup

### Project Structure

```
rockstar-web-framework/
├── cmd/                   # Main applications
├── pkg/                   # Framework library code
├── examples/              # Example applications
├── tests/                 # Integration tests
├── docs/                  # Documentation
├── sql/                   # Database queries
├── plugins/               # Plugin directory
└── scripts/               # Build scripts
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run plugin tests
make test-plugins

# Run specific package tests
go test ./pkg/...

# Run integration tests
go test ./tests/...

# Run benchmarks
go test -bench=. ./tests/
```

### Running Examples

```bash
# Run a specific example
go run examples/getting_started.go

# Build and run an example
go build -o app examples/getting_started.go
./app
```

## Code Style Guidelines

### Go Code Style

We follow standard Go conventions and best practices:

#### Formatting

- **Use `gofmt`**: All code must be formatted with `gofmt`
- **Run `go vet`**: Code must pass `go vet` checks
- **Line Length**: Aim for 100 characters, but readability takes precedence
- **Imports**: Group imports into standard library, external, and internal

```go
import (
    // Standard library
    "context"
    "fmt"
    "time"
    
    // External packages
    "github.com/gorilla/websocket"
    "gopkg.in/yaml.v3"
    
    // Internal packages
    "github.com/echterhof/rockstar-web-framework/pkg"
)
```

#### Naming Conventions

- **Interfaces**: Use descriptive names ending in `Manager` or `Engine` (e.g., `DatabaseManager`, `RouterEngine`)
- **Implementations**: Use `{name}Impl` suffix or `default{Name}` prefix (e.g., `routerImpl`, `defaultDatabase`)
- **Exported Names**: Use clear, descriptive names (e.g., `NewFramework`, `HandleRequest`)
- **Private Names**: Use camelCase for unexported names (e.g., `handleError`, `parseConfig`)
- **Constants**: Use ALL_CAPS for constants (e.g., `MAX_CONNECTIONS`, `DEFAULT_TIMEOUT`)

#### Comments

- **Package Comments**: Every package should have a package comment
- **Exported Functions**: All exported functions must have comments
- **Complex Logic**: Comment complex algorithms or non-obvious code
- **TODO Comments**: Use `// TODO:` for future improvements

```go
// DatabaseManager provides database operations and connection management.
// It supports multiple database drivers including MySQL, PostgreSQL, SQLite, and MSSQL.
type DatabaseManager interface {
    // Connect establishes a connection to the database using the provided configuration.
    // Returns an error if the connection fails.
    Connect(config DatabaseConfig) error
    
    // Query executes a SQL query and returns the results.
    Query(query string, args ...interface{}) ([]map[string]interface{}, error)
}
```

#### Error Handling

- **Return Errors**: Return errors rather than panicking
- **Wrap Errors**: Use `fmt.Errorf` with `%w` to wrap errors
- **Check Errors**: Always check and handle errors
- **Error Messages**: Provide clear, actionable error messages

```go
// Good
if err := db.Connect(config); err != nil {
    return fmt.Errorf("failed to connect to database: %w", err)
}

// Bad
db.Connect(config) // Ignoring error
```

#### Testing

- **Test Files**: Name test files `{file}_test.go`
- **Test Functions**: Name test functions `Test{FunctionName}`
- **Table-Driven Tests**: Use table-driven tests for multiple cases
- **Test Coverage**: Aim for >80% test coverage for new code

```go
func TestDatabaseConnect(t *testing.T) {
    tests := []struct {
        name    string
        config  DatabaseConfig
        wantErr bool
    }{
        {
            name:    "valid config",
            config:  DatabaseConfig{Driver: "sqlite3", DSN: ":memory:"},
            wantErr: false,
        },
        {
            name:    "invalid driver",
            config:  DatabaseConfig{Driver: "invalid", DSN: ""},
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            db := NewDatabaseManager()
            err := db.Connect(tt.config)
            if (err != nil) != tt.wantErr {
                t.Errorf("Connect() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Documentation Style

- **Markdown**: Use Markdown for all documentation
- **Code Blocks**: Always specify language for syntax highlighting
- **Examples**: Include working code examples
- **Links**: Use relative links for internal documentation
- **Headings**: Use proper heading hierarchy (H1 → H2 → H3)

## Testing Requirements

### Test Coverage

All contributions must include appropriate tests:

- **Unit Tests**: Test individual functions and methods
- **Integration Tests**: Test component interactions
- **Property-Based Tests**: Test properties that should hold for all inputs (where applicable)
- **Benchmark Tests**: Include benchmarks for performance-critical code

### Test Requirements by Contribution Type

#### Bug Fixes
- Add a test that reproduces the bug
- Verify the test fails before the fix
- Verify the test passes after the fix

#### New Features
- Unit tests for all new functions
- Integration tests for feature workflows
- Property-based tests for critical logic
- Example code demonstrating the feature

#### Performance Improvements
- Benchmark tests showing improvement
- Verify no regression in existing benchmarks
- Document performance characteristics

### Running Tests Locally

Before submitting a PR, ensure all tests pass:

```bash
# Format code
make fmt

# Run static analysis
make vet

# Run all tests
make test

# Run all checks
make check
```

### Test Quality Standards

- **Descriptive Names**: Test names should clearly describe what is being tested
- **Independent**: Tests should not depend on each other
- **Deterministic**: Tests should produce consistent results
- **Fast**: Unit tests should run quickly
- **Isolated**: Use mocks/stubs for external dependencies where appropriate

## Commit Message Guidelines

### Commit Message Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- **feat**: New feature
- **fix**: Bug fix
- **docs**: Documentation changes
- **style**: Code style changes (formatting, no logic change)
- **refactor**: Code refactoring
- **perf**: Performance improvements
- **test**: Adding or updating tests
- **chore**: Build process or auxiliary tool changes

### Examples

```
feat(router): add support for route groups

Add RouteGroup functionality to organize routes with shared prefixes
and middleware. This improves code organization for large applications.

Closes #123
```

```
fix(database): prevent connection leak on query timeout

Ensure database connections are properly returned to the pool when
queries timeout. Previously, timed-out connections were not released,
leading to pool exhaustion.

Fixes #456
```

```
docs(security): add OAuth2 configuration examples

Add comprehensive examples for OAuth2 configuration including
authorization code flow, client credentials, and refresh tokens.
```

### Commit Message Rules

- Use present tense ("add feature" not "added feature")
- Use imperative mood ("move cursor to..." not "moves cursor to...")
- First line should be 50 characters or less
- Reference issues and pull requests in the footer
- Explain what and why, not how (code shows how)

## Pull Request Process

### Before Submitting

1. **Update Documentation**: Update relevant documentation
2. **Add Tests**: Include tests for your changes
3. **Run Tests**: Ensure all tests pass locally
4. **Format Code**: Run `make fmt` and `make vet`
5. **Update CHANGELOG**: Add entry to CHANGELOG.md (for significant changes)
6. **Rebase**: Rebase your branch on the latest `main`

### PR Title and Description

- **Title**: Use the same format as commit messages
- **Description**: Provide a clear description of the changes
- **Motivation**: Explain why the change is needed
- **Testing**: Describe how you tested the changes
- **Screenshots**: Include screenshots for UI changes
- **Breaking Changes**: Clearly mark any breaking changes

### PR Template

```markdown
## Description
Brief description of the changes

## Motivation
Why is this change needed?

## Changes
- List of changes made
- Another change

## Testing
How were these changes tested?

## Checklist
- [ ] Tests added/updated
- [ ] Documentation updated
- [ ] Code formatted (`make fmt`)
- [ ] Static analysis passed (`make vet`)
- [ ] All tests pass (`make test`)
- [ ] CHANGELOG updated (if applicable)

## Related Issues
Closes #123
```

### Review Process

1. **Automated Checks**: CI/CD will run automated tests
2. **Code Review**: Maintainers will review your code
3. **Feedback**: Address any feedback or requested changes
4. **Approval**: Once approved, your PR will be merged
5. **Merge**: Maintainers will merge your PR

### Review Criteria

- **Code Quality**: Clean, readable, maintainable code
- **Tests**: Adequate test coverage
- **Documentation**: Clear documentation
- **Performance**: No performance regressions
- **Compatibility**: Maintains backward compatibility (unless breaking change is justified)
- **Style**: Follows project style guidelines

## Bug Reports

### Before Reporting

1. **Search Existing Issues**: Check if the bug has already been reported
2. **Verify Bug**: Ensure it's actually a bug and not expected behavior
3. **Minimal Reproduction**: Create a minimal example that reproduces the bug
4. **Latest Version**: Verify the bug exists in the latest version

### Bug Report Template

```markdown
## Bug Description
Clear and concise description of the bug

## Steps to Reproduce
1. Step one
2. Step two
3. Step three

## Expected Behavior
What you expected to happen

## Actual Behavior
What actually happened

## Minimal Reproduction
```go
// Minimal code that reproduces the bug
```

## Environment
- OS: [e.g., Ubuntu 22.04]
- Go Version: [e.g., 1.25.4]
- Framework Version: [e.g., 1.0.0]

## Additional Context
Any other relevant information
```

### Bug Report Guidelines

- **Be Specific**: Provide detailed information
- **Be Concise**: Keep it focused on the issue
- **Be Respectful**: Remember that maintainers are volunteers
- **Follow Up**: Respond to questions and provide additional information

## Feature Requests

### Before Requesting

1. **Search Existing Issues**: Check if the feature has been requested
2. **Consider Scope**: Ensure the feature fits the framework's goals
3. **Think Through Design**: Consider implementation implications
4. **Provide Use Cases**: Explain real-world use cases

### Feature Request Template

```markdown
## Feature Description
Clear and concise description of the feature

## Motivation
Why is this feature needed? What problem does it solve?

## Proposed Solution
How should this feature work?

## Use Cases
Real-world scenarios where this feature would be useful

## Alternatives Considered
Other approaches you've considered

## Additional Context
Any other relevant information
```

### Feature Request Guidelines

- **Be Clear**: Clearly describe the feature
- **Be Realistic**: Consider implementation complexity
- **Be Open**: Be open to alternative solutions
- **Be Patient**: Feature implementation takes time

## Documentation

### Documentation Contributions

Documentation improvements are always welcome:

- **Fix Typos**: Correct spelling and grammar errors
- **Clarify Content**: Improve unclear explanations
- **Add Examples**: Provide additional code examples
- **Update Content**: Keep documentation current with code changes
- **Add Guides**: Create new guides for features

### Documentation Standards

- **Accuracy**: Ensure technical accuracy
- **Clarity**: Write clear, understandable content
- **Completeness**: Cover all aspects of the topic
- **Examples**: Include working code examples
- **Formatting**: Follow Markdown best practices
- **Links**: Verify all links work correctly

### Documentation Structure

```
docs/
├── README.md              # Documentation index
├── GETTING_STARTED.md     # Quick start guide
├── INSTALLATION.md        # Installation instructions
├── CHANGELOG.md           # Version history
├── CONTRIBUTING.md        # This file
├── guides/                # Feature guides
├── api/                   # API reference
├── architecture/          # Architecture docs
├── examples/              # Example documentation
├── migration/             # Migration guides
└── troubleshooting/       # Troubleshooting guides
```

## Community

### Getting Help

- **Documentation**: Check the [documentation](README.md) first
- **FAQ**: Review the [FAQ](troubleshooting/faq.md)
- **GitHub Issues**: Search existing issues
- **GitHub Discussions**: Ask questions in discussions

### Staying Updated

- **Watch Repository**: Watch the repository for updates
- **Follow Releases**: Subscribe to release notifications
- **Read CHANGELOG**: Review the [CHANGELOG](CHANGELOG.md) for updates

### Communication Channels

- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: Questions and general discussion
- **Pull Requests**: Code contributions and reviews

### Recognition

Contributors will be recognized in:

- **CHANGELOG**: Significant contributions mentioned in release notes
- **Contributors List**: All contributors listed in the repository
- **Release Notes**: Major contributions highlighted in releases

## License

By contributing to Rockstar Web Framework, you agree that your contributions will be licensed under the MIT License.

## Questions?

If you have questions about contributing, please:

1. Check this document thoroughly
2. Search existing issues and discussions
3. Create a new discussion if your question isn't answered

---

**Thank you for contributing to Rockstar Web Framework!** Your contributions help make this project better for everyone.
