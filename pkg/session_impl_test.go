//go:build test
// +build test

package pkg

import (
	"fmt"
	"time"
)

// Mock session storage for testing
type mockSessionStorage struct {
	sessions map[string]*Session
}

func newMockSessionStorage() *mockSessionStorage {
	return &mockSessionStorage{
		sessions: make(map[string]*Session),
	}
}

// SaveSession saves a session to mock storage
func (mss *mockSessionStorage) SaveSession(session *Session) error {
	if session == nil {
		return fmt.Errorf("session is nil")
	}

	now := time.Now()
	if session.CreatedAt.IsZero() {
		session.CreatedAt = now
	}
	session.UpdatedAt = now

	// Deep copy the session to avoid reference issues
	sessionCopy := &Session{
		ID:        session.ID,
		UserID:    session.UserID,
		TenantID:  session.TenantID,
		Data:      make(map[string]interface{}),
		ExpiresAt: session.ExpiresAt,
		CreatedAt: session.CreatedAt,
		UpdatedAt: session.UpdatedAt,
		IPAddress: session.IPAddress,
		UserAgent: session.UserAgent,
	}

	// Deep copy data
	for k, v := range session.Data {
		sessionCopy.Data[k] = v
	}

	mss.sessions[session.ID] = sessionCopy
	return nil
}

// LoadSession loads a session from mock storage
func (mss *mockSessionStorage) LoadSession(sessionID string) (*Session, error) {
	session, exists := mss.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found or expired")
	}

	// Check if expired
	if session.ExpiresAt.Before(time.Now()) {
		delete(mss.sessions, sessionID)
		return nil, fmt.Errorf("session not found or expired")
	}

	// Return a copy to avoid reference issues
	sessionCopy := &Session{
		ID:        session.ID,
		UserID:    session.UserID,
		TenantID:  session.TenantID,
		Data:      make(map[string]interface{}),
		ExpiresAt: session.ExpiresAt,
		CreatedAt: session.CreatedAt,
		UpdatedAt: session.UpdatedAt,
		IPAddress: session.IPAddress,
		UserAgent: session.UserAgent,
	}

	// Deep copy data
	for k, v := range session.Data {
		sessionCopy.Data[k] = v
	}

	return sessionCopy, nil
}

// DeleteSession deletes a session from mock storage
func (mss *mockSessionStorage) DeleteSession(sessionID string) error {
	delete(mss.sessions, sessionID)
	return nil
}

// CleanupExpiredSessions removes expired sessions from mock storage
func (mss *mockSessionStorage) CleanupExpiredSessions() error {
	now := time.Now()
	for id, session := range mss.sessions {
		if session.ExpiresAt.Before(now) {
			delete(mss.sessions, id)
		}
	}
	return nil
}

// Mock database manager implementation for testing
var mockSessionStore = newMockSessionStorage()

// SaveSession saves a session (mock implementation)
func (dm *databaseManager) SaveSession(session *Session) error {
	return mockSessionStore.SaveSession(session)
}

// LoadSession loads a session (mock implementation)
func (dm *databaseManager) LoadSession(sessionID string) (*Session, error) {
	return mockSessionStore.LoadSession(sessionID)
}

// DeleteSession deletes a session (mock implementation)
func (dm *databaseManager) DeleteSession(sessionID string) error {
	return mockSessionStore.DeleteSession(sessionID)
}

// CleanupExpiredSessions removes expired sessions (mock implementation)
func (dm *databaseManager) CleanupExpiredSessions() error {
	return mockSessionStore.CleanupExpiredSessions()
}
