# GraphQL API Example

The GraphQL API example (`examples/graphql_example.go`) demonstrates how to build a GraphQL API with the Rockstar Web Framework. It showcases schema definition, queries, mutations, introspection, and the GraphQL Playground for interactive testing.

## What This Example Demonstrates

- **GraphQL schema** definition and implementation
- **Query resolvers** for data retrieval
- **Mutation resolvers** for data modification
- **Schema introspection** for tooling support
- **GraphQL Playground** for interactive testing
- **Rate limiting** for API protection
- **Error handling** with GraphQL-compliant responses
- **Request validation** and size limits

## Prerequisites

- Go 1.25 or higher
- SQLite (included with the framework)

## Setup Instructions

### Run the Example

```bash
go run examples/graphql_example.go
```

The server will start on `http://localhost:8080` with sample blog data pre-loaded.

## GraphQL Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/graphql` | GraphQL API endpoint |
| GET | `/graphql` | GraphQL Playground (interactive UI) |

## Testing the API

### Using GraphQL Playground

Open `http://localhost:8080/graphql` in your browser to access the interactive GraphQL Playground. This provides:
- Query editor with syntax highlighting
- Auto-completion based on schema
- Query history
- Documentation explorer
- Response viewer

### Using curl

```bash
# List all users
curl -X POST http://localhost:8080/graphql \
  -H 'Content-Type: application/json' \
  -d '{"query": "{ users { id name email role } }"}'

# Get user by ID
curl -X POST http://localhost:8080/graphql \
  -H 'Content-Type: application/json' \
  -d '{"query": "{ user(id: \"1\") { id name email role } }"}'

# List all posts
curl -X POST http://localhost:8080/graphql \
  -H 'Content-Type: application/json' \
  -d '{"query": "{ posts { id title content author_id published } }"}'

# Create user (mutation)
curl -X POST http://localhost:8080/graphql \
  -H 'Content-Type: application/json' \
  -d '{"query": "mutation { createUser(name: \"Alice\", email: \"alice@example.com\", role: \"admin\") { id name email } }"}'

# Create post (mutation)
curl -X POST http://localhost:8080/graphql \
  -H 'Content-Type: application/json' \
  -d '{"query": "mutation { createPost(title: \"My Post\", content: \"Hello World\", author_id: \"1\") { id title } }"}'
```

## GraphQL Schema

### Types

**User**
```graphql
type User {
  id: String!
  name: String!
  email: String!
  role: String!
  created_at: String!
}
```

**Post**
```graphql
type Post {
  id: String!
  title: String!
  content: String!
  author_id: String!
  published: Boolean!
  created_at: String!
}
```

### Queries

```graphql
type Query {
  # Get all users
  users: [User!]!
  
  # Get user by ID
  user(id: String!): User
  
  # Get all posts
  posts: [Post!]!
  
  # Get post by ID
  post(id: String!): Post
}
```

### Mutations

```graphql
type Mutation {
  # Create a new user
  createUser(name: String!, email: String!, role: String): User!
  
  # Create a new post
  createPost(title: String!, content: String!, author_id: String!): Post!
  
  # Update post publication status
  updatePost(id: String!, published: Boolean!): Post!
  
  # Delete a post
  deletePost(id: String!): DeleteResult!
}

type DeleteResult {
  success: Boolean!
  id: String!
}
```

## Example Queries

### List All Users

```graphql
{
  users {
    id
    name
    email
    role
    created_at
  }
}
```

**Response**:
```json
{
  "data": {
    "users": [
      {
        "id": "1",
        "name": "John Doe",
        "email": "john@example.com",
        "role": "admin",
        "created_at": "2025-01-15T10:00:00Z"
      },
      {
        "id": "2",
        "name": "Jane Smith",
        "email": "jane@example.com",
        "role": "user",
        "created_at": "2025-01-15T10:00:00Z"
      }
    ]
  }
}
```

### Get User by ID

```graphql
{
  user(id: "1") {
    id
    name
    email
    role
  }
}
```

### List All Posts

```graphql
{
  posts {
    id
    title
    content
    author_id
    published
    created_at
  }
}
```

