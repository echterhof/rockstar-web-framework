# AGENT.md - AI Agent Guide for Rockstar Web Framework

**Version:** 1.0.0  
**Last Updated:** 2025-11-30  
**Framework Status:** 100% Complete, Production-Ready

---

## Purpose of This Document

This document provides critical guidance for AI agents (like GPT-4, Claude, Copilot, etc.) working on the Rockstar Web Framework. It explains the framework's creation philosophy, architectural decisions, and what to pay attention to when making changes or additions.

---

## Framework Creation Philosophy

### Design Principles

1. **Context-Driven Architecture**
   - Everything flows through the `Context` interface
   - No global state - all services accessed via context
   - Enables dependency injection and testability
   - **Critical:** Never bypass context to access framework features

2. **Flat Package Structure**
   - All framework code lives in `pkg/` as a single package
   - No sub-packages or nested modules
   - Simplifies imports and reduces complexity
   - **Critical:** Do not create sub-packages in `pkg/`

3. **Interface + Implementation Pattern**
   - Public interfaces define contracts (e.g., `DatabaseManager`)
   - Private implementations provide functionality (e.g., `databaseImpl`)
   - Enables testing with mocks and alternative implementations
   - **Critical:** Always define interfaces before implementations

4. **Manager Pattern**
   - Framework services exposed as "managers" (DatabaseManager, CacheManager, etc.)
   - Managers are accessed through Context
   - Provides consistent API across all features
   - **Critical:** New features should follow the manager pattern

5. **Pipeline-Based Request Processing**
   - Requests flow through: Middleware → Pipelines → Views
   - Pipelines can chain, multiplex, or terminate early
   - Enables complex request processing without tight coupling
   - **Critical:** Understand pipeline flow before modifying request handling

6. **Multi-Protocol, Multi-API Support**
   - Framework supports HTTP/1, HTTP/2, QUIC, WebSocket
   - Supports REST, GraphQL, gRPC, SOAP
   - Protocol/API choice is configuration-driven
   - **Critical:** Changes must not break protocol/API neutrality

7. **Enterprise-First Features**
   - Multi-tenancy with tenant isolation
   - Security (OAuth2, JWT, RBAC, CSRF/XSS)
   - Internationalization (i18n)
   - Plugin system for extensibility
   - **Critical:** Maintain enterprise-grade quality standards

---

## Critical Architectural Decisions

### 1. Context is King

**Decision:** All framework features are accessed through the `Context` interface.

**Rationale:**
- Eliminates global state
- Enables request-scoped resources (cache, sessions)
- Simplifies testing with mock contexts
- Provides consistent API

**What Agents Must Do:**
- ✅ Always access framework features via `ctx.DB()`, `ctx.Cache()`, etc.
- ✅ Pass context through function chains
- ❌ Never create global variables for framework services
- ❌ Never bypass context to access internal implementations

**Example:**
```go
// ✅ CORRECT
func MyHandler(ctx pkg.Context) error {
    user, err := ctx.DB().QueryOne("SELECT * FROM users WHERE id = ?", ctx.Param("id"))
    return ctx.JSON(200, user)
}

// ❌ WRONG - bypassing context
var globalDB pkg.DatabaseManager
func MyHandler(ctx pkg.Context) error {
    user, err := globalDB.QueryOne(...) // DON'T DO THIS
}
```

### 2. Flat Package Structure

**Decision:** All framework code in `pkg/` as a single package, no sub-packages.

**Rationale:**
- Reduces import complexity
- Prevents circular dependencies
- Simplifies API surface
- Easier to understand and navigate

**What Agents Must Do:**
- ✅ Place all new framework code in `pkg/`
- ✅ Use file naming conventions: `{component}.go`, `{component}_impl.go`, `{component}_test.go`
- ❌ Never create sub-packages like `pkg/database/`, `pkg/router/`
- ❌ Never split related code across multiple packages

**File Organization:**
```
pkg/
├── framework.go           # Main framework struct
├── context.go             # Context interface
├── context_impl.go        # Context implementation
├── database.go            # DatabaseManager interface
├── database_impl.go       # DatabaseManager implementation
├── database_test.go       # DatabaseManager tests
└── ...
```

### 3. Externalized SQL Queries

**Decision:** SQL queries are stored in `sql/{driver}/` directories, not embedded in code.

**Rationale:**
- Supports multiple database drivers (MySQL, PostgreSQL, SQLite, MSSQL)
- Allows driver-specific SQL syntax
- Easier to review and modify queries
- Enables query versioning and migration

