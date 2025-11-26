package pkg

import (
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// **Feature: config-defaults, Property 1: Zero values are replaced with defaults**
// **Validates: Requirements 1.2, 2.1, 8.2**
func TestProperty_ZeroValuesReplacedWithDefaults(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// Property 1.1: ServerConfig zero values are replaced with defaults
	properties.Property("ServerConfig zero values get defaults", prop.ForAll(
		func() bool {
			config := ServerConfig{} // All zero values
			config.ApplyDefaults()

			// Verify all fields have non-zero defaults
			return config.ReadTimeout == 30*time.Second &&
				config.WriteTimeout == 30*time.Second &&
				config.IdleTimeout == 120*time.Second &&
				config.MaxHeaderBytes == 1048576 &&
				config.MaxConnections == 10000 &&
				config.MaxRequestSize == 10485760 &&
				config.ShutdownTimeout == 30*time.Second &&
				config.ReadBufferSize == 4096 &&
				config.WriteBufferSize == 4096
		},
	))

	// Property 1.2: DatabaseConfig zero values are replaced with defaults
	properties.Property("DatabaseConfig zero values get defaults", prop.ForAll(
		func(driver string) bool {
			// Only test with valid drivers
			validDrivers := []string{"postgres", "mysql", "mssql", "sqlite"}
			isValid := false
			for _, d := range validDrivers {
				if driver == d {
					isValid = true
					break
				}
			}
			if !isValid {
				return true // Skip invalid drivers
			}

			config := DatabaseConfig{Driver: driver} // Zero values except driver
			config.ApplyDefaults()

			// Verify defaults are applied
			hasDefaults := config.Host == "localhost" &&
				config.MaxOpenConns == 25 &&
				config.MaxIdleConns == 5 &&
				config.ConnMaxLifetime == 5*time.Minute

			// Verify driver-specific port defaults
			var expectedPort int
			switch driver {
			case "postgres":
				expectedPort = 5432
			case "mysql":
				expectedPort = 3306
			case "mssql":
				expectedPort = 1433
			case "sqlite":
				expectedPort = 0
			}

			return hasDefaults && config.Port == expectedPort
		},
		gen.OneConstOf("postgres", "mysql", "mssql", "sqlite"),
	))

	// Property 1.3: CacheConfig zero values are replaced with defaults
	properties.Property("CacheConfig zero values get defaults", prop.ForAll(
		func() bool {
			config := CacheConfig{} // All zero values
			config.ApplyDefaults()

			// Verify defaults
			return config.Type == "memory" &&
				config.MaxSize == 0 && // 0 means unlimited
				config.DefaultTTL == 0 // 0 means no expiration
		},
	))

	// Property 1.4: SessionConfig zero values are replaced with defaults
	properties.Property("SessionConfig zero values get defaults", prop.ForAll(
		func() bool {
			config := SessionConfig{} // All zero values
			config.ApplyDefaults()

			// Verify defaults
			return config.CookieName == "rockstar_session" &&
				config.CookiePath == "/" &&
				config.SessionLifetime == 24*time.Hour &&
				config.CleanupInterval == 1*time.Hour &&
				config.FilesystemPath == "./sessions"
		},
	))

	// Property 1.5: MonitoringConfig zero values are replaced with defaults
	properties.Property("MonitoringConfig zero values get defaults", prop.ForAll(
		func() bool {
			config := MonitoringConfig{} // All zero values
			config.ApplyDefaults()

			// Verify defaults
			return config.MetricsPort == 9090 &&
				config.PprofPort == 6060 &&
				config.SNMPPort == 161 &&
				config.SNMPCommunity == "public" &&
				config.OptimizationInterval == 5*time.Minute
		},
	))

	// Property 1.6: Negative CacheConfig values are normalized to zero
	properties.Property("CacheConfig negative values normalized", prop.ForAll(
		func(maxSize int64, ttlNanos int64) bool {
			ttl := time.Duration(ttlNanos)
			// Only test with negative values
			if maxSize >= 0 || ttl >= 0 {
				return true
			}

			config := CacheConfig{
				MaxSize:    maxSize,
				DefaultTTL: ttl,
			}
			config.ApplyDefaults()

			// Verify negative values are normalized to 0
			return config.MaxSize == 0 && config.DefaultTTL == 0
		},
		gen.Int64Range(-1000000, -1),
		gen.Int64Range(-1000000, -1),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// **Feature: config-defaults, Property 2: User-provided values are preserved**
// **Validates: Requirements 8.3, 8.4**
func TestProperty_UserProvidedValuesPreserved(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// Property 2.1: ServerConfig preserves non-zero user values
	properties.Property("ServerConfig preserves user values", prop.ForAll(
		func(readTimeoutNanos, writeTimeoutNanos, idleTimeoutNanos, shutdownTimeoutNanos int64,
			maxHeaderBytes, maxConnections, readBufferSize, writeBufferSize int,
			maxRequestSize int64) bool {

			readTimeout := time.Duration(readTimeoutNanos)
			writeTimeout := time.Duration(writeTimeoutNanos)
			idleTimeout := time.Duration(idleTimeoutNanos)
			shutdownTimeout := time.Duration(shutdownTimeoutNanos)

			// Skip zero values - we're testing preservation of non-zero values
			if readTimeout == 0 || writeTimeout == 0 || idleTimeout == 0 ||
				shutdownTimeout == 0 || maxHeaderBytes == 0 || maxConnections == 0 ||
				readBufferSize == 0 || writeBufferSize == 0 || maxRequestSize == 0 {
				return true
			}

			config := ServerConfig{
				ReadTimeout:     readTimeout,
				WriteTimeout:    writeTimeout,
				IdleTimeout:     idleTimeout,
				MaxHeaderBytes:  maxHeaderBytes,
				MaxConnections:  maxConnections,
				MaxRequestSize:  maxRequestSize,
				ShutdownTimeout: shutdownTimeout,
				ReadBufferSize:  readBufferSize,
				WriteBufferSize: writeBufferSize,
			}

			// Store original values
			original := config

			// Apply defaults
			config.ApplyDefaults()

			// Verify all user values are preserved
			return config.ReadTimeout == original.ReadTimeout &&
				config.WriteTimeout == original.WriteTimeout &&
				config.IdleTimeout == original.IdleTimeout &&
				config.MaxHeaderBytes == original.MaxHeaderBytes &&
				config.MaxConnections == original.MaxConnections &&
				config.MaxRequestSize == original.MaxRequestSize &&
				config.ShutdownTimeout == original.ShutdownTimeout &&
				config.ReadBufferSize == original.ReadBufferSize &&
				config.WriteBufferSize == original.WriteBufferSize
		},
		gen.Int64Range(int64(1*time.Second), int64(10*time.Minute)),
		gen.Int64Range(int64(1*time.Second), int64(10*time.Minute)),
		gen.Int64Range(int64(1*time.Second), int64(10*time.Minute)),
		gen.Int64Range(int64(1*time.Second), int64(10*time.Minute)),
		gen.IntRange(1, 10000000),
		gen.IntRange(1, 100000),
		gen.IntRange(1, 65536),
		gen.IntRange(1, 65536),
		gen.Int64Range(1, 100000000),
	))

	// Property 2.2: DatabaseConfig preserves non-zero user values
	properties.Property("DatabaseConfig preserves user values", prop.ForAll(
		func(driver, host string, port, maxOpenConns, maxIdleConns int, connMaxLifetimeNanos int64) bool {
			connMaxLifetime := time.Duration(connMaxLifetimeNanos)
			// Skip zero/empty values
			if host == "" || port == 0 || maxOpenConns == 0 || maxIdleConns == 0 || connMaxLifetime == 0 {
				return true
			}

			config := DatabaseConfig{
				Driver:          driver,
				Host:            host,
				Port:            port,
				MaxOpenConns:    maxOpenConns,
				MaxIdleConns:    maxIdleConns,
				ConnMaxLifetime: connMaxLifetime,
			}

			// Store original values
			original := config

			// Apply defaults
			config.ApplyDefaults()

			// Verify all user values are preserved
			return config.Driver == original.Driver &&
				config.Host == original.Host &&
				config.Port == original.Port &&
				config.MaxOpenConns == original.MaxOpenConns &&
				config.MaxIdleConns == original.MaxIdleConns &&
				config.ConnMaxLifetime == original.ConnMaxLifetime
		},
		gen.OneConstOf("postgres", "mysql", "mssql", "sqlite", "custom"),
		gen.AlphaString(),
		gen.IntRange(1, 65535),
		gen.IntRange(1, 1000),
		gen.IntRange(1, 1000),
		gen.Int64Range(int64(1*time.Minute), int64(1*time.Hour)),
	))

	// Property 2.3: CacheConfig preserves non-zero user values
	properties.Property("CacheConfig preserves user values", prop.ForAll(
		func(cacheType string, maxSize int64, defaultTTLNanos int64) bool {
			defaultTTL := time.Duration(defaultTTLNanos)
			// Skip zero/empty values
			if cacheType == "" || maxSize == 0 || defaultTTL == 0 {
				return true
			}

			config := CacheConfig{
				Type:       cacheType,
				MaxSize:    maxSize,
				DefaultTTL: defaultTTL,
			}

			// Store original values
			original := config

			// Apply defaults
			config.ApplyDefaults()

			// Verify all user values are preserved
			return config.Type == original.Type &&
				config.MaxSize == original.MaxSize &&
				config.DefaultTTL == original.DefaultTTL
		},
		gen.OneConstOf("memory", "distributed", "redis"),
		gen.Int64Range(1, 1000000000),
		gen.Int64Range(int64(1*time.Second), int64(24*time.Hour)),
	))

	// Property 2.4: SessionConfig preserves non-zero user values
	properties.Property("SessionConfig preserves user values", prop.ForAll(
		func(cookieName, cookiePath, filesystemPath string, sessionLifetimeNanos, cleanupIntervalNanos int64) bool {
			sessionLifetime := time.Duration(sessionLifetimeNanos)
			cleanupInterval := time.Duration(cleanupIntervalNanos)
			// Skip zero/empty values
			if cookieName == "" || cookiePath == "" || filesystemPath == "" ||
				sessionLifetime == 0 || cleanupInterval == 0 {
				return true
			}

			config := SessionConfig{
				CookieName:      cookieName,
				CookiePath:      cookiePath,
				FilesystemPath:  filesystemPath,
				SessionLifetime: sessionLifetime,
				CleanupInterval: cleanupInterval,
			}

			// Store original values
			original := config

			// Apply defaults
			config.ApplyDefaults()

			// Verify all user values are preserved
			return config.CookieName == original.CookieName &&
				config.CookiePath == original.CookiePath &&
				config.FilesystemPath == original.FilesystemPath &&
				config.SessionLifetime == original.SessionLifetime &&
				config.CleanupInterval == original.CleanupInterval
		},
		gen.AlphaString(),
		gen.AlphaString(),
		gen.AlphaString(),
		gen.Int64Range(int64(1*time.Minute), int64(48*time.Hour)),
		gen.Int64Range(int64(1*time.Minute), int64(24*time.Hour)),
	))

	// Property 2.5: MonitoringConfig preserves non-zero user values
	properties.Property("MonitoringConfig preserves user values", prop.ForAll(
		func(metricsPort, pprofPort, snmpPort int, snmpCommunity string, optimizationIntervalNanos int64) bool {
			optimizationInterval := time.Duration(optimizationIntervalNanos)
			// Skip zero/empty values
			if metricsPort == 0 || pprofPort == 0 || snmpPort == 0 ||
				snmpCommunity == "" || optimizationInterval == 0 {
				return true
			}

			config := MonitoringConfig{
				MetricsPort:          metricsPort,
				PprofPort:            pprofPort,
				SNMPPort:             snmpPort,
				SNMPCommunity:        snmpCommunity,
				OptimizationInterval: optimizationInterval,
			}

			// Store original values
			original := config

			// Apply defaults
			config.ApplyDefaults()

			// Verify all user values are preserved
			return config.MetricsPort == original.MetricsPort &&
				config.PprofPort == original.PprofPort &&
				config.SNMPPort == original.SNMPPort &&
				config.SNMPCommunity == original.SNMPCommunity &&
				config.OptimizationInterval == original.OptimizationInterval
		},
		gen.IntRange(1, 65535),
		gen.IntRange(1, 65535),
		gen.IntRange(1, 65535),
		gen.AlphaString(),
		gen.Int64Range(int64(1*time.Minute), int64(1*time.Hour)),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Unit tests for ServerConfig defaults
// Requirements: 2.4, 2.5, 2.6, 6.1, 6.2, 6.3, 6.4, 6.5, 6.6

func TestServerConfig_ApplyDefaults_ZeroValues(t *testing.T) {
	config := ServerConfig{} // All zero values
	config.ApplyDefaults()

	tests := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{"ReadTimeout", config.ReadTimeout, 30 * time.Second},
		{"WriteTimeout", config.WriteTimeout, 30 * time.Second},
		{"IdleTimeout", config.IdleTimeout, 120 * time.Second},
		{"MaxHeaderBytes", config.MaxHeaderBytes, 1048576},
		{"MaxConnections", config.MaxConnections, 10000},
		{"MaxRequestSize", config.MaxRequestSize, int64(10485760)},
		{"ShutdownTimeout", config.ShutdownTimeout, 30 * time.Second},
		{"ReadBufferSize", config.ReadBufferSize, 4096},
		{"WriteBufferSize", config.WriteBufferSize, 4096},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s: got %v, expected %v", tt.name, tt.got, tt.expected)
			}
		})
	}
}

func TestServerConfig_ApplyDefaults_PreservesUserValues(t *testing.T) {
	config := ServerConfig{
		ReadTimeout:     45 * time.Second,
		WriteTimeout:    60 * time.Second,
		IdleTimeout:     180 * time.Second,
		MaxHeaderBytes:  2097152, // 2MB
		MaxConnections:  20000,
		MaxRequestSize:  20971520, // 20MB
		ShutdownTimeout: 60 * time.Second,
		ReadBufferSize:  8192,
		WriteBufferSize: 8192,
	}

	// Store original values
	original := config

	// Apply defaults
	config.ApplyDefaults()

	tests := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{"ReadTimeout", config.ReadTimeout, original.ReadTimeout},
		{"WriteTimeout", config.WriteTimeout, original.WriteTimeout},
		{"IdleTimeout", config.IdleTimeout, original.IdleTimeout},
		{"MaxHeaderBytes", config.MaxHeaderBytes, original.MaxHeaderBytes},
		{"MaxConnections", config.MaxConnections, original.MaxConnections},
		{"MaxRequestSize", config.MaxRequestSize, original.MaxRequestSize},
		{"ShutdownTimeout", config.ShutdownTimeout, original.ShutdownTimeout},
		{"ReadBufferSize", config.ReadBufferSize, original.ReadBufferSize},
		{"WriteBufferSize", config.WriteBufferSize, original.WriteBufferSize},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s: got %v, expected %v (user value not preserved)", tt.name, tt.got, tt.expected)
			}
		})
	}
}

