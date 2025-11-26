package pkg

import (
	"testing"
	"time"
)

func TestNewInMemoryTokenStorage(t *testing.T) {
	storage := newInMemoryTokenStorage()
	if storage == nil {
		t.Fatal("expected non-nil storage")
	}
	if storage.tokens == nil {
		t.Fatal("expected initialized tokens map")
	}
	if storage.Count() != 0 {
		t.Errorf("expected empty storage, got %d tokens", storage.Count())
	}
}

func TestInMemoryTokenStorage_Save(t *testing.T) {
	storage := newInMemoryTokenStorage()

	token := &AccessToken{
		Token:     "test-token-123",
		UserID:    "user-1",
		TenantID:  "tenant-1",
		Scopes:    []string{"read", "write"},
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
	}

	err := storage.Save(token)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if storage.Count() != 1 {
		t.Errorf("expected 1 token, got %d", storage.Count())
	}
}

func TestInMemoryTokenStorage_SaveNilToken(t *testing.T) {
	storage := newInMemoryTokenStorage()

	err := storage.Save(nil)
	if err == nil {
		t.Fatal("expected error for nil token")
	}
	if err.Error() != "token is nil" {
		t.Errorf("expected 'token is nil' error, got %v", err)
	}
}

func TestInMemoryTokenStorage_SaveEmptyTokenValue(t *testing.T) {
	storage := newInMemoryTokenStorage()

	token := &AccessToken{
		Token:  "",
		UserID: "user-1",
	}

	err := storage.Save(token)
	if err == nil {
		t.Fatal("expected error for empty token value")
	}
	if err.Error() != "token value is required" {
		t.Errorf("expected 'token value is required' error, got %v", err)
	}
}

func TestInMemoryTokenStorage_Load(t *testing.T) {
	storage := newInMemoryTokenStorage()

	originalToken := &AccessToken{
		Token:     "test-token-456",
		UserID:    "user-2",
		TenantID:  "tenant-2",
		Scopes:    []string{"admin"},
		ExpiresAt: time.Now().Add(2 * time.Hour),
		CreatedAt: time.Now(),
	}

	err := storage.Save(originalToken)
	if err != nil {
		t.Fatalf("failed to save token: %v", err)
	}

	loadedToken, err := storage.Load("test-token-456")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if loadedToken.Token != originalToken.Token {
		t.Errorf("expected token %s, got %s", originalToken.Token, loadedToken.Token)
	}
	if loadedToken.UserID != originalToken.UserID {
		t.Errorf("expected user ID %s, got %s", originalToken.UserID, loadedToken.UserID)
	}
	if loadedToken.TenantID != originalToken.TenantID {
		t.Errorf("expected tenant ID %s, got %s", originalToken.TenantID, loadedToken.TenantID)
	}
	if len(loadedToken.Scopes) != len(originalToken.Scopes) {
		t.Errorf("expected %d scopes, got %d", len(originalToken.Scopes), len(loadedToken.Scopes))
	}
}

func TestInMemoryTokenStorage_LoadNotFound(t *testing.T) {
	storage := newInMemoryTokenStorage()

	_, err := storage.Load("non-existent-token")
	if err == nil {
		t.Fatal("expected error for non-existent token")
	}
	if err.Error() != "token not found" {
		t.Errorf("expected 'token not found' error, got %v", err)
	}
}

func TestInMemoryTokenStorage_LoadEmptyTokenValue(t *testing.T) {
	storage := newInMemoryTokenStorage()

	_, err := storage.Load("")
	if err == nil {
		t.Fatal("expected error for empty token value")
	}
	if err.Error() != "token value is required" {
		t.Errorf("expected 'token value is required' error, got %v", err)
	}
}

func TestInMemoryTokenStorage_Validate(t *testing.T) {
	storage := newInMemoryTokenStorage()

	token := &AccessToken{
		Token:     "valid-token",
		UserID:    "user-3",
		TenantID:  "tenant-3",
		Scopes:    []string{"read"},
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
	}

	err := storage.Save(token)
	if err != nil {
		t.Fatalf("failed to save token: %v", err)
	}

	validatedToken, err := storage.Validate("valid-token")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if validatedToken.Token != token.Token {
		t.Errorf("expected token %s, got %s", token.Token, validatedToken.Token)
	}
}

func TestInMemoryTokenStorage_ValidateExpired(t *testing.T) {
	storage := newInMemoryTokenStorage()

	expiredToken := &AccessToken{
		Token:     "expired-token",
		UserID:    "user-4",
		TenantID:  "tenant-4",
		Scopes:    []string{"read"},
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
		CreatedAt: time.Now().Add(-2 * time.Hour),
	}

	err := storage.Save(expiredToken)
	if err != nil {
		t.Fatalf("failed to save token: %v", err)
	}

	_, err = storage.Validate("expired-token")
	if err == nil {
		t.Fatal("expected error for expired token")
	}
	if err.Error() != "token expired" {
		t.Errorf("expected 'token expired' error, got %v", err)
	}
}

func TestInMemoryTokenStorage_ValidateNotFound(t *testing.T) {
	storage := newInMemoryTokenStorage()

	_, err := storage.Validate("non-existent-token")
	if err == nil {
		t.Fatal("expected error for non-existent token")
	}
	if err.Error() != "token not found" {
		t.Errorf("expected 'token not found' error, got %v", err)
	}
}

