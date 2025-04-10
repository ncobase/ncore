package helper

import (
	"context"
	"github.com/ncobase/ncore/pkg/config"
	"github.com/ncobase/ncore/pkg/consts"
	"github.com/ncobase/ncore/pkg/uuid"

	"github.com/gin-gonic/gin"
)

const (
	ginContextKey = consts.GinContextKey
	userIDKey     = consts.UserKey
	tenantIDKey   = consts.TenantKey
	tokenKey      = consts.TokenKey
	providerKey   = "provider"
	profileKey    = "profile"
	configKey     = "config"
	emailSender   = "email_sender"
	storageKey    = "storage"
	TraceIDKey    = "trace_id"
)

// FromGinContext extracts the context.Context from *gin.Context.
func FromGinContext(c *gin.Context) context.Context {
	return c.Request.Context()
}

// WithGinContext returns a context.Context that embeds the *gin.Context.
func WithGinContext(ctx context.Context, c *gin.Context) context.Context {
	return context.WithValue(ctx, ginContextKey, c)
}

// GetGinContext extracts *gin.Context from context.Context if it exists.
func GetGinContext(ctx context.Context) (*gin.Context, bool) {
	if c, ok := ctx.Value(ginContextKey).(*gin.Context); ok {
		return c, ok
	}
	return nil, false
}

// GetValue retrieves a value from the context.
func GetValue(ctx context.Context, key string) any {
	if c, ok := GetGinContext(ctx); ok {
		if val, exists := c.Get(string(key)); exists {
			return val
		}
	}
	return ctx.Value(key)
}

// SetValue sets a value to the context.
func SetValue(ctx context.Context, key string, val any) context.Context {
	if c, ok := GetGinContext(ctx); ok {
		c.Set(string(key), val)
	}
	return context.WithValue(ctx, key, val)
}

// SetConfig sets config to context.Context.
func SetConfig(ctx context.Context, conf *config.Config) context.Context {
	return SetValue(ctx, configKey, conf)
}

// GetConfig gets config from context.Context.
func GetConfig(ctx context.Context) *config.Config {
	if conf, ok := GetValue(ctx, configKey).(*config.Config); ok {
		return conf
	}
	// Context does not contain config, load it from config.
	conf, err := config.GetConfig()
	if err != nil {
		return nil
	}
	return conf
}

// SetUserID sets user id to context.Context.
func SetUserID(ctx context.Context, uid string) context.Context {
	return SetValue(ctx, userIDKey, uid)
}

// GetUserID gets user id from context.Context.
func GetUserID(ctx context.Context) string {
	if uid, ok := GetValue(ctx, userIDKey).(string); ok {
		return uid
	}
	return ""
}

// SetTenantID sets tenant id to context.Context.
func SetTenantID(ctx context.Context, uid string) context.Context {
	return SetValue(ctx, tenantIDKey, uid)
}

// GetTenantID gets tenant id from context.Context.
func GetTenantID(ctx context.Context) string {
	if uid, ok := GetValue(ctx, tenantIDKey).(string); ok {
		return uid
	}
	return ""
}

// SetToken sets token to context.Context.
func SetToken(ctx context.Context, token string) context.Context {
	return SetValue(ctx, tokenKey, token)
}

// GetToken gets token from context.Context.
func GetToken(ctx context.Context) string {
	if token, ok := GetValue(ctx, tokenKey).(string); ok {
		return token
	}
	return ""
}

// SetProvider sets provider to context.Context.
func SetProvider(ctx context.Context, provider string) context.Context {
	return SetValue(ctx, providerKey, provider)
}

// GetProvider gets provider from context.Context.
func GetProvider(ctx context.Context) string {
	if provider, ok := GetValue(ctx, providerKey).(string); ok {
		return provider
	}
	return ""
}

// SetProfile sets profile to context.Context.
func SetProfile(ctx context.Context, profile any) context.Context {
	return SetValue(ctx, profileKey, profile)
}

// GetProfile gets profile from context.Context.
func GetProfile(ctx context.Context) any {
	if profile, ok := GetValue(ctx, profileKey).(any); ok {
		return profile
	}
	return nil
}

// GetTraceID gets trace id from context.Context or gin.Context.
func GetTraceID(ctx context.Context) string {
	if traceID, ok := GetValue(ctx, TraceIDKey).(string); ok {
		return traceID
	}
	return ""
}

// SetTraceID sets trace id to context.Context and gin.Context if available.
func SetTraceID(ctx context.Context, traceID string) context.Context {
	return SetValue(ctx, TraceIDKey, traceID)
}

// EnsureTraceID ensures that a trace ID exists in the context.
func EnsureTraceID(ctx context.Context) (context.Context, string) {
	if traceID := GetTraceID(ctx); traceID != "" {
		return ctx, traceID
	}
	traceID := uuid.NewString()
	return SetTraceID(ctx, traceID), traceID
}