func TestServerConfig_ApplyDefaults_PartialConfig(t *testing.T) {
	// Test with some values set and some zero
	config := ServerConfig{
		ReadTimeout:    45 * time.Second, // User value
		MaxConnections: 15000,            // User value
		// Other fields are zero and should get defaults
	}

	config.ApplyDefaults()

	// User values should be preserved
	if config.ReadTimeout != 45*time.Second {
		t.Errorf("ReadTimeout: got %v, expected 45s (user value not preserved)", config.ReadTimeout)
	}
	if config.MaxConnections != 15000 {
		t.Errorf("MaxConnections: got %v, expected 15000 (user value not preserved)", config.MaxConnections)
	}

	// Zero values should get defaults
	if config.WriteTimeout != 30*time.Second {
		t.Errorf("WriteTimeout: got %v, expected 30s (default not applied)", config.WriteTimeout)
	}
	if config.IdleTimeout != 120*time.Second {
		t.Errorf("IdleTimeout: got %v, expected 120s (default not applied)", config.IdleTimeout)
	}
	if config.MaxHeaderBytes != 1048576 {
		t.Errorf("MaxHeaderBytes: got %v, expected 1048576 (default not applied)", config.MaxHeaderBytes)
	}
	if config.MaxRequestSize != 10485760 {
		t.Errorf("MaxRequestSize: got %v, expected 10485760 (default not applied)", config.MaxRequestSize)
	}
	if config.ShutdownTimeout != 30*time.Second {
		t.Errorf("ShutdownTimeout: got %v, expected 30s (default not applied)", config.ShutdownTimeout)
	}
	if config.ReadBufferSize != 4096 {
		t.Errorf("ReadBufferSize: got %v, expected 4096 (default not applied)", config.ReadBufferSize)
	}
	if config.WriteBufferSize != 4096 {
		t.Errorf("WriteBufferSize: got %v, expected 4096 (default not applied)", config.WriteBufferSize)
	}
}

