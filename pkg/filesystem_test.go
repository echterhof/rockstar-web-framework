package pkg

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

// TestOSFileSystem tests the OS-based virtual filesystem
func TestOSFileSystem(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tempDir, "test.txt")
	testData := []byte("Hello, World!")
	if err := os.WriteFile(testFile, testData, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create filesystem
	fs := NewOSFileSystem(tempDir)

	t.Run("Open existing file", func(t *testing.T) {
		file, err := fs.Open("test.txt")
		if err != nil {
			t.Fatalf("Failed to open file: %v", err)
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}

		if string(data) != string(testData) {
			t.Errorf("Expected %s, got %s", testData, data)
		}
	})

	t.Run("Open non-existent file", func(t *testing.T) {
		_, err := fs.Open("nonexistent.txt")
		if err != ErrFileNotFound {
			t.Errorf("Expected ErrFileNotFound, got %v", err)
		}
	})

	t.Run("Exists returns true for existing file", func(t *testing.T) {
		if !fs.Exists("test.txt") {
			t.Error("Expected file to exist")
		}
	})

	t.Run("Exists returns false for non-existent file", func(t *testing.T) {
		if fs.Exists("nonexistent.txt") {
			t.Error("Expected file to not exist")
		}
	})

	t.Run("Prevent directory traversal", func(t *testing.T) {
		_, err := fs.Open("../../../etc/passwd")
		if err != ErrInvalidPath && err != ErrFileNotFound {
			t.Errorf("Expected ErrInvalidPath or ErrFileNotFound for directory traversal, got %v", err)
		}
	})
}

// TestMemoryFileSystem tests the in-memory virtual filesystem
func TestMemoryFileSystem(t *testing.T) {
	fs := NewMemoryFileSystem().(*MemoryFileSystem)

	t.Run("Add and open file", func(t *testing.T) {
		testData := []byte("Hello, Memory!")
		if err := fs.AddFile("/test.txt", testData); err != nil {
			t.Fatalf("Failed to add file: %v", err)
		}

		file, err := fs.Open("/test.txt")
		if err != nil {
			t.Fatalf("Failed to open file: %v", err)
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}

		if string(data) != string(testData) {
			t.Errorf("Expected %s, got %s", testData, data)
		}
	})

	t.Run("Open non-existent file", func(t *testing.T) {
		_, err := fs.Open("/nonexistent.txt")
		if err != ErrFileNotFound {
			t.Errorf("Expected ErrFileNotFound, got %v", err)
		}
	})

	t.Run("Exists returns true for existing file", func(t *testing.T) {
		fs.AddFile("/exists.txt", []byte("data"))
		if !fs.Exists("/exists.txt") {
			t.Error("Expected file to exist")
		}
	})

	t.Run("Exists returns false for non-existent file", func(t *testing.T) {
		if fs.Exists("/nonexistent.txt") {
			t.Error("Expected file to not exist")
		}
	})

	t.Run("Add and open directory", func(t *testing.T) {
		if err := fs.AddDir("/testdir"); err != nil {
			t.Fatalf("Failed to add directory: %v", err)
		}

		file, err := fs.Open("/testdir")
		if err != nil {
			t.Fatalf("Failed to open directory: %v", err)
		}
		defer file.Close()

		stat, err := file.Stat()
		if err != nil {
			t.Fatalf("Failed to stat directory: %v", err)
		}

		if !stat.IsDir() {
			t.Error("Expected directory")
		}
	})

	t.Run("Read from directory returns error", func(t *testing.T) {
		fs.AddDir("/readdir")
		file, err := fs.Open("/readdir")
		if err != nil {
			t.Fatalf("Failed to open directory: %v", err)
		}
		defer file.Close()

		buf := make([]byte, 10)
		_, err = file.Read(buf)
		if err != ErrIsDirectory {
			t.Errorf("Expected ErrIsDirectory, got %v", err)
		}
	})
}

