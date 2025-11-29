# Utilities API Reference

## Overview

The Rockstar Web Framework provides a collection of utility functions and types for common operations including cryptography, password hashing, path validation, regex validation, and time utilities. These utilities are designed with security and safety in mind.

## Cryptography Utilities

### GenerateEncryptionKey()

Generates a cryptographically secure encryption key of the specified length.

**Signature:**
```go
func GenerateEncryptionKey(length int) ([]byte, error)
```

**Parameters:**
- `length` - Key length in bytes (use 32 for AES-256)

**Returns:**
- `[]byte` - Generated key
- `error` - Error if length is invalid or generation fails

**Example:**
```go
// Generate AES-256 key (32 bytes)
key, err := pkg.GenerateEncryptionKey(32)
if err != nil {
    log.Fatal(err)
}
```

### GenerateEncryptionKeyHex()

Generates a cryptographically secure encryption key and returns it as a hex-encoded string.

**Signature:**
```go
func GenerateEncryptionKeyHex(length int) (string, error)
```

**Parameters:**
- `length` - Key length in bytes

**Returns:**
- `string` - Hex-encoded key
- `error` - Error if generation fails

**Example:**
```go
// Generate hex-encoded key for configuration
keyHex, err := pkg.GenerateEncryptionKeyHex(32)
if err != nil {
    log.Fatal(err)
}
// keyHex: "a1b2c3d4e5f6..."
```

### GenerateJWTSecret()

Generates a cryptographically secure JWT secret. Recommended: at least 32 bytes for HS256.

**Signature:**
```go
func GenerateJWTSecret(length int) (string, error)
```

**Parameters:**
- `length` - Secret length in bytes (minimum 32 recommended)

**Returns:**
- `string` - Hex-encoded secret
- `error` - Error if generation fails

**Example:**
```go
// Generate JWT secret
secret, err := pkg.GenerateJWTSecret(32)
if err != nil {
    log.Fatal(err)
}
```

### MustGenerateEncryptionKey()

Generates a key or panics. Use only for testing or initialization code.

**Signature:**
```go
func MustGenerateEncryptionKey(length int) []byte
```

**Parameters:**
- `length` - Key length in bytes

**Returns:**
- `[]byte` - Generated key

**Panics:** If key generation fails

**Example:**
```go
// In test code
key := pkg.MustGenerateEncryptionKey(32)
```

### MustGenerateEncryptionKeyHex()

Generates a hex-encoded key or panics. Use only for testing or initialization code.

**Signature:**
```go
func MustGenerateEncryptionKeyHex(length int) string
```

**Parameters:**
- `length` - Key length in bytes

**Returns:**
- `string` - Hex-encoded key

**Panics:** If key generation fails

**Example:**
```go
// In test code
keyHex := pkg.MustGenerateEncryptionKeyHex(32)
```

## Password Hashing

### PasswordHasher Interface

Interface for password hashing operations.

```go
type PasswordHasher interface {
    Hash(password string) (string, error)
    Verify(password, hash string) (bool, error)
    NeedsRehash(hash string) bool
}
```

### PasswordHashAlgorithm

Represents the hashing algorithm to use.

**Constants:**
```go
const (
    AlgorithmBcrypt   PasswordHashAlgorithm = "bcrypt"    // Recommended for most use cases
    AlgorithmArgon2id PasswordHashAlgorithm = "argon2id"  // High-security applications
)
```

### NewPasswordHasher()

Creates a password hasher with the specified algorithm.

**Signature:**
```go
func NewPasswordHasher(algorithm PasswordHashAlgorithm) PasswordHasher
```

**Parameters:**
- `algorithm` - Hashing algorithm (AlgorithmBcrypt or AlgorithmArgon2id)

**Returns:**
- `PasswordHasher` - Password hasher instance

**Example:**
```go
// Create bcrypt hasher
hasher := pkg.NewPasswordHasher(pkg.AlgorithmBcrypt)

// Hash password
hash, err := hasher.Hash("mypassword123")
if err != nil {
    log.Fatal(err)
}

// Verify password
valid, err := hasher.Verify("mypassword123", hash)
if err != nil {
    log.Fatal(err)
}
```