func TestInMemoryTokenStorage_ValidateEmptyTokenValue(t *testing.T) {
	storage := newInMemoryTokenStorage()

	_, err := storage.Validate("")
	if err == nil {
		t.Fatal("expected error for empty token value")
	}
	if err.Error() != "token value is required" {
		t.Errorf("expected 'token value is required' error, got %v", err)
	}
}

func TestInMemoryTokenStorage_Delete(t *testing.T) {
	storage := newInMemoryTokenStorage()

	token := &AccessToken{
		Token:     "delete-me",
		UserID:    "user-5",
		TenantID:  "tenant-5",
		Scopes:    []string{"write"},
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
	}

	err := storage.Save(token)
	if err != nil {
		t.Fatalf("failed to save token: %v", err)
	}

	if storage.Count() != 1 {
		t.Errorf("expected 1 token before delete, got %d", storage.Count())
	}

	err = storage.Delete("delete-me")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if storage.Count() != 0 {
		t.Errorf("expected 0 tokens after delete, got %d", storage.Count())
	}

	_, err = storage.Load("delete-me")
	if err == nil {
		t.Fatal("expected error loading deleted token")
	}
}

func TestInMemoryTokenStorage_DeleteEmptyTokenValue(t *testing.T) {
	storage := newInMemoryTokenStorage()

	err := storage.Delete("")
	if err == nil {
		t.Fatal("expected error for empty token value")
	}
	if err.Error() != "token value is required" {
		t.Errorf("expected 'token value is required' error, got %v", err)
	}
}

func TestInMemoryTokenStorage_DeleteNonExistent(t *testing.T) {
	storage := newInMemoryTokenStorage()

	// Deleting non-existent token should not error
	err := storage.Delete("non-existent")
	if err != nil {
		t.Fatalf("expected no error for deleting non-existent token, got %v", err)
	}
}

func TestInMemoryTokenStorage_Cleanup(t *testing.T) {
	storage := newInMemoryTokenStorage()

	// Add valid token
	validToken := &AccessToken{
		Token:     "valid-token",
		UserID:    "user-6",
		TenantID:  "tenant-6",
		Scopes:    []string{"read"},
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
	}

	// Add expired token
	expiredToken := &AccessToken{
		Token:     "expired-token",
		UserID:    "user-7",
		TenantID:  "tenant-7",
		Scopes:    []string{"write"},
		ExpiresAt: time.Now().Add(-1 * time.Hour),
		CreatedAt: time.Now().Add(-2 * time.Hour),
	}

	err := storage.Save(validToken)
	if err != nil {
		t.Fatalf("failed to save valid token: %v", err)
	}

	err = storage.Save(expiredToken)
	if err != nil {
		t.Fatalf("failed to save expired token: %v", err)
	}

	if storage.Count() != 2 {
		t.Errorf("expected 2 tokens before cleanup, got %d", storage.Count())
	}

	err = storage.Cleanup()
	if err != nil {
		t.Fatalf("expected no error during cleanup, got %v", err)
	}

	if storage.Count() != 1 {
		t.Errorf("expected 1 token after cleanup, got %d", storage.Count())
	}

	// Verify valid token still exists
	_, err = storage.Load("valid-token")
	if err != nil {
		t.Errorf("expected valid token to still exist after cleanup, got error: %v", err)
	}

	// Verify expired token was removed
	_, err = storage.Load("expired-token")
	if err == nil {
		t.Error("expected expired token to be removed after cleanup")
	}
}

func TestInMemoryTokenStorage_DeepCopy(t *testing.T) {
	storage := newInMemoryTokenStorage()

	originalScopes := []string{"read", "write", "admin"}
	token := &AccessToken{
		Token:     "test-copy",
		UserID:    "user-8",
		TenantID:  "tenant-8",
		Scopes:    originalScopes,
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
	}

	err := storage.Save(token)
	if err != nil {
		t.Fatalf("failed to save token: %v", err)
	}

	// Modify original token's scopes
	token.Scopes[0] = "modified"

	// Load token and verify it wasn't affected by external modification
	loadedToken, err := storage.Load("test-copy")
	if err != nil {
		t.Fatalf("failed to load token: %v", err)
	}

	if loadedToken.Scopes[0] != "read" {
		t.Errorf("expected scope 'read', got '%s' - deep copy failed", loadedToken.Scopes[0])
	}
}

func TestInMemoryTokenStorage_ConcurrentAccess(t *testing.T) {
	storage := newInMemoryTokenStorage()

	// Test concurrent writes
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			token := &AccessToken{
				Token:     string(rune('A' + id)),
				UserID:    "user",
				TenantID:  "tenant",
				Scopes:    []string{"read"},
				ExpiresAt: time.Now().Add(1 * time.Hour),
				CreatedAt: time.Now(),
			}
			storage.Save(token)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Test concurrent reads
	for i := 0; i < 10; i++ {
		go func(id int) {
			storage.Load(string(rune('A' + id)))
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify count
	if storage.Count() != 10 {
		t.Errorf("expected 10 tokens after concurrent access, got %d", storage.Count())
	}
}