// TestMemoryHTTPFile tests the memory HTTP file implementation
func TestMemoryHTTPFile(t *testing.T) {
	fs := NewMemoryFileSystem().(*MemoryFileSystem)
	testData := []byte("Hello, World!")
	fs.AddFile("/test.txt", testData)

	file, err := fs.Open("/test.txt")
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	t.Run("Read file", func(t *testing.T) {
		buf := make([]byte, 5)
		n, err := file.Read(buf)
		if err != nil {
			t.Fatalf("Failed to read: %v", err)
		}
		if n != 5 {
			t.Errorf("Expected to read 5 bytes, got %d", n)
		}
		if string(buf) != "Hello" {
			t.Errorf("Expected 'Hello', got '%s'", buf)
		}
	})

	t.Run("Seek file", func(t *testing.T) {
		offset, err := file.Seek(0, 0)
		if err != nil {
			t.Fatalf("Failed to seek: %v", err)
		}
		if offset != 0 {
			t.Errorf("Expected offset 0, got %d", offset)
		}

		offset, err = file.Seek(7, 0)
		if err != nil {
			t.Fatalf("Failed to seek: %v", err)
		}
		if offset != 7 {
			t.Errorf("Expected offset 7, got %d", offset)
		}
	})

	t.Run("Stat file", func(t *testing.T) {
		stat, err := file.Stat()
		if err != nil {
			t.Fatalf("Failed to stat: %v", err)
		}

		if stat.Size() != int64(len(testData)) {
			t.Errorf("Expected size %d, got %d", len(testData), stat.Size())
		}

		if stat.IsDir() {
			t.Error("Expected file, not directory")
		}
	})
}

// TestHostFileSystemManager tests per-host filesystem management
func TestHostFileSystemManager(t *testing.T) {
	manager := NewHostFileSystemManager()

	t.Run("Register and get filesystem", func(t *testing.T) {
		fs := NewMemoryFileSystem()
		manager.RegisterHost("example.com", fs)

		retrieved, exists := manager.GetFileSystem("example.com")
		if !exists {
			t.Error("Expected filesystem to exist")
		}
		if retrieved != fs {
			t.Error("Expected same filesystem instance")
		}
	})

	t.Run("Get non-existent filesystem", func(t *testing.T) {
		_, exists := manager.GetFileSystem("nonexistent.com")
		if exists {
			t.Error("Expected filesystem to not exist")
		}
	})

	t.Run("Unregister filesystem", func(t *testing.T) {
		fs := NewMemoryFileSystem()
		manager.RegisterHost("remove.com", fs)

		_, exists := manager.GetFileSystem("remove.com")
		if !exists {
			t.Error("Expected filesystem to exist before removal")
		}

		manager.UnregisterHost("remove.com")

		_, exists = manager.GetFileSystem("remove.com")
		if exists {
			t.Error("Expected filesystem to not exist after removal")
		}
	})

	t.Run("Multiple hosts with different filesystems", func(t *testing.T) {
		fs1 := NewMemoryFileSystem().(*MemoryFileSystem)
		fs2 := NewMemoryFileSystem().(*MemoryFileSystem)

		fs1.AddFile("/host1.txt", []byte("Host 1"))
		fs2.AddFile("/host2.txt", []byte("Host 2"))

		manager.RegisterHost("host1.com", fs1)
		manager.RegisterHost("host2.com", fs2)

		// Verify host1 has its file
		retrieved1, _ := manager.GetFileSystem("host1.com")
		if !retrieved1.Exists("/host1.txt") {
			t.Error("Expected host1.txt to exist in host1.com")
		}
		if retrieved1.Exists("/host2.txt") {
			t.Error("Expected host2.txt to not exist in host1.com")
		}

		// Verify host2 has its file
		retrieved2, _ := manager.GetFileSystem("host2.com")
		if !retrieved2.Exists("/host2.txt") {
			t.Error("Expected host2.txt to exist in host2.com")
		}
		if retrieved2.Exists("/host1.txt") {
			t.Error("Expected host1.txt to not exist in host2.com")
		}
	})
}

// TestFileManager tests the file manager implementation
func TestFileManager(t *testing.T) {
	t.Run("FileManager with memory filesystem", func(t *testing.T) {
		vfs := NewMemoryFileSystem()
		fm := NewFileManager(vfs)

		// Test Write
		testData := []byte("Test data")
		if err := fm.Write("/test.txt", testData); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}

		// Test Exists
		if !fm.Exists("/test.txt") {
			t.Error("Expected file to exist")
		}

		// Test Read
		data, err := fm.Read("/test.txt")
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}
		if string(data) != string(testData) {
			t.Errorf("Expected %s, got %s", testData, data)
		}

		// Test Delete
		if err := fm.Delete("/test.txt"); err != nil {
			t.Fatalf("Failed to delete file: %v", err)
		}
		if fm.Exists("/test.txt") {
			t.Error("Expected file to not exist after deletion")
		}
	})

	t.Run("FileManager with OS filesystem", func(t *testing.T) {
		tempDir := t.TempDir()
		vfs := NewOSFileSystem(tempDir)
		fm := NewFileManager(vfs)

		// Test Write
		testData := []byte("OS Test data")
		if err := fm.Write("/ostest.txt", testData); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}

		// Test Exists
		if !fm.Exists("/ostest.txt") {
			t.Error("Expected file to exist")
		}

		// Test Read
		data, err := fm.Read("/ostest.txt")
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}
		if string(data) != string(testData) {
			t.Errorf("Expected %s, got %s", testData, data)
		}

		// Test Delete
		if err := fm.Delete("/ostest.txt"); err != nil {
			t.Fatalf("Failed to delete file: %v", err)
		}
		if fm.Exists("/ostest.txt") {
			t.Error("Expected file to not exist after deletion")
		}
	})

	t.Run("FileManager CreateDir", func(t *testing.T) {
		vfs := NewMemoryFileSystem()
		fm := NewFileManager(vfs)

		if err := fm.CreateDir("/testdir"); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}

		if !fm.Exists("/testdir") {
			t.Error("Expected directory to exist")
		}
	})
}

