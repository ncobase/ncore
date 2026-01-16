# Example 07: Authentication & Authorization

Demonstrates comprehensive authentication and authorization using NCore's security module, JWT tokens, middleware
patterns, and role-based access control.

## Features

- **JWT Authentication**: Token-based authentication with NCore's `security/jwt`
- **User Registration & Login**: Complete auth flow
- **Password Hashing**: Secure password storage with bcrypt
- **Middleware**: Authentication and authorization middleware
- **Role-Based Access Control (RBAC)**: User roles and permissions
- **Token Refresh**: Access and refresh token pattern
- **Session Management**: Track user sessions

## Architecture

```text
┌──────────────┐                    ┌──────────────┐
│    Client    │───── Credentials ──►│    Auth      │
└──────────────┘                    │   Service    │
        │                           └──────┬───────┘
        │ JWT Token                        │
        ▼                                  ▼
┌──────────────┐                    ┌──────────────┐
│  Protected   │◄──── Verify ──────│     JWT      │
│   Resource   │                    │TokenManager  │
└──────────────┘                    └──────────────┘
```

## Features Demonstrated

### 1. User Registration

```go
type RegisterRequest struct {
    Name     string `json:"name" binding:"required"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=8"`
}

func (s *AuthService) Register(ctx context.Context, req *RegisterRequest) (*User, error) {
    // Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword(
        []byte(req.Password),
        bcrypt.DefaultCost,
    )

    // Create user
    user := &User{
        Name:     req.Name,
        Email:    req.Email,
        Password: string(hashedPassword),
        Role:     "user",
    }

    return s.repo.Create(ctx, user)
}
```

### 2. Login & Token Generation

```go
func (s *AuthService) Login(ctx context.Context, email, password string) (*TokenPair, error) {
    user, err := s.repo.FindByEmail(ctx, email)
    if err != nil {
        return nil, ErrInvalidCredentials
    }

    // Verify password
    if err := bcrypt.CompareHashAndPassword(
        []byte(user.Password),
        []byte(password),
    ); err != nil {
        return nil, ErrInvalidCredentials
    }

    // Generate tokens
    accessToken, err := s.tokenManager.Generate(user.ID, 15*time.Minute)
    refreshToken, err := s.tokenManager.Generate(user.ID, 7*24*time.Hour)

    return &TokenPair{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
    }, nil
}
```

### 3. Authentication Middleware

```go
func AuthMiddleware(tokenManager *jwt.TokenManager) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Extract token from Authorization header
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
            c.Abort()
            return
        }

        token := strings.TrimPrefix(authHeader, "Bearer ")

        // Validate token
        claims, err := tokenManager.Validate(token)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
            c.Abort()
            return
        }

        // Set user ID in context
        c.Set("user_id", claims.UserID)
        c.Next()
    }
}
```

### 4. Authorization Middleware (RBAC)

```go
func RequireRole(roles ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetString("user_id")
        user, err := getUserByID(userID)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
            c.Abort()
            return
        }

        // Check if user has required role
        hasRole := false
        for _, role := range roles {
            if user.Role == role {
                hasRole = true
                break
            }
        }

        if !hasRole {
            c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
            c.Abort()
            return
        }

        c.Next()
    }
}
```

## Project Structure

```text
07-authentication/
├── data/
│   └── data.go          # SQLite connection
├── service/
│   ├── service.go       # Auth business logic
│   ├── structs/         # Auth models
│   └── data/
│       └── repository/  # SQLite repositories
├── handler/
│   └── auth.go          # HTTP handlers
└── middleware/
│       └── auth.go          # Auth/authz middleware
├── main.go
├── config.yaml
└── README.md
```

## API Endpoints

### Register

```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "securepass123"
  }'
```

### Login

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "securepass123"
  }'

# Response:
{
  "access_token": "eyJhbGc...",
  "refresh_token": "eyJhbGc...",
  "expires_in": 900
}
```

### Access Protected Resource

```bash
curl http://localhost:8080/api/profile \
  -H "Authorization: Bearer eyJhbGc..."
```

### Admin-Only Endpoint

```bash
curl http://localhost:8080/api/admin/users \
  -H "Authorization: Bearer eyJhbGc..."
```

### Refresh Token

