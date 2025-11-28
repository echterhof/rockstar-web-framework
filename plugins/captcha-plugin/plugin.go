package captchaplugin

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

// ============================================================================
// CAPTCHA Plugin (Compile-Time)
// ============================================================================
//
// This plugin demonstrates:
// - CAPTCHA generation and validation
// - Service export for other plugins
// - Session-based CAPTCHA storage
// - Middleware for CAPTCHA protection
// - Compile-time plugin registration
//
// Requirements: 9.6, 14.1, 14.2, 14.3
// ============================================================================

// init registers the plugin at compile time
func init() {
	pkg.RegisterPlugin("captcha-plugin", func() pkg.Plugin {
		return &CaptchaPlugin{}
	})
}

// CaptchaPlugin implements CAPTCHA validation functionality
type CaptchaPlugin struct {
	ctx pkg.PluginContext
	
	// Configuration
	enabled           bool
	captchaLength     int
	captchaExpiry     time.Duration
	protectedPaths    []string
	protectedMethods  []string
	caseSensitive     bool
	allowedCharacters string
	maxAttempts       int
	
	// Storage for CAPTCHA challenges
	storage pkg.PluginStorage
}

// CaptchaService is exported for other plugins to use
type CaptchaService struct {
	plugin *CaptchaPlugin
}

// GenerateCaptcha creates a new CAPTCHA challenge
func (s *CaptchaService) GenerateCaptcha() (id string, challenge string, err error) {
	return s.plugin.generateCaptcha()
}

// ValidateCaptcha checks if the provided answer matches the challenge
func (s *CaptchaService) ValidateCaptcha(id string, answer string) (bool, error) {
	return s.plugin.validateCaptcha(id, answer)
}

// ============================================================================
// Plugin Interface Implementation
// ============================================================================

func (p *CaptchaPlugin) Name() string {
	return "captcha-plugin"
}

func (p *CaptchaPlugin) Version() string {
	return "1.0.0"
}

func (p *CaptchaPlugin) Description() string {
	return "CAPTCHA generation and validation plugin for bot protection"
}

func (p *CaptchaPlugin) Author() string {
	return "Rockstar Framework Team"
}

func (p *CaptchaPlugin) Dependencies() []pkg.PluginDependency {
	return []pkg.PluginDependency{}
}

func (p *CaptchaPlugin) Initialize(ctx pkg.PluginContext) error {
	fmt.Printf("[%s] Initializing CAPTCHA plugin...\n", p.Name())
	
	p.ctx = ctx
	p.storage = ctx.PluginStorage()
	
	// Parse configuration
	if err := p.parseConfig(ctx.PluginConfig()); err != nil {
		return fmt.Errorf("failed to parse configuration: %w", err)
	}
	
	// Register middleware
	if err := p.registerMiddleware(); err != nil {
		return fmt.Errorf("failed to register middleware: %w", err)
	}
	
	// Export CAPTCHA service
	if err := p.exportService(); err != nil {
		return fmt.Errorf("failed to export service: %w", err)
	}
	
	fmt.Printf("[%s] Initialization complete\n", p.Name())
	return nil
}

func (p *CaptchaPlugin) Start() error {
	fmt.Printf("[%s] Starting CAPTCHA plugin...\n", p.Name())
	
	// Start background cleanup task
	go p.cleanupExpiredCaptchas()
	
	fmt.Printf("[%s] Plugin started\n", p.Name())
	return nil
}

func (p *CaptchaPlugin) Stop() error {
	fmt.Printf("[%s] Stopping CAPTCHA plugin...\n", p.Name())
	fmt.Printf("[%s] Plugin stopped\n", p.Name())
	return nil
}

func (p *CaptchaPlugin) Cleanup() error {
	fmt.Printf("[%s] Cleaning up CAPTCHA plugin...\n", p.Name())
	
	// Clear all CAPTCHA data
	if p.storage != nil {
		p.storage.Clear()
	}
	
	p.ctx = nil
	p.storage = nil
	
	fmt.Printf("[%s] Cleanup complete\n", p.Name())
	return nil
}

