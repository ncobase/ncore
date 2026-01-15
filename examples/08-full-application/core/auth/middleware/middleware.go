// Package middleware provides JWT middleware for the full app.
package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/resp"
	securityjwt "github.com/ncobase/ncore/security/jwt"
)

type Middleware struct {
	tokenManager *securityjwt.TokenManager
	logger       *logger.Logger
}

func NewMiddleware(tokenManager *securityjwt.TokenManager, log *logger.Logger) *Middleware {
	return &Middleware{
		tokenManager: tokenManager,
		logger:       log,
	}
}

// AuthMiddleware validates JWT tokens and adds user info to context.
func (m *Middleware) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			resp.Fail(c.Writer, resp.UnAuthorized("missing authorization header"))
			c.Abort()
			return
		}

		// Extract token (format: "Bearer <token>")
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			resp.Fail(c.Writer, resp.UnAuthorized("invalid authorization format"))
			c.Abort()
			return
		}

		token := parts[1]

		// Validate token
		tokenObj, err := m.tokenManager.ValidateToken(token)
		if err != nil {
			m.logger.Error(c.Request.Context(), "Token validation failed", "error", err)
			resp.Fail(c.Writer, resp.UnAuthorized("invalid or expired token"))
			c.Abort()
			return
		}

		claims, ok := tokenObj.Claims.(jwtv5.MapClaims)
		if !ok {
			resp.Fail(c.Writer, resp.UnAuthorized("invalid token payload"))
			c.Abort()
			return
		}

		if !securityjwt.IsAccessToken(claims) {
			resp.Fail(c.Writer, resp.UnAuthorized("invalid token type"))
			c.Abort()
			return
		}

		userID := securityjwt.GetPayloadString(claims, "user_id")
		if userID == "" {
			userID = securityjwt.GetTokenID(claims)
		}
		if userID == "" {
			resp.Fail(c.Writer, resp.UnAuthorized("invalid token payload"))
			c.Abort()
			return
		}

		email := securityjwt.GetPayloadString(claims, "email")
		role := securityjwt.GetPayloadString(claims, "role")
		name := securityjwt.GetPayloadString(claims, "name")

		// Add user info to context
		c.Set("user_id", userID)
		c.Set("email", email)
		c.Set("role", role)
		c.Set("name", name)

		m.logger.Debug(c.Request.Context(), "User authenticated", "user_id", userID, "email", email)
		c.Next()
	}
}

// RequireRole ensures user has required role.
func (m *Middleware) RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user role from context (set by AuthMiddleware)
		userRole, exists := c.Get("role")
		if !exists {
			resp.Fail(c.Writer, resp.UnAuthorized("not authenticated"))
			c.Abort()
			return
		}

		roleStr, ok := userRole.(string)
		if !ok {
			resp.Fail(c.Writer, resp.InternalServer("invalid user role"))
			c.Abort()
			return
		}

		// Check if user has required role
		hasRole := false
		for _, requiredRole := range roles {
			if roleStr == requiredRole {
				hasRole = true
				break
			}
		}

		if !hasRole {
			m.logger.Warn(c.Request.Context(), "Access denied", "user_role", roleStr, "required_roles", roles)
			resp.Fail(c.Writer, resp.Forbidden("insufficient permissions"))
			c.Abort()
			return
		}

		m.logger.Debug(c.Request.Context(), "Role check passed", "user_role", roleStr, "required_roles", roles)
		c.Next()
	}
}

// RequireAdmin ensures user is an admin.
func (m *Middleware) RequireAdmin() gin.HandlerFunc {
	return m.RequireRole("admin")
}

// OptionalAuth attempts to authenticate but doesn't require it.
func (m *Middleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No auth header, continue without user context
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		token := parts[1]
		tokenObj, err := m.tokenManager.ValidateToken(token)
		if err != nil {
			c.Next()
			return
		}

		claims, ok := tokenObj.Claims.(jwtv5.MapClaims)
		if !ok {
			c.Next()
			return
		}

		if !securityjwt.IsAccessToken(claims) {
			c.Next()
			return
		}

		userID := securityjwt.GetPayloadString(claims, "user_id")
		if userID == "" {
			userID = securityjwt.GetTokenID(claims)
		}
		if userID != "" {
			email := securityjwt.GetPayloadString(claims, "email")
			role := securityjwt.GetPayloadString(claims, "role")
			name := securityjwt.GetPayloadString(claims, "name")
			c.Set("user_id", userID)
			c.Set("email", email)
			c.Set("role", role)
			c.Set("name", name)
		}

		c.Next()
	}
}

// GetCurrentUserID gets the current user ID from context.
func GetCurrentUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", false
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return "", false
	}

	return userIDStr, true
}

// GetCurrentRole gets the current user role from context.
func GetCurrentRole(c *gin.Context) (string, bool) {
	role, exists := c.Get("role")
	if !exists {
		return "", false
	}

	roleStr, ok := role.(string)
	if !ok {
		return "", false
	}

	return roleStr, true
}

// IsAdmin checks if the current user is an admin.
func IsAdmin(c *gin.Context) bool {
	role, exists := GetCurrentRole(c)
	if !exists {
		return false
	}
	return role == "admin"
}
