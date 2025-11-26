package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

// This example demonstrates configuration management in the Rockstar Web Framework.
// It shows how to load configuration from different formats (JSON, YAML, TOML),
// access configuration values, work with nested configuration, and handle
// environment-specific settings.

func main() {
	fmt.Println("🎸 Rockstar Web Framework - Configuration Management Example\n")

	// Example 1: Load from JSON file
	fmt.Println("1. Loading configuration from JSON file:")
	jsonExample()

	// Example 2: Load from YAML file
	fmt.Println("\n2. Loading configuration from YAML file:")
	yamlExample()

	// Example 3: Load from TOML file
	fmt.Println("\n3. Loading configuration from TOML file:")
	tomlExample()

	// Example 4: Load from environment variables
	fmt.Println("\n4. Loading configuration from environment variables:")
	envExample()

	// Example 5: Environment-specific configuration
	fmt.Println("\n5. Environment-specific configuration:")
	environmentExample()

	// Example 6: Configuration validation
	fmt.Println("\n6. Configuration validation:")
	validationExample()

	// Example 7: Using default values
	fmt.Println("\n7. Using default values:")
	defaultsExample()

	// Example 8: Nested configuration access
	fmt.Println("\n8. Working with nested configuration:")
	nestedExample()

	// Example 9: Configuration watching
	fmt.Println("\n9. Configuration watching (file changes):")
	watchExample()

	fmt.Println("\n✅ Configuration examples completed!")
}

func jsonExample() {
	// Create temporary JSON configuration file
	tmpDir := os.TempDir()
	configPath := filepath.Join(tmpDir, "app.json")

	// JSON configuration with application settings
	jsonContent := `{
  "app_name": "MyRockstarApp",
  "port": 8080,
  "debug": true,
  "timeout": "30s",
  "features": ["auth", "api", "websocket"],
  "database": {
    "host": "localhost",
    "port": 5432,
    "name": "mydb",
    "username": "admin",
    "max_connections": 100
  },
  "cache": {
    "enabled": true,
    "ttl": 3600,
    "max_size": 1000
  }
}`

	err := os.WriteFile(configPath, []byte(jsonContent), 0644)
	if err != nil {
		log.Fatalf("Failed to create config file: %v", err)
	}
	defer os.Remove(configPath)

	// Load configuration from JSON file
	config := pkg.NewConfigManager()
	err = config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Access top-level configuration values
	fmt.Printf("  App Name: %s\n", config.GetString("app_name"))
	fmt.Printf("  Port: %d\n", config.GetInt("port"))
	fmt.Printf("  Debug: %v\n", config.GetBool("debug"))
	fmt.Printf("  Timeout: %v\n", config.GetDuration("timeout"))
	fmt.Printf("  Features: %v\n", config.GetStringSlice("features"))

	// Access nested database configuration
	dbConfig := config.Sub("database")
	fmt.Printf("  Database Host: %s\n", dbConfig.GetString("host"))
	fmt.Printf("  Database Port: %d\n", dbConfig.GetInt("port"))
	fmt.Printf("  Database Name: %s\n", dbConfig.GetString("name"))
	fmt.Printf("  Max Connections: %d\n", dbConfig.GetInt("max_connections"))
}

func tomlExample() {
	// Create temporary TOML configuration file
	tmpDir := os.TempDir()
	configPath := filepath.Join(tmpDir, "app.toml")

	// TOML configuration with nested sections
	tomlContent := `# Application Configuration
app_name = "MyRockstarApp"
port = 8080
debug = true
features = ["auth", "api", "websocket"]

[database]
host = "localhost"
port = 5432
name = "mydb"
max_connections = 100

[server]
timeout = "30s"
max_body_size = 1048576

[server.tls]
enabled = true
cert_file = "/path/to/cert.pem"
key_file = "/path/to/key.pem"

[cache]
enabled = true
ttl = 3600
`

	err := os.WriteFile(configPath, []byte(tomlContent), 0644)
	if err != nil {
		log.Fatalf("Failed to create config file: %v", err)
	}
	defer os.Remove(configPath)

	// Load configuration from TOML file
	config := pkg.NewConfigManager()
	err = config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Access top-level values
	fmt.Printf("  App Name: %s\n", config.GetString("app_name"))
	fmt.Printf("  Port: %d\n", config.GetInt("port"))
	fmt.Printf("  Features: %v\n", config.GetStringSlice("features"))

	// Access nested server configuration
	serverConfig := config.Sub("server")
	fmt.Printf("  Server Timeout: %v\n", serverConfig.GetDuration("timeout"))
	fmt.Printf("  Max Body Size: %d\n", serverConfig.GetInt("max_body_size"))

	// Access deeply nested TLS configuration
	tlsConfig := serverConfig.Sub("tls")
	fmt.Printf("  TLS Enabled: %v\n", tlsConfig.GetBool("enabled"))
	fmt.Printf("  Cert File: %s\n", tlsConfig.GetString("cert_file"))
}