**What Agents Must Do:**
- ✅ Store SQL in `sql/{driver}/*.sql` files
- ✅ Create identical query files for each driver with driver-specific syntax
- ✅ Load queries at runtime using `LoadSQLQuery(driver, queryName)`
- ❌ Never hardcode SQL in Go code
- ❌ Never assume a single database driver

**Example:**
```
sql/
├── mysql/
│   ├── create_users_table.sql
│   └── load_user_by_id.sql
├── postgres/
│   ├── create_users_table.sql
│   └── load_user_by_id.sql
└── sqlite/
    ├── create_users_table.sql
    └── load_user_by_id.sql
```

### 4. Plugin System Architecture

**Decision:** Framework is extensible via plugins, not by modifying core code.

**Rationale:**
- Keeps core framework stable
- Allows third-party extensions
- Provides security boundaries

**What Agents Must Do:**
- ✅ Use plugin hooks for extensibility points
- ✅ Document plugin APIs clearly
- ✅ Maintain backward compatibility in plugin interfaces
- ❌ Never require core modifications for new features that could be plugins
- ❌ Never break existing plugin APIs without major version bump

### 5. Multi-Tenancy Design

**Decision:** Framework supports multi-tenancy with host-based routing and tenant isolation.

**Rationale:**
- Enterprise applications often need multi-tenancy
- Host-based routing is most flexible
- Tenant isolation prevents data leaks
- Resource limits prevent tenant abuse

**What Agents Must Do:**
- ✅ Ensure new features respect tenant boundaries
- ✅ Test features in multi-tenant scenarios
- ✅ Document tenant-specific configuration
- ❌ Never share resources across tenants without explicit configuration
- ❌ Never assume single-tenant deployment

### 6. Configuration Management

**Decision:** Configuration via YAML files or environment variables, with defaults.

**Rationale:**
- YAML is human-readable and supports complex structures
- Environment variables enable 12-factor app deployment
- Defaults reduce configuration burden
- Type-safe configuration structs prevent errors

**What Agents Must Do:**
- ✅ Provide sensible defaults for all configuration
- ✅ Support both YAML and environment variable configuration
- ✅ Validate configuration at startup
- ✅ Document all configuration options
- ❌ Never require configuration for basic usage
- ❌ Never use configuration formats other than YAML/ENV (INI/TOML planned but not implemented)

---

## What AI Agents Must Pay Attention To

### 1. Backward Compatibility

**Why It Matters:** Framework is production-ready (v1.0.0). Breaking changes harm users.

**What to Check:**
- ✅ Public interfaces remain unchanged
- ✅ Configuration format stays compatible
- ✅ Plugin APIs maintain compatibility
- ✅ Database schema changes are migrations, not replacements
- ❌ Never remove or rename public methods without deprecation
- ❌ Never change function signatures in public APIs

**Deprecation Process:**
1. Mark old method as deprecated with comment
2. Add new method with improved API
3. Update documentation
4. Remove deprecated method in next major version

### 2. Error Handling

**Why It Matters:** Framework uses i18n for error messages. Errors must be translatable.

**What to Check:**
- ✅ Use `I18nManager` for user-facing errors
- ✅ Provide error keys for translation
- ✅ Include context in error messages
- ✅ Log errors with appropriate severity
- ❌ Never return raw error strings to users
- ❌ Never panic in production code (use recovery)

**Example:**
```go
// ✅ CORRECT
return ctx.I18n().Error("database.connection_failed", map[string]interface{}{
    "driver": "mysql",
    "error": err.Error(),
})

// ❌ WRONG
return fmt.Errorf("database connection failed: %v", err)
```

### 3. Security Considerations

**Why It Matters:** Framework handles sensitive data. Security bugs are critical.

**What to Check:**
- ✅ Validate all user input
- ✅ Use parameterized queries (never string concatenation)
- ✅ Encrypt sensitive data (sessions, cookies)
- ✅ Implement rate limiting for APIs
- ✅ Check authorization before data access
- ❌ Never trust user input
- ❌ Never log sensitive data (passwords, tokens)
- ❌ Never disable security features by default

**Security Checklist:**
- [ ] Input validation implemented?
- [ ] SQL injection prevented?
- [ ] XSS protection enabled?
- [ ] CSRF tokens validated?
- [ ] Rate limiting configured?
- [ ] Authorization checked?
- [ ] Sensitive data encrypted?

