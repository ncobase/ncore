// Package middleware provides auth and RBAC middleware.
package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/examples/07-authentication/service"
	auth "github.com/ncobase/ncore/examples/07-authentication/service"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/resp"
	securityjwt "github.com/ncobase/ncore/security/jwt"
)

// AuthMiddleware creates authentication middleware.
func AuthMiddleware(authService *service.Service, logger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			resp.Fail(c.Writer, resp.UnAuthorized("missing authorization header"))
			c.Abort()
			return
		}

		// Check Bearer prefix
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			resp.Fail(c.Writer, resp.UnAuthorized("invalid authorization header format"))
			c.Abort()
			return
		}

		token := parts[1]

		// Validate token
		claims, err := authService.ValidateToken(token)
		if err != nil {
			logger.Warn(c.Request.Context(), "Invalid token", "error", err)
			resp.Fail(c.Writer, resp.UnAuthorized("invalid token"))
			c.Abort()
			return
		}

		userID := securityjwt.GetPayloadString(claims, "user_id")
		role := securityjwt.GetPayloadString(claims, "role")
		if userID == "" {
			resp.Fail(c.Writer, resp.UnAuthorized("invalid token"))
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Set("user_role", role)

		c.Next()
	}
}

// RequireRole creates role-based authorization middleware.
func RequireRole(authService *auth.Service, roles ...auth.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			resp.Fail(c.Writer, resp.UnAuthorized("unauthorized"))
			c.Abort()
			return
		}

		// Check if user has required role
		hasRole := false
		for _, role := range roles {
			if userRole == string(role) {
				hasRole = true
				break
			}
		}

		if !hasRole {
			resp.Fail(c.Writer, resp.Forbidden("insufficient permissions"))
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetCurrentUserID retrieves the current user ID from context.
func GetCurrentUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", false
	}
	return userID.(string), true
}

// GetCurrentUserRole retrieves the current user role from context.
func GetCurrentUserRole(c *gin.Context) (string, bool) {
	role, exists := c.Get("user_role")
	if !exists {
		return "", false
	}
	return role.(string), true
}
