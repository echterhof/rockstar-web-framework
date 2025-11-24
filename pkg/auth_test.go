package pkg

import (
	"database/sql"
	"testing"
	"time"
)

// MockDatabaseManager is a mock implementation of DatabaseManager for testing
type MockDatabaseManager struct {
	accessTokens map[string]*AccessToken
	sessions     map[string]*Session
}

func NewMockDatabaseManager() *MockDatabaseManager {
	return &MockDatabaseManager{
		accessTokens: make(map[string]*AccessToken),
		sessions:     make(map[string]*Session),
	}
}

func (m *MockDatabaseManager) SaveAccessToken(token *AccessToken) error {
	m.accessTokens[token.Token] = token
	return nil
}

func (m *MockDatabaseManager) LoadAccessToken(tokenValue string) (*AccessToken, error) {
	token, exists := m.accessTokens[tokenValue]
	if !exists {
		return nil, NewDatabaseError("token not found", "LoadAccessToken")
	}
	return token, nil
}

func (m *MockDatabaseManager) ValidateAccessToken(tokenValue string) (*AccessToken, error) {
	token, err := m.LoadAccessToken(tokenValue)
	if err != nil {
		return nil, err
	}

	if token.ExpiresAt.Before(time.Now()) {
		return nil, NewDatabaseError("token expired", "ValidateAccessToken")
	}

	return token, nil
}

func (m *MockDatabaseManager) DeleteAccessToken(tokenValue string) error {
	delete(m.accessTokens, tokenValue)
	return nil
}

func (m *MockDatabaseManager) CleanupExpiredTokens() error {
	return nil
}

func (m *MockDatabaseManager) SaveSession(session *Session) error {
	m.sessions[session.ID] = session
	return nil
}

func (m *MockDatabaseManager) LoadSession(sessionID string) (*Session, error) {
	session, exists := m.sessions[sessionID]
	if !exists {
		return nil, NewDatabaseError("session not found", "LoadSession")
	}
	return session, nil
}

func (m *MockDatabaseManager) DeleteSession(sessionID string) error {
	delete(m.sessions, sessionID)
	return nil
}

func (m *MockDatabaseManager) CleanupExpiredSessions() error {
	return nil
}

// Stub implementations for other DatabaseManager methods
func (m *MockDatabaseManager) Connect(config DatabaseConfig) error { return nil }
func (m *MockDatabaseManager) Close() error                        { return nil }
func (m *MockDatabaseManager) Ping() error                         { return nil }
func (m *MockDatabaseManager) Stats() DatabaseStats                { return DatabaseStats{} }
func (m *MockDatabaseManager) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return nil, nil
}
func (m *MockDatabaseManager) QueryRow(query string, args ...interface{}) *sql.Row {
	return nil
}
func (m *MockDatabaseManager) Exec(query string, args ...interface{}) (sql.Result, error) {
	return nil, nil
}
func (m *MockDatabaseManager) Prepare(query string) (*sql.Stmt, error) { return nil, nil }
func (m *MockDatabaseManager) Begin() (Transaction, error)             { return nil, nil }
func (m *MockDatabaseManager) BeginTx(opts *sql.TxOptions) (Transaction, error) {
	return nil, nil
}
func (m *MockDatabaseManager) SaveTenant(tenant *Tenant) error { return nil }
func (m *MockDatabaseManager) LoadTenant(tenantID string) (*Tenant, error) {
	return nil, nil
}
func (m *MockDatabaseManager) LoadTenantByHost(hostname string) (*Tenant, error) {
	return nil, nil
}
func (m *MockDatabaseManager) SaveWorkloadMetrics(metrics *WorkloadMetrics) error {
	return nil
}
func (m *MockDatabaseManager) GetWorkloadMetrics(tenantID string, from, to time.Time) ([]*WorkloadMetrics, error) {
	return nil, nil
}
func (m *MockDatabaseManager) CheckRateLimit(key string, limit int, window time.Duration) (bool, error) {
	return true, nil
}
func (m *MockDatabaseManager) IncrementRateLimit(key string, window time.Duration) error {
	return nil
}
func (m *MockDatabaseManager) Migrate() error      { return nil }
func (m *MockDatabaseManager) CreateTables() error { return nil }
func (m *MockDatabaseManager) DropTables() error   { return nil }

