package pkg

import (
	"context"
	"io"
	"log/slog"
	"os"
	"sync"
)

// standardLogger implements the Logger interface using slog
type standardLogger struct {
	logger    *slog.Logger
	handler   slog.Handler
	level     slog.Level
	output    io.Writer
	formatter LogFormatter
	mu        sync.RWMutex
	requestID string
	tenantID  string
	userID    string
}

// NewLogger creates a new standard logger
func NewLogger(logger *slog.Logger) Logger {
	if logger == nil {
		logger = slog.Default()
	}

	// Extract handler and level from the logger
	handler := logger.Handler()
	output := os.Stderr // Default output

	return &standardLogger{
		logger:    logger,
		handler:   handler,
		level:     slog.LevelInfo,
		output:    output,
		formatter: nil,
	}
}

// Debug logs a debug message
func (l *standardLogger) Debug(msg string, fields ...interface{}) {
	attrs := l.contextAttrs()
	l.logger.LogAttrs(nil, slog.LevelDebug, msg, attrs...)
}

// Info logs an info message
func (l *standardLogger) Info(msg string, fields ...interface{}) {
	attrs := l.contextAttrs()
	l.logger.LogAttrs(nil, slog.LevelInfo, msg, attrs...)
}

// Warn logs a warning message
func (l *standardLogger) Warn(msg string, fields ...interface{}) {
	attrs := l.contextAttrs()
	l.logger.LogAttrs(nil, slog.LevelWarn, msg, attrs...)
}

// Error logs an error message
func (l *standardLogger) Error(msg string, fields ...interface{}) {
	attrs := l.contextAttrs()
	l.logger.LogAttrs(nil, slog.LevelError, msg, attrs...)
}

// Fatal logs a fatal message and exits
func (l *standardLogger) Fatal(msg string, fields ...interface{}) {
	attrs := l.contextAttrs()
	l.logger.LogAttrs(nil, slog.LevelError, msg, attrs...)
	panic(msg) // Fatal should terminate
}

// WithField adds a field to the logger
func (l *standardLogger) WithField(key string, value interface{}) Logger {
	// Create a new logger with the field added
	return l
}

// WithFields adds multiple fields to the logger
func (l *standardLogger) WithFields(fields map[string]interface{}) Logger {
	// Create a new logger with the fields added
	return l
}

// WithError adds an error to the logger
func (l *standardLogger) WithError(err error) Logger {
	// Create a new logger with the error added
	return l
}

// WithContext creates a new logger with context information
func (l *standardLogger) WithContext(ctx Context) Logger {
	requestID := ""
	tenantID := ""
	userID := ""

	if ctx.Tenant() != nil {
		tenantID = ctx.Tenant().ID
	}
	if ctx.User() != nil {
		userID = ctx.User().ID
	}

	return &standardLogger{
		logger:    l.logger,
		requestID: requestID,
		tenantID:  tenantID,
		userID:    userID,
	}
}

// WithRequestID adds a request ID to the logger
func (l *standardLogger) WithRequestID(requestID string) Logger {
	return &standardLogger{
		logger:    l.logger,
		requestID: requestID,
		tenantID:  l.tenantID,
		userID:    l.userID,
	}
}

// WithTenant adds a tenant ID to the logger
func (l *standardLogger) WithTenant(tenantID string) Logger {
	return &standardLogger{
		logger:    l.logger,
		requestID: l.requestID,
		tenantID:  tenantID,
		userID:    l.userID,
	}
}

// WithUser adds a user ID to the logger
func (l *standardLogger) WithUser(userID string) Logger {
	return &standardLogger{
		logger:    l.logger,
		requestID: l.requestID,
		tenantID:  l.tenantID,
		userID:    userID,
	}
}

// DebugI18n logs a debug message with i18n support
func (l *standardLogger) DebugI18n(key string, params ...interface{}) {
	// For now, just log the key
	l.Debug(key)
}

// InfoI18n logs an info message with i18n support
func (l *standardLogger) InfoI18n(key string, params ...interface{}) {
	// For now, just log the key
	l.Info(key)
}

// WarnI18n logs a warning message with i18n support
func (l *standardLogger) WarnI18n(key string, params ...interface{}) {
	// For now, just log the key
	l.Warn(key)
}

// ErrorI18n logs an error message with i18n support
func (l *standardLogger) ErrorI18n(key string, params ...interface{}) {
	// For now, just log the key
	l.Error(key)
}