func (p *CaptchaPlugin) ConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"enabled": map[string]interface{}{
			"type":        "bool",
			"default":     true,
			"description": "Enable CAPTCHA protection",
		},
		"captcha_length": map[string]interface{}{
			"type":        "int",
			"default":     6,
			"description": "Length of CAPTCHA challenge",
		},
		"captcha_expiry": map[string]interface{}{
			"type":        "duration",
			"default":     "5m",
			"description": "CAPTCHA expiration time",
		},
		"protected_paths": map[string]interface{}{
			"type":        "array",
			"items":       "string",
			"default":     []string{"/api/login", "/api/register", "/api/contact"},
			"description": "Paths that require CAPTCHA validation",
		},
		"protected_methods": map[string]interface{}{
			"type":        "array",
			"items":       "string",
			"default":     []string{"POST"},
			"description": "HTTP methods that require CAPTCHA validation",
		},
		"case_sensitive": map[string]interface{}{
			"type":        "bool",
			"default":     false,
			"description": "Whether CAPTCHA validation is case-sensitive",
		},
		"allowed_characters": map[string]interface{}{
			"type":        "string",
			"default":     "ABCDEFGHJKLMNPQRSTUVWXYZ23456789",
			"description": "Characters allowed in CAPTCHA (excluding ambiguous ones)",
		},
		"max_attempts": map[string]interface{}{
			"type":        "int",
			"default":     3,
			"description": "Maximum validation attempts per CAPTCHA",
		},
	}
}

func (p *CaptchaPlugin) OnConfigChange(config map[string]interface{}) error {
	fmt.Printf("[%s] Configuration updated\n", p.Name())
	return p.parseConfig(config)
}

// ============================================================================
// Configuration Parsing
// ============================================================================

func (p *CaptchaPlugin) parseConfig(config map[string]interface{}) error {
	// Enabled
	if val, ok := config["enabled"].(bool); ok {
		p.enabled = val
	} else {
		p.enabled = true
	}
	
	// CAPTCHA Length
	if val, ok := config["captcha_length"].(int); ok {
		p.captchaLength = val
	} else if val, ok := config["captcha_length"].(float64); ok {
		p.captchaLength = int(val)
	} else {
		p.captchaLength = 6
	}
	
	// CAPTCHA Expiry
	if duration, ok := config["captcha_expiry"].(string); ok {
		d, err := time.ParseDuration(duration)
		if err != nil {
			return fmt.Errorf("invalid captcha_expiry: %w", err)
		}
		p.captchaExpiry = d
	} else {
		p.captchaExpiry = 5 * time.Minute
	}
	
	// Protected Paths
	if paths, ok := config["protected_paths"].([]interface{}); ok {
		p.protectedPaths = make([]string, 0, len(paths))
		for _, path := range paths {
			if pathStr, ok := path.(string); ok {
				p.protectedPaths = append(p.protectedPaths, pathStr)
			}
		}
	} else {
		p.protectedPaths = []string{"/api/login", "/api/register", "/api/contact"}
	}
	
	// Protected Methods
	if methods, ok := config["protected_methods"].([]interface{}); ok {
		p.protectedMethods = make([]string, 0, len(methods))
		for _, method := range methods {
			if methodStr, ok := method.(string); ok {
				p.protectedMethods = append(p.protectedMethods, methodStr)
			}
		}
	} else {
		p.protectedMethods = []string{"POST"}
	}
	
	// Case Sensitive
	if val, ok := config["case_sensitive"].(bool); ok {
		p.caseSensitive = val
	} else {
		p.caseSensitive = false
	}
	
	// Allowed Characters
	if val, ok := config["allowed_characters"].(string); ok {
		p.allowedCharacters = val
	} else {
		p.allowedCharacters = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	}
	
	// Max Attempts
	if val, ok := config["max_attempts"].(int); ok {
		p.maxAttempts = val
	} else if val, ok := config["max_attempts"].(float64); ok {
		p.maxAttempts = int(val)
	} else {
		p.maxAttempts = 3
	}
	
	return nil
}

// ============================================================================
// Middleware Registration
// ============================================================================

func (p *CaptchaPlugin) registerMiddleware() error {
	// Register CAPTCHA validation middleware
	captchaMiddleware := func(ctx pkg.Context, next pkg.HandlerFunc) error {
		if !p.enabled {
			return next(ctx)
		}
		
		req := ctx.Request()
		if req == nil || req.URL == nil {
			return next(ctx)
		}
		
		// Check if path and method require CAPTCHA
		if !p.requiresCaptcha(req.URL.Path, req.Method) {
			return next(ctx)
		}
		
		// Get CAPTCHA ID and answer from request
		captchaID := ctx.GetHeader("X-Captcha-ID")
		captchaAnswer := ctx.GetHeader("X-Captcha-Answer")
		
		if captchaID == "" || captchaAnswer == "" {
			return ctx.JSON(400, map[string]interface{}{
				"error": "CAPTCHA validation required",
			})
		}
		
		// Validate CAPTCHA
		valid, err := p.validateCaptcha(captchaID, captchaAnswer)
		if err != nil {
			return ctx.JSON(500, map[string]interface{}{
				"error": "CAPTCHA validation failed",
			})
		}
		
		if !valid {
			return ctx.JSON(403, map[string]interface{}{
				"error": "Invalid CAPTCHA",
			})
		}
		
		// CAPTCHA valid, proceed
		return next(ctx)
	}
	
	// Register the middleware
	err := p.ctx.RegisterMiddleware("captcha", captchaMiddleware, 180, []string{})
	if err != nil {
		return fmt.Errorf("failed to register captcha middleware: %w", err)
	}
	
	fmt.Printf("[%s] Registered CAPTCHA middleware\n", p.Name())
	return nil
}