### HashPassword()

Convenience function to hash a password using bcrypt with default cost.

**Signature:**
```go
func HashPassword(password string) (string, error)
```

**Parameters:**
- `password` - Plain text password

**Returns:**
- `string` - Hashed password
- `error` - Error if hashing fails

**Example:**
```go
hash, err := pkg.HashPassword("mypassword123")
if err != nil {
    return err
}
// Store hash in database
```

### VerifyPassword()

Convenience function to verify a password against a hash. Automatically detects the algorithm (bcrypt or argon2id).

**Signature:**
```go
func VerifyPassword(password, hash string) (bool, error)
```

**Parameters:**
- `password` - Plain text password
- `hash` - Hashed password

**Returns:**
- `bool` - true if password matches, false otherwise
- `error` - Error if verification fails

**Example:**
```go
// Retrieve hash from database
hash := user.PasswordHash

// Verify password
valid, err := pkg.VerifyPassword("mypassword123", hash)
if err != nil {
    return err
}
if !valid {
    return errors.New("invalid password")
}
```

### ValidatePasswordStrength()

Validates password strength requirements.

**Signature:**
```go
func ValidatePasswordStrength(password string, minLength int) error
```

**Parameters:**
- `password` - Password to validate
- `minLength` - Minimum required length

**Returns:**
- `error` - Error describing validation failure, or nil if valid

**Requirements:**
- Minimum length
- At least one uppercase letter
- At least one lowercase letter
- At least one digit
- At least one special character

**Example:**
```go
err := pkg.ValidatePasswordStrength("MyPass123!", 8)
if err != nil {
    return ctx.JSON(400, map[string]string{
        "error": err.Error(),
    })
}
```

### Bcrypt Hasher

#### NewBcryptHasher()

Creates a new bcrypt password hasher with specified cost.

**Signature:**
```go
func NewBcryptHasher(cost int) *BcryptHasher
```

**Parameters:**
- `cost` - Cost factor (4-31, default 12)

**Returns:**
- `*BcryptHasher` - Bcrypt hasher instance

**Example:**
```go
// Create hasher with higher cost for extra security
hasher := pkg.NewBcryptHasher(14)
hash, _ := hasher.Hash("password")
```

#### BcryptHasher.Hash()

Generates a bcrypt hash from a password.

**Signature:**
```go
func (b *BcryptHasher) Hash(password string) (string, error)
```

**Limitations:**
- Password cannot be empty
- Maximum password length: 72 bytes

#### BcryptHasher.Verify()

Checks if a password matches a bcrypt hash.

**Signature:**
```go
func (b *BcryptHasher) Verify(password, hash string) (bool, error)
```

#### BcryptHasher.NeedsRehash()

Checks if a bcrypt hash needs to be regenerated due to updated cost parameters.

**Signature:**
```go
func (b *BcryptHasher) NeedsRehash(hash string) bool
```

### Argon2 Hasher

#### Argon2Params

Parameters for Argon2id hashing.

```go
type Argon2Params struct {
    Memory      uint32 // Memory in KiB (default: 64MB = 65536)
    Iterations  uint32 // Number of iterations (default: 3)
    Parallelism uint8  // Degree of parallelism (default: 4)
    SaltLength  uint32 // Salt length in bytes (default: 16)
    KeyLength   uint32 // Key length in bytes (default: 32)
}
```

#### DefaultArgon2Params()

Returns recommended Argon2id parameters.

**Signature:**
```go
func DefaultArgon2Params() Argon2Params
```

**Returns:**
- `Argon2Params` - Default parameters (64MB memory, 3 iterations, 4 parallelism)

#### NewArgon2Hasher()

Creates a new Argon2id password hasher.

**Signature:**
```go
func NewArgon2Hasher(params Argon2Params) *Argon2Hasher
```

**Parameters:**
- `params` - Argon2 parameters

**Returns:**
- `*Argon2Hasher` - Argon2 hasher instance

**Example:**
```go
// Use default parameters
hasher := pkg.NewArgon2Hasher(pkg.DefaultArgon2Params())

// Or customize
params := pkg.Argon2Params{
    Memory:      128 * 1024, // 128 MB
    Iterations:  4,
    Parallelism: 8,
    SaltLength:  16,
    KeyLength:   32,
}
hasher := pkg.NewArgon2Hasher(params)
```

