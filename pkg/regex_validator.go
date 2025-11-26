package pkg

import (
	"context"
	"errors"
	"regexp"
	"time"
)

// RegexValidator provides safe regex matching with timeout protection
type RegexValidator struct {
	timeout time.Duration
}

// NewRegexValidator creates a new regex validator with a default timeout
func NewRegexValidator(timeout time.Duration) *RegexValidator {
	if timeout <= 0 {
		timeout = 100 * time.Millisecond // Default 100ms timeout
	}
	return &RegexValidator{
		timeout: timeout,
	}
}

// DefaultRegexValidator returns a validator with 100ms timeout
func DefaultRegexValidator() *RegexValidator {
	return NewRegexValidator(100 * time.Millisecond)
}

// MatchString matches a pattern against input with timeout protection
// This prevents ReDoS (Regular Expression Denial of Service) attacks
func (rv *RegexValidator) MatchString(pattern, input string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), rv.timeout)
	defer cancel()

	resultChan := make(chan matchResult, 1)

	go func() {
		matched, err := regexp.MatchString(pattern, input)
		resultChan <- matchResult{matched: matched, err: err}
	}()

	select {
	case result := <-resultChan:
		return result.matched, result.err
	case <-ctx.Done():
		return false, &FrameworkError{
			Code:       ErrCodeRegexTimeout,
			Message:    "regex matching timeout exceeded",
			StatusCode: 400,
			I18nKey:    "error.validation.regex_timeout",
		}
	}
}

// Match matches a compiled regex against input with timeout protection
func (rv *RegexValidator) Match(re *regexp.Regexp, input string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), rv.timeout)
	defer cancel()

	resultChan := make(chan bool, 1)

	go func() {
		matched := re.MatchString(input)
		resultChan <- matched
	}()

	select {
	case matched := <-resultChan:
		return matched, nil
	case <-ctx.Done():
		return false, &FrameworkError{
			Code:       ErrCodeRegexTimeout,
			Message:    "regex matching timeout exceeded",
			StatusCode: 400,
			I18nKey:    "error.validation.regex_timeout",
		}
	}
}

// FindStringSubmatch finds submatches with timeout protection
func (rv *RegexValidator) FindStringSubmatch(re *regexp.Regexp, input string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), rv.timeout)
	defer cancel()

	resultChan := make(chan []string, 1)

	go func() {
		matches := re.FindStringSubmatch(input)
		resultChan <- matches
	}()

	select {
	case matches := <-resultChan:
		return matches, nil
	case <-ctx.Done():
		return nil, &FrameworkError{
			Code:       ErrCodeRegexTimeout,
			Message:    "regex matching timeout exceeded",
			StatusCode: 400,
			I18nKey:    "error.validation.regex_timeout",
		}
	}
}

// ValidatePattern checks if a regex pattern is safe (not too complex)
func ValidatePattern(pattern string) error {
	// Check pattern length
	if len(pattern) > 1000 {
		return errors.New("regex pattern too long")
	}

	// Check for potentially dangerous nested quantifier patterns
	// These patterns look for actual nested quantifiers like (.*)+, not just the presence of + or *
	dangerousPatterns := []string{
		`\([^)]*\*\)[*+]`, // Matches (anything*)[*+] - nested quantifiers with *
		`\([^)]*\+\)[*+]`, // Matches (anything+)[*+] - nested quantifiers with +
	}

	for _, dangerous := range dangerousPatterns {
		matched, _ := regexp.MatchString(dangerous, pattern)
		if matched {
			return errors.New("potentially dangerous regex pattern detected")
		}
	}

	// Try to compile to ensure it's valid
	_, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}

	return nil
}

// matchResult holds the result of a regex match operation
type matchResult struct {
	matched bool
	err     error
}

// SafeMatchString is a convenience function for safe regex matching with default timeout
func SafeMatchString(pattern, input string) (bool, error) {
	validator := DefaultRegexValidator()
	return validator.MatchString(pattern, input)
}

// SafeMatch is a convenience function for safe regex matching with compiled regex
func SafeMatch(re *regexp.Regexp, input string) (bool, error) {
	validator := DefaultRegexValidator()
	return validator.Match(re, input)
}