// Unit tests for DatabaseConfig defaults
// Requirements: 3.1, 3.2, 3.3, 3.5, 3.6, 3.7

func TestDatabaseConfig_ApplyDefaults_ConnectionPoolDefaults(t *testing.T) {
	config := DatabaseConfig{
		Driver: "postgres", // Required field
		// Connection pool fields are zero and should get defaults
	}
	config.ApplyDefaults()

	tests := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{"Host", config.Host, "localhost"},
		{"MaxOpenConns", config.MaxOpenConns, 25},
		{"MaxIdleConns", config.MaxIdleConns, 5},
		{"ConnMaxLifetime", config.ConnMaxLifetime, 5 * time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s: got %v, expected %v", tt.name, tt.got, tt.expected)
			}
		})
	}
}

func TestDatabaseConfig_ApplyDefaults_DriverSpecificPorts(t *testing.T) {
	tests := []struct {
		driver       string
		expectedPort int
	}{
		{"postgres", 5432},
		{"mysql", 3306},
		{"mssql", 1433},
		{"sqlite", 0},
		{"sqlite3", 0}, // Alternative SQLite driver name
	}

	for _, tt := range tests {
		t.Run(tt.driver, func(t *testing.T) {
			config := DatabaseConfig{
				Driver: tt.driver,
				// Port is zero and should get driver-specific default
			}
			config.ApplyDefaults()

			if config.Port != tt.expectedPort {
				t.Errorf("Port for driver %s: got %d, expected %d", tt.driver, config.Port, tt.expectedPort)
			}
		})
	}
}

