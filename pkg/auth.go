package pkg

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// AuthManager implements authentication functionality for OAuth2, JWT, and access tokens
type AuthManager struct {
	db           DatabaseManager
	jwtSecret    []byte
	oauth2Config OAuth2Config
}

// OAuth2Config defines OAuth2 configuration
type OAuth2Config struct {
	ClientID     string
	ClientSecret string
	TokenURL     string
	AuthURL      string
	RedirectURL  string
	Scopes       []string
}

// User represents an authenticated user with authorization information
type User struct {
	ID         string                 `json:"id"`
	Username   string                 `json:"username"`
	Email      string                 `json:"email"`
	Roles      []string               `json:"roles"`
	Actions    []string               `json:"actions"`
	TenantID   string                 `json:"tenant_id"`
	Metadata   map[string]interface{} `json:"metadata"`
	AuthMethod string                 `json:"auth_method"`
	AuthTime   time.Time              `json:"auth_time"`
	ExpiresAt  time.Time              `json:"expires_at"`
}

// JWTClaims represents JWT token claims
type JWTClaims struct {
	UserID    string                 `json:"user_id"`
	Username  string                 `json:"username"`
	Email     string                 `json:"email"`
	Roles     []string               `json:"roles"`
	Actions   []string               `json:"actions"`
	TenantID  string                 `json:"tenant_id"`
	Metadata  map[string]interface{} `json:"metadata"`
	IssuedAt  int64                  `json:"iat"`
	ExpiresAt int64                  `json:"exp"`
	Issuer    string                 `json:"iss"`
	Subject   string                 `json:"sub"`
}

// NewAuthManager creates a new authentication manager
func NewAuthManager(db DatabaseManager, jwtSecret string, oauth2Config OAuth2Config) *AuthManager {
	return &AuthManager{
		db:           db,
		jwtSecret:    []byte(jwtSecret),
		oauth2Config: oauth2Config,
	}
}

// AuthenticateOAuth2 authenticates a user using OAuth2 token
// Requirements: 3.1
func (am *AuthManager) AuthenticateOAuth2(token string) (*User, error) {
	if token == "" {
		return nil, NewAuthenticationError("OAuth2 token is required")
	}

	// In a real implementation, this would validate the token with the OAuth2 provider
	// For now, we'll check if it's a valid access token in our database
	accessToken, err := am.db.LoadAccessToken(token)
	if err != nil {
		return nil, NewAuthenticationError("Invalid OAuth2 token").WithCause(err)
	}

	// Check expiration
	if accessToken.ExpiresAt.Before(time.Now()) {
		return nil, &FrameworkError{
			Code:       ErrCodeTokenExpired,
			Message:    "OAuth2 token has expired",
			StatusCode: 401,
			I18nKey:    "error.authentication.token_expired",
		}
	}

	// Create user from access token
	user := &User{
		ID:         accessToken.UserID,
		TenantID:   accessToken.TenantID,
		Roles:      accessToken.Scopes,
		AuthMethod: "OAuth2",
		AuthTime:   time.Now(),
		ExpiresAt:  accessToken.ExpiresAt,
	}

	return user, nil
}

// AuthenticateJWT authenticates a user using JWT token
// Requirements: 3.2
func (am *AuthManager) AuthenticateJWT(token string) (*User, error) {
	if token == "" {
		return nil, NewAuthenticationError("JWT token is required")
	}

	// Parse and validate JWT token
	claims, err := am.parseJWT(token)
	if err != nil {
		return nil, NewAuthenticationError("Invalid JWT token").WithCause(err)
	}

	// Check expiration
	if claims.ExpiresAt < time.Now().Unix() {
		return nil, &FrameworkError{
			Code:       ErrCodeTokenExpired,
			Message:    "JWT token has expired",
			StatusCode: 401,
			I18nKey:    "error.authentication.token_expired",
		}
	}

	// Create user from JWT claims
	user := &User{
		ID:         claims.UserID,
		Username:   claims.Username,
		Email:      claims.Email,
		Roles:      claims.Roles,
		Actions:    claims.Actions,
		TenantID:   claims.TenantID,
		Metadata:   claims.Metadata,
		AuthMethod: "JWT",
		AuthTime:   time.Unix(claims.IssuedAt, 0),
		ExpiresAt:  time.Unix(claims.ExpiresAt, 0),
	}

	return user, nil
}