func yamlExample() {
	// Create temporary YAML configuration file
	tmpDir := os.TempDir()
	configPath := filepath.Join(tmpDir, "app.yaml")

	// YAML configuration with hierarchical structure
	yamlContent := `# Application Configuration
app_name: MyRockstarApp
port: 8080
debug: true
timeout: 30s
features:
  - auth
  - api
  - websocket

database:
  host: localhost
  port: 5432
  name: mydb
  username: admin
  max_connections: 100

cache:
  enabled: yes
  ttl: 3600
  max_size: 1000

server:
  read_timeout: 10s
  write_timeout: 10s
`

	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	if err != nil {
		log.Fatalf("Failed to create config file: %v", err)
	}
	defer os.Remove(configPath)

	// Load configuration from YAML file
	config := pkg.NewConfigManager()
	err = config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Access top-level configuration values
	fmt.Printf("  App Name: %s\n", config.GetString("app_name"))
	fmt.Printf("  Port: %d\n", config.GetInt("port"))
	fmt.Printf("  Debug: %v\n", config.GetBool("debug"))
	fmt.Printf("  Features: %v\n", config.GetStringSlice("features"))

	// Access nested cache configuration
	cacheConfig := config.Sub("cache")
	fmt.Printf("  Cache Enabled: %v\n", cacheConfig.GetBool("enabled"))
	fmt.Printf("  Cache TTL: %d seconds\n", cacheConfig.GetInt("ttl"))
	fmt.Printf("  Cache Max Size: %d\n", cacheConfig.GetInt("max_size"))
}

func envExample() {
	// Environment variables provide a way to configure applications
	// without modifying configuration files. This is especially useful
	// for containerized deployments and CI/CD pipelines.

	// Set environment variables with ROCKSTAR_ prefix
	os.Setenv("ROCKSTAR_APP_NAME", "EnvApp")
	os.Setenv("ROCKSTAR_PORT", "9090")
	os.Setenv("ROCKSTAR_DEBUG", "true")
	os.Setenv("ROCKSTAR_DATABASE_HOST", "envhost")
	os.Setenv("ROCKSTAR_DATABASE_PORT", "3306")
	os.Setenv("ROCKSTAR_CACHE_ENABLED", "true")
	os.Setenv("ROCKSTAR_CACHE_TTL", "7200")

	defer func() {
		// Clean up environment variables
		os.Unsetenv("ROCKSTAR_APP_NAME")
		os.Unsetenv("ROCKSTAR_PORT")
		os.Unsetenv("ROCKSTAR_DEBUG")
		os.Unsetenv("ROCKSTAR_DATABASE_HOST")
		os.Unsetenv("ROCKSTAR_DATABASE_PORT")
		os.Unsetenv("ROCKSTAR_CACHE_ENABLED")
		os.Unsetenv("ROCKSTAR_CACHE_TTL")
	}()

	// Load configuration from environment variables
	// Variables with ROCKSTAR_ prefix are automatically loaded
	// Underscores in variable names are converted to dots for nested access
	config := pkg.NewConfigManager()
	err := config.LoadFromEnv()
	if err != nil {
		log.Fatalf("Failed to load from env: %v", err)
	}

	// Access environment-based configuration values
	fmt.Printf("  App Name: %s\n", config.GetString("app.name"))
	fmt.Printf("  Port: %d\n", config.GetInt("port"))
	fmt.Printf("  Debug: %v\n", config.GetBool("debug"))
	fmt.Printf("  Database Host: %s\n", config.GetString("database.host"))
	fmt.Printf("  Database Port: %d\n", config.GetInt("database.port"))
	fmt.Printf("  Cache Enabled: %v\n", config.GetBool("cache.enabled"))
	fmt.Printf("  Cache TTL: %d seconds\n", config.GetInt("cache.ttl"))
}

