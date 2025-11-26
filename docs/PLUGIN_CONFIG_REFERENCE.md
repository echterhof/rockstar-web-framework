# Plugin Configuration Reference

This document provides a comprehensive reference for all plugin configuration options in the Rockstar Web Framework.

## Table of Contents

- [Overview](#overview)
- [Configuration Formats](#configuration-formats)
- [Global Plugin Settings](#global-plugin-settings)
- [Plugin-Specific Settings](#plugin-specific-settings)
- [Configuration Options Reference](#configuration-options-reference)
- [Permissions Reference](#permissions-reference)
- [Best Practices](#best-practices)
- [Examples](#examples)

## Overview

The plugin system allows you to extend the Rockstar Web Framework with custom functionality through dynamically loadable modules. Plugins can:

- Hook into framework lifecycle events
- Register custom middleware and routes
- Access framework services (database, cache, router, etc.)
- Communicate with other plugins
- Store plugin-specific configuration and data

**Requirements Addressed:**
- **3.1**: Support for YAML, JSON, and TOML configuration formats
- **3.3**: Enabled flag filtering for selective plugin loading
- **3.4**: Initialization parameters passed to plugins
- **3.5**: Load order preservation through configuration sequence

## Configuration Formats

The plugin system supports three configuration formats:

### YAML
```yaml
plugins:
  enabled: true
  plugins:
    - name: my-plugin
      enabled: true
      path: ./plugins/my-plugin
```

**File**: `examples/plugin-config.yaml`

### JSON
```json
{
  "plugins": {
    "enabled": true,
    "plugins": [
      {
        "name": "my-plugin",
        "enabled": true,
        "path": "./plugins/my-plugin"
      }
    ]
  }
}
```

**File**: `examples/plugin-config.json`

### TOML
```toml
[plugins]
enabled = true

[[plugins.plugins]]
name = "my-plugin"
enabled = true
path = "./plugins/my-plugin"
```

**File**: `examples/plugin-config.toml`

## Global Plugin Settings

These settings control the overall behavior of the plugin system:

| Setting | Type | Default | Description | Requirement |
|---------|------|---------|-------------|-------------|
| `enabled` | boolean | `true` | Enable or disable the entire plugin system | 3.3 |
| `directory` | string | `./plugins` | Base directory for plugin discovery | 3.2 |
| `max_concurrent_operations` | integer | `10` | Maximum number of concurrent plugin operations | - |
| `error_threshold` | integer | `100` | Number of errors before auto-disabling a plugin | 9.5 |
| `hot_reload_enabled` | boolean | `true` | Enable hot reload support | 7.1 |
| `load_timeout` | integer | `30` | Plugin load timeout in seconds | - |

### Example

```yaml
plugins:
  enabled: true
  directory: ./plugins
  max_concurrent_operations: 10
  error_threshold: 100
  hot_reload_enabled: true
  load_timeout: 30
```

## Plugin-Specific Settings

Each plugin in the `plugins` array can have the following settings:

### Required Settings

| Setting | Type | Description | Requirement |
|---------|------|-------------|-------------|
| `name` | string | Unique identifier for the plugin | 1.3 |
| `enabled` | boolean | Whether to load this plugin | 3.3 |
| `path` | string | Path to the plugin binary or directory | 3.2 |

### Optional Settings

| Setting | Type | Default | Description | Requirement |
|---------|------|---------|-------------|-------------|
| `priority` | integer | `100` | Hook execution priority (0-1000, higher = earlier) | 2.6 |
| `config` | object | `{}` | Plugin-specific configuration parameters | 3.4, 8.1 |
| `permissions` | object | See below | Framework service access permissions | 11.1 |

## Configuration Options Reference

### Name
**Type**: `string`  
**Required**: Yes  
**Requirement**: 1.3

Unique identifier for the plugin. Used for:
- Plugin registry lookups
- Dependency resolution
- Inter-plugin communication
- Logging and metrics

```yaml
name: "auth-plugin"
```

### Enabled
**Type**: `boolean`  
**Required**: Yes  
**Requirement**: 3.3

Controls whether the plugin is loaded. When `false`, the plugin configuration is kept but the plugin is not loaded.

```yaml
enabled: true  # Plugin will be loaded
enabled: false # Plugin configuration kept but not loaded
```

**Use Cases**:
- Temporarily disable a plugin without removing its configuration
- Environment-specific plugin loading
- A/B testing different plugin configurations

### Path
**Type**: `string`  
**Required**: Yes  
**Requirement**: 3.2

Path to the plugin binary or directory. Can be:
- **Absolute**: `/usr/local/plugins/my-plugin`
- **Relative to `plugins.directory`**: `./my-plugin`
- **Relative to application**: `../plugins/my-plugin`

```yaml
path: "./plugins/auth-plugin"
path: "/usr/local/plugins/auth-plugin"
path: "../shared-plugins/auth-plugin"
```

### Priority
**Type**: `integer`  
**Range**: 0-1000  
**Default**: 100  
**Requirement**: 2.6

Controls hook execution order. Higher priority plugins execute first.

```yaml
priority: 200  # Executes before priority 100
priority: 100  # Default priority
priority: 50   # Executes after priority 100
```

**Common Priority Ranges**:
- **300-1000**: Critical plugins (authentication, security)
- **200-299**: High priority (rate limiting, validation)
- **100-199**: Normal priority (default)
- **50-99**: Low priority (logging, metrics)
- **0-49**: Lowest priority (cleanup, finalization)

### Config
**Type**: `object`  
**Default**: `{}`  
**Requirement**: 3.4, 8.1, 8.4, 8.5

Plugin-specific configuration parameters passed to the plugin's `Initialize()` method.

```yaml
config:
  api_key: "secret-key"
  timeout: "30s"
  enabled_features:
    - feature1
    - feature2
  nested:
    setting: value
```

**Features**:
- Supports nested structures (Requirement 8.4)
- Isolated per plugin (Requirement 8.1)
- Can specify defaults in plugin manifest (Requirement 8.5)
- Triggers `OnConfigChange` callback when updated (Requirement 8.2)

### Permissions
**Type**: `object`  
**Requirement**: 11.1, 11.4

Controls what framework services the plugin can access.

```yaml
permissions:
  database: true
  cache: true
  router: true
  config: true
  filesystem: false
  network: false
  exec: false
  custom:
    custom_permission: true
```

## Permissions Reference

### Standard Permissions

| Permission | Description | Risk Level | Requirement |
|------------|-------------|------------|-------------|
| `database` | Access to DatabaseManager | Medium | 11.4 |
| `cache` | Access to CacheManager | Low | 11.4 |
| `router` | Access to Router (register routes/middleware) | Medium | 11.4 |
| `config` | Access to ConfigManager | Low | 11.4 |
| `filesystem` | Access to filesystem operations | High | 11.4 |
| `network` | Access to network operations | High | 11.4 |
| `exec` | Access to execute external commands | Critical | 11.4 |

### Permission Enforcement

**Requirement 11.2**: The system verifies permissions before allowing access to framework services.

**Requirement 11.3**: Unauthorized access attempts are denied and logged as security violations.

```yaml
# Example: Minimal permissions (safest)
permissions:
  database: false
  cache: false
  router: true
  config: true
  filesystem: false
  network: false
  exec: false

# Example: Full permissions (use with caution)
permissions:
  database: true
  cache: true
  router: true
  config: true
  filesystem: true
  network: true
  exec: false  # Never enable unless absolutely necessary
```

### Custom Permissions

You can define custom permissions for plugin-specific access control:

```yaml
permissions:
  database: true
  custom:
    user_management: true
    role_management: true
    billing_access: false
```

## Best Practices

### 1. Load Order (Requirement 3.5)

Plugins are loaded in the order they appear in the configuration. This is important for:
- Dependency resolution
- Hook execution order
- Initialization sequence

```yaml
plugins:
  plugins:
    - name: base-plugin      # Loaded first
    - name: dependent-plugin # Can depend on base-plugin
    - name: final-plugin     # Loaded last
```

### 2. Principle of Least Privilege

Only grant permissions that a plugin actually needs:

```yaml
# ❌ Bad: Granting unnecessary permissions
permissions:
  database: true
  cache: true
  router: true
  config: true
  filesystem: true  # Not needed
  network: true     # Not needed
  exec: true        # Dangerous!

# ✅ Good: Minimal permissions
permissions:
  database: true
  cache: true
  router: true
  config: true
  filesystem: false
  network: false
  exec: false
```

### 3. Priority Assignment

Use priority to control execution order:

```yaml
# Authentication should run first
- name: auth-plugin
  priority: 200

# Rate limiting after auth
- name: rate-limit-plugin
  priority: 180

# Caching after rate limiting
- name: cache-plugin
  priority: 150

# Logging runs last
- name: logging-plugin
  priority: 50
```

### 4. Configuration Organization

Keep related settings together:

```yaml
config:
  # Authentication settings
  jwt_secret: "secret"
  token_duration: "2h"
  
  # Session settings
  session_cookie_name: "session"
  session_duration: "24h"
  
  # Security settings
  enable_csrf: true
  max_login_attempts: 5
```

### 5. Environment-Specific Configuration

Use enabled flags for environment-specific plugins:

```yaml
# Development only
- name: debug-plugin
  enabled: true  # Set to false in production

# Production only
- name: monitoring-plugin
  enabled: false  # Set to true in production
```

## Examples

### Example 1: Authentication Plugin

```yaml
- name: auth-plugin
  enabled: true
  path: ./plugins/auth-plugin
  priority: 200
  
  config:
    jwt_secret: "your-secret-key"
    token_duration: "2h"
    require_auth: true
    excluded_paths:
      - "/health"
      - "/public"
  
  permissions:
    database: true
    cache: true
    router: true
    config: true
    filesystem: false
    network: false
    exec: false
```

**Addresses Requirements**:
- 3.3: Enabled flag
- 3.4: Initialization parameters (config)
- 11.1: Permission assignment

### Example 2: Logging Plugin

```yaml
- name: logging-plugin
  enabled: true
  path: ./plugins/logging-plugin
  priority: 50
  
  config:
    log_requests: true
    log_responses: true
    output_format: "json"
    mask_headers:
      - "Authorization"
      - "Cookie"
  
  permissions:
    database: false
    cache: false
    router: true
    config: true
    filesystem: true  # For log files
    network: false
    exec: false
```

**Addresses Requirements**:
- 2.6: Priority ordering (low priority for logging)
- 8.4: Nested configuration (mask_headers array)
- 11.4: Granular permissions (filesystem only)

### Example 3: Disabled Plugin

```yaml
- name: experimental-plugin
  enabled: false  # Not loaded
  path: ./plugins/experimental-plugin
  priority: 100
  
  config:
    experimental_feature: true
  
  permissions:
    database: false
    cache: false
    router: false
    config: false
    filesystem: false
    network: false
    exec: false
```

**Addresses Requirements**:
- 3.3: Enabled flag filtering (plugin not loaded)

## Loading Configuration

### From Code

```go
// Load from YAML
err := framework.PluginManager().LoadPluginsFromConfig("config.yaml")

// Load from JSON
err := framework.PluginManager().LoadPluginsFromConfig("config.json")

// Load from TOML
err := framework.PluginManager().LoadPluginsFromConfig("config.toml")
```

### From Command Line

```bash
# Specify config file
./myapp --plugin-config=plugin-config.yaml

# Use default config
./myapp  # Looks for plugin-config.yaml in current directory
```

## Troubleshooting

### Plugin Not Loading

**Symptoms**: Plugin doesn't appear in loaded plugins list, no initialization logs

**Common Causes and Solutions**:

1. **Enabled flag is false**
   - Check: `enabled: true` in plugin configuration
   - Solution: Set `enabled: true` or remove the plugin from config

2. **Incorrect path**
   - Check: Verify the path exists and is accessible
   - Solution: Use absolute path or correct relative path
   - Example: `./plugins/my-plugin` or `/usr/local/plugins/my-plugin`

3. **Invalid plugin manifest**
   - Check: Validate plugin.yaml syntax
   - Solution: Ensure all required fields are present (name, version, description)
   - Use: `go run examples/test_manifest_parser.go plugin.yaml` to validate

4. **Missing dependencies**
   - Check: Review plugin manifest dependencies section
   - Solution: Load dependency plugins first or install missing dependencies
   - Example: If plugin requires `auth-plugin`, ensure it's loaded before

5. **Framework version incompatibility**
   - Check: Plugin manifest `framework.version` field
   - Solution: Update plugin or framework to compatible versions
   - Example: Plugin requires `>=1.0.0,<2.0.0`

6. **Plugin binary not executable**
   - Check: File permissions on plugin binary
   - Solution: `chmod +x plugins/my-plugin/plugin` (Unix/Linux)

7. **Load timeout exceeded**
   - Check: Plugin initialization takes too long
   - Solution: Increase `load_timeout` in global plugin settings
   - Example: `load_timeout: 60` (seconds)

**Debugging Steps**:
```bash
# Enable debug logging
export ROCKSTAR_LOG_LEVEL=debug

# Check plugin directory
ls -la ./plugins/my-plugin

# Validate plugin manifest
go run examples/test_manifest_parser.go ./plugins/my-plugin/plugin.yaml

# Check framework logs
tail -f logs/framework.log | grep plugin
```

### Permission Denied Errors

**Symptoms**: Plugin fails with "permission denied" or "access denied" errors

**Common Causes and Solutions**:

1. **Missing required permissions**
   - Check: Plugin tries to access service without permission
   - Solution: Grant required permission in configuration
   - Example:
     ```yaml
     permissions:
       database: true  # If plugin needs database access
     ```

2. **Filesystem permission denied**
   - Check: Plugin tries to read/write files without filesystem permission
   - Solution: Grant `filesystem: true` permission
   - Warning: Only grant if plugin needs file access

3. **Network permission denied**
   - Check: Plugin tries to make HTTP requests without network permission
   - Solution: Grant `network: true` permission
   - Use case: OAuth providers, external APIs, Redis connections

4. **Custom permission not defined**
   - Check: Plugin requires custom permission not in configuration
   - Solution: Add custom permission to config
   - Example:
     ```yaml
     permissions:
       custom:
         user_management: true
     ```

5. **Security policy violation**
   - Check: Security logs for policy violations
   - Solution: Review and update security policies
   - Location: Check framework security configuration

**Debugging Steps**:
```bash
# Check security logs
grep "permission denied" logs/security.log

# Review plugin permissions
cat plugin-config.yaml | grep -A 10 "permissions:"

# Test with full permissions (temporarily)
# Then reduce to minimum required
```

**Security Best Practices**:
- Start with minimal permissions
- Add permissions only when needed
- Never grant `exec: true` unless absolutely necessary
- Review permissions regularly
- Use custom permissions for fine-grained control

### Load Order Issues

**Symptoms**: Plugin fails because dependency not available, hooks execute in wrong order

**Common Causes and Solutions**:

1. **Dependency loaded after dependent plugin**
   - Check: Plugin order in configuration
   - Solution: Move dependency plugins earlier in the list
   - Example:
     ```yaml
     plugins:
       - name: base-plugin      # Load first
       - name: dependent-plugin # Load after base-plugin
     ```

2. **Circular dependencies**
   - Check: Plugin A depends on B, B depends on A
   - Solution: Refactor to remove circular dependency
   - Alternative: Use event-based communication instead

3. **Priority conflicts**
   - Check: Multiple plugins with same priority
   - Solution: Assign unique priorities or use load order
   - Example:
     ```yaml
     - name: auth-plugin
       priority: 200
     - name: rate-limit-plugin
       priority: 180  # Different priority
     ```

4. **Hook execution order incorrect**
   - Check: Plugin hooks execute in unexpected order
   - Solution: Adjust priority values (higher = earlier)
   - Example: Authentication (200) before rate limiting (180)

5. **Plugin initialization order matters**
   - Check: Plugin B needs plugin A to be fully initialized
   - Solution: Use startup hooks or event subscriptions
   - Alternative: Implement lazy initialization

**Debugging Steps**:
```bash
# Check plugin load order
grep "Loading plugin" logs/framework.log

# Check hook execution order
grep "Executing hook" logs/framework.log

# Visualize dependencies
# Create a dependency graph from manifests
```

**Best Practices**:
- Document plugin dependencies clearly
- Use semantic versioning for dependencies
- Test plugin load order in development
- Use priority ranges consistently (see Priority section)

### Configuration Errors

**Symptoms**: Plugin fails to initialize, configuration validation errors

**Common Causes and Solutions**:

1. **Invalid YAML/JSON/TOML syntax**
   - Check: Configuration file syntax
   - Solution: Validate with linter or parser
   - Tools: `yamllint`, `jsonlint`, or online validators

2. **Missing required configuration**
   - Check: Plugin manifest config schema
   - Solution: Add required configuration values
   - Example:
     ```yaml
     config:
       api_key: "required-value"  # Was missing
     ```

3. **Type mismatch**
   - Check: Configuration value types match schema
   - Solution: Correct value types
   - Example: `timeout: "30s"` not `timeout: 30`

4. **Invalid duration format**
   - Check: Duration strings use correct format
   - Solution: Use Go duration format
   - Valid: `"30s"`, `"5m"`, `"2h"`, `"24h"`
   - Invalid: `"30 seconds"`, `"5 minutes"`

5. **Configuration not reloading**
   - Check: Hot reload enabled
   - Solution: Enable hot reload or restart application
   - Example: `hot_reload_enabled: true`

**Debugging Steps**:
```bash
# Validate YAML syntax
yamllint plugin-config.yaml

# Validate JSON syntax
jsonlint plugin-config.json

# Check configuration loading
grep "Loading configuration" logs/framework.log

# Test configuration parsing
go run examples/test_manifest_parser.go plugin-config.yaml
```

### Performance Issues

**Symptoms**: Slow plugin loading, high memory usage, slow request processing

**Common Causes and Solutions**:

1. **Too many concurrent operations**
   - Check: `max_concurrent_operations` setting
   - Solution: Reduce concurrent operations
   - Example: `max_concurrent_operations: 5`

2. **Plugin initialization timeout**
   - Check: Plugin takes too long to initialize
   - Solution: Optimize plugin initialization or increase timeout
   - Example: `load_timeout: 60`

3. **Memory leaks in plugin**
   - Check: Memory usage grows over time
   - Solution: Review plugin code for leaks
   - Tools: Use pprof for memory profiling

4. **Too many plugins loaded**
   - Check: Number of active plugins
   - Solution: Disable unused plugins
   - Example: Set `enabled: false` for unused plugins

5. **Hook execution overhead**
   - Check: Too many hooks or slow hook handlers
   - Solution: Optimize hook handlers, reduce hook count
   - Profile: Use framework profiling tools

**Debugging Steps**:
```bash
# Check memory usage
ps aux | grep myapp

# Profile memory
curl http://localhost:8080/debug/pprof/heap > heap.prof
go tool pprof heap.prof

# Profile CPU
curl http://localhost:8080/debug/pprof/profile?seconds=30 > cpu.prof
go tool pprof cpu.prof

# Check plugin metrics
curl http://localhost:8080/metrics | grep plugin
```

### Common Error Messages

**"Plugin not found"**
- Cause: Plugin path incorrect or plugin not installed
- Solution: Verify path and install plugin

**"Version mismatch"**
- Cause: Plugin requires different framework version
- Solution: Update plugin or framework

**"Dependency not satisfied"**
- Cause: Required dependency plugin not loaded
- Solution: Load dependency plugin first

**"Permission denied: database"**
- Cause: Plugin tries to access database without permission
- Solution: Add `database: true` to permissions

**"Configuration validation failed"**
- Cause: Required configuration missing or invalid
- Solution: Check plugin manifest for required config

**"Hook registration failed"**
- Cause: Invalid hook type or priority
- Solution: Check hook type and priority values

**"Plugin initialization timeout"**
- Cause: Plugin takes too long to initialize
- Solution: Increase `load_timeout` or optimize plugin

**"Circular dependency detected"**
- Cause: Plugin A depends on B, B depends on A
- Solution: Refactor to remove circular dependency

## See Also

- [Plugin Development Guide](../docs/PLUGIN_DEVELOPMENT.md)
- [Plugin System API](../docs/PLUGIN_SYSTEM.md)
- [Example Plugins](./plugins/README.md)
- [Plugin Manifest Reference](./plugin-manifest-example.yaml)
