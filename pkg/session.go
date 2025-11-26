package pkg

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// SessionStorageType defines the type of session storage backend
type SessionStorageType string

const (
	SessionStorageDatabase   SessionStorageType = "database"
	SessionStorageCache      SessionStorageType = "cache"
	SessionStorageFilesystem SessionStorageType = "filesystem"
)

// SessionConfig defines configuration for session management
type SessionConfig struct {
	// StorageType specifies the session storage backend.
	// Default: SessionStorageDatabase
	StorageType SessionStorageType `json:"storage_type"`

	// CookieName is the name of the session cookie.
	// Default: "rockstar_session"
	CookieName string `json:"cookie_name"`

	// CookiePath is the path scope for the session cookie.
	// Default: "/"
	CookiePath string `json:"cookie_path"`

	// CookieDomain is the domain scope for the session cookie.
	// Default: ""
	CookieDomain string `json:"cookie_domain"`

	// CookieSecure indicates if the cookie should only be sent over HTTPS.
	// Default: true
	CookieSecure bool `json:"cookie_secure"`

	// CookieHTTPOnly indicates if the cookie should be inaccessible to JavaScript.
	// Default: true
	CookieHTTPOnly bool `json:"cookie_http_only"`

	// CookieSameSite specifies the SameSite attribute for the cookie (Strict, Lax, None).
	// Default: "Lax"
	CookieSameSite string `json:"cookie_same_site"`

	// SessionLifetime is the duration before a session expires.
	// Default: 24 hours
	SessionLifetime time.Duration `json:"session_lifetime"`

	// EncryptionKey is the AES-256 key (32 bytes) for encrypting session data.
	// Required, no default
	EncryptionKey []byte `json:"-"`

	// FilesystemPath is the directory path for filesystem-based session storage.
	// Default: "./sessions"
	FilesystemPath string `json:"filesystem_path"`

	// CleanupInterval is the interval for cleaning up expired sessions.
	// Default: 1 hour
	CleanupInterval time.Duration `json:"cleanup_interval"`
}

// DefaultSessionConfig returns default session configuration
func DefaultSessionConfig() *SessionConfig {
	return &SessionConfig{
		StorageType:     SessionStorageDatabase,
		CookieName:      "rockstar_session",
		CookiePath:      "/",
		CookieDomain:    "",
		CookieSecure:    true,
		CookieHTTPOnly:  true,
		CookieSameSite:  "Lax",
		SessionLifetime: 24 * time.Hour,
		FilesystemPath:  "./sessions",
		CleanupInterval: 1 * time.Hour,
	}
}

// sessionManager implements the SessionManager interface
type sessionManager struct {
	config        *SessionConfig
	db            DatabaseManager
	cache         CacheManager
	cipher        cipher.Block
	mu            sync.RWMutex
	sessions      map[string]*Session // In-memory cache for filesystem storage
	stopCleanup   chan struct{}
	memoryStorage *inMemorySessionStorage // In-memory storage when no database is available
	usingInMemory bool                    // Flag to indicate if using in-memory storage
}

