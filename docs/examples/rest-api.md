# REST API Example

The REST API example (`examples/rest_api_example.go`) demonstrates how to build a production-ready RESTful API with the Rockstar Web Framework. It showcases best practices for CRUD operations, request validation, pagination, filtering, rate limiting, and CORS configuration.

## What This Example Demonstrates

- **RESTful API design** following REST principles
- **CRUD operations** (Create, Read, Update, Delete)
- **Request validation** with detailed error messages
- **Pagination** for list endpoints
- **Filtering** by category and other parameters
- **Rate limiting** to prevent abuse
- **CORS configuration** for cross-origin requests
- **REST API Manager** for consistent API patterns
- **Error handling** with proper HTTP status codes
- **Health check endpoint** for monitoring

## Prerequisites

- Go 1.25 or higher
- SQLite (included with the framework)

## Setup Instructions

### Run the Example

```bash
go run examples/rest_api_example.go
```

The server will start on `http://localhost:8080` with sample product data pre-loaded.

## API Endpoints

### Products

| Method | Endpoint | Description | Query Parameters |
|--------|----------|-------------|------------------|
| GET | `/api/products` | List all products | `page`, `per_page`, `category` |
| GET | `/api/products/:id` | Get product by ID | - |
| POST | `/api/products` | Create new product | - |
| PUT | `/api/products/:id` | Update product (full) | - |
| PATCH | `/api/products/:id` | Update product (partial) | - |
| DELETE | `/api/products/:id` | Delete product | - |

### Other Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/categories` | List all categories |
| GET | `/health` | Health check |

## Testing the API

### List Products

```bash
# Get all products (first page)
curl http://localhost:8080/api/products

# Get products with pagination
curl "http://localhost:8080/api/products?page=1&per_page=10"

# Filter by category
curl "http://localhost:8080/api/products?category=Electronics"

# Combine pagination and filtering
curl "http://localhost:8080/api/products?category=Books&page=1&per_page=5"
```

**Response**:
```json
{
  "products": [
    {
      "id": 1,
      "name": "Wireless Headphones",
      "description": "High-quality wireless headphones with noise cancellation",
      "price": 149.99,
      "stock": 50,
      "category": "Electronics",
      "created_at": "2025-01-15T10:00:00Z",
      "updated_at": "2025-01-15T10:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "per_page": 10,
    "total": 3,
    "total_pages": 1
  }
}
```

### Get Product by ID

```bash
curl http://localhost:8080/api/products/1
```

**Response**:
```json
{
  "id": 1,
  "name": "Wireless Headphones",
  "description": "High-quality wireless headphones with noise cancellation",
  "price": 149.99,
  "stock": 50,
  "category": "Electronics",
  "created_at": "2025-01-15T10:00:00Z",
  "updated_at": "2025-01-15T10:00:00Z"
}
```

### Create Product

```bash
curl -X POST http://localhost:8080/api/products \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Mechanical Keyboard",
    "description": "RGB mechanical keyboard with Cherry MX switches",
    "price": 129.99,
    "stock": 75,
    "category": "Electronics"
  }'
```

**Response** (201 Created):
```json
{
  "id": 4,
  "name": "Mechanical Keyboard",
  "description": "RGB mechanical keyboard with Cherry MX switches",
  "price": 129.99,
  "stock": 75,
  "category": "Electronics",
  "created_at": "2025-01-15T10:05:00Z",
  "updated_at": "2025-01-15T10:05:00Z"
}
```

### Update Product (Full)

```bash
curl -X PUT http://localhost:8080/api/products/1 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Premium Wireless Headphones",
    "description": "Updated description",
    "price": 179.99,
    "stock": 45,
    "category": "Electronics"
  }'
```

**Note**: PUT requires all fields to be provided.

### Update Product (Partial)

```bash
curl -X PATCH http://localhost:8080/api/products/1 \
  -H "Content-Type: application/json" \
  -d '{
    "price": 139.99,
    "stock": 60
  }'
```

**Note**: PATCH allows updating only specific fields.

### Delete Product

```bash
curl -X DELETE http://localhost:8080/api/products/1
```

**Response**:
```json
{
  "deleted": true,
  "id": 1
}
```

### List Categories

```bash
curl http://localhost:8080/api/categories
```

**Response**:
```json
{
  "categories": ["Electronics", "Books", "Clothing", "Food"]
}
```

### Health Check

```bash
curl http://localhost:8080/health
```

**Response**:
```json
{
  "status": "healthy",
  "timestamp": "2025-01-15T10:00:00Z",
  "version": "1.0.0",
  "products": 3
}
```

## Code Walkthrough

### REST API Manager

The example uses the `RESTAPIManager` for consistent API patterns:

```go
restAPI := pkg.NewRESTAPIManager(router, db)
```

This manager provides:
- Consistent JSON response formatting
- Error response standardization
- Request parsing utilities
- Rate limiting integration
- CORS handling

### Route Configuration

Configure API routes with rate limiting and CORS:

