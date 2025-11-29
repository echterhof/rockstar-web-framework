# Forms and Validation API

Complete API reference for form parsing and validation in the Rockstar Web Framework.

## Overview

The framework provides comprehensive form handling with support for:
- URL-encoded forms
- Multipart forms with file uploads
- Field validation with custom rules
- File validation (size, type, required)
- Automatic error collection

## FormParser Interface

Handles parsing of form data from HTTP requests.

```go
type FormParser interface {
    ParseForm(ctx Context) error
    ParseMultipartForm(ctx Context, maxMemory int64) error
    GetFormValue(ctx Context, key string) string
    GetFormValues(ctx Context, key string) []string
    GetFormFile(ctx Context, key string) (*FormFile, error)
    GetFormFiles(ctx Context, key string) ([]*FormFile, error)
    HasFormValue(ctx Context, key string) bool
}
```

### Methods

#### ParseForm

```go
ParseForm(ctx Context) error
```

Parses URL-encoded form data from the request body.

**Parameters:**
- `ctx`: Request context

**Returns:**
- `error`: Error if parsing fails

**Example:**
```go
func handler(ctx pkg.Context) error {
    parser := pkg.NewFormParser()
    if err := parser.ParseForm(ctx); err != nil {
        return err
    }
    
    name := parser.GetFormValue(ctx, "name")
    email := parser.GetFormValue(ctx, "email")
    
    return ctx.JSON(200, map[string]string{
        "name": name,
        "email": email,
    })
}
```

#### ParseMultipartForm

```go
ParseMultipartForm(ctx Context, maxMemory int64) error
```

Parses multipart form data, including file uploads.

**Parameters:**
- `ctx`: Request context
- `maxMemory`: Maximum memory in bytes to use for parsing (remaining data is stored in temp files)

**Returns:**
- `error`: Error if parsing fails

**Example:**
```go
func uploadHandler(ctx pkg.Context) error {
    parser := pkg.NewFormParser()
    
    // Parse with 10MB max memory
    if err := parser.ParseMultipartForm(ctx, 10<<20); err != nil {
        return err
    }
    
    file, err := parser.GetFormFile(ctx, "upload")
    if err != nil {
        return err
    }
    
    // Process file...
    return ctx.JSON(200, map[string]string{
        "filename": file.Filename,
        "size": fmt.Sprintf("%d", file.Size),
    })
}
```

#### GetFormValue

```go
GetFormValue(ctx Context, key string) string
```

Retrieves a single form value by key.

**Parameters:**
- `ctx`: Request context
- `key`: Form field name

**Returns:**
- `string`: Form value (empty string if not found)

#### GetFormValues

```go
GetFormValues(ctx Context, key string) []string
```

Retrieves all values for a form key (useful for checkboxes, multi-select).

**Parameters:**
- `ctx`: Request context
- `key`: Form field name

**Returns:**
- `[]string`: All values for the key

**Example:**
```go
// For <input type="checkbox" name="interests" value="sports">
// <input type="checkbox" name="interests" value="music">
interests := parser.GetFormValues(ctx, "interests")
// interests = ["sports", "music"]
```

#### GetFormFile

```go
GetFormFile(ctx Context, key string) (*FormFile, error)
```

Retrieves a single uploaded file.

**Parameters:**
- `ctx`: Request context
- `key`: Form field name

**Returns:**
- `*FormFile`: File information and content
- `error`: Error if file not found or invalid

#### GetFormFiles

```go
GetFormFiles(ctx Context, key string) ([]*FormFile, error)
```

Retrieves multiple uploaded files for a single field.

**Parameters:**
- `ctx`: Request context
- `key`: Form field name

**Returns:**
- `[]*FormFile`: Array of uploaded files
- `error`: Error if files not found or invalid

#### HasFormValue

```go
HasFormValue(ctx Context, key string) bool
```

Checks if a form field exists.

**Parameters:**
- `ctx`: Request context
- `key`: Form field name

**Returns:**
- `bool`: True if field exists

## FormValidator Interface

Validates form data with built-in and custom rules.

```go
type FormValidator interface {
    ValidateRequired(ctx Context, fields []string) map[string]string
    ValidateRequiredFiles(ctx Context, fields []string) map[string]string
    ValidateFileSize(file *FormFile, maxSize int64) error
    ValidateFileType(file *FormFile, allowedTypes []string) error
    ValidateField(value string, rules []ValidationRule) error
    ValidateForm(ctx Context, rules map[string][]ValidationRule) map[string]string
}
```

### Methods

#### ValidateRequired

```go
ValidateRequired(ctx Context, fields []string) map[string]string
```

Validates that required fields are present and non-empty.

**Parameters:**
- `ctx`: Request context
- `fields`: List of required field names

**Returns:**
- `map[string]string`: Map of field names to error messages (empty if valid)

