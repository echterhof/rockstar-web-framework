package main

import (
	"fmt"
	"log"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
	fmt.Println("=== Rockstar Web Framework - Virtual File System Example ===\n")

	// Example 1: Basic Memory File System
	fmt.Println("1. Basic Memory File System")
	memoryFSExample()

	// Example 2: OS File System
	fmt.Println("\n2. OS File System")
	osFSExample()

	// Example 3: Per-Host Virtual Filesystems
	fmt.Println("\n3. Per-Host Virtual Filesystems")
	perHostFSExample()

	// Example 4: Static File Serving with Router
	fmt.Println("\n4. Static File Serving with Router")
	staticFileServingExample()

	// Example 5: File Manager Operations
	fmt.Println("\n5. File Manager Operations")
	fileManagerExample()
}

// Example 1: Basic Memory File System
func memoryFSExample() {
	// Create a memory filesystem
	fs := pkg.NewMemoryFileSystem().(*pkg.MemoryFileSystem)

	// Add files
	fs.AddFile("/index.html", []byte("<html><body>Welcome!</body></html>"))
	fs.AddFile("/style.css", []byte("body { color: blue; }"))
	fs.AddFile("/js/app.js", []byte("console.log('Hello');"))

	// Add directory
	fs.AddDir("/images")

	// Check if files exist
	fmt.Printf("  /index.html exists: %v\n", fs.Exists("/index.html"))
	fmt.Printf("  /style.css exists: %v\n", fs.Exists("/style.css"))
	fmt.Printf("  /nonexistent.txt exists: %v\n", fs.Exists("/nonexistent.txt"))

	// Open and read a file
	file, err := fs.Open("/index.html")
	if err != nil {
		log.Printf("  Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	stat, _ := file.Stat()
	fmt.Printf("  File size: %d bytes\n", stat.Size())
}

// Example 2: OS File System
func osFSExample() {
	// Create an OS filesystem rooted at current directory
	fs := pkg.NewOSFileSystem(".")

	// Check if files exist
	fmt.Printf("  go.mod exists: %v\n", fs.Exists("go.mod"))
	fmt.Printf("  README.md exists: %v\n", fs.Exists("README.md"))

	// Try to open go.mod
	file, err := fs.Open("go.mod")
	if err != nil {
		fmt.Printf("  Could not open go.mod: %v\n", err)
	} else {
		defer file.Close()
		stat, _ := file.Stat()
		fmt.Printf("  go.mod size: %d bytes\n", stat.Size())
	}

	// Security: Try directory traversal (should fail)
	_, err = fs.Open("../../../etc/passwd")
	if err == pkg.ErrInvalidPath {
		fmt.Println("  ✓ Directory traversal prevented")
	}
}

// Example 3: Per-Host Virtual Filesystems
func perHostFSExample() {
	// Create a host filesystem manager
	manager := pkg.NewHostFileSystemManager()

	// Create filesystems for different hosts
	host1FS := pkg.NewMemoryFileSystem().(*pkg.MemoryFileSystem)
	host1FS.AddFile("/index.html", []byte("<html>Welcome to Host 1</html>"))
	host1FS.AddFile("/about.html", []byte("<html>About Host 1</html>"))

	host2FS := pkg.NewMemoryFileSystem().(*pkg.MemoryFileSystem)
	host2FS.AddFile("/index.html", []byte("<html>Welcome to Host 2</html>"))
	host2FS.AddFile("/contact.html", []byte("<html>Contact Host 2</html>"))

	// Register hosts
	manager.RegisterHost("host1.example.com", host1FS)
	manager.RegisterHost("host2.example.com", host2FS)

	// Access files from different hosts
	if fs, exists := manager.GetFileSystem("host1.example.com"); exists {
		fmt.Printf("  host1.example.com has /index.html: %v\n", fs.Exists("/index.html"))
		fmt.Printf("  host1.example.com has /about.html: %v\n", fs.Exists("/about.html"))
		fmt.Printf("  host1.example.com has /contact.html: %v\n", fs.Exists("/contact.html"))
	}

	if fs, exists := manager.GetFileSystem("host2.example.com"); exists {
		fmt.Printf("  host2.example.com has /index.html: %v\n", fs.Exists("/index.html"))
		fmt.Printf("  host2.example.com has /about.html: %v\n", fs.Exists("/about.html"))
		fmt.Printf("  host2.example.com has /contact.html: %v\n", fs.Exists("/contact.html"))
	}

	// Unregister a host
	manager.UnregisterHost("host1.example.com")
	_, exists := manager.GetFileSystem("host1.example.com")
	fmt.Printf("  host1.example.com exists after unregister: %v\n", exists)
}

// Example 4: Static File Serving with Router
func staticFileServingExample() {
	// Create a router
	router := pkg.NewRouter()

	// Create filesystems for different purposes
	publicFS := pkg.NewMemoryFileSystem().(*pkg.MemoryFileSystem)
	publicFS.AddFile("/index.html", []byte("<html>Public Index</html>"))
	publicFS.AddFile("/style.css", []byte("body { margin: 0; }"))

	assetsFS := pkg.NewMemoryFileSystem().(*pkg.MemoryFileSystem)
	assetsFS.AddFile("/logo.png", []byte("PNG data..."))
	assetsFS.AddFile("/icon.svg", []byte("<svg>...</svg>"))

	// Register static routes
	router.Static("/public", publicFS)
	router.Static("/assets", assetsFS)

	// Test route matching
	route, params, found := router.Match("GET", "/public/index.html", "")
	if found {
		fmt.Printf("  ✓ Matched route: %s (static: %v)\n", route.Path, route.IsStatic)
		fmt.Printf("    Filepath parameter: %s\n", params["filepath"])
	}

	route, params, found = router.Match("GET", "/assets/logo.png", "")
	if found {
		fmt.Printf("  ✓ Matched route: %s (static: %v)\n", route.Path, route.IsStatic)
		fmt.Printf("    Filepath parameter: %s\n", params["filepath"])
	}
}

// Example 5: File Manager Operations
func fileManagerExample() {
	// Create a file manager with memory filesystem
	vfs := pkg.NewMemoryFileSystem()
	fm := pkg.NewFileManager(vfs)

	// Write files
	fmt.Println("  Writing files...")
	fm.Write("/config.json", []byte(`{"port": 8080, "host": "localhost"}`))
	fm.Write("/data.txt", []byte("Some important data"))

	// Create directory
	fm.CreateDir("/uploads")

	// Check existence
	fmt.Printf("  /config.json exists: %v\n", fm.Exists("/config.json"))
	fmt.Printf("  /data.txt exists: %v\n", fm.Exists("/data.txt"))
	fmt.Printf("  /uploads exists: %v\n", fm.Exists("/uploads"))

	// Read file
	data, err := fm.Read("/config.json")
	if err != nil {
		fmt.Printf("  Error reading file: %v\n", err)
	} else {
		fmt.Printf("  Config content: %s\n", string(data))
	}

	// Delete file
	fmt.Println("  Deleting /data.txt...")
	fm.Delete("/data.txt")
	fmt.Printf("  /data.txt exists after delete: %v\n", fm.Exists("/data.txt"))
}

// Example handler using FileManager in context
func exampleHandler(ctx pkg.Context) error {
	fm := ctx.Files()

	// Read a configuration file
	config, err := fm.Read("/config.json")
	if err != nil {
		return ctx.JSON(500, map[string]string{
			"error": "Failed to read config",
		})
	}

	// Write a log file
	logData := []byte(fmt.Sprintf("Request from %s",
		ctx.Request().RemoteAddr))
	fm.Write("/logs/access.log", logData)

	return ctx.JSON(200, map[string]interface{}{
		"config": string(config),
		"status": "success",
	})
}

// Example: Multi-tenant file serving
func multiTenantExample() {
	router := pkg.NewRouter()

	// Create tenant-specific filesystems
	tenant1FS := pkg.NewMemoryFileSystem().(*pkg.MemoryFileSystem)
	tenant1FS.AddFile("/index.html", []byte("<html>Tenant 1 Portal</html>"))
	tenant1FS.AddFile("/logo.png", []byte("Tenant 1 logo data"))

	tenant2FS := pkg.NewMemoryFileSystem().(*pkg.MemoryFileSystem)
	tenant2FS.AddFile("/index.html", []byte("<html>Tenant 2 Portal</html>"))
	tenant2FS.AddFile("/logo.png", []byte("Tenant 2 logo data"))

	// Register per-host routes for tenants
	router.Host("tenant1.example.com").Static("/", tenant1FS)
	router.Host("tenant2.example.com").Static("/", tenant2FS)

	// Each tenant gets their own isolated filesystem
	// tenant1.example.com/index.html -> Tenant 1 content
	// tenant2.example.com/index.html -> Tenant 2 content
}