// NewSessionManager creates a new session manager instance
func NewSessionManager(config *SessionConfig, db DatabaseManager, cache CacheManager) (SessionManager, error) {
	if config == nil {
		config = DefaultSessionConfig()
	}

	// Apply defaults to ensure all fields have sensible values
	config.ApplyDefaults()

	// Validate encryption key
	if len(config.EncryptionKey) != 32 {
		return nil, errors.New("encryption key must be 32 bytes for AES-256")
	}

	// Create AES cipher
	block, err := aes.NewCipher(config.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	sm := &sessionManager{
		config:      config,
		db:          db,
		cache:       cache,
		cipher:      block,
		sessions:    make(map[string]*Session),
		stopCleanup: make(chan struct{}),
	}

	// Check if database is available and storage type is database
	if isNoopDatabase(db) && config.StorageType == SessionStorageDatabase {
		// Switch to in-memory storage when no database is available and database storage is requested
		sm.memoryStorage = newInMemorySessionStorage()
		sm.usingInMemory = true
		fmt.Println("WARN: SessionManager using in-memory storage. Sessions will not persist across restarts.")
	} else {
		// Create filesystem directory if using filesystem storage
		if config.StorageType == SessionStorageFilesystem {
			if err := os.MkdirAll(config.FilesystemPath, 0755); err != nil {
				return nil, fmt.Errorf("failed to create session directory: %w", err)
			}
		}
	}

	// Start cleanup goroutine
	go sm.cleanupLoop()

	return sm, nil
}

// Create creates a new session
func (sm *sessionManager) Create(ctx Context) (*Session, error) {
	sessionID, err := generateSessionID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %w", err)
	}

	now := time.Now()
	session := &Session{
		ID:        sessionID,
		Data:      make(map[string]interface{}),
		ExpiresAt: now.Add(sm.config.SessionLifetime),
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Set user and tenant from context if available
	if ctx != nil {
		if user := ctx.User(); user != nil {
			session.UserID = user.ID
		}
		if tenant := ctx.Tenant(); tenant != nil {
			session.TenantID = tenant.ID
		}

		// Extract IP and User-Agent from request
		if req := ctx.Request(); req != nil {
			session.IPAddress = req.RemoteAddr
			session.UserAgent = ctx.GetHeader("User-Agent")
		}
	}

	// Save session to storage
	if err := sm.saveToStorage(session); err != nil {
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	return session, nil
}

// Load loads a session by ID
func (sm *sessionManager) Load(ctx Context, sessionID string) (*Session, error) {
	if sessionID == "" {
		return nil, errors.New("session ID is required")
	}

	session, err := sm.loadFromStorage(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to load session: %w", err)
	}

	// Check if session is expired
	if session.ExpiresAt.Before(time.Now()) {
		// Clean up expired session
		_ = sm.Destroy(ctx, sessionID)
		return nil, errors.New("session expired")
	}

	return session, nil
}

// Save saves a session
func (sm *sessionManager) Save(ctx Context, session *Session) error {
	if session == nil {
		return errors.New("session is nil")
	}

	session.UpdatedAt = time.Now()

	if err := sm.saveToStorage(session); err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	return nil
}

// Destroy destroys a session
func (sm *sessionManager) Destroy(ctx Context, sessionID string) error {
	if sessionID == "" {
		return errors.New("session ID is required")
	}

	if err := sm.deleteFromStorage(sessionID); err != nil {
		return fmt.Errorf("failed to destroy session: %w", err)
	}

	return nil
}

// Refresh refreshes a session's expiration time
func (sm *sessionManager) Refresh(ctx Context, sessionID string) error {
	session, err := sm.Load(ctx, sessionID)
	if err != nil {
		return err
	}

	session.ExpiresAt = time.Now().Add(sm.config.SessionLifetime)
	return sm.Save(ctx, session)
}

// Get retrieves a value from session data
func (sm *sessionManager) Get(sessionID, key string) (interface{}, error) {
	session, err := sm.loadFromStorage(sessionID)
	if err != nil {
		return nil, err
	}

	value, exists := session.Data[key]
	if !exists {
		return nil, fmt.Errorf("key %s not found in session", key)
	}

	return value, nil
}

// Set sets a value in session data
func (sm *sessionManager) Set(sessionID, key string, value interface{}) error {
	session, err := sm.loadFromStorage(sessionID)
	if err != nil {
		return err
	}

	session.Data[key] = value
	session.UpdatedAt = time.Now()

	return sm.saveToStorage(session)
}

// Delete deletes a key from session data
func (sm *sessionManager) Delete(sessionID, key string) error {
	session, err := sm.loadFromStorage(sessionID)
	if err != nil {
		return err
	}

	delete(session.Data, key)
	session.UpdatedAt = time.Now()

	return sm.saveToStorage(session)
}

// Clear clears all data from a session
func (sm *sessionManager) Clear(sessionID string) error {
	session, err := sm.loadFromStorage(sessionID)
	if err != nil {
		return err
	}

	session.Data = make(map[string]interface{})
	session.UpdatedAt = time.Now()

	return sm.saveToStorage(session)
}

// SetCookie sets an encrypted session cookie
func (sm *sessionManager) SetCookie(ctx Context, session *Session) error {
	if ctx == nil || session == nil {
		return errors.New("context and session are required")
	}

	// Encrypt session ID
	encryptedID, err := sm.encryptSessionID(session.ID)
	if err != nil {
		return fmt.Errorf("failed to encrypt session ID: %w", err)
	}

	// Create cookie
	cookie := &Cookie{
		Name:      sm.config.CookieName,
		Value:     encryptedID,
		Path:      sm.config.CookiePath,
		Domain:    sm.config.CookieDomain,
		Expires:   session.ExpiresAt,
		Secure:    sm.config.CookieSecure,
		HttpOnly:  sm.config.CookieHTTPOnly,
		Encrypted: true,
	}

	return ctx.SetCookie(cookie)
}

// GetSessionFromCookie retrieves session from encrypted cookie
func (sm *sessionManager) GetSessionFromCookie(ctx Context) (*Session, error) {
	if ctx == nil {
		return nil, errors.New("context is required")
	}

	// Get cookie
	cookie, err := ctx.GetCookie(sm.config.CookieName)
	if err != nil {
		return nil, fmt.Errorf("session cookie not found: %w", err)
	}

	// Decrypt session ID
	sessionID, err := sm.decryptSessionID(cookie.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt session ID: %w", err)
	}

	// Load session
	return sm.Load(ctx, sessionID)
}

// IsValid checks if a session is valid
func (sm *sessionManager) IsValid(sessionID string) bool {
	session, err := sm.loadFromStorage(sessionID)
	if err != nil {
		return false
	}

	return !session.ExpiresAt.Before(time.Now())
}

// IsExpired checks if a session is expired
func (sm *sessionManager) IsExpired(sessionID string) bool {
	session, err := sm.loadFromStorage(sessionID)
	if err != nil {
		return true
	}

	return session.ExpiresAt.Before(time.Now())
}

// CleanupExpired removes expired sessions
func (sm *sessionManager) CleanupExpired() error {
	// Use in-memory storage cleanup if no database is available
	if sm.usingInMemory {
		return sm.memoryStorage.Cleanup()
	}

	switch sm.config.StorageType {
	case SessionStorageDatabase:
		if sm.db != nil {
			return sm.db.CleanupExpiredSessions()
		}
	case SessionStorageCache:
		// Cache typically handles expiration automatically
		return nil
	case SessionStorageFilesystem:
		return sm.cleanupFilesystemSessions()
	}

	return nil
}

// Storage backend methods

func (sm *sessionManager) saveToStorage(session *Session) error {
	// Use in-memory storage if no database is available
	if sm.usingInMemory {
		return sm.memoryStorage.Save(session)
	}

	switch sm.config.StorageType {
	case SessionStorageDatabase:
		if sm.db == nil {
			return errors.New("database manager not configured")
		}
		return sm.db.SaveSession(session)

	case SessionStorageCache:
		if sm.cache == nil {
			return errors.New("cache manager not configured")
		}
		ttl := time.Until(session.ExpiresAt)
		return sm.cache.Set(sessionKey(session.ID), session, ttl)

	case SessionStorageFilesystem:
		return sm.saveToFilesystem(session)

	default:
		return fmt.Errorf("unsupported storage type: %s", sm.config.StorageType)
	}
}

func (sm *sessionManager) loadFromStorage(sessionID string) (*Session, error) {
	// Use in-memory storage if no database is available
	if sm.usingInMemory {
		return sm.memoryStorage.Load(sessionID)
	}

	switch sm.config.StorageType {
	case SessionStorageDatabase:
		if sm.db == nil {
			return nil, errors.New("database manager not configured")
		}
		return sm.db.LoadSession(sessionID)

	case SessionStorageCache:
		if sm.cache == nil {
			return nil, errors.New("cache manager not configured")
		}
		value, err := sm.cache.Get(sessionKey(sessionID))
		if err != nil {
			return nil, err
		}
		session, ok := value.(*Session)
		if !ok {
			return nil, errors.New("invalid session data in cache")
		}
		return session, nil

	case SessionStorageFilesystem:
		return sm.loadFromFilesystem(sessionID)

	default:
		return nil, fmt.Errorf("unsupported storage type: %s", sm.config.StorageType)
	}
}

func (sm *sessionManager) deleteFromStorage(sessionID string) error {
	// Use in-memory storage if no database is available
	if sm.usingInMemory {
		return sm.memoryStorage.Delete(sessionID)
	}

	switch sm.config.StorageType {
	case SessionStorageDatabase:
		if sm.db == nil {
			return errors.New("database manager not configured")
		}
		return sm.db.DeleteSession(sessionID)

	case SessionStorageCache:
		if sm.cache == nil {
			return errors.New("cache manager not configured")
		}
		return sm.cache.Delete(sessionKey(sessionID))

	case SessionStorageFilesystem:
		return sm.deleteFromFilesystem(sessionID)

	default:
		return fmt.Errorf("unsupported storage type: %s", sm.config.StorageType)
	}
}

// Filesystem storage methods

func (sm *sessionManager) saveToFilesystem(session *Session) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	filename := filepath.Join(sm.config.FilesystemPath, session.ID+".json")
	if err := os.WriteFile(filename, data, 0600); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}

	sm.sessions[session.ID] = session
	return nil
}