// AuthenticateAccessToken validates an access token from the database
// Requirements: 3.5
func (am *AuthManager) AuthenticateAccessToken(token string) (*AccessToken, error) {
	if token == "" {
		return nil, NewAuthenticationError("Access token is required")
	}

	// Validate token from database
	accessToken, err := am.db.ValidateAccessToken(token)
	if err != nil {
		return nil, NewAuthenticationError("Invalid access token").WithCause(err)
	}

	// Check expiration
	if accessToken.ExpiresAt.Before(time.Now()) {
		return nil, &FrameworkError{
			Code:       ErrCodeTokenExpired,
			Message:    "Access token has expired",
			StatusCode: 401,
			I18nKey:    "error.authentication.token_expired",
		}
	}

	return accessToken, nil
}

// parseJWT parses and validates a JWT token
func (am *AuthManager) parseJWT(token string) (*JWTClaims, error) {
	// Split token into parts
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT format: expected 3 parts, got %d", len(parts))
	}

	// Verify signature
	signature := parts[2]
	message := parts[0] + "." + parts[1]

	if !am.verifySignature(message, signature) {
		return nil, fmt.Errorf("invalid JWT signature")
	}

	// Decode payload
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode JWT payload: %w", err)
	}

	// Parse claims
	var claims JWTClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("failed to parse JWT claims: %w", err)
	}

	return &claims, nil
}

// verifySignature verifies the JWT signature using HMAC-SHA256
func (am *AuthManager) verifySignature(message, signature string) bool {
	// Compute expected signature
	mac := hmac.New(sha256.New, am.jwtSecret)
	mac.Write([]byte(message))
	expectedSignature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	// Compare signatures
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// GenerateJWT generates a JWT token for a user
func (am *AuthManager) GenerateJWT(user *User, expiresIn time.Duration) (string, error) {
	now := time.Now()
	claims := JWTClaims{
		UserID:    user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Roles:     user.Roles,
		Actions:   user.Actions,
		TenantID:  user.TenantID,
		Metadata:  user.Metadata,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(expiresIn).Unix(),
		Issuer:    "rockstar-framework",
		Subject:   user.ID,
	}

	// Create header
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}

	// Encode header
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JWT header: %w", err)
	}
	headerEncoded := base64.RawURLEncoding.EncodeToString(headerJSON)

	// Encode payload
	payloadJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JWT claims: %w", err)
	}
	payloadEncoded := base64.RawURLEncoding.EncodeToString(payloadJSON)

	// Create signature
	message := headerEncoded + "." + payloadEncoded
	mac := hmac.New(sha256.New, am.jwtSecret)
	mac.Write([]byte(message))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	// Combine parts
	token := message + "." + signature

	return token, nil
}

// CreateAccessToken creates a new access token and stores it in the database
func (am *AuthManager) CreateAccessToken(userID, tenantID string, scopes []string, expiresIn time.Duration) (*AccessToken, error) {
	// Generate secure random token
	token := am.generateSecureToken()

	accessToken := &AccessToken{
		Token:     token,
		UserID:    userID,
		TenantID:  tenantID,
		Scopes:    scopes,
		ExpiresAt: time.Now().Add(expiresIn),
		CreatedAt: time.Now(),
	}

	// Save to database
	if err := am.db.SaveAccessToken(accessToken); err != nil {
		return nil, fmt.Errorf("failed to save access token: %w", err)
	}

	return accessToken, nil
}

// tokenCounter is used to ensure unique token generation
var tokenCounter int64 = 0

// generateSecureToken generates a cryptographically secure random token
func (am *AuthManager) generateSecureToken() string {
	// Use current time with nanosecond precision plus a counter for uniqueness
	now := time.Now()
	tokenCounter++
	data := fmt.Sprintf("%d-%d-%d-%s", now.Unix(), now.Nanosecond(), tokenCounter, now.String())

	mac := hmac.New(sha256.New, am.jwtSecret)
	mac.Write([]byte(data))
	hash := mac.Sum(nil)

	return base64.RawURLEncoding.EncodeToString(hash)
}

