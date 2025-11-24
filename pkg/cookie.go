package pkg

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// CookieManager provides cookie management functionality with encryption support
type CookieManager interface {
	// Cookie operations
	SetCookie(ctx Context, cookie *Cookie) error
	GetCookie(ctx Context, name string) (*Cookie, error)
	GetAllCookies(ctx Context) ([]*Cookie, error)
	DeleteCookie(ctx Context, name string) error

	// Encrypted cookie operations
	SetEncryptedCookie(ctx Context, cookie *Cookie) error
	GetEncryptedCookie(ctx Context, name string) (*Cookie, error)

	// Cookie value encryption/decryption
	EncryptValue(value string) (string, error)
	DecryptValue(encryptedValue string) (string, error)
}

// CookieConfig holds cookie manager configuration
type CookieConfig struct {
	EncryptionKey   []byte        // AES-256 key (32 bytes)
	DefaultPath     string        // Default cookie path
	DefaultDomain   string        // Default cookie domain
	DefaultSecure   bool          // Default secure flag
	DefaultHTTPOnly bool          // Default HTTP-only flag
	DefaultSameSite http.SameSite // Default SameSite policy
	DefaultMaxAge   int           // Default max age in seconds
}

// DefaultCookieConfig returns default cookie configuration
func DefaultCookieConfig() *CookieConfig {
	return &CookieConfig{
		DefaultPath:     "/",
		DefaultDomain:   "",
		DefaultSecure:   true,
		DefaultHTTPOnly: true,
		DefaultSameSite: http.SameSiteLaxMode,
		DefaultMaxAge:   86400, // 24 hours
	}
}

// cookieManager implements CookieManager interface
type cookieManager struct {
	config *CookieConfig
	cipher cipher.Block
}

// NewCookieManager creates a new cookie manager instance
func NewCookieManager(config *CookieConfig) (CookieManager, error) {
	if config == nil {
		config = DefaultCookieConfig()
	}

	var block cipher.Block
	var err error

	// Create cipher if encryption key is provided
	if len(config.EncryptionKey) > 0 {
		if len(config.EncryptionKey) != 32 {
			return nil, errors.New("encryption key must be 32 bytes for AES-256")
		}

		block, err = aes.NewCipher(config.EncryptionKey)
		if err != nil {
			return nil, fmt.Errorf("failed to create cipher: %w", err)
		}
	}

	return &cookieManager{
		config: config,
		cipher: block,
	}, nil
}

// SetCookie sets a cookie in the response
func (cm *cookieManager) SetCookie(ctx Context, cookie *Cookie) error {
	if ctx == nil {
		return errors.New("context is required")
	}

	if cookie == nil {
		return errors.New("cookie is required")
	}

	// Apply defaults if not set
	cm.applyDefaults(cookie)

	// Set cookie through context
	return ctx.SetCookie(cookie)
}

// GetCookie retrieves a cookie from the request
func (cm *cookieManager) GetCookie(ctx Context, name string) (*Cookie, error) {
	if ctx == nil {
		return nil, errors.New("context is required")
	}

	if name == "" {
		return nil, errors.New("cookie name is required")
	}

	return ctx.GetCookie(name)
}

// GetAllCookies retrieves all cookies from the request
func (cm *cookieManager) GetAllCookies(ctx Context) ([]*Cookie, error) {
	if ctx == nil {
		return nil, errors.New("context is required")
	}

	req := ctx.Request()
	if req == nil {
		return nil, errors.New("request is required")
	}

	cookieHeader := req.Header.Get("Cookie")
	if cookieHeader == "" {
		return []*Cookie{}, nil
	}

	// Parse all cookies from header
	cookies := []*Cookie{}
	cookiePairs := strings.Split(cookieHeader, ";")

	for _, pair := range cookiePairs {
		parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(parts) == 2 {
			cookies = append(cookies, &Cookie{
				Name:  parts[0],
				Value: parts[1],
			})
		}
	}

	return cookies, nil
}

// DeleteCookie deletes a cookie by setting it to expire immediately
func (cm *cookieManager) DeleteCookie(ctx Context, name string) error {
	if ctx == nil {
		return errors.New("context is required")
	}

	if name == "" {
		return errors.New("cookie name is required")
	}

	// Create an expired cookie
	cookie := &Cookie{
		Name:     name,
		Value:    "",
		Path:     cm.config.DefaultPath,
		Domain:   cm.config.DefaultDomain,
		Expires:  time.Unix(0, 0), // Epoch time
		MaxAge:   -1,              // Delete immediately
		Secure:   cm.config.DefaultSecure,
		HttpOnly: cm.config.DefaultHTTPOnly,
		SameSite: cm.config.DefaultSameSite,
	}

	return ctx.SetCookie(cookie)
}