func (sm *sessionManager) loadFromFilesystem(sessionID string) (*Session, error) {
	sm.mu.RLock()
	if session, exists := sm.sessions[sessionID]; exists {
		sm.mu.RUnlock()
		return session, nil
	}
	sm.mu.RUnlock()

	filename := filepath.Join(sm.config.FilesystemPath, sessionID+".json")
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("session not found")
		}
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	sm.mu.Lock()
	sm.sessions[sessionID] = &session
	sm.mu.Unlock()

	return &session, nil
}

func (sm *sessionManager) deleteFromFilesystem(sessionID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	delete(sm.sessions, sessionID)

	filename := filepath.Join(sm.config.FilesystemPath, sessionID+".json")
	if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete session file: %w", err)
	}

	return nil
}

func (sm *sessionManager) cleanupFilesystemSessions() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	files, err := os.ReadDir(sm.config.FilesystemPath)
	if err != nil {
		return fmt.Errorf("failed to read session directory: %w", err)
	}

	now := time.Now()
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		sessionID := file.Name()[:len(file.Name())-5] // Remove .json extension
		if session, exists := sm.sessions[sessionID]; exists {
			if session.ExpiresAt.Before(now) {
				delete(sm.sessions, sessionID)
				os.Remove(filepath.Join(sm.config.FilesystemPath, file.Name()))
			}
		} else {
			// Load and check expiration
			filename := filepath.Join(sm.config.FilesystemPath, file.Name())
			data, err := os.ReadFile(filename)
			if err != nil {
				continue
			}

			var session Session
			if err := json.Unmarshal(data, &session); err != nil {
				continue
			}

			if session.ExpiresAt.Before(now) {
				os.Remove(filename)
			}
		}
	}

	return nil
}

// Encryption methods

func (sm *sessionManager) encryptSessionID(sessionID string) (string, error) {
	plaintext := []byte(sessionID)

	// Create a new GCM cipher mode
	gcm, err := cipher.NewGCM(sm.cipher)
	if err != nil {
		return "", err
	}

	// Create a nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt the data
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	// Encode to base64 for cookie storage
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

func (sm *sessionManager) decryptSessionID(encryptedID string) (string, error) {
	// Decode from base64
	ciphertext, err := base64.URLEncoding.DecodeString(encryptedID)
	if err != nil {
		return "", err
	}

	// Create a new GCM cipher mode
	gcm, err := cipher.NewGCM(sm.cipher)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	// Extract nonce and ciphertext
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt the data
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// Cleanup loop

func (sm *sessionManager) cleanupLoop() {
	ticker := time.NewTicker(sm.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			_ = sm.CleanupExpired()
		case <-sm.stopCleanup:
			return
		}
	}
}

// Stop stops the session manager and cleanup goroutine
func (sm *sessionManager) Stop() {
	close(sm.stopCleanup)
}

// Helper functions

func generateSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func sessionKey(sessionID string) string {
	return "session:" + sessionID
}
