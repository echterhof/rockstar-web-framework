# Form Handling API Reference

## Overview

The Rockstar Web Framework provides comprehensive form handling capabilities including parsing URL-encoded and multipart form data, file uploads, and validation. The form handling system is designed to be secure, efficient, and easy to use.

## FormParser Interface

Interface for parsing form data from requests.

```go
type FormParser interface {
    ParseForm(req *Request) error
    ParseMultipartForm(req *Request, maxMemory int64) error
    GetFormValue(req *Request, key string) string
    GetFormValues(req *Request, key string) []string
    GetFormFile(req *Request, key string) (*FormFile, error)
    GetFormFiles(req *Request, key string) ([]*FormFile, error)
    GetAllFormFiles(req *Request) map[string][]*FormFile
}
```

### NewFormParser()

Creates a new form parser with default settings (32MB max memory for multipart forms).

**Signature:**
```go
func NewFormParser() FormParser
```

**Returns:**
- `FormParser` - Form parser instance

**Example:**
```go
parser := pkg.NewFormParser()
```

### ParseForm()

Parses form data from the request. Automatically detects content type and parses accordingly:
- `application/x-www-form-urlencoded` - URL-encoded form data
- `multipart/form-data` - Multipart form data with files
- Query parameters as fallback

**Signature:**
```go
ParseForm(req *Request) error
```

**Parameters:**
- `req` - Request object

**Returns:**
- `error` - Error if parsing fails

**Example:**
```go
router.POST("/submit", func(ctx pkg.Context) error {
    parser := pkg.NewFormParser()
    req := ctx.Request()
    
    if err := parser.ParseForm(req); err != nil {
        return ctx.JSON(400, map[string]string{"error": "Failed to parse form"})
    }
    
    username := parser.GetFormValue(req, "username")
    return ctx.JSON(200, map[string]string{"username": username})
})
```

### ParseMultipartForm()

Parses multipart form data with a specified max memory limit. Used for file uploads.

**Signature:**
```go
ParseMultipartForm(req *Request, maxMemory int64) error
```

**Parameters:**
- `req` - Request object
- `maxMemory` - Maximum memory in bytes (e.g., 32 << 20 for 32MB)

**Returns:**
- `error` - Error if parsing fails

**Example:**
```go
router.POST("/upload", func(ctx pkg.Context) error {
    parser := pkg.NewFormParser()
    req := ctx.Request()
    
    // Parse with 10MB limit
    maxMemory := int64(10 << 20)
    if err := parser.ParseMultipartForm(req, maxMemory); err != nil {
        return ctx.JSON(400, map[string]string{"error": "Failed to parse upload"})
    }
    
    file, _ := parser.GetFormFile(req, "file")
    return ctx.JSON(200, map[string]interface{}{
        "filename": file.Filename,
        "size":     file.Size,
    })
})
```

### GetFormValue()

Gets a single form value by key.

**Signature:**
```go
GetFormValue(req *Request, key string) string
```

**Parameters:**
- `req` - Request object
- `key` - Form field name

**Returns:**
- `string` - Form value, or empty string if not found

**Example:**
```go
parser := pkg.NewFormParser()
parser.ParseForm(req)

username := parser.GetFormValue(req, "username")
email := parser.GetFormValue(req, "email")
```

### GetFormValues()

Gets all values for a key (returns as slice).

**Signature:**
```go
GetFormValues(req *Request, key string) []string
```

**Parameters:**
- `req` - Request object
- `key` - Form field name

**Returns:**
- `[]string` - Array of values, or nil if not found

**Example:**
```go
parser := pkg.NewFormParser()
parser.ParseForm(req)

tags := parser.GetFormValues(req, "tags")
for _, tag := range tags {
    fmt.Println(tag)
}
```

### GetFormFile()

Gets an uploaded file by key.

**Signature:**
```go
GetFormFile(req *Request, key string) (*FormFile, error)
```

**Parameters:**
- `req` - Request object
- `key` - Form field name

**Returns:**
- `*FormFile` - Uploaded file object
- `error` - ErrFileRequired if file not found, ErrFormNotParsed if form not parsed

**Example:**
```go
parser := pkg.NewFormParser()
parser.ParseMultipartForm(req, 32<<20)

file, err := parser.GetFormFile(req, "avatar")
if err != nil {
    return ctx.JSON(400, map[string]string{"error": "No file uploaded"})
}

fmt.Printf("Uploaded: %s (%d bytes)\n", file.Filename, file.Size)
```