#### Argon2Hasher.Hash()

Generates an Argon2id hash from a password.

**Signature:**
```go
func (a *Argon2Hasher) Hash(password string) (string, error)
```

**Hash Format:**
```
$argon2id$v=19$m=65536,t=3,p=4$salt$hash
```

#### Argon2Hasher.Verify()

Checks if a password matches an Argon2id hash using constant-time comparison.

**Signature:**
```go
func (a *Argon2Hasher) Verify(password, encodedHash string) (bool, error)
```

#### Argon2Hasher.NeedsRehash()

Checks if an Argon2id hash needs to be regenerated due to updated parameters.

**Signature:**
```go
func (a *Argon2Hasher) NeedsRehash(encodedHash string) bool
```

## Path Validation

### PathValidator

Provides secure path validation to prevent directory traversal attacks.

```go
type PathValidator struct {
    // private fields
}
```

### NewPathValidator()

Creates a new path validator with a base directory.

**Signature:**
```go
func NewPathValidator(baseDir string) *PathValidator
```

**Parameters:**
- `baseDir` - Base directory for path validation

**Returns:**
- `*PathValidator` - Path validator instance

**Example:**
```go
validator := pkg.NewPathValidator("/var/www/uploads")
```

### PathValidator.ValidatePath()

Validates that a path is safe and within the base directory. Prevents directory traversal attacks.

**Signature:**
```go
func (pv *PathValidator) ValidatePath(requestedPath string) error
```

**Parameters:**
- `requestedPath` - Path to validate

**Returns:**
- `error` - Error if path is unsafe, nil if valid

**Example:**
```go
validator := pkg.NewPathValidator("/var/www/uploads")

// Safe path
err := validator.ValidatePath("images/photo.jpg")  // OK

// Unsafe paths
err := validator.ValidatePath("../../../etc/passwd")  // Error
err := validator.ValidatePath("/etc/passwd")          // Error
```

### PathValidator.ResolvePath()

Validates and resolves a path to its absolute form within the base directory.

**Signature:**
```go
func (pv *PathValidator) ResolvePath(requestedPath string) (string, error)
```

**Parameters:**
- `requestedPath` - Path to validate and resolve

**Returns:**
- `string` - Absolute path
- `error` - Error if path is unsafe

**Example:**
```go
validator := pkg.NewPathValidator("/var/www/uploads")
fullPath, err := validator.ResolvePath("images/photo.jpg")
// fullPath: "/var/www/uploads/images/photo.jpg"
```

### ValidateAndResolvePath()

Convenience function that validates and resolves a path in one call.

**Signature:**
```go
func ValidateAndResolvePath(baseDir, requestedPath string) (string, error)
```

**Parameters:**
- `baseDir` - Base directory
- `requestedPath` - Path to validate and resolve

**Returns:**
- `string` - Absolute path
- `error` - Error if path is unsafe

**Example:**
```go
fullPath, err := pkg.ValidateAndResolvePath("/var/www/uploads", "images/photo.jpg")
if err != nil {
    return ctx.JSON(400, map[string]string{"error": "Invalid path"})
}
```

### IsPathSafe()

Checks if a path is safe without returning an error.

**Signature:**
```go
func IsPathSafe(baseDir, requestedPath string) bool
```

**Parameters:**
- `baseDir` - Base directory
- `requestedPath` - Path to check

**Returns:**
- `bool` - true if path is safe, false otherwise

**Example:**
```go
if !pkg.IsPathSafe("/var/www/uploads", userPath) {
    return ctx.JSON(400, map[string]string{"error": "Invalid path"})
}
```

## Regex Validation

### RegexValidator

Provides safe regex matching with timeout protection to prevent ReDoS (Regular Expression Denial of Service) attacks.

```go
type RegexValidator struct {
    // private fields
}
```

### NewRegexValidator()

Creates a new regex validator with a specified timeout.

**Signature:**
```go
func NewRegexValidator(timeout time.Duration) *RegexValidator
```

**Parameters:**
- `timeout` - Maximum time allowed for regex matching (default: 100ms if <= 0)

