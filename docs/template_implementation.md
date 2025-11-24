# Template System Implementation

## Overview

The Rockstar Web Framework provides a comprehensive template system built on Go's `html/template` package. The template system supports:

- Go template language syntax
- Context parameter passing to views
- Custom template functions
- Multiple template loading (from strings, files, or filesystems)
- View functions that return response functions
- Template rendering with status codes

## Architecture

### Components

1. **TemplateManager**: Core interface for managing and rendering templates
2. **TemplateRenderer**: Wrapper around TemplateManager for convenience
3. **View Functions**: Helper functions that create view handlers
4. **ResponseWriter Integration**: Templates are rendered through the ResponseWriter

### Template Manager

The `TemplateManager` interface provides the following methods:

```go
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
```

## Usage

### Basic Template Rendering

```go
// Create a template manager
tm := pkg.NewTemplateManager()

// Load a template
err := tm.LoadTemplate("index.html", `
<!DOCTYPE html>
<html>
<head><title>{{.Title}}</title></head>
<body>
    <h1>{{.Title}}</h1>
    <p>{{.Content}}</p>
</body>
</html>
`)

// In a route handler
router.GET("/", func(ctx pkg.Context) error {
    // Set the template manager on the response writer
    ctx.Response().SetTemplateManager(tm)
    
    // Render the template
    return ctx.HTML(200, "index.html", map[string]interface{}{
        "Title": "Welcome",
        "Content": "Hello, World!",
    })
})
```

### Loading Templates from Filesystem

```go
// Load templates from a directory
tm := pkg.NewTemplateManager()
err := tm.LoadTemplates(os.DirFS("templates"), "*.html")
if err != nil {
    log.Fatal(err)
}
```

### Custom Template Functions

```go
tm := pkg.NewTemplateManager()

// Add custom functions before loading templates
err := tm.AddFunc("upper", strings.ToUpper)
err = tm.AddFunc("formatDate", func(t time.Time) string {
    return t.Format("2006-01-02")
})

// Load template that uses custom functions
tm.LoadTemplate("page.html", `
<h1>{{upper .Title}}</h1>
<p>Date: {{formatDate .Date}}</p>
`)
```

### View Functions

View functions provide a clean way to separate template rendering logic:

```go
// Create a view function
homeView := pkg.View("home.html", func(ctx pkg.Context) interface{} {
    return map[string]interface{}{
        "Title": "Home Page",
        "User": ctx.User(),
    }
})

// Use in a route
router.GET("/", func(ctx pkg.Context) error {
    ctx.Response().SetTemplateManager(tm)
    responseFunc := homeView(ctx)
    return responseFunc()
})
```

### View Functions with Custom Status Codes

```go
// Create an error view with 404 status
notFoundView := pkg.ViewWithStatus(404, "error.html", func(ctx pkg.Context) interface{} {
    return map[string]interface{}{
        "StatusCode": 404,
        "Message": "Page not found",
    }
})

router.GET("/not-found", func(ctx pkg.Context) error {
    ctx.Response().SetTemplateManager(tm)
    responseFunc := notFoundView(ctx)
    return responseFunc()
})
```

## Template Syntax

The template system uses Go's standard template syntax:

### Variables

```html
<h1>{{.Title}}</h1>
<p>{{.Content}}</p>
```

### Conditionals

```html
{{if .IsLoggedIn}}
    <p>Welcome, {{.Username}}!</p>
{{else}}
    <p>Please log in.</p>
{{end}}
```

### Loops

```html
<ul>
{{range .Items}}
    <li>{{.}}</li>
{{end}}
</ul>
```

### Nested Data

```html
<h1>{{.User.Name}}</h1>
<p>Email: {{.User.Email}}</p>
```

### Template Functions

```html
<!-- Built-in functions -->
<p>{{len .Items}} items</p>

<!-- Custom functions -->
<h1>{{upper .Title}}</h1>
```

## Best Practices

### 1. Load Templates at Startup

Load all templates when the application starts, not on every request:

```go
func main() {
    tm := pkg.NewTemplateManager()
    
    // Load all templates at startup
    err := tm.LoadTemplates(os.DirFS("templates"), "*.html")
    if err != nil {
        log.Fatal(err)
    }
    
    // Use the same template manager for all requests
    server := pkg.NewServer(config)
    // ... configure routes
}
```

### 2. Set Template Manager Once

For convenience, you can create a middleware to set the template manager:

```go
func templateMiddleware(tm pkg.TemplateManager) pkg.MiddlewareFunc {
    return func(ctx pkg.Context) error {
        ctx.Response().SetTemplateManager(tm)
        return nil
    }
}

// Use the middleware
server.Use(templateMiddleware(tm))
```

### 3. Use View Functions for Complex Logic

Separate template data preparation from rendering:

