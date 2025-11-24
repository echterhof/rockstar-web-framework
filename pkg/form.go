package pkg

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"strings"
)

var (
	// ErrFormNotParsed is returned when form data has not been parsed
	ErrFormNotParsed = errors.New("form data not parsed")
	// ErrFieldRequired is returned when a required field is missing
	ErrFieldRequired = errors.New("required field missing")
	// ErrFieldInvalid is returned when a field value is invalid
	ErrFieldInvalid = errors.New("field value invalid")
	// ErrFileRequired is returned when a required file is missing
	ErrFileRequired = errors.New("required file missing")
	// ErrFileTooLarge is returned when a file exceeds size limits
	ErrFileTooLarge = errors.New("file too large")
	// ErrInvalidFileType is returned when a file type is not allowed
	ErrInvalidFileType = errors.New("invalid file type")
)

// FormParser handles parsing of form data from requests
type FormParser interface {
	// ParseForm parses form data from the request
	ParseForm(req *Request) error

	// ParseMultipartForm parses multipart form data with a max memory limit
	ParseMultipartForm(req *Request, maxMemory int64) error

	// GetFormValue gets a form value by key
	GetFormValue(req *Request, key string) string

	// GetFormValues gets all values for a key
	GetFormValues(req *Request, key string) []string

	// GetFormFile gets an uploaded file by key
	GetFormFile(req *Request, key string) (*FormFile, error)

	// GetFormFiles gets all uploaded files for a key
	GetFormFiles(req *Request, key string) ([]*FormFile, error)

	// GetAllFormFiles gets all uploaded files
	GetAllFormFiles(req *Request) map[string][]*FormFile
}

// formParser implements FormParser
type formParser struct {
	maxMemory int64 // Default max memory for multipart forms (32MB)
}

// NewFormParser creates a new form parser
func NewFormParser() FormParser {
	return &formParser{
		maxMemory: 32 << 20, // 32MB default
	}
}

// ParseForm parses form data from the request
func (fp *formParser) ParseForm(req *Request) error {
	if req == nil {
		return errors.New("request is nil")
	}

	// Check if already parsed
	if req.Form != nil {
		return nil
	}

	// Initialize form map
	req.Form = make(map[string]string)

	// Check content type
	contentType := ""
	if req.Header != nil {
		contentType = req.Header.Get("Content-Type")
	}

	// Parse based on content type
	if strings.HasPrefix(contentType, "multipart/form-data") {
		return fp.ParseMultipartForm(req, fp.maxMemory)
	}

	// Parse URL-encoded form data
	if strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
		return fp.parseURLEncodedForm(req)
	}

	// If no content type or unsupported, try to parse query parameters
	if req.URL != nil {
		for key, values := range req.URL.Query() {
			if len(values) > 0 {
				req.Form[key] = values[0]
			}
		}
	}

	return nil
}

// parseURLEncodedForm parses URL-encoded form data
func (fp *formParser) parseURLEncodedForm(req *Request) error {
	if req.RawBody == nil || len(req.RawBody) == 0 {
		return nil
	}

	// Parse the form data
	body := string(req.RawBody)
	pairs := strings.Split(body, "&")

	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			value := parts[1]
			req.Form[key] = value
		}
	}

	return nil
}