**Example:**
```go
validator := pkg.NewFormValidator()
errors := validator.ValidateRequired(ctx, []string{"name", "email", "password"})

if len(errors) > 0 {
    return ctx.JSON(400, map[string]interface{}{
        "errors": errors,
    })
}
```

#### ValidateRequiredFiles

```go
ValidateRequiredFiles(ctx Context, fields []string) map[string]string
```

Validates that required file uploads are present.

**Parameters:**
- `ctx`: Request context
- `fields`: List of required file field names

**Returns:**
- `map[string]string`: Map of field names to error messages

#### ValidateFileSize

```go
ValidateFileSize(file *FormFile, maxSize int64) error
```

Validates that an uploaded file doesn't exceed maximum size.

**Parameters:**
- `file`: Uploaded file
- `maxSize`: Maximum size in bytes

**Returns:**
- `error`: Error if file exceeds size limit

**Example:**
```go
file, _ := parser.GetFormFile(ctx, "avatar")

// Max 5MB
if err := validator.ValidateFileSize(file, 5<<20); err != nil {
    return ctx.JSON(400, map[string]string{
        "error": "File too large (max 5MB)",
    })
}
```

#### ValidateFileType

```go
ValidateFileType(file *FormFile, allowedTypes []string) error
```

Validates that an uploaded file has an allowed MIME type.

**Parameters:**
- `file`: Uploaded file
- `allowedTypes`: List of allowed MIME types

**Returns:**
- `error`: Error if file type not allowed

**Example:**
```go
file, _ := parser.GetFormFile(ctx, "document")

allowedTypes := []string{"application/pdf", "image/jpeg", "image/png"}
if err := validator.ValidateFileType(file, allowedTypes); err != nil {
    return ctx.JSON(400, map[string]string{
        "error": "Invalid file type",
    })
}
```

#### ValidateField

```go
ValidateField(value string, rules []ValidationRule) error
```

Validates a single field value against multiple rules.

**Parameters:**
- `value`: Field value to validate
- `rules`: Array of validation rules

**Returns:**
- `error`: First validation error encountered

**Example:**
```go
email := parser.GetFormValue(ctx, "email")

rules := []pkg.ValidationRule{
    pkg.RequiredRule{},
    pkg.EmailRule{},
    pkg.MaxLengthRule{Max: 255},
}

if err := validator.ValidateField(email, rules); err != nil {
    return ctx.JSON(400, map[string]string{
        "error": err.Error(),
    })
}
```

#### ValidateForm

```go
ValidateForm(ctx Context, rules map[string][]ValidationRule) map[string]string
```

Validates multiple form fields at once.

**Parameters:**
- `ctx`: Request context
- `rules`: Map of field names to validation rules

**Returns:**
- `map[string]string`: Map of field names to error messages

**Example:**
```go
rules := map[string][]pkg.ValidationRule{
    "name": {
        pkg.RequiredRule{},
        pkg.MinLengthRule{Min: 2},
        pkg.MaxLengthRule{Max: 100},
    },
    "email": {
        pkg.RequiredRule{},
        pkg.EmailRule{},
    },
    "age": {
        pkg.RequiredRule{},
        pkg.IntegerRule{},
        pkg.MinValueRule{Min: 18},
    },
}

errors := validator.ValidateForm(ctx, rules)
if len(errors) > 0 {
    return ctx.JSON(400, map[string]interface{}{
        "errors": errors,
    })
}
```

## Types

### FormFile

Represents an uploaded file.

```go
type FormFile struct {
    Filename    string      // Original filename
    Size        int64       // File size in bytes
    ContentType string      // MIME type
    Header      textproto.MIMEHeader // File headers
    Content     []byte      // File content
}
```

**Methods:**

```go
// Save file to disk
func (f *FormFile) Save(path string) error

// Get file reader
func (f *FormFile) Open() (io.ReadCloser, error)
```

### ValidationRule

Interface for custom validation rules.

```go
type ValidationRule interface {
    Validate(value string) error
    Message() string
}
```

## Built-in Validation Rules

### RequiredRule

Validates that a field is not empty.

```go
pkg.RequiredRule{}
```

### EmailRule

Validates email format.

```go
pkg.EmailRule{}
```

### MinLengthRule

Validates minimum string length.

```go
pkg.MinLengthRule{Min: 5}
```

### MaxLengthRule

Validates maximum string length.

```go
pkg.MaxLengthRule{Max: 100}
```

### IntegerRule

Validates that value is an integer.

```go
pkg.IntegerRule{}
```

### MinValueRule

Validates minimum numeric value.

```go
pkg.MinValueRule{Min: 0}
```

### MaxValueRule

Validates maximum numeric value.

```go
pkg.MaxValueRule{Max: 100}
```

### RegexRule

Validates against a regular expression.

