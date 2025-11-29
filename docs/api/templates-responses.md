# Templates and Response API Reference

## Overview

The Rockstar Web Framework provides comprehensive template rendering and response writing capabilities with support for HTML templates, JSON, XML, streaming, and more.

## Template Management

### TemplateManager Interface

Interface for managing and rendering templates.

```go
type TemplateManager interface {
    LoadTemplates(fsys fs.FS, pattern string) error
    LoadTemplate(name string, content string) error
    Render(name string, data interface{}) (string, error)
    RenderTo(w io.Writer, name string, data interface{}) error
    HasTemplate(name string) bool
    AddFunc(name string, fn interface{}) error
    Clear()
}
```

### NewTemplateManager()

Creates a new template manager instance.

**Signature:**
```go
func NewTemplateManager() TemplateManager
```

**Returns:**
- `TemplateManager` - Template manager instance

**Example:**
```go
tm := pkg.NewTemplateManager()
```

### LoadTemplates()

Loads templates from a filesystem with a pattern.

**Signature:**
```go
LoadTemplates(fsys fs.FS, pattern string) error
```

**Parameters:**
- `fsys` - Filesystem to load from
- `pattern` - Glob pattern for template files (e.g., "*.html", "templates/*.tmpl")

**Returns:**
- `error` - Error if loading fails

**Example:**
```go
tm := pkg.NewTemplateManager()

// Load from embedded filesystem
//go:embed templates/*.html
var templatesFS embed.FS

err := tm.LoadTemplates(templatesFS, "templates/*.html")
if err != nil {
    log.Fatal(err)
}
```

### LoadTemplate()

Loads a single template from a string.

**Signature:**
```go
LoadTemplate(name string, content string) error
```

**Parameters:**
- `name` - Template name
- `content` - Template content

**Returns:**
- `error` - Error if parsing fails

**Example:**
```go
tm := pkg.NewTemplateManager()

template := `
<!DOCTYPE html>
<html>
<head><title>{{.Title}}</title></head>
<body>
    <h1>{{.Heading}}</h1>
    <p>{{.Content}}</p>
</body>
</html>
`

err := tm.LoadTemplate("page.html", template)
if err != nil {
    log.Fatal(err)
}
```

### Render()

Renders a template with data and returns the result as a string.

**Signature:**
```go
Render(name string, data interface{}) (string, error)
```

**Parameters:**
- `name` - Template name
- `data` - Data to pass to template

**Returns:**
- `string` - Rendered HTML
- `error` - Error if rendering fails

**Example:**
```go
tm := pkg.NewTemplateManager()
tm.LoadTemplate("greeting.html", "<h1>Hello, {{.Name}}!</h1>")

html, err := tm.Render("greeting.html", map[string]interface{}{
    "Name": "Alice",
})
if err != nil {
    log.Fatal(err)
}
// html: "<h1>Hello, Alice!</h1>"
```

### RenderTo()

Renders a template with data to a writer.

**Signature:**
```go
RenderTo(w io.Writer, name string, data interface{}) error
```

**Parameters:**
- `w` - Writer to render to
- `name` - Template name
- `data` - Data to pass to template

**Returns:**
- `error` - Error if rendering fails

**Example:**
```go
tm := pkg.NewTemplateManager()
tm.LoadTemplate("page.html", "<h1>{{.Title}}</h1>")

var buf bytes.Buffer
err := tm.RenderTo(&buf, "page.html", map[string]interface{}{
    "Title": "Welcome",
})
if err != nil {
    log.Fatal(err)
}
```

### HasTemplate()

Checks if a template exists.

**Signature:**
```go
HasTemplate(name string) bool
```

**Parameters:**
- `name` - Template name

**Returns:**
- `bool` - true if template exists, false otherwise

**Example:**
```go
tm := pkg.NewTemplateManager()

if tm.HasTemplate("page.html") {
    // Template exists
} else {
    // Template not found
}
```

### AddFunc()

Adds a custom function to the template function map. Must be called before loading templates.

**Signature:**
```go
AddFunc(name string, fn interface{}) error
```

**Parameters:**
- `name` - Function name
- `fn` - Function implementation

**Returns:**
- `error` - Error if templates already loaded