// Test OAuth2 authentication
func TestAuthenticateOAuth2_ValidToken(t *testing.T) {
	// Setup
	db := NewMockDatabaseManager()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	// Create a valid access token
	token := &AccessToken{
		Token:     "valid-oauth2-token",
		UserID:    "user123",
		TenantID:  "tenant456",
		Scopes:    []string{"read", "write"},
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
	}
	db.SaveAccessToken(token)

	// Test
	user, err := authManager.AuthenticateOAuth2("valid-oauth2-token")

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if user == nil {
		t.Fatal("Expected user, got nil")
	}
	if user.ID != "user123" {
		t.Errorf("Expected user ID 'user123', got '%s'", user.ID)
	}
	if user.TenantID != "tenant456" {
		t.Errorf("Expected tenant ID 'tenant456', got '%s'", user.TenantID)
	}
	if user.AuthMethod != "OAuth2" {
		t.Errorf("Expected auth method 'OAuth2', got '%s'", user.AuthMethod)
	}
	if len(user.Roles) != 2 {
		t.Errorf("Expected 2 roles, got %d", len(user.Roles))
	}
}

func TestAuthenticateOAuth2_EmptyToken(t *testing.T) {
	// Setup
	db := NewMockDatabaseManager()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	// Test
	user, err := authManager.AuthenticateOAuth2("")

	// Assert
	if err == nil {
		t.Fatal("Expected error for empty token, got nil")
	}
	if user != nil {
		t.Errorf("Expected nil user, got %v", user)
	}
}

func TestAuthenticateOAuth2_InvalidToken(t *testing.T) {
	// Setup
	db := NewMockDatabaseManager()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	// Test
	user, err := authManager.AuthenticateOAuth2("invalid-token")

	// Assert
	if err == nil {
		t.Fatal("Expected error for invalid token, got nil")
	}
	if user != nil {
		t.Errorf("Expected nil user, got %v", user)
	}
}

func TestAuthenticateOAuth2_ExpiredToken(t *testing.T) {
	// Setup
	db := NewMockDatabaseManager()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	// Create an expired access token
	token := &AccessToken{
		Token:     "expired-oauth2-token",
		UserID:    "user123",
		TenantID:  "tenant456",
		Scopes:    []string{"read"},
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
		CreatedAt: time.Now().Add(-2 * time.Hour),
	}
	db.SaveAccessToken(token)

	// Test
	user, err := authManager.AuthenticateOAuth2("expired-oauth2-token")

	// Assert
	if err == nil {
		t.Fatal("Expected error for expired token, got nil")
	}
	if user != nil {
		t.Errorf("Expected nil user, got %v", user)
	}

	// Check error code
	if frameworkErr, ok := err.(*FrameworkError); ok {
		if frameworkErr.Code != ErrCodeTokenExpired {
			t.Errorf("Expected error code '%s', got '%s'", ErrCodeTokenExpired, frameworkErr.Code)
		}
	}
}

// Test JWT authentication
func TestAuthenticateJWT_ValidToken(t *testing.T) {
	// Setup
	db := NewMockDatabaseManager()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	// Create a user and generate JWT
	user := &User{
		ID:       "user123",
		Username: "testuser",
		Email:    "test@example.com",
		Roles:    []string{"admin", "user"},
		Actions:  []string{"read", "write"},
		TenantID: "tenant456",
		Metadata: map[string]interface{}{"department": "engineering"},
	}

	token, err := authManager.GenerateJWT(user, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate JWT: %v", err)
	}

	// Test
	authenticatedUser, err := authManager.AuthenticateJWT(token)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if authenticatedUser == nil {
		t.Fatal("Expected user, got nil")
	}
	if authenticatedUser.ID != user.ID {
		t.Errorf("Expected user ID '%s', got '%s'", user.ID, authenticatedUser.ID)
	}
	if authenticatedUser.Username != user.Username {
		t.Errorf("Expected username '%s', got '%s'", user.Username, authenticatedUser.Username)
	}
	if authenticatedUser.Email != user.Email {
		t.Errorf("Expected email '%s', got '%s'", user.Email, authenticatedUser.Email)
	}
	if authenticatedUser.TenantID != user.TenantID {
		t.Errorf("Expected tenant ID '%s', got '%s'", user.TenantID, authenticatedUser.TenantID)
	}
	if authenticatedUser.AuthMethod != "JWT" {
		t.Errorf("Expected auth method 'JWT', got '%s'", authenticatedUser.AuthMethod)
	}
}

func TestAuthenticateJWT_EmptyToken(t *testing.T) {
	// Setup
	db := NewMockDatabaseManager()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	// Test
	user, err := authManager.AuthenticateJWT("")

	// Assert
	if err == nil {
		t.Fatal("Expected error for empty token, got nil")
	}
	if user != nil {
		t.Errorf("Expected nil user, got %v", user)
	}
}

func TestAuthenticateJWT_InvalidFormat(t *testing.T) {
	// Setup
	db := NewMockDatabaseManager()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	// Test with invalid format (not 3 parts)
	user, err := authManager.AuthenticateJWT("invalid.token")

	// Assert
	if err == nil {
		t.Fatal("Expected error for invalid token format, got nil")
	}
	if user != nil {
		t.Errorf("Expected nil user, got %v", user)
	}
}