```go
func userProfileView(userID string) pkg.ViewFunc {
    return pkg.View("profile.html", func(ctx pkg.Context) interface{} {
        // Fetch user data
        user, err := getUserByID(ctx, userID)
        if err != nil {
            return map[string]interface{}{
                "Error": err.Error(),
            }
        }
        
        // Prepare template data
        return map[string]interface{}{
            "User": user,
            "Posts": getUserPosts(ctx, userID),
            "Followers": getUserFollowers(ctx, userID),
        }
    })
}
```

### 4. Handle Template Errors

Always check for template rendering errors:

```go
router.GET("/", func(ctx pkg.Context) error {
    ctx.Response().SetTemplateManager(tm)
    
    err := ctx.HTML(200, "index.html", data)
    if err != nil {
        // Log the error
        ctx.Logger().Error("Template rendering failed", "error", err)
        
        // Return a simple error response
        return ctx.String(500, "Internal Server Error")
    }
    
    return nil
})
```

### 5. Use Template Inheritance

Create base templates and extend them:

```go
// base.html
tm.LoadTemplate("base.html", `
<!DOCTYPE html>
<html>
<head>
    <title>{{.Title}}</title>
</head>
<body>
    {{template "content" .}}
</body>
</html>
`)

// page.html
tm.LoadTemplate("page.html", `
{{define "content"}}
    <h1>{{.Heading}}</h1>
    <p>{{.Content}}</p>
{{end}}
`)
```

## Performance Considerations

### Template Caching

Templates are parsed once and cached in memory. Subsequent renders use the cached templates, making rendering very fast.

### Memory Usage

The template manager stores all parsed templates in memory. For applications with many templates, consider:

1. Loading only necessary templates
2. Using template groups for different sections of your application
3. Clearing unused templates with `Clear()` if needed

### Concurrent Access

The template manager is thread-safe and can be used concurrently from multiple goroutines.

## Error Handling

### Template Not Found

```go
err := ctx.HTML(200, "nonexistent.html", data)
// Returns: "template nonexistent.html not found"
```

### Template Execution Error

```go
// If template references missing data
err := ctx.HTML(200, "page.html", map[string]string{})
// Returns: "failed to execute template page.html: ..."
```

### No Template Manager

```go
// If template manager is not set
err := ctx.HTML(200, "page.html", data)
// Returns: "template manager not configured"
```

## Integration with Framework Features

### With Internationalization

```go
homeView := pkg.View("home.html", func(ctx pkg.Context) interface{} {
    i18n := ctx.I18n()
    
    return map[string]interface{}{
        "Title": i18n.Translate("home.title"),
        "Welcome": i18n.Translate("home.welcome"),
    }
})
```

### With Authentication

```go
dashboardView := pkg.View("dashboard.html", func(ctx pkg.Context) interface{} {
    if !ctx.IsAuthenticated() {
        return map[string]interface{}{
            "Error": "Please log in",
        }
    }
    
    return map[string]interface{}{
        "User": ctx.User(),
        "Data": getDashboardData(ctx),
    }
})
```

### With Database

```go
listView := pkg.View("list.html", func(ctx pkg.Context) interface{} {
    db := ctx.DB()
    
    var items []Item
    err := db.Find(&items)
    if err != nil {
        return map[string]interface{}{
            "Error": err.Error(),
        }
    }
    
    return map[string]interface{}{
        "Items": items,
    }
})
```

## Testing

### Unit Testing Templates

```go
func TestTemplateRendering(t *testing.T) {
    tm := pkg.NewTemplateManager()
    
    err := tm.LoadTemplate("test", "<h1>{{.Title}}</h1>")
    if err != nil {
        t.Fatal(err)
    }
    
    result, err := tm.Render("test", map[string]string{
        "Title": "Test",
    })
    if err != nil {
        t.Fatal(err)
    }
    
    expected := "<h1>Test</h1>"
    if result != expected {
        t.Errorf("Expected %q, got %q", expected, result)
    }
}
```

### Testing View Functions

```go
func TestViewFunction(t *testing.T) {
    tm := pkg.NewTemplateManager()
    tm.LoadTemplate("test", "{{.Message}}")
    
    // Create a mock context
    ctx := createMockContext()
    ctx.Response().SetTemplateManager(tm)
    
    // Create and execute view
    view := pkg.View("test", func(c pkg.Context) interface{} {
        return map[string]string{"Message": "Hello"}
    })
    
    responseFunc := view(ctx)
    err := responseFunc()
    
    if err != nil {
        t.Fatal(err)
    }
}
```

## Requirements Validation

This implementation satisfies the following requirements:

- **Requirement 13.1**: Supports Go template language ✓
- **Requirement 13.2**: Provides context parameter to views ✓
- **Requirement 13.3**: All views return response functions ✓
- **Requirement 13.4**: Comprehensive unit tests for template rendering ✓

## Conclusion

The template system provides a powerful and flexible way to render HTML responses in the Rockstar Web Framework. It integrates seamlessly with other framework features and follows Go best practices for template management.