**Example:**
```go
tm := pkg.NewTemplateManager()

// Add custom function
tm.AddFunc("upper", strings.ToUpper)
tm.AddFunc("formatDate", func(t time.Time) string {
    return t.Format("2006-01-02")
})

// Load templates (functions must be added before this)
tm.LoadTemplate("page.html", "<h1>{{upper .Title}}</h1>")

html, _ := tm.Render("page.html", map[string]interface{}{
    "Title": "hello",
})
// html: "<h1>HELLO</h1>"
```

### Clear()

Clears all loaded templates.

**Signature:**
```go
Clear()
```

**Example:**
```go
tm := pkg.NewTemplateManager()
tm.LoadTemplate("page.html", "<h1>Test</h1>")

tm.Clear() // All templates removed
```

## Template Renderer

### TemplateRenderer

Wrapper around TemplateManager with additional convenience methods.

```go
type TemplateRenderer struct {
    // private fields
}
```

### NewTemplateRenderer()

Creates a new template renderer.

**Signature:**
```go
func NewTemplateRenderer(manager TemplateManager) *TemplateRenderer
```

**Parameters:**
- `manager` - Template manager instance

**Returns:**
- `*TemplateRenderer` - Template renderer

**Example:**
```go
tm := pkg.NewTemplateManager()
renderer := pkg.NewTemplateRenderer(tm)
```

### TemplateRenderer.Render()

Renders a template to a writer.

**Signature:**
```go
Render(w io.Writer, name string, data interface{}) error
```

### TemplateRenderer.LoadFromFS()

Loads templates from a filesystem.

**Signature:**
```go
LoadFromFS(fsys fs.FS, pattern string) error
```

### TemplateRenderer.LoadFromString()

Loads a template from a string.

**Signature:**
```go
LoadFromString(name string, content string) error
```

### TemplateRenderer.AddFunc()

Adds a custom template function.

**Signature:**
```go
AddFunc(name string, fn interface{}) error
```

## View Functions

### ViewFunc

Type representing a view function that returns a response function.

```go
type ViewFunc func(ctx Context) ResponseFunc
type ResponseFunc func() error
```

### View()

Creates a view function that renders a template with status 200.

**Signature:**
```go
func View(templateName string, dataFunc func(ctx Context) interface{}) ViewFunc
```

**Parameters:**
- `templateName` - Template name to render
- `dataFunc` - Function that returns template data

**Returns:**
- `ViewFunc` - View function

**Example:**
```go
router.GET("/", pkg.View("index.html", func(ctx pkg.Context) interface{} {
    return map[string]interface{}{
        "Title": "Home",
        "User":  ctx.User(),
    }
}))
```

### ViewWithStatus()

Creates a view function that renders a template with a custom status code.

**Signature:**
```go
func ViewWithStatus(statusCode int, templateName string, dataFunc func(ctx Context) interface{}) ViewFunc
```

**Parameters:**
- `statusCode` - HTTP status code
- `templateName` - Template name to render
- `dataFunc` - Function that returns template data

**Returns:**
- `ViewFunc` - View function

**Example:**
```go
router.GET("/error", pkg.ViewWithStatus(500, "error.html", func(ctx pkg.Context) interface{} {
    return map[string]interface{}{
        "Error": "Something went wrong",
    }
}))
```

## Response Writer

### ResponseWriter Interface

Interface for writing HTTP responses.

```go
type ResponseWriter interface {
    http.ResponseWriter
    
    // Response methods
    WriteJSON(statusCode int, data interface{}) error
    WriteXML(statusCode int, data interface{}) error
    WriteHTML(statusCode int, templateName string, data interface{}) error
    WriteString(statusCode int, message string) error
    WriteStream(statusCode int, contentType string, reader io.Reader) error
    
    // Cookie and header methods
    SetCookie(cookie *Cookie) error
    SetHeader(key, value string)
    SetContentType(contentType string)
    
    // Status and metadata
    Status() int
    Size() int64
    Written() bool
    
    // Advanced features
    Flush() error
    Close() error
    Hijack() (net.Conn, *bufio.ReadWriter, error)
    Push(target string, opts *http.PushOptions) error
    
    // Template support
    SetTemplateManager(tm TemplateManager)
}
```

### NewResponseWriter()

Creates a new response writer.

**Signature:**
```go
func NewResponseWriter(w http.ResponseWriter) ResponseWriter
```

**Parameters:**
- `w` - HTTP response writer

**Returns:**
- `ResponseWriter` - Response writer instance

### NewResponseWriterWithTemplates()

Creates a new response writer with template support.