### GetFormFiles()

Gets all uploaded files for a key (returns as slice).

**Signature:**
```go
GetFormFiles(req *Request, key string) ([]*FormFile, error)
```

**Parameters:**
- `req` - Request object
- `key` - Form field name

**Returns:**
- `[]*FormFile` - Array of uploaded files
- `error` - Error if files not found

**Example:**
```go
parser := pkg.NewFormParser()
parser.ParseMultipartForm(req, 32<<20)

files, err := parser.GetFormFiles(req, "attachments")
if err != nil {
    return ctx.JSON(400, map[string]string{"error": "No files uploaded"})
}

for _, file := range files {
    fmt.Printf("File: %s\n", file.Filename)
}
```

### GetAllFormFiles()

Gets all uploaded files from the request.

**Signature:**
```go
GetAllFormFiles(req *Request) map[string][]*FormFile
```

**Parameters:**
- `req` - Request object

**Returns:**
- `map[string][]*FormFile` - Map of field names to file arrays

**Example:**
```go
parser := pkg.NewFormParser()
parser.ParseMultipartForm(req, 32<<20)

allFiles := parser.GetAllFormFiles(req)
for fieldName, files := range allFiles {
    fmt.Printf("Field %s has %d file(s)\n", fieldName, len(files))
}
```

## FormValidator Interface

Interface for validating form data.

```go
type FormValidator interface {
    ValidateRequired(req *Request, fields []string) error
    ValidateRequiredFiles(req *Request, fields []string) error
    ValidateFileSize(req *Request, field string, maxSize int64) error
    ValidateFileType(req *Request, field string, allowedTypes []string) error
    ValidateField(req *Request, field string, validator func(string) error) error
    ValidateAll(req *Request, rules FormValidationRules) error
}
```

### NewFormValidator()

Creates a new form validator.

**Signature:**
```go
func NewFormValidator(parser FormParser) FormValidator
```

**Parameters:**
- `parser` - Form parser instance

**Returns:**
- `FormValidator` - Form validator instance

**Example:**
```go
parser := pkg.NewFormParser()
validator := pkg.NewFormValidator(parser)
```

### ValidateRequired()

Checks if required fields are present and non-empty.

**Signature:**
```go
ValidateRequired(req *Request, fields []string) error
```

**Parameters:**
- `req` - Request object
- `fields` - Array of required field names

**Returns:**
- `error` - ErrFieldRequired if any field is missing

**Example:**
```go
parser := pkg.NewFormParser()
validator := pkg.NewFormValidator(parser)

parser.ParseForm(req)

err := validator.ValidateRequired(req, []string{"username", "email", "password"})
if err != nil {
    return ctx.JSON(400, map[string]string{"error": err.Error()})
}
```

### ValidateRequiredFiles()

Checks if required files are present.

**Signature:**
```go
ValidateRequiredFiles(req *Request, fields []string) error
```

**Parameters:**
- `req` - Request object
- `fields` - Array of required file field names

**Returns:**
- `error` - ErrFileRequired if any file is missing

**Example:**
```go
parser := pkg.NewFormParser()
validator := pkg.NewFormValidator(parser)

parser.ParseMultipartForm(req, 32<<20)

err := validator.ValidateRequiredFiles(req, []string{"avatar", "document"})
if err != nil {
    return ctx.JSON(400, map[string]string{"error": "Missing required files"})
}
```

### ValidateFileSize()

Checks if a file is within size limits.

**Signature:**
```go
ValidateFileSize(req *Request, field string, maxSize int64) error
```

**Parameters:**
- `req` - Request object
- `field` - File field name
- `maxSize` - Maximum size in bytes

**Returns:**
- `error` - ErrFileTooLarge if file exceeds limit

**Example:**
```go
parser := pkg.NewFormParser()
validator := pkg.NewFormValidator(parser)

parser.ParseMultipartForm(req, 32<<20)

// Limit to 5MB
maxSize := int64(5 << 20)
err := validator.ValidateFileSize(req, "avatar", maxSize)
if err != nil {
    return ctx.JSON(400, map[string]string{"error": "File too large"})
}
```

### ValidateFileType()

Checks if a file has an allowed content type.

**Signature:**
```go
ValidateFileType(req *Request, field string, allowedTypes []string) error
```