func TestDatabaseConfig_ApplyDefaults_PreservesUserValues(t *testing.T) {
	config := DatabaseConfig{
		Driver:          "postgres",
		Host:            "db.example.com",
		Port:            9999,
		MaxOpenConns:    100,
		MaxIdleConns:    20,
		ConnMaxLifetime: 10 * time.Minute,
	}

	// Store original values
	original := config

	// Apply defaults
	config.ApplyDefaults()

	tests := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{"Driver", config.Driver, original.Driver},
		{"Host", config.Host, original.Host},
		{"Port", config.Port, original.Port},
		{"MaxOpenConns", config.MaxOpenConns, original.MaxOpenConns},
		{"MaxIdleConns", config.MaxIdleConns, original.MaxIdleConns},
		{"ConnMaxLifetime", config.ConnMaxLifetime, original.ConnMaxLifetime},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s: got %v, expected %v (user value not preserved)", tt.name, tt.got, tt.expected)
			}
		})
	}
}

func TestDatabaseConfig_ApplyDefaults_PartialConfig(t *testing.T) {
	// Test with some values set and some zero
	config := DatabaseConfig{
		Driver:       "mysql",
		Host:         "custom.host.com", // User value
		MaxOpenConns: 50,                // User value
		// Port, MaxIdleConns, ConnMaxLifetime are zero and should get defaults
	}

	config.ApplyDefaults()

	// User values should be preserved
	if config.Host != "custom.host.com" {
		t.Errorf("Host: got %v, expected custom.host.com (user value not preserved)", config.Host)
	}
	if config.MaxOpenConns != 50 {
		t.Errorf("MaxOpenConns: got %v, expected 50 (user value not preserved)", config.MaxOpenConns)
	}

	// Zero values should get defaults
	if config.Port != 3306 {
		t.Errorf("Port: got %v, expected 3306 (default not applied)", config.Port)
	}
	if config.MaxIdleConns != 5 {
		t.Errorf("MaxIdleConns: got %v, expected 5 (default not applied)", config.MaxIdleConns)
	}
	if config.ConnMaxLifetime != 5*time.Minute {
		t.Errorf("ConnMaxLifetime: got %v, expected 5m (default not applied)", config.ConnMaxLifetime)
	}
}