```bash
curl -X POST http://localhost:8080/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "eyJhbGc..."
  }'
```

## Security Features

### 1. Password Security

- **Hashing**: bcrypt with default cost
- **Validation**: Minimum 8 characters, complexity requirements
- **No Plain Storage**: Never store plain passwords

### 2. Token Security

- **JWT Signing**: HMAC-SHA256 algorithm
- **Short Expiry**: Access tokens expire in 15 minutes
- **Refresh Tokens**: Long-lived (7 days) for token renewal
- **Secret Management**: Load secret from environment/config

### 3. Middleware Chain

```go
// Public routes
public := r.Group("/auth")
public.POST("/register", authHandler.Register)
public.POST("/login", authHandler.Login)

// Protected routes
api := r.Group("/api")
api.Use(AuthMiddleware(tokenManager))
api.GET("/profile", userHandler.GetProfile)

// Admin-only routes
admin := r.Group("/api/admin")
admin.Use(AuthMiddleware(tokenManager))
admin.Use(RequireRole("admin"))
admin.GET("/users", adminHandler.ListUsers)
```

## Role Hierarchy

```text
admin
  ├── Can manage all users
  ├── Can access all resources
  └── Can modify system settings

moderator
  ├── Can moderate content
  └── Can access user data

user
  └── Can access own resources
```

## Configuration

```yaml
data:
  database:
    master:
      driver: sqlite3
      source: "file:auth.db?cache=shared&_fk=1"

auth:
  jwt:
    secret: ${JWT_SECRET} # Load from environment
    access_token_ttl: 900 # 15 minutes
    refresh_token_ttl: 604800 # 7 days
  password:
    min_length: 8
    require_uppercase: true
    require_number: true
    require_special: true
  session:
    max_concurrent: 5 # Max sessions per user
```

## Token Payload

```json
{
  "user_id": "123",
  "email": "user@example.com",
  "role": "user",
  "exp": 1704067200,
  "iat": 1704066300
}
```

## Testing

### Unit Tests

```go
func TestAuthService_Login(t *testing.T) {
    svc := NewAuthService(mockRepo, mockTokenManager)

    tokens, err := svc.Login(ctx, "user@example.com", "password123")
    assert.NoError(t, err)
    assert.NotEmpty(t, tokens.AccessToken)
    assert.NotEmpty(t, tokens.RefreshToken)
}
```

### Integration Tests

```go
func TestAuth_EndToEnd(t *testing.T) {
    // Register
    user := register(t, "test@example.com", "password123")

    // Login
    tokens := login(t, "test@example.com", "password123")

    // Access protected resource
    profile := getProfile(t, tokens.AccessToken)
    assert.Equal(t, user.Email, profile.Email)
}
```

## Security Best Practices

1. **HTTPS Only**: Always use HTTPS in production
2. **Rate Limiting**: Limit login attempts
3. **Token Rotation**: Rotate refresh tokens periodically
4. **Secure Cookies**: Use httpOnly, secure flags
5. **CORS**: Configure properly for web clients
6. **Input Validation**: Validate all user inputs
7. **SQL Injection**: Use parameterized queries
8. **XSS Protection**: Sanitize outputs

## Common Patterns

### 1. Current User Helper

```go
func GetCurrentUser(c *gin.Context) (*User, error) {
    userID := c.GetString("user_id")
    return getUserByID(userID)
}
```

### 2. Permission Checks

```go
func CanModifyResource(user *User, resource *Resource) bool {
    return user.Role == "admin" || resource.OwnerID == user.ID
}
```

### 3. Audit Logging

```go
func (s *AuthService) Login(ctx context.Context, email, password string) (*TokenPair, error) {
    // ... login logic ...

    s.logger.Info("user logged in", "user_id", user.ID, "ip", getIP(ctx))
    return tokens, nil
}
```

## Use Cases

- User registration and login
- API authentication
- Admin dashboards
- Multi-tenant applications
- Microservice authentication
- Mobile app backends

## Next Steps

- Integrate with [multi-module app](../03-multi-module)
- Add OAuth 2.0 providers (Google, GitHub)
- Implement 2FA (two-factor authentication)
- Add session management
- Integrate with [full application](../08-full-application)

## License

This example is part of the NCore project.
