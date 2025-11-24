package pkg

/*
Framework Data Models Documentation

This file provides documentation for the core data models used in the Rockstar Web Framework.
All models support JSON serialization/deserialization and database persistence.

================================================================================
SESSION MODEL
================================================================================

The Session model represents a user session stored in the database, cache, or filesystem.
Sessions are used to maintain user state across requests in a distributed environment.

Fields:
  - ID:        Unique session identifier (typically a UUID or secure random string)
  - UserID:    ID of the user associated with this session
  - TenantID:  ID of the tenant (for multi-tenancy support)
  - Data:      Arbitrary session data stored as key-value pairs
  - ExpiresAt: Timestamp when the session expires
  - CreatedAt: Timestamp when the session was created
  - UpdatedAt: Timestamp when the session was last updated
  - IPAddress: IP address of the client that created the session
  - UserAgent: User agent string of the client

Usage Example:
  session := &Session{
      ID:        "sess_abc123",
      UserID:    "user_456",
      TenantID:  "tenant_789",
      Data:      map[string]interface{}{"cart": []string{"item1", "item2"}},
      ExpiresAt: time.Now().Add(24 * time.Hour),
      IPAddress: "192.168.1.1",
      UserAgent: "Mozilla/5.0...",
  }

  // Save to database
  err := dbManager.SaveSession(session)

  // Load from database
  loadedSession, err := dbManager.LoadSession("sess_abc123")

Database Schema:
  - Primary Key: id
  - Indexes: expires_at, user_id, tenant_id
  - Storage: MySQL, PostgreSQL, MSSQL, SQLite

Requirements Satisfied: 4.1, 4.2, 4.3, 4.4, 4.5, 4.6, 10.5

================================================================================
ACCESS TOKEN MODEL
================================================================================

The AccessToken model represents an API access token for authentication.
Tokens are used to authenticate API requests across GraphQL, REST, gRPC, and SOAP protocols.

Fields:
  - Token:     The actual token string (should be cryptographically secure)
  - UserID:    ID of the user this token belongs to
  - TenantID:  ID of the tenant (for multi-tenancy support)
  - Scopes:    Array of permission scopes granted to this token
  - ExpiresAt: Timestamp when the token expires
  - CreatedAt: Timestamp when the token was created

Usage Example:
  token := &AccessToken{
      Token:     "tok_xyz789",
      UserID:    "user_456",
      TenantID:  "tenant_789",
      Scopes:    []string{"read:users", "write:posts", "admin:settings"},
      ExpiresAt: time.Now().Add(1 * time.Hour),
  }

  // Save to database
  err := dbManager.SaveAccessToken(token)

  // Validate token
  validToken, err := dbManager.ValidateAccessToken("tok_xyz789")
  if err != nil {
      // Token is invalid or expired
  }

Database Schema:
  - Primary Key: token
  - Indexes: expires_at, user_id, tenant_id
  - Storage: MySQL, PostgreSQL, MSSQL, SQLite

Requirements Satisfied: 3.5, 3.6, 2.5, 2.6, 10.5

================================================================================
TENANT MODEL
================================================================================

The Tenant model represents a tenant in the multi-tenant architecture.
Each tenant can have multiple hosts, custom configuration, and resource limits.

Fields:
  - ID:          Unique tenant identifier
  - Name:        Human-readable tenant name
  - Hosts:       Array of hostnames associated with this tenant
  - Config:      Tenant-specific configuration as key-value pairs
  - IsActive:    Whether the tenant is currently active
  - CreatedAt:   Timestamp when the tenant was created
  - UpdatedAt:   Timestamp when the tenant was last updated
  - MaxUsers:    Maximum number of users allowed for this tenant
  - MaxStorage:  Maximum storage in bytes allowed for this tenant
  - MaxRequests: Maximum number of requests per time period

Usage Example:
  tenant := &Tenant{
      ID:          "tenant_123",
      Name:        "Acme Corporation",
      Hosts:       []string{"acme.example.com", "www.acme.example.com"},
      Config:      map[string]interface{}{"theme": "dark", "language": "en"},
      IsActive:    true,
      MaxUsers:    1000,
      MaxStorage:  10737418240, // 10GB
      MaxRequests: 100000,
  }

  // Save to database
  err := dbManager.SaveTenant(tenant)

  // Load by ID
  loadedTenant, err := dbManager.LoadTenant("tenant_123")

  // Load by hostname
  tenantByHost, err := dbManager.LoadTenantByHost("acme.example.com")

Database Schema:
  - Primary Key: id
  - Indexes: is_active
  - Storage: MySQL, PostgreSQL, MSSQL, SQLite
  - Note: hosts field uses JSON storage for array support

Requirements Satisfied: 8.1, 8.2, 8.6, 10.5

================================================================================
WORKLOAD METRICS MODEL
================================================================================

The WorkloadMetrics model represents performance and usage metrics for monitoring.
Metrics are collected for each request to track system performance and resource usage.

Fields:
  - ID:            Auto-incrementing unique identifier
  - Timestamp:     When the metric was recorded
  - TenantID:      ID of the tenant (for multi-tenancy support)
  - UserID:        ID of the user who made the request (optional)
  - RequestID:     Unique identifier for the request
  - Duration:      Request duration in milliseconds
  - ContextSize:   Size of the request context in bytes
  - MemoryUsage:   Memory used by the request in bytes
  - CPUUsage:      CPU usage percentage during the request
  - Path:          Request path (e.g., "/api/users")
  - Method:        HTTP method (GET, POST, etc.)
  - StatusCode:    HTTP status code returned
  - ResponseSize:  Size of the response in bytes
  - ErrorMessage:  Error message if the request failed (empty if successful)

Usage Example:
  metrics := &WorkloadMetrics{
      Timestamp:    time.Now(),
      TenantID:     "tenant_123",
      UserID:       "user_456",
      RequestID:    "req_789",
      Duration:     150,
      ContextSize:  2048,
      MemoryUsage:  4096000,
      CPUUsage:     45.5,
      Path:         "/api/users",
      Method:       "GET",
      StatusCode:   200,
      ResponseSize: 1024,
  }

  // Save to database
  err := dbManager.SaveWorkloadMetrics(metrics)

  // Query metrics for a tenant
  from := time.Now().Add(-24 * time.Hour)
  to := time.Now()
  metricsList, err := dbManager.GetWorkloadMetrics("tenant_123", from, to)

Database Schema:
  - Primary Key: id (auto-increment)
  - Indexes: timestamp, tenant_id, user_id
  - Storage: MySQL, PostgreSQL, MSSQL, SQLite

Requirements Satisfied: 14.3, 14.4, 14.5, 14.6, 17.4, 10.5

================================================================================
BEST PRACTICES
================================================================================

1. Session Management:
   - Always set appropriate expiration times for sessions
   - Clean up expired sessions regularly using CleanupExpiredSessions()
   - Use AES-encrypted cookies for session IDs
   - Store sensitive data in the database, not in cookies

2. Access Tokens:
   - Use cryptographically secure random strings for tokens
   - Set appropriate expiration times based on security requirements
   - Clean up expired tokens regularly using CleanupExpiredTokens()
   - Use scopes to implement fine-grained access control

3. Multi-Tenancy:
   - Always validate tenant_id in all operations
   - Use LoadTenantByHost() for host-based routing
   - Enforce resource limits (MaxUsers, MaxStorage, MaxRequests)
   - Keep tenant configuration flexible using the Config map

4. Workload Metrics:
   - Collect metrics asynchronously to avoid impacting request performance
   - Aggregate metrics for reporting and analysis
   - Use metrics for capacity planning and optimization
   - Monitor error rates and response times

5. Database Operations:
   - Use transactions for operations that modify multiple records
   - Handle database errors gracefully with proper error messages
   - Use connection pooling for optimal performance
   - Implement proper indexing for query performance

================================================================================
*/