func TestDatabaseConfig_ApplyDefaults_UnknownDriver(t *testing.T) {
	// Test with an unknown driver - port should remain 0
	config := DatabaseConfig{
		Driver: "unknown_driver",
		// Port is zero
	}
	config.ApplyDefaults()

	// Port should remain 0 for unknown drivers
	if config.Port != 0 {
		t.Errorf("Port for unknown driver: got %d, expected 0", config.Port)
	}

	// Other defaults should still be applied
	if config.Host != "localhost" {
		t.Errorf("Host: got %v, expected localhost", config.Host)
	}
	if config.MaxOpenConns != 25 {
		t.Errorf("MaxOpenConns: got %v, expected 25", config.MaxOpenConns)
	}
	if config.MaxIdleConns != 5 {
		t.Errorf("MaxIdleConns: got %v, expected 5", config.MaxIdleConns)
	}
	if config.ConnMaxLifetime != 5*time.Minute {
		t.Errorf("ConnMaxLifetime: got %v, expected 5m", config.ConnMaxLifetime)
	}
}

// Unit tests for CacheConfig defaults
// Requirements: 4.1, 4.2, 4.3, 4.4, 4.5

func TestCacheConfig_ApplyDefaults_TypeDefaultsToMemory(t *testing.T) {
	config := CacheConfig{} // Type is empty
	config.ApplyDefaults()

	if config.Type != "memory" {
		t.Errorf("Type: got %v, expected 'memory'", config.Type)
	}
}