```go
pkg.RegexRule{Pattern: `^[A-Z][a-z]+$`}
```

### URLRule

Validates URL format.

```go
pkg.URLRule{}
```

### AlphanumericRule

Validates alphanumeric characters only.

```go
pkg.AlphanumericRule{}
```

## Complete Example

```go
package main

import (
    "github.com/echterhof/rockstar-web-framework/pkg"
)

func registrationHandler(ctx pkg.Context) error {
    parser := pkg.NewFormParser()
    validator := pkg.NewFormValidator()
    
    // Parse form
    if err := parser.ParseMultipartForm(ctx, 10<<20); err != nil {
        return ctx.JSON(400, map[string]string{
            "error": "Invalid form data",
        })
    }
    
    // Validate fields
    rules := map[string][]pkg.ValidationRule{
        "username": {
            pkg.RequiredRule{},
            pkg.MinLengthRule{Min: 3},
            pkg.MaxLengthRule{Max: 20},
            pkg.AlphanumericRule{},
        },
        "email": {
            pkg.RequiredRule{},
            pkg.EmailRule{},
        },
        "password": {
            pkg.RequiredRule{},
            pkg.MinLengthRule{Min: 8},
        },
    }
    
    errors := validator.ValidateForm(ctx, rules)
    if len(errors) > 0 {
        return ctx.JSON(400, map[string]interface{}{
            "errors": errors,
        })
    }
    
    // Validate avatar upload (optional)
    if parser.HasFormValue(ctx, "avatar") {
        avatar, err := parser.GetFormFile(ctx, "avatar")
        if err != nil {
            return ctx.JSON(400, map[string]string{
                "error": "Invalid avatar file",
            })
        }
        
        // Validate file size (max 2MB)
        if err := validator.ValidateFileSize(avatar, 2<<20); err != nil {
            return ctx.JSON(400, map[string]string{
                "error": "Avatar too large (max 2MB)",
            })
        }
        
        // Validate file type
        allowedTypes := []string{"image/jpeg", "image/png", "image/gif"}
        if err := validator.ValidateFileType(avatar, allowedTypes); err != nil {
            return ctx.JSON(400, map[string]string{
                "error": "Invalid avatar type (JPEG, PNG, or GIF only)",
            })
        }
        
        // Save avatar
        if err := avatar.Save("uploads/" + avatar.Filename); err != nil {
            return ctx.JSON(500, map[string]string{
                "error": "Failed to save avatar",
            })
        }
    }
    
    // Create user...
    return ctx.JSON(201, map[string]string{
        "message": "Registration successful",
    })
}
```

## Custom Validation Rules

Create custom validation rules by implementing the `ValidationRule` interface:

```go
type PasswordStrengthRule struct {
    RequireUppercase bool
    RequireLowercase bool
    RequireDigit     bool
    RequireSpecial   bool
}

func (r PasswordStrengthRule) Validate(value string) error {
    if r.RequireUppercase && !strings.ContainsAny(value, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
        return fmt.Errorf("password must contain uppercase letter")
    }
    if r.RequireLowercase && !strings.ContainsAny(value, "abcdefghijklmnopqrstuvwxyz") {
        return fmt.Errorf("password must contain lowercase letter")
    }
    if r.RequireDigit && !strings.ContainsAny(value, "0123456789") {
        return fmt.Errorf("password must contain digit")
    }
    if r.RequireSpecial && !strings.ContainsAny(value, "!@#$%^&*") {
        return fmt.Errorf("password must contain special character")
    }
    return nil
}

func (r PasswordStrengthRule) Message() string {
    return "password does not meet strength requirements"
}

// Usage
rules := []pkg.ValidationRule{
    pkg.RequiredRule{},
    pkg.MinLengthRule{Min: 8},
    PasswordStrengthRule{
        RequireUppercase: true,
        RequireLowercase: true,
        RequireDigit:     true,
        RequireSpecial:   true,
    },
}
```

## Error Handling

All validation methods return errors or error maps that can be directly returned to clients:

```go
errors := validator.ValidateForm(ctx, rules)
if len(errors) > 0 {
    return ctx.JSON(400, map[string]interface{}{
        "success": false,
        "errors": errors,
    })
}
```

## Best Practices

1. **Always validate on the server**: Never trust client-side validation alone
2. **Validate file uploads**: Always check size and type for security
3. **Use appropriate max memory**: Set `maxMemory` based on expected upload sizes
4. **Sanitize input**: Validate and sanitize all user input
5. **Provide clear error messages**: Help users understand what went wrong
6. **Use custom rules**: Create reusable validation rules for business logic
7. **Validate early**: Validate input before processing to fail fast

## See Also

- [Context API](context.md) - Request context interface
- [Security Guide](../guides/security.md) - Input validation and sanitization
- [Router API](router.md) - Route handlers