func TestAuthenticateJWT_InvalidSignature(t *testing.T) {
	// Setup
	db := NewMockDatabaseManager()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	// Create a JWT with wrong secret
	wrongAuthManager := NewAuthManager(db, "wrong-secret", OAuth2Config{})
	user := &User{
		ID:       "user123",
		Username: "testuser",
		TenantID: "tenant456",
	}
	token, _ := wrongAuthManager.GenerateJWT(user, 1*time.Hour)

	// Test with correct auth manager (different secret)
	authenticatedUser, err := authManager.AuthenticateJWT(token)

	// Assert
	if err == nil {
		t.Fatal("Expected error for invalid signature, got nil")
	}
	if authenticatedUser != nil {
		t.Errorf("Expected nil user, got %v", authenticatedUser)
	}
}

func TestAuthenticateJWT_ExpiredToken(t *testing.T) {
	// Setup
	db := NewMockDatabaseManager()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	// Create a user and generate expired JWT
	user := &User{
		ID:       "user123",
		Username: "testuser",
		TenantID: "tenant456",
	}

	// Generate token with negative expiration (already expired)
	token, err := authManager.GenerateJWT(user, -1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate JWT: %v", err)
	}

	// Test
	authenticatedUser, err := authManager.AuthenticateJWT(token)

	// Assert
	if err == nil {
		t.Fatal("Expected error for expired token, got nil")
	}
	if authenticatedUser != nil {
		t.Errorf("Expected nil user, got %v", authenticatedUser)
	}

	// Check error code
	if frameworkErr, ok := err.(*FrameworkError); ok {
		if frameworkErr.Code != ErrCodeTokenExpired {
			t.Errorf("Expected error code '%s', got '%s'", ErrCodeTokenExpired, frameworkErr.Code)
		}
	}
}

// Test access token authentication
func TestAuthenticateAccessToken_ValidToken(t *testing.T) {
	// Setup
	db := NewMockDatabaseManager()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	// Create a valid access token
	token := &AccessToken{
		Token:     "valid-access-token",
		UserID:    "user123",
		TenantID:  "tenant456",
		Scopes:    []string{"api:read", "api:write"},
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
	}
	db.SaveAccessToken(token)

	// Test
	accessToken, err := authManager.AuthenticateAccessToken("valid-access-token")

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if accessToken == nil {
		t.Fatal("Expected access token, got nil")
	}
	if accessToken.Token != "valid-access-token" {
		t.Errorf("Expected token 'valid-access-token', got '%s'", accessToken.Token)
	}
	if accessToken.UserID != "user123" {
		t.Errorf("Expected user ID 'user123', got '%s'", accessToken.UserID)
	}
}

func TestAuthenticateAccessToken_EmptyToken(t *testing.T) {
	// Setup
	db := NewMockDatabaseManager()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	// Test
	token, err := authManager.AuthenticateAccessToken("")

	// Assert
	if err == nil {
		t.Fatal("Expected error for empty token, got nil")
	}
	if token != nil {
		t.Errorf("Expected nil token, got %v", token)
	}
}

func TestAuthenticateAccessToken_InvalidToken(t *testing.T) {
	// Setup
	db := NewMockDatabaseManager()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	// Test
	token, err := authManager.AuthenticateAccessToken("invalid-token")

	// Assert
	if err == nil {
		t.Fatal("Expected error for invalid token, got nil")
	}
	if token != nil {
		t.Errorf("Expected nil token, got %v", token)
	}
}

func TestAuthenticateAccessToken_ExpiredToken(t *testing.T) {
	// Setup
	db := NewMockDatabaseManager()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	// Create an expired access token
	token := &AccessToken{
		Token:     "expired-access-token",
		UserID:    "user123",
		TenantID:  "tenant456",
		Scopes:    []string{"api:read"},
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
		CreatedAt: time.Now().Add(-2 * time.Hour),
	}
	db.SaveAccessToken(token)

	// Test
	accessToken, err := authManager.AuthenticateAccessToken("expired-access-token")

	// Assert
	if err == nil {
		t.Fatal("Expected error for expired token, got nil")
	}
	if accessToken != nil {
		t.Errorf("Expected nil token, got %v", accessToken)
	}
}

