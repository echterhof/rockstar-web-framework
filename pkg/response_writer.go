package pkg

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
)

// responseWriter implements ResponseWriter interface
type responseWriter struct {
	http.ResponseWriter

	status          int
	size            int64
	written         bool
	mu              sync.RWMutex
	wroteHeader     bool
	templateManager TemplateManager
}

// newResponseWriter creates a new response writer
func newResponseWriter(w http.ResponseWriter) ResponseWriter {
	return &responseWriter{
		ResponseWriter: w,
		status:         http.StatusOK,
	}
}

// newResponseWriterWithTemplates creates a new response writer with template support
func newResponseWriterWithTemplates(w http.ResponseWriter, tm TemplateManager) ResponseWriter {
	return &responseWriter{
		ResponseWriter:  w,
		status:          http.StatusOK,
		templateManager: tm,
	}
}

// Header returns the response headers
func (w *responseWriter) Header() http.Header {
	if w.ResponseWriter == nil {
		return http.Header{}
	}
	return w.ResponseWriter.Header()
}

// Write writes data to the response
func (w *responseWriter) Write(data []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.ResponseWriter == nil {
		return 0, fmt.Errorf("response writer is nil")
	}

	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	n, err := w.ResponseWriter.Write(data)
	w.size += int64(n)
	w.written = true
	return n, err
}

// WriteHeader writes the HTTP status code
func (w *responseWriter) WriteHeader(statusCode int) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.wroteHeader {
		return
	}

	if w.ResponseWriter == nil {
		return
	}

	w.status = statusCode
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(statusCode)
}

// WriteJSON writes JSON response
func (w *responseWriter) WriteJSON(statusCode int, data interface{}) error {
	w.SetContentType("application/json")
	w.WriteHeader(statusCode)

	encoder := json.NewEncoder(w)
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

// WriteXML writes XML response
func (w *responseWriter) WriteXML(statusCode int, data interface{}) error {
	w.SetContentType("application/xml")
	w.WriteHeader(statusCode)

	encoder := xml.NewEncoder(w)
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode XML: %w", err)
	}

	return nil
}

// WriteHTML writes HTML response
func (w *responseWriter) WriteHTML(statusCode int, templateName string, data interface{}) error {
	w.SetContentType("text/html; charset=utf-8")
	w.WriteHeader(statusCode)

	// If no template manager is set, return an error
	if w.templateManager == nil {
		return fmt.Errorf("template manager not configured")
	}

	// Render the template
	if err := w.templateManager.RenderTo(w, templateName, data); err != nil {
		return fmt.Errorf("failed to render template %s: %w", templateName, err)
	}

	return nil
}

// WriteString writes string response
func (w *responseWriter) WriteString(statusCode int, message string) error {
	w.SetContentType("text/plain; charset=utf-8")
	w.WriteHeader(statusCode)

	_, err := w.Write([]byte(message))
	return err
}

// WriteStream writes streaming response
func (w *responseWriter) WriteStream(statusCode int, contentType string, reader io.Reader) error {
	w.SetContentType(contentType)
	w.WriteHeader(statusCode)

	_, err := io.Copy(w, reader)
	if err != nil {
		return fmt.Errorf("failed to stream response: %w", err)
	}

	return nil
}

// SetCookie sets a cookie
func (w *responseWriter) SetCookie(cookie *Cookie) error {
	if cookie == nil {
		return errors.New("cookie cannot be nil")
	}

	httpCookie := &http.Cookie{
		Name:     cookie.Name,
		Value:    cookie.Value,
		Path:     cookie.Path,
		Domain:   cookie.Domain,
		Expires:  cookie.Expires,
		MaxAge:   cookie.MaxAge,
		Secure:   cookie.Secure,
		HttpOnly: cookie.HttpOnly,
		SameSite: cookie.SameSite,
	}

	http.SetCookie(w.ResponseWriter, httpCookie)
	return nil
}

// SetHeader sets a response header
func (w *responseWriter) SetHeader(key, value string) {
	w.Header().Set(key, value)
}

// SetContentType sets the Content-Type header
func (w *responseWriter) SetContentType(contentType string) {
	w.SetHeader("Content-Type", contentType)
}

// Status returns the HTTP status code
func (w *responseWriter) Status() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.status
}

// Size returns the response size in bytes
func (w *responseWriter) Size() int64 {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.size
}

// Written returns whether the response has been written
func (w *responseWriter) Written() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.written
}

// Flush flushes the response
func (w *responseWriter) Flush() error {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
		return nil
	}
	return errors.New("response writer does not support flushing")
}

// Close closes the response writer
func (w *responseWriter) Close() error {
	// HTTP response writers don't need explicit closing
	return nil
}

// Hijack implements http.Hijacker interface for WebSocket support
func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, errors.New("response writer does not support hijacking")
}

// Push implements http.Pusher interface for HTTP/2 server push
func (w *responseWriter) Push(target string, opts *http.PushOptions) error {
	if pusher, ok := w.ResponseWriter.(http.Pusher); ok {
		return pusher.Push(target, opts)
	}
	return errors.New("response writer does not support HTTP/2 push")
}

// SetTemplateManager sets the template manager for the response writer
func (w *responseWriter) SetTemplateManager(tm TemplateManager) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.templateManager = tm
}

// NewResponseWriter creates a new response writer (exported for testing)
func NewResponseWriter(w http.ResponseWriter) ResponseWriter {
	return newResponseWriter(w)
}

// NewResponseWriterWithTemplates creates a new response writer with template support (exported for testing)
func NewResponseWriterWithTemplates(w http.ResponseWriter, tm TemplateManager) ResponseWriter {
	return newResponseWriterWithTemplates(w, tm)
}