func environmentExample() {
	// Environment-specific configuration allows different settings
	// for development, staging, and production environments.

	// Set the environment (normally set via ROCKSTAR_ENV or ENV variable)
	os.Setenv("ROCKSTAR_ENV", "production")
	defer os.Unsetenv("ROCKSTAR_ENV")

	config := pkg.NewConfigManager()

	// Check current environment
	fmt.Printf("  Current Environment: %s\n", config.GetEnv())
	fmt.Printf("  Is Production: %v\n", config.IsProduction())
	fmt.Printf("  Is Development: %v\n", config.IsDevelopment())
	fmt.Printf("  Is Test: %v\n", config.IsTest())

	// Load environment-specific configuration
	// In practice, you would load different config files based on environment
	// e.g., config.production.yaml, config.development.yaml
	fmt.Println("  (In production, you would load config.production.yaml)")

	// Change to development environment
	os.Setenv("ROCKSTAR_ENV", "development")
	config2 := pkg.NewConfigManager()
	fmt.Printf("\n  Switched to: %s\n", config2.GetEnv())
	fmt.Printf("  Is Development: %v\n", config2.IsDevelopment())
}

func validationExample() {
	// Configuration validation ensures required settings are present
	// and have valid values before the application starts.

	tmpDir := os.TempDir()
	configPath := filepath.Join(tmpDir, "app.json")

	// Create a valid configuration
	validConfig := `{
  "app_name": "ValidApp",
  "port": 8080,
  "database": {
    "host": "localhost",
    "port": 5432
  }
}`

	err := os.WriteFile(configPath, []byte(validConfig), 0644)
	if err != nil {
		log.Fatalf("Failed to create config file: %v", err)
	}
	defer os.Remove(configPath)

	// Load and validate configuration
	config := pkg.NewConfigManager()
	err = config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Validate configuration
	err = config.Validate()
	if err != nil {
		fmt.Printf("  ❌ Configuration validation failed: %v\n", err)
	} else {
		fmt.Printf("  ✅ Configuration is valid\n")
	}

	// Check if specific keys are set
	fmt.Printf("  Has app_name: %v\n", config.IsSet("app_name"))
	fmt.Printf("  Has database: %v\n", config.IsSet("database"))
	fmt.Printf("  Has redis: %v\n", config.IsSet("redis"))

	// Validate empty configuration
	emptyConfig := pkg.NewConfigManager()
	err = emptyConfig.Validate()
	if err != nil {
		fmt.Printf("  ❌ Empty config validation failed (expected): %v\n", err)
	}
}

func defaultsExample() {
	// Default values provide fallback configuration when values
	// are not explicitly set, making applications more resilient.

	config := pkg.NewConfigManager()

	// Get values with defaults (config is empty, so defaults are used)
	appName := config.GetStringWithDefault("app_name", "DefaultApp")
	port := config.GetIntWithDefault("port", 8080)
	debug := config.GetBoolWithDefault("debug", false)
	timeout := config.GetWithDefault("timeout", 30*time.Second)

	fmt.Printf("  App Name (default): %s\n", appName)
	fmt.Printf("  Port (default): %d\n", port)
	fmt.Printf("  Debug (default): %v\n", debug)
	fmt.Printf("  Timeout (default): %v\n", timeout)

	// Now load some configuration
	tmpDir := os.TempDir()
	configPath := filepath.Join(tmpDir, "app.json")
	configContent := `{"app_name": "ConfiguredApp", "port": 9000}`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		log.Fatalf("Failed to create config file: %v", err)
	}
	defer os.Remove(configPath)

	config2 := pkg.NewConfigManager()
	err = config2.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Get values with defaults (some are set, some use defaults)
	appName2 := config2.GetStringWithDefault("app_name", "DefaultApp")
	port2 := config2.GetIntWithDefault("port", 8080)
	debug2 := config2.GetBoolWithDefault("debug", false)

	fmt.Printf("\n  With config file:\n")
	fmt.Printf("  App Name (from config): %s\n", appName2)
	fmt.Printf("  Port (from config): %d\n", port2)
	fmt.Printf("  Debug (default, not in config): %v\n", debug2)
}

