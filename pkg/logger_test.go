package pkg

import (
	"bytes"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"testing/quick"
)

// **Feature: todo-implementations, Property 5: Log level filtering**
// **Validates: Requirements 2.1**
func TestProperty_LogLevelFiltering(t *testing.T) {
	// Property: For any valid log level, when set, only messages at that level or higher should be output

	f := func(levelIndex uint8) bool {
		// Map index to valid log levels
		levels := []string{"debug", "info", "warn", "error"}
		levelIdx := int(levelIndex) % len(levels)
		level := levels[levelIdx]

		// Create a buffer to capture output
		var buf bytes.Buffer

		// Create initial logger
		handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
		logger := slog.New(handler)

		stdLogger := NewLogger(logger).(*standardLogger)

		// Set output to our buffer first
		stdLogger.SetOutput(&buf)

		// Set the level
		err := stdLogger.SetLevel(level)
		if err != nil {
			t.Logf("Failed to set level %s: %v", level, err)
			return false
		}

		// Clear buffer after setup
		buf.Reset()

		// Log messages at all levels
		stdLogger.Debug("debug message")
		stdLogger.Info("info message")
		stdLogger.Warn("warn message")
		stdLogger.Error("error message")

		output := buf.String()

		// Verify filtering based on level
		switch level {
		case "debug":
			// All messages should appear
			return strings.Contains(output, "debug message") &&
				strings.Contains(output, "info message") &&
				strings.Contains(output, "warn message") &&
				strings.Contains(output, "error message")
		case "info":
			// Info and above should appear, debug should not
			return !strings.Contains(output, "debug message") &&
				strings.Contains(output, "info message") &&
				strings.Contains(output, "warn message") &&
				strings.Contains(output, "error message")
		case "warn":
			// Warn and above should appear, debug and info should not
			return !strings.Contains(output, "debug message") &&
				!strings.Contains(output, "info message") &&
				strings.Contains(output, "warn message") &&
				strings.Contains(output, "error message")
		case "error":
			// Only error should appear
			return !strings.Contains(output, "debug message") &&
				!strings.Contains(output, "info message") &&
				!strings.Contains(output, "warn message") &&
				strings.Contains(output, "error message")
		}

		return false
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(f, config); err != nil {
		t.Errorf("Property violated: %v", err)
	}
}

// **Feature: todo-implementations, Property 8: Invalid level rejection**
// **Validates: Requirements 2.4**
func TestProperty_InvalidLevelRejection(t *testing.T) {
	// Property: For any invalid log level string, SetLevel should return an error and the current level should remain unchanged

	f := func(invalidLevel string) bool {
		// Skip valid levels
		validLevels := map[string]bool{
			"debug": true,
			"info":  true,
			"warn":  true,
			"error": true,
			"fatal": true,
		}

		if validLevels[invalidLevel] {
			return true // Skip valid levels
		}

		// Create logger
		logger := NewLogger(nil).(*standardLogger)

		// Set to a known valid level first
		logger.SetLevel("info")
		originalLevel := logger.GetLevel()

		// Try to set invalid level
		err := logger.SetLevel(invalidLevel)

		// Should return an error
		if err == nil {
			t.Logf("Expected error for invalid level %q, got nil", invalidLevel)
			return false
		}

		// Level should remain unchanged
		currentLevel := logger.GetLevel()
		if currentLevel != originalLevel {
			t.Logf("Level changed from %q to %q after invalid level %q", originalLevel, currentLevel, invalidLevel)
			return false
		}

		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(f, config); err != nil {
		t.Errorf("Property violated: %v", err)
	}
}

// **Feature: todo-implementations, Property 6: Output redirection**
// **Validates: Requirements 2.2**
func TestProperty_OutputRedirection(t *testing.T) {
	// Property: For any io.Writer, when set as output, all subsequent log messages should be written to that writer

	f := func(messageIndex uint8) bool {
		// Generate a unique message using printable characters
		message := "test message " + string(rune('A'+int(messageIndex%26)))

		// Create two buffers
		var buf1, buf2 bytes.Buffer

		// Create logger with first buffer
		logger := NewLogger(nil).(*standardLogger)
		logger.SetLevel("info") // Ensure level is set
		logger.SetOutput(&buf1)

		// Log to first buffer
		logger.Info(message + " first")

		// Verify message is in first buffer
		if !strings.Contains(buf1.String(), message+" first") {
			t.Logf("Message not found in first buffer: %q", buf1.String())
			return false
		}

		// Switch to second buffer
		logger.SetOutput(&buf2)

		// Log to second buffer
		logger.Info(message + " second")

		// Verify message is in second buffer
		if !strings.Contains(buf2.String(), message+" second") {
			t.Logf("Message not found in second buffer: %q", buf2.String())
			return false
		}

		// Verify first buffer didn't get the second message
		if strings.Contains(buf1.String(), message+" second") {
			t.Logf("Second message incorrectly appeared in first buffer")
			return false
		}

		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(f, config); err != nil {
		t.Errorf("Property violated: %v", err)
	}
}

// testFormatter is a simple formatter for testing
type testFormatter struct {
	prefix string
}

func (f *testFormatter) Format(level string, message string, args ...interface{}) string {
	return f.prefix + " [" + level + "] " + message
}

// **Feature: todo-implementations, Property 7: Custom formatter application**
// **Validates: Requirements 2.3**
func TestProperty_CustomFormatterApplication(t *testing.T) {
	// Property: For any LogFormatter, when set, all subsequent log messages should be formatted using that formatter

	f := func(prefixIndex uint8) bool {
		// Generate a unique prefix
		prefix := "PREFIX" + string(rune('A'+int(prefixIndex%26)))

		// Create a buffer to capture output
		var buf bytes.Buffer

		// Create logger
		logger := NewLogger(nil).(*standardLogger)
		logger.SetLevel("info")
		logger.SetOutput(&buf)

		// Set custom formatter
		formatter := &testFormatter{prefix: prefix}
		err := logger.SetFormatter(formatter)
		if err != nil {
			t.Logf("Failed to set formatter: %v", err)
			return false
		}

		// Log a message
		message := "test message"
		logger.Info(message)

		output := buf.String()

		// Verify the formatter was applied (prefix should be in output)
		if !strings.Contains(output, prefix) {
			t.Logf("Prefix %q not found in output: %q", prefix, output)
			return false
		}

		// Verify the message is in output
		if !strings.Contains(output, message) {
			t.Logf("Message %q not found in output: %q", message, output)
			return false
		}

		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(f, config); err != nil {
		t.Errorf("Property violated: %v", err)
	}
}

// Unit tests for logger configuration

func TestLoggerConfiguration_ConcurrentChanges(t *testing.T) {
	// Test concurrent configuration changes
	logger := NewLogger(nil).(*standardLogger)

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent level changes
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			logger.SetLevel("debug")
			logger.SetLevel("info")
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			logger.SetLevel("warn")
			logger.SetLevel("error")
		}
	}()

	// Concurrent logging
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			logger.Info("test message")
		}
	}()

	wg.Wait()

	// Should not panic or deadlock
}

