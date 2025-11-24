package pkg

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestConfigManager_LoadINI(t *testing.T) {
	// Create temporary INI file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.ini")

	iniContent := `
# Root level config
app_name = MyApp
port = 8080
debug = true
timeout = 30s

[database]
host = localhost
port = 5432
name = mydb
max_connections = 100

[cache]
enabled = true
ttl = 3600
`

	err := os.WriteFile(configPath, []byte(iniContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Test loading
	config := NewConfigManager()
	err = config.Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test root level values
	if config.GetString("app_name") != "MyApp" {
		t.Errorf("Expected app_name to be 'MyApp', got '%s'", config.GetString("app_name"))
	}

	if config.GetInt("port") != 8080 {
		t.Errorf("Expected port to be 8080, got %d", config.GetInt("port"))
	}

	if !config.GetBool("debug") {
		t.Error("Expected debug to be true")
	}

	if config.GetDuration("timeout") != 30*time.Second {
		t.Errorf("Expected timeout to be 30s, got %v", config.GetDuration("timeout"))
	}

	// Test nested values
	dbConfig := config.Sub("database")
	if dbConfig.GetString("host") != "localhost" {
		t.Errorf("Expected database.host to be 'localhost', got '%s'", dbConfig.GetString("host"))
	}

	if dbConfig.GetInt("port") != 5432 {
		t.Errorf("Expected database.port to be 5432, got %d", dbConfig.GetInt("port"))
	}

	if dbConfig.GetInt("max_connections") != 100 {
		t.Errorf("Expected database.max_connections to be 100, got %d", dbConfig.GetInt("max_connections"))
	}

	// Test cache section
	cacheConfig := config.Sub("cache")
	if !cacheConfig.GetBool("enabled") {
		t.Error("Expected cache.enabled to be true")
	}

	if cacheConfig.GetInt("ttl") != 3600 {
		t.Errorf("Expected cache.ttl to be 3600, got %d", cacheConfig.GetInt("ttl"))
	}
}

func TestConfigManager_LoadTOML(t *testing.T) {
	// Create temporary TOML file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	tomlContent := `
# Root level config
app_name = "MyApp"
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
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Test loading
	config := NewConfigManager()
	err = config.Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test root level values
	if config.GetString("app_name") != "MyApp" {
		t.Errorf("Expected app_name to be 'MyApp', got '%s'", config.GetString("app_name"))
	}

	if config.GetInt("port") != 8080 {
		t.Errorf("Expected port to be 8080, got %d", config.GetInt("port"))
	}

	if !config.GetBool("debug") {
		t.Error("Expected debug to be true")
	}

	// Test array
	features := config.GetStringSlice("features")
	if len(features) != 3 {
		t.Errorf("Expected 3 features, got %d", len(features))
	}

	// Test nested values
	dbConfig := config.Sub("database")
	if dbConfig.GetString("host") != "localhost" {
		t.Errorf("Expected database.host to be 'localhost', got '%s'", dbConfig.GetString("host"))
	}

	// Test deeply nested values
	serverConfig := config.Sub("server")
	tlsConfig := serverConfig.Sub("tls")
	if !tlsConfig.GetBool("enabled") {
		t.Error("Expected server.tls.enabled to be true")
	}

	if tlsConfig.GetString("cert_file") != "/path/to/cert.pem" {
		t.Errorf("Expected cert_file, got '%s'", tlsConfig.GetString("cert_file"))
	}
}

func TestConfigManager_LoadYAML(t *testing.T) {
	// Create temporary YAML file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	yamlContent := `
# Root level config
app_name: MyApp
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
`

	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Test loading
	config := NewConfigManager()
	err = config.Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test root level values
	if config.GetString("app_name") != "MyApp" {
		t.Errorf("Expected app_name to be 'MyApp', got '%s'", config.GetString("app_name"))
	}

	if config.GetInt("port") != 8080 {
		t.Errorf("Expected port to be 8080, got %d", config.GetInt("port"))
	}

	if !config.GetBool("debug") {
		t.Error("Expected debug to be true")
	}

	// Test nested values
	dbConfig := config.Sub("database")
	if dbConfig.GetString("host") != "localhost" {
		t.Errorf("Expected database.host to be 'localhost', got '%s'", dbConfig.GetString("host"))
	}

	if dbConfig.GetInt("port") != 5432 {
		t.Errorf("Expected database.port to be 5432, got %d", dbConfig.GetInt("port"))
	}

	// Test cache section
	cacheConfig := config.Sub("cache")
	if !cacheConfig.GetBool("enabled") {
		t.Error("Expected cache.enabled to be true")
	}
}

func TestConfigManager_LoadFromEnv(t *testing.T) {
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

	// Test loading from environment
	config := NewConfigManager()
	err := config.LoadFromEnv()
	if err != nil {
		t.Fatalf("Failed to load from env: %v", err)
	}

	// Test values - environment variables are stored with dot notation
	appName := config.GetString("app.name")
	if appName != "EnvApp" {
		t.Logf("Available keys: %v", config.(*configManager).data)
		t.Errorf("Expected app.name to be 'EnvApp', got '%s'", appName)
	}

	port := config.GetInt("port")
	if port != 9090 {
		t.Errorf("Expected port to be 9090, got %d", port)
	}

	debug := config.GetBool("debug")
	if !debug {
		t.Error("Expected debug to be true")
	}

	dbHost := config.GetString("database.host")
	if dbHost != "envhost" {
		t.Errorf("Expected database.host to be 'envhost', got '%s'", dbHost)
	}

	dbPort := config.GetInt("database.port")
	if dbPort != 3306 {
		t.Errorf("Expected database.port to be 3306, got %d", dbPort)
	}
}

func TestConfigManager_GetWithDefault(t *testing.T) {
	config := NewConfigManager()

	// Test with non-existent key
	if config.GetStringWithDefault("missing", "default") != "default" {
		t.Error("Expected default value for missing key")
	}

	if config.GetIntWithDefault("missing", 42) != 42 {
		t.Error("Expected default int value for missing key")
	}

	if config.GetBoolWithDefault("missing", true) != true {
		t.Error("Expected default bool value for missing key")
	}
}

func TestConfigManager_TypeConversions(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.ini")

	iniContent := `
string_val = hello
int_val = 42
float_val = 3.14
bool_val = true
duration_val = 5m
`

	err := os.WriteFile(configPath, []byte(iniContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	config := NewConfigManager()
	err = config.Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test string
	if config.GetString("string_val") != "hello" {
		t.Errorf("Expected 'hello', got '%s'", config.GetString("string_val"))
	}

	// Test int
	if config.GetInt("int_val") != 42 {
		t.Errorf("Expected 42, got %d", config.GetInt("int_val"))
	}

	// Test int64
	if config.GetInt64("int_val") != 42 {
		t.Errorf("Expected 42, got %d", config.GetInt64("int_val"))
	}

	// Test float64
	if config.GetFloat64("float_val") != 3.14 {
		t.Errorf("Expected 3.14, got %f", config.GetFloat64("float_val"))
	}

	// Test bool
	if !config.GetBool("bool_val") {
		t.Error("Expected true")
	}

	// Test duration
	if config.GetDuration("duration_val") != 5*time.Minute {
		t.Errorf("Expected 5m, got %v", config.GetDuration("duration_val"))
	}
}

func TestConfigManager_IsSet(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.ini")

	iniContent := `
existing_key = value
`

	err := os.WriteFile(configPath, []byte(iniContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	config := NewConfigManager()
	err = config.Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test existing key
	if !config.IsSet("existing_key") {
		t.Error("Expected existing_key to be set")
	}

	// Test non-existent key
	if config.IsSet("missing_key") {
		t.Error("Expected missing_key to not be set")
	}
}

func TestConfigManager_Environment(t *testing.T) {
	// Test default environment
	config := NewConfigManager()

	if config.GetEnv() == "" {
		t.Error("Expected environment to be set")
	}

	// Test environment checks
	os.Setenv("ROCKSTAR_ENV", "production")
	defer os.Unsetenv("ROCKSTAR_ENV")

	config2 := NewConfigManager()
	if !config2.IsProduction() {
		t.Error("Expected IsProduction to be true")
	}

	if config2.IsDevelopment() {
		t.Error("Expected IsDevelopment to be false")
	}

	if config2.IsTest() {
		t.Error("Expected IsTest to be false")
	}
}

func TestConfigManager_Validate(t *testing.T) {
	// Test empty config
	config := NewConfigManager()
	err := config.Validate()
	if err == nil {
		t.Error("Expected validation error for empty config")
	}

	// Test with data
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.ini")

	iniContent := `
key = value
`

	err = os.WriteFile(configPath, []byte(iniContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	err = config.Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	err = config.Validate()
	if err != nil {
		t.Errorf("Expected validation to pass, got error: %v", err)
	}
}

func TestConfigManager_Reload(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.ini")

	// Write initial config
	iniContent := `
value = initial
`

	err := os.WriteFile(configPath, []byte(iniContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	config := NewConfigManager()
	err = config.Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if config.GetString("value") != "initial" {
		t.Errorf("Expected 'initial', got '%s'", config.GetString("value"))
	}

	// Update config file
	iniContent = `
value = updated
`

	err = os.WriteFile(configPath, []byte(iniContent), 0644)
	if err != nil {
		t.Fatalf("Failed to update test config file: %v", err)
	}

	// Reload
	err = config.Reload()
	if err != nil {
		t.Fatalf("Failed to reload config: %v", err)
	}

	if config.GetString("value") != "updated" {
		t.Errorf("Expected 'updated', got '%s'", config.GetString("value"))
	}
}

func TestConfigManager_Sub(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.ini")

	iniContent := `
[database]
host = localhost
port = 5432

[cache]
enabled = true
`

	err := os.WriteFile(configPath, []byte(iniContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	config := NewConfigManager()
	err = config.Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test sub-config
	dbConfig := config.Sub("database")
	if dbConfig.GetString("host") != "localhost" {
		t.Errorf("Expected 'localhost', got '%s'", dbConfig.GetString("host"))
	}

	if dbConfig.GetInt("port") != 5432 {
		t.Errorf("Expected 5432, got %d", dbConfig.GetInt("port"))
	}

	// Test non-existent sub-config
	missingConfig := config.Sub("missing")
	if missingConfig.GetString("anything") != "" {
		t.Error("Expected empty string for missing sub-config")
	}
}

func TestConfigManager_ConcurrentAccess(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.ini")

	iniContent := `
value = test
count = 42
`

	err := os.WriteFile(configPath, []byte(iniContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	config := NewConfigManager()
	err = config.Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test concurrent reads
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_ = config.GetString("value")
				_ = config.GetInt("count")
				_ = config.IsSet("value")
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}
