//go:build ignore

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
	fmt.Println("🔐 Rockstar Web Framework - Cryptographic Key Generator")
	fmt.Println("=" + "==========================================================")
	fmt.Println()

	// Generate Session Encryption Key (AES-256 requires 32 bytes)
	sessionKey, err := pkg.GenerateEncryptionKeyHex(32)
	if err != nil {
		log.Fatalf("Failed to generate session encryption key: %v", err)
	}

	// Generate JWT Secret (recommended 32 bytes minimum for HS256)
	jwtSecret, err := pkg.GenerateJWTSecret(32)
	if err != nil {
		log.Fatalf("Failed to generate JWT secret: %v", err)
	}

	// Generate Security Encryption Key (for SecurityConfig)
	securityKey, err := pkg.GenerateEncryptionKeyHex(32)
	if err != nil {
		log.Fatalf("Failed to generate security encryption key: %v", err)
	}

	fmt.Println("✅ Generated Cryptographically Secure Keys")
	fmt.Println()
	fmt.Println("IMPORTANT: Store these keys securely!")
	fmt.Println("- Never commit them to version control")
	fmt.Println("- Use environment variables or a secrets manager")
	fmt.Println("- Rotate keys periodically")
	fmt.Println()
	fmt.Println("=" + "==========================================================")
	fmt.Println()

	// Display keys
	fmt.Println("📋 Session Encryption Key (32 bytes for AES-256):")
	fmt.Printf("   %s\n", sessionKey)
	fmt.Println()

	fmt.Println("📋 JWT Secret (32 bytes for HS256):")
	fmt.Printf("   %s\n", jwtSecret)
	fmt.Println()

	fmt.Println("📋 Security Encryption Key (32 bytes for AES-256):")
	fmt.Printf("   %s\n", securityKey)
	fmt.Println()

	fmt.Println("=" + "==========================================================")
	fmt.Println()

	// Generate .env file example
	fmt.Println("💾 Environment Variables (.env file format):")
	fmt.Println()
	fmt.Printf("SESSION_ENCRYPTION_KEY=%s\n", sessionKey)
	fmt.Printf("JWT_SECRET=%s\n", jwtSecret)
	fmt.Printf("SECURITY_ENCRYPTION_KEY=%s\n", securityKey)
	fmt.Println()

	// Offer to save to file
	fmt.Println("=" + "==========================================================")
	fmt.Println()
	fmt.Print("💾 Save to .env.example file? (y/N): ")

	var response string
	fmt.Scanln(&response)

	if response == "y" || response == "Y" {
		envContent := fmt.Sprintf(`# Rockstar Web Framework - Environment Variables
# Generated: %s
# 
# SECURITY WARNING:
# - Never commit the actual .env file to version control
# - Add .env to your .gitignore
# - Use different keys for each environment (dev, staging, prod)
# - Rotate keys periodically

# Session Encryption Key (32 bytes for AES-256)
SESSION_ENCRYPTION_KEY=%s

# JWT Secret (32 bytes for HS256)
JWT_SECRET=%s

# Security Encryption Key (32 bytes for AES-256)
SECURITY_ENCRYPTION_KEY=%s

# Database Configuration (example)
# DB_DRIVER=postgres
# DB_HOST=localhost
# DB_PORT=5432
# DB_NAME=myapp
# DB_USER=myuser
# DB_PASSWORD=mypassword

# Server Configuration (example)
# SERVER_PORT=8080
# SERVER_HOST=localhost
# ENABLE_HTTPS=true
# TLS_CERT_FILE=/path/to/cert.pem
# TLS_KEY_FILE=/path/to/key.pem
`,
			pkg.TimeNow().Format("2006-01-02 15:04:05"),
			sessionKey,
			jwtSecret,
			securityKey,
		)

		if err := os.WriteFile(".env.example", []byte(envContent), 0644); err != nil {
			log.Fatalf("Failed to write .env.example: %v", err)
		}

		fmt.Println("✅ Saved to .env.example")
		fmt.Println()
		fmt.Println("Next steps:")
		fmt.Println("1. Copy .env.example to .env")
		fmt.Println("2. Add .env to your .gitignore")
		fmt.Println("3. Load environment variables in your application")
		fmt.Println()
	}

	fmt.Println("=" + "==========================================================")
	fmt.Println()
	fmt.Println("🔒 Usage in Go Code:")
	fmt.Println()
	fmt.Println(`  import (
      "encoding/hex"
      "os"
  )

  // Load encryption key from environment
  encryptionKeyHex := os.Getenv("SESSION_ENCRYPTION_KEY")
  encryptionKey, err := hex.DecodeString(encryptionKeyHex)
  if err != nil {
      log.Fatal("Invalid encryption key")
  }

  // Use in configuration
  config := pkg.FrameworkConfig{
      SessionConfig: pkg.SessionConfig{
          EncryptionKey: encryptionKey,
      },
      SecurityConfig: pkg.SecurityConfig{
          EncryptionKey: os.Getenv("SECURITY_ENCRYPTION_KEY"),
          JWTSecret:     os.Getenv("JWT_SECRET"),
      },
  }`)
	fmt.Println()
	fmt.Println("=" + "==========================================================")
}