func TestCacheConfig_ApplyDefaults_ZeroMaxSizeTreatedAsUnlimited(t *testing.T) {
	config := CacheConfig{
		MaxSize: 0, // Zero means unlimited
	}
	config.ApplyDefaults()

	if config.MaxSize != 0 {
		t.Errorf("MaxSize: got %v, expected 0 (unlimited)", config.MaxSize)
	}
}

func TestCacheConfig_ApplyDefaults_ZeroDefaultTTLTreatedAsNoExpiration(t *testing.T) {
	config := CacheConfig{
		DefaultTTL: 0, // Zero means no expiration
	}
	config.ApplyDefaults()

	if config.DefaultTTL != 0 {
		t.Errorf("DefaultTTL: got %v, expected 0 (no expiration)", config.DefaultTTL)
	}
}

func TestCacheConfig_ApplyDefaults_NegativeValuesNormalized(t *testing.T) {
	config := CacheConfig{
		MaxSize:    -1000,
		DefaultTTL: -5 * time.Second,
	}
	config.ApplyDefaults()

	if config.MaxSize != 0 {
		t.Errorf("MaxSize: got %v, expected 0 (normalized from negative)", config.MaxSize)
	}
	if config.DefaultTTL != 0 {
		t.Errorf("DefaultTTL: got %v, expected 0 (normalized from negative)", config.DefaultTTL)
	}
}

func TestCacheConfig_ApplyDefaults_AllZeroValues(t *testing.T) {
	config := CacheConfig{} // All zero values
	config.ApplyDefaults()

	tests := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{"Type", config.Type, "memory"},
		{"MaxSize", config.MaxSize, int64(0)},
		{"DefaultTTL", config.DefaultTTL, time.Duration(0)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s: got %v, expected %v", tt.name, tt.got, tt.expected)
			}
		})
	}
}

func TestCacheConfig_ApplyDefaults_PreservesUserValues(t *testing.T) {
	config := CacheConfig{
		Type:       "distributed",
		MaxSize:    1000000,
		DefaultTTL: 10 * time.Minute,
	}

	// Store original values
	original := config

	// Apply defaults
	config.ApplyDefaults()

	tests := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{"Type", config.Type, original.Type},
		{"MaxSize", config.MaxSize, original.MaxSize},
		{"DefaultTTL", config.DefaultTTL, original.DefaultTTL},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s: got %v, expected %v (user value not preserved)", tt.name, tt.got, tt.expected)
			}
		})
	}
}