**Parameters:**
- `req` - Request object
- `field` - File field name
- `allowedTypes` - Array of allowed MIME type prefixes

**Returns:**
- `error` - ErrInvalidFileType if file type not allowed

**Example:**
```go
parser := pkg.NewFormParser()
validator := pkg.NewFormValidator(parser)

parser.ParseMultipartForm(req, 32<<20)

// Only allow images
allowedTypes := []string{"image/jpeg", "image/png", "image/gif"}
err := validator.ValidateFileType(req, "avatar", allowedTypes)
if err != nil {
    return ctx.JSON(400, map[string]string{"error": "Invalid file type"})
}
```

### ValidateField()

Validates a field with a custom validator function.

**Signature:**
```go
ValidateField(req *Request, field string, validator func(string) error) error
```

**Parameters:**
- `req` - Request object
- `field` - Field name
- `validator` - Custom validation function

**Returns:**
- `error` - Error from validator function

**Example:**
```go
parser := pkg.NewFormParser()
validator := pkg.NewFormValidator(parser)

parser.ParseForm(req)

// Custom email validator
err := validator.ValidateField(req, "email", func(value string) error {
    if !strings.Contains(value, "@") {
        return errors.New("invalid email format")
    }
    return nil
})
```

### ValidateAll()

Validates all rules at once.

**Signature:**
```go
ValidateAll(req *Request, rules FormValidationRules) error
```

**Parameters:**
- `req` - Request object
- `rules` - Validation rules

**Returns:**
- `error` - First validation error encountered

**Example:**
```go
parser := pkg.NewFormParser()
validator := pkg.NewFormValidator(parser)

parser.ParseMultipartForm(req, 32<<20)

rules := pkg.FormValidationRules{
    RequiredFields: []string{"username", "email"},
    RequiredFiles:  []string{"avatar"},
    FileSizeLimits: map[string]int64{
        "avatar": 5 << 20, // 5MB
    },
    FileTypes: map[string][]string{
        "avatar": {"image/jpeg", "image/png"},
    },
}

err := validator.ValidateAll(req, rules)
if err != nil {
    return ctx.JSON(400, map[string]string{"error": err.Error()})
}
```

## FormValidationRules

Defines validation rules for form data.

```go
type FormValidationRules struct {
    RequiredFields []string
    RequiredFiles  []string
    FileSizeLimits map[string]int64
    FileTypes      map[string][]string
    CustomRules    map[string]func(string) error
}
```

**Fields:**
- `RequiredFields` - Array of required field names
- `RequiredFiles` - Array of required file field names
- `FileSizeLimits` - Map of field names to max size in bytes
- `FileTypes` - Map of field names to allowed MIME types
- `CustomRules` - Map of field names to custom validator functions

**Example:**
```go
rules := pkg.FormValidationRules{
    RequiredFields: []string{"username", "email", "password"},
    RequiredFiles:  []string{"avatar"},
    FileSizeLimits: map[string]int64{
        "avatar":   5 << 20,  // 5MB
        "document": 10 << 20, // 10MB
    },
    FileTypes: map[string][]string{
        "avatar":   {"image/jpeg", "image/png", "image/gif"},
        "document": {"application/pdf", "application/msword"},
    },
    CustomRules: map[string]func(string) error{
        "email": func(value string) error {
            if !strings.Contains(value, "@") {
                return errors.New("invalid email")
            }
            return nil
        },
        "age": func(value string) error {
            age, err := strconv.Atoi(value)
            if err != nil || age < 18 {
                return errors.New("must be 18 or older")
            }
            return nil
        },
    },
}
```

## Pipeline Functions

### FormDataPipeline()

Creates a pipeline for form data validation.

**Signature:**
```go
func FormDataPipeline(rules FormValidationRules) PipelineFunc
```

**Parameters:**
- `rules` - Validation rules

**Returns:**
- `PipelineFunc` - Pipeline function

**Example:**
```go
rules := pkg.FormValidationRules{
    RequiredFields: []string{"username", "email"},
}

pipeline := pkg.NewPipelineEngine()
pipeline.Register("form-validation", pkg.FormDataPipeline(rules))

router.POST("/register", func(ctx pkg.Context) error {
    result, err := pipeline.Execute(ctx, "form-validation")
    if err != nil {
        return ctx.JSON(400, map[string]string{"error": err.Error()})
    }
    
    // Form is valid, process registration
    return ctx.JSON(200, map[string]string{"status": "registered"})
})
```