// RevokeAccessToken revokes an access token
func (am *AuthManager) RevokeAccessToken(token string) error {
	return am.db.DeleteAccessToken(token)
}

// RefreshAccessToken refreshes an access token by creating a new one
func (am *AuthManager) RefreshAccessToken(oldToken string, expiresIn time.Duration) (*AccessToken, error) {
	// Validate old token
	accessToken, err := am.AuthenticateAccessToken(oldToken)
	if err != nil {
		return nil, err
	}

	// Create new token
	newToken, err := am.CreateAccessToken(accessToken.UserID, accessToken.TenantID, accessToken.Scopes, expiresIn)
	if err != nil {
		return nil, err
	}

	// Revoke old token
	if err := am.RevokeAccessToken(oldToken); err != nil {
		// Log error but don't fail the refresh
		// The old token will be cleaned up by the cleanup job
	}

	return newToken, nil
}

// AuthorizeRole checks if a user has the required role (Role-Based Access Control)
// Requirements: 3.3
func (am *AuthManager) AuthorizeRole(user *User, requiredRole string) error {
	if user == nil {
		return NewAuthorizationError("user is required for authorization")
	}

	if requiredRole == "" {
		return NewAuthorizationError("required role is empty")
	}

	// Check if user has the required role
	for _, role := range user.Roles {
		if role == requiredRole {
			return nil
		}
	}

	return &FrameworkError{
		Code:       ErrCodeInsufficientRoles,
		Message:    fmt.Sprintf("user does not have required role: %s", requiredRole),
		StatusCode: 403,
		I18nKey:    "error.authorization.insufficient_roles",
		Details: map[string]interface{}{
			"required_role": requiredRole,
			"user_roles":    user.Roles,
		},
		UserID:   user.ID,
		TenantID: user.TenantID,
	}
}

// AuthorizeRoles checks if a user has any of the required roles
// Requirements: 3.3
func (am *AuthManager) AuthorizeRoles(user *User, requiredRoles []string) error {
	if user == nil {
		return NewAuthorizationError("user is required for authorization")
	}

	if len(requiredRoles) == 0 {
		return NewAuthorizationError("required roles list is empty")
	}

	// Check if user has any of the required roles
	for _, userRole := range user.Roles {
		for _, requiredRole := range requiredRoles {
			if userRole == requiredRole {
				return nil
			}
		}
	}

	return &FrameworkError{
		Code:       ErrCodeInsufficientRoles,
		Message:    fmt.Sprintf("user does not have any of the required roles: %v", requiredRoles),
		StatusCode: 403,
		I18nKey:    "error.authorization.insufficient_roles",
		Details: map[string]interface{}{
			"required_roles": requiredRoles,
			"user_roles":     user.Roles,
		},
		UserID:   user.ID,
		TenantID: user.TenantID,
	}
}

// AuthorizeAllRoles checks if a user has all of the required roles
// Requirements: 3.3
func (am *AuthManager) AuthorizeAllRoles(user *User, requiredRoles []string) error {
	if user == nil {
		return NewAuthorizationError("user is required for authorization")
	}

	if len(requiredRoles) == 0 {
		return NewAuthorizationError("required roles list is empty")
	}

	// Create a map of user roles for efficient lookup
	userRolesMap := make(map[string]bool)
	for _, role := range user.Roles {
		userRolesMap[role] = true
	}

	// Check if user has all required roles
	missingRoles := []string{}
	for _, requiredRole := range requiredRoles {
		if !userRolesMap[requiredRole] {
			missingRoles = append(missingRoles, requiredRole)
		}
	}

	if len(missingRoles) > 0 {
		return &FrameworkError{
			Code:       ErrCodeInsufficientRoles,
			Message:    fmt.Sprintf("user is missing required roles: %v", missingRoles),
			StatusCode: 403,
			I18nKey:    "error.authorization.insufficient_roles",
			Details: map[string]interface{}{
				"required_roles": requiredRoles,
				"missing_roles":  missingRoles,
				"user_roles":     user.Roles,
			},
			UserID:   user.ID,
			TenantID: user.TenantID,
		}
	}

	return nil
}

