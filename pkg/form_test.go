package pkg

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

// TestFormParser_ParseURLEncodedForm tests parsing URL-encoded form data
func TestFormParser_ParseURLEncodedForm(t *testing.T) {
	parser := NewFormParser()

	// Create request with URL-encoded form data
	body := "name=John&email=john@example.com&age=30"
	req := &Request{
		Method:  "POST",
		Header:  http.Header{"Content-Type": []string{"application/x-www-form-urlencoded"}},
		RawBody: []byte(body),
	}

	// Parse form
	err := parser.ParseForm(req)
	if err != nil {
		t.Fatalf("ParseForm failed: %v", err)
	}

	// Verify form data
	if req.Form["name"] != "John" {
		t.Errorf("Expected name=John, got %s", req.Form["name"])
	}
	if req.Form["email"] != "john@example.com" {
		t.Errorf("Expected email=john@example.com, got %s", req.Form["email"])
	}
	if req.Form["age"] != "30" {
		t.Errorf("Expected age=30, got %s", req.Form["age"])
	}
}

// TestFormParser_ParseMultipartForm tests parsing multipart form data
func TestFormParser_ParseMultipartForm(t *testing.T) {
	parser := NewFormParser()

	// Create multipart form data
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add text fields
	writer.WriteField("name", "Jane")
	writer.WriteField("email", "jane@example.com")

	// Add file field
	fileWriter, err := writer.CreateFormFile("avatar", "avatar.jpg")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
	fileWriter.Write([]byte("fake image content"))

	writer.Close()

	// Create request
	req := &Request{
		Method:  "POST",
		Header:  http.Header{"Content-Type": []string{writer.FormDataContentType()}},
		RawBody: buf.Bytes(),
	}

	// Parse form
	err = parser.ParseMultipartForm(req, 32<<20)
	if err != nil {
		t.Fatalf("ParseMultipartForm failed: %v", err)
	}

	// Verify form data
	if req.Form["name"] != "Jane" {
		t.Errorf("Expected name=Jane, got %s", req.Form["name"])
	}
	if req.Form["email"] != "jane@example.com" {
		t.Errorf("Expected email=jane@example.com, got %s", req.Form["email"])
	}

	// Verify file
	file, exists := req.Files["avatar"]
	if !exists {
		t.Fatal("Expected avatar file to exist")
	}
	if file.Filename != "avatar.jpg" {
		t.Errorf("Expected filename=avatar.jpg, got %s", file.Filename)
	}
	if string(file.Content) != "fake image content" {
		t.Errorf("Expected file content='fake image content', got %s", string(file.Content))
	}
}

// TestFormParser_GetFormValue tests getting form values
func TestFormParser_GetFormValue(t *testing.T) {
	parser := NewFormParser()

	req := &Request{
		Form: map[string]string{
			"username": "testuser",
			"password": "secret",
		},
	}

	// Test existing value
	value := parser.GetFormValue(req, "username")
	if value != "testuser" {
		t.Errorf("Expected username=testuser, got %s", value)
	}

	// Test non-existing value
	value = parser.GetFormValue(req, "nonexistent")
	if value != "" {
		t.Errorf("Expected empty string for non-existent field, got %s", value)
	}
}

// TestFormParser_GetFormFile tests getting uploaded files
func TestFormParser_GetFormFile(t *testing.T) {
	parser := NewFormParser()

	req := &Request{
		Files: map[string]*FormFile{
			"document": {
				Filename: "test.pdf",
				Content:  []byte("PDF content"),
				Size:     11,
			},
		},
	}

	// Test existing file
	file, err := parser.GetFormFile(req, "document")
	if err != nil {
		t.Fatalf("GetFormFile failed: %v", err)
	}
	if file.Filename != "test.pdf" {
		t.Errorf("Expected filename=test.pdf, got %s", file.Filename)
	}

	// Test non-existing file
	_, err = parser.GetFormFile(req, "nonexistent")
	if err != ErrFileRequired {
		t.Errorf("Expected ErrFileRequired, got %v", err)
	}
}

