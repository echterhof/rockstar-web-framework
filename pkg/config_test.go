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

// Property-Based Tests

// **Feature: todo-implementations, Property 13: Multi-format configuration loading**
// **Validates: Requirements 4.1**
func TestProperty_MultiFormatConfigurationLoading(t *testing.T) {
	// Property: For any valid configuration file in JSON, YAML, or TOML format,
	// the ConfigManager should successfully parse and load the configuration

	testCases := []struct {
		format  string
		ext     string
		content string
	}{
		{
			format: "JSON",
			ext:    ".json",
			content: `{
				"app_name": "TestApp",
				"port": 8080,
				"debug": true,
				"timeout": "30s",
				"database": {
					"host": "localhost",
					"port": 5432,
					"name": "testdb"
				}
			}`,
		},
		{
			format: "YAML",
			ext:    ".yaml",
			content: `
app_name: TestApp
port: 8080
debug: true
timeout: 30s
database:
  host: localhost
  port: 5432
  name: testdb
`,
		},
		{
			format: "TOML",
			ext:    ".toml",
			content: `
app_name = "TestApp"
port = 8080
debug = true
timeout = "30s"

[database]
host = "localhost"
port = 5432
name = "testdb"
`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.format, func(t *testing.T) {
			// Create temporary config file
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config"+tc.ext)

			err := os.WriteFile(configPath, []byte(tc.content), 0644)
			if err != nil {
				t.Fatalf("Failed to create test config file: %v", err)
			}

			// Load configuration
			config := NewConfigManager()
			err = config.Load(configPath)
			if err != nil {
				t.Fatalf("Failed to load %s config: %v", tc.format, err)
			}

			// Verify that configuration was loaded successfully
			// All formats should produce the same logical structure
			if config.GetString("app_name") != "TestApp" {
				t.Errorf("%s: Expected app_name to be 'TestApp', got '%s'", tc.format, config.GetString("app_name"))
			}

			if config.GetInt("port") != 8080 {
				t.Errorf("%s: Expected port to be 8080, got %d", tc.format, config.GetInt("port"))
			}

			if !config.GetBool("debug") {
				t.Errorf("%s: Expected debug to be true", tc.format)
			}

			if config.GetDuration("timeout") != 30*time.Second {
				t.Errorf("%s: Expected timeout to be 30s, got %v", tc.format, config.GetDuration("timeout"))
			}

			// Verify nested configuration
			dbConfig := config.Sub("database")
			if dbConfig.GetString("host") != "localhost" {
				t.Errorf("%s: Expected database.host to be 'localhost', got '%s'", tc.format, dbConfig.GetString("host"))
			}

			if dbConfig.GetInt("port") != 5432 {
				t.Errorf("%s: Expected database.port to be 5432, got %d", tc.format, dbConfig.GetInt("port"))
			}

			if dbConfig.GetString("name") != "testdb" {
				t.Errorf("%s: Expected database.name to be 'testdb', got '%s'", tc.format, dbConfig.GetString("name"))
			}
		})
	}
}

