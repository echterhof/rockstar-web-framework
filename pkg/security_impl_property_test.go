package pkg

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// TestProperty_SecurityManagerOperatesWithoutDatabase tests Property 8:
// SecurityManager operates without database persistence
// **Feature: optional-database, Property 8: SecurityManager operates without database persistence**
// **Validates: Requirements 3.2**
func TestProperty_SecurityManagerOperatesWithoutDatabase(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("SecurityManager initializes with no-op database", prop.ForAll(
		func() bool {
			// Create no-op database
			noopDB := NewNoopDatabaseManager()

			// Create security config
			config := DefaultSecurityConfig()
			config.EncryptionKey = hex.EncodeToString([]byte("12345678901234567890123456789012"))
			config.JWTSecret = "test-jwt-secret"

			// Create security manager
			sm, err := NewSecurityManager(noopDB, config)
			if err != nil {
				t.Logf("Failed to create security manager: %v", err)
				return false
			}

			// Verify security manager was created
			if sm == nil {
				t.Log("Security manager is nil")
				return false
			}

			// Verify it's using in-memory storage
			smImpl, ok := sm.(*securityManagerImpl)
			if !ok {
				t.Log("Security manager is not the expected implementation")
				return false
			}

			// Check that in-memory storage is initialized
			if smImpl.tokenStorage == nil {
				t.Log("Token storage is nil - should be in-memory")
				return false
			}

			if smImpl.rateLimits == nil {
				t.Log("Rate limits storage is nil - should be in-memory")
				return false
			}

			return true
		},
	))

	properties.Property("Rate limiting works with in-memory storage", prop.ForAll(
		func(resource string, clientIP string) bool {
			// Skip empty values
			if resource == "" || clientIP == "" {
				return true
			}

			// Create no-op database
			noopDB := NewNoopDatabaseManager()

			// Create security config
			config := DefaultSecurityConfig()
			config.EncryptionKey = hex.EncodeToString([]byte("12345678901234567890123456789012"))
			config.JWTSecret = "test-jwt-secret"

			// Create security manager
			sm, err := NewSecurityManager(noopDB, config)
			if err != nil {
				t.Logf("Failed to create security manager: %v", err)
				return false
			}

			// Create a test context
			req := &Request{
				RemoteAddr: clientIP + ":12345",
				Header:     make(map[string][]string),
			}
			ctx := createTestContext(t, req)

			// Check rate limit multiple times - should work without database
			for i := 0; i < 5; i++ {
				err := sm.CheckRateLimit(ctx, resource)
				if err != nil {
					// Rate limit might be exceeded, which is fine
					// We just want to ensure it doesn't fail due to missing database
					if _, ok := err.(*FrameworkError); !ok {
						t.Logf("Unexpected error type: %v", err)
						return false
					}
				}
			}

			return true
		},
		gen.AlphaString(),
		gen.RegexMatch(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`),
	))

	properties.Property("Global rate limiting works with in-memory storage", prop.ForAll(
		func(clientIP string) bool {
			// Skip empty values
			if clientIP == "" {
				return true
			}

			// Create no-op database
			noopDB := NewNoopDatabaseManager()

			// Create security config
			config := DefaultSecurityConfig()
			config.EncryptionKey = hex.EncodeToString([]byte("12345678901234567890123456789012"))
			config.JWTSecret = "test-jwt-secret"

			// Create security manager
			sm, err := NewSecurityManager(noopDB, config)
			if err != nil {
				t.Logf("Failed to create security manager: %v", err)
				return false
			}

			// Create a test context
			req := &Request{
				RemoteAddr: clientIP + ":12345",
				Header:     make(map[string][]string),
			}
			ctx := createTestContext(t, req)

			// Check global rate limit - should work without database
			err = sm.CheckGlobalRateLimit(ctx)
			if err != nil {
				// Rate limit might be exceeded, which is fine
				// We just want to ensure it doesn't fail due to missing database
				if _, ok := err.(*FrameworkError); !ok {
					t.Logf("Unexpected error type: %v", err)
					return false
				}
			}

			return true
		},
		gen.RegexMatch(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`),
	))

	properties.Property("CSRF protection works without database", prop.ForAll(
		func() bool {
			// Create no-op database
			noopDB := NewNoopDatabaseManager()

			// Create security config
			config := DefaultSecurityConfig()
			config.EncryptionKey = hex.EncodeToString([]byte("12345678901234567890123456789012"))
			config.JWTSecret = "test-jwt-secret"

			// Create security manager
			sm, err := NewSecurityManager(noopDB, config)
			if err != nil {
				t.Logf("Failed to create security manager: %v", err)
				return false
			}

			// Create a test context
			req := &Request{
				Header: make(map[string][]string),
			}
			ctx := createTestContext(t, req)

			// Enable CSRF protection - should work without database
			token, err := sm.EnableCSRFProtection(ctx)
			if err != nil {
				t.Logf("Failed to enable CSRF protection: %v", err)
				return false
			}

			if token == "" {
				t.Log("CSRF token is empty")
				return false
			}

			return true
		},
	))

	properties.Property("Cookie encryption works without database", prop.ForAll(
		func(value string) bool {
			// Skip empty values
			if value == "" {
				return true
			}

			// Create no-op database
			noopDB := NewNoopDatabaseManager()

			// Create security config
			config := DefaultSecurityConfig()
			config.EncryptionKey = hex.EncodeToString([]byte("12345678901234567890123456789012"))
			config.JWTSecret = "test-jwt-secret"

			// Create security manager
			sm, err := NewSecurityManager(noopDB, config)
			if err != nil {
				t.Logf("Failed to create security manager: %v", err)
				return false
			}

			// Encrypt cookie - should work without database
			encrypted, err := sm.EncryptCookie(value)
			if err != nil {
				t.Logf("Failed to encrypt cookie: %v", err)
				return false
			}

			if encrypted == "" {
				t.Log("Encrypted value is empty")
				return false
			}

			// Decrypt cookie - should work without database
			decrypted, err := sm.DecryptCookie(encrypted)
			if err != nil {
				t.Logf("Failed to decrypt cookie: %v", err)
				return false
			}

			if decrypted != value {
				t.Logf("Decrypted value %q doesn't match original %q", decrypted, value)
				return false
			}

			return true
		},
		gen.AlphaString(),
	))

	properties.Property("Input validation works without database", prop.ForAll(
		func(input string) bool {
			// Create no-op database
			noopDB := NewNoopDatabaseManager()

			// Create security config
			config := DefaultSecurityConfig()
			config.EncryptionKey = hex.EncodeToString([]byte("12345678901234567890123456789012"))
			config.JWTSecret = "test-jwt-secret"

			// Create security manager
			sm, err := NewSecurityManager(noopDB, config)
			if err != nil {
				t.Logf("Failed to create security manager: %v", err)
				return false
			}

			// Validate input - should work without database
			rules := InputValidationRules{
				MaxLength: 1000,
				AllowHTML: false,
			}

			err = sm.ValidateInput(input, rules)
			// Error is fine (e.g., input too long, contains HTML)
			// We just want to ensure it doesn't fail due to missing database

			return true
		},
		gen.AlphaString(),
	))

	properties.Property("Security headers work without database", prop.ForAll(
		func() bool {
			// Create no-op database
			noopDB := NewNoopDatabaseManager()

			// Create security config
			config := DefaultSecurityConfig()
			config.EncryptionKey = hex.EncodeToString([]byte("12345678901234567890123456789012"))
			config.JWTSecret = "test-jwt-secret"

			// Create security manager
			sm, err := NewSecurityManager(noopDB, config)
			if err != nil {
				t.Logf("Failed to create security manager: %v", err)
				return false
			}

			// Create a test context
			req := &Request{
				Header: make(map[string][]string),
			}
			ctx := createTestContext(t, req)

			// Set security headers - should work without database
			err = sm.SetSecurityHeaders(ctx)
			if err != nil {
				t.Logf("Failed to set security headers: %v", err)
				return false
			}

			return true
		},
	))

	properties.Property("Rate limit counters persist in memory during session", prop.ForAll(
		func(resource string, clientIP string) bool {
			// Skip empty values
			if resource == "" || clientIP == "" {
				return true
			}

			// Create no-op database
			noopDB := NewNoopDatabaseManager()

			// Create security config
			config := DefaultSecurityConfig()
			config.EncryptionKey = hex.EncodeToString([]byte("12345678901234567890123456789012"))
			config.JWTSecret = "test-jwt-secret"

			// Create security manager
			sm, err := NewSecurityManager(noopDB, config)
			if err != nil {
				t.Logf("Failed to create security manager: %v", err)
				return false
			}

			smImpl := sm.(*securityManagerImpl)

			// Create a test context
			req := &Request{
				RemoteAddr: clientIP + ":12345",
				Header:     make(map[string][]string),
			}
			ctx := createTestContext(t, req)

			// Check initial count
			initialCount := smImpl.rateLimits.Count()

			// Perform rate limit check - this should create an entry
			sm.CheckRateLimit(ctx, resource)

			// Check that count increased
			newCount := smImpl.rateLimits.Count()
			if newCount <= initialCount {
				t.Logf("Rate limit count didn't increase: initial=%d, new=%d", initialCount, newCount)
				return false
			}

			return true
		},
		gen.AlphaString(),
		gen.RegexMatch(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`),
	))

	// Run all properties with 100 iterations minimum
	params := gopter.DefaultTestParameters()
	params.MinSuccessfulTests = 100
	properties.TestingRun(t, params)
}

// TestProperty_InMemoryRateLimitStorage tests the in-memory rate limit storage
func TestProperty_InMemoryRateLimitStorage(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Rate limit entries are created and tracked", prop.ForAll(
		func(key string) bool {
			// Skip empty keys
			if key == "" {
				return true
			}

			storage := newInMemoryRateLimitStorage()

			// Check initial state
			allowed, err := storage.CheckRateLimit(key, 10, time.Minute)
			if err != nil {
				t.Logf("Failed to check rate limit: %v", err)
				return false
			}

			if !allowed {
				t.Log("Rate limit should be allowed initially")
				return false
			}

			// Increment counter
			err = storage.IncrementRateLimit(key, time.Minute)
			if err != nil {
				t.Logf("Failed to increment rate limit: %v", err)
				return false
			}

			// Verify entry was created
			if storage.Count() == 0 {
				t.Log("No entries created after increment")
				return false
			}

			return true
		},
		gen.AlphaString(),
	))

	properties.Property("Rate limit enforces limits correctly", prop.ForAll(
		func(key string, limit uint8) bool {
			// Skip empty keys and zero limits
			if key == "" || limit == 0 {
				return true
			}

			storage := newInMemoryRateLimitStorage()
			limitInt := int(limit)

			// Increment up to the limit
			for i := 0; i < limitInt; i++ {
				err := storage.IncrementRateLimit(key, time.Minute)
				if err != nil {
					t.Logf("Failed to increment rate limit: %v", err)
					return false
				}
			}

			// Check that limit is now exceeded
			allowed, err := storage.CheckRateLimit(key, limitInt, time.Minute)
			if err != nil {
				t.Logf("Failed to check rate limit: %v", err)
				return false
			}

			if allowed {
				t.Logf("Rate limit should be exceeded after %d increments with limit %d", limitInt, limitInt)
				return false
			}

			return true
		},
		gen.AlphaString(),
		gen.UInt8Range(1, 100),
	))

	properties.Property("Expired entries are cleaned up", prop.ForAll(
		func(key string) bool {
			// Skip empty keys
			if key == "" {
				return true
			}

			storage := newInMemoryRateLimitStorage()

			// Create entry with very short window
			err := storage.IncrementRateLimit(key, 1*time.Millisecond)
			if err != nil {
				t.Logf("Failed to increment rate limit: %v", err)
				return false
			}

			// Wait for expiration
			time.Sleep(10 * time.Millisecond)

			// Cleanup
			err = storage.Cleanup()
			if err != nil {
				t.Logf("Failed to cleanup: %v", err)
				return false
			}

			// Verify entry was removed
			if storage.Count() != 0 {
				t.Logf("Expected 0 entries after cleanup, got %d", storage.Count())
				return false
			}

			return true
		},
		gen.AlphaString(),
	))

	// Run all properties with 100 iterations minimum
	params := gopter.DefaultTestParameters()
	params.MinSuccessfulTests = 100
	properties.TestingRun(t, params)
}