// ParseMultipartForm parses multipart form data with a max memory limit
func (fp *formParser) ParseMultipartForm(req *Request, maxMemory int64) error {
	if req == nil {
		return errors.New("request is nil")
	}

	// Check if already parsed
	if req.Form != nil && req.Files != nil {
		return nil
	}

	// Initialize maps
	if req.Form == nil {
		req.Form = make(map[string]string)
	}
	if req.Files == nil {
		req.Files = make(map[string]*FormFile)
	}

	// Get content type and boundary
	contentType := ""
	if req.Header != nil {
		contentType = req.Header.Get("Content-Type")
	}

	if contentType == "" {
		return errors.New("missing content-type header")
	}

	// Parse media type and params
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return fmt.Errorf("failed to parse content-type: %w", err)
	}

	if !strings.HasPrefix(mediaType, "multipart/") {
		return errors.New("not a multipart form")
	}

	boundary, ok := params["boundary"]
	if !ok {
		return errors.New("missing boundary in content-type")
	}

	// Create multipart reader
	var reader io.Reader
	if req.Body != nil {
		reader = req.Body
	} else if req.RawBody != nil {
		reader = bytes.NewReader(req.RawBody)
	} else {
		return errors.New("no request body")
	}

	multipartReader := multipart.NewReader(reader, boundary)

	// Parse each part
	for {
		part, err := multipartReader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read multipart part: %w", err)
		}

		// Get field name
		fieldName := part.FormName()
		if fieldName == "" {
			part.Close()
			continue
		}

		// Check if it's a file
		filename := part.FileName()
		if filename != "" {
			// Read file content
			content, err := io.ReadAll(part)
			if err != nil {
				part.Close()
				return fmt.Errorf("failed to read file content: %w", err)
			}

			// Create FormFile
			formFile := &FormFile{
				Filename: filename,
				Header:   make(map[string][]string),
				Size:     int64(len(content)),
				Content:  content,
			}

			// Copy headers
			for key, values := range part.Header {
				formFile.Header[key] = values
			}

			// Store file (only first file per field name for now)
			if _, exists := req.Files[fieldName]; !exists {
				req.Files[fieldName] = formFile
			}
		} else {
			// Read form value
			value, err := io.ReadAll(part)
			if err != nil {
				part.Close()
				return fmt.Errorf("failed to read form value: %w", err)
			}

			// Store value (only first value per field name)
			if _, exists := req.Form[fieldName]; !exists {
				req.Form[fieldName] = string(value)
			}
		}

		part.Close()
	}

	return nil
}

// GetFormValue gets a form value by key
func (fp *formParser) GetFormValue(req *Request, key string) string {
	if req.Form == nil {
		return ""
	}
	return req.Form[key]
}

// GetFormValues gets all values for a key (simplified - returns single value as slice)
func (fp *formParser) GetFormValues(req *Request, key string) []string {
	if req.Form == nil {
		return nil
	}
	value := req.Form[key]
	if value == "" {
		return nil
	}
	return []string{value}
}

// GetFormFile gets an uploaded file by key
func (fp *formParser) GetFormFile(req *Request, key string) (*FormFile, error) {
	if req.Files == nil {
		return nil, ErrFormNotParsed
	}

	file, exists := req.Files[key]
	if !exists {
		return nil, ErrFileRequired
	}

	return file, nil
}

// GetFormFiles gets all uploaded files for a key (simplified - returns single file as slice)
func (fp *formParser) GetFormFiles(req *Request, key string) ([]*FormFile, error) {
	file, err := fp.GetFormFile(req, key)
	if err != nil {
		return nil, err
	}
	return []*FormFile{file}, nil
}

// GetAllFormFiles gets all uploaded files
func (fp *formParser) GetAllFormFiles(req *Request) map[string][]*FormFile {
	if req.Files == nil {
		return nil
	}

	result := make(map[string][]*FormFile)
	for key, file := range req.Files {
		result[key] = []*FormFile{file}
	}
	return result
}

// FormValidator validates form data
type FormValidator interface {
	// ValidateRequired checks if required fields are present
	ValidateRequired(req *Request, fields []string) error

	// ValidateRequiredFiles checks if required files are present
	ValidateRequiredFiles(req *Request, fields []string) error

	// ValidateFileSize checks if files are within size limits
	ValidateFileSize(req *Request, field string, maxSize int64) error

	// ValidateFileType checks if files have allowed types
	ValidateFileType(req *Request, field string, allowedTypes []string) error

	// ValidateField validates a field with a custom validator function
	ValidateField(req *Request, field string, validator func(string) error) error

	// ValidateAll validates all rules
	ValidateAll(req *Request, rules FormValidationRules) error
}

// FormValidationRules defines validation rules for form data
type FormValidationRules struct {
	RequiredFields []string
	RequiredFiles  []string
	FileSizeLimits map[string]int64
	FileTypes      map[string][]string
	CustomRules    map[string]func(string) error
}

// formValidator implements FormValidator
type formValidator struct {
	parser FormParser
}

// NewFormValidator creates a new form validator
func NewFormValidator(parser FormParser) FormValidator {
	return &formValidator{
		parser: parser,
	}
}

// ValidateRequired checks if required fields are present
func (fv *formValidator) ValidateRequired(req *Request, fields []string) error {
	for _, field := range fields {
		value := fv.parser.GetFormValue(req, field)
		if value == "" {
			return fmt.Errorf("%w: %s", ErrFieldRequired, field)
		}
	}
	return nil
}

