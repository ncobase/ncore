package oauth

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const (
	tokenContextKey   contextKey = "oauth_token"
	profileContextKey contextKey = "oauth_profile"
)

// MiddlewareOptions represents middleware configuration options
type MiddlewareOptions struct {
	TokenLookup      string                                                    // "header:Authorization,query:token,cookie:oauth_token"
	AuthScheme       string                                                    // "Bearer" or "Token"
	Provider         Provider                                                  // Static provider used for token validation
	ProviderResolver func(*http.Request) (Provider, error)                     // Optional request-aware provider resolver
	ProfileLoader    func(context.Context, Provider, string) (*Profile, error) // Optional profile loader
	SkipPaths        []string                                                  // Paths to skip OAuth validation
	ErrorHandler     func(http.ResponseWriter, *http.Request)                  // Custom error handler
	SuccessHandler   func(http.ResponseWriter, *http.Request, *Profile)        // Success handler
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

			token := extractToken(r, opts.TokenLookup, opts.AuthScheme)
			if token == "" {
				writeMiddlewareError(w, r, opts)
				return
			}

			provider, err := resolveProvider(r, opts)
			if err != nil {
				writeMiddlewareError(w, r, opts)
				return
			}

			tokenInfo, err := client.ValidateToken(r.Context(), provider, token)
			if err != nil || tokenInfo == nil || !tokenInfo.Valid {
				writeMiddlewareError(w, r, opts)
				return
			}

			profile, err := loadProfile(r.Context(), client, opts, provider, token)
			if err != nil || profile == nil {
				writeMiddlewareError(w, r, opts)
				return
			}

			ctx := context.WithValue(r.Context(), tokenContextKey, token)
			ctx = context.WithValue(ctx, profileContextKey, profile)

			if opts.SuccessHandler != nil {
				opts.SuccessHandler(w, r.WithContext(ctx), profile)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func resolveProvider(r *http.Request, opts *MiddlewareOptions) (Provider, error) {
	if opts.ProviderResolver != nil {
		provider, err := opts.ProviderResolver(r)
		if err != nil {
			return "", err
		}
		if provider == "" {
			return "", ErrProviderNotSupported
		}
		return provider, nil
	}
	if opts.Provider == "" {
		return "", ErrProviderNotSupported
	}
	return opts.Provider, nil
}

func loadProfile(ctx context.Context, client ClientInterface, opts *MiddlewareOptions, provider Provider, token string) (*Profile, error) {
	if opts.ProfileLoader != nil {
		return opts.ProfileLoader(ctx, provider, token)
	}
	return client.GetUserProfile(ctx, provider, token)
}

func writeMiddlewareError(w http.ResponseWriter, r *http.Request, opts *MiddlewareOptions) {
	if opts.ErrorHandler != nil {
		opts.ErrorHandler(w, r)
		return
	}
	http.Error(w, "Missing or invalid token", http.StatusUnauthorized)
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
				if authScheme == "" {
					return token
				}
				parts := strings.SplitN(token, " ", 2)
				if len(parts) == 2 && strings.EqualFold(parts[0], authScheme) {
					return strings.TrimSpace(parts[1])
				}
				return ""
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
	if token, ok := ctx.Value(tokenContextKey).(string); ok {
		return token
	}
	return ""
}

// GetProfileFromContext extracts user profile from context
func GetProfileFromContext(ctx context.Context) *Profile {
	if profile, ok := ctx.Value(profileContextKey).(*Profile); ok {
		return profile
	}
	return nil
}
