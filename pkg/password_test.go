package pkg

import (
	"strings"
	"testing"
)

func TestBcryptHasher_Hash(t *testing.T) {
	hasher := NewBcryptHasher(10) // Use cost 10 for faster tests

	password := "MySecurePassword123!"
	hash, err := hasher.Hash(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if hash == "" {
		t.Error("Hash is empty")
	}

	if !strings.HasPrefix(hash, "$2") {
		t.Errorf("Hash doesn't start with bcrypt prefix: %s", hash)
	}
}

func TestBcryptHasher_Verify(t *testing.T) {
	hasher := NewBcryptHasher(10)

	password := "MySecurePassword123!"
	hash, err := hasher.Hash(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Test correct password
	valid, err := hasher.Verify(password, hash)
	if err != nil {
		t.Fatalf("Failed to verify password: %v", err)
	}
	if !valid {
		t.Error("Valid password was rejected")
	}

	// Test incorrect password
	valid, err = hasher.Verify("WrongPassword", hash)
	if err != nil {
		t.Fatalf("Failed to verify password: %v", err)
	}
	if valid {
		t.Error("Invalid password was accepted")
	}
}

func TestBcryptHasher_NeedsRehash(t *testing.T) {
	hasher10 := NewBcryptHasher(10)
	hasher12 := NewBcryptHasher(12)

	password := "MySecurePassword123!"
	hash, err := hasher10.Hash(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Hash with cost 10 should need rehash when using cost 12
	if !hasher12.NeedsRehash(hash) {
		t.Error("Expected hash to need rehash")
	}

	// Hash with cost 10 should not need rehash when using cost 10
	if hasher10.NeedsRehash(hash) {
		t.Error("Expected hash to not need rehash")
	}
}

func TestArgon2Hasher_Hash(t *testing.T) {
	hasher := NewArgon2Hasher(DefaultArgon2Params())

	password := "MySecurePassword123!"
	hash, err := hasher.Hash(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if hash == "" {
		t.Error("Hash is empty")
	}

	if !strings.HasPrefix(hash, "$argon2id$") {
		t.Errorf("Hash doesn't start with argon2id prefix: %s", hash)
	}
}

func TestArgon2Hasher_Verify(t *testing.T) {
	hasher := NewArgon2Hasher(DefaultArgon2Params())

	password := "MySecurePassword123!"
	hash, err := hasher.Hash(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Test correct password
	valid, err := hasher.Verify(password, hash)
	if err != nil {
		t.Fatalf("Failed to verify password: %v", err)
	}
	if !valid {
		t.Error("Valid password was rejected")
	}

	// Test incorrect password
	valid, err = hasher.Verify("WrongPassword", hash)
	if err != nil {
		t.Fatalf("Failed to verify password: %v", err)
	}
	if valid {
		t.Error("Invalid password was accepted")
	}
}

func TestArgon2Hasher_NeedsRehash(t *testing.T) {
	params1 := DefaultArgon2Params()
	params2 := DefaultArgon2Params()
	params2.Memory = 128 * 1024 // Double the memory

	hasher1 := NewArgon2Hasher(params1)
	hasher2 := NewArgon2Hasher(params2)

	password := "MySecurePassword123!"
	hash, err := hasher1.Hash(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Hash with params1 should need rehash when using params2
	if !hasher2.NeedsRehash(hash) {
		t.Error("Expected hash to need rehash")
	}

	// Hash with params1 should not need rehash when using params1
	if hasher1.NeedsRehash(hash) {
		t.Error("Expected hash to not need rehash")
	}
}

func TestHashPassword(t *testing.T) {
	password := "MySecurePassword123!"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if hash == "" {
		t.Error("Hash is empty")
	}
}

func TestVerifyPassword(t *testing.T) {
	password := "MySecurePassword123!"

	// Test bcrypt
	bcryptHash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	valid, err := VerifyPassword(password, bcryptHash)
	if err != nil {
		t.Fatalf("Failed to verify bcrypt password: %v", err)
	}
	if !valid {
		t.Error("Valid bcrypt password was rejected")
	}

	// Test argon2id
	argon2Hasher := NewArgon2Hasher(DefaultArgon2Params())
	argon2Hash, err := argon2Hasher.Hash(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	valid, err = VerifyPassword(password, argon2Hash)
	if err != nil {
		t.Fatalf("Failed to verify argon2id password: %v", err)
	}
	if !valid {
		t.Error("Valid argon2id password was rejected")
	}

	// Test wrong password
	valid, err = VerifyPassword("WrongPassword", bcryptHash)
	if err != nil {
		t.Fatalf("Failed to verify password: %v", err)
	}
	if valid {
		t.Error("Invalid password was accepted")
	}
}

func TestValidatePasswordStrength(t *testing.T) {
	tests := []struct {
		password  string
		minLength int
		shouldErr bool
		errMsg    string
	}{
		{"Short1!", 8, true, "too short"},
		{"nouppercase1!", 8, true, "no uppercase"},
		{"NOLOWERCASE1!", 8, true, "no lowercase"},
		{"NoDigits!", 8, true, "no digit"},
		{"NoSpecial1", 8, true, "no special"},
		{"ValidPass1!", 8, false, ""},
		{"AnotherGood123#", 8, false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.password, func(t *testing.T) {
			err := ValidatePasswordStrength(tt.password, tt.minLength)
			if tt.shouldErr && err == nil {
				t.Errorf("Expected error for %s (%s), got nil", tt.password, tt.errMsg)
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Expected no error for %s, got: %v", tt.password, err)
			}
		})
	}
}

func TestPasswordHasher_EmptyPassword(t *testing.T) {
	bcryptHasher := NewBcryptHasher(10)
	argon2Hasher := NewArgon2Hasher(DefaultArgon2Params())

	// Test bcrypt with empty password
	_, err := bcryptHasher.Hash("")
	if err == nil {
		t.Error("Expected error for empty password with bcrypt")
	}

	// Test argon2id with empty password
	_, err = argon2Hasher.Hash("")
	if err == nil {
		t.Error("Expected error for empty password with argon2id")
	}
}

func TestPasswordHasher_LongPassword(t *testing.T) {
	bcryptHasher := NewBcryptHasher(10)

	// Bcrypt has a 72-byte limit
	longPassword := strings.Repeat("a", 73)
	_, err := bcryptHasher.Hash(longPassword)
	if err == nil {
		t.Error("Expected error for password exceeding 72 bytes with bcrypt")
	}

	// Argon2id should handle long passwords
	argon2Hasher := NewArgon2Hasher(DefaultArgon2Params())
	veryLongPassword := strings.Repeat("a", 1000)
	hash, err := argon2Hasher.Hash(veryLongPassword)
	if err != nil {
		t.Errorf("Argon2id should handle long passwords: %v", err)
	}

	valid, err := argon2Hasher.Verify(veryLongPassword, hash)
	if err != nil || !valid {
		t.Error("Failed to verify long password with argon2id")
	}
}

func BenchmarkBcryptHash(b *testing.B) {
	hasher := NewBcryptHasher(10)
	password := "BenchmarkPassword123!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = hasher.Hash(password)
	}
}

func BenchmarkArgon2Hash(b *testing.B) {
	hasher := NewArgon2Hasher(DefaultArgon2Params())
	password := "BenchmarkPassword123!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = hasher.Hash(password)
	}
}

func BenchmarkBcryptVerify(b *testing.B) {
	hasher := NewBcryptHasher(10)
	password := "BenchmarkPassword123!"
	hash, _ := hasher.Hash(password)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = hasher.Verify(password, hash)
	}
}

func BenchmarkArgon2Verify(b *testing.B) {
	hasher := NewArgon2Hasher(DefaultArgon2Params())
	password := "BenchmarkPassword123!"
	hash, _ := hasher.Hash(password)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = hasher.Verify(password, hash)
	}
}