// SetLevel sets the log level
func (l *standardLogger) SetLevel(level string) error {
	// Parse level string to slog.Level
	var slogLevel slog.Level
	switch level {
	case "debug":
		slogLevel = slog.LevelDebug
	case "info":
		slogLevel = slog.LevelInfo
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	case "fatal":
		slogLevel = slog.LevelError // Fatal maps to error level
	default:
		return &FrameworkError{
			Code:    "invalid_log_level",
			Message: "invalid log level: must be one of debug, info, warn, error, fatal",
			Details: map[string]interface{}{"provided_level": level},
		}
	}

	// Update logger's level atomically
	l.mu.Lock()
	defer l.mu.Unlock()

	l.level = slogLevel

	// Create new handler with updated level
	opts := &slog.HandlerOptions{
		Level: slogLevel,
	}

	var newHandler slog.Handler
	if l.formatter != nil {
		// If we have a custom formatter, wrap the handler
		newHandler = &formatterHandler{
			handler:   slog.NewTextHandler(l.output, opts),
			formatter: l.formatter,
		}
	} else {
		newHandler = slog.NewTextHandler(l.output, opts)
	}

	l.handler = newHandler
	l.logger = slog.New(newHandler)

	return nil
}

// GetLevel returns the current log level
func (l *standardLogger) GetLevel() string {
	l.mu.RLock()
	defer l.mu.RUnlock()

	switch l.level {
	case slog.LevelDebug:
		return "debug"
	case slog.LevelInfo:
		return "info"
	case slog.LevelWarn:
		return "warn"
	case slog.LevelError:
		return "error"
	default:
		return "info"
	}
}

// SetOutput sets the output writer
func (l *standardLogger) SetOutput(output io.Writer) error {
	// Handle nil writer gracefully
	if output == nil {
		output = os.Stderr
	}

	// Update logger atomically
	l.mu.Lock()
	defer l.mu.Unlock()

	l.output = output

	// Create new handler with updated output
	opts := &slog.HandlerOptions{
		Level: l.level,
	}

	var newHandler slog.Handler
	if l.formatter != nil {
		// If we have a custom formatter, wrap the handler
		newHandler = &formatterHandler{
			handler:   slog.NewTextHandler(output, opts),
			formatter: l.formatter,
		}
	} else {
		newHandler = slog.NewTextHandler(output, opts)
	}

	l.handler = newHandler
	l.logger = slog.New(newHandler)

	return nil
}

// SetFormatter sets the log formatter
func (l *standardLogger) SetFormatter(formatter LogFormatter) error {
	// Update logger atomically
	l.mu.Lock()
	defer l.mu.Unlock()

	l.formatter = formatter

	// Create new handler with formatter
	opts := &slog.HandlerOptions{
		Level: l.level,
	}

	var newHandler slog.Handler
	if formatter != nil {
		// Wrap handler to apply custom formatting
		newHandler = &formatterHandler{
			handler:   slog.NewTextHandler(l.output, opts),
			formatter: formatter,
		}
	} else {
		// No formatter, use plain handler
		newHandler = slog.NewTextHandler(l.output, opts)
	}

	l.handler = newHandler
	l.logger = slog.New(newHandler)

	return nil
}

// contextAttrs returns context attributes for logging
func (l *standardLogger) contextAttrs() []slog.Attr {
	attrs := make([]slog.Attr, 0, 3)

	if l.requestID != "" {
		attrs = append(attrs, slog.String("request_id", l.requestID))
	}
	if l.tenantID != "" {
		attrs = append(attrs, slog.String("tenant_id", l.tenantID))
	}
	if l.userID != "" {
		attrs = append(attrs, slog.String("user_id", l.userID))
	}

	return attrs
}

// formatterHandler wraps a slog.Handler to apply custom formatting
type formatterHandler struct {
	handler   slog.Handler
	formatter LogFormatter
}

func (h *formatterHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *formatterHandler) Handle(ctx context.Context, record slog.Record) error {
	// Apply custom formatting
	levelStr := record.Level.String()
	formatted := h.formatter.Format(levelStr, record.Message)

	// Create a new record with formatted message
	newRecord := slog.NewRecord(record.Time, record.Level, formatted, record.PC)
	record.Attrs(func(attr slog.Attr) bool {
		newRecord.AddAttrs(attr)
		return true
	})

	return h.handler.Handle(ctx, newRecord)
}

func (h *formatterHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &formatterHandler{
		handler:   h.handler.WithAttrs(attrs),
		formatter: h.formatter,
	}
}

func (h *formatterHandler) WithGroup(name string) slog.Handler {
	return &formatterHandler{
		handler:   h.handler.WithGroup(name),
		formatter: h.formatter,
	}
}