// Test token creation and management
func TestCreateAccessToken(t *testing.T) {
	// Setup
	db := NewMockDatabaseManager()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	// Test
	token, err := authManager.CreateAccessToken(
		"user123",
		"tenant456",
		[]string{"read", "write", "admin"},
		24*time.Hour,
	)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if token == nil {
		t.Fatal("Expected token, got nil")
	}
	if token.Token == "" {
		t.Error("Expected non-empty token string")
	}
	if token.UserID != "user123" {
		t.Errorf("Expected user ID 'user123', got '%s'", token.UserID)
	}
	if token.TenantID != "tenant456" {
		t.Errorf("Expected tenant ID 'tenant456', got '%s'", token.TenantID)
	}
	if len(token.Scopes) != 3 {
		t.Errorf("Expected 3 scopes, got %d", len(token.Scopes))
	}

	// Verify token is saved in database
	savedToken, err := db.LoadAccessToken(token.Token)
	if err != nil {
		t.Fatalf("Expected token to be saved in database, got error: %v", err)
	}
	if savedToken.Token != token.Token {
		t.Errorf("Expected saved token to match created token")
	}
}

func TestRevokeAccessToken(t *testing.T) {
	// Setup
	db := NewMockDatabaseManager()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	// Create a token
	token, _ := authManager.CreateAccessToken("user123", "tenant456", []string{"read"}, 1*time.Hour)

	// Test
	err := authManager.RevokeAccessToken(token.Token)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify token is deleted from database
	_, err = db.LoadAccessToken(token.Token)
	if err == nil {
		t.Error("Expected token to be deleted from database")
	}
}

func TestRefreshAccessToken(t *testing.T) {
	// Setup
	db := NewMockDatabaseManager()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	// Create an initial token
	oldToken, _ := authManager.CreateAccessToken("user123", "tenant456", []string{"read", "write"}, 1*time.Hour)

	// Test
	newToken, err := authManager.RefreshAccessToken(oldToken.Token, 2*time.Hour)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if newToken == nil {
		t.Fatal("Expected new token, got nil")
	}
	if newToken.Token == oldToken.Token {
		t.Error("Expected new token to be different from old token")
	}
	if newToken.UserID != oldToken.UserID {
		t.Errorf("Expected same user ID, got '%s' vs '%s'", newToken.UserID, oldToken.UserID)
	}
	if newToken.TenantID != oldToken.TenantID {
		t.Errorf("Expected same tenant ID, got '%s' vs '%s'", newToken.TenantID, oldToken.TenantID)
	}

	// Verify new token is in database
	_, err = db.LoadAccessToken(newToken.Token)
	if err != nil {
		t.Errorf("Expected new token to be in database, got error: %v", err)
	}
}

func TestGenerateJWT(t *testing.T) {
	// Setup
	db := NewMockDatabaseManager()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	user := &User{
		ID:       "user123",
		Username: "testuser",
		Email:    "test@example.com",
		Roles:    []string{"admin"},
		Actions:  []string{"read", "write"},
		TenantID: "tenant456",
		Metadata: map[string]interface{}{"key": "value"},
	}

	// Test
	token, err := authManager.GenerateJWT(user, 1*time.Hour)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if token == "" {
		t.Fatal("Expected non-empty token")
	}

	// Verify token can be parsed back
	claims, err := authManager.parseJWT(token)
	if err != nil {
		t.Fatalf("Expected token to be parseable, got error: %v", err)
	}
	if claims.UserID != user.ID {
		t.Errorf("Expected user ID '%s', got '%s'", user.ID, claims.UserID)
	}
	if claims.Username != user.Username {
		t.Errorf("Expected username '%s', got '%s'", user.Username, claims.Username)
	}
	if claims.Email != user.Email {
		t.Errorf("Expected email '%s', got '%s'", user.Email, claims.Email)
	}
}

// Test role-based authorization
func TestAuthorizeRole_ValidRole(t *testing.T) {
	// Setup
	db := NewMockDatabaseManager()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	user := &User{
		ID:       "user123",
		Roles:    []string{"admin", "user", "moderator"},
		TenantID: "tenant456",
	}

	// Test
	err := authManager.AuthorizeRole(user, "admin")

	// Assert
	if err != nil {
		t.Fatalf("Expected no error for valid role, got: %v", err)
	}
}

func TestAuthorizeRole_InvalidRole(t *testing.T) {
	// Setup
	db := NewMockDatabaseManager()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	user := &User{
		ID:       "user123",
		Roles:    []string{"user", "moderator"},
		TenantID: "tenant456",
	}

	// Test
	err := authManager.AuthorizeRole(user, "admin")

	// Assert
	if err == nil {
		t.Fatal("Expected error for invalid role, got nil")
	}

	// Check error code
	if frameworkErr, ok := err.(*FrameworkError); ok {
		if frameworkErr.Code != ErrCodeInsufficientRoles {
			t.Errorf("Expected error code '%s', got '%s'", ErrCodeInsufficientRoles, frameworkErr.Code)
		}
		if frameworkErr.StatusCode != 403 {
			t.Errorf("Expected status code 403, got %d", frameworkErr.StatusCode)
		}
	} else {
		t.Error("Expected FrameworkError type")
	}
}

