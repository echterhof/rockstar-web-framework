package pkg

import (
	"strings"
	"testing"
	"time"
)

func TestGenerateSecureToken(t *testing.T) {
	auth := NewAuthManager(nil, "test-secret-key-32-bytes-long!", OAuth2Config{})

	// Generate multiple tokens
	tokens := make(map[string]bool)
	numTokens := 1000

	for i := 0; i < numTokens; i++ {
		token := auth.generateSecureToken()

		// Check token is not empty
		if token == "" {
			t.Fatal("Generated token is empty")
		}

		// Check token length (32 bytes base64-encoded should be 43 characters without padding)
		if len(token) != 43 {
			t.Errorf("Expected token length 43, got %d", len(token))
		}

		// Check for uniqueness
		if tokens[token] {
			t.Fatalf("Generated duplicate token: %s", token)
		}
		tokens[token] = true
	}

	t.Logf("Successfully generated %d unique tokens", numTokens)
}

func TestGenerateSecureTokenUniqueness(t *testing.T) {
	auth := NewAuthManager(nil, "test-secret-key-32-bytes-long!", OAuth2Config{})

	// Generate tokens rapidly to test for time-based collisions
	token1 := auth.generateSecureToken()
	token2 := auth.generateSecureToken()
	token3 := auth.generateSecureToken()

	if token1 == token2 || token1 == token3 || token2 == token3 {
		t.Fatal("Generated duplicate tokens in rapid succession")
	}
}

func TestCreateAccessToken(t *testing.T) {
	// Use mock database for testing
	db := NewMockDatabaseManager()
	err := db.Connect(DatabaseConfig{Driver: "mock"})
	if err != nil {
		t.Fatalf("Failed to connect mock database: %v", err)
	}
	defer db.Close()

	auth := NewAuthManager(db, "test-secret-key-32-bytes-long!", OAuth2Config{})

	// Create access token
	token, err := auth.CreateAccessToken("user123", "tenant456", []string{"read", "write"}, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to create access token: %v", err)
	}

	// Verify token properties
	if token.Token == "" {
		t.Error("Token value is empty")
	}

	if token.UserID != "user123" {
		t.Errorf("Expected UserID 'user123', got '%s'", token.UserID)
	}

	if token.TenantID != "tenant456" {
		t.Errorf("Expected TenantID 'tenant456', got '%s'", token.TenantID)
	}

	if len(token.Scopes) != 2 {
		t.Errorf("Expected 2 scopes, got %d", len(token.Scopes))
	}

	// Verify expiration is in the future
	if token.ExpiresAt.Before(time.Now()) {
		t.Error("Token expiration is in the past")
	}

	// Verify expiration is approximately 1 hour from now
	expectedExpiry := time.Now().Add(1 * time.Hour)
	diff := token.ExpiresAt.Sub(expectedExpiry)
	if diff < -1*time.Second || diff > 1*time.Second {
		t.Errorf("Token expiration is not approximately 1 hour from now: %v", diff)
	}
}

func TestGenerateJWT(t *testing.T) {
	auth := NewAuthManager(nil, "test-secret-key-32-bytes-long!", OAuth2Config{})

	user := &User{
		ID:       "user123",
		Username: "testuser",
		Email:    "test@example.com",
		Roles:    []string{"admin", "user"},
		Actions:  []string{"read", "write"},
		Scopes:   []string{"api:read", "api:write"},
		TenantID: "tenant456",
	}

	// Generate JWT
	token, err := auth.GenerateJWT(user, 24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate JWT: %v", err)
	}

	if token == "" {
		t.Fatal("Generated JWT is empty")
	}

	// JWT should have 3 parts separated by dots
	parts := len(token) - len(replaceAll(token, ".", ""))
	if parts != 2 {
		t.Errorf("Expected JWT to have 3 parts (2 dots), got %d dots", parts)
	}

	// Verify JWT can be parsed back
	parsedUser, err := auth.AuthenticateJWT(token)
	if err != nil {
		t.Fatalf("Failed to authenticate JWT: %v", err)
	}

	if parsedUser.ID != user.ID {
		t.Errorf("Expected user ID '%s', got '%s'", user.ID, parsedUser.ID)
	}

	if parsedUser.Username != user.Username {
		t.Errorf("Expected username '%s', got '%s'", user.Username, parsedUser.Username)
	}

	if parsedUser.Email != user.Email {
		t.Errorf("Expected email '%s', got '%s'", user.Email, parsedUser.Email)
	}
}

func TestAuthenticateJWTExpired(t *testing.T) {
	auth := NewAuthManager(nil, "test-secret-key-32-bytes-long!", OAuth2Config{})

	user := &User{
		ID:       "user123",
		Username: "testuser",
		Email:    "test@example.com",
	}

	// Generate JWT with negative expiration (already expired)
	token, err := auth.GenerateJWT(user, -1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate JWT: %v", err)
	}

	// Try to authenticate expired token
	_, err = auth.AuthenticateJWT(token)
	if err == nil {
		t.Fatal("Expected error for expired JWT, got nil")
	}

	// Check error message contains "expired"
	errMsg := err.Error()
	if !strings.Contains(errMsg, "expired") {
		t.Errorf("Expected error message to contain 'expired', got: %v", errMsg)
	}
}

func TestAuthenticateJWTInvalidSignature(t *testing.T) {
	auth := NewAuthManager(nil, "test-secret-key-32-bytes-long!", OAuth2Config{})

	user := &User{
		ID:       "user123",
		Username: "testuser",
	}

	// Generate JWT
	token, err := auth.GenerateJWT(user, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate JWT: %v", err)
	}

	// Tamper with the token (change last character)
	tamperedToken := token[:len(token)-1] + "X"

	// Try to authenticate tampered token
	_, err = auth.AuthenticateJWT(tamperedToken)
	if err == nil {
		t.Fatal("Expected error for tampered JWT, got nil")
	}
}

// Helper function to replace all occurrences
func replaceAll(s, old, new string) string {
	result := ""
	for _, c := range s {
		if string(c) == old {
			result += new
		} else {
			result += string(c)
		}
	}
	return result
}