func TestLoggerConfiguration_FormatterWithVariousMessageTypes(t *testing.T) {
	// Test formatter with various message types
	var buf bytes.Buffer
	logger := NewLogger(nil).(*standardLogger)
	logger.SetLevel("info")
	logger.SetOutput(&buf)

	formatter := &testFormatter{prefix: "TEST"}
	logger.SetFormatter(formatter)

	// Test different message types
	testCases := []struct {
		name    string
		logFunc func(string, ...interface{})
		message string
	}{
		{"info", logger.Info, "info message"},
		{"warn", logger.Warn, "warning message"},
		{"error", logger.Error, "error message"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf.Reset()
			tc.logFunc(tc.message)

			output := buf.String()
			if !strings.Contains(output, "TEST") {
				t.Errorf("Expected formatter prefix in output, got: %s", output)
			}
			if !strings.Contains(output, tc.message) {
				t.Errorf("Expected message %q in output, got: %s", tc.message, output)
			}
		})
	}
}

func TestLoggerConfiguration_OutputToDifferentWriters(t *testing.T) {
	// Test output to different writers
	logger := NewLogger(nil).(*standardLogger)
	logger.SetLevel("info")

	// Test with buffer
	var buf bytes.Buffer
	logger.SetOutput(&buf)
	logger.Info("buffer message")

	if !strings.Contains(buf.String(), "buffer message") {
		t.Errorf("Expected message in buffer, got: %s", buf.String())
	}

	// Test with another buffer
	var buf2 bytes.Buffer
	logger.SetOutput(&buf2)
	logger.Info("buffer2 message")

	if !strings.Contains(buf2.String(), "buffer2 message") {
		t.Errorf("Expected message in second buffer, got: %s", buf2.String())
	}

	// First buffer should not have second message
	if strings.Contains(buf.String(), "buffer2 message") {
		t.Errorf("Second message should not be in first buffer")
	}
}

func TestLoggerConfiguration_NilFormatterHandling(t *testing.T) {
	// Test that nil formatter is handled gracefully
	var buf bytes.Buffer
	logger := NewLogger(nil).(*standardLogger)
	logger.SetLevel("info")
	logger.SetOutput(&buf)

	// Set a formatter
	formatter := &testFormatter{prefix: "TEST"}
	logger.SetFormatter(formatter)

	buf.Reset()
	logger.Info("with formatter")

	if !strings.Contains(buf.String(), "TEST") {
		t.Errorf("Expected formatter prefix")
	}

	// Set nil formatter
	logger.SetFormatter(nil)

	buf.Reset()
	logger.Info("without formatter")

	// Should still log, just without custom formatting
	if !strings.Contains(buf.String(), "without formatter") {
		t.Errorf("Expected message without formatter")
	}

	// Should not have the prefix anymore
	if strings.Contains(buf.String(), "TEST") {
		t.Errorf("Should not have formatter prefix after setting nil")
	}
}

func TestLoggerConfiguration_NilOutputHandling(t *testing.T) {
	// Test that nil output is handled gracefully (defaults to stderr)
	logger := NewLogger(nil).(*standardLogger)

	// Set nil output - should default to stderr
	err := logger.SetOutput(nil)
	if err != nil {
		t.Errorf("SetOutput(nil) should not return error, got: %v", err)
	}

	// Should not panic when logging
	logger.Info("test message")
}
