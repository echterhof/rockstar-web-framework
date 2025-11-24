//go:build !test
// +build !test

package pkg

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// sessionStorage implements session storage operations for DatabaseManager
type sessionStorage struct {
	db DatabaseManager
}

// newSessionStorage creates a new session storage instance
func newSessionStorage(db DatabaseManager) *sessionStorage {
	return &sessionStorage{db: db}
}

// SaveSession saves a session to the database
func (ss *sessionStorage) SaveSession(session *Session) error {
	dataJSON, err := json.Marshal(session.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}

	query := `INSERT INTO sessions (id, user_id, tenant_id, data, expires_at, created_at, updated_at, ip_address, user_agent) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?) 
			  ON DUPLICATE KEY UPDATE data = ?, expires_at = ?, updated_at = ?, ip_address = ?, user_agent = ?`

	now := time.Now()
	if session.CreatedAt.IsZero() {
		session.CreatedAt = now
	}
	session.UpdatedAt = now

	_, err = ss.db.Exec(query,
		session.ID, session.UserID, session.TenantID, string(dataJSON),
		session.ExpiresAt, session.CreatedAt, session.UpdatedAt,
		session.IPAddress, session.UserAgent,
		// ON DUPLICATE KEY UPDATE values
		string(dataJSON), session.ExpiresAt, session.UpdatedAt,
		session.IPAddress, session.UserAgent)

	return err
}

// LoadSession loads a session from the database
func (ss *sessionStorage) LoadSession(sessionID string) (*Session, error) {
	query := `SELECT id, user_id, tenant_id, data, expires_at, created_at, updated_at, ip_address, user_agent 
			  FROM sessions WHERE id = ? AND expires_at > ?`

	row := ss.db.QueryRow(query, sessionID, time.Now())

	session := &Session{}
	var dataJSON string

	err := row.Scan(&session.ID, &session.UserID, &session.TenantID, &dataJSON,
		&session.ExpiresAt, &session.CreatedAt, &session.UpdatedAt,
		&session.IPAddress, &session.UserAgent)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found or expired")
		}
		return nil, fmt.Errorf("failed to load session: %w", err)
	}

	if err := json.Unmarshal([]byte(dataJSON), &session.Data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	return session, nil
}

// DeleteSession deletes a session from the database
func (ss *sessionStorage) DeleteSession(sessionID string) error {
	query := `DELETE FROM sessions WHERE id = ?`
	_, err := ss.db.Exec(query, sessionID)
	return err
}

// CleanupExpiredSessions removes expired sessions from the database
func (ss *sessionStorage) CleanupExpiredSessions() error {
	query := `DELETE FROM sessions WHERE expires_at <= ?`
	_, err := ss.db.Exec(query, time.Now())
	return err
}

// DatabaseManager session methods - these delegate to sessionStorage

// SaveSession saves a session to the database
func (dm *databaseManager) SaveSession(session *Session) error {
	ss := newSessionStorage(dm)
	return ss.SaveSession(session)
}

// LoadSession loads a session from the database
func (dm *databaseManager) LoadSession(sessionID string) (*Session, error) {
	ss := newSessionStorage(dm)
	return ss.LoadSession(sessionID)
}

// DeleteSession deletes a session from the database
func (dm *databaseManager) DeleteSession(sessionID string) error {
	ss := newSessionStorage(dm)
	return ss.DeleteSession(sessionID)
}

// CleanupExpiredSessions removes expired sessions from the database
func (dm *databaseManager) CleanupExpiredSessions() error {
	ss := newSessionStorage(dm)
	return ss.CleanupExpiredSessions()
}
