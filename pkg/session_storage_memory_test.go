package pkg

import (
	"sync"
	"testing"
	"time"
)

func TestInMemorySessionStorage_Save(t *testing.T) {
	storage := newInMemorySessionStorage()

	session := &Session{
		ID:        "test-session-1",
		UserID:    "user-1",
		TenantID:  "tenant-1",
		Data:      map[string]interface{}{"key": "value"},
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IPAddress: "127.0.0.1",
		UserAgent: "test-agent",
	}

	err := storage.Save(session)
	if err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	if storage.Count() != 1 {
		t.Errorf("Expected 1 session, got %d", storage.Count())
	}
}

func TestInMemorySessionStorage_SaveNilSession(t *testing.T) {
	storage := newInMemorySessionStorage()

	err := storage.Save(nil)
	if err == nil {
		t.Error("Expected error when saving nil session")
	}
	if err.Error() != "session is nil" {
		t.Errorf("Expected 'session is nil' error, got: %v", err)
	}
}

func TestInMemorySessionStorage_SaveEmptyID(t *testing.T) {
	storage := newInMemorySessionStorage()

	session := &Session{
		ID:        "",
		Data:      map[string]interface{}{},
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	err := storage.Save(session)
	if err == nil {
		t.Error("Expected error when saving session with empty ID")
	}
	if err.Error() != "session ID is required" {
		t.Errorf("Expected 'session ID is required' error, got: %v", err)
	}
}

func TestInMemorySessionStorage_Load(t *testing.T) {
	storage := newInMemorySessionStorage()

	originalSession := &Session{
		ID:        "test-session-2",
		UserID:    "user-2",
		TenantID:  "tenant-2",
		Data:      map[string]interface{}{"foo": "bar"},
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IPAddress: "192.168.1.1",
		UserAgent: "test-agent-2",
	}

	err := storage.Save(originalSession)
	if err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	loadedSession, err := storage.Load("test-session-2")
	if err != nil {
		t.Fatalf("Failed to load session: %v", err)
	}

	if loadedSession.ID != originalSession.ID {
		t.Errorf("Expected ID %s, got %s", originalSession.ID, loadedSession.ID)
	}
	if loadedSession.UserID != originalSession.UserID {
		t.Errorf("Expected UserID %s, got %s", originalSession.UserID, loadedSession.UserID)
	}
	if loadedSession.TenantID != originalSession.TenantID {
		t.Errorf("Expected TenantID %s, got %s", originalSession.TenantID, loadedSession.TenantID)
	}
	if loadedSession.Data["foo"] != "bar" {
		t.Errorf("Expected Data['foo'] = 'bar', got %v", loadedSession.Data["foo"])
	}
}

func TestInMemorySessionStorage_LoadNonExistent(t *testing.T) {
	storage := newInMemorySessionStorage()

	_, err := storage.Load("non-existent")
	if err == nil {
		t.Error("Expected error when loading non-existent session")
	}
	if err.Error() != "session not found" {
		t.Errorf("Expected 'session not found' error, got: %v", err)
	}
}

func TestInMemorySessionStorage_LoadEmptyID(t *testing.T) {
	storage := newInMemorySessionStorage()

	_, err := storage.Load("")
	if err == nil {
		t.Error("Expected error when loading session with empty ID")
	}
	if err.Error() != "session ID is required" {
		t.Errorf("Expected 'session ID is required' error, got: %v", err)
	}
}

func TestInMemorySessionStorage_LoadExpired(t *testing.T) {
	storage := newInMemorySessionStorage()

	expiredSession := &Session{
		ID:        "expired-session",
		UserID:    "user-3",
		Data:      map[string]interface{}{},
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
		CreatedAt: time.Now().Add(-2 * time.Hour),
		UpdatedAt: time.Now().Add(-1 * time.Hour),
	}

	err := storage.Save(expiredSession)
	if err != nil {
		t.Fatalf("Failed to save expired session: %v", err)
	}

	_, err = storage.Load("expired-session")
	if err == nil {
		t.Error("Expected error when loading expired session")
	}
	if err.Error() != "session expired" {
		t.Errorf("Expected 'session expired' error, got: %v", err)
	}
}

func TestInMemorySessionStorage_Delete(t *testing.T) {
	storage := newInMemorySessionStorage()

	session := &Session{
		ID:        "test-session-3",
		Data:      map[string]interface{}{},
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := storage.Save(session)
	if err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	if storage.Count() != 1 {
		t.Errorf("Expected 1 session before delete, got %d", storage.Count())
	}

	err = storage.Delete("test-session-3")
	if err != nil {
		t.Fatalf("Failed to delete session: %v", err)
	}

	if storage.Count() != 0 {
		t.Errorf("Expected 0 sessions after delete, got %d", storage.Count())
	}

	_, err = storage.Load("test-session-3")
	if err == nil {
		t.Error("Expected error when loading deleted session")
	}
}

func TestInMemorySessionStorage_DeleteNonExistent(t *testing.T) {
	storage := newInMemorySessionStorage()

	// Deleting non-existent session should not error
	err := storage.Delete("non-existent")
	if err != nil {
		t.Errorf("Unexpected error when deleting non-existent session: %v", err)
	}
}

func TestInMemorySessionStorage_DeleteEmptyID(t *testing.T) {
	storage := newInMemorySessionStorage()

	err := storage.Delete("")
	if err == nil {
		t.Error("Expected error when deleting session with empty ID")
	}
	if err.Error() != "session ID is required" {
		t.Errorf("Expected 'session ID is required' error, got: %v", err)
	}
}

func TestInMemorySessionStorage_Cleanup(t *testing.T) {
	storage := newInMemorySessionStorage()

	// Add valid session
	validSession := &Session{
		ID:        "valid-session",
		Data:      map[string]interface{}{},
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := storage.Save(validSession)
	if err != nil {
		t.Fatalf("Failed to save valid session: %v", err)
	}

	// Add expired session
	expiredSession := &Session{
		ID:        "expired-session",
		Data:      map[string]interface{}{},
		ExpiresAt: time.Now().Add(-1 * time.Hour),
		CreatedAt: time.Now().Add(-2 * time.Hour),
		UpdatedAt: time.Now().Add(-1 * time.Hour),
	}
	err = storage.Save(expiredSession)
	if err != nil {
		t.Fatalf("Failed to save expired session: %v", err)
	}

	if storage.Count() != 2 {
		t.Errorf("Expected 2 sessions before cleanup, got %d", storage.Count())
	}

	err = storage.Cleanup()
	if err != nil {
		t.Fatalf("Failed to cleanup sessions: %v", err)
	}

	if storage.Count() != 1 {
		t.Errorf("Expected 1 session after cleanup, got %d", storage.Count())
	}

	// Verify valid session still exists
	_, err = storage.Load("valid-session")
	if err != nil {
		t.Errorf("Valid session should still exist after cleanup: %v", err)
	}

	// Verify expired session was removed
	_, err = storage.Load("expired-session")
	if err == nil {
		t.Error("Expired session should have been removed during cleanup")
	}
}

func TestInMemorySessionStorage_DataIsolation(t *testing.T) {
	storage := newInMemorySessionStorage()

	session := &Session{
		ID:        "test-isolation",
		Data:      map[string]interface{}{"key": "original"},
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := storage.Save(session)
	if err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	// Modify the original session data
	session.Data["key"] = "modified"
	session.Data["new_key"] = "new_value"

	// Load the session and verify it wasn't affected by external modifications
	loadedSession, err := storage.Load("test-isolation")
	if err != nil {
		t.Fatalf("Failed to load session: %v", err)
	}

	if loadedSession.Data["key"] != "original" {
		t.Errorf("Expected Data['key'] = 'original', got %v", loadedSession.Data["key"])
	}
	if _, exists := loadedSession.Data["new_key"]; exists {
		t.Error("External modifications should not affect stored session")
	}

	// Modify loaded session and verify storage wasn't affected
	loadedSession.Data["key"] = "changed"

	loadedAgain, err := storage.Load("test-isolation")
	if err != nil {
		t.Fatalf("Failed to load session again: %v", err)
	}

	if loadedAgain.Data["key"] != "original" {
		t.Errorf("Expected Data['key'] = 'original', got %v", loadedAgain.Data["key"])
	}
}

func TestInMemorySessionStorage_ConcurrentAccess(t *testing.T) {
	storage := newInMemorySessionStorage()

	const numGoroutines = 100
	const numOperations = 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				session := &Session{
					ID:        string(rune(id*numOperations + j)),
					Data:      map[string]interface{}{"value": id*numOperations + j},
					ExpiresAt: time.Now().Add(1 * time.Hour),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				storage.Save(session)
			}
		}(i)
	}

	wg.Wait()

	// Verify all sessions were saved
	expectedCount := numGoroutines * numOperations
	if storage.Count() != expectedCount {
		t.Errorf("Expected %d sessions, got %d", expectedCount, storage.Count())
	}

	// Concurrent reads and deletes
	wg.Add(numGoroutines * 2)

	for i := 0; i < numGoroutines; i++ {
		// Readers
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				storage.Load(string(rune(id*numOperations + j)))
			}
		}(i)

		// Deleters
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				storage.Delete(string(rune(id*numOperations + j)))
			}
		}(i)
	}

	wg.Wait()

	// After all deletes, storage should be empty
	if storage.Count() != 0 {
		t.Errorf("Expected 0 sessions after concurrent deletes, got %d", storage.Count())
	}
}

func TestInMemorySessionStorage_UpdateSession(t *testing.T) {
	storage := newInMemorySessionStorage()

	session := &Session{
		ID:        "update-test",
		UserID:    "user-1",
		Data:      map[string]interface{}{"counter": 1},
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := storage.Save(session)
	if err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	// Update the session
	session.Data["counter"] = 2
	session.UserID = "user-2"
	session.UpdatedAt = time.Now()

	err = storage.Save(session)
	if err != nil {
		t.Fatalf("Failed to update session: %v", err)
	}

	// Verify update
	loadedSession, err := storage.Load("update-test")
	if err != nil {
		t.Fatalf("Failed to load updated session: %v", err)
	}

	if loadedSession.UserID != "user-2" {
		t.Errorf("Expected UserID 'user-2', got %s", loadedSession.UserID)
	}
	if loadedSession.Data["counter"] != 2 {
		t.Errorf("Expected Data['counter'] = 2, got %v", loadedSession.Data["counter"])
	}

	// Verify only one session exists (not duplicated)
	if storage.Count() != 1 {
		t.Errorf("Expected 1 session after update, got %d", storage.Count())
	}
}