**Signature:**
```go
func NewResponseWriterWithTemplates(w http.ResponseWriter, tm TemplateManager) ResponseWriter
```

**Parameters:**
- `w` - HTTP response writer
- `tm` - Template manager

**Returns:**
- `ResponseWriter` - Response writer instance

### WriteJSON()

Writes a JSON response.

**Signature:**
```go
WriteJSON(statusCode int, data interface{}) error
```

**Parameters:**
- `statusCode` - HTTP status code
- `data` - Data to serialize as JSON

**Returns:**
- `error` - Error if serialization fails

**Example:**
```go
router.GET("/api/users", func(ctx pkg.Context) error {
    users := []User{{ID: 1, Name: "Alice"}, {ID: 2, Name: "Bob"}}
    return ctx.Response().WriteJSON(200, map[string]interface{}{
        "users": users,
        "total": len(users),
    })
})
```

### WriteXML()

Writes an XML response.

**Signature:**
```go
WriteXML(statusCode int, data interface{}) error
```

**Parameters:**
- `statusCode` - HTTP status code
- `data` - Data to serialize as XML

**Returns:**
- `error` - Error if serialization fails

**Example:**
```go
type User struct {
    ID   int    `xml:"id"`
    Name string `xml:"name"`
}

router.GET("/api/users.xml", func(ctx pkg.Context) error {
    user := User{ID: 1, Name: "Alice"}
    return ctx.Response().WriteXML(200, user)
})
```

### WriteHTML()

Writes an HTML response using a template.

**Signature:**
```go
WriteHTML(statusCode int, templateName string, data interface{}) error
```

**Parameters:**
- `statusCode` - HTTP status code
- `templateName` - Template name
- `data` - Data to pass to template

**Returns:**
- `error` - Error if rendering fails

**Example:**
```go
router.GET("/", func(ctx pkg.Context) error {
    return ctx.Response().WriteHTML(200, "index.html", map[string]interface{}{
        "Title": "Home",
        "User":  ctx.User(),
    })
})
```

### WriteString()

Writes a plain text response.

**Signature:**
```go
WriteString(statusCode int, message string) error
```

**Parameters:**
- `statusCode` - HTTP status code
- `message` - Text message

**Returns:**
- `error` - Error if writing fails

**Example:**
```go
router.GET("/health", func(ctx pkg.Context) error {
    return ctx.Response().WriteString(200, "OK")
})
```

### WriteStream()

Writes a streaming response.

**Signature:**
```go
WriteStream(statusCode int, contentType string, reader io.Reader) error
```

**Parameters:**
- `statusCode` - HTTP status code
- `contentType` - Content type
- `reader` - Reader to stream from

**Returns:**
- `error` - Error if streaming fails

**Example:**
```go
router.GET("/download/:file", func(ctx pkg.Context) error {
    filename := ctx.Param("file")
    file, err := os.Open("/uploads/" + filename)
    if err != nil {
        return ctx.JSON(404, map[string]string{"error": "File not found"})
    }
    defer file.Close()
    
    return ctx.Response().WriteStream(200, "application/octet-stream", file)
})
```

### Status()

Returns the HTTP status code.

**Signature:**
```go
Status() int
```

**Returns:**
- `int` - HTTP status code

### Size()

Returns the response size in bytes.

**Signature:**
```go
Size() int64
```

**Returns:**
- `int64` - Response size

### Written()

Returns whether the response has been written.

**Signature:**
```go
Written() bool
```

**Returns:**
- `bool` - true if response written, false otherwise

### Flush()

Flushes the response buffer.

**Signature:**
```go
Flush() error
```

**Returns:**
- `error` - Error if flushing not supported

**Example:**
```go
router.GET("/stream", func(ctx pkg.Context) error {
    resp := ctx.Response()
    
    for i := 0; i < 10; i++ {
        resp.Write([]byte(fmt.Sprintf("Chunk %d\n", i)))
        resp.Flush() // Send immediately
        time.Sleep(time.Second)
    }
    
    return nil
})
```

### Hijack()

Hijacks the connection for WebSocket or custom protocols.

**Signature:**
```go
Hijack() (net.Conn, *bufio.ReadWriter, error)
```

**Returns:**
- `net.Conn` - Network connection
- `*bufio.ReadWriter` - Buffered reader/writer
- `error` - Error if hijacking not supported

### Push()

Pushes a resource using HTTP/2 server push.