// ============================================================================
// Service Export
// ============================================================================

func (p *CaptchaPlugin) exportService() error {
	service := &CaptchaService{plugin: p}
	
	err := p.ctx.ExportService("CaptchaService", service)
	if err != nil {
		return fmt.Errorf("failed to export CaptchaService: %w", err)
	}
	
	fmt.Printf("[%s] Exported CaptchaService\n", p.Name())
	return nil
}

// ============================================================================
// CAPTCHA Logic
// ============================================================================

func (p *CaptchaPlugin) generateCaptcha() (string, string, error) {
	// Generate random CAPTCHA ID
	idBytes := make([]byte, 16)
	if _, err := rand.Read(idBytes); err != nil {
		return "", "", fmt.Errorf("failed to generate CAPTCHA ID: %w", err)
	}
	id := base64.URLEncoding.EncodeToString(idBytes)
	
	// Generate random CAPTCHA challenge
	challenge := make([]byte, p.captchaLength)
	for i := range challenge {
		idx := make([]byte, 1)
		rand.Read(idx)
		challenge[i] = p.allowedCharacters[int(idx[0])%len(p.allowedCharacters)]
	}
	challengeStr := string(challenge)
	
	// Store CAPTCHA with expiry
	captchaData := map[string]interface{}{
		"challenge": challengeStr,
		"created":   time.Now(),
		"expires":   time.Now().Add(p.captchaExpiry),
		"attempts":  0,
	}
	
	if err := p.storage.Set(id, captchaData); err != nil {
		return "", "", fmt.Errorf("failed to store CAPTCHA: %w", err)
	}
	
	fmt.Printf("[%s] Generated CAPTCHA: %s\n", p.Name(), id)
	return id, challengeStr, nil
}

func (p *CaptchaPlugin) validateCaptcha(id string, answer string) (bool, error) {
	// Retrieve CAPTCHA data
	data, err := p.storage.Get(id)
	if err != nil {
		return false, fmt.Errorf("CAPTCHA not found")
	}
	
	captchaData, ok := data.(map[string]interface{})
	if !ok {
		return false, fmt.Errorf("invalid CAPTCHA data")
	}
	
	// Check expiry
	if expires, ok := captchaData["expires"].(time.Time); ok {
		if time.Now().After(expires) {
			p.storage.Delete(id)
			return false, fmt.Errorf("CAPTCHA expired")
		}
	}
	
	// Check attempts
	attempts, _ := captchaData["attempts"].(int)
	if attempts >= p.maxAttempts {
		p.storage.Delete(id)
		return false, fmt.Errorf("maximum attempts exceeded")
	}
	
	// Increment attempts
	attempts++
	captchaData["attempts"] = attempts
	p.storage.Set(id, captchaData)
	
	// Validate answer
	challenge, ok := captchaData["challenge"].(string)
	if !ok {
		return false, fmt.Errorf("invalid CAPTCHA challenge")
	}
	
	// Compare (case-insensitive by default)
	if !p.caseSensitive {
		challenge = strings.ToUpper(challenge)
		answer = strings.ToUpper(answer)
	}
	
	if challenge == answer {
		// Valid - delete CAPTCHA
		p.storage.Delete(id)
		fmt.Printf("[%s] CAPTCHA validated: %s\n", p.Name(), id)
		return true, nil
	}
	
	return false, nil
}

func (p *CaptchaPlugin) requiresCaptcha(path string, method string) bool {
	// Check if method requires CAPTCHA
	methodMatch := false
	for _, m := range p.protectedMethods {
		if m == method {
			methodMatch = true
			break
		}
	}
	if !methodMatch {
		return false
	}
	
	// Check if path requires CAPTCHA
	for _, protectedPath := range p.protectedPaths {
		if strings.HasPrefix(path, protectedPath) {
			return true
		}
	}
	
	return false
}

func (p *CaptchaPlugin) cleanupExpiredCaptchas() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		if p.storage == nil {
			continue
		}
		
		keys, err := p.storage.List()
		if err != nil {
			continue
		}
		
		for _, key := range keys {
			data, err := p.storage.Get(key)
			if err != nil {
				continue
			}
			
			if captchaData, ok := data.(map[string]interface{}); ok {
				if expires, ok := captchaData["expires"].(time.Time); ok {
					if time.Now().After(expires) {
						p.storage.Delete(key)
						fmt.Printf("[%s] Deleted expired CAPTCHA: %s\n", p.Name(), key)
					}
				}
			}
		}
	}
}
