package pkg

import (
	"regexp"
	"testing"
	"time"
)

func TestRegexValidator_MatchString(t *testing.T) {
	validator := NewRegexValidator(100 * time.Millisecond)

	tests := []struct {
		name          string
		pattern       string
		input         string
		expectMatch   bool
		expectTimeout bool
	}{
		{
			name:        "simple match",
			pattern:     "^[a-z]+$",
			input:       "hello",
			expectMatch: true,
		},
		{
			name:        "simple no match",
			pattern:     "^[a-z]+$",
			input:       "Hello123",
			expectMatch: false,
		},
		{
			name:        "email pattern",
			pattern:     `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
			input:       "test@example.com",
			expectMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, err := validator.MatchString(tt.pattern, tt.input)

			if tt.expectTimeout {
				if err == nil {
					t.Error("Expected timeout error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				if matched != tt.expectMatch {
					t.Errorf("Expected match=%v, got match=%v", tt.expectMatch, matched)
				}
			}
		})
	}
}

func TestRegexValidator_Match(t *testing.T) {
	validator := NewRegexValidator(100 * time.Millisecond)

	pattern := regexp.MustCompile("^[a-z]+$")

	tests := []struct {
		name        string
		input       string
		expectMatch bool
	}{
		{
			name:        "match",
			input:       "hello",
			expectMatch: true,
		},
		{
			name:        "no match",
			input:       "Hello123",
			expectMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, err := validator.Match(pattern, tt.input)

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if matched != tt.expectMatch {
				t.Errorf("Expected match=%v, got match=%v", tt.expectMatch, matched)
			}
		})
	}
}

func TestValidatePattern(t *testing.T) {
	tests := []struct {
		name        string
		pattern     string
		expectError bool
	}{
		{
			name:        "safe pattern",
			pattern:     "^[a-z]+$",
			expectError: false,
		},
		{
			name:        "email pattern",
			pattern:     `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
			expectError: false,
		},
		{
			name:        "invalid pattern",
			pattern:     "[",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePattern(tt.pattern)

			if tt.expectError && err == nil {
				t.Error("Expected error, but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, but got: %v", err)
			}
		})
	}
}

func TestSafeMatchString(t *testing.T) {
	matched, err := SafeMatchString("^[a-z]+$", "hello")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !matched {
		t.Error("Expected match, but got no match")
	}
}

func TestSafeMatch(t *testing.T) {
	pattern := regexp.MustCompile("^[a-z]+$")
	matched, err := SafeMatch(pattern, "hello")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !matched {
		t.Error("Expected match, but got no match")
	}
}