### 4. Performance Implications

**Why It Matters:** Framework targets high-performance applications.

**What to Check:**
- ✅ Use arena-based memory management for request-scoped data
- ✅ Pool expensive resources (connections, buffers)
- ✅ Cache frequently accessed data
- ✅ Avoid allocations in hot paths
- ✅ Use goroutines for concurrent operations
- ❌ Never block request handling
- ❌ Never create unbounded goroutines
- ❌ Never ignore memory leaks

**Performance Checklist:**
- [ ] Allocations minimized?
- [ ] Resources pooled?
- [ ] Caching implemented?
- [ ] Goroutines bounded?
- [ ] Benchmarks added?

### 5. Testing Requirements

**Why It Matters:** Framework must be reliable. Tests prevent regressions.

**What to Check:**
- ✅ Unit tests for all public methods
- ✅ Integration tests for complex features
- ✅ Benchmark tests for performance-critical code
- ✅ Test error paths, not just happy paths
- ✅ Mock external dependencies
- ❌ Never commit code without tests
- ❌ Never skip error path testing
- ❌ Never ignore failing tests

**Test Coverage Targets:**
- Unit tests: 80%+ coverage
- Integration tests: All major features
- Benchmarks: Performance-critical paths

### 6. Documentation Standards

**Why It Matters:** Framework is complex. Good docs are essential.

**What to Check:**
- ✅ Godoc comments for all public types/functions
- ✅ Examples in documentation
- ✅ Configuration options documented
- ✅ Migration guides for breaking changes
- ❌ Never add public APIs without documentation
- ❌ Never assume users understand internal details

**Documentation Checklist:**
- [ ] Godoc comments added?
- [ ] Examples provided?
- [ ] Configuration documented?
- [ ] Migration guide written (if breaking)?

### 7. Platform Compatibility

**Why It Matters:** Framework supports AIX, Unix, Windows.