// SetEncryptedCookie sets an encrypted cookie
func (cm *cookieManager) SetEncryptedCookie(ctx Context, cookie *Cookie) error {
	if cm.cipher == nil {
		return errors.New("encryption not configured")
	}

	if cookie == nil {
		return errors.New("cookie is required")
	}

	// Encrypt the cookie value
	encryptedValue, err := cm.EncryptValue(cookie.Value)
	if err != nil {
		return fmt.Errorf("failed to encrypt cookie value: %w", err)
	}

	// Create new cookie with encrypted value
	encryptedCookie := &Cookie{
		Name:      cookie.Name,
		Value:     encryptedValue,
		Path:      cookie.Path,
		Domain:    cookie.Domain,
		Expires:   cookie.Expires,
		MaxAge:    cookie.MaxAge,
		Secure:    cookie.Secure,
		HttpOnly:  cookie.HttpOnly,
		SameSite:  cookie.SameSite,
		Encrypted: true,
	}

	return cm.SetCookie(ctx, encryptedCookie)
}

// GetEncryptedCookie retrieves and decrypts an encrypted cookie
func (cm *cookieManager) GetEncryptedCookie(ctx Context, name string) (*Cookie, error) {
	if cm.cipher == nil {
		return nil, errors.New("encryption not configured")
	}

	// Get the encrypted cookie
	cookie, err := cm.GetCookie(ctx, name)
	if err != nil {
		return nil, err
	}

	// Decrypt the value
	decryptedValue, err := cm.DecryptValue(cookie.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt cookie value: %w", err)
	}

	// Return cookie with decrypted value
	cookie.Value = decryptedValue
	cookie.Encrypted = false

	return cookie, nil
}

// EncryptValue encrypts a cookie value using AES-GCM
func (cm *cookieManager) EncryptValue(value string) (string, error) {
	if cm.cipher == nil {
		return "", errors.New("encryption not configured")
	}

	plaintext := []byte(value)

	// Create GCM cipher mode
	gcm, err := cipher.NewGCM(cm.cipher)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt the data
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	// Encode to base64 for cookie storage
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

// DecryptValue decrypts an encrypted cookie value
func (cm *cookieManager) DecryptValue(encryptedValue string) (string, error) {
	if cm.cipher == nil {
		return "", errors.New("encryption not configured")
	}

	// Decode from base64
	ciphertext, err := base64.URLEncoding.DecodeString(encryptedValue)
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted value: %w", err)
	}

	// Create GCM cipher mode
	gcm, err := cipher.NewGCM(cm.cipher)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
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
		return "", fmt.Errorf("failed to decrypt value: %w", err)
	}

	return string(plaintext), nil
}

// applyDefaults applies default configuration to a cookie
func (cm *cookieManager) applyDefaults(cookie *Cookie) {
	if cookie.Path == "" {
		cookie.Path = cm.config.DefaultPath
	}

	if cookie.Domain == "" {
		cookie.Domain = cm.config.DefaultDomain
	}

	// Only apply defaults if not explicitly set
	if cookie.MaxAge == 0 && cookie.Expires.IsZero() {
		cookie.MaxAge = cm.config.DefaultMaxAge
	}

	// Apply security defaults if not explicitly set to false
	if !cookie.Secure && cm.config.DefaultSecure {
		cookie.Secure = cm.config.DefaultSecure
	}

	if !cookie.HttpOnly && cm.config.DefaultHTTPOnly {
		cookie.HttpOnly = cm.config.DefaultHTTPOnly
	}

	if cookie.SameSite == 0 {
		cookie.SameSite = cm.config.DefaultSameSite
	}
}

// HeaderManager provides header management functionality
type HeaderManager interface {
	// Request header operations
	GetHeader(ctx Context, key string) string
	GetAllHeaders(ctx Context) map[string]string
	GetHeaderValues(ctx Context, key string) []string
	HasHeader(ctx Context, key string) bool

	// Response header operations
	SetHeader(ctx Context, key, value string) error
	SetHeaders(ctx Context, headers map[string]string) error
	AddHeader(ctx Context, key, value string) error
	DeleteHeader(ctx Context, key string) error

	// Common header helpers
	SetContentType(ctx Context, contentType string) error
	SetCacheControl(ctx Context, directive string) error
	SetLocation(ctx Context, url string) error
	SetAuthorization(ctx Context, token string) error
	GetAuthorization(ctx Context) string
	GetUserAgent(ctx Context) string
	GetReferer(ctx Context) string
	GetContentType(ctx Context) string
}