// TestFormValidator_ValidateRequired tests required field validation
func TestFormValidator_ValidateRequired(t *testing.T) {
	parser := NewFormParser()
	validator := NewFormValidator(parser)

	tests := []struct {
		name           string
		form           map[string]string
		requiredFields []string
		expectError    bool
	}{
		{
			name: "all required fields present",
			form: map[string]string{
				"name":  "John",
				"email": "john@example.com",
			},
			requiredFields: []string{"name", "email"},
			expectError:    false,
		},
		{
			name: "missing required field",
			form: map[string]string{
				"name": "John",
			},
			requiredFields: []string{"name", "email"},
			expectError:    true,
		},
		{
			name: "empty required field",
			form: map[string]string{
				"name":  "John",
				"email": "",
			},
			requiredFields: []string{"name", "email"},
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &Request{Form: tt.form}
			err := validator.ValidateRequired(req, tt.requiredFields)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

// TestFormValidator_ValidateRequiredFiles tests required file validation
func TestFormValidator_ValidateRequiredFiles(t *testing.T) {
	parser := NewFormParser()
	validator := NewFormValidator(parser)

	tests := []struct {
		name          string
		files         map[string]*FormFile
		requiredFiles []string
		expectError   bool
	}{
		{
			name: "all required files present",
			files: map[string]*FormFile{
				"avatar": {Filename: "avatar.jpg"},
				"resume": {Filename: "resume.pdf"},
			},
			requiredFiles: []string{"avatar", "resume"},
			expectError:   false,
		},
		{
			name: "missing required file",
			files: map[string]*FormFile{
				"avatar": {Filename: "avatar.jpg"},
			},
			requiredFiles: []string{"avatar", "resume"},
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &Request{Files: tt.files}
			err := validator.ValidateRequiredFiles(req, tt.requiredFiles)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

// TestFormValidator_ValidateFileSize tests file size validation
func TestFormValidator_ValidateFileSize(t *testing.T) {
	parser := NewFormParser()
	validator := NewFormValidator(parser)

	tests := []struct {
		name        string
		file        *FormFile
		maxSize     int64
		expectError bool
	}{
		{
			name: "file within size limit",
			file: &FormFile{
				Filename: "small.txt",
				Content:  []byte("small content"),
				Size:     13,
			},
			maxSize:     100,
			expectError: false,
		},
		{
			name: "file exceeds size limit",
			file: &FormFile{
				Filename: "large.txt",
				Content:  make([]byte, 1000),
				Size:     1000,
			},
			maxSize:     100,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &Request{
				Files: map[string]*FormFile{
					"document": tt.file,
				},
			}
			err := validator.ValidateFileSize(req, "document", tt.maxSize)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

// TestFormValidator_ValidateFileType tests file type validation
func TestFormValidator_ValidateFileType(t *testing.T) {
	parser := NewFormParser()
	validator := NewFormValidator(parser)

	tests := []struct {
		name         string
		file         *FormFile
		allowedTypes []string
		expectError  bool
	}{
		{
			name: "allowed file type",
			file: &FormFile{
				Filename: "image.jpg",
				Header: map[string][]string{
					"Content-Type": {"image/jpeg"},
				},
			},
			allowedTypes: []string{"image/jpeg", "image/png"},
			expectError:  false,
		},
		{
			name: "disallowed file type",
			file: &FormFile{
				Filename: "script.exe",
				Header: map[string][]string{
					"Content-Type": {"application/x-msdownload"},
				},
			},
			allowedTypes: []string{"image/jpeg", "image/png"},
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &Request{
				Files: map[string]*FormFile{
					"upload": tt.file,
				},
			}
			err := validator.ValidateFileType(req, "upload", tt.allowedTypes)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

// TestFormValidator_ValidateAll tests comprehensive validation
func TestFormValidator_ValidateAll(t *testing.T) {
	parser := NewFormParser()
	validator := NewFormValidator(parser)

	req := &Request{
		Form: map[string]string{
			"name":  "John Doe",
			"email": "john@example.com",
		},
		Files: map[string]*FormFile{
			"avatar": {
				Filename: "avatar.jpg",
				Content:  make([]byte, 500),
				Size:     500,
				Header: map[string][]string{
					"Content-Type": {"image/jpeg"},
				},
			},
		},
	}

	rules := FormValidationRules{
		RequiredFields: []string{"name", "email"},
		RequiredFiles:  []string{"avatar"},
		FileSizeLimits: map[string]int64{
			"avatar": 1000,
		},
		FileTypes: map[string][]string{
			"avatar": {"image/jpeg", "image/png"},
		},
	}

	err := validator.ValidateAll(req, rules)
	if err != nil {
		t.Errorf("ValidateAll failed: %v", err)
	}
}

// TestFormDataPipeline tests the form data validation pipeline
func TestFormDataPipeline(t *testing.T) {
	rules := FormValidationRules{
		RequiredFields: []string{"username", "password"},
	}

	pipeline := FormDataPipeline(rules)

	// Create request with form data
	body := "username=testuser&password=secret123"
	req := &Request{
		Method:  "POST",
		Header:  http.Header{"Content-Type": []string{"application/x-www-form-urlencoded"}},
		RawBody: []byte(body),
	}

	// Create context
	ctx := NewContext(req, nil, nil)

	// Execute pipeline
	result, err := pipeline(ctx)
	if err != nil {
		t.Errorf("Pipeline failed: %v", err)
	}
	if result != PipelineResultContinue {
		t.Errorf("Expected PipelineResultContinue, got %v", result)
	}
}

// TestFileUploadPipeline tests the file upload validation pipeline
func TestFileUploadPipeline(t *testing.T) {
	rules := FormValidationRules{
		RequiredFiles: []string{"document"},
		FileSizeLimits: map[string]int64{
			"document": 1000,
		},
	}

	pipeline := FileUploadPipeline(32<<20, rules)

	// Create multipart form data
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	fileWriter, _ := writer.CreateFormFile("document", "test.pdf")
	fileWriter.Write([]byte("PDF content"))
	writer.Close()

	// Create request
	req := &Request{
		Method:  "POST",
		Header:  http.Header{"Content-Type": []string{writer.FormDataContentType()}},
		RawBody: buf.Bytes(),
	}

	// Create context
	ctx := NewContext(req, nil, nil)

	// Execute pipeline
	result, err := pipeline(ctx)
	if err != nil {
		t.Errorf("Pipeline failed: %v", err)
	}
	if result != PipelineResultContinue {
		t.Errorf("Expected PipelineResultContinue, got %v", result)
	}
}

// TestContext_FormValue tests context form value access
func TestContext_FormValue(t *testing.T) {
	body := "username=testuser&email=test@example.com"
	req := &Request{
		Method:  "POST",
		Header:  http.Header{"Content-Type": []string{"application/x-www-form-urlencoded"}},
		RawBody: []byte(body),
		URL:     &url.URL{},
	}

	ctx := NewContext(req, nil, nil)

	// Test form value access
	username := ctx.FormValue("username")
	if username != "testuser" {
		t.Errorf("Expected username=testuser, got %s", username)
	}

	email := ctx.FormValue("email")
	if email != "test@example.com" {
		t.Errorf("Expected email=test@example.com, got %s", email)
	}
}

// TestContext_FormFile tests context form file access
func TestContext_FormFile(t *testing.T) {
	// Create multipart form data
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	fileWriter, _ := writer.CreateFormFile("upload", "test.txt")
	fileWriter.Write([]byte("test content"))
	writer.Close()

	req := &Request{
		Method:  "POST",
		Header:  http.Header{"Content-Type": []string{writer.FormDataContentType()}},
		RawBody: buf.Bytes(),
		URL:     &url.URL{},
	}

	ctx := NewContext(req, nil, nil)

	// Test form file access
	file, err := ctx.FormFile("upload")
	if err != nil {
		t.Fatalf("FormFile failed: %v", err)
	}
	if file.Filename != "test.txt" {
		t.Errorf("Expected filename=test.txt, got %s", file.Filename)
	}
	if string(file.Content) != "test content" {
		t.Errorf("Expected content='test content', got %s", string(file.Content))
	}
}

// TestFileManager_SaveUploadedFile tests saving uploaded files
func TestFileManager_SaveUploadedFile(t *testing.T) {
	// Create memory filesystem
	memFS := NewMemoryFileSystem()
	fm := NewFileManager(memFS)

	// Create multipart form data
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	fileWriter, _ := writer.CreateFormFile("upload", "test.txt")
	fileWriter.Write([]byte("test content"))
	writer.Close()

	req := &Request{
		Method:  "POST",
		Header:  http.Header{"Content-Type": []string{writer.FormDataContentType()}},
		RawBody: buf.Bytes(),
		URL:     &url.URL{},
	}

	ctx := NewContext(req, nil, nil)

	// Save uploaded file
	err := fm.SaveUploadedFile(ctx, "upload", "/uploads/test.txt")
	if err != nil {
		t.Fatalf("SaveUploadedFile failed: %v", err)
	}

	// Verify file was saved
	content, err := fm.Read("/uploads/test.txt")
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}
	if string(content) != "test content" {
		t.Errorf("Expected content='test content', got %s", string(content))
	}
}

// TestFormParser_EmptyForm tests parsing empty form data
func TestFormParser_EmptyForm(t *testing.T) {
	parser := NewFormParser()

	req := &Request{
		Method:  "POST",
		Header:  http.Header{"Content-Type": []string{"application/x-www-form-urlencoded"}},
		RawBody: []byte(""),
	}

	err := parser.ParseForm(req)
	if err != nil {
		t.Errorf("ParseForm should not fail on empty form: %v", err)
	}

	if req.Form == nil {
		t.Error("Form map should be initialized")
	}
}

// TestFormValidator_CustomRules tests custom validation rules
func TestFormValidator_CustomRules(t *testing.T) {
	parser := NewFormParser()
	validator := NewFormValidator(parser)

	req := &Request{
		Form: map[string]string{
			"age": "25",
		},
	}

	rules := FormValidationRules{
		CustomRules: map[string]func(string) error{
			"age": func(value string) error {
				if value == "" {
					return ErrFieldRequired
				}
				// Simple validation - just check if it's not empty
				return nil
			},
		},
	}

	err := validator.ValidateAll(req, rules)
	if err != nil {
		t.Errorf("ValidateAll with custom rules failed: %v", err)
	}
}

// TestFormParser_MultipleFieldsWithSameName tests handling multiple values (simplified)
func TestFormParser_MultipleFieldsWithSameName(t *testing.T) {
	parser := NewFormParser()

	// In our simplified implementation, only the first value is stored
	body := "tag=go&tag=web&tag=framework"
	req := &Request{
		Method:  "POST",
		Header:  http.Header{"Content-Type": []string{"application/x-www-form-urlencoded"}},
		RawBody: []byte(body),
	}

	err := parser.ParseForm(req)
	if err != nil {
		t.Fatalf("ParseForm failed: %v", err)
	}

	// Should have at least one value
	if req.Form["tag"] == "" {
		t.Error("Expected at least one tag value")
	}
}

// TestFormParser_InvalidMultipartForm tests handling invalid multipart data
func TestFormParser_InvalidMultipartForm(t *testing.T) {
	parser := NewFormParser()

	req := &Request{
		Method:  "POST",
		Header:  http.Header{"Content-Type": []string{"multipart/form-data; boundary=invalid"}},
		RawBody: []byte("invalid multipart data"),
	}

	err := parser.ParseMultipartForm(req, 32<<20)
	if err == nil {
		t.Error("Expected error for invalid multipart data")
	}
}

// TestFormValidator_ValidateField tests custom field validation
func TestFormValidator_ValidateField(t *testing.T) {
	parser := NewFormParser()
	validator := NewFormValidator(parser)

	req := &Request{
		Form: map[string]string{
			"email": "test@example.com",
		},
	}

	// Test valid email
	err := validator.ValidateField(req, "email", func(value string) error {
		if !strings.Contains(value, "@") {
			return ErrFieldInvalid
		}
		return nil
	})
	if err != nil {
		t.Errorf("Expected no error for valid email, got: %v", err)
	}

	// Test invalid email
	req.Form["email"] = "invalid-email"
	err = validator.ValidateField(req, "email", func(value string) error {
		if !strings.Contains(value, "@") {
			return ErrFieldInvalid
		}
		return nil
	})
	if err == nil {
		t.Error("Expected error for invalid email")
	}
}
