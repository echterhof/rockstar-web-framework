package pkg

import (
	"crypto/rand"
	"time"
)

// ApplyDefaults applies default values to ServerConfig for any zero-valued fields
// Default: ReadTimeout=30s, WriteTimeout=30s, IdleTimeout=120s, MaxHeaderBytes=1MB,
// MaxConnections=10000, MaxRequestSize=10MB, ShutdownTimeout=30s,
// ReadBufferSize=4096, WriteBufferSize=4096
func (c *ServerConfig) ApplyDefaults() {
	if c.ReadTimeout == 0 {
		c.ReadTimeout = 30 * time.Second
	}
	if c.WriteTimeout == 0 {
		c.WriteTimeout = 30 * time.Second
	}
	if c.IdleTimeout == 0 {
		c.IdleTimeout = 120 * time.Second
	}
	if c.MaxHeaderBytes == 0 {
		c.MaxHeaderBytes = 1048576 // 1 MB
	}
	if c.MaxConnections == 0 {
		c.MaxConnections = 10000
	}
	if c.MaxRequestSize == 0 {
		c.MaxRequestSize = 10485760 // 10 MB
	}
	if c.ShutdownTimeout == 0 {
		c.ShutdownTimeout = 30 * time.Second
	}
	if c.ReadBufferSize == 0 {
		c.ReadBufferSize = 4096
	}
	if c.WriteBufferSize == 0 {
		c.WriteBufferSize = 4096
	}
}

// ApplyDefaults applies default values to DatabaseConfig for any zero-valued fields
// Default: Host="localhost", MaxOpenConns=25, MaxIdleConns=5, ConnMaxLifetime=5m
// Port defaults are driver-specific: postgres=5432, mysql=3306, mssql=1433, sqlite=0
func (c *DatabaseConfig) ApplyDefaults() {
	if c.Host == "" {
		c.Host = "localhost"
	}
	if c.Port == 0 {
		// Apply driver-specific port defaults
		switch c.Driver {
		case "postgres":
			c.Port = 5432
		case "mysql":
			c.Port = 3306
		case "mssql":
			c.Port = 1433
		case "sqlite":
			c.Port = 0 // SQLite doesn't use ports
		}
	}
	if c.MaxOpenConns == 0 {
		c.MaxOpenConns = 25
	}
	if c.MaxIdleConns == 0 {
		c.MaxIdleConns = 5
	}
	if c.ConnMaxLifetime == 0 {
		c.ConnMaxLifetime = 5 * time.Minute
	}
}

// ApplyDefaults applies default values to CacheConfig for any zero-valued fields
// Default: Type="memory", MaxSize=0 (unlimited), DefaultTTL=0 (no expiration)
// Negative values are normalized to 0
func (c *CacheConfig) ApplyDefaults() {
	if c.Type == "" {
		c.Type = "memory"
	}
	// Normalize negative values to 0 (unlimited/no expiration)
	if c.MaxSize < 0 {
		c.MaxSize = 0
	}
	if c.DefaultTTL < 0 {
		c.DefaultTTL = 0
	}
}

// ApplyDefaults applies default values to SessionConfig for any zero-valued fields
// Default: CookieName="rockstar_session", CookiePath="/", SessionLifetime=24h,
// CleanupInterval=1h, FilesystemPath="./sessions", EncryptionKey=random 32 bytes
func (c *SessionConfig) ApplyDefaults() {
	if c.CookieName == "" {
		c.CookieName = "rockstar_session"
	}
	if c.CookiePath == "" {
		c.CookiePath = "/"
	}
	if c.SessionLifetime == 0 {
		c.SessionLifetime = 24 * time.Hour
	}
	if c.CleanupInterval == 0 {
		c.CleanupInterval = 1 * time.Hour
	}
	if c.FilesystemPath == "" {
		c.FilesystemPath = "./sessions"
	}
	if len(c.EncryptionKey) == 0 {
		// Generate a random 32-byte key for AES-256
		// WARNING: Using a random key means sessions won't persist across restarts
		// For production, provide a fixed encryption key
		c.EncryptionKey = make([]byte, 32)
		if _, err := rand.Read(c.EncryptionKey); err != nil {
			// Fallback to a deterministic key if random generation fails
			// This should never happen in practice
			copy(c.EncryptionKey, []byte("default-insecure-key-change-me!!"))
		}
	}
}

// ApplyDefaults applies default values to MonitoringConfig for any zero-valued fields
// Default: MetricsPort=9090, PprofPort=6060, SNMPPort=161, SNMPCommunity="public",
// OptimizationInterval=5m
func (c *MonitoringConfig) ApplyDefaults() {
	if c.MetricsPort == 0 {
		c.MetricsPort = 9090
	}
	if c.PprofPort == 0 {
		c.PprofPort = 6060
	}
	if c.SNMPPort == 0 {
		c.SNMPPort = 161
	}
	if c.SNMPCommunity == "" {
		c.SNMPCommunity = "public"
	}
	if c.OptimizationInterval == 0 {
		c.OptimizationInterval = 5 * time.Minute
	}
}