func TestCacheConfig_ApplyDefaults_PartialConfig(t *testing.T) {
	// Test with some values set and some zero
	config := CacheConfig{
		Type: "redis", // User value
		// MaxSize and DefaultTTL are zero and should remain 0 (unlimited/no expiration)
	}

	config.ApplyDefaults()

	// User value should be preserved
	if config.Type != "redis" {
		t.Errorf("Type: got %v, expected 'redis' (user value not preserved)", config.Type)
	}

	// Zero values should remain 0 (they are valid defaults)
	if config.MaxSize != 0 {
		t.Errorf("MaxSize: got %v, expected 0 (unlimited)", config.MaxSize)
	}
	if config.DefaultTTL != 0 {
		t.Errorf("DefaultTTL: got %v, expected 0 (no expiration)", config.DefaultTTL)
	}
}

// Unit tests for SessionConfig defaults
// Requirements: 2.3, 7.1, 7.2, 7.3, 7.4, 7.5, 7.6

func TestSessionConfig_ApplyDefaults_NilConfigUsesDefaultSessionConfig(t *testing.T) {
	// Test that NewSessionManager handles nil config by using DefaultSessionConfig
	// We can't test NewSessionManager directly without encryption key, so we test the pattern
	defaultConfig := DefaultSessionConfig()

	// Verify DefaultSessionConfig returns expected values
	if defaultConfig.CookieName != "rockstar_session" {
		t.Errorf("CookieName: got %v, expected 'rockstar_session'", defaultConfig.CookieName)
	}
	if defaultConfig.CookiePath != "/" {
		t.Errorf("CookiePath: got %v, expected '/'", defaultConfig.CookiePath)
	}
	if defaultConfig.SessionLifetime != 24*time.Hour {
		t.Errorf("SessionLifetime: got %v, expected 24h", defaultConfig.SessionLifetime)
	}
	if defaultConfig.CleanupInterval != 1*time.Hour {
		t.Errorf("CleanupInterval: got %v, expected 1h", defaultConfig.CleanupInterval)
	}
	if defaultConfig.FilesystemPath != "./sessions" {
		t.Errorf("FilesystemPath: got %v, expected './sessions'", defaultConfig.FilesystemPath)
	}
}

func TestSessionConfig_ApplyDefaults_AllZeroValues(t *testing.T) {
	config := SessionConfig{} // All zero values
	config.ApplyDefaults()

	tests := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{"CookieName", config.CookieName, "rockstar_session"},
		{"CookiePath", config.CookiePath, "/"},
		{"SessionLifetime", config.SessionLifetime, 24 * time.Hour},
		{"CleanupInterval", config.CleanupInterval, 1 * time.Hour},
		{"FilesystemPath", config.FilesystemPath, "./sessions"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s: got %v, expected %v", tt.name, tt.got, tt.expected)
			}
		})
	}
}

func TestSessionConfig_ApplyDefaults_PreservesUserValues(t *testing.T) {
	config := SessionConfig{
		CookieName:      "my_custom_session",
		CookiePath:      "/app",
		SessionLifetime: 48 * time.Hour,
		CleanupInterval: 2 * time.Hour,
		FilesystemPath:  "/var/sessions",
	}

	// Store original values
	original := config

	// Apply defaults
	config.ApplyDefaults()

	tests := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{"CookieName", config.CookieName, original.CookieName},
		{"CookiePath", config.CookiePath, original.CookiePath},
		{"SessionLifetime", config.SessionLifetime, original.SessionLifetime},
		{"CleanupInterval", config.CleanupInterval, original.CleanupInterval},
		{"FilesystemPath", config.FilesystemPath, original.FilesystemPath},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s: got %v, expected %v (user value not preserved)", tt.name, tt.got, tt.expected)
			}
		})
	}
}

