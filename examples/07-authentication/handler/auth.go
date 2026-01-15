// Package handler exposes authentication HTTP endpoints.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/examples/07-authentication/middleware"
	auth "github.com/ncobase/ncore/examples/07-authentication/service"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/resp"
)

// AuthHandler handles authentication HTTP requests.
type AuthHandler struct {
	authService *auth.Service
	logger      *logger.Logger
}

// NewAuthHandler creates a new auth handler.
func NewAuthHandler(authService *auth.Service, logger *logger.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

// Register handles user registration.
func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Name     string    `json:"name" binding:"required"`
		Email    string    `json:"email" binding:"required,email"`
		Password string    `json:"password" binding:"required,min=8"`
		Role     auth.Role `json:"role"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	// Default role is user
	if req.Role == "" {
		req.Role = auth.RoleUser
	}

	user, err := h.authService.Register(c.Request.Context(), req.Name, req.Email, req.Password, req.Role)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.WithStatusCode(c.Writer, http.StatusCreated, user)
}

// Login handles user login.
func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	tokens, err := h.authService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		resp.Fail(c.Writer, resp.UnAuthorized("invalid credentials"))
		return
	}

	resp.Success(c.Writer, tokens)
}

// RefreshToken handles token refresh.
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	tokens, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		resp.Fail(c.Writer, resp.UnAuthorized("invalid refresh token"))
		return
	}

	resp.Success(c.Writer, tokens)
}

// Logout handles user logout.
func (h *AuthHandler) Logout(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	if err := h.authService.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, map[string]string{"message": "logged out successfully"})
}

// GetProfile retrieves the current user's profile.
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		resp.Fail(c.Writer, resp.UnAuthorized("unauthorized"))
		return
	}

	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		resp.Fail(c.Writer, resp.NotFound("user not found"))
		return
	}

	resp.Success(c.Writer, user)
}

// AdminHandler handles admin-only operations.
type AdminHandler struct {
	authService *auth.Service
	logger      *logger.Logger
}

// NewAdminHandler creates a new admin handler.
func NewAdminHandler(authService *auth.Service, logger *logger.Logger) *AdminHandler {
	return &AdminHandler{
		authService: authService,
		logger:      logger,
	}
}

// ListUsers lists all users (admin only).
func (h *AdminHandler) ListUsers(c *gin.Context) {
	users, err := h.authService.ListUsers(c.Request.Context())
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer("failed to list users"))
		return
	}

	resp.Success(c.Writer, users)
}

// DeleteUser deletes a user (admin only).
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("user_id")
	if err := h.authService.DeleteUser(c.Request.Context(), userID); err != nil {
		resp.Fail(c.Writer, resp.InternalServer("failed to delete user"))
		return
	}

	resp.Success(c.Writer, map[string]any{
		"message": "User deleted",
		"user_id": userID,
	})
}
