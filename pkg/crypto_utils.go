package pkg

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// GenerateEncryptionKey generates a cryptographically secure encryption key
// of the specified length in bytes. For AES-256, use 32 bytes.
func GenerateEncryptionKey(length int) ([]byte, error) {
	if length <= 0 {
		return nil, fmt.Errorf("key length must be positive, got %d", length)
	}

	key := make([]byte, length)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate random key: %w", err)
	}

	return key, nil
}

// GenerateEncryptionKeyHex generates a cryptographically secure encryption key
// and returns it as a hex-encoded string. For AES-256, use 32 bytes.
func GenerateEncryptionKeyHex(length int) (string, error) {
	key, err := GenerateEncryptionKey(length)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(key), nil
}

// GenerateJWTSecret generates a cryptographically secure JWT secret
// Returns a hex-encoded string of the specified length in bytes.
// Recommended: at least 32 bytes for HS256.
func GenerateJWTSecret(length int) (string, error) {
	return GenerateEncryptionKeyHex(length)
}

// MustGenerateEncryptionKey generates a key or panics. Use only for testing.
func MustGenerateEncryptionKey(length int) []byte {
	key, err := GenerateEncryptionKey(length)
	if err != nil {
		panic(fmt.Sprintf("failed to generate encryption key: %v", err))
	}
	return key
}

// MustGenerateEncryptionKeyHex generates a hex key or panics. Use only for testing.
func MustGenerateEncryptionKeyHex(length int) string {
	key, err := GenerateEncryptionKeyHex(length)
	if err != nil {
		panic(fmt.Sprintf("failed to generate encryption key: %v", err))
	}
	return key
}
