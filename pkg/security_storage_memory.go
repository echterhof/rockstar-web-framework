package pkg

import (
	"errors"
	"sync"
	"time"
)

// inMemoryTokenStorage implements token storage in memory
type inMemoryTokenStorage struct {
	mu     sync.RWMutex
	tokens map[string]*AccessToken
}

// newInMemoryTokenStorage creates a new in-memory token storage instance
func newInMemoryTokenStorage() *inMemoryTokenStorage {
	return &inMemoryTokenStorage{
		tokens: make(map[string]*AccessToken),
	}
}

// Save saves a token to memory
func (s *inMemoryTokenStorage) Save(token *AccessToken) error {
	if token == nil {
		return errors.New("token is nil")
	}
	if token.Token == "" {
		return errors.New("token value is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Create a deep copy of the token to avoid external modifications
	tokenCopy := &AccessToken{
		Token:     token.Token,
		UserID:    token.UserID,
		TenantID:  token.TenantID,
		Scopes:    make([]string, len(token.Scopes)),
		ExpiresAt: token.ExpiresAt,
		CreatedAt: token.CreatedAt,
	}

	// Deep copy the scopes slice
	copy(tokenCopy.Scopes, token.Scopes)

	s.tokens[token.Token] = tokenCopy
	return nil
}

// Load loads a token from memory
func (s *inMemoryTokenStorage) Load(tokenValue string) (*AccessToken, error) {
	if tokenValue == "" {
		return nil, errors.New("token value is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	token, exists := s.tokens[tokenValue]
	if !exists {
		return nil, errors.New("token not found")
	}

	// Return a copy to avoid external modifications
	tokenCopy := &AccessToken{
		Token:     token.Token,
		UserID:    token.UserID,
		TenantID:  token.TenantID,
		Scopes:    make([]string, len(token.Scopes)),
		ExpiresAt: token.ExpiresAt,
		CreatedAt: token.CreatedAt,
	}

	// Deep copy the scopes slice
	copy(tokenCopy.Scopes, token.Scopes)

	return tokenCopy, nil
}

// Validate validates a token and returns it if valid
func (s *inMemoryTokenStorage) Validate(tokenValue string) (*AccessToken, error) {
	if tokenValue == "" {
		return nil, errors.New("token value is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	token, exists := s.tokens[tokenValue]
	if !exists {
		return nil, errors.New("token not found")
	}

	// Check if token is expired
	if token.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("token expired")
	}

	// Return a copy to avoid external modifications
	tokenCopy := &AccessToken{
		Token:     token.Token,
		UserID:    token.UserID,
		TenantID:  token.TenantID,
		Scopes:    make([]string, len(token.Scopes)),
		ExpiresAt: token.ExpiresAt,
		CreatedAt: token.CreatedAt,
	}

	// Deep copy the scopes slice
	copy(tokenCopy.Scopes, token.Scopes)

	return tokenCopy, nil
}

// Delete deletes a token from memory
func (s *inMemoryTokenStorage) Delete(tokenValue string) error {
	if tokenValue == "" {
		return errors.New("token value is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.tokens, tokenValue)
	return nil
}

// Cleanup removes expired tokens from memory
func (s *inMemoryTokenStorage) Cleanup() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for tokenValue, token := range s.tokens {
		if token.ExpiresAt.Before(now) {
			delete(s.tokens, tokenValue)
		}
	}

	return nil
}

// Count returns the number of tokens in memory (useful for testing)
func (s *inMemoryTokenStorage) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.tokens)
}

// inMemoryRateLimitStorage implements rate limiting in memory
type inMemoryRateLimitStorage struct {
	mu     sync.RWMutex
	limits map[string]*rateLimitEntry
}

// rateLimitEntry stores rate limit information for a key
type rateLimitEntry struct {
	count     int
	expiresAt time.Time
}

// newInMemoryRateLimitStorage creates a new in-memory rate limit storage instance
func newInMemoryRateLimitStorage() *inMemoryRateLimitStorage {
	return &inMemoryRateLimitStorage{
		limits: make(map[string]*rateLimitEntry),
	}
}

// CheckRateLimit checks if a rate limit has been exceeded
func (s *inMemoryRateLimitStorage) CheckRateLimit(key string, limit int, window time.Duration) (bool, error) {
	if key == "" {
		return false, errors.New("rate limit key is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exists := s.limits[key]
	if !exists {
		// No entry exists, so limit is not exceeded
		return true, nil
	}

	// Check if entry has expired
	if entry.expiresAt.Before(time.Now()) {
		// Entry expired, so limit is not exceeded
		return true, nil
	}

	// Check if count exceeds limit
	return entry.count < limit, nil
}

// IncrementRateLimit increments the rate limit counter for a key
func (s *inMemoryRateLimitStorage) IncrementRateLimit(key string, window time.Duration) error {
	if key == "" {
		return errors.New("rate limit key is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	entry, exists := s.limits[key]

	if !exists || entry.expiresAt.Before(now) {
		// Create new entry
		s.limits[key] = &rateLimitEntry{
			count:     1,
			expiresAt: now.Add(window),
		}
	} else {
		// Increment existing entry
		entry.count++
	}

	return nil
}

// Cleanup removes expired rate limit entries from memory
func (s *inMemoryRateLimitStorage) Cleanup() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for key, entry := range s.limits {
		if entry.expiresAt.Before(now) {
			delete(s.limits, key)
		}
	}

	return nil
}

// Count returns the number of rate limit entries in memory (useful for testing)
func (s *inMemoryRateLimitStorage) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.limits)
}