// headerManager implements HeaderManager interface
type headerManager struct{}

// NewHeaderManager creates a new header manager instance
func NewHeaderManager() HeaderManager {
	return &headerManager{}
}

// GetHeader retrieves a request header value
func (hm *headerManager) GetHeader(ctx Context, key string) string {
	if ctx == nil {
		return ""
	}

	return ctx.GetHeader(key)
}

// GetAllHeaders retrieves all request headers
func (hm *headerManager) GetAllHeaders(ctx Context) map[string]string {
	if ctx == nil {
		return make(map[string]string)
	}

	return ctx.Headers()
}

// GetHeaderValues retrieves all values for a request header
func (hm *headerManager) GetHeaderValues(ctx Context, key string) []string {
	if ctx == nil {
		return []string{}
	}

	req := ctx.Request()
	if req == nil || req.Header == nil {
		return []string{}
	}

	return req.Header.Values(key)
}

// HasHeader checks if a request header exists
func (hm *headerManager) HasHeader(ctx Context, key string) bool {
	if ctx == nil {
		return false
	}

	req := ctx.Request()
	if req == nil || req.Header == nil {
		return false
	}

	return req.Header.Get(key) != ""
}

// SetHeader sets a response header
func (hm *headerManager) SetHeader(ctx Context, key, value string) error {
	if ctx == nil {
		return errors.New("context is required")
	}

	if key == "" {
		return errors.New("header key is required")
	}

	ctx.SetHeader(key, value)
	return nil
}

// SetHeaders sets multiple response headers
func (hm *headerManager) SetHeaders(ctx Context, headers map[string]string) error {
	if ctx == nil {
		return errors.New("context is required")
	}

	for key, value := range headers {
		ctx.SetHeader(key, value)
	}

	return nil
}

// AddHeader adds a response header (allows multiple values)
func (hm *headerManager) AddHeader(ctx Context, key, value string) error {
	if ctx == nil {
		return errors.New("context is required")
	}

	if key == "" {
		return errors.New("header key is required")
	}

	resp := ctx.Response()
	if resp == nil {
		return errors.New("response is required")
	}

	resp.Header().Add(key, value)
	return nil
}

// DeleteHeader deletes a response header
func (hm *headerManager) DeleteHeader(ctx Context, key string) error {
	if ctx == nil {
		return errors.New("context is required")
	}

	if key == "" {
		return errors.New("header key is required")
	}

	resp := ctx.Response()
	if resp == nil {
		return errors.New("response is required")
	}

	resp.Header().Del(key)
	return nil
}

// SetContentType sets the Content-Type header
func (hm *headerManager) SetContentType(ctx Context, contentType string) error {
	return hm.SetHeader(ctx, "Content-Type", contentType)
}

// SetCacheControl sets the Cache-Control header
func (hm *headerManager) SetCacheControl(ctx Context, directive string) error {
	return hm.SetHeader(ctx, "Cache-Control", directive)
}

// SetLocation sets the Location header (for redirects)
func (hm *headerManager) SetLocation(ctx Context, url string) error {
	return hm.SetHeader(ctx, "Location", url)
}

// SetAuthorization sets the Authorization header
func (hm *headerManager) SetAuthorization(ctx Context, token string) error {
	return hm.SetHeader(ctx, "Authorization", token)
}

// GetAuthorization retrieves the Authorization header
func (hm *headerManager) GetAuthorization(ctx Context) string {
	return hm.GetHeader(ctx, "Authorization")
}

// GetUserAgent retrieves the User-Agent header
func (hm *headerManager) GetUserAgent(ctx Context) string {
	return hm.GetHeader(ctx, "User-Agent")
}

// GetReferer retrieves the Referer header
func (hm *headerManager) GetReferer(ctx Context) string {
	return hm.GetHeader(ctx, "Referer")
}

// GetContentType retrieves the Content-Type header
func (hm *headerManager) GetContentType(ctx Context) string {
	return hm.GetHeader(ctx, "Content-Type")
}