func nestedExample() {
	// Nested configuration allows organizing related settings
	// into hierarchical structures for better maintainability.

	tmpDir := os.TempDir()
	configPath := filepath.Join(tmpDir, "app.yaml")

	// YAML with deeply nested configuration
	yamlContent := `
database:
  host: localhost
  port: 5432
  pool:
    min_size: 5
    max_size: 20
    timeout: 30s
  replicas:
    - host: replica1.example.com
      port: 5432
    - host: replica2.example.com
      port: 5432

cache:
  enabled: true
  ttl: 3600
  redis:
    host: localhost
    port: 6379
    db: 0

server:
  http:
    port: 8080
    timeout: 30s
  https:
    port: 8443
    cert: /path/to/cert.pem
    key: /path/to/key.pem
`

	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	if err != nil {
		log.Fatalf("Failed to create config file: %v", err)
	}
	defer os.Remove(configPath)

	// Load configuration
	config := pkg.NewConfigManager()
	err = config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Access nested configuration using Sub()
	dbConfig := config.Sub("database")
	fmt.Printf("  Database Host: %s\n", dbConfig.GetString("host"))
	fmt.Printf("  Database Port: %d\n", dbConfig.GetInt("port"))

	// Access deeply nested configuration
	poolConfig := dbConfig.Sub("pool")
	fmt.Printf("  Pool Min Size: %d\n", poolConfig.GetInt("min_size"))
	fmt.Printf("  Pool Max Size: %d\n", poolConfig.GetInt("max_size"))
	fmt.Printf("  Pool Timeout: %v\n", poolConfig.GetDuration("timeout"))

	// Access cache configuration
	cacheConfig := config.Sub("cache")
	redisConfig := cacheConfig.Sub("redis")
	fmt.Printf("  Redis Host: %s\n", redisConfig.GetString("host"))
	fmt.Printf("  Redis Port: %d\n", redisConfig.GetInt("port"))

	// Access server configuration
	serverConfig := config.Sub("server")
	httpConfig := serverConfig.Sub("http")
	httpsConfig := serverConfig.Sub("https")
	fmt.Printf("  HTTP Port: %d\n", httpConfig.GetInt("port"))
	fmt.Printf("  HTTPS Port: %d\n", httpsConfig.GetInt("port"))

	// Check if keys are set
	fmt.Printf("\n  Has cache config: %v\n", config.IsSet("cache"))
	fmt.Printf("  Has redis config: %v\n", config.IsSet("cache.redis"))
	fmt.Printf("  Has mongodb config: %v\n", config.IsSet("mongodb"))
}

func watchExample() {
	// Configuration watching allows applications to automatically
	// reload configuration when files change, without requiring a restart.

	tmpDir := os.TempDir()
	configPath := filepath.Join(tmpDir, "app.yaml")

	// Create initial configuration
	initialConfig := `
app_name: WatchedApp
port: 8080
debug: false
`

	err := os.WriteFile(configPath, []byte(initialConfig), 0644)
	if err != nil {
		log.Fatalf("Failed to create config file: %v", err)
	}
	defer os.Remove(configPath)

	// Load configuration
	config := pkg.NewConfigManager()
	err = config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("  Initial app_name: %s\n", config.GetString("app_name"))
	fmt.Printf("  Initial port: %d\n", config.GetInt("port"))
	fmt.Printf("  Initial debug: %v\n", config.GetBool("debug"))

	// Set up configuration watching with callback
	changeCount := 0
	err = config.Watch(func() {
		changeCount++
		fmt.Printf("  📢 Configuration changed! (change #%d)\n", changeCount)
		fmt.Printf("     New app_name: %s\n", config.GetString("app_name"))
		fmt.Printf("     New port: %d\n", config.GetInt("port"))
		fmt.Printf("     New debug: %v\n", config.GetBool("debug"))
	})
	if err != nil {
		log.Fatalf("Failed to watch config: %v", err)
	}

	fmt.Println("\n  👀 Watching for configuration changes...")
	fmt.Println("  (File watching runs in background goroutine)")

	// Simulate configuration change
	time.Sleep(1 * time.Second)
	updatedConfig := `
app_name: UpdatedApp
port: 9090
debug: true
`
	err = os.WriteFile(configPath, []byte(updatedConfig), 0644)
	if err != nil {
		log.Fatalf("Failed to update config file: %v", err)
	}

	// Wait for the watcher to detect the change
	fmt.Println("  ✏️  Updated configuration file...")
	time.Sleep(6 * time.Second) // Watcher checks every 5 seconds

	// Stop watching
	fmt.Println("\n  🛑 Stopping configuration watcher...")
	config.StopWatching()
	fmt.Println("  Configuration watching stopped")
}
