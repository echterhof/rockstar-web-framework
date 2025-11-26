//go:build ignore

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
	// Create a new framework instance with proper session configuration
	framework, err := pkg.New(pkg.FrameworkConfig{
		DatabaseConfig: pkg.DatabaseConfig{
			Driver:   "sqlite3",
			Database: ":memory:",
		},
		SessionConfig: pkg.SessionConfig{
			StorageType:     "cache",
			SessionLifetime: 24 * time.Hour,
			EncryptionKey:   []byte("12345678901234567890123456789012"), // 32 bytes for AES-256
			CleanupInterval: 10 * time.Minute,
		},
	})
	if err != nil {
		log.Fatalf("Failed to create framework: %v", err)
	}

	// Create a template manager
	templateManager := pkg.NewTemplateManager()

	// Add custom template functions
	if err = templateManager.AddFunc("upper", func(s string) string {
		return fmt.Sprintf("%s", s)
	}); err != nil {
		log.Fatalf("Failed to add template function: %v", err)
	}

	// Load templates from strings (in production, you'd load from files)
	templates := map[string]string{
		"index.html": `
<!DOCTYPE html>
<html>
<head>
	<title>{{.Title}}</title>
	<style>
		body { font-family: Arial, sans-serif; margin: 40px; }
		h1 { color: #333; }
		.user-info { background: #f0f0f0; padding: 20px; border-radius: 5px; }
	</style>
</head>
<body>
	<h1>{{.Title}}</h1>
	<div class="user-info">
		<p><strong>Welcome, {{.Username}}!</strong></p>
		<p>Email: {{.Email}}</p>
	</div>
	<h2>Features:</h2>
	<ul>
	{{range .Features}}
		<li>{{.}}</li>
	{{end}}
	</ul>
</body>
</html>
`,
		"about.html": `
<!DOCTYPE html>
<html>
<head>
	<title>About - {{.AppName}}</title>
</head>
<body>
	<h1>About {{.AppName}}</h1>
	<p>{{.Description}}</p>
	<p>Version: {{.Version}}</p>
</body>
</html>
`,
		"error.html": `
<!DOCTYPE html>
<html>
<head>
	<title>Error {{.StatusCode}}</title>
	<style>
		body { font-family: Arial, sans-serif; margin: 40px; text-align: center; }
		.error { color: #d32f2f; }
	</style>
</head>
<body>
	<h1 class="error">Error {{.StatusCode}}</h1>
	<p>{{.Message}}</p>
	<a href="/">Go back home</a>
</body>
</html>
`,
	}

	// Load all templates
	for name, content := range templates {
		if err := templateManager.LoadTemplate(name, content); err != nil {
			log.Fatalf("Failed to load template %s: %v", name, err)
		}
	}

	// Get the router
	router := framework.Router()

	// Example 1: Using Context.HTML directly
	router.GET("/", func(ctx pkg.Context) error {
		// Set the template manager on the response writer
		ctx.Response().SetTemplateManager(templateManager)

		// Render the template
		return ctx.HTML(http.StatusOK, "index.html", map[string]interface{}{
			"Title":    "Rockstar Web Framework",
			"Username": "John Doe",
			"Email":    "john@example.com",
			"Features": []string{
				"Go template language support",
				"Context parameter passing",
				"Custom template functions",
				"Multiple template loading",
			},
		})
	})

	// Example 2: Using the View helper function
	aboutView := pkg.View("about.html", func(ctx pkg.Context) interface{} {
		return map[string]interface{}{
			"AppName":     "Rockstar Web Framework",
			"Description": "A high-performance, enterprise-grade Go web framework",
			"Version":     "1.0.0",
		}
	})

	router.GET("/about", func(ctx pkg.Context) error {
		// Set the template manager
		ctx.Response().SetTemplateManager(templateManager)

		// Execute the view
		responseFunc := aboutView(ctx)
		return responseFunc()
	})

	// Example 3: Using ViewWithStatus for error pages
	errorView := pkg.ViewWithStatus(http.StatusNotFound, "error.html", func(ctx pkg.Context) interface{} {
		return map[string]interface{}{
			"StatusCode": 404,
			"Message":    "The page you're looking for doesn't exist.",
		}
	})

	router.GET("/error", func(ctx pkg.Context) error {
		// Set the template manager
		ctx.Response().SetTemplateManager(templateManager)

		// Execute the error view
		responseFunc := errorView(ctx)
		return responseFunc()
	})

	// Example 4: Dynamic template rendering based on parameters
	router.GET("/user/:id", func(ctx pkg.Context) error {
		// Set the template manager
		ctx.Response().SetTemplateManager(templateManager)

		userID := ctx.Params()["id"]

		// In a real application, you'd fetch user data from a database
		userData := map[string]interface{}{
			"Title":    fmt.Sprintf("User Profile - %s", userID),
			"Username": fmt.Sprintf("User %s", userID),
			"Email":    fmt.Sprintf("user%s@example.com", userID),
			"Features": []string{
				"Profile management",
				"Settings",
				"Activity history",
			},
		}

		return ctx.HTML(http.StatusOK, "index.html", userData)
	})

	// Example 5: Template with conditional rendering
	conditionalTemplate := `
<!DOCTYPE html>
<html>
<head>
	<title>{{.Title}}</title>
</head>
<body>
	<h1>{{.Title}}</h1>
	{{if .IsLoggedIn}}
		<p>Welcome back, {{.Username}}!</p>
		<a href="/logout">Logout</a>
	{{else}}
		<p>Please log in to continue.</p>
		<a href="/login">Login</a>
	{{end}}
</body>
</html>
`

	if err := templateManager.LoadTemplate("conditional.html", conditionalTemplate); err != nil {
		log.Fatalf("Failed to load conditional template: %v", err)
	}

	router.GET("/dashboard", func(ctx pkg.Context) error {
		// Set the template manager
		ctx.Response().SetTemplateManager(templateManager)

		// Check if user is authenticated (simplified example)
		isLoggedIn := ctx.IsAuthenticated()

		data := map[string]interface{}{
			"Title":      "Dashboard",
			"IsLoggedIn": isLoggedIn,
		}

		if isLoggedIn {
			data["Username"] = ctx.User().Username
		}

		return ctx.HTML(http.StatusOK, "conditional.html", data)
	})

	// Start the server
	fmt.Println("Server starting on http://localhost:8080")
	fmt.Println("Available routes:")
	fmt.Println("  - http://localhost:8080/")
	fmt.Println("  - http://localhost:8080/about")
	fmt.Println("  - http://localhost:8080/error")
	fmt.Println("  - http://localhost:8080/user/123")
	fmt.Println("  - http://localhost:8080/dashboard")

	if err := framework.Listen(":8080"); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