```go
apiConfig := pkg.RESTRouteConfig{
    RateLimit: &pkg.RESTRateLimitConfig{
        Limit:  100,         // 100 requests
        Window: time.Minute, // per minute
        Key:    "ip_address",
    },
    MaxRequestSize: 1024 * 1024, // 1MB max
    Timeout:        30 * time.Second,
    CORS: &pkg.CORSConfig{
        AllowOrigins:     []string{"*"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
        AllowHeaders:     []string{"Content-Type", "Authorization"},
        AllowCredentials: false,
        MaxAge:           3600,
    },
}
```

**Configuration Options**:
- **RateLimit**: Limit requests per IP address
- **MaxRequestSize**: Prevent large payloads
- **Timeout**: Request timeout duration
- **CORS**: Cross-origin resource sharing settings

### Registering Routes

Register routes with the REST API manager:

```go
restAPI.RegisterRoute("GET", "/api/products", listProducts(restAPI), apiConfig)
restAPI.RegisterRoute("POST", "/api/products", createProduct(restAPI), apiConfig)
```

### Handler Pattern

REST handlers use the `RESTHandler` type:

```go
func listProducts(restAPI pkg.RESTAPIManager) pkg.RESTHandler {
    return func(ctx pkg.Context) error {
        // Handler logic
        return restAPI.SendJSONResponse(ctx, 200, data)
    }
}
```

### Pagination Implementation

The list endpoint implements pagination:

```go
// Parse pagination parameters
page := 1
perPage := 10
if pageStr := query["page"]; pageStr != "" {
    if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
        page = p
    }
}

// Calculate pagination
total := len(filtered)
totalPages := (total + perPage - 1) / perPage
start := (page - 1) * perPage
end := start + perPage

// Get page of results
pageResults := filtered[start:end]

// Build response with metadata
response := map[string]interface{}{
    "products": pageResults,
    "pagination": map[string]interface{}{
        "page":        page,
        "per_page":    perPage,
        "total":       total,
        "total_pages": totalPages,
    },
}
```

**Pagination Best Practices**:
- Default to reasonable page size (10-50 items)
- Limit maximum page size (e.g., 100)
- Include pagination metadata in response
- Handle edge cases (empty results, invalid page numbers)

### Filtering Implementation

Filter results based on query parameters:

```go
category := query["category"]

var filtered []*Product
for _, product := range products {
    if category == "" || product.Category == category {
        filtered = append(filtered, product)
    }
}
```

**Filtering Best Practices**:
- Support multiple filter parameters
- Use exact matches or pattern matching as appropriate
- Combine filters with AND logic
- Document available filters in API docs

### Request Validation

Validate input before processing:

```go
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
```

**Validation Best Practices**:
- Validate all required fields
- Check data types and ranges
- Provide specific error messages
- Return 400 Bad Request for validation errors

### Error Responses

Use consistent error response format:

```go
return restAPI.SendErrorResponse(ctx, 404, "Product not found", map[string]interface{}{
    "id": id,
})
```

**Error Response Format**:
```json
{
  "error": "Product not found",
  "details": {
    "id": 123
  }
}
```

### PUT vs PATCH

The example demonstrates both full and partial updates:

**PUT** - Full update (all fields required):
```go
func updateProduct(restAPI pkg.RESTAPIManager) pkg.RESTHandler {
    return func(ctx pkg.Context) error {
        // Parse all fields
        // Validate all fields
        // Replace entire resource
        product.Name = input.Name
        product.Description = input.Description
        product.Price = input.Price
        // ... update all fields
    }
}
```

**PATCH** - Partial update (only provided fields):
```go
func patchProduct(restAPI pkg.RESTAPIManager) pkg.RESTHandler {
    return func(ctx pkg.Context) error {
        // Parse as map for optional fields
        var input map[string]interface{}
        
        // Update only provided fields
        if name, ok := input["name"].(string); ok && name != "" {
            product.Name = name
        }
        if price, ok := input["price"].(float64); ok && price > 0 {
            product.Price = price
        }
    }
}
```

## RESTful Design Principles

This example follows REST best practices:

### Resource-Oriented URLs

```
GET    /api/products      - Collection
POST   /api/products      - Create in collection
GET    /api/products/:id  - Individual resource
PUT    /api/products/:id  - Replace resource
PATCH  /api/products/:id  - Update resource
DELETE /api/products/:id  - Delete resource
```

### HTTP Methods

- **GET**: Retrieve resources (safe, idempotent)
- **POST**: Create new resources
- **PUT**: Replace entire resource (idempotent)
- **PATCH**: Partial update (idempotent)
- **DELETE**: Remove resource (idempotent)

### Status Codes

- **200 OK**: Successful GET, PUT, PATCH, DELETE
- **201 Created**: Successful POST
- **400 Bad Request**: Invalid input
- **404 Not Found**: Resource doesn't exist
- **500 Internal Server Error**: Server error

### Response Format

