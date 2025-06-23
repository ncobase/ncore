# JWT Package

A simple and efficient JWT token management library for Go applications.

## Basic Usage

```go
import "your-project/security/jwt"

// Create token manager
tm := jwt.NewTokenManager("your-secret-key")

// With custom configuration
config := &jwt.TokenConfig{
    AccessTokenExpiry:   2 * time.Hour,
    RefreshTokenExpiry:  7 * 24 * time.Hour,
    RegisterTokenExpiry: 30 * time.Minute,
}
tm := jwt.NewTokenManager("secret", config)
```

## Token Generation

```go
payload := map[string]any{"username": "john", "role": "admin"}

// Generate tokens
accessToken, err := tm.GenerateAccessToken("user-123", payload)
refreshToken, err := tm.GenerateRefreshToken("user-123", payload)
registerToken, err := tm.GenerateRegisterToken("user-123", payload, "register")

// With custom expiry
customConfig := &jwt.TokenConfig{Expiry: 1 * time.Hour}
token, err := tm.GenerateAccessToken("user-123", payload, customConfig)
```

## Token Validation & Decoding

```go
// Validate token
token, err := tm.ValidateToken(tokenString)

// Decode claims
claims, err := tm.DecodeToken(tokenString)

// Get payload only
payload, err := tm.GetPayload(tokenString)

// Check expiry
expired := tm.IsTokenExpired(tokenString)
expiryTime, err := tm.GetTokenExpiry(tokenString)
```

## Token Refresh

```go
// Refresh if needed (refresh when < 30 minutes remaining)
newToken, refreshed, err := tm.RefreshTokenIfNeeded(tokenString, 30*time.Minute)
if refreshed {
    // Use new token
}
```

## Configuration Methods

```go
tm.SetSecret("new-secret")
secret := tm.GetSecret()
tm.SetAccessTokenExpiry(3 * time.Hour)
tm.SetRefreshTokenExpiry(14 * 24 * time.Hour)
tm.SetRegisterTokenExpiry(1 * time.Hour)
```

## Claim Extraction Utilities

```go
// Standard claims
tokenID := jwt.GetTokenID(claims)
subject := jwt.GetSubject(claims)
issuer := jwt.GetIssuer(claims)
audience := jwt.GetAudience(claims)
expiry := jwt.GetExpiration(claims)
issuedAt := jwt.GetIssuedAt(claims)
notBefore := jwt.GetNotBefore(claims)

// Payload extraction
payload := jwt.GetPayload(claims)
username := jwt.GetPayloadString(claims, "username")
isAdmin := jwt.GetPayloadBool(claims, "admin")
level := jwt.GetPayloadInt(claims, "level")
roles := jwt.GetPayloadStringSlice(claims, "roles")
hasKey := jwt.HasPayloadValue(claims, "key")

// Safe type extraction
str := jwt.GetString(data, "key")
num := jwt.GetInt(data, "key")
flag := jwt.GetBool(data, "key")
slice := jwt.GetStringSlice(data, "key")
nested := jwt.GetMap(data, "key")
```

## Token Type Validation

```go
isAccess := jwt.IsAccessToken(claims)
isRefresh := jwt.IsRefreshToken(claims)

// Validate specific type
err := jwt.ValidateTokenType(claims, "access")
```

## Token Timing Validation

```go
// Check all timing constraints
err := jwt.ValidateTokenTiming(claims)

// Individual checks
expired := jwt.IsTokenExpired(claims)
active := jwt.IsTokenActive(claims)
stale := jwt.IsTokenStale(claims, 24*time.Hour)
```

## Utility Functions

```go
// Check slice membership
contains := jwt.ContainsValue(slice, "value")
containsAny := jwt.ContainsAnyValue(slice, "val1", "val2")
```
