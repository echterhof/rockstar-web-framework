package pkg

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

// PasswordHasher defines the interface for password hashing operations
type PasswordHasher interface {
	// Hash generates a hash from a password
	Hash(password string) (string, error)

	// Verify checks if a password matches a hash
	Verify(password, hash string) (bool, error)

	// NeedsRehash checks if a hash needs to be regenerated (e.g., due to updated parameters)
	NeedsRehash(hash string) bool
}

// PasswordHashAlgorithm represents the hashing algorithm to use
type PasswordHashAlgorithm string

const (
	// AlgorithmBcrypt uses bcrypt (recommended for most use cases)
	AlgorithmBcrypt PasswordHashAlgorithm = "bcrypt"

	// AlgorithmArgon2id uses Argon2id (recommended for high-security applications)
	AlgorithmArgon2id PasswordHashAlgorithm = "argon2id"
)

// BcryptHasher implements password hashing using bcrypt
type BcryptHasher struct {
	cost int // Cost factor (4-31, default 12)
}

// NewBcryptHasher creates a new bcrypt password hasher
func NewBcryptHasher(cost int) *BcryptHasher {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = bcrypt.DefaultCost // 12
	}
	return &BcryptHasher{cost: cost}
}

// Hash generates a bcrypt hash from a password
func (b *BcryptHasher) Hash(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	// Bcrypt has a maximum password length of 72 bytes
	if len(password) > 72 {
		return "", fmt.Errorf("password exceeds maximum length of 72 bytes")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), b.cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hash), nil
}

// Verify checks if a password matches a bcrypt hash
func (b *BcryptHasher) Verify(password, hash string) (bool, error) {
	if password == "" || hash == "" {
		return false, fmt.Errorf("password and hash cannot be empty")
	}

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return false, nil
		}
		return false, fmt.Errorf("failed to verify password: %w", err)
	}

	return true, nil
}

// NeedsRehash checks if a bcrypt hash needs to be regenerated
func (b *BcryptHasher) NeedsRehash(hash string) bool {
	cost, err := bcrypt.Cost([]byte(hash))
	if err != nil {
		return true // If we can't determine cost, rehash to be safe
	}
	return cost < b.cost
}

// Argon2Params holds parameters for Argon2id hashing
type Argon2Params struct {
	Memory      uint32 // Memory in KiB (default: 64MB = 65536)
	Iterations  uint32 // Number of iterations (default: 3)
	Parallelism uint8  // Degree of parallelism (default: 4)
	SaltLength  uint32 // Salt length in bytes (default: 16)
	KeyLength   uint32 // Key length in bytes (default: 32)
}

// DefaultArgon2Params returns recommended Argon2id parameters
func DefaultArgon2Params() Argon2Params {
	return Argon2Params{
		Memory:      64 * 1024, // 64 MB
		Iterations:  3,
		Parallelism: 4,
		SaltLength:  16,
		KeyLength:   32,
	}
}

// Argon2Hasher implements password hashing using Argon2id
type Argon2Hasher struct {
	params Argon2Params
}

// NewArgon2Hasher creates a new Argon2id password hasher
func NewArgon2Hasher(params Argon2Params) *Argon2Hasher {
	// Validate and set defaults
	if params.Memory == 0 {
		params.Memory = 64 * 1024
	}
	if params.Iterations == 0 {
		params.Iterations = 3
	}
	if params.Parallelism == 0 {
		params.Parallelism = 4
	}
	if params.SaltLength == 0 {
		params.SaltLength = 16
	}
	if params.KeyLength == 0 {
		params.KeyLength = 32
	}

	return &Argon2Hasher{params: params}
}

// Hash generates an Argon2id hash from a password
func (a *Argon2Hasher) Hash(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	// Generate random salt
	salt := make([]byte, a.params.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Generate hash
	hash := argon2.IDKey(
		[]byte(password),
		salt,
		a.params.Iterations,
		a.params.Memory,
		a.params.Parallelism,
		a.params.KeyLength,
	)

	// Encode hash in format: $argon2id$v=19$m=65536,t=3,p=4$salt$hash
	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		a.params.Memory,
		a.params.Iterations,
		a.params.Parallelism,
		encodedSalt,
		encodedHash,
	), nil
}