### FileUploadPipeline()

Creates a pipeline for file upload validation.

**Signature:**
```go
func FileUploadPipeline(maxMemory int64, rules FormValidationRules) PipelineFunc
```

**Parameters:**
- `maxMemory` - Maximum memory for multipart parsing
- `rules` - Validation rules

**Returns:**
- `PipelineFunc` - Pipeline function

**Example:**
```go
rules := pkg.FormValidationRules{
    RequiredFiles: []string{"avatar"},
    FileSizeLimits: map[string]int64{
        "avatar": 5 << 20, // 5MB
    },
    FileTypes: map[string][]string{
        "avatar": {"image/jpeg", "image/png"},
    },
}

pipeline := pkg.NewPipelineEngine()
pipeline.Register("file-upload", pkg.FileUploadPipeline(32<<20, rules))

router.POST("/upload", func(ctx pkg.Context) error {
    result, err := pipeline.Execute(ctx, "file-upload")
    if err != nil {
        return ctx.JSON(400, map[string]string{"error": err.Error()})
    }
    
    // File is valid, save it
    return ctx.JSON(200, map[string]string{"status": "uploaded"})
})
```

## Error Variables

```go
var (
    ErrFormNotParsed   = errors.New("form data not parsed")
    ErrFieldRequired   = errors.New("required field missing")
    ErrFieldInvalid    = errors.New("field value invalid")
    ErrFileRequired    = errors.New("required file missing")
    ErrFileTooLarge    = errors.New("file too large")
    ErrInvalidFileType = errors.New("invalid file type")
)
```

## Complete Examples

### Simple Form Submission

```go
router.POST("/contact", func(ctx pkg.Context) error {
    parser := pkg.NewFormParser()
    validator := pkg.NewFormValidator(parser)
    req := ctx.Request()
    
    // Parse form
    if err := parser.ParseForm(req); err != nil {
        return ctx.JSON(400, map[string]string{"error": "Invalid form data"})
    }
    
    // Validate required fields
    err := validator.ValidateRequired(req, []string{"name", "email", "message"})
    if err != nil {
        return ctx.JSON(400, map[string]string{"error": err.Error()})
    }
    
    // Get values
    name := parser.GetFormValue(req, "name")
    email := parser.GetFormValue(req, "email")
    message := parser.GetFormValue(req, "message")
    
    // Process contact form
    // ...
    
    return ctx.JSON(200, map[string]string{
        "status": "Message sent",
        "name":   name,
    })
})
```

### File Upload with Validation

```go
router.POST("/profile/avatar", func(ctx pkg.Context) error {
    parser := pkg.NewFormParser()
    validator := pkg.NewFormValidator(parser)
    req := ctx.Request()
    
    // Parse multipart form (10MB limit)
    maxMemory := int64(10 << 20)
    if err := parser.ParseMultipartForm(req, maxMemory); err != nil {
        return ctx.JSON(400, map[string]string{"error": "Failed to parse upload"})
    }
    
    // Validate file exists
    err := validator.ValidateRequiredFiles(req, []string{"avatar"})
    if err != nil {
        return ctx.JSON(400, map[string]string{"error": "Avatar is required"})
    }
    
    // Validate file size (5MB max)
    err = validator.ValidateFileSize(req, "avatar", 5<<20)
    if err != nil {
        return ctx.JSON(400, map[string]string{"error": "Avatar too large (max 5MB)"})
    }
    
    // Validate file type
    allowedTypes := []string{"image/jpeg", "image/png", "image/gif"}
    err = validator.ValidateFileType(req, "avatar", allowedTypes)
    if err != nil {
        return ctx.JSON(400, map[string]string{"error": "Invalid file type"})
    }
    
    // Get file
    file, _ := parser.GetFormFile(req, "avatar")
    
    // Save file
    files := ctx.Files()
    destPath := fmt.Sprintf("/uploads/avatars/%s", file.Filename)
    err = files.SaveUploadedFile(ctx, file.Filename, destPath)
    if err != nil {
        return ctx.JSON(500, map[string]string{"error": "Failed to save file"})
    }
    
    return ctx.JSON(200, map[string]interface{}{
        "status":   "uploaded",
        "filename": file.Filename,
        "size":     file.Size,
        "path":     destPath,
    })
})
```

