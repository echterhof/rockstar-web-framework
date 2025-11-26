package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

// 🎸 REST API Example
// This example demonstrates RESTful API design with the Rockstar Web Framework
// Features: CRUD operations, request validation, pagination, filtering, rate limiting

// Product represents a product in our API
type Product struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	Category    string    `json:"category"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// In-memory storage for demonstration
var (
	products     = make(map[int]*Product)
	productIDSeq = 1
	categories   = []string{"Electronics", "Books", "Clothing", "Food"}
)

func main() {
	// Initialize sample data
	initSampleProducts()

	// Create framework configuration
	config := pkg.FrameworkConfig{
		ServerConfig: pkg.ServerConfig{
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			EnableHTTP1:  true,
			EnableHTTP2:  true,
		},
		DatabaseConfig: pkg.DatabaseConfig{
			Driver:   "sqlite",
			Database: ":memory:",
		},
	}

	// Create framework instance
	app, err := pkg.New(config)
	if err != nil {
		log.Fatalf("Failed to create framework: %v", err)
	}

	// Get router and database
	router := app.Router()
	db := app.Database()

	// Initialize database tables
	if err := db.CreateTables(); err != nil {
		log.Fatalf("Failed to create database tables: %v", err)
	}

	// Create REST API manager
	restAPI := pkg.NewRESTAPIManager(router, db)

	// Configure rate limiting and CORS for API routes
	apiConfig := pkg.RESTRouteConfig{
		RateLimit: &pkg.RESTRateLimitConfig{
			Limit:  100,         // 100 requests
			Window: time.Minute, // per minute
			Key:    "ip_address",
		},
		MaxRequestSize: 1024 * 1024, // 1MB max request size
		Timeout:        30 * time.Second,
		CORS: &pkg.CORSConfig{
			AllowOrigins:     []string{"*"},
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
			AllowHeaders:     []string{"Content-Type", "Authorization"},
			AllowCredentials: false,
			MaxAge:           3600,
		},
	}

	// Register REST API routes

	// GET /api/products - List products with pagination and filtering
	restAPI.RegisterRoute("GET", "/api/products", listProducts(restAPI), apiConfig)

	// GET /api/products/:id - Get product by ID
	restAPI.RegisterRoute("GET", "/api/products/:id", getProduct(restAPI), apiConfig)

	// POST /api/products - Create new product
	restAPI.RegisterRoute("POST", "/api/products", createProduct(restAPI), apiConfig)

	// PUT /api/products/:id - Update product
	restAPI.RegisterRoute("PUT", "/api/products/:id", updateProduct(restAPI), apiConfig)

	// PATCH /api/products/:id - Partial update product
	restAPI.RegisterRoute("PATCH", "/api/products/:id", patchProduct(restAPI), apiConfig)

	// DELETE /api/products/:id - Delete product
	restAPI.RegisterRoute("DELETE", "/api/products/:id", deleteProduct(restAPI), apiConfig)

	// GET /api/categories - List categories
	restAPI.RegisterRoute("GET", "/api/categories", listCategories(restAPI), apiConfig)

	// Health check endpoint (no rate limiting)
	restAPI.RegisterRoute("GET", "/health", healthCheck(restAPI), pkg.RESTRouteConfig{})

	// Print startup information
	fmt.Println("🎸 REST API Example")
	fmt.Println("==================")
	fmt.Println()
	fmt.Println("Server listening on http://localhost:8080")
	fmt.Println()
	fmt.Println("Available endpoints:")
	fmt.Println("  GET    /api/products          - List products (supports ?page, ?per_page, ?category)")
	fmt.Println("  GET    /api/products/:id      - Get product by ID")
	fmt.Println("  POST   /api/products          - Create new product")
	fmt.Println("  PUT    /api/products/:id      - Update product")
	fmt.Println("  PATCH  /api/products/:id      - Partial update product")
	fmt.Println("  DELETE /api/products/:id      - Delete product")
	fmt.Println("  GET    /api/categories        - List categories")
	fmt.Println("  GET    /health                - Health check")
	fmt.Println()
	fmt.Println("Example requests:")
	fmt.Println("  curl http://localhost:8080/api/products")
	fmt.Println("  curl http://localhost:8080/api/products/1")
	fmt.Println("  curl http://localhost:8080/api/products?category=Electronics&page=1&per_page=10")
	fmt.Println("  curl -X POST http://localhost:8080/api/products -H 'Content-Type: application/json' \\")
	fmt.Println("       -d '{\"name\":\"New Product\",\"price\":29.99,\"stock\":100,\"category\":\"Electronics\"}'")
	fmt.Println()
	fmt.Println("Rate limits:")
	fmt.Println("  API routes: 100 requests/minute per IP")
	fmt.Println()

	// Start server
	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// listProducts returns a handler that lists products with pagination and filtering
func listProducts(restAPI pkg.RESTAPIManager) pkg.RESTHandler {
	return func(ctx pkg.Context) error {
		query := ctx.Query()

		// Parse pagination parameters
		page := 1
		perPage := 10
		if pageStr := query["page"]; pageStr != "" {
			if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
				page = p
			}
		}
		if perPageStr := query["per_page"]; perPageStr != "" {
			if pp, err := strconv.Atoi(perPageStr); err == nil && pp > 0 && pp <= 100 {
				perPage = pp
			}
		}

		// Filter by category if provided
		category := query["category"]

		// Collect matching products
		var filtered []*Product
		for _, product := range products {
			if category == "" || product.Category == category {
				filtered = append(filtered, product)
			}
		}

		// Calculate pagination
		total := len(filtered)
		totalPages := (total + perPage - 1) / perPage
		start := (page - 1) * perPage
		end := start + perPage
		if end > total {
			end = total
		}
		if start > total {
			start = total
		}

		// Get page of results
		var pageResults []*Product
		if start < total {
			pageResults = filtered[start:end]
		}

		// Build response with pagination metadata
		response := map[string]interface{}{
			"products": pageResults,
			"pagination": map[string]interface{}{
				"page":        page,
				"per_page":    perPage,
				"total":       total,
				"total_pages": totalPages,
			},
		}

		return restAPI.SendJSONResponse(ctx, 200, response)
	}
}

// getProduct returns a handler that gets a product by ID
func getProduct(restAPI pkg.RESTAPIManager) pkg.RESTHandler {
	return func(ctx pkg.Context) error {
		// Get product ID from URL parameters
		idStr := ctx.Params()["id"]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return restAPI.SendErrorResponse(ctx, 400, "Invalid product ID", map[string]interface{}{
				"id": idStr,
			})
		}

		// Find product
		product, exists := products[id]
		if !exists {
			return restAPI.SendErrorResponse(ctx, 404, "Product not found", map[string]interface{}{
				"id": id,
			})
		}

		return restAPI.SendJSONResponse(ctx, 200, product)
	}
}

// createProduct returns a handler that creates a new product
func createProduct(restAPI pkg.RESTAPIManager) pkg.RESTHandler {
	return func(ctx pkg.Context) error {
		// Parse request body
		var input struct {
			Name        string  `json:"name"`
			Description string  `json:"description"`
			Price       float64 `json:"price"`
			Stock       int     `json:"stock"`
			Category    string  `json:"category"`
		}

		if err := restAPI.ParseJSONRequest(ctx, &input); err != nil {
			return restAPI.SendErrorResponse(ctx, 400, "Invalid JSON", map[string]interface{}{
				"error": err.Error(),
			})
		}

		// Validate required fields
		if input.Name == "" {
			return restAPI.SendErrorResponse(ctx, 400, "Missing required field: name", nil)
		}
		if input.Price <= 0 {
			return restAPI.SendErrorResponse(ctx, 400, "Price must be greater than 0", nil)
		}
		if input.Stock < 0 {
			return restAPI.SendErrorResponse(ctx, 400, "Stock cannot be negative", nil)
		}
		if input.Category == "" {
			return restAPI.SendErrorResponse(ctx, 400, "Missing required field: category", nil)
		}

		// Validate category
		validCategory := false
		for _, cat := range categories {
			if cat == input.Category {
				validCategory = true
				break
			}
		}
		if !validCategory {
			return restAPI.SendErrorResponse(ctx, 400, "Invalid category", map[string]interface{}{
				"category":         input.Category,
				"valid_categories": categories,
			})
		}

		// Create product
		product := &Product{
			ID:          productIDSeq,
			Name:        input.Name,
			Description: input.Description,
			Price:       input.Price,
			Stock:       input.Stock,
			Category:    input.Category,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		products[productIDSeq] = product
		productIDSeq++

		return restAPI.SendJSONResponse(ctx, 201, product)
	}
}

// updateProduct returns a handler that updates a product
func updateProduct(restAPI pkg.RESTAPIManager) pkg.RESTHandler {
	return func(ctx pkg.Context) error {
		// Get product ID
		idStr := ctx.Params()["id"]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return restAPI.SendErrorResponse(ctx, 400, "Invalid product ID", map[string]interface{}{
				"id": idStr,
			})
		}

		// Find product
		product, exists := products[id]
		if !exists {
			return restAPI.SendErrorResponse(ctx, 404, "Product not found", map[string]interface{}{
				"id": id,
			})
		}

		// Parse request body
		var input struct {
			Name        string  `json:"name"`
			Description string  `json:"description"`
			Price       float64 `json:"price"`
			Stock       int     `json:"stock"`
			Category    string  `json:"category"`
		}

		if err := restAPI.ParseJSONRequest(ctx, &input); err != nil {
			return restAPI.SendErrorResponse(ctx, 400, "Invalid JSON", map[string]interface{}{
				"error": err.Error(),
			})
		}

		// Validate fields
		if input.Name == "" {
			return restAPI.SendErrorResponse(ctx, 400, "Missing required field: name", nil)
		}
		if input.Price <= 0 {
			return restAPI.SendErrorResponse(ctx, 400, "Price must be greater than 0", nil)
		}
		if input.Stock < 0 {
			return restAPI.SendErrorResponse(ctx, 400, "Stock cannot be negative", nil)
		}

		// Update product
		product.Name = input.Name
		product.Description = input.Description
		product.Price = input.Price
		product.Stock = input.Stock
		product.Category = input.Category
		product.UpdatedAt = time.Now()

		return restAPI.SendJSONResponse(ctx, 200, product)
	}
}

// patchProduct returns a handler that partially updates a product
func patchProduct(restAPI pkg.RESTAPIManager) pkg.RESTHandler {
	return func(ctx pkg.Context) error {
		// Get product ID
		idStr := ctx.Params()["id"]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return restAPI.SendErrorResponse(ctx, 400, "Invalid product ID", map[string]interface{}{
				"id": idStr,
			})
		}

		// Find product
		product, exists := products[id]
		if !exists {
			return restAPI.SendErrorResponse(ctx, 404, "Product not found", map[string]interface{}{
				"id": id,
			})
		}

		// Parse request body (all fields optional)
		var input map[string]interface{}
		if err := restAPI.ParseJSONRequest(ctx, &input); err != nil {
			return restAPI.SendErrorResponse(ctx, 400, "Invalid JSON", map[string]interface{}{
				"error": err.Error(),
			})
		}

		// Update only provided fields
		if name, ok := input["name"].(string); ok && name != "" {
			product.Name = name
		}
		if description, ok := input["description"].(string); ok {
			product.Description = description
		}
		if price, ok := input["price"].(float64); ok && price > 0 {
			product.Price = price
		}
		if stock, ok := input["stock"].(float64); ok && stock >= 0 {
			product.Stock = int(stock)
		}
		if category, ok := input["category"].(string); ok && category != "" {
			product.Category = category
		}

		product.UpdatedAt = time.Now()

		return restAPI.SendJSONResponse(ctx, 200, product)
	}
}

// deleteProduct returns a handler that deletes a product
func deleteProduct(restAPI pkg.RESTAPIManager) pkg.RESTHandler {
	return func(ctx pkg.Context) error {
		// Get product ID
		idStr := ctx.Params()["id"]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return restAPI.SendErrorResponse(ctx, 400, "Invalid product ID", map[string]interface{}{
				"id": idStr,
			})
		}

		// Find product
		_, exists := products[id]
		if !exists {
			return restAPI.SendErrorResponse(ctx, 404, "Product not found", map[string]interface{}{
				"id": id,
			})
		}

		// Delete product
		delete(products, id)

		return restAPI.SendJSONResponse(ctx, 200, map[string]interface{}{
			"deleted": true,
			"id":      id,
		})
	}
}

// listCategories returns a handler that lists all categories
func listCategories(restAPI pkg.RESTAPIManager) pkg.RESTHandler {
	return func(ctx pkg.Context) error {
		return restAPI.SendJSONResponse(ctx, 200, map[string]interface{}{
			"categories": categories,
		})
	}
}

// healthCheck returns a handler for health check endpoint
func healthCheck(restAPI pkg.RESTAPIManager) pkg.RESTHandler {
	return func(ctx pkg.Context) error {
		health := map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now(),
			"version":   "1.0.0",
			"products":  len(products),
		}
		return restAPI.SendJSONResponse(ctx, 200, health)
	}
}

// initSampleProducts initializes sample product data
func initSampleProducts() {
	products[productIDSeq] = &Product{
		ID:          productIDSeq,
		Name:        "Wireless Headphones",
		Description: "High-quality wireless headphones with noise cancellation",
		Price:       149.99,
		Stock:       50,
		Category:    "Electronics",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	productIDSeq++

	products[productIDSeq] = &Product{
		ID:          productIDSeq,
		Name:        "Programming Book",
		Description: "Learn Go programming from scratch",
		Price:       39.99,
		Stock:       100,
		Category:    "Books",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	productIDSeq++

	products[productIDSeq] = &Product{
		ID:          productIDSeq,
		Name:        "T-Shirt",
		Description: "Comfortable cotton t-shirt",
		Price:       19.99,
		Stock:       200,
		Category:    "Clothing",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	productIDSeq++
}