### Get Post by ID

```graphql
{
  post(id: "1") {
    id
    title
    content
    published
  }
}
```

## Example Mutations

### Create User

```graphql
mutation {
  createUser(
    name: "Alice Johnson"
    email: "alice@example.com"
    role: "admin"
  ) {
    id
    name
    email
    role
    created_at
  }
}
```

**Response**:
```json
{
  "data": {
    "createUser": {
      "id": "3",
      "name": "Alice Johnson",
      "email": "alice@example.com",
      "role": "admin",
      "created_at": "2025-01-15T10:05:00Z"
    }
  }
}
```

### Create Post

```graphql
mutation {
  createPost(
    title: "Getting Started with GraphQL"
    content: "GraphQL is a powerful query language..."
    author_id: "1"
  ) {
    id
    title
    content
    author_id
    published
  }
}
```

### Publish Post

```graphql
mutation {
  updatePost(id: "1", published: true) {
    id
    title
    published
  }
}
```

### Delete Post

```graphql
mutation {
  deletePost(id: "1") {
    success
    id
  }
}
```

## Code Walkthrough

### Schema Implementation

The example implements a custom GraphQL schema:

```go
type BlogSchema struct{}

func (s *BlogSchema) Execute(query string, variables map[string]interface{}) (interface{}, error) {
    // Parse and execute GraphQL query
    // Route to appropriate resolver
}
```

**Note**: This is a simplified implementation for demonstration. In production, use a proper GraphQL library like `graphql-go` or `gqlgen`.

### Query Resolvers

Query resolvers fetch data:

```go
func (s *BlogSchema) queryUsers() (interface{}, error) {
    userList := make([]*User, 0, len(users))
    for _, user := range users {
        userList = append(userList, user)
    }
    return map[string]interface{}{
        "users": userList,
    }, nil
}

func (s *BlogSchema) queryUser(id string) (interface{}, error) {
    user, exists := users[id]
    if !exists {
        return nil, fmt.Errorf("user not found: %s", id)
    }
    return map[string]interface{}{
        "user": user,
    }, nil
}
```

### Mutation Resolvers

Mutation resolvers modify data:

```go
func (s *BlogSchema) mutationCreateUser(name, email, role string) (interface{}, error) {
    if name == "" || email == "" {
        return nil, fmt.Errorf("name and email are required")
    }
    
    user := &User{
        ID:        generateID(),
        Name:      name,
        Email:     email,
        Role:      role,
        CreatedAt: time.Now(),
    }
    users[user.ID] = user
    
    return map[string]interface{}{
        "createUser": user,
    }, nil
}
```

### GraphQL Manager Configuration

Configure the GraphQL endpoint:

```go
graphqlManager := pkg.NewGraphQLManager(router, db, authManager)

config := pkg.GraphQLConfig{
    EnableIntrospection: true,
    EnablePlayground:    true,
    MaxRequestSize:      1024 * 1024, // 1MB
    Timeout:             30 * time.Second,
    RateLimit: &pkg.GraphQLRateLimitConfig{
        Limit:  100,
        Window: time.Minute,
        Key:    "ip_address",
    },
}

err = graphqlManager.RegisterSchema("/graphql", schema, config)
```

**Configuration Options**:
- **EnableIntrospection**: Allow schema introspection (disable in production if needed)
- **EnablePlayground**: Enable GraphQL Playground UI
- **MaxRequestSize**: Limit request payload size
- **Timeout**: Query execution timeout
- **RateLimit**: Protect against abuse

### Schema Introspection

The example supports introspection for tooling:

```go
func (s *BlogSchema) introspection() (interface{}, error) {
    return map[string]interface{}{
        "__schema": map[string]interface{}{
            "types": []map[string]interface{}{
                {
                    "name": "User",
                    "fields": []map[string]string{
                        {"name": "id", "type": "String!"},
                        {"name": "name", "type": "String!"},
                        // ... more fields
                    },
                },
            },
        },
    }, nil
}
```

## GraphQL Best Practices

### Field Selection

Only request fields you need:

