package pkg

import (
	"errors"
	"sync"
	"time"
)

// inMemorySessionStorage implements session storage in memory
type inMemorySessionStorage struct {
	mu       sync.RWMutex
	sessions map[string]*Session
}

// newInMemorySessionStorage creates a new in-memory session storage instance
func newInMemorySessionStorage() *inMemorySessionStorage {
	return &inMemorySessionStorage{
		sessions: make(map[string]*Session),
	}
}

// Save saves a session to memory
func (s *inMemorySessionStorage) Save(session *Session) error {
	if session == nil {
		return errors.New("session is nil")
	}
	if session.ID == "" {
		return errors.New("session ID is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Create a deep copy of the session to avoid external modifications
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

	// Deep copy the data map
	for k, v := range session.Data {
		sessionCopy.Data[k] = v
	}

	s.sessions[session.ID] = sessionCopy
	return nil
}

// Load loads a session from memory
func (s *inMemorySessionStorage) Load(sessionID string) (*Session, error) {
	if sessionID == "" {
		return nil, errors.New("session ID is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return nil, errors.New("session not found")
	}

	// Check if session is expired
	if session.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("session expired")
	}

	// Return a copy to avoid external modifications
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

	// Deep copy the data map
	for k, v := range session.Data {
		sessionCopy.Data[k] = v
	}

	return sessionCopy, nil
}

// Delete deletes a session from memory
func (s *inMemorySessionStorage) Delete(sessionID string) error {
	if sessionID == "" {
		return errors.New("session ID is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.sessions, sessionID)
	return nil
}

// Cleanup removes expired sessions from memory
func (s *inMemorySessionStorage) Cleanup() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for id, session := range s.sessions {
		if session.ExpiresAt.Before(now) {
			delete(s.sessions, id)
		}
	}

	return nil
}

// Count returns the number of sessions in memory (useful for testing)
func (s *inMemorySessionStorage) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.sessions)
}