**What to Check:**
- ✅ Use platform-specific files: `{file}_windows.go`, `{file}_unix.go`
- ✅ Test on all supported platforms
- ✅ Use `filepath` package for paths (not hardcoded `/` or `\`)
- ✅ Handle platform-specific features gracefully
- ❌ Never assume Unix-only features
- ❌ Never hardcode platform-specific paths
- ❌ Never use platform-specific syscalls without fallbacks

**Platform-Specific Files:**
```
pkg/
├── listener.go           # Platform-agnostic interface
├── listener_unix.go      # Unix/Linux implementation
├── listener_windows.go   # Windows implementation
└── listener_aix.go       # AIX implementation
```

---

## Common Pitfalls to Avoid

### 1. Breaking the Context Pattern

**Pitfall:** Creating global variables or singletons for framework services.

**Why It's Bad:**
- Breaks testability
- Creates hidden dependencies
- Prevents request-scoped resources
- Makes concurrent requests unsafe

**How to Avoid:**
- Always access services via `Context`
- Pass context through function chains
- Never store context in structs (it's request-scoped)

### 2. Ignoring Multi-Tenancy

**Pitfall:** Assuming single-tenant deployment.

**Why It's Bad:**
- Data leaks between tenants
- Resource exhaustion by one tenant
- Security vulnerabilities

**How to Avoid:**
- Always check tenant context
- Isolate resources per tenant
- Test with multiple tenants

### 3. Hardcoding Configuration

**Pitfall:** Using hardcoded values instead of configuration.

**Why It's Bad:**
- Reduces flexibility
- Requires code changes for deployment
- Breaks 12-factor app principles

**How to Avoid:**
- Use `ConfigManager` for all settings
- Provide sensible defaults
- Document configuration options

### 4. Blocking Request Handling

**Pitfall:** Performing slow operations in request handlers.

**Why It's Bad:**
- Reduces throughput
- Causes timeouts
- Wastes resources

**How to Avoid:**
- Use goroutines for slow operations
- Implement timeouts
- Use async pipelines

### 5. Ignoring Error Paths

**Pitfall:** Only testing happy paths.

**Why It's Bad:**
- Production errors are unhandled
- Security vulnerabilities
- Poor user experience

**How to Avoid:**
- Test all error conditions
- Use table-driven tests
- Check error messages

### 6. Skipping Documentation

**Pitfall:** Adding features without documentation.

**Why It's Bad:**
- Users can't discover features
- Increases support burden
- Reduces adoption

**How to Avoid:**
- Write docs before code
- Include examples
- Update API reference

---

## Framework-Specific Conventions

### Naming Conventions

**Interfaces:**
- Use descriptive names: `DatabaseManager`, `RouterEngine`, `CacheManager`
- Suffix with `Manager` for service interfaces
- Suffix with `Engine` for processing interfaces

**Implementations:**
- Suffix with `Impl`: `databaseImpl`, `routerImpl`
- Or prefix with `default`: `defaultDatabase`, `defaultRouter`
- Keep implementations private (lowercase first letter)

**Files:**
- Interface: `{component}.go`
- Implementation: `{component}_impl.go`
- Tests: `{component}_test.go`
- Platform-specific: `{component}_{os}.go`

**Functions:**
- Middleware: `func MyMiddleware(ctx Context) error`
- Pipelines: `func MyPipeline(ctx Context) PipelineResult`
- Views: `func MyView(ctx Context) error`
- Handlers: `func MyHandler(ctx Context) error`

### Code Organization

**File Size:**
- Keep files under 1000 lines
- Split large files by functionality
- Use `{component}_impl.go`, `{component}_helpers.go`, etc.

**Function Size:**
- Keep functions under 50 lines
- Extract complex logic to helper functions
- Use descriptive names for helpers

**Comment Style:**
- Godoc comments for all public types/functions
- Inline comments for complex logic
- TODO comments for future work

### Testing Conventions

**Test Files:**
- One test file per source file
- Use table-driven tests for multiple cases
- Use subtests for related tests

**Test Names:**
- `TestComponentMethod` for unit tests
- `TestComponentMethod_ErrorCase` for error tests
- `BenchmarkComponentMethod` for benchmarks

**Test Structure:**
```go
func TestDatabaseQuery(t *testing.T) {
    tests := []struct {
        name    string
        query   string
        args    []interface{}
        want    interface{}
        wantErr bool
    }{
        // test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test logic
        })
    }
}
```

---

## Implementation Status

**Current Status:** 100% Complete (80/80 features) ✅

### All Features Implemented ✅
- Context-based architecture
- Multi-protocol support (HTTP/1, HTTP/2, QUIC)
- Multi-API support (REST, GraphQL, gRPC, SOAP)
- Database support (MySQL, PostgreSQL, SQLite, MSSQL)
- Security (OAuth2, JWT, RBAC, CSRF/XSS)
- Multi-tenancy with tenant isolation
- Plugin system with hot reload
- Internationalization (i18n)
- Session management
- Caching
- Monitoring and metrics
- Proxy with circuit breakers
- WebSocket support
- Configuration parsers (YAML ✅, ENV ✅, INI ✅, TOML ✅)
- Workload prediction with forecasting algorithms

### Priority for AI Agents

**High Priority:**
1. Maintain backward compatibility (critical for v1.0.0)
2. Fix security issues immediately
3. Address bug reports promptly

**Medium Priority:**
1. Improve test coverage (target: 90%+)
2. Performance optimizations
3. Documentation improvements

**Low Priority:**
1. Code refactoring for maintainability
2. Additional examples
3. Community contributions

---

## When Making Changes

### Before You Start

1. **Read the relevant documentation:**
   - `docs/GETTING_STARTED.md` - Understand basic usage
   - `docs/guides/{feature}.md` - Understand feature details
   - `docs/api/{component}.md` - Understand API contracts

2. **Check existing code:**
   - Look for similar implementations
   - Follow existing patterns
   - Maintain consistency

3. **Understand the impact:**
   - Will this break existing code?
   - Does this affect performance?
   - Are there security implications?

### While You Work

1. **Follow the patterns:**
   - Use Context for all framework access
   - Follow interface + implementation pattern
   - Use manager pattern for services

2. **Write tests first:**
   - Define expected behavior
   - Test error cases
   - Add benchmarks for performance-critical code

3. **Document as you go:**
   - Add Godoc comments
   - Update API reference
   - Add examples

### After You Finish

1. **Run all checks:**
   ```bash
   make check  # Runs fmt, vet, test
   ```

2. **Verify documentation:**
   - Godoc comments complete?
   - Examples working?
   - API reference updated?

3. **Test on all platforms:**
   - Windows
   - Linux/Unix
   - AIX (if applicable)

4. **Check backward compatibility:**
   - No breaking changes?
   - Deprecation warnings added?
   - Migration guide written?

---

## Getting Help

### Documentation
- `docs/` - Comprehensive documentation
- `examples/` - Working code examples
- `IMPLEMENTATION_STATUS.md` - Feature completion status

### Code References
- `pkg/framework.go` - Main framework struct
- `pkg/context.go` - Context interface (the heart of the framework)
- `pkg/*_impl.go` - Implementation examples

### Testing
- `pkg/*_test.go` - Unit test examples
- `tests/*_test.go` - Integration test examples
- `tests/benchmark_test.go` - Benchmark examples

---

## Maintenance Mode

**Status:** The framework is now in **maintenance mode** following 100% feature completion.

### What This Means

1. **Feature Freeze:** No new features will be added to the core framework
2. **Stability Focus:** Emphasis on bug fixes, security, and performance
3. **Plugin-Based Extensions:** New functionality should be implemented as plugins
4. **Breaking Changes:** Require major version bump (v2.0.0)

### Allowed Changes

✅ **Permitted:**
- Bug fixes
- Security patches
- Performance optimizations
- Documentation improvements
- Test coverage improvements
- Dependency updates
- Plugin development

❌ **Not Permitted:**
- New core features
- Breaking API changes (without major version bump)
- Architectural changes
- Removal of existing features

---

## Final Notes for AI Agents

### Remember

1. **Context is everything** - All framework features flow through Context
2. **Flat is better than nested** - Keep everything in `pkg/`
3. **Interfaces define contracts** - Implementation details are private
4. **Security is not optional** - Validate, encrypt, authorize
5. **Performance matters** - This is a high-performance framework
6. **Tests are required** - No code without tests
7. **Documentation is essential** - No public API without docs
8. **Backward compatibility is sacred** - Don't break existing code
9. **Feature complete** - Framework is in maintenance mode, focus on stability

### When in Doubt

1. Look at existing code for patterns
2. Check documentation for guidance
3. Ask for clarification before making breaking changes
4. Err on the side of caution with security
5. Write tests to verify your understanding

### Success Criteria

Your changes are successful when:
- ✅ All tests pass
- ✅ Documentation is complete
- ✅ No breaking changes (or properly deprecated)
- ✅ Security is maintained
- ✅ Performance is not degraded
- ✅ Code follows framework patterns
- ✅ Platform compatibility is maintained
- ✅ Changes align with maintenance mode (bug fixes, not new features)

---

## Framework Completion

**🎉 The Rockstar Web Framework is 100% feature complete!**

All 80 planned features have been implemented, tested, and documented. The framework is production-ready and enters maintenance mode as of November 30, 2025.

### What Was Achieved

- ✅ Complete multi-protocol web framework (HTTP/1, HTTP/2, QUIC)
- ✅ Multi-API support (REST, GraphQL, gRPC, SOAP)
- ✅ Enterprise features (multi-tenancy, security, monitoring)
- ✅ Plugin system for extensibility
- ✅ Comprehensive documentation (100+ examples)
- ✅ 85%+ test coverage
- ✅ Production-ready performance

### Moving Forward

The framework is now stable and ready for production use. Future work focuses on:
- Maintaining stability and security
- Supporting the user community
- Developing plugins for extended functionality
- Keeping dependencies current

---

# Estimated Effort to Produce the Framework

This illustrates one way to estimate total engineering time using the repository's current size.

## Line counts

* **Go source (implementation, tests, examples, plugins):** > 96,200 lines
* **Entire repository (code + docs + SQL + configs):** > 174,400 lines

## Example time calculation

Assuming an average of **1 minute per line of Go code** (writing, debugging, reviews included):

* Total minutes: `96,217 lines × 1 min/line = 96,217 minutes`
* Total hours: `96,217 ÷ 60 ≈ 1,603.6 hours`
* 40-hour work weeks: `1,603.6 ÷ 40 ≈ 40.1 weeks`

Alternate speeds for comparison:

| Minutes/line | Hours | 40-hr weeks |
| --- | --- | --- |
| 0.5 | ≈ 801.8 | ≈ 20.0 |
| 1.0 | ≈ 1,603.6 | ≈ 40.1 |
| 2.0 | ≈ 3,207.2 | ≈ 80.2 |

---

**Thank you for contributing to the Rockstar Web Framework!**

This framework represents a significant effort to provide a complete, enterprise-grade web framework for Go. Your careful attention to these guidelines helps maintain the quality and consistency that makes this framework valuable to its users.

**See [FEATURE_COMPLETE.md](FEATURE_COMPLETE.md) for detailed completion status.**