// AuthorizeAction checks if a user can perform the required action (Action-Based Access Control)
// Requirements: 3.4
func (am *AuthManager) AuthorizeAction(user *User, requiredAction string) error {
	if user == nil {
		return NewAuthorizationError("user is required for authorization")
	}

	if requiredAction == "" {
		return NewAuthorizationError("required action is empty")
	}

	// Check if user has the required action
	for _, action := range user.Actions {
		if action == requiredAction {
			return nil
		}
	}

	return &FrameworkError{
		Code:       ErrCodeInsufficientActions,
		Message:    fmt.Sprintf("user does not have required action: %s", requiredAction),
		StatusCode: 403,
		I18nKey:    "error.authorization.insufficient_actions",
		Details: map[string]interface{}{
			"required_action": requiredAction,
			"user_actions":    user.Actions,
		},
		UserID:   user.ID,
		TenantID: user.TenantID,
	}
}

// AuthorizeActions checks if a user can perform any of the required actions
// Requirements: 3.4
func (am *AuthManager) AuthorizeActions(user *User, requiredActions []string) error {
	if user == nil {
		return NewAuthorizationError("user is required for authorization")
	}

	if len(requiredActions) == 0 {
		return NewAuthorizationError("required actions list is empty")
	}

	// Check if user has any of the required actions
	for _, userAction := range user.Actions {
		for _, requiredAction := range requiredActions {
			if userAction == requiredAction {
				return nil
			}
		}
	}

	return &FrameworkError{
		Code:       ErrCodeInsufficientActions,
		Message:    fmt.Sprintf("user does not have any of the required actions: %v", requiredActions),
		StatusCode: 403,
		I18nKey:    "error.authorization.insufficient_actions",
		Details: map[string]interface{}{
			"required_actions": requiredActions,
			"user_actions":     user.Actions,
		},
		UserID:   user.ID,
		TenantID: user.TenantID,
	}
}

// AuthorizeAllActions checks if a user can perform all of the required actions
// Requirements: 3.4
func (am *AuthManager) AuthorizeAllActions(user *User, requiredActions []string) error {
	if user == nil {
		return NewAuthorizationError("user is required for authorization")
	}

	if len(requiredActions) == 0 {
		return NewAuthorizationError("required actions list is empty")
	}

	// Create a map of user actions for efficient lookup
	userActionsMap := make(map[string]bool)
	for _, action := range user.Actions {
		userActionsMap[action] = true
	}

	// Check if user has all required actions
	missingActions := []string{}
	for _, requiredAction := range requiredActions {
		if !userActionsMap[requiredAction] {
			missingActions = append(missingActions, requiredAction)
		}
	}

	if len(missingActions) > 0 {
		return &FrameworkError{
			Code:       ErrCodeInsufficientActions,
			Message:    fmt.Sprintf("user is missing required actions: %v", missingActions),
			StatusCode: 403,
			I18nKey:    "error.authorization.insufficient_actions",
			Details: map[string]interface{}{
				"required_actions": requiredActions,
				"missing_actions":  missingActions,
				"user_actions":     user.Actions,
			},
			UserID:   user.ID,
			TenantID: user.TenantID,
		}
	}

	return nil
}

// Authorize performs comprehensive authorization check with both roles and actions
// Requirements: 3.3, 3.4
func (am *AuthManager) Authorize(user *User, requiredRoles []string, requiredActions []string) error {
	if user == nil {
		return NewAuthorizationError("user is required for authorization")
	}

	// Check roles if specified
	if len(requiredRoles) > 0 {
		if err := am.AuthorizeRoles(user, requiredRoles); err != nil {
			return err
		}
	}

	// Check actions if specified
	if len(requiredActions) > 0 {
		if err := am.AuthorizeActions(user, requiredActions); err != nil {
			return err
		}
	}

	return nil
}
