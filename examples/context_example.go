package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
	fmt.Println("Rockstar Web Framework - Context Manager Example")
	fmt.Println("================================================")

	// Create a sample HTTP request
	u, _ := url.Parse("/users/123?filter=active&page=1")
	req := &pkg.Request{
		Method: "GET",
		URL:    u,
		Header: http.Header{
			"Content-Type":  []string{"application/json"},
			"Authorization": []string{"Bearer abc123"},
			"User-Agent":    []string{"RockstarFramework/1.0"},
		},
		Params: map[string]string{
			"id": "123",
		},
		Query: map[string]string{
			"sort": "name",
		},
		Form: map[string]string{
			"username": "john_doe",
			"email":    "john@example.com",
		},
		ID:       "req-12345",
		TenantID: "tenant-abc",
		UserID:   "user-456",
	}

	// Create a response writer
	w := httptest.NewRecorder()
	resp := pkg.NewResponseWriter(w)

	// Create the context
	baseCtx := context.Background()
	ctx := pkg.NewContext(req, resp, baseCtx)

	// Demonstrate context functionality
	demonstrateContextUsage(ctx)

	fmt.Println("\nResponse Status:", resp.Status())
	fmt.Println("Response Body:", w.Body.String())
}

func demonstrateContextUsage(ctx pkg.Context) {
	fmt.Println("\n1. Request Information:")
	fmt.Printf("   Method: %s\n", ctx.Request().Method)
	fmt.Printf("   URL: %s\n", ctx.Request().URL.String())
	fmt.Printf("   Request ID: %s\n", ctx.Request().ID)
	fmt.Printf("   Tenant ID: %s\n", ctx.Request().TenantID)
	fmt.Printf("   User ID: %s\n", ctx.Request().UserID)

	fmt.Println("\n2. Route Parameters:")
	params := ctx.Params()
	for key, value := range params {
		fmt.Printf("   %s = %s\n", key, value)
	}

	fmt.Println("\n3. Query Parameters:")
	query := ctx.Query()
	for key, value := range query {
		fmt.Printf("   %s = %s\n", key, value)
	}

	fmt.Println("\n4. Headers:")
	headers := ctx.Headers()
	for key, value := range headers {
		fmt.Printf("   %s = %s\n", key, value)
	}

	fmt.Println("\n5. Form Data:")
	fmt.Printf("   Username: %s\n", ctx.FormValue("username"))
	fmt.Printf("   Email: %s\n", ctx.FormValue("email"))

	fmt.Println("\n6. Header Access (case-insensitive):")
	fmt.Printf("   Content-Type: %s\n", ctx.GetHeader("content-type"))
	fmt.Printf("   Authorization: %s\n", ctx.GetHeader("AUTHORIZATION"))

	fmt.Println("\n7. Authentication Status:")
	fmt.Printf("   Is Authenticated: %t\n", ctx.IsAuthenticated())

	fmt.Println("\n8. Context Control:")
	// Demonstrate timeout context
	timeoutCtx := ctx.WithTimeout(5 * 1000000000) // 5 seconds
	deadline, hasDeadline := timeoutCtx.Context().Deadline()
	fmt.Printf("   Has Timeout: %t\n", hasDeadline)
	if hasDeadline {
		fmt.Printf("   Deadline: %s\n", deadline.Format("15:04:05"))
	}

	// Demonstrate cancel context
	cancelCtx, cancel := ctx.WithCancel()
	fmt.Printf("   Context Cancelled (before): %t\n", cancelCtx.Context().Err() != nil)
	cancel()
	fmt.Printf("   Context Cancelled (after): %t\n", cancelCtx.Context().Err() != nil)

	fmt.Println("\n9. Response Operations:")
	// Set some headers
	ctx.SetHeader("X-Request-ID", ctx.Request().ID)
	ctx.SetHeader("X-Tenant-ID", ctx.Request().TenantID)

	// Set a cookie
	cookie := &pkg.Cookie{
		Name:     "session",
		Value:    "session-token-xyz",
		Path:     "/",
		HttpOnly: true,
	}
	ctx.SetCookie(cookie)

	// Send JSON response
	responseData := map[string]interface{}{
		"message":    "Context manager working correctly!",
		"request_id": ctx.Request().ID,
		"tenant_id":  ctx.Request().TenantID,
		"params":     ctx.Params(),
		"query":      ctx.Query(),
	}

	err := ctx.JSON(200, responseData)
	if err != nil {
		fmt.Printf("   Error sending JSON response: %v\n", err)
	} else {
		fmt.Println("   JSON response sent successfully")
	}
}