// **Feature: todo-implementations, Property 14: Type-safe value retrieval**
// **Validates: Requirements 4.2**
func TestProperty_TypeSafeValueRetrieval(t *testing.T) {
	// Property: For any configuration value, when retrieved, it should maintain
	// its original type (string, int, bool, etc.)

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// Create a config with various types
	configContent := `{
		"string_value": "hello world",
		"int_value": 42,
		"int64_value": 1234567890,
		"float_value": 3.14159,
		"bool_true": true,
		"bool_false": false,
		"duration_value": "5m30s",
		"string_slice": ["item1", "item2", "item3"],
		"nested": {
			"inner_string": "nested value",
			"inner_int": 100
		}
	}`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	config := NewConfigManager()
	err = config.Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test string retrieval
	stringVal := config.GetString("string_value")
	if stringVal != "hello world" {
		t.Errorf("Expected string 'hello world', got '%s'", stringVal)
	}

	// Test int retrieval
	intVal := config.GetInt("int_value")
	if intVal != 42 {
		t.Errorf("Expected int 42, got %d", intVal)
	}

	// Test int64 retrieval
	int64Val := config.GetInt64("int64_value")
	if int64Val != 1234567890 {
		t.Errorf("Expected int64 1234567890, got %d", int64Val)
	}

	// Test float64 retrieval
	floatVal := config.GetFloat64("float_value")
	if floatVal < 3.14 || floatVal > 3.15 {
		t.Errorf("Expected float ~3.14159, got %f", floatVal)
	}

	// Test bool retrieval (true)
	boolTrue := config.GetBool("bool_true")
	if !boolTrue {
		t.Error("Expected bool true, got false")
	}

	// Test bool retrieval (false)
	boolFalse := config.GetBool("bool_false")
	if boolFalse {
		t.Error("Expected bool false, got true")
	}

	// Test duration retrieval
	durationVal := config.GetDuration("duration_value")
	expectedDuration := 5*time.Minute + 30*time.Second
	if durationVal != expectedDuration {
		t.Errorf("Expected duration %v, got %v", expectedDuration, durationVal)
	}

	// Test string slice retrieval
	sliceVal := config.GetStringSlice("string_slice")
	if len(sliceVal) != 3 {
		t.Errorf("Expected slice length 3, got %d", len(sliceVal))
	}
	if sliceVal[0] != "item1" || sliceVal[1] != "item2" || sliceVal[2] != "item3" {
		t.Errorf("Expected ['item1', 'item2', 'item3'], got %v", sliceVal)
	}

	// Test nested value retrieval
	nestedConfig := config.Sub("nested")
	nestedString := nestedConfig.GetString("inner_string")
	if nestedString != "nested value" {
		t.Errorf("Expected nested string 'nested value', got '%s'", nestedString)
	}

	nestedInt := nestedConfig.GetInt("inner_int")
	if nestedInt != 100 {
		t.Errorf("Expected nested int 100, got %d", nestedInt)
	}

	// Test type conversion: int to string
	intAsString := config.GetString("int_value")
	if intAsString != "42" {
		t.Errorf("Expected int converted to string '42', got '%s'", intAsString)
	}

	// Test type conversion: string to int (should work for numeric strings)
	tmpDir2 := t.TempDir()
	configPath2 := filepath.Join(tmpDir2, "config2.json")
	configContent2 := `{"numeric_string": "123"}`
	err = os.WriteFile(configPath2, []byte(configContent2), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	config2 := NewConfigManager()
	err = config2.Load(configPath2)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	numericInt := config2.GetInt("numeric_string")
	if numericInt != 123 {
		t.Errorf("Expected string converted to int 123, got %d", numericInt)
	}
}

// **Feature: todo-implementations, Property 15: Missing key handling**
// **Validates: Requirements 4.3**
func TestProperty_MissingKeyHandling(t *testing.T) {
	// Property: For any non-existent configuration key, the ConfigManager should
	// either return a default value or an error as appropriate

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// Create a config with some keys
	configContent := `{
		"existing_string": "value",
		"existing_int": 42,
		"existing_bool": true
	}`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	config := NewConfigManager()
	err = config.Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test missing keys return zero values
	missingString := config.GetString("missing_key")
	if missingString != "" {
		t.Errorf("Expected empty string for missing key, got '%s'", missingString)
	}

	missingInt := config.GetInt("missing_key")
	if missingInt != 0 {
		t.Errorf("Expected 0 for missing int key, got %d", missingInt)
	}

	missingInt64 := config.GetInt64("missing_key")
	if missingInt64 != 0 {
		t.Errorf("Expected 0 for missing int64 key, got %d", missingInt64)
	}

	missingFloat := config.GetFloat64("missing_key")
	if missingFloat != 0.0 {
		t.Errorf("Expected 0.0 for missing float key, got %f", missingFloat)
	}

	missingBool := config.GetBool("missing_key")
	if missingBool != false {
		t.Error("Expected false for missing bool key, got true")
	}

	missingDuration := config.GetDuration("missing_key")
	if missingDuration != 0 {
		t.Errorf("Expected 0 duration for missing key, got %v", missingDuration)
	}

	missingSlice := config.GetStringSlice("missing_key")
	if missingSlice != nil {
		t.Errorf("Expected nil for missing slice key, got %v", missingSlice)
	}

	// Test IsSet correctly identifies missing keys
	if config.IsSet("missing_key") {
		t.Error("Expected IsSet to return false for missing key")
	}

	if !config.IsSet("existing_string") {
		t.Error("Expected IsSet to return true for existing key")
	}

	// Test GetWithDefault returns default for missing keys
	defaultValue := config.GetWithDefault("missing_key", "default_value")
	if defaultValue != "default_value" {
		t.Errorf("Expected default value 'default_value', got '%v'", defaultValue)
	}

	// Test GetWithDefault returns actual value for existing keys
	existingValue := config.GetWithDefault("existing_string", "default_value")
	if existingValue != "value" {
		t.Errorf("Expected existing value 'value', got '%v'", existingValue)
	}

	// Test typed GetWithDefault methods
	defaultString := config.GetStringWithDefault("missing_key", "default")
	if defaultString != "default" {
		t.Errorf("Expected default string 'default', got '%s'", defaultString)
	}

	defaultInt := config.GetIntWithDefault("missing_key", 999)
	if defaultInt != 999 {
		t.Errorf("Expected default int 999, got %d", defaultInt)
	}

	defaultBool := config.GetBoolWithDefault("missing_key", true)
	if !defaultBool {
		t.Error("Expected default bool true, got false")
	}

	// Test that existing keys don't return defaults
	existingString := config.GetStringWithDefault("existing_string", "default")
	if existingString != "value" {
		t.Errorf("Expected existing value 'value', got '%s'", existingString)
	}

	existingInt := config.GetIntWithDefault("existing_int", 999)
	if existingInt != 42 {
		t.Errorf("Expected existing value 42, got %d", existingInt)
	}

	existingBool := config.GetBoolWithDefault("existing_bool", false)
	if !existingBool {
		t.Error("Expected existing value true, got false")
	}

	// Test Sub with missing key returns empty config
	missingSubConfig := config.Sub("missing_section")
	if missingSubConfig.GetString("any_key") != "" {
		t.Error("Expected empty string from missing sub-config")
	}

	if missingSubConfig.IsSet("any_key") {
		t.Error("Expected IsSet to return false for keys in missing sub-config")
	}
}

// **Feature: todo-implementations, Property 16: Configuration reload**
// **Validates: Requirements 4.5**
func TestProperty_ConfigurationReload(t *testing.T) {
	// Property: For any configuration file, when reloaded after modification,
	// the ConfigManager should reflect the new values without application restart

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// Create initial configuration
	initialConfig := `{
		"app_name": "InitialApp",
		"port": 8080,
		"debug": false,
		"database": {
			"host": "localhost",
			"port": 5432
		}
	}`

	err := os.WriteFile(configPath, []byte(initialConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Load initial configuration
	config := NewConfigManager()
	err = config.Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify initial values
	if config.GetString("app_name") != "InitialApp" {
		t.Errorf("Expected initial app_name 'InitialApp', got '%s'", config.GetString("app_name"))
	}

	if config.GetInt("port") != 8080 {
		t.Errorf("Expected initial port 8080, got %d", config.GetInt("port"))
	}

	if config.GetBool("debug") {
		t.Error("Expected initial debug to be false")
	}

	dbConfig := config.Sub("database")
	if dbConfig.GetString("host") != "localhost" {
		t.Errorf("Expected initial database.host 'localhost', got '%s'", dbConfig.GetString("host"))
	}

	// Modify configuration file
	updatedConfig := `{
		"app_name": "UpdatedApp",
		"port": 9090,
		"debug": true,
		"new_key": "new_value",
		"database": {
			"host": "remotehost",
			"port": 3306,
			"new_db_key": "db_value"
		}
	}`

	// Wait a bit to ensure file modification time changes
	time.Sleep(10 * time.Millisecond)

	err = os.WriteFile(configPath, []byte(updatedConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to update test config file: %v", err)
	}

	// Reload configuration
	err = config.Reload()
	if err != nil {
		t.Fatalf("Failed to reload config: %v", err)
	}

	// Verify updated values
	if config.GetString("app_name") != "UpdatedApp" {
		t.Errorf("Expected updated app_name 'UpdatedApp', got '%s'", config.GetString("app_name"))
	}

	if config.GetInt("port") != 9090 {
		t.Errorf("Expected updated port 9090, got %d", config.GetInt("port"))
	}

	if !config.GetBool("debug") {
		t.Error("Expected updated debug to be true")
	}

	// Verify new key was added
	if config.GetString("new_key") != "new_value" {
		t.Errorf("Expected new_key 'new_value', got '%s'", config.GetString("new_key"))
	}

	// Verify nested values were updated
	dbConfig = config.Sub("database")
	if dbConfig.GetString("host") != "remotehost" {
		t.Errorf("Expected updated database.host 'remotehost', got '%s'", dbConfig.GetString("host"))
	}

	if dbConfig.GetInt("port") != 3306 {
		t.Errorf("Expected updated database.port 3306, got %d", dbConfig.GetInt("port"))
	}

	// Verify new nested key was added
	if dbConfig.GetString("new_db_key") != "db_value" {
		t.Errorf("Expected new_db_key 'db_value', got '%s'", dbConfig.GetString("new_db_key"))
	}

	// Test reload with different formats
	formats := []struct {
		ext     string
		content string
	}{
		{
			ext: ".yaml",
			content: `
app_name: YAMLApp
port: 7070
debug: false
`,
		},
		{
			ext: ".toml",
			content: `
app_name = "TOMLApp"
port = 6060
debug = true
`,
		},
	}

	for _, format := range formats {
		formatConfigPath := filepath.Join(tmpDir, "config"+format.ext)
		err = os.WriteFile(formatConfigPath, []byte(format.content), 0644)
		if err != nil {
			t.Fatalf("Failed to create %s config file: %v", format.ext, err)
		}

		formatConfig := NewConfigManager()
		err = formatConfig.Load(formatConfigPath)
		if err != nil {
			t.Fatalf("Failed to load %s config: %v", format.ext, err)
		}

		// Modify and reload
		updatedContent := format.content
		if format.ext == ".yaml" {
			updatedContent = `
app_name: UpdatedYAMLApp
port: 7777
debug: true
`
		} else if format.ext == ".toml" {
			updatedContent = `
app_name = "UpdatedTOMLApp"
port = 6666
debug = false
`
		}

		time.Sleep(10 * time.Millisecond)
		err = os.WriteFile(formatConfigPath, []byte(updatedContent), 0644)
		if err != nil {
			t.Fatalf("Failed to update %s config file: %v", format.ext, err)
		}

		err = formatConfig.Reload()
		if err != nil {
			t.Fatalf("Failed to reload %s config: %v", format.ext, err)
		}

		// Verify reload worked for this format
		appName := formatConfig.GetString("app_name")
		if format.ext == ".yaml" && appName != "UpdatedYAMLApp" {
			t.Errorf("Expected UpdatedYAMLApp, got '%s'", appName)
		} else if format.ext == ".toml" && appName != "UpdatedTOMLApp" {
			t.Errorf("Expected UpdatedTOMLApp, got '%s'", appName)
		}
	}
}

func TestConfigManager_LoadJSON(t *testing.T) {
	// Create temporary JSON file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	jsonContent := `{
		"app_name": "MyApp",
		"port": 8080,
		"debug": true,
		"timeout": "30s",
		"features": ["auth", "api", "websocket"],
		"database": {
			"host": "localhost",
			"port": 5432,
			"name": "mydb",
			"max_connections": 100
		},
		"server": {
			"timeout": "30s",
			"max_body_size": 1048576,
			"tls": {
				"enabled": true,
				"cert_file": "/path/to/cert.pem",
				"key_file": "/path/to/key.pem"
			}
		}
	}`

	err := os.WriteFile(configPath, []byte(jsonContent), 0644)
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

	// Test array
	features := config.GetStringSlice("features")
	if len(features) != 3 {
		t.Errorf("Expected 3 features, got %d", len(features))
	}
	if features[0] != "auth" || features[1] != "api" || features[2] != "websocket" {
		t.Errorf("Expected ['auth', 'api', 'websocket'], got %v", features)
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