// TestStaticFileServing tests static file serving through the router
func TestStaticFileServing(t *testing.T) {
	// Create a memory filesystem with test files
	fs := NewMemoryFileSystem().(*MemoryFileSystem)
	fs.AddFile("/index.html", []byte("<html>Index</html>"))
	fs.AddFile("/style.css", []byte("body { color: red; }"))
	fs.AddFile("/js/app.js", []byte("console.log('app');"))

	// Create router and register static route
	router := NewRouter()
	router.Static("/static", fs)

	t.Run("Match static route", func(t *testing.T) {
		_, params, found := router.Match("GET", "/static/index.html", "")
		if !found {
			t.Fatal("Expected to find static route")
		}
		if params["filepath"] != "index.html" {
			t.Errorf("Expected filepath 'index.html', got '%s'", params["filepath"])
		}
	})

	t.Run("Match nested static file", func(t *testing.T) {
		_, params, found := router.Match("GET", "/static/js/app.js", "")
		if !found {
			t.Fatal("Expected to find static route")
		}
		if params["filepath"] != "js/app.js" {
			t.Errorf("Expected filepath 'js/app.js', got '%s'", params["filepath"])
		}
	})
}

// TestPerHostVirtualFS tests per-host virtual filesystem support
func TestPerHostVirtualFS(t *testing.T) {
	// Create filesystems for different hosts
	fs1 := NewMemoryFileSystem().(*MemoryFileSystem)
	fs1.AddFile("/index.html", []byte("<html>Host 1</html>"))

	fs2 := NewMemoryFileSystem().(*MemoryFileSystem)
	fs2.AddFile("/index.html", []byte("<html>Host 2</html>"))

	// Create router with host-specific static routes
	router := NewRouter()
	router.Host("host1.com").Static("/", fs1)
	router.Host("host2.com").Static("/", fs2)

	t.Run("Match host1 static file", func(t *testing.T) {
		_, params, found := router.Match("GET", "/index.html", "host1.com")
		if !found {
			t.Fatal("Expected to find static route for host1.com")
		}
		if params["filepath"] != "index.html" {
			t.Errorf("Expected filepath 'index.html', got '%s'", params["filepath"])
		}
	})

	t.Run("Match host2 static file", func(t *testing.T) {
		_, _, found := router.Match("GET", "/index.html", "host2.com")
		if !found {
			t.Fatal("Expected to find static route for host2.com")
		}
	})

	t.Run("Different hosts have isolated filesystems", func(t *testing.T) {
		// Verify that host1's filesystem has its file
		file1, err := fs1.Open("/index.html")
		if err != nil {
			t.Fatalf("Failed to open file from host1: %v", err)
		}
		defer file1.Close()

		data1, _ := io.ReadAll(file1)
		if string(data1) != "<html>Host 1</html>" {
			t.Errorf("Expected Host 1 content, got %s", data1)
		}

		// Verify that host2's filesystem has its file
		file2, err := fs2.Open("/index.html")
		if err != nil {
			t.Fatalf("Failed to open file from host2: %v", err)
		}
		defer file2.Close()

		data2, _ := io.ReadAll(file2)
		if string(data2) != "<html>Host 2</html>" {
			t.Errorf("Expected Host 2 content, got %s", data2)
		}
	})
}

// TestConcurrentFileSystemAccess tests concurrent access to filesystems
func TestConcurrentFileSystemAccess(t *testing.T) {
	fs := NewMemoryFileSystem().(*MemoryFileSystem)

	// Add initial files
	for i := 0; i < 10; i++ {
		fs.AddFile("/file"+string(rune('0'+i))+".txt", []byte("data"))
	}

	// Concurrent reads
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(idx int) {
			for j := 0; j < 100; j++ {
				filename := "/file" + string(rune('0'+idx)) + ".txt"
				if !fs.Exists(filename) {
					t.Errorf("File %s should exist", filename)
				}
				file, err := fs.Open(filename)
				if err != nil {
					t.Errorf("Failed to open file: %v", err)
				}
				file.Close()
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}