**Returns:**
- `*RegexValidator` - Regex validator instance

**Example:**
```go
validator := pkg.NewRegexValidator(200 * time.Millisecond)
```

### DefaultRegexValidator()

Returns a validator with 100ms timeout.

**Signature:**
```go
func DefaultRegexValidator() *RegexValidator
```

**Returns:**
- `*RegexValidator` - Validator with default timeout

### RegexValidator.MatchString()

Matches a pattern against input with timeout protection.

**Signature:**
```go
func (rv *RegexValidator) MatchString(pattern, input string) (bool, error)
```

**Parameters:**
- `pattern` - Regular expression pattern
- `input` - String to match against

**Returns:**
- `bool` - true if matched, false otherwise
- `error` - Error if timeout or invalid pattern

**Example:**
```go
validator := pkg.DefaultRegexValidator()
matched, err := validator.MatchString(`^\d{3}-\d{4}$`, "123-4567")
if err != nil {
    log.Printf("Regex error: %v", err)
}
if matched {
    // Valid format
}
```

### RegexValidator.Match()

Matches a compiled regex against input with timeout protection.

**Signature:**
```go
func (rv *RegexValidator) Match(re *regexp.Regexp, input string) (bool, error)
```

**Parameters:**
- `re` - Compiled regular expression
- `input` - String to match against

**Returns:**
- `bool` - true if matched, false otherwise
- `error` - Error if timeout

**Example:**
```go
validator := pkg.DefaultRegexValidator()
re := regexp.MustCompile(`^\d{3}-\d{4}$`)
matched, err := validator.Match(re, "123-4567")
```

### RegexValidator.FindStringSubmatch()

Finds submatches with timeout protection.

**Signature:**
```go
func (rv *RegexValidator) FindStringSubmatch(re *regexp.Regexp, input string) ([]string, error)
```

**Parameters:**
- `re` - Compiled regular expression with capture groups
- `input` - String to match against

**Returns:**
- `[]string` - Array of matches (full match + capture groups)
- `error` - Error if timeout

**Example:**
```go
validator := pkg.DefaultRegexValidator()
re := regexp.MustCompile(`^(\d{3})-(\d{4})$`)
matches, err := validator.FindStringSubmatch(re, "123-4567")
// matches: ["123-4567", "123", "4567"]
```

### ValidatePattern()

Checks if a regex pattern is safe (not too complex).

**Signature:**
```go
func ValidatePattern(pattern string) error
```

**Parameters:**
- `pattern` - Regular expression pattern to validate

**Returns:**
- `error` - Error if pattern is unsafe, nil if valid

**Checks:**
- Pattern length (max 1000 characters)
- Nested quantifiers (e.g., `(.*)+`)
- Pattern compilation validity

**Example:**
```go
err := pkg.ValidatePattern(userPattern)
if err != nil {
    return ctx.JSON(400, map[string]string{
        "error": "Invalid regex pattern",
    })
}
```

### SafeMatchString()

Convenience function for safe regex matching with default timeout.

**Signature:**
```go
func SafeMatchString(pattern, input string) (bool, error)
```

**Parameters:**
- `pattern` - Regular expression pattern
- `input` - String to match against

**Returns:**
- `bool` - true if matched, false otherwise
- `error` - Error if timeout or invalid pattern

**Example:**
```go
matched, err := pkg.SafeMatchString(`^\w+@\w+\.\w+$`, email)
if err != nil {
    log.Printf("Regex error: %v", err)
}
```

### SafeMatch()

Convenience function for safe regex matching with compiled regex.

**Signature:**
```go
func SafeMatch(re *regexp.Regexp, input string) (bool, error)
```

**Parameters:**
- `re` - Compiled regular expression
- `input` - String to match against

**Returns:**
- `bool` - true if matched, false otherwise
- `error` - Error if timeout

**Example:**
```go
re := regexp.MustCompile(`^\d+$`)
matched, err := pkg.SafeMatch(re, userInput)
```

## Time Utilities

### TimeNow()

Returns the current time. This function exists to allow mocking time in tests.

**Signature:**
```go
func TimeNow() time.Time
```