```graphql
# Good - only request needed fields
{
  users {
    id
    name
  }
}

# Avoid - requesting all fields when not needed
{
  users {
    id
    name
    email
    role
    created_at
  }
}
```

### Error Handling

GraphQL returns errors in a standard format:

```json
{
  "errors": [
    {
      "message": "user not found: 999",
      "path": ["user"]
    }
  ],
  "data": {
    "user": null
  }
}
```

### Mutations

Use mutations for data modification:

```graphql
# Good - use mutation for creating data
mutation {
  createUser(name: "Alice", email: "alice@example.com") {
    id
  }
}

# Bad - don't use queries for modifications
{
  createUser(name: "Alice", email: "alice@example.com") {
    id
  }
}
```

### Variables

Use variables for dynamic values:

```graphql
mutation CreateUser($name: String!, $email: String!) {
  createUser(name: $name, email: $email) {
    id
    name
  }
}
```

**Variables**:
```json
{
  "name": "Alice",
  "email": "alice@example.com"
}
```

## Production Considerations

### Use a GraphQL Library

For production, use a proper GraphQL library:

```bash
# graphql-go
go get github.com/graphql-go/graphql

# gqlgen (code generation)
go get github.com/99designs/gqlgen
```

### Define Schema with SDL

Use Schema Definition Language (SDL):

```graphql
type User {
  id: ID!
  name: String!
  email: String!
  posts: [Post!]!
}

type Post {
  id: ID!
  title: String!
  author: User!
}

type Query {
  user(id: ID!): User
  users: [User!]!
}

type Mutation {
  createUser(input: CreateUserInput!): User!
}

input CreateUserInput {
  name: String!
  email: String!
}
```

### Add Authentication

Protect sensitive operations:

```go
config := pkg.GraphQLConfig{
    RequireAuth:   true,
    RequiredRoles: []string{"admin"},
    // ... other config
}
```

### Implement DataLoader

Prevent N+1 queries with DataLoader pattern:

```go
// Batch load users
func (l *UserLoader) Load(ids []string) ([]*User, error) {
    // Single database query for all IDs
    return fetchUsersByIDs(ids)
}
```

### Add Caching

Cache frequently accessed data:

```go
// Cache user data
cacheKey := fmt.Sprintf("user:%s", id)
if cached, err := ctx.Cache().Get(cacheKey); err == nil {
    return cached.(*User), nil
}

user := fetchUser(id)
ctx.Cache().Set(cacheKey, user, 5*time.Minute)
```

### Limit Query Depth

Prevent deeply nested queries:

```go
config := pkg.GraphQLConfig{
    MaxQueryDepth: 10,
    // ... other config
}
```

### Implement Pagination

Use cursor-based pagination:

```graphql
type Query {
  users(first: Int, after: String): UserConnection!
}

type UserConnection {
  edges: [UserEdge!]!
  pageInfo: PageInfo!
}

type UserEdge {
  node: User!
  cursor: String!
}

type PageInfo {
  hasNextPage: Boolean!
  endCursor: String
}
```

## Common Issues

### "Query too complex"

**Solution**: Reduce query depth or implement query complexity analysis

### "Rate limit exceeded"

**Solution**: Wait for rate limit window to reset, or implement backoff

### "Introspection disabled"

**Solution**: Enable introspection in config (only for development)

## Next Steps

After understanding this example:

1. **Use a GraphQL library**: Implement with `graphql-go` or `gqlgen`
2. **Add authentication**: Protect sensitive queries and mutations
3. **Implement DataLoader**: Optimize database queries
4. **Add subscriptions**: Real-time updates with WebSocket
5. **Study production patterns**: Review [Full Featured App](full-featured-app.md)

## Related Documentation

- [API Styles Guide](../guides/api-styles.md) - REST, GraphQL, gRPC, SOAP
- [Router API](../api/router.md) - Routing reference
- [Security Guide](../guides/security.md) - Authentication and authorization
- [WebSocket Guide](../guides/websockets.md) - Real-time subscriptions

## Source Code

The complete source code for this example is available at `examples/graphql_example.go` in the repository.
