package pkg

import (
	"testing"
	"time"
)

func TestSessionStorage(t *testing.T) {
	// Create a mock database manager
	db := NewDatabaseManager()

	// Connect to mock database
	err := db.Connect(DatabaseConfig{Driver: "mock"})
	if err != nil {
		t.Fatalf("Failed to connect to mock database: %v", err)
	}
	defer db.Close()

	// Create a test session
	session := &Session{
		ID:        "test-session-123",
		UserID:    "user-456",
		TenantID:  "tenant-789",
		Data:      map[string]interface{}{"username": "testuser", "role": "admin"},
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IPAddress: "127.0.0.1",
		UserAgent: "Test Agent",
	}

	// Test SaveSession
	err = db.SaveSession(session)
	if err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	// Test LoadSession
	loadedSession, err := db.LoadSession(session.ID)
	if err != nil {
		t.Fatalf("Failed to load session: %v", err)
	}

	// Verify session data
	if loadedSession.ID != session.ID {
		t.Errorf("Session ID mismatch: expected %s, got %s", session.ID, loadedSession.ID)
	}

	if loadedSession.UserID != session.UserID {
		t.Errorf("User ID mismatch: expected %s, got %s", session.UserID, loadedSession.UserID)
	}

	if loadedSession.TenantID != session.TenantID {
		t.Errorf("Tenant ID mismatch: expected %s, got %s", session.TenantID, loadedSession.TenantID)
	}

	if loadedSession.Data["username"] != "testuser" {
		t.Errorf("Session data mismatch: expected 'testuser', got %v", loadedSession.Data["username"])
	}

	// Test updating session
	loadedSession.Data["last_activity"] = time.Now()
	err = db.SaveSession(loadedSession)
	if err != nil {
		t.Fatalf("Failed to update session: %v", err)
	}

	// Test DeleteSession
	err = db.DeleteSession(session.ID)
	if err != nil {
		t.Fatalf("Failed to delete session: %v", err)
	}

	// Verify session is deleted
	_, err = db.LoadSession(session.ID)
	if err == nil {
		t.Error("Expected error when loading deleted session, got nil")
	}
}

func TestSessionExpiration(t *testing.T) {
	db := NewDatabaseManager()
	db.Connect(DatabaseConfig{Driver: "mock"})
	defer db.Close()

	// Create an expired session
	expiredSession := &Session{
		ID:        "expired-session-123",
		UserID:    "user-456",
		TenantID:  "tenant-789",
		Data:      map[string]interface{}{"test": "data"},
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
		CreatedAt: time.Now().Add(-2 * time.Hour),
		UpdatedAt: time.Now().Add(-2 * time.Hour),
	}

	// Save expired session
	err := db.SaveSession(expiredSession)
	if err != nil {
		t.Fatalf("Failed to save expired session: %v", err)
	}

	// Try to load expired session - should fail
	_, err = db.LoadSession(expiredSession.ID)
	if err == nil {
		t.Error("Expected error when loading expired session, got nil")
	}
}

func TestCleanupExpiredSessions(t *testing.T) {
	db := NewDatabaseManager()
	db.Connect(DatabaseConfig{Driver: "mock"})
	defer db.Close()

	// Create multiple sessions with different expiration times
	sessions := []*Session{
		{
			ID:        "session-1",
			UserID:    "user-1",
			TenantID:  "tenant-1",
			Data:      map[string]interface{}{},
			ExpiresAt: time.Now().Add(1 * time.Hour), // Valid
		},
		{
			ID:        "session-2",
			UserID:    "user-2",
			TenantID:  "tenant-1",
			Data:      map[string]interface{}{},
			ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired
		},
		{
			ID:        "session-3",
			UserID:    "user-3",
			TenantID:  "tenant-1",
			Data:      map[string]interface{}{},
			ExpiresAt: time.Now().Add(-2 * time.Hour), // Expired
		},
	}

	// Save all sessions
	for _, session := range sessions {
		if err := db.SaveSession(session); err != nil {
			t.Fatalf("Failed to save session %s: %v", session.ID, err)
		}
	}

	// Run cleanup
	err := db.CleanupExpiredSessions()
	if err != nil {
		t.Fatalf("Failed to cleanup expired sessions: %v", err)
	}

	// Verify valid session still exists
	_, err = db.LoadSession("session-1")
	if err != nil {
		t.Errorf("Valid session should still exist: %v", err)
	}

	// Verify expired sessions are gone
	_, err = db.LoadSession("session-2")
	if err == nil {
		t.Error("Expired session-2 should have been cleaned up")
	}

	_, err = db.LoadSession("session-3")
	if err == nil {
		t.Error("Expired session-3 should have been cleaned up")
	}
}

func TestMultiTenantSessionIsolation(t *testing.T) {
	db := NewDatabaseManager()
	db.Connect(DatabaseConfig{Driver: "mock"})
	defer db.Close()

	// Create sessions for different tenants
	tenant1Session := &Session{
		ID:        "tenant1-session",
		UserID:    "user-1",
		TenantID:  "tenant-1",
		Data:      map[string]interface{}{"tenant": "tenant-1"},
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	tenant2Session := &Session{
		ID:        "tenant2-session",
		UserID:    "user-2",
		TenantID:  "tenant-2",
		Data:      map[string]interface{}{"tenant": "tenant-2"},
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	// Save both sessions
	if err := db.SaveSession(tenant1Session); err != nil {
		t.Fatalf("Failed to save tenant1 session: %v", err)
	}

	if err := db.SaveSession(tenant2Session); err != nil {
		t.Fatalf("Failed to save tenant2 session: %v", err)
	}

	// Load and verify tenant isolation
	loaded1, err := db.LoadSession(tenant1Session.ID)
	if err != nil {
		t.Fatalf("Failed to load tenant1 session: %v", err)
	}

	if loaded1.TenantID != "tenant-1" {
		t.Errorf("Tenant1 session has wrong tenant ID: %s", loaded1.TenantID)
	}

	loaded2, err := db.LoadSession(tenant2Session.ID)
	if err != nil {
		t.Fatalf("Failed to load tenant2 session: %v", err)
	}

	if loaded2.TenantID != "tenant-2" {
		t.Errorf("Tenant2 session has wrong tenant ID: %s", loaded2.TenantID)
	}
}
