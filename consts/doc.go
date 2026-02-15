// Package consts defines application-wide constants including context keys,
// character sets, metadata keys, and social platform identifiers.
//
// This package provides:
//   - Context key constants for storing/retrieving values
//   - Character sets and encoding definitions
//   - Metadata field name constants
//   - Social platform identifiers
//   - Common application constants
//
// # Context Keys
//
// Standard keys for storing values in context.Context:
//
//	const (
//	    GinContextKey = "GinContext"     // Gin context
//	    UserKey       = "UserID"         // User ID
//	    UsernameKey   = "Username"       // Username
//	    TraceIDKey    = "TraceID"        // Trace ID for logging
//	    SpanIDKey     = "SpanID"         // Span ID for tracing
//	)
//
// Usage with ctxutil:
//
//	import "github.com/ncobase/ncore/consts"
//	import "github.com/ncobase/ncore/ctxutil"
//
//	ctx = ctxutil.SetValue(ctx, consts.UserKey, "user-123")
//	userID := ctxutil.GetValue(ctx, consts.UserKey)
//
// # Character Sets
//
// Predefined character sets for various purposes:
//
//	consts.Numeric            // "0123456789"
//	consts.Alphabetic         // "a-zA-Z"
//	consts.Alphanumeric       // "a-zA-Z0-9"
//	consts.AlphanumericSymbol // "a-zA-Z0-9!@#$%"
//
// Usage for validation or generation:
//
//	import "strings"
//
//	func isNumeric(s string) bool {
//	    return strings.ContainsOnly(s, consts.Numeric)
//	}
//
// # Metadata Keys
//
// Standard metadata field names:
//
//	consts.CreatedAt   // "created_at"
//	consts.UpdatedAt   // "updated_at"
//	consts.DeletedAt   // "deleted_at"
//	consts.CreatedBy   // "created_by"
//	consts.UpdatedBy   // "updated_by"
//
// Use for consistent field naming:
//
//	type Model struct {
//	    CreatedAt time.Time `json:"created_at"`
//	    UpdatedAt time.Time `json:"updated_at"`
//	}
//
// # Social Platforms
//
// Identifiers for social authentication providers:
//
//	consts.PlatformWeChat   // "wechat"
//	consts.PlatformGoogle   // "google"
//	consts.PlatformGitHub   // "github"
//	consts.PlatformFacebook // "facebook"
//
// Usage in OAuth flows:
//
//	switch provider {
//	case consts.PlatformGoogle:
//	    return googleAuth.Authenticate(token)
//	case consts.PlatformGitHub:
//	    return githubAuth.Authenticate(token)
//	}
//
// # Best Practices
//
//   - Use package constants instead of string literals
//   - Reference consts for consistent naming
//   - Add new constants to appropriate categories
//   - Document custom constant additions
//   - Use typed constants where appropriate
//   - Avoid duplicating constant values
package consts
