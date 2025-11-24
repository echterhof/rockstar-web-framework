package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
	fmt.Println("=== Rockstar Web Framework - Configuration Management Example ===\n")

	// Example 1: Load from INI file
	fmt.Println("1. Loading configuration from INI file:")
	iniExample()

	// Example 2: Load from TOML file
	fmt.Println("\n2. Loading configuration from TOML file:")
	tomlExample()

	// Example 3: Load from YAML file
	fmt.Println("\n3. Loading configuration from YAML file:")
	yamlExample()

	// Example 4: Load from environment variables
	fmt.Println("\n4. Loading configuration from environment variables:")
	envExample()

	// Example 5: Using default values
	fmt.Println("\n5. Using default values:")
	defaultsExample()

	// Example 6: Nested configuration
	fmt.Println("\n6. Working with nested configuration:")
	nestedExample()

	// Example 7: Configuration watching
	fmt.Println("\n7. Configuration watching (file changes):")
	watchExample()
}

func iniExample() {
	// Create temporary INI file
	tmpDir := os.TempDir()
	configPath := filepath.Join(tmpDir, "app.ini")

	iniContent := `
# Application Configuration
app_name = MyRockstarApp
port = 8080
debug = true
timeout = 30s

[database]
host = localhost
port = 5432
name = mydb
username = admin
password = secret

[cache]
enabled = true
ttl = 3600
max_size = 1000
`

	err := os.WriteFile(configPath, []byte(iniContent), 0644)
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

	// Access configuration values
	fmt.Printf("  App Name: %s\n", config.GetString("app_name"))
	fmt.Printf("  Port: %d\n", config.GetInt("port"))
	fmt.Printf("  Debug: %v\n", config.GetBool("debug"))
	fmt.Printf("  Timeout: %v\n", config.GetDuration("timeout"))

	// Access nested values
	dbConfig := config.Sub("database")
	fmt.Printf("  Database Host: %s\n", dbConfig.GetString("host"))
	fmt.Printf("  Database Port: %d\n", dbConfig.GetInt("port"))
	fmt.Printf("  Database Name: %s\n", dbConfig.GetString("name"))
}

func tomlExample() {
	// Create temporary TOML file
	tmpDir := os.TempDir()
	configPath := filepath.Join(tmpDir, "app.toml")

	tomlContent := `
# Application Configuration
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
`

	err := os.WriteFile(configPath, []byte(tomlContent), 0644)
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

	// Access configuration values
	fmt.Printf("  App Name: %s\n", config.GetString("app_name"))
	fmt.Printf("  Features: %v\n", config.GetStringSlice("features"))

	// Access deeply nested values
	serverConfig := config.Sub("server")
	tlsConfig := serverConfig.Sub("tls")
	fmt.Printf("  TLS Enabled: %v\n", tlsConfig.GetBool("enabled"))
	fmt.Printf("  Cert File: %s\n", tlsConfig.GetString("cert_file"))
}

func yamlExample() {
	// Create temporary YAML file
	tmpDir := os.TempDir()
	configPath := filepath.Join(tmpDir, "app.yaml")

	yamlContent := `
# Application Configuration
app_name: MyRockstarApp
port: 8080
debug: true
timeout: 30s

database:
  host: localhost
  port: 5432
  name: mydb
  max_connections: 100

cache:
  enabled: yes
  ttl: 3600
  max_size: 1000
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

	// Access configuration values
	fmt.Printf("  App Name: %s\n", config.GetString("app_name"))
	fmt.Printf("  Port: %d\n", config.GetInt("port"))

	// Access nested values
	cacheConfig := config.Sub("cache")
	fmt.Printf("  Cache Enabled: %v\n", cacheConfig.GetBool("enabled"))
	fmt.Printf("  Cache TTL: %d\n", cacheConfig.GetInt("ttl"))
}

func envExample() {
	// Set environment variables
	os.Setenv("ROCKSTAR_APP_NAME", "EnvApp")
	os.Setenv("ROCKSTAR_PORT", "9090")
	os.Setenv("ROCKSTAR_DEBUG", "true")
	os.Setenv("ROCKSTAR_DATABASE_HOST", "envhost")
	os.Setenv("ROCKSTAR_DATABASE_PORT", "3306")

	defer func() {
		os.Unsetenv("ROCKSTAR_APP_NAME")
		os.Unsetenv("ROCKSTAR_PORT")
		os.Unsetenv("ROCKSTAR_DEBUG")
		os.Unsetenv("ROCKSTAR_DATABASE_HOST")
		os.Unsetenv("ROCKSTAR_DATABASE_PORT")
	}()

	// Load from environment
	config := pkg.NewConfigManager()
	err := config.LoadFromEnv()
	if err != nil {
		log.Fatalf("Failed to load from env: %v", err)
	}

	// Access values
	fmt.Printf("  App Name: %s\n", config.GetString("app.name"))
	fmt.Printf("  Port: %d\n", config.GetInt("port"))
	fmt.Printf("  Debug: %v\n", config.GetBool("debug"))
	fmt.Printf("  Database Host: %s\n", config.GetString("database.host"))
	fmt.Printf("  Database Port: %d\n", config.GetInt("database.port"))
}

func defaultsExample() {
	config := pkg.NewConfigManager()

	// Get values with defaults (config is empty)
	appName := config.GetStringWithDefault("app_name", "DefaultApp")
	port := config.GetIntWithDefault("port", 8080)
	debug := config.GetBoolWithDefault("debug", false)

	fmt.Printf("  App Name (default): %s\n", appName)
	fmt.Printf("  Port (default): %d\n", port)
	fmt.Printf("  Debug (default): %v\n", debug)
}

func nestedExample() {
	// Create temporary config file
	tmpDir := os.TempDir()
	configPath := filepath.Join(tmpDir, "app.ini")

	iniContent := `
[database]
host = localhost
port = 5432

[database.pool]
min_size = 5
max_size = 20

[cache]
enabled = true
ttl = 3600
`

	err := os.WriteFile(configPath, []byte(iniContent), 0644)
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

	// Access nested configuration
	dbConfig := config.Sub("database")
	fmt.Printf("  Database Host: %s\n", dbConfig.GetString("host"))
	fmt.Printf("  Database Port: %d\n", dbConfig.GetInt("port"))

	// Check if keys are set
	fmt.Printf("  Has cache config: %v\n", config.IsSet("cache"))
	fmt.Printf("  Has redis config: %v\n", config.IsSet("redis"))
}

func watchExample() {
	// Create temporary config file
	tmpDir := os.TempDir()
	configPath := filepath.Join(tmpDir, "app.ini")

	iniContent := `
value = initial
`

	err := os.WriteFile(configPath, []byte(iniContent), 0644)
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

	fmt.Printf("  Initial value: %s\n", config.GetString("value"))

	// Set up watching (in real app, this would run continuously)
	err = config.Watch(func() {
		fmt.Println("  Configuration changed!")
	})
	if err != nil {
		log.Fatalf("Failed to watch config: %v", err)
	}

	fmt.Println("  Watching for changes... (in production, this runs continuously)")
	fmt.Println("  Note: File watching runs in background goroutine")

	// Stop watching
	config.StopWatching()
}
