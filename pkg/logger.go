package pkg

import (
	"io"
	"log/slog"
)

// standardLogger implements the Logger interface using slog
type standardLogger struct {
	logger    *slog.Logger
	requestID string
	tenantID  string
	userID    string
}

// NewLogger creates a new standard logger
func NewLogger(logger *slog.Logger) Logger {
	if logger == nil {
		logger = slog.Default()
	}

	return &standardLogger{
		logger: logger,
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
	// TODO: Implement level setting
	return nil
}

// GetLevel returns the current log level
func (l *standardLogger) GetLevel() string {
	return "info"
}

// SetOutput sets the output writer
func (l *standardLogger) SetOutput(output io.Writer) error {
	// TODO: Implement output setting
	return nil
}

// SetFormatter sets the log formatter
func (l *standardLogger) SetFormatter(formatter LogFormatter) error {
	// TODO: Implement formatter setting
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