### Complete Form with All Validations

```go
router.POST("/register", func(ctx pkg.Context) error {
    parser := pkg.NewFormParser()
    validator := pkg.NewFormValidator(parser)
    req := ctx.Request()
    
    // Parse multipart form
    if err := parser.ParseMultipartForm(req, 32<<20); err != nil {
        return ctx.JSON(400, map[string]string{"error": "Invalid form data"})
    }
    
    // Define validation rules
    rules := pkg.FormValidationRules{
        RequiredFields: []string{"username", "email", "password"},
        RequiredFiles:  []string{"avatar"},
        FileSizeLimits: map[string]int64{
            "avatar": 5 << 20, // 5MB
        },
        FileTypes: map[string][]string{
            "avatar": {"image/jpeg", "image/png"},
        },
        CustomRules: map[string]func(string) error{
            "email": func(value string) error {
                if !strings.Contains(value, "@") {
                    return errors.New("invalid email format")
                }
                return nil
            },
            "password": func(value string) error {
                if len(value) < 8 {
                    return errors.New("password must be at least 8 characters")
                }
                return nil
            },
        },
    }
    
    // Validate all rules
    if err := validator.ValidateAll(req, rules); err != nil {
        return ctx.JSON(400, map[string]string{"error": err.Error()})
    }
    
    // Get form values
    username := parser.GetFormValue(req, "username")
    email := parser.GetFormValue(req, "email")
    password := parser.GetFormValue(req, "password")
    
    // Get uploaded file
    avatar, _ := parser.GetFormFile(req, "avatar")
    
    // Process registration
    // ...
    
    return ctx.JSON(201, map[string]interface{}{
        "status":   "registered",
        "username": username,
        "email":    email,
        "avatar":   avatar.Filename,
    })
})
```

### Using Context Helper Methods

```go
router.POST("/submit", func(ctx pkg.Context) error {
    // Context provides convenient form methods
    username := ctx.FormValue("username")
    email := ctx.FormValue("email")
    
    if username == "" || email == "" {
        return ctx.JSON(400, map[string]string{
            "error": "Username and email are required",
        })
    }
    
    // Get uploaded file
    file, err := ctx.FormFile("document")
    if err != nil {
        return ctx.JSON(400, map[string]string{
            "error": "Document is required",
        })
    }
    
    // Validate file size
    maxSize := int64(10 << 20) // 10MB
    if file.Size > maxSize {
        return ctx.JSON(400, map[string]string{
            "error": "File too large",
        })
    }
    
    return ctx.JSON(200, map[string]interface{}{
        "username": username,
        "email":    email,
        "file":     file.Filename,
    })
})
```

## Best Practices

1. **Always Parse Before Access:** Call `ParseForm()` or `ParseMultipartForm()` before accessing form data

2. **Set Memory Limits:** Use appropriate memory limits for multipart forms to prevent DoS attacks

3. **Validate Early:** Validate form data before processing to fail fast

4. **Use Validation Rules:** Use `FormValidationRules` for complex validation scenarios

5. **Check File Sizes:** Always validate file sizes before processing uploads

6. **Validate File Types:** Check MIME types to prevent malicious file uploads

7. **Sanitize Input:** Sanitize form values before using them in queries or output

8. **Use Custom Validators:** Implement custom validation logic for business rules

9. **Handle Errors Gracefully:** Return clear error messages for validation failures

10. **Use Pipelines:** Use pipeline functions for reusable validation logic

## Security Considerations

1. **File Upload Limits:** Always set reasonable file size limits
2. **MIME Type Validation:** Validate file types by MIME type, not just extension
3. **Path Traversal:** Validate filenames to prevent directory traversal attacks
4. **Memory Limits:** Set appropriate memory limits for multipart parsing
5. **Input Sanitization:** Sanitize all form input before use
6. **CSRF Protection:** Use CSRF tokens for form submissions
7. **Rate Limiting:** Implement rate limiting on form endpoints
8. **Content Type Validation:** Verify Content-Type header matches expected format

## See Also

- [Context API](context.md) - Context form methods
- [Security Guide](../guides/security.md) - Security best practices
- [File Manager API](filesystem.md) - File operations
- [Pipeline API](pipeline.md) - Pipeline functions
- [Validation Guide](../guides/validation.md) - Input validation patterns
