package pkg

import (
	"bytes"
	"net/http"
	"strings"
	"testing"
	"testing/fstest"
)

// TestTemplateManager_LoadTemplate tests loading a single template
func TestTemplateManager_LoadTemplate(t *testing.T) {
	tm := NewTemplateManager()

	// Load a simple template
	err := tm.LoadTemplate("test", "<h1>{{.Title}}</h1>")
	if err != nil {
		t.Fatalf("Failed to load template: %v", err)
	}

	// Check if template exists
	if !tm.HasTemplate("test") {
		t.Error("Template should exist after loading")
	}

	// Render the template
	result, err := tm.Render("test", map[string]string{"Title": "Hello World"})
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	expected := "<h1>Hello World</h1>"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

// TestTemplateManager_LoadTemplates tests loading templates from filesystem
func TestTemplateManager_LoadTemplates(t *testing.T) {
	// Create a mock filesystem
	mockFS := fstest.MapFS{
		"templates/index.html": &fstest.MapFile{
			Data: []byte("<html><body>{{.Content}}</body></html>"),
		},
		"templates/about.html": &fstest.MapFile{
			Data: []byte("<html><body><h1>{{.Title}}</h1></body></html>"),
		},
	}

	tm := NewTemplateManager()

	// Load templates from filesystem
	err := tm.LoadTemplates(mockFS, "templates/*.html")
	if err != nil {
		t.Fatalf("Failed to load templates: %v", err)
	}

	// Check if templates exist
	if !tm.HasTemplate("index.html") {
		t.Error("index.html template should exist")
	}

	if !tm.HasTemplate("about.html") {
		t.Error("about.html template should exist")
	}

	// Render index template
	result, err := tm.Render("index.html", map[string]string{"Content": "Welcome"})
	if err != nil {
		t.Fatalf("Failed to render index template: %v", err)
	}

	if !strings.Contains(result, "Welcome") {
		t.Errorf("Expected result to contain 'Welcome', got %q", result)
	}
}

// TestTemplateManager_RenderTo tests rendering to a writer
func TestTemplateManager_RenderTo(t *testing.T) {
	tm := NewTemplateManager()

	err := tm.LoadTemplate("test", "Hello {{.Name}}!")
	if err != nil {
		t.Fatalf("Failed to load template: %v", err)
	}

	var buf bytes.Buffer
	err = tm.RenderTo(&buf, "test", map[string]string{"Name": "Alice"})
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	expected := "Hello Alice!"
	if buf.String() != expected {
		t.Errorf("Expected %q, got %q", expected, buf.String())
	}
}

// TestTemplateManager_AddFunc tests adding custom functions
func TestTemplateManager_AddFunc(t *testing.T) {
	tm := NewTemplateManager()

	// Add a custom function
	err := tm.AddFunc("upper", strings.ToUpper)
	if err != nil {
		t.Fatalf("Failed to add function: %v", err)
	}

	// Load a template that uses the custom function
	err = tm.LoadTemplate("test", "{{upper .Text}}")
	if err != nil {
		t.Fatalf("Failed to load template: %v", err)
	}

	result, err := tm.Render("test", map[string]string{"Text": "hello"})
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	expected := "HELLO"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

// TestTemplateManager_AddFuncAfterLoad tests that adding functions after loading fails
func TestTemplateManager_AddFuncAfterLoad(t *testing.T) {
	tm := NewTemplateManager()

	// Load a template first
	err := tm.LoadTemplate("test", "{{.Text}}")
	if err != nil {
		t.Fatalf("Failed to load template: %v", err)
	}

	// Try to add a function after loading
	err = tm.AddFunc("upper", strings.ToUpper)
	if err == nil {
		t.Error("Expected error when adding function after templates are loaded")
	}
}

// TestTemplateManager_Clear tests clearing templates
func TestTemplateManager_Clear(t *testing.T) {
	tm := NewTemplateManager()

	err := tm.LoadTemplate("test", "{{.Text}}")
	if err != nil {
		t.Fatalf("Failed to load template: %v", err)
	}

	if !tm.HasTemplate("test") {
		t.Error("Template should exist before clear")
	}

	tm.Clear()

	if tm.HasTemplate("test") {
		t.Error("Template should not exist after clear")
	}
}

// TestTemplateManager_RenderNonExistent tests rendering a non-existent template
func TestTemplateManager_RenderNonExistent(t *testing.T) {
	tm := NewTemplateManager()

	_, err := tm.Render("nonexistent", nil)
	if err == nil {
		t.Error("Expected error when rendering non-existent template")
	}
}

// TestTemplateManager_RenderWithoutLoading tests rendering without loading templates
func TestTemplateManager_RenderWithoutLoading(t *testing.T) {
	tm := NewTemplateManager()

	_, err := tm.Render("test", nil)
	if err == nil {
		t.Error("Expected error when rendering without loading templates")
	}
}

// TestView tests the View helper function
func TestView(t *testing.T) {
	// Create a mock context
	tm := NewTemplateManager()
	err := tm.LoadTemplate("test", "<h1>{{.Title}}</h1>")
	if err != nil {
		t.Fatalf("Failed to load template: %v", err)
	}

	// Create a test HTTP response writer
	mockWriter := &mockHTTPResponseWriter{
		header: make(http.Header),
		body:   &bytes.Buffer{},
	}

	rw := NewResponseWriterWithTemplates(mockWriter, tm)

	// Create a basic context
	ctx := &contextImpl{
		response: rw,
	}

	// Create a view function
	viewFunc := View("test", func(c Context) interface{} {
		return map[string]string{"Title": "Test Page"}
	})

	// Execute the view
	responseFunc := viewFunc(ctx)
	err = responseFunc()
	if err != nil {
		t.Fatalf("Failed to execute view: %v", err)
	}

	// Check the response
	result := mockWriter.body.String()
	expected := "<h1>Test Page</h1>"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

// TestViewWithStatus tests the ViewWithStatus helper function
func TestViewWithStatus(t *testing.T) {
	tm := NewTemplateManager()
	err := tm.LoadTemplate("error", "<h1>Error: {{.Message}}</h1>")
	if err != nil {
		t.Fatalf("Failed to load template: %v", err)
	}

	mockWriter := &mockHTTPResponseWriter{
		header: make(http.Header),
		body:   &bytes.Buffer{},
	}

	rw := NewResponseWriterWithTemplates(mockWriter, tm)

	ctx := &contextImpl{
		response: rw,
	}

	// Create a view with custom status code
	viewFunc := ViewWithStatus(404, "error", func(c Context) interface{} {
		return map[string]string{"Message": "Not Found"}
	})

	responseFunc := viewFunc(ctx)
	err = responseFunc()
	if err != nil {
		t.Fatalf("Failed to execute view: %v", err)
	}

	// Check status code
	if mockWriter.statusCode != 404 {
		t.Errorf("Expected status code 404, got %d", mockWriter.statusCode)
	}

	// Check the response
	result := mockWriter.body.String()
	if !strings.Contains(result, "Not Found") {
		t.Errorf("Expected result to contain 'Not Found', got %q", result)
	}
}

// TestTemplateRenderer tests the TemplateRenderer wrapper
func TestTemplateRenderer(t *testing.T) {
	tm := NewTemplateManager()
	renderer := NewTemplateRenderer(tm)

	// Load a template
	err := renderer.LoadFromString("test", "Hello {{.Name}}!")
	if err != nil {
		t.Fatalf("Failed to load template: %v", err)
	}

	// Render the template
	var buf bytes.Buffer
	err = renderer.Render(&buf, "test", map[string]string{"Name": "Bob"})
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	expected := "Hello Bob!"
	if buf.String() != expected {
		t.Errorf("Expected %q, got %q", expected, buf.String())
	}
}

// TestTemplateRenderer_LoadFromFS tests loading templates from filesystem
func TestTemplateRenderer_LoadFromFS(t *testing.T) {
	mockFS := fstest.MapFS{
		"page.html": &fstest.MapFile{
			Data: []byte("<html>{{.Content}}</html>"),
		},
	}

	tm := NewTemplateManager()
	renderer := NewTemplateRenderer(tm)

	err := renderer.LoadFromFS(mockFS, "*.html")
	if err != nil {
		t.Fatalf("Failed to load templates: %v", err)
	}

	var buf bytes.Buffer
	err = renderer.Render(&buf, "page.html", map[string]string{"Content": "Test"})
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	if !strings.Contains(buf.String(), "Test") {
		t.Errorf("Expected result to contain 'Test', got %q", buf.String())
	}
}

// TestTemplateRenderer_AddFunc tests adding custom functions via renderer
func TestTemplateRenderer_AddFunc(t *testing.T) {
	tm := NewTemplateManager()
	renderer := NewTemplateRenderer(tm)

	// Add custom function
	err := renderer.AddFunc("lower", strings.ToLower)
	if err != nil {
		t.Fatalf("Failed to add function: %v", err)
	}

	// Load template
	err = renderer.LoadFromString("test", "{{lower .Text}}")
	if err != nil {
		t.Fatalf("Failed to load template: %v", err)
	}

	var buf bytes.Buffer
	err = renderer.Render(&buf, "test", map[string]string{"Text": "HELLO"})
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	expected := "hello"
	if buf.String() != expected {
		t.Errorf("Expected %q, got %q", expected, buf.String())
	}
}

// TestResponseWriter_WriteHTML tests HTML rendering with templates
func TestResponseWriter_WriteHTML(t *testing.T) {
	tm := NewTemplateManager()
	err := tm.LoadTemplate("page", "<html><body>{{.Content}}</body></html>")
	if err != nil {
		t.Fatalf("Failed to load template: %v", err)
	}

	mockWriter := &mockHTTPResponseWriter{
		header: make(http.Header),
		body:   &bytes.Buffer{},
	}

	rw := NewResponseWriterWithTemplates(mockWriter, tm)

	err = rw.WriteHTML(200, "page", map[string]string{"Content": "Hello"})
	if err != nil {
		t.Fatalf("Failed to write HTML: %v", err)
	}

	result := mockWriter.body.String()
	if !strings.Contains(result, "Hello") {
		t.Errorf("Expected result to contain 'Hello', got %q", result)
	}

	// Check content type
	contentType := mockWriter.header.Get("Content-Type")
	if contentType != "text/html; charset=utf-8" {
		t.Errorf("Expected content type 'text/html; charset=utf-8', got %q", contentType)
	}
}

// TestResponseWriter_WriteHTML_NoTemplateManager tests error when no template manager is set
func TestResponseWriter_WriteHTML_NoTemplateManager(t *testing.T) {
	mockWriter := &mockHTTPResponseWriter{
		header: make(http.Header),
		body:   &bytes.Buffer{},
	}

	rw := NewResponseWriter(mockWriter)

	err := rw.WriteHTML(200, "page", nil)
	if err == nil {
		t.Error("Expected error when template manager is not configured")
	}
}

// TestResponseWriter_SetTemplateManager tests setting template manager
func TestResponseWriter_SetTemplateManager(t *testing.T) {
	tm := NewTemplateManager()
	err := tm.LoadTemplate("test", "{{.Text}}")
	if err != nil {
		t.Fatalf("Failed to load template: %v", err)
	}

	mockWriter := &mockHTTPResponseWriter{
		header: make(http.Header),
		body:   &bytes.Buffer{},
	}

	rw := NewResponseWriter(mockWriter)
	rw.SetTemplateManager(tm)

	err = rw.WriteHTML(200, "test", map[string]string{"Text": "Success"})
	if err != nil {
		t.Fatalf("Failed to write HTML: %v", err)
	}

	result := mockWriter.body.String()
	if result != "Success" {
		t.Errorf("Expected 'Success', got %q", result)
	}
}

// TestTemplateManager_ComplexData tests rendering with complex data structures
func TestTemplateManager_ComplexData(t *testing.T) {
	tm := NewTemplateManager()

	template := `
<html>
<head><title>{{.Title}}</title></head>
<body>
	<h1>{{.Title}}</h1>
	<ul>
	{{range .Items}}
		<li>{{.}}</li>
	{{end}}
	</ul>
</body>
</html>
`

	err := tm.LoadTemplate("complex", template)
	if err != nil {
		t.Fatalf("Failed to load template: %v", err)
	}

	data := map[string]interface{}{
		"Title": "My List",
		"Items": []string{"Item 1", "Item 2", "Item 3"},
	}

	result, err := tm.Render("complex", data)
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	if !strings.Contains(result, "My List") {
		t.Error("Expected result to contain 'My List'")
	}

	if !strings.Contains(result, "Item 1") {
		t.Error("Expected result to contain 'Item 1'")
	}

	if !strings.Contains(result, "Item 2") {
		t.Error("Expected result to contain 'Item 2'")
	}
}

// Mock HTTP response writer for testing
type mockHTTPResponseWriter struct {
	header     http.Header
	body       *bytes.Buffer
	statusCode int
}

func (m *mockHTTPResponseWriter) Header() http.Header {
	return m.header
}

func (m *mockHTTPResponseWriter) Write(data []byte) (int, error) {
	return m.body.Write(data)
}

func (m *mockHTTPResponseWriter) WriteHeader(statusCode int) {
	m.statusCode = statusCode
}
