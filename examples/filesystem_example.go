//go:build ignore

package main

import (
	"fmt"
	"log"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
	// ============================================================================
	// Configuration Setup
	// ============================================================================
	config := pkg.FrameworkConfig{
		ServerConfig: pkg.ServerConfig{
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			EnableHTTP1:  true,
			EnableHTTP2:  true,
		},
	}

	// ============================================================================
	// Framework Initialization
	// ============================================================================
	app, err := pkg.New(config)
	if err != nil {
		log.Fatalf("Failed to create framework: %v", err)
	}

	// Initialize file manager with memory filesystem for demonstration
	memFS := pkg.NewMemoryFileSystem().(*pkg.MemoryFileSystem)
	fileManager := pkg.NewFileManager(memFS)

	// Add some sample files to the virtual filesystem
	memFS.AddFile("/readme.txt", []byte("Welcome to the Rockstar Web Framework!"))
	memFS.AddFile("/data/users.json", []byte(`{"users": [{"id": 1, "name": "John"}]}`))
	memFS.AddDir("/uploads")

	// ============================================================================
	// Route Registration
	// ============================================================================
	router := app.Router()

	// File upload handling
	router.POST("/api/files/upload", uploadFileHandler(fileManager))
	router.POST("/api/files/upload-multiple", uploadMultipleFilesHandler(fileManager))

	// File download serving
	router.GET("/api/files/download/:filename", downloadFileHandler(fileManager))
	router.GET("/api/files/read/:filename", readFileHandler(fileManager))

	// Virtual filesystem operations
	router.GET("/api/files/list", listFilesHandler(fileManager))
	router.GET("/api/files/exists/:filename", fileExistsHandler(fileManager))
	router.GET("/api/files/info/:filename", fileInfoHandler(fileManager))
	router.DELETE("/api/files/delete/:filename", deleteFileHandler(fileManager))

	// Directory operations
	router.POST("/api/files/mkdir", createDirectoryHandler(fileManager))
	router.GET("/api/files/listdir/:dirname", listDirectoryHandler(fileManager))

	// File validation and security
	router.POST("/api/files/upload-validated", uploadValidatedFileHandler(fileManager))

	// ============================================================================
	// Server Startup
	// ============================================================================
	fmt.Println("🎸 Rockstar Web Framework - Filesystem Example")
	fmt.Println("=" + "==========================================================")
	fmt.Println()
	fmt.Println("Server listening on: http://localhost:8080")
	fmt.Println()
	fmt.Println("Try these commands:")
	fmt.Println("  # File upload handling")
	fmt.Println("  curl -X POST -F \"file=@yourfile.txt\" http://localhost:8080/api/files/upload")
	fmt.Println("  curl -X POST -F \"file1=@file1.txt\" -F \"file2=@file2.txt\" http://localhost:8080/api/files/upload-multiple")
	fmt.Println()
	fmt.Println("  # File download serving")
	fmt.Println("  curl http://localhost:8080/api/files/download/readme.txt")
	fmt.Println("  curl http://localhost:8080/api/files/read/readme.txt")
	fmt.Println()
	fmt.Println("  # Virtual filesystem operations")
	fmt.Println("  curl http://localhost:8080/api/files/list")
	fmt.Println("  curl http://localhost:8080/api/files/exists/readme.txt")
	fmt.Println("  curl http://localhost:8080/api/files/info/readme.txt")
	fmt.Println("  curl -X DELETE http://localhost:8080/api/files/delete/readme.txt")
	fmt.Println()
	fmt.Println("  # Directory operations")
	fmt.Println("  curl -X POST -d 'dirname=newdir' http://localhost:8080/api/files/mkdir")
	fmt.Println("  curl http://localhost:8080/api/files/listdir/data")
	fmt.Println()
	fmt.Println("  # File validation")
	fmt.Println("  curl -X POST -F \"file=@yourfile.txt\" http://localhost:8080/api/files/upload-validated")
	fmt.Println()
	fmt.Println("=" + "==========================================================")
	fmt.Println()

	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// ============================================================================
// Handler Functions - File Upload Handling
// ============================================================================

// uploadFileHandler demonstrates handling file uploads
func uploadFileHandler(fm pkg.FileManager) func(pkg.Context) error {
	return func(ctx pkg.Context) error {
		// Get uploaded file
		file, err := ctx.FormFile("file")
		if err != nil {
			return ctx.JSON(400, map[string]interface{}{
				"error":   "No file uploaded",
				"message": "Please provide a file with the 'file' field",
			})
		}

		// Save file to virtual filesystem
		uploadPath := "/uploads/" + file.Filename
		err = fm.Write(uploadPath, file.Content)
		if err != nil {
			return ctx.JSON(500, map[string]interface{}{
				"error": "Failed to save file",
			})
		}

		return ctx.JSON(200, map[string]interface{}{
			"message":  "File uploaded successfully",
			"filename": file.Filename,
			"size":     file.Size,
			"path":     uploadPath,
		})
	}
}

// uploadMultipleFilesHandler demonstrates handling multiple file uploads
func uploadMultipleFilesHandler(fm pkg.FileManager) func(pkg.Context) error {
	return func(ctx pkg.Context) error {
		// In this example, we'll handle files named file1, file2, etc.
		uploadedFiles := []map[string]interface{}{}

		// Try to get multiple files
		for i := 1; i <= 10; i++ {
			fieldName := fmt.Sprintf("file%d", i)
			file, err := ctx.FormFile(fieldName)
			if err != nil {
				// No more files
				break
			}

			// Save file
			uploadPath := "/uploads/" + file.Filename
			err = fm.Write(uploadPath, file.Content)
			if err != nil {
				return ctx.JSON(500, map[string]interface{}{
					"error": fmt.Sprintf("Failed to save file %s", file.Filename),
				})
			}

			uploadedFiles = append(uploadedFiles, map[string]interface{}{
				"filename": file.Filename,
				"size":     file.Size,
				"path":     uploadPath,
			})
		}

		if len(uploadedFiles) == 0 {
			return ctx.JSON(400, map[string]interface{}{
				"error":   "No files uploaded",
				"message": "Please provide files with fields named file1, file2, etc.",
			})
		}

		return ctx.JSON(200, map[string]interface{}{
			"message": "Files uploaded successfully",
			"count":   len(uploadedFiles),
			"files":   uploadedFiles,
		})
	}
}

// ============================================================================
// Handler Functions - File Download Serving
// ============================================================================

// downloadFileHandler demonstrates serving files for download
func downloadFileHandler(fm pkg.FileManager) func(pkg.Context) error {
	return func(ctx pkg.Context) error {
		filename := ctx.Params()["filename"]
		if filename == "" {
			return ctx.JSON(400, map[string]interface{}{
				"error": "Filename is required",
			})
		}

		// Read file from virtual filesystem
		filePath := "/" + filename
		content, err := fm.Read(filePath)
		if err != nil {
			return ctx.JSON(404, map[string]interface{}{
				"error":   "File not found",
				"message": fmt.Sprintf("File '%s' does not exist", filename),
			})
		}

		// Set headers for file download
		ctx.SetHeader("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
		ctx.SetHeader("Content-Type", "application/octet-stream")

		// Write the file content as response
		resp := ctx.Response()
		resp.WriteHeader(200)
		resp.Write(content)
		return nil
	}
}

// readFileHandler demonstrates reading and displaying file content
func readFileHandler(fm pkg.FileManager) func(pkg.Context) error {
	return func(ctx pkg.Context) error {
		filename := ctx.Params()["filename"]
		if filename == "" {
			return ctx.JSON(400, map[string]interface{}{
				"error": "Filename is required",
			})
		}

		// Read file from virtual filesystem
		filePath := "/" + filename
		content, err := fm.Read(filePath)
		if err != nil {
			return ctx.JSON(404, map[string]interface{}{
				"error":   "File not found",
				"message": fmt.Sprintf("File '%s' does not exist", filename),
			})
		}

		return ctx.JSON(200, map[string]interface{}{
			"message":  "File read successfully",
			"filename": filename,
			"size":     len(content),
			"content":  string(content),
		})
	}
}

// ============================================================================
// Handler Functions - Virtual Filesystem Operations
// ============================================================================

// listFilesHandler demonstrates listing all files in the virtual filesystem
func listFilesHandler(fm pkg.FileManager) func(pkg.Context) error {
	return func(ctx pkg.Context) error {
		// For memory filesystem, we can list files
		// This is a simplified example
		files := []string{
			"/readme.txt",
			"/data/users.json",
			"/uploads/",
		}

		return ctx.JSON(200, map[string]interface{}{
			"message": "Files listed successfully",
			"count":   len(files),
			"files":   files,
			"note":    "This is a simplified list from the memory filesystem",
		})
	}
}

// fileExistsHandler demonstrates checking if a file exists
func fileExistsHandler(fm pkg.FileManager) func(pkg.Context) error {
	return func(ctx pkg.Context) error {
		filename := ctx.Params()["filename"]
		if filename == "" {
			return ctx.JSON(400, map[string]interface{}{
				"error": "Filename is required",
			})
		}

		filePath := "/" + filename
		exists := fm.Exists(filePath)

		return ctx.JSON(200, map[string]interface{}{
			"filename": filename,
			"path":     filePath,
			"exists":   exists,
		})
	}
}

// fileInfoHandler demonstrates getting file information
func fileInfoHandler(fm pkg.FileManager) func(pkg.Context) error {
	return func(ctx pkg.Context) error {
		filename := ctx.Params()["filename"]
		if filename == "" {
			return ctx.JSON(400, map[string]interface{}{
				"error": "Filename is required",
			})
		}

		filePath := "/" + filename

		// Check if file exists
		if !fm.Exists(filePath) {
			return ctx.JSON(404, map[string]interface{}{
				"error":   "File not found",
				"message": fmt.Sprintf("File '%s' does not exist", filename),
			})
		}

		// Read file to get size
		content, err := fm.Read(filePath)
		if err != nil {
			return ctx.JSON(500, map[string]interface{}{
				"error": "Failed to read file",
			})
		}

		return ctx.JSON(200, map[string]interface{}{
			"message":  "File information retrieved",
			"filename": filename,
			"path":     filePath,
			"size":     len(content),
		})
	}
}

// deleteFileHandler demonstrates deleting a file
func deleteFileHandler(fm pkg.FileManager) func(pkg.Context) error {
	return func(ctx pkg.Context) error {
		filename := ctx.Params()["filename"]
		if filename == "" {
			return ctx.JSON(400, map[string]interface{}{
				"error": "Filename is required",
			})
		}

		filePath := "/" + filename

		// Check if file exists
		if !fm.Exists(filePath) {
			return ctx.JSON(404, map[string]interface{}{
				"error":   "File not found",
				"message": fmt.Sprintf("File '%s' does not exist", filename),
			})
		}

		// Delete file
		err := fm.Delete(filePath)
		if err != nil {
			return ctx.JSON(500, map[string]interface{}{
				"error": "Failed to delete file",
			})
		}

		return ctx.JSON(200, map[string]interface{}{
			"message":  "File deleted successfully",
			"filename": filename,
			"path":     filePath,
		})
	}
}

// ============================================================================
// Handler Functions - Directory Operations
// ============================================================================

// createDirectoryHandler demonstrates creating a directory
func createDirectoryHandler(fm pkg.FileManager) func(pkg.Context) error {
	return func(ctx pkg.Context) error {
		dirname := ctx.FormValue("dirname")
		if dirname == "" {
			return ctx.JSON(400, map[string]interface{}{
				"error":   "Directory name is required",
				"message": "Please provide 'dirname' in the request body",
			})
		}

		dirPath := "/" + dirname

		// Create directory
		err := fm.CreateDir(dirPath)
		if err != nil {
			return ctx.JSON(500, map[string]interface{}{
				"error": "Failed to create directory",
			})
		}

		return ctx.JSON(200, map[string]interface{}{
			"message": "Directory created successfully",
			"dirname": dirname,
			"path":    dirPath,
		})
	}
}

// listDirectoryHandler demonstrates listing directory contents
func listDirectoryHandler(fm pkg.FileManager) func(pkg.Context) error {
	return func(ctx pkg.Context) error {
		dirname := ctx.Params()["dirname"]
		if dirname == "" {
			return ctx.JSON(400, map[string]interface{}{
				"error": "Directory name is required",
			})
		}

		dirPath := "/" + dirname

		// Check if directory exists
		if !fm.Exists(dirPath) {
			return ctx.JSON(404, map[string]interface{}{
				"error":   "Directory not found",
				"message": fmt.Sprintf("Directory '%s' does not exist", dirname),
			})
		}

		// For this example, return a simplified list
		// In a real application, you would implement proper directory listing
		entries := []string{"users.json"}

		return ctx.JSON(200, map[string]interface{}{
			"message": "Directory listed successfully",
			"dirname": dirname,
			"path":    dirPath,
			"count":   len(entries),
			"entries": entries,
			"note":    "This is a simplified example",
		})
	}
}

// ============================================================================
// Handler Functions - File Validation and Security
// ============================================================================

// uploadValidatedFileHandler demonstrates file upload with validation
func uploadValidatedFileHandler(fm pkg.FileManager) func(pkg.Context) error {
	return func(ctx pkg.Context) error {
		// Get uploaded file
		file, err := ctx.FormFile("file")
		if err != nil {
			return ctx.JSON(400, map[string]interface{}{
				"error":   "No file uploaded",
				"message": "Please provide a file with the 'file' field",
			})
		}

		// Validate file size (max 10MB)
		maxSize := int64(10 * 1024 * 1024)
		if file.Size > maxSize {
			return ctx.JSON(400, map[string]interface{}{
				"error":   "File too large",
				"message": fmt.Sprintf("File size must be less than %d bytes", maxSize),
				"size":    file.Size,
			})
		}

		// Validate file extension (allow only .txt, .json, .md)
		allowedExtensions := []string{".txt", ".json", ".md"}
		validExtension := false
		for _, ext := range allowedExtensions {
			if len(file.Filename) >= len(ext) && file.Filename[len(file.Filename)-len(ext):] == ext {
				validExtension = true
				break
			}
		}

		if !validExtension {
			return ctx.JSON(400, map[string]interface{}{
				"error":              "Invalid file type",
				"message":            "Only .txt, .json, and .md files are allowed",
				"allowed_extensions": allowedExtensions,
			})
		}

		// Sanitize filename (remove path traversal attempts)
		sanitizedFilename := sanitizeFilename(file.Filename)

		// Save file to virtual filesystem
		uploadPath := "/uploads/" + sanitizedFilename
		err = fm.Write(uploadPath, file.Content)
		if err != nil {
			return ctx.JSON(500, map[string]interface{}{
				"error": "Failed to save file",
			})
		}

		return ctx.JSON(200, map[string]interface{}{
			"message":  "File uploaded and validated successfully",
			"filename": sanitizedFilename,
			"size":     file.Size,
			"path":     uploadPath,
			"validation": map[string]interface{}{
				"size_check":      "passed",
				"extension_check": "passed",
				"sanitization":    "applied",
			},
		})
	}
}

// sanitizeFilename removes potentially dangerous characters from filenames
func sanitizeFilename(filename string) string {
	// Remove path separators and other dangerous characters
	sanitized := ""
	for _, char := range filename {
		if char == '/' || char == '\\' {
			continue
		}
		sanitized += string(char)
	}
	// Also remove any ".." sequences
	for i := 0; i < len(sanitized)-1; i++ {
		if sanitized[i] == '.' && sanitized[i+1] == '.' {
			sanitized = sanitized[:i] + sanitized[i+2:]
			i--
		}
	}
	return sanitized
}
