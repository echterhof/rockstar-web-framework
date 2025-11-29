# Request and Response API

Complete API documentation for HTTP request and response handling in the Rockstar Web Framework.

## Overview

The framework provides comprehensive request and response handling with support for multiple content types, file uploads, cookies, headers, and streaming. All request/response operations are accessed through the `Context` interface.

## Table of Contents

- [Request API](#request-api)
- [Response API](#response-api)
- [Request Data Access](#request-data-access)
- [Response Methods](#response-methods)
- [File Uploads](#file-uploads)
- [Cookies](#cookies)
- [Headers](#headers)
- [Streaming](#streaming)

---

## Request API

### Request Structure

```go
type Request struct {
    // Standard HTTP fields
    Method     string
    URL        *url.URL
    Proto      string
    Header     http.Header
    Body       io.ReadCloser
    Host       string
    RemoteAddr string
    RequestURI string
    
    // Framework-specific fields
    ID        string               // Unique request ID
    TenantID  string               // Multi-tenancy support
    UserID    string               // Authenticated user ID
    StartTime time.Time            // Request start time
    Params    map[string]string    // Route parameters
    Query     map[string]string    // Query parameters
    Form      map[string]string    // Form data
    Files     map[string]*FormFile // Uploaded files
    
    // Security context
    AccessToken string // API access token
    SessionID   string // Session identifier
    
    // Protocol information
    IsWebSocket bool   // WebSocket upgrade request
    Protocol    string // HTTP/1, HTTP/2, QUIC, WebSocket
    
    // Raw request data
    RawBody []byte // Cached body content
}
```

### Accessing Request

```go
func handler(ctx pkg.Context) error {
    req := ctx.Request()
    
    // HTTP method
    method := req.Method // "GET", "POST", etc.
    
    // Request path
    path := req.URL.Path
    
    // Request ID (unique per request)
    requestID := req.ID
    
    // Client information
    clientIP := req.RemoteAddr
    userAgent := req.Header.Get("User-Agent")
    
    return ctx.JSON(200, map[string]interface{}{
        "method": method,
        "path": path,
        "request_id": requestID,
    })
}
```

---

## Response API

### ResponseWriter Interface

```go
type ResponseWriter interface {
    // Standard HTTP response methods
    Header() http.Header
    Write([]byte) (int, error)
    WriteHeader(statusCode int)
    
    // Framework-specific methods
    WriteJSON(statusCode int, data interface{}) error
    WriteXML(statusCode int, data interface{}) error
    WriteHTML(statusCode int, template string, data interface{}) error
    WriteString(statusCode int, message string) error
    
    // Stream support
    WriteStream(statusCode int, contentType string, reader io.Reader) error
    
    // Cookie support
    SetCookie(cookie *Cookie) error
    
    // Header helpers
    SetHeader(key, value string)
    SetContentType(contentType string)
    
    // Status and size tracking
    Status() int
    Size() int64
    Written() bool
    
    // Response control
    Flush() error
    Close() error
    
    // Template support
    SetTemplateManager(tm TemplateManager)
}
```

### Response Structure

```go
type Response struct {
    StatusCode int
    Header     http.Header
    Body       []byte
    Size       int64
    
    // Framework-specific fields
    Template string      // Template name for HTML responses
    Data     interface{} // Response data
    Cookies  []*Cookie   // Cookies to set
    
    // Streaming support
    Stream     io.Reader // Stream reader for large responses
    StreamType string    // Content type for streaming
}
```

---

## Request Data Access

### Route Parameters

Extract parameters from URL path patterns.

```go
// Route: /users/:id/posts/:postID
app.Router().GET("/users/:id/posts/:postID", func(ctx pkg.Context) error {
    userID := ctx.Param("id")
    postID := ctx.Param("postID")
    
    // Or get all params
    params := ctx.Params()
    
    return ctx.JSON(200, map[string]interface{}{
        "user_id": userID,
        "post_id": postID,
        "all_params": params,
    })
})
```

### Query Parameters

Extract parameters from URL query string.

```go
// URL: /search?q=golang&page=2&limit=10
func searchHandler(ctx pkg.Context) error {
    query := ctx.Query()
    
    searchTerm := query["q"]      // "golang"
    page := query["page"]          // "2"
    limit := query["limit"]        // "10"
    
    // With default values
    pageNum, _ := strconv.Atoi(query["page"])
    if pageNum == 0 {
        pageNum = 1
    }
    
    return ctx.JSON(200, map[string]interface{}{
        "search": searchTerm,
        "page": pageNum,
    })
}
```

### Request Body

Access raw request body data.

```go
func handler(ctx pkg.Context) error {
    // Get raw body
    body := ctx.Body()
    
    // Parse JSON
    var data map[string]interface{}
    if err := json.Unmarshal(body, &data); err != nil {
        return pkg.NewValidationError("Invalid JSON", "body")
    }
    
    return ctx.JSON(200, data)
}
```

### Form Data

Access form-encoded data.

```go
func formHandler(ctx pkg.Context) error {
    // Get single form value
    username := ctx.FormValue("username")
    email := ctx.FormValue("email")
    
    // Validate
    if username == "" {
        return pkg.NewMissingFieldError("username")
    }
    
    return ctx.JSON(200, map[string]interface{}{
        "username": username,
        "email": email,
    })
}
```

---

## Response Methods

### JSON Response

```go
func jsonHandler(ctx pkg.Context) error {
    data := map[string]interface{}{
        "message": "Success",
        "user": map[string]interface{}{
            "id": 123,
            "name": "John Doe",
        },
        "timestamp": time.Now(),
    }
    
    return ctx.JSON(200, data)
}

// Response:
// Content-Type: application/json
// {
//   "message": "Success",
//   "user": {
//     "id": 123,
//     "name": "John Doe"
//   },
//   "timestamp": "2025-11-29T10:00:00Z"
// }
```

### XML Response

```go
type User struct {
    XMLName xml.Name `xml:"user"`
    ID      int      `xml:"id"`
    Name    string   `xml:"name"`
}

func xmlHandler(ctx pkg.Context) error {
    user := User{
        ID: 123,
        Name: "John Doe",
    }
    
    return ctx.XML(200, user)
}

// Response:
// Content-Type: application/xml
// <?xml version="1.0"?>
// <user>
//   <id>123</id>
//   <name>John Doe</name>
// </user>
```

### HTML Response

```go
func htmlHandler(ctx pkg.Context) error {
    data := map[string]interface{}{
        "Title": "Welcome",
        "User": map[string]string{
            "Name": "John Doe",
        },
    }
    
    return ctx.HTML(200, "welcome.html", data)
}

// Template: welcome.html
// <html>
//   <head><title>{{.Title}}</title></head>
//   <body>
//     <h1>Welcome, {{.User.Name}}!</h1>
//   </body>
// </html>
```

### Plain Text Response

```go
func textHandler(ctx pkg.Context) error {
    return ctx.String(200, "Hello, World!")
}

// Response:
// Content-Type: text/plain; charset=utf-8
// Hello, World!
```

### Redirect Response

```go
func redirectHandler(ctx pkg.Context) error {
    // Temporary redirect (302)
    return ctx.Redirect(302, "/new-location")
    
    // Permanent redirect (301)
    // return ctx.Redirect(301, "/new-location")
}
```

---

## File Uploads

### FormFile Structure

```go
type FormFile struct {
    Filename string
    Header   map[string][]string
    Size     int64
    Content  []byte
}
```

### Single File Upload

```go
func uploadHandler(ctx pkg.Context) error {
    // Get uploaded file
    file, err := ctx.FormFile("avatar")
    if err != nil {
        return pkg.NewValidationError("File required", "avatar")
    }
    
    // Validate file size (5MB limit)
    maxSize := int64(5 * 1024 * 1024)
    if file.Size > maxSize {
        return pkg.NewFrameworkError(
            pkg.ErrCodeFileTooLarge,
            "File too large",
            413,
        ).WithDetails(map[string]interface{}{
            "max_size": maxSize,
            "file_size": file.Size,
        })
    }
    
    // Validate file type
    contentType := file.Header["Content-Type"][0]
    allowedTypes := []string{"image/jpeg", "image/png", "image/gif"}
    valid := false
    for _, t := range allowedTypes {
        if contentType == t {
            valid = true
            break
        }
    }
    
    if !valid {
        return pkg.NewFrameworkError(
            pkg.ErrCodeInvalidFileType,
            "Invalid file type",
            400,
        ).WithDetails(map[string]interface{}{
            "allowed_types": allowedTypes,
            "file_type": contentType,
        })
    }
    
    // Save file
    destPath := fmt.Sprintf("uploads/%s", file.Filename)
    if err := ctx.Files().Write(destPath, file.Content); err != nil {
        return err
    }
    
    return ctx.JSON(200, map[string]interface{}{
        "message": "File uploaded successfully",
        "filename": file.Filename,
        "size": file.Size,
        "path": destPath,
    })
}
```

### Multiple File Upload

```go
func multiUploadHandler(ctx pkg.Context) error {
    req := ctx.Request()
    
    uploaded := []map[string]interface{}{}
    
    for fieldName, file := range req.Files {
        // Process each file
        destPath := fmt.Sprintf("uploads/%s", file.Filename)
        if err := ctx.Files().Write(destPath, file.Content); err != nil {
            return err
        }
        
        uploaded = append(uploaded, map[string]interface{}{
            "field": fieldName,
            "filename": file.Filename,
            "size": file.Size,
            "path": destPath,
        })
    }
    
    return ctx.JSON(200, map[string]interface{}{
        "message": "Files uploaded successfully",
        "files": uploaded,
    })
}
```

### File Upload with Form Data

```go
func uploadWithDataHandler(ctx pkg.Context) error {
    // Get form data
    title := ctx.FormValue("title")
    description := ctx.FormValue("description")
    
    // Get file
    file, err := ctx.FormFile("document")
    if err != nil {
        return pkg.NewValidationError("File required", "document")
    }
    
    // Save file
    destPath := fmt.Sprintf("documents/%s", file.Filename)
    if err := ctx.Files().Write(destPath, file.Content); err != nil {
        return err
    }
    
    // Save metadata to database
    doc := &Document{
        Title: title,
        Description: description,
        Filename: file.Filename,
        Path: destPath,
        Size: file.Size,
    }
    
    if err := ctx.DB().SaveDocument(doc); err != nil {
        return err
    }
    
    return ctx.JSON(201, doc)
}
```

---

## Cookies

### Cookie Structure

```go
type Cookie struct {
    Name     string
    Value    string
    Path     string
    Domain   string
    Expires  time.Time
    MaxAge   int
    Secure   bool
    HttpOnly bool
    SameSite http.SameSite
    
    // Framework extensions
    Encrypted bool // Whether cookie value is encrypted
}
```

### Setting Cookies

```go
func setCookieHandler(ctx pkg.Context) error {
    cookie := &pkg.Cookie{
        Name:     "user_preference",
        Value:    "dark_mode",
        Path:     "/",
        MaxAge:   86400, // 1 day
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteLaxMode,
    }
    
    if err := ctx.SetCookie(cookie); err != nil {
        return err
    }
    
    return ctx.String(200, "Cookie set")
}
```

### Reading Cookies

```go
func getCookieHandler(ctx pkg.Context) error {
    cookie, err := ctx.GetCookie("user_preference")
    if err != nil {
        return ctx.String(200, "Cookie not found")
    }
    
    return ctx.JSON(200, map[string]interface{}{
        "name": cookie.Name,
        "value": cookie.Value,
    })
}
```

### Deleting Cookies

```go
func deleteCookieHandler(ctx pkg.Context) error {
    cookie := &pkg.Cookie{
        Name:   "user_preference",
        Value:  "",
        Path:   "/",
        MaxAge: -1, // Delete cookie
    }
    
    if err := ctx.SetCookie(cookie); err != nil {
        return err
    }
    
    return ctx.String(200, "Cookie deleted")
}
```

---

## Headers

### Setting Response Headers

```go
func headersHandler(ctx pkg.Context) error {
    // Set individual headers
    ctx.SetHeader("X-Custom-Header", "value")
    ctx.SetHeader("X-Request-ID", ctx.Request().ID)
    
    // Set content type
    ctx.Response().SetContentType("application/json")
    
    // Set cache control
    ctx.SetHeader("Cache-Control", "public, max-age=3600")
    
    // Set CORS headers
    ctx.SetHeader("Access-Control-Allow-Origin", "*")
    ctx.SetHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
    
    return ctx.JSON(200, map[string]string{"status": "ok"})
}
```

### Reading Request Headers

```go
func readHeadersHandler(ctx pkg.Context) error {
    // Get single header
    userAgent := ctx.GetHeader("User-Agent")
    authorization := ctx.GetHeader("Authorization")
    
    // Get all headers
    headers := ctx.Headers()
    
    return ctx.JSON(200, map[string]interface{}{
        "user_agent": userAgent,
        "authorization": authorization,
        "all_headers": headers,
    })
}
```

### Common Headers

```go
// Content negotiation
func contentNegotiationHandler(ctx pkg.Context) error {
    accept := ctx.GetHeader("Accept")
    
    switch {
    case strings.Contains(accept, "application/json"):
        return ctx.JSON(200, data)
    case strings.Contains(accept, "application/xml"):
        return ctx.XML(200, data)
    case strings.Contains(accept, "text/html"):
        return ctx.HTML(200, "template.html", data)
    default:
        return ctx.String(200, "text response")
    }
}

// Authentication
func authHeaderHandler(ctx pkg.Context) error {
    auth := ctx.GetHeader("Authorization")
    if !strings.HasPrefix(auth, "Bearer ") {
        return pkg.NewAuthenticationError("Invalid authorization header")
    }
    
    token := strings.TrimPrefix(auth, "Bearer ")
    // Validate token...
    
    return ctx.JSON(200, map[string]string{"status": "authenticated"})
}
```

---

## Streaming

### Streaming Response

```go
func streamHandler(ctx pkg.Context) error {
    // Create a reader (e.g., file, database cursor, etc.)
    file, err := os.Open("large-file.dat")
    if err != nil {
        return err
    }
    defer file.Close()
    
    // Stream the response
    return ctx.Response().WriteStream(
        200,
        "application/octet-stream",
        file,
    )
}
```

### Chunked Response

```go
func chunkedHandler(ctx pkg.Context) error {
    ctx.SetHeader("Transfer-Encoding", "chunked")
    ctx.Response().WriteHeader(200)
    
    // Write chunks
    for i := 0; i < 10; i++ {
        chunk := fmt.Sprintf("Chunk %d\n", i)
        ctx.Response().Write([]byte(chunk))
        ctx.Response().Flush()
        time.Sleep(100 * time.Millisecond)
    }
    
    return nil
}
```

### Server-Sent Events (SSE)

```go
func sseHandler(ctx pkg.Context) error {
    ctx.SetHeader("Content-Type", "text/event-stream")
    ctx.SetHeader("Cache-Control", "no-cache")
    ctx.SetHeader("Connection", "keep-alive")
    
    ctx.Response().WriteHeader(200)
    
    // Send events
    for i := 0; i < 10; i++ {
        event := fmt.Sprintf("data: {\"message\": \"Event %d\"}\n\n", i)
        ctx.Response().Write([]byte(event))
        ctx.Response().Flush()
        time.Sleep(1 * time.Second)
    }
    
    return nil
}
```

---

## Advanced Patterns

### Request Validation Pipeline

```go
func validateRequest(ctx pkg.Context) error {
    // Validate content type
    contentType := ctx.GetHeader("Content-Type")
    if !strings.Contains(contentType, "application/json") {
        return pkg.NewValidationError("Content-Type must be application/json", "header")
    }
    
    // Validate body size
    if len(ctx.Body()) > 1024*1024 { // 1MB
        return pkg.NewRequestTooLargeError(1024 * 1024)
    }
    
    // Parse and validate JSON
    var data map[string]interface{}
    if err := json.Unmarshal(ctx.Body(), &data); err != nil {
        return pkg.NewValidationError("Invalid JSON", "body")
    }
    
    // Validate required fields
    required := []string{"name", "email"}
    for _, field := range required {
        if data[field] == nil || data[field] == "" {
            return pkg.NewMissingFieldError(field)
        }
    }
    
    return nil
}
```

### Response Caching

```go
func cachedHandler(ctx pkg.Context) error {
    cacheKey := "response:" + ctx.Request().Path
    
    // Check cache
    if cached, err := ctx.Cache().Get(cacheKey); err == nil {
        ctx.SetHeader("X-Cache", "HIT")
        return ctx.JSON(200, cached)
    }
    
    // Generate response
    data := generateExpensiveData()
    
    // Cache response
    ctx.Cache().Set(cacheKey, data, 5*time.Minute)
    
    ctx.SetHeader("X-Cache", "MISS")
    return ctx.JSON(200, data)
}
```

### Conditional Responses

```go
func conditionalHandler(ctx pkg.Context) error {
    // Get resource
    resource := getResource(ctx.Param("id"))
    
    // Check If-None-Match (ETag)
    etag := fmt.Sprintf(`"%s"`, resource.Hash)
    if ctx.GetHeader("If-None-Match") == etag {
        return ctx.String(304, "")
    }
    
    // Check If-Modified-Since
    if modifiedSince := ctx.GetHeader("If-Modified-Since"); modifiedSince != "" {
        t, _ := time.Parse(http.TimeFormat, modifiedSince)
        if resource.UpdatedAt.Before(t) || resource.UpdatedAt.Equal(t) {
            return ctx.String(304, "")
        }
    }
    
    // Set caching headers
    ctx.SetHeader("ETag", etag)
    ctx.SetHeader("Last-Modified", resource.UpdatedAt.Format(http.TimeFormat))
    ctx.SetHeader("Cache-Control", "max-age=3600")
    
    return ctx.JSON(200, resource)
}
```

### Content Compression

```go
func compressedHandler(ctx pkg.Context) error {
    // Check if client accepts compression
    acceptEncoding := ctx.GetHeader("Accept-Encoding")
    
    data := generateLargeResponse()
    
    if strings.Contains(acceptEncoding, "gzip") {
        // Compress response
        var buf bytes.Buffer
        gz := gzip.NewWriter(&buf)
        json.NewEncoder(gz).Encode(data)
        gz.Close()
        
        ctx.SetHeader("Content-Encoding", "gzip")
        ctx.Response().WriteHeader(200)
        return ctx.Response().Write(buf.Bytes())
    }
    
    return ctx.JSON(200, data)
}
```

---

## Best Practices

### 1. Always Validate Input

```go
func handler(ctx pkg.Context) error {
    // Validate content type
    if ctx.GetHeader("Content-Type") != "application/json" {
        return pkg.NewValidationError("Invalid content type", "header")
    }
    
    // Validate body
    if len(ctx.Body()) == 0 {
        return pkg.NewValidationError("Empty body", "body")
    }
    
    // Parse and validate
    var input InputData
    if err := json.Unmarshal(ctx.Body(), &input); err != nil {
        return pkg.NewValidationError("Invalid JSON", "body")
    }
    
    if err := input.Validate(); err != nil {
        return err
    }
    
    return ctx.JSON(200, input)
}
```

### 2. Set Appropriate Headers

```go
func handler(ctx pkg.Context) error {
    // Security headers
    ctx.SetHeader("X-Content-Type-Options", "nosniff")
    ctx.SetHeader("X-Frame-Options", "DENY")
    ctx.SetHeader("X-XSS-Protection", "1; mode=block")
    
    // Cache control
    ctx.SetHeader("Cache-Control", "no-store, no-cache, must-revalidate")
    
    // CORS (if needed)
    ctx.SetHeader("Access-Control-Allow-Origin", "https://example.com")
    
    return ctx.JSON(200, data)
}
```

### 3. Handle Errors Gracefully

```go
func handler(ctx pkg.Context) error {
    data, err := fetchData(ctx)
    if err != nil {
        // Log error
        ctx.Logger().Error("failed to fetch data", "error", err)
        
        // Return appropriate error
        if errors.Is(err, ErrNotFound) {
            return pkg.NewNotFoundError("Resource")
        }
        
        return pkg.NewInternalError("Failed to process request")
    }
    
    return ctx.JSON(200, data)
}
```

### 4. Use Appropriate Status Codes

```go
// 200 OK - Success
return ctx.JSON(200, data)

// 201 Created - Resource created
return ctx.JSON(201, newResource)

// 204 No Content - Success with no body
return ctx.String(204, "")

// 400 Bad Request - Client error
return pkg.NewValidationError("Invalid input", "field")

// 401 Unauthorized - Authentication required
return pkg.NewAuthenticationError("Login required")

// 403 Forbidden - Insufficient permissions
return pkg.NewAuthorizationError("Access denied")

// 404 Not Found - Resource not found
return pkg.NewNotFoundError("User")

// 500 Internal Server Error - Server error
return pkg.NewInternalError("Processing failed")
```

### 5. Sanitize Output

```go
func handler(ctx pkg.Context) error {
    user := getUser(ctx.Param("id"))
    
    // Don't expose sensitive data
    response := map[string]interface{}{
        "id": user.ID,
        "username": user.Username,
        "email": user.Email,
        // Don't include: password, tokens, etc.
    }
    
    return ctx.JSON(200, response)
}
```

---

## See Also

- [Context API](context.md) - Request context interface
- [Forms and Validation API](forms-validation.md) - Form handling
- [Cookies and Headers API](cookies-headers.md) - Cookie/header management
- [Templates and Responses API](templates-responses.md) - Template rendering
- [Error Codes Reference](error-codes.md) - Error handling

---

**Last Updated**: 2025-11-29  
**Framework Version**: 1.0.0