func TestAuthorizeRole_NilUser(t *testing.T) {
	// Setup
	db := NewMockDatabaseManager()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	// Test
	err := authManager.AuthorizeRole(nil, "admin")

	// Assert
	if err == nil {
		t.Fatal("Expected error for nil user, got nil")
	}
}

func TestAuthorizeRole_EmptyRole(t *testing.T) {
	// Setup
	db := NewMockDatabaseManager()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	user := &User{
		ID:    "user123",
		Roles: []string{"admin"},
	}

	// Test
	err := authManager.AuthorizeRole(user, "")

	// Assert
	if err == nil {
		t.Fatal("Expected error for empty role, got nil")
	}
}

func TestAuthorizeRoles_HasOneOfRoles(t *testing.T) {
	// Setup
	db := NewMockDatabaseManager()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	user := &User{
		ID:       "user123",
		Roles:    []string{"user", "moderator"},
		TenantID: "tenant456",
	}

	// Test - user has "moderator" which is one of the required roles
	err := authManager.AuthorizeRoles(user, []string{"admin", "moderator", "superuser"})

	// Assert
	if err != nil {
		t.Fatalf("Expected no error when user has one of the required roles, got: %v", err)
	}
}

func TestAuthorizeRoles_HasNoneOfRoles(t *testing.T) {
	// Setup
	db := NewMockDatabaseManager()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	user := &User{
		ID:       "user123",
		Roles:    []string{"user"},
		TenantID: "tenant456",
	}

	// Test
	err := authManager.AuthorizeRoles(user, []string{"admin", "moderator", "superuser"})

	// Assert
	if err == nil {
		t.Fatal("Expected error when user has none of the required roles, got nil")
	}

	if frameworkErr, ok := err.(*FrameworkError); ok {
		if frameworkErr.Code != ErrCodeInsufficientRoles {
			t.Errorf("Expected error code '%s', got '%s'", ErrCodeInsufficientRoles, frameworkErr.Code)
		}
	}
}

func TestAuthorizeAllRoles_HasAllRoles(t *testing.T) {
	// Setup
	db := NewMockDatabaseManager()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	user := &User{
		ID:       "user123",
		Roles:    []string{"admin", "user", "moderator", "editor"},
		TenantID: "tenant456",
	}

	// Test
	err := authManager.AuthorizeAllRoles(user, []string{"admin", "moderator"})

	// Assert
	if err != nil {
		t.Fatalf("Expected no error when user has all required roles, got: %v", err)
	}
}

func TestAuthorizeAllRoles_MissingOneRole(t *testing.T) {
	// Setup
	db := NewMockDatabaseManager()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	user := &User{
		ID:       "user123",
		Roles:    []string{"admin", "user"},
		TenantID: "tenant456",
	}

	// Test
	err := authManager.AuthorizeAllRoles(user, []string{"admin", "moderator", "editor"})

	// Assert
	if err == nil {
		t.Fatal("Expected error when user is missing required roles, got nil")
	}

	if frameworkErr, ok := err.(*FrameworkError); ok {
		if frameworkErr.Code != ErrCodeInsufficientRoles {
			t.Errorf("Expected error code '%s', got '%s'", ErrCodeInsufficientRoles, frameworkErr.Code)
		}

		// Check that missing roles are reported
		if details, ok := frameworkErr.Details["missing_roles"].([]string); ok {
			if len(details) != 2 {
				t.Errorf("Expected 2 missing roles, got %d", len(details))
			}
		}
	}
}

// Test action-based authorization
func TestAuthorizeAction_ValidAction(t *testing.T) {
	// Setup
	db := NewMockDatabaseManager()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	user := &User{
		ID:       "user123",
		Actions:  []string{"read", "write", "delete"},
		TenantID: "tenant456",
	}

	// Test
	err := authManager.AuthorizeAction(user, "write")

	// Assert
	if err != nil {
		t.Fatalf("Expected no error for valid action, got: %v", err)
	}
}

func TestAuthorizeAction_InvalidAction(t *testing.T) {
	// Setup
	db := NewMockDatabaseManager()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	user := &User{
		ID:       "user123",
		Actions:  []string{"read"},
		TenantID: "tenant456",
	}

	// Test
	err := authManager.AuthorizeAction(user, "delete")

	// Assert
	if err == nil {
		t.Fatal("Expected error for invalid action, got nil")
	}

	// Check error code
	if frameworkErr, ok := err.(*FrameworkError); ok {
		if frameworkErr.Code != ErrCodeInsufficientActions {
			t.Errorf("Expected error code '%s', got '%s'", ErrCodeInsufficientActions, frameworkErr.Code)
		}
		if frameworkErr.StatusCode != 403 {
			t.Errorf("Expected status code 403, got %d", frameworkErr.StatusCode)
		}
	} else {
		t.Error("Expected FrameworkError type")
	}
}