// ValidateRequiredFiles checks if required files are present
func (fv *formValidator) ValidateRequiredFiles(req *Request, fields []string) error {
	for _, field := range fields {
		_, err := fv.parser.GetFormFile(req, field)
		if err != nil {
			return fmt.Errorf("%w: %s", ErrFileRequired, field)
		}
	}
	return nil
}

// ValidateFileSize checks if files are within size limits
func (fv *formValidator) ValidateFileSize(req *Request, field string, maxSize int64) error {
	file, err := fv.parser.GetFormFile(req, field)
	if err != nil {
		return err
	}

	if file.Size > maxSize {
		return fmt.Errorf("%w: %s (size: %d, max: %d)", ErrFileTooLarge, field, file.Size, maxSize)
	}

	return nil
}

// ValidateFileType checks if files have allowed types
func (fv *formValidator) ValidateFileType(req *Request, field string, allowedTypes []string) error {
	file, err := fv.parser.GetFormFile(req, field)
	if err != nil {
		return err
	}

	// Get content type from header
	contentType := ""
	if file.Header != nil {
		if ct, ok := file.Header["Content-Type"]; ok && len(ct) > 0 {
			contentType = ct[0]
		}
	}

	// Check if content type is allowed
	for _, allowed := range allowedTypes {
		if strings.HasPrefix(contentType, allowed) {
			return nil
		}
	}

	return fmt.Errorf("%w: %s (type: %s)", ErrInvalidFileType, field, contentType)
}

// ValidateField validates a field with a custom validator function
func (fv *formValidator) ValidateField(req *Request, field string, validator func(string) error) error {
	value := fv.parser.GetFormValue(req, field)
	return validator(value)
}

// ValidateAll validates all rules
func (fv *formValidator) ValidateAll(req *Request, rules FormValidationRules) error {
	// Validate required fields
	if err := fv.ValidateRequired(req, rules.RequiredFields); err != nil {
		return err
	}

	// Validate required files
	if err := fv.ValidateRequiredFiles(req, rules.RequiredFiles); err != nil {
		return err
	}

	// Validate file sizes
	for field, maxSize := range rules.FileSizeLimits {
		if err := fv.ValidateFileSize(req, field, maxSize); err != nil {
			return err
		}
	}

	// Validate file types
	for field, allowedTypes := range rules.FileTypes {
		if err := fv.ValidateFileType(req, field, allowedTypes); err != nil {
			return err
		}
	}

	// Validate custom rules
	for field, validator := range rules.CustomRules {
		if err := fv.ValidateField(req, field, validator); err != nil {
			return err
		}
	}

	return nil
}

// FormDataPipeline creates a pipeline for form data validation
func FormDataPipeline(rules FormValidationRules) PipelineFunc {
	parser := NewFormParser()
	validator := NewFormValidator(parser)

	return func(ctx Context) (PipelineResult, error) {
		req := ctx.Request()
		if req == nil {
			return PipelineResultClose, errors.New("request is nil")
		}

		// Parse form data
		if err := parser.ParseForm(req); err != nil {
			return PipelineResultClose, fmt.Errorf("failed to parse form: %w", err)
		}

		// Validate form data
		if err := validator.ValidateAll(req, rules); err != nil {
			return PipelineResultClose, fmt.Errorf("form validation failed: %w", err)
		}

		return PipelineResultContinue, nil
	}
}

// FileUploadPipeline creates a pipeline for file upload validation
func FileUploadPipeline(maxMemory int64, rules FormValidationRules) PipelineFunc {
	parser := NewFormParser()
	validator := NewFormValidator(parser)

	return func(ctx Context) (PipelineResult, error) {
		req := ctx.Request()
		if req == nil {
			return PipelineResultClose, errors.New("request is nil")
		}

		// Parse multipart form data
		if err := parser.ParseMultipartForm(req, maxMemory); err != nil {
			return PipelineResultClose, fmt.Errorf("failed to parse multipart form: %w", err)
		}

		// Validate form data
		if err := validator.ValidateAll(req, rules); err != nil {
			return PipelineResultClose, fmt.Errorf("file upload validation failed: %w", err)
		}

		return PipelineResultContinue, nil
	}
}