**Signature:**
```go
Push(target string, opts *http.PushOptions) error
```

**Parameters:**
- `target` - Resource path to push
- `opts` - Push options

**Returns:**
- `error` - Error if push not supported

**Example:**
```go
router.GET("/", func(ctx pkg.Context) error {
    resp := ctx.Response()
    
    // Push CSS file
    resp.Push("/static/style.css", nil)
    
    return ctx.HTML(200, "index.html", nil)
})
```

## Complete Examples

### Template-Based Web Application

```go
package main

import (
    "embed"
    "github.com/echterhof/rockstar-web-framework/pkg"
)

//go:embed templates/*.html
var templatesFS embed.FS

func main() {
    config := pkg.FrameworkConfig{}
    app, _ := pkg.New(config)
    
    // Setup template manager
    tm := pkg.NewTemplateManager()
    tm.AddFunc("formatDate", func(t time.Time) string {
        return t.Format("Jan 02, 2006")
    })
    tm.LoadTemplates(templatesFS, "templates/*.html")
    
    router := app.Router()
    
    // Home page
    router.GET("/", func(ctx pkg.Context) error {
        return ctx.HTML(200, "index.html", map[string]interface{}{
            "Title": "Welcome",
            "User":  ctx.User(),
        })
    })
    
    // Blog post
    router.GET("/posts/:id", func(ctx pkg.Context) error {
        postID := ctx.Param("id")
        
        // Fetch post from database
        post := fetchPost(postID)
        
        return ctx.HTML(200, "post.html", map[string]interface{}{
            "Post": post,
        })
    })
    
    app.Listen(":8080")
}
```

### API with Multiple Response Formats

```go
router.GET("/api/users/:id", func(ctx pkg.Context) error {
    userID := ctx.Param("id")
    user := fetchUser(userID)
    
    // Check Accept header
    accept := ctx.GetHeader("Accept")
    
    switch {
    case strings.Contains(accept, "application/json"):
        return ctx.JSON(200, user)
    case strings.Contains(accept, "application/xml"):
        return ctx.XML(200, user)
    case strings.Contains(accept, "text/html"):
        return ctx.HTML(200, "user.html", user)
    default:
        return ctx.JSON(200, user)
    }
})
```

### Streaming Response

```go
router.GET("/logs/stream", func(ctx pkg.Context) error {
    resp := ctx.Response()
    resp.SetContentType("text/event-stream")
    resp.SetHeader("Cache-Control", "no-cache")
    resp.SetHeader("Connection", "keep-alive")
    resp.WriteHeader(200)
    
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()
    
    for i := 0; i < 10; i++ {
        <-ticker.C
        
        data := fmt.Sprintf("data: Log entry %d\n\n", i)
        resp.Write([]byte(data))
        resp.Flush()
    }
    
    return nil
})
```

## Best Practices

### Templates

1. **Preload Templates:** Load templates at startup, not per request
2. **Use Embedded FS:** Embed templates in binary for easy deployment
3. **Add Functions Early:** Add custom functions before loading templates
4. **Cache Rendered Output:** Cache frequently rendered templates
5. **Escape Data:** Use `{{.}}` for auto-escaping, `{{. | html}}` for manual
6. **Template Inheritance:** Use `{{template}}` for reusable components
7. **Error Handling:** Always check template rendering errors

### Responses

1. **Set Content-Type:** Always set appropriate Content-Type header
2. **Use Appropriate Method:** Use JSON() for APIs, HTML() for pages
3. **Check Written:** Check if response already written before writing again
4. **Flush Strategically:** Only flush when needed for streaming
5. **Handle Errors:** Return errors from response methods
6. **Set Status First:** Set status code before writing body
7. **Use Streaming:** Use streaming for large responses

## Security Considerations

1. **Template Injection:** Validate template names from user input
2. **XSS Prevention:** Use auto-escaping in templates
3. **Content-Type:** Set correct Content-Type to prevent MIME sniffing
4. **Cache Headers:** Set appropriate cache headers for sensitive data
5. **Error Messages:** Don't expose internal errors in responses
6. **File Streaming:** Validate file paths before streaming
7. **Template Data:** Sanitize data passed to templates

## See Also

- [Context API](context.md) - Context response methods
- [Router API](router.md) - Route registration
- [WebSocket API](websockets.md) - WebSocket support
- [Middleware Guide](../guides/middleware.md) - Middleware patterns