func TestAuthorizeAction_NilUser(t *testing.T) {
	// Setup
	db := NewMockDatabaseManager()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	// Test
	err := authManager.AuthorizeAction(nil, "read")

	// Assert
	if err == nil {
		t.Fatal("Expected error for nil user, got nil")
	}
}

func TestAuthorizeAction_EmptyAction(t *testing.T) {
	// Setup
	db := NewMockDatabaseManager()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	user := &User{
		ID:      "user123",
		Actions: []string{"read"},
	}

	// Test
	err := authManager.AuthorizeAction(user, "")

	// Assert
	if err == nil {
		t.Fatal("Expected error for empty action, got nil")
	}
}

func TestAuthorizeActions_HasOneOfActions(t *testing.T) {
	// Setup
	db := NewMockDatabaseManager()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	user := &User{
		ID:       "user123",
		Actions:  []string{"read", "write"},
		TenantID: "tenant456",
	}

	// Test - user has "write" which is one of the required actions
	err := authManager.AuthorizeActions(user, []string{"delete", "write", "admin"})

	// Assert
	if err != nil {
		t.Fatalf("Expected no error when user has one of the required actions, got: %v", err)
	}
}

func TestAuthorizeActions_HasNoneOfActions(t *testing.T) {
	// Setup
	db := NewMockDatabaseManager()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	user := &User{
		ID:       "user123",
		Actions:  []string{"read"},
		TenantID: "tenant456",
	}

	// Test
	err := authManager.AuthorizeActions(user, []string{"delete", "write", "admin"})

	// Assert
	if err == nil {
		t.Fatal("Expected error when user has none of the required actions, got nil")
	}

	if frameworkErr, ok := err.(*FrameworkError); ok {
		if frameworkErr.Code != ErrCodeInsufficientActions {
			t.Errorf("Expected error code '%s', got '%s'", ErrCodeInsufficientActions, frameworkErr.Code)
		}
	}
}

func TestAuthorizeAllActions_HasAllActions(t *testing.T) {
	// Setup
	db := NewMockDatabaseManager()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	user := &User{
		ID:       "user123",
		Actions:  []string{"read", "write", "delete", "admin"},
		TenantID: "tenant456",
	}

	// Test
	err := authManager.AuthorizeAllActions(user, []string{"read", "write"})

	// Assert
	if err != nil {
		t.Fatalf("Expected no error when user has all required actions, got: %v", err)
	}
}

func TestAuthorizeAllActions_MissingOneAction(t *testing.T) {
	// Setup
	db := NewMockDatabaseManager()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	user := &User{
		ID:       "user123",
		Actions:  []string{"read", "write"},
		TenantID: "tenant456",
	}

	// Test
	err := authManager.AuthorizeAllActions(user, []string{"read", "write", "delete"})

	// Assert
	if err == nil {
		t.Fatal("Expected error when user is missing required actions, got nil")
	}

	if frameworkErr, ok := err.(*FrameworkError); ok {
		if frameworkErr.Code != ErrCodeInsufficientActions {
			t.Errorf("Expected error code '%s', got '%s'", ErrCodeInsufficientActions, frameworkErr.Code)
		}

		// Check that missing actions are reported
		if details, ok := frameworkErr.Details["missing_actions"].([]string); ok {
			if len(details) != 1 {
				t.Errorf("Expected 1 missing action, got %d", len(details))
			}
			if details[0] != "delete" {
				t.Errorf("Expected missing action 'delete', got '%s'", details[0])
			}
		}
	}
}

// Test comprehensive authorization
func TestAuthorize_BothRolesAndActions(t *testing.T) {
	// Setup
	db := NewMockDatabaseManager()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	user := &User{
		ID:       "user123",
		Roles:    []string{"admin", "user"},
		Actions:  []string{"read", "write", "delete"},
		TenantID: "tenant456",
	}

	// Test
	err := authManager.Authorize(user, []string{"admin"}, []string{"write"})

	// Assert
	if err != nil {
		t.Fatalf("Expected no error when user has both required roles and actions, got: %v", err)
	}
}