Consistent JSON structure:
```json
{
  "data": { ... },           // For single resource
  "items": [ ... ],          // For collections
  "pagination": { ... },     // For paginated lists
  "error": "message",        // For errors
  "details": { ... }         // Additional error info
}
```

## Rate Limiting

The example implements rate limiting per IP address:

```go
RateLimit: &pkg.RESTRateLimitConfig{
    Limit:  100,         // 100 requests
    Window: time.Minute, // per minute
    Key:    "ip_address",
}
```

**Rate Limit Behavior**:
- Tracks requests per IP address
- Returns 429 Too Many Requests when limit exceeded
- Includes `X-RateLimit-*` headers in responses
- Resets after the time window

**Rate Limit Headers**:
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1642252800
```

## CORS Configuration

Enable cross-origin requests:

```go
CORS: &pkg.CORSConfig{
    AllowOrigins:     []string{"*"},
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
    AllowHeaders:     []string{"Content-Type", "Authorization"},
    AllowCredentials: false,
    MaxAge:           3600,
}
```

**CORS Headers**:
- `Access-Control-Allow-Origin`: Allowed origins
- `Access-Control-Allow-Methods`: Allowed HTTP methods
- `Access-Control-Allow-Headers`: Allowed request headers
- `Access-Control-Max-Age`: Preflight cache duration

## Production Considerations

### Database Integration

Replace in-memory storage with database:

```go
// Create product
result, err := ctx.DB().Exec(
    "INSERT INTO products (name, description, price, stock, category) VALUES (?, ?, ?, ?, ?)",
    input.Name, input.Description, input.Price, input.Stock, input.Category,
)

// Get product
var product Product
err := ctx.DB().QueryRow(
    "SELECT id, name, description, price, stock, category, created_at, updated_at FROM products WHERE id = ?",
    id,
).Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.Stock, &product.Category, &product.CreatedAt, &product.UpdatedAt)
```

### Authentication

Add authentication middleware:

```go
authAPI := router.Group("/api", authenticationMiddleware)
authAPI.POST("/products", createProduct(restAPI), apiConfig)
authAPI.PUT("/products/:id", updateProduct(restAPI), apiConfig)
authAPI.DELETE("/products/:id", deleteProduct(restAPI), apiConfig)
```

### Input Sanitization

Sanitize user input to prevent injection attacks:

```go
import "html"

input.Name = html.EscapeString(input.Name)
input.Description = html.EscapeString(input.Description)
```

### Caching

Cache frequently accessed resources:

```go
// Try cache first
cacheKey := fmt.Sprintf("product:%d", id)
if cached, err := ctx.Cache().Get(cacheKey); err == nil {
    return ctx.JSON(200, cached)
}

// Fetch from database
product := fetchProduct(id)

// Store in cache
ctx.Cache().Set(cacheKey, product, 5*time.Minute)

return ctx.JSON(200, product)
```

### Logging

Add structured logging:

```go
logger := ctx.Logger()
logger.Info("Product created",
    "product_id", product.ID,
    "user_id", ctx.User().ID,
    "ip", ctx.Request().RemoteAddr,
)
```

## Testing the API

### Manual Testing

Use curl or tools like Postman, Insomnia, or HTTPie:

```bash
# HTTPie examples
http GET localhost:8080/api/products
http POST localhost:8080/api/products name="New Product" price:=29.99
http PUT localhost:8080/api/products/1 name="Updated" price:=39.99
http DELETE localhost:8080/api/products/1
```

### Automated Testing

Write integration tests:

```go
func TestListProducts(t *testing.T) {
    // Setup test server
    // Make request
    // Assert response
}
```

## Common Issues

### "Invalid JSON"

**Cause**: Malformed JSON in request body

**Solution**: Validate JSON syntax, ensure proper Content-Type header

### "Missing required field"

**Cause**: Required field not provided in request

**Solution**: Include all required fields in request body

### "Invalid category"

**Cause**: Category not in allowed list

**Solution**: Use one of: Electronics, Books, Clothing, Food

### Rate limit exceeded

**Cause**: Too many requests from same IP

**Solution**: Wait for rate limit window to reset, or implement backoff

## Next Steps

After understanding this example:

1. **Add authentication**: Implement JWT or OAuth2 authentication
2. **Add database**: Replace in-memory storage with PostgreSQL/MySQL
3. **Add caching**: Cache frequently accessed products
4. **Add search**: Implement full-text search
5. **Add sorting**: Allow sorting by price, name, date
6. **Study advanced patterns**: Review [Full Featured App](full-featured-app.md)

## Related Documentation

- [API Styles Guide](../guides/api-styles.md) - REST, GraphQL, gRPC, SOAP
- [Router API](../api/router.md) - Routing reference
- [Context API](../api/context.md) - Context interface
- [Security Guide](../guides/security.md) - Authentication and authorization
- [Database Guide](../guides/database.md) - Database integration

## Source Code

The complete source code for this example is available at `examples/rest_api_example.go` in the repository.
