package ctxutil

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/config"
	"github.com/ncobase/ncore/consts"
	"github.com/ncobase/ncore/utils/uuid"
)

const (
	ginContextKey   = consts.GinContextKey
	userIDKey       = consts.UserKey
	usernameKey     = consts.UsernameKey
	userEmailKey    = consts.UserEmailKey
	userStatusKey   = consts.UserStatusKey
	userIsCertified = consts.UserIsCertifiedKey
	spaceIDKey      = consts.SpaceKey
	tokenKey        = consts.TokenKey
	providerKey     = "provider"
	profileKey      = "profile"
	configKey       = "config"
	emailSender     = "email_sender"
	storageKey      = "storage"
	TraceIDKey      = "trace_id"
	userRolesKey    = "user_roles"
	userPermissions = "user_permissions"
	userIsAdminKey  = "user_is_admin"
	spaceIDsKey     = "user_space_ids"
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

// SetUsername sets username to context.Context.
func SetUsername(ctx context.Context, username string) context.Context {
	return SetValue(ctx, usernameKey, username)
}

// GetUsername gets username from context.Context.
func GetUsername(ctx context.Context) string {
	if username, ok := GetValue(ctx, usernameKey).(string); ok {
		return username
	}
	return ""
}

// SetUserEmail sets user email to context.Context.
func SetUserEmail(ctx context.Context, email string) context.Context {
	return SetValue(ctx, userEmailKey, email)
}

// GetUserEmail gets user email from context.Context.
func GetUserEmail(ctx context.Context) string {
	if email, ok := GetValue(ctx, userEmailKey).(string); ok {
		return email
	}
	return ""
}

// SetUserStatus sets user status to context.Context.
func SetUserStatus(ctx context.Context, status int) context.Context {
	return SetValue(ctx, userStatusKey, status)
}

// GetUserStatus gets user status from context.Context.
func GetUserStatus(ctx context.Context) int {
	if status, ok := GetValue(ctx, userStatusKey).(int); ok {
		return status
	}
	return 3 // unknown
}

// SetUserIsCertified sets user is certified to context.Context.
func SetUserIsCertified(ctx context.Context, isCertified bool) context.Context {
	return SetValue(ctx, userIsCertified, isCertified)
}

// GetUserIsCertified gets user is certified from context.Context.
func GetUserIsCertified(ctx context.Context) bool {
	if isCertified, ok := GetValue(ctx, userIsCertified).(bool); ok {
		return isCertified
	}
	return false
}

// SetSpaceID sets space id to context.Context.
func SetSpaceID(ctx context.Context, uid string) context.Context {
	return SetValue(ctx, spaceIDKey, uid)
}

// GetSpaceID gets space id from context.Context.
func GetSpaceID(ctx context.Context) string {
	if uid, ok := GetValue(ctx, spaceIDKey).(string); ok {
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

// ExtractContext extracts context from payload map safely
func ExtractContext(payload *map[string]any) context.Context {
	if payload == nil {
		return context.Background()
	}

	if ctxVal, exists := (*payload)["ctx"]; exists {
		if ctx, ok := ctxVal.(context.Context); ok {
			delete(*payload, "ctx")
			return ctx
		}
	}
	return context.Background()
}

// SetUserRoles sets user roles to context.Context.
func SetUserRoles(ctx context.Context, roles []string) context.Context {
	return SetValue(ctx, userRolesKey, roles)
}

// GetUserRoles gets user roles from context.Context.
func GetUserRoles(ctx context.Context) []string {
	if roles, ok := GetValue(ctx, userRolesKey).([]string); ok {
		return roles
	}
	return []string{}
}

// SetUserPermissions sets user permissions to context.Context.
func SetUserPermissions(ctx context.Context, permissions []string) context.Context {
	return SetValue(ctx, userPermissions, permissions)
}

// GetUserPermissions gets user permissions from context.Context.
func GetUserPermissions(ctx context.Context) []string {
	if perms, ok := GetValue(ctx, userPermissions).([]string); ok {
		return perms
	}
	return []string{}
}

// SetUserIsAdmin sets user admin status to context.Context.
func SetUserIsAdmin(ctx context.Context, isAdmin bool) context.Context {
	return SetValue(ctx, userIsAdminKey, isAdmin)
}

// GetUserIsAdmin gets user admin status from context.Context.
func GetUserIsAdmin(ctx context.Context) bool {
	if isAdmin, ok := GetValue(ctx, userIsAdminKey).(bool); ok {
		return isAdmin
	}
	return false
}

// SetUserSpaceIDs sets user space IDs to context.Context.
func SetUserSpaceIDs(ctx context.Context, spaceIDs []string) context.Context {
	return SetValue(ctx, spaceIDsKey, spaceIDs)
}

// GetUserSpaceIDs gets user space IDs from context.Context.
func GetUserSpaceIDs(ctx context.Context) []string {
	if spaceIDs, ok := GetValue(ctx, spaceIDsKey).([]string); ok {
		return spaceIDs
	}
	return []string{}
}