func TestAuthorize_MissingRole(t *testing.T) {
	// Setup
	db := NewMockDatabaseManager()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	user := &User{
		ID:       "user123",
		Roles:    []string{"user"},
		Actions:  []string{"read", "write"},
		TenantID: "tenant456",
	}

	// Test
	err := authManager.Authorize(user, []string{"admin"}, []string{"write"})

	// Assert
	if err == nil {
		t.Fatal("Expected error when user is missing required role, got nil")
	}

	if frameworkErr, ok := err.(*FrameworkError); ok {
		if frameworkErr.Code != ErrCodeInsufficientRoles {
			t.Errorf("Expected error code '%s', got '%s'", ErrCodeInsufficientRoles, frameworkErr.Code)
		}
	}
}

func TestAuthorize_MissingAction(t *testing.T) {
	// Setup
	db := NewMockDatabaseManager()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	user := &User{
		ID:       "user123",
		Roles:    []string{"admin", "user"},
		Actions:  []string{"read"},
		TenantID: "tenant456",
	}

	// Test
	err := authManager.Authorize(user, []string{"admin"}, []string{"delete"})

	// Assert
	if err == nil {
		t.Fatal("Expected error when user is missing required action, got nil")
	}

	if frameworkErr, ok := err.(*FrameworkError); ok {
		if frameworkErr.Code != ErrCodeInsufficientActions {
			t.Errorf("Expected error code '%s', got '%s'", ErrCodeInsufficientActions, frameworkErr.Code)
		}
	}
}

func TestAuthorize_OnlyRoles(t *testing.T) {
	// Setup
	db := NewMockDatabaseManager()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	user := &User{
		ID:       "user123",
		Roles:    []string{"admin"},
		TenantID: "tenant456",
	}

	// Test - only check roles, no actions required
	err := authManager.Authorize(user, []string{"admin"}, nil)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error when only checking roles, got: %v", err)
	}
}

func TestAuthorize_OnlyActions(t *testing.T) {
	// Setup
	db := NewMockDatabaseManager()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	user := &User{
		ID:       "user123",
		Actions:  []string{"read", "write"},
		TenantID: "tenant456",
	}

	// Test - only check actions, no roles required
	err := authManager.Authorize(user, nil, []string{"write"})

	// Assert
	if err != nil {
		t.Fatalf("Expected no error when only checking actions, got: %v", err)
	}
}

func TestAuthorize_NoRequirements(t *testing.T) {
	// Setup
	db := NewMockDatabaseManager()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	user := &User{
		ID:       "user123",
		TenantID: "tenant456",
	}

	// Test - no roles or actions required
	err := authManager.Authorize(user, nil, nil)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error when no requirements specified, got: %v", err)
	}
}

func TestAuthorize_NilUser(t *testing.T) {
	// Setup
	db := NewMockDatabaseManager()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	// Test
	err := authManager.Authorize(nil, []string{"admin"}, []string{"read"})

	// Assert
	if err == nil {
		t.Fatal("Expected error for nil user, got nil")
	}
}

// Property-based tests for scope authorization
// Feature: todo-implementations, Property 1: Complete scope verification
// Validates: Requirements 1.1
func TestProperty_CompleteScopeVerification(t *testing.T) {
	db := NewMockDatabaseManager()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	// Property: For any user and any set of required scopes, when all required scopes
	// are present in the user's scope list, authorization should succeed
	property := func(userScopes []string, requiredScopes []string) bool {
		// Skip if no required scopes (trivial case)
		if len(requiredScopes) == 0 {
			return true
		}

		// Create user with all required scopes plus potentially more
		allScopes := make([]string, len(userScopes))
		copy(allScopes, userScopes)

		// Ensure all required scopes are in user's scopes
		scopeMap := make(map[string]bool)
		for _, s := range allScopes {
			scopeMap[s] = true
		}
		for _, rs := range requiredScopes {
			if !scopeMap[rs] {
				allScopes = append(allScopes, rs)
			}
		}

		user := &User{
			ID:       "test-user",
			Scopes:   allScopes,
			TenantID: "test-tenant",
		}

		// Test authorization - should succeed
		err := authManager.AuthorizeAllScopes(user, requiredScopes)
		return err == nil
	}

	// Run property test with 100 iterations
	for i := 0; i < 100; i++ {
		// Generate random scopes
		userScopes := generateRandomScopes(i % 10)
		requiredScopes := generateRandomScopes((i % 5) + 1)

		if !property(userScopes, requiredScopes) {
			t.Errorf("Property failed: user with all required scopes should be authorized")
			t.Errorf("User scopes: %v, Required scopes: %v", userScopes, requiredScopes)
		}
	}
}