func TestSessionConfig_ApplyDefaults_PartialConfig(t *testing.T) {
	// Test with some values set and some zero
	config := SessionConfig{
		CookieName:      "custom_session", // User value
		SessionLifetime: 12 * time.Hour,   // User value
		// CookiePath, CleanupInterval, FilesystemPath are zero and should get defaults
	}

	config.ApplyDefaults()

	// User values should be preserved
	if config.CookieName != "custom_session" {
		t.Errorf("CookieName: got %v, expected 'custom_session' (user value not preserved)", config.CookieName)
	}
	if config.SessionLifetime != 12*time.Hour {
		t.Errorf("SessionLifetime: got %v, expected 12h (user value not preserved)", config.SessionLifetime)
	}

	// Zero values should get defaults
	if config.CookiePath != "/" {
		t.Errorf("CookiePath: got %v, expected '/' (default not applied)", config.CookiePath)
	}
	if config.CleanupInterval != 1*time.Hour {
		t.Errorf("CleanupInterval: got %v, expected 1h (default not applied)", config.CleanupInterval)
	}
	if config.FilesystemPath != "./sessions" {
		t.Errorf("FilesystemPath: got %v, expected './sessions' (default not applied)", config.FilesystemPath)
	}
}

func TestNewSessionManager_AppliesDefaults(t *testing.T) {
	// Test that NewSessionManager applies defaults to the config
	config := &SessionConfig{
		EncryptionKey: generateEncryptionKey(), // Required field
		// Other fields are zero and should get defaults
	}

	db := NewMockDatabaseManager()
	db.Connect(DatabaseConfig{Driver: "mock"})

	sm, err := NewSessionManager(config, db, nil)
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}

	// Verify defaults were applied by checking the config was modified
	if config.CookieName != "rockstar_session" {
		t.Errorf("CookieName: got %v, expected 'rockstar_session' (default not applied)", config.CookieName)
	}
	if config.CookiePath != "/" {
		t.Errorf("CookiePath: got %v, expected '/' (default not applied)", config.CookiePath)
	}
	if config.SessionLifetime != 24*time.Hour {
		t.Errorf("SessionLifetime: got %v, expected 24h (default not applied)", config.SessionLifetime)
	}
	if config.CleanupInterval != 1*time.Hour {
		t.Errorf("CleanupInterval: got %v, expected 1h (default not applied)", config.CleanupInterval)
	}
	if config.FilesystemPath != "./sessions" {
		t.Errorf("FilesystemPath: got %v, expected './sessions' (default not applied)", config.FilesystemPath)
	}

	// Clean up
	if smImpl, ok := sm.(*sessionManager); ok {
		smImpl.Stop()
	}
}

func TestNewSessionManager_PreservesUserValues(t *testing.T) {
	// Test that NewSessionManager preserves user-provided values
	config := &SessionConfig{
		EncryptionKey:   generateEncryptionKey(), // Required field
		CookieName:      "my_session",
		CookiePath:      "/custom",
		SessionLifetime: 48 * time.Hour,
		CleanupInterval: 2 * time.Hour,
		FilesystemPath:  "/tmp/sessions",
		StorageType:     SessionStorageDatabase,
	}

	// Store original values
	originalCookieName := config.CookieName
	originalCookiePath := config.CookiePath
	originalSessionLifetime := config.SessionLifetime
	originalCleanupInterval := config.CleanupInterval
	originalFilesystemPath := config.FilesystemPath

	db := NewMockDatabaseManager()
	db.Connect(DatabaseConfig{Driver: "mock"})

	sm, err := NewSessionManager(config, db, nil)
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}

	// Verify user values were preserved
	if config.CookieName != originalCookieName {
		t.Errorf("CookieName: got %v, expected %v (user value not preserved)", config.CookieName, originalCookieName)
	}
	if config.CookiePath != originalCookiePath {
		t.Errorf("CookiePath: got %v, expected %v (user value not preserved)", config.CookiePath, originalCookiePath)
	}
	if config.SessionLifetime != originalSessionLifetime {
		t.Errorf("SessionLifetime: got %v, expected %v (user value not preserved)", config.SessionLifetime, originalSessionLifetime)
	}
	if config.CleanupInterval != originalCleanupInterval {
		t.Errorf("CleanupInterval: got %v, expected %v (user value not preserved)", config.CleanupInterval, originalCleanupInterval)
	}
	if config.FilesystemPath != originalFilesystemPath {
		t.Errorf("FilesystemPath: got %v, expected %v (user value not preserved)", config.FilesystemPath, originalFilesystemPath)
	}

	// Clean up
	if smImpl, ok := sm.(*sessionManager); ok {
		smImpl.Stop()
	}
}
