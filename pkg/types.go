package pkg

import (
	"net/http"
	"time"
)

// VirtualFS represents a virtual file system interface
// This interface is compatible with http.FileSystem
type VirtualFS interface {
	Open(name string) (http.File, error)
	Exists(name string) bool
}

// Minimal interface definitions for Context compatibility
// Full implementations are in their respective files

// Minimal interface definitions to avoid circular dependencies
// These match the actual implementations in their respective files

type SessionManager interface {
	Create(ctx Context) (*Session, error)
	Load(ctx Context, sessionID string) (*Session, error)
	Save(ctx Context, session *Session) error
	Destroy(ctx Context, sessionID string) error
	Set(sessionID, key string, value interface{}) error
	Get(sessionID, key string) (interface{}, error)
	Delete(sessionID, key string) error
	Clear(sessionID string) error
	SetCookie(ctx Context, session *Session) error
	GetSessionFromCookie(ctx Context) (*Session, error)
	IsValid(sessionID string) bool
	IsExpired(sessionID string) bool
	Refresh(ctx Context, sessionID string) error
	CleanupExpired() error
}

type CacheManager interface {
	Get(key string) (interface{}, error)
	Set(key string, value interface{}, ttl time.Duration) error
	Delete(key string) error
	Exists(key string) bool
	TTL(key string) (time.Duration, error)
	SetMultiple(items map[string]interface{}, ttl time.Duration) error
	GetMultiple(keys []string) (map[string]interface{}, error)
	DeleteMultiple(keys []string) error
	Increment(key string, delta int64) (int64, error)
	Decrement(key string, delta int64) (int64, error)
	Expire(key string, ttl time.Duration) error
	Clear() error
	Invalidate(pattern string) error
	GetRequestCache(requestID string) RequestCache
	ClearRequestCache(requestID string) error
}

type ConfigManager interface {
	Load(configPath string) error
	LoadFromEnv() error
	Reload() error
	GetString(key string) string
	GetInt(key string) int
	GetInt64(key string) int64
	GetFloat64(key string) float64
	GetBool(key string) bool
	GetDuration(key string) time.Duration
	GetStringSlice(key string) []string
	GetWithDefault(key string, defaultValue interface{}) interface{}
	GetStringWithDefault(key, defaultValue string) string
	GetIntWithDefault(key string, defaultValue int) int
	GetBoolWithDefault(key string, defaultValue bool) bool
	GetEnv() string
	Sub(key string) ConfigManager
	IsSet(key string) bool
	IsProduction() bool
	IsDevelopment() bool
	IsTest() bool
	Validate() error
	Watch(callback func()) error
	StopWatching() error
}

type I18nManager interface {
	Translate(key string, params ...interface{}) string
	TranslatePlural(key string, count int, params ...interface{}) string
	SetLanguage(lang string) error
	GetLanguage() string
	LoadLocale(locale string, data map[string]interface{}) error
	GetSupportedLanguages() []string
	LoadLocaleFromFile(locale, filepath string) error
}

type FileManager interface {
	Read(path string) ([]byte, error)
	Write(path string, data []byte) error
	Delete(path string) error
	Exists(path string) bool
	CreateDir(path string) error
	SaveUploadedFile(ctx Context, filename string, destPath string) error
}

type Logger interface {
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	WithRequestID(requestID string) Logger
}

type MetricsCollector interface {
	// Request metrics
	Start(requestID string) *RequestMetrics
	Record(metrics *RequestMetrics) error
	RecordRequest(ctx Context, duration time.Duration, statusCode int) error
	RecordError(ctx Context, err error) error

	// Workload metrics
	GetMetrics(tenantID string, from, to time.Time) ([]*WorkloadMetrics, error)
	GetAggregatedMetrics(tenantID string, from, to time.Time) (*AggregatedMetrics, error)
	PredictLoad(tenantID string, duration time.Duration) (*LoadPrediction, error)
	RecordWorkloadMetrics(metrics *WorkloadMetrics) error
	GetWorkloadMetrics(tenantID string, from, to time.Time) ([]*WorkloadMetrics, error)

	// Counter and gauge metrics
	IncrementCounter(name string, tags map[string]string) error
	IncrementCounterBy(name string, value int64, tags map[string]string) error
	SetGauge(name string, value float64, tags map[string]string) error
	IncrementGauge(name string, value float64, tags map[string]string) error
	DecrementGauge(name string, value float64, tags map[string]string) error

	// Histogram and timing metrics
	RecordHistogram(name string, value float64, tags map[string]string) error
	RecordTiming(name string, duration time.Duration, tags map[string]string) error
	StartTimer(name string, tags map[string]string) Timer

	// System metrics
	RecordMemoryUsage(usage int64) error
	RecordCPUUsage(usage float64) error
	RecordCustomMetric(name string, value interface{}, tags map[string]string) error

	// Export
	Export() (map[string]interface{}, error)
	ExportPrometheus() ([]byte, error)
}

// RequestCache provides request-specific caching
type RequestCache interface {
	Get(key string) interface{}
	Set(key string, value interface{})
	Delete(key string)
	Clear()
	Size() int64
	Keys() []string
}

// LogFormatter formats log messages
type LogFormatter interface {
	Format(level string, message string, args ...interface{}) string
}

// Timer provides timing functionality for metrics
type Timer interface {
	Stop() time.Duration
	Elapsed() time.Duration
}