// Verify checks if a password matches an Argon2id hash
func (a *Argon2Hasher) Verify(password, encodedHash string) (bool, error) {
	if password == "" || encodedHash == "" {
		return false, fmt.Errorf("password and hash cannot be empty")
	}

	// Parse the encoded hash
	params, salt, hash, err := a.decodeHash(encodedHash)
	if err != nil {
		return false, fmt.Errorf("failed to decode hash: %w", err)
	}

	// Generate hash with same parameters
	computedHash := argon2.IDKey(
		[]byte(password),
		salt,
		params.Iterations,
		params.Memory,
		params.Parallelism,
		params.KeyLength,
	)

	// Use constant-time comparison to prevent timing attacks
	if subtle.ConstantTimeCompare(hash, computedHash) == 1 {
		return true, nil
	}

	return false, nil
}

// NeedsRehash checks if an Argon2id hash needs to be regenerated
func (a *Argon2Hasher) NeedsRehash(encodedHash string) bool {
	params, _, _, err := a.decodeHash(encodedHash)
	if err != nil {
		return true // If we can't parse, rehash to be safe
	}

	// Check if parameters have changed
	return params.Memory != a.params.Memory ||
		params.Iterations != a.params.Iterations ||
		params.Parallelism != a.params.Parallelism ||
		params.KeyLength != a.params.KeyLength
}

// decodeHash parses an Argon2id encoded hash
func (a *Argon2Hasher) decodeHash(encodedHash string) (Argon2Params, []byte, []byte, error) {
	// Format: $argon2id$v=19$m=65536,t=3,p=4$salt$hash
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return Argon2Params{}, nil, nil, fmt.Errorf("invalid hash format")
	}

	if parts[1] != "argon2id" {
		return Argon2Params{}, nil, nil, fmt.Errorf("not an argon2id hash")
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return Argon2Params{}, nil, nil, fmt.Errorf("invalid version: %w", err)
	}

	var params Argon2Params
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &params.Memory, &params.Iterations, &params.Parallelism); err != nil {
		return Argon2Params{}, nil, nil, fmt.Errorf("invalid parameters: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return Argon2Params{}, nil, nil, fmt.Errorf("invalid salt: %w", err)
	}
	params.SaltLength = uint32(len(salt))

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return Argon2Params{}, nil, nil, fmt.Errorf("invalid hash: %w", err)
	}
	params.KeyLength = uint32(len(hash))

	return params, salt, hash, nil
}

// NewPasswordHasher creates a password hasher with the specified algorithm
func NewPasswordHasher(algorithm PasswordHashAlgorithm) PasswordHasher {
	switch algorithm {
	case AlgorithmBcrypt:
		return NewBcryptHasher(bcrypt.DefaultCost)
	case AlgorithmArgon2id:
		return NewArgon2Hasher(DefaultArgon2Params())
	default:
		// Default to bcrypt
		return NewBcryptHasher(bcrypt.DefaultCost)
	}
}

// HashPassword is a convenience function to hash a password using bcrypt
func HashPassword(password string) (string, error) {
	hasher := NewBcryptHasher(bcrypt.DefaultCost)
	return hasher.Hash(password)
}

// VerifyPassword is a convenience function to verify a password against a hash
// Automatically detects the algorithm (bcrypt or argon2id)
func VerifyPassword(password, hash string) (bool, error) {
	if strings.HasPrefix(hash, "$argon2id$") {
		hasher := NewArgon2Hasher(DefaultArgon2Params())
		return hasher.Verify(password, hash)
	} else if strings.HasPrefix(hash, "$2") { // bcrypt hashes start with $2a$, $2b$, $2y$
		hasher := NewBcryptHasher(bcrypt.DefaultCost)
		return hasher.Verify(password, hash)
	}

	return false, fmt.Errorf("unknown hash format")
}

// ValidatePasswordStrength validates password strength
func ValidatePasswordStrength(password string, minLength int) error {
	if len(password) < minLength {
		return fmt.Errorf("password must be at least %d characters", minLength)
	}

	// Check for at least one uppercase letter
	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;:,.<>?", char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !hasDigit {
		return fmt.Errorf("password must contain at least one digit")
	}
	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}