**Returns:**
- `time.Time` - Current time

**Example:**
```go
now := pkg.TimeNow()
timestamp := now.Unix()
```

**Testing:**
```go
// In tests, you can mock this function if needed
// by replacing it with a custom implementation
```

## Best Practices

### Cryptography

1. **Key Length:** Use 32 bytes for AES-256 encryption
2. **Storage:** Store keys securely (environment variables, secrets manager)
3. **Rotation:** Rotate encryption keys periodically
4. **Testing:** Use `Must*` functions only in test code

### Password Hashing

1. **Algorithm Choice:**
   - Use bcrypt for most applications (good balance of security and performance)
   - Use Argon2id for high-security applications (more resistant to GPU attacks)

2. **Cost/Parameters:**
   - Bcrypt: Use cost 12-14 (higher = more secure but slower)
   - Argon2id: Use default parameters or higher for sensitive applications

3. **Rehashing:** Check `NeedsRehash()` after successful login and rehash if needed

4. **Validation:** Always validate password strength before hashing

**Example:**
```go
// Registration
err := pkg.ValidatePasswordStrength(password, 8)
if err != nil {
    return ctx.JSON(400, map[string]string{"error": err.Error()})
}

hash, err := pkg.HashPassword(password)
if err != nil {
    return ctx.JSON(500, map[string]string{"error": "Failed to hash password"})
}
// Store hash in database

// Login
valid, err := pkg.VerifyPassword(password, storedHash)
if err != nil || !valid {
    return ctx.JSON(401, map[string]string{"error": "Invalid credentials"})
}

// Check if rehash needed
hasher := pkg.NewPasswordHasher(pkg.AlgorithmBcrypt)
if hasher.NeedsRehash(storedHash) {
    newHash, _ := hasher.Hash(password)
    // Update hash in database
}
```

### Path Validation

1. **Always Validate:** Never trust user-provided paths
2. **Base Directory:** Set a strict base directory for file operations
3. **Early Validation:** Validate paths before any file operations
4. **Error Handling:** Return generic errors to users (don't expose internal paths)

**Example:**
```go
router.GET("/files/:filename", func(ctx pkg.Context) error {
    filename := ctx.Param("filename")
    
    // Validate path
    fullPath, err := pkg.ValidateAndResolvePath("/var/www/uploads", filename)
    if err != nil {
        return ctx.JSON(400, map[string]string{"error": "Invalid file path"})
    }
    
    // Safe to use fullPath
    data, err := os.ReadFile(fullPath)
    // ...
})
```

### Regex Validation

1. **Timeout:** Always use timeout protection for user-provided patterns
2. **Pattern Validation:** Validate patterns before using them
3. **Complexity:** Keep patterns simple to avoid performance issues
4. **Testing:** Test patterns with various inputs including edge cases

**Example:**
```go
router.POST("/search", func(ctx pkg.Context) error {
    pattern := ctx.FormValue("pattern")
    
    // Validate pattern
    if err := pkg.ValidatePattern(pattern); err != nil {
        return ctx.JSON(400, map[string]string{"error": "Invalid pattern"})
    }
    
    // Safe matching with timeout
    validator := pkg.DefaultRegexValidator()
    matched, err := validator.MatchString(pattern, searchText)
    if err != nil {
        return ctx.JSON(400, map[string]string{"error": "Pattern matching failed"})
    }
    
    // Use matched result
})
```

## Security Considerations

1. **Cryptographic Keys:**
   - Never hardcode keys in source code
   - Use environment variables or secrets management
   - Rotate keys periodically

2. **Password Hashing:**
   - Never store plain text passwords
   - Use timing-safe comparison (built into Verify methods)
   - Implement rate limiting on login attempts

3. **Path Traversal:**
   - Always validate paths before file operations
   - Use absolute paths internally
   - Don't expose internal paths in error messages

4. **ReDoS Protection:**
   - Always use timeout protection for regex
   - Validate patterns before use
   - Limit pattern complexity

## See Also

- [Security Guide](../guides/security.md) - Security best practices
- [Security API](security.md) - Security manager reference
- [Configuration Guide](../guides/configuration.md) - Configuration management
- [Error Handling](errors.md) - Error types and handling