// Feature: todo-implementations, Property 2: Missing scope detection
// Validates: Requirements 1.2
func TestProperty_MissingScopeDetection(t *testing.T) {
	db := NewMockDatabaseManager()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	// Property: For any user missing at least one required scope, authorization should
	// fail with a 403 response containing details about the missing scopes
	property := func(userScopes []string, missingScope string) bool {
		// Skip if missing scope is in user scopes
		for _, s := range userScopes {
			if s == missingScope {
				return true // Skip this test case
			}
		}

		user := &User{
			ID:       "test-user",
			Scopes:   userScopes,
			TenantID: "test-tenant",
		}

		// Test authorization with a scope the user doesn't have
		err := authManager.AuthorizeAllScopes(user, []string{missingScope})

		if err == nil {
			return false // Should have failed
		}

		// Check that it's a FrameworkError with correct code and status
		frameworkErr, ok := err.(*FrameworkError)
		if !ok {
			return false
		}

		if frameworkErr.Code != ErrCodeInsufficientScopes {
			return false
		}

		if frameworkErr.StatusCode != 403 {
			return false
		}

		// Check that missing scopes are reported in details
		if details, ok := frameworkErr.Details["missing_scopes"].([]string); ok {
			if len(details) == 0 {
				return false
			}
			// Verify the missing scope is in the details
			found := false
			for _, ms := range details {
				if ms == missingScope {
					found = true
					break
				}
			}
			return found
		}

		return false
	}

	// Run property test with 100 iterations
	for i := 0; i < 100; i++ {
		userScopes := generateRandomScopes(i % 10)
		missingScope := generateRandomScope(i + 1000)

		if !property(userScopes, missingScope) {
			t.Errorf("Property failed: user missing required scope should get 403 with details")
			t.Errorf("User scopes: %v, Missing scope: %s", userScopes, missingScope)
		}
	}
}

// Feature: todo-implementations, Property 3: Wildcard scope grants universal access
// Validates: Requirements 1.4
func TestProperty_WildcardScopeGrantsUniversalAccess(t *testing.T) {
	db := NewMockDatabaseManager()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	// Property: For any user with the wildcard scope "*", authorization should succeed
	// regardless of which scopes are required
	property := func(requiredScopes []string) bool {
		// Skip empty required scopes
		if len(requiredScopes) == 0 {
			return true
		}

		user := &User{
			ID:       "test-user",
			Scopes:   []string{"*"},
			TenantID: "test-tenant",
		}

		// Test authorization - should always succeed with wildcard
		err := authManager.AuthorizeAllScopes(user, requiredScopes)
		return err == nil
	}

	// Run property test with 100 iterations
	for i := 0; i < 100; i++ {
		requiredScopes := generateRandomScopes((i % 10) + 1)

		if !property(requiredScopes) {
			t.Errorf("Property failed: wildcard scope should grant access to all scopes")
			t.Errorf("Required scopes: %v", requiredScopes)
		}
	}
}

// Feature: todo-implementations, Property 4: Hierarchical scope matching
// Validates: Requirements 1.5
func TestProperty_HierarchicalScopeMatching(t *testing.T) {
	db := NewMockDatabaseManager()
	authManager := NewAuthManager(db, "test-secret", OAuth2Config{})

	// Property: For any user scope in hierarchical format (e.g., "admin"), it should
	// match any required scope with that prefix (e.g., "admin:read", "admin:write")
	property := func(baseScope string, subAction string) bool {
		// Skip empty inputs
		if baseScope == "" || subAction == "" {
			return true
		}

		// Create hierarchical scope
		hierarchicalScope := baseScope + ":" + subAction

		user := &User{
			ID:       "test-user",
			Scopes:   []string{baseScope},
			TenantID: "test-tenant",
		}

		// Test authorization - base scope should match hierarchical scope
		err := authManager.AuthorizeScope(user, hierarchicalScope)
		return err == nil
	}

	// Run property test with 100 iterations
	for i := 0; i < 100; i++ {
		baseScope := generateRandomScope(i)
		subAction := generateRandomScope(i + 500)

		if !property(baseScope, subAction) {
			t.Errorf("Property failed: base scope should match hierarchical scope")
			t.Errorf("Base scope: %s, Sub-action: %s", baseScope, subAction)
		}
	}
}

// Helper functions for generating random test data
func generateRandomScopes(count int) []string {
	if count <= 0 {
		return []string{}
	}

	scopes := make([]string, count)
	for i := 0; i < count; i++ {
		scopes[i] = generateRandomScope(i)
	}
	return scopes
}

func generateRandomScope(seed int) string {
	resources := []string{"users", "posts", "comments", "admin", "api", "data", "files"}
	actions := []string{"read", "write", "delete", "create", "update", "list"}

	resource := resources[seed%len(resources)]
	action := actions[(seed/len(resources))%len(actions)]

	return resource + ":" + action
}
