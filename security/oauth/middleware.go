package oauth

import (
	"context"
	"net/http"
	"strings"
)

// MiddlewareOptions represents middleware configuration options
type MiddlewareOptions struct {
	TokenLookup    string                                             // "header:Authorization,query:token,cookie:oauth_token"
	AuthScheme     string                                             // "Bearer" or "Token"
	SkipPaths      []string                                           // Paths to skip OAuth validation
	ErrorHandler   func(http.ResponseWriter, *http.Request)           // Custom error handler
	SuccessHandler func(http.ResponseWriter, *http.Request, *Profile) // Success handler
}

// Middleware creates OAuth validation middleware
func Middleware(client ClientInterface, opts *MiddlewareOptions) func(http.Handler) http.Handler {
	if opts == nil {
		opts = &MiddlewareOptions{
			TokenLookup: "header:Authorization",
			AuthScheme:  "Bearer",
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if path should be skipped
			for _, path := range opts.SkipPaths {
				if strings.HasPrefix(r.URL.Path, path) {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Extract token
			token := extractToken(r, opts.TokenLookup, opts.AuthScheme)
			if token == "" {
				if opts.ErrorHandler != nil {
					opts.ErrorHandler(w, r)
				} else {
					http.Error(w, "Missing or invalid token", http.StatusUnauthorized)
				}
				return
			}

			// Here you would validate the token and get user profile
			// This is a simplified example
			ctx := context.WithValue(r.Context(), "oauth_token", token)

			if opts.SuccessHandler != nil {
				// In real implementation, you'd validate token and get profile
				profile := &Profile{} // This should come from token validation
				opts.SuccessHandler(w, r.WithContext(ctx), profile)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractToken extracts token from request based on lookup configuration
func extractToken(r *http.Request, tokenLookup, authScheme string) string {
	lookups := strings.Split(tokenLookup, ",")

	for _, lookup := range lookups {
		parts := strings.Split(strings.TrimSpace(lookup), ":")
		if len(parts) != 2 {
			continue
		}

		source, key := parts[0], parts[1]

		switch source {
		case "header":
			token := r.Header.Get(key)
			if token != "" {
				// Remove auth scheme prefix
				if authScheme != "" && strings.HasPrefix(token, authScheme+" ") {
					return token[len(authScheme)+1:]
				}
				return token
			}
		case "query":
			token := r.URL.Query().Get(key)
			if token != "" {
				return token
			}
		case "cookie":
			if cookie, err := r.Cookie(key); err == nil && cookie.Value != "" {
				return cookie.Value
			}
		}
	}

	return ""
}

// GetTokenFromContext extracts OAuth token from context
func GetTokenFromContext(ctx context.Context) string {
	if token, ok := ctx.Value("oauth_token").(string); ok {
		return token
	}
	return ""
}

// GetProfileFromContext extracts user profile from context
func GetProfileFromContext(ctx context.Context) *Profile {
	if profile, ok := ctx.Value("oauth_profile").(*Profile); ok {
		return profile
	}
	return nil
}
