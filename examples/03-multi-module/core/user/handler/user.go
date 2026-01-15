// Package handler provides user HTTP handlers for the multi-module example.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/examples/03-multi-module/core/user/service"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/resp"
)

// UserHandler handles user HTTP requests.
type UserHandler struct {
	service *service.UserService
	logger  *logger.Logger
}

// NewUserHandler creates a new user handler.
func NewUserHandler(svc *service.UserService, logger *logger.Logger) *UserHandler {
	return &UserHandler{
		service: svc,
		logger:  logger,
	}
}

// Create handles user creation.
func (h *UserHandler) Create(c *gin.Context) {
	var req struct {
		Name  string `json:"name" binding:"required"`
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	user, err := h.service.CreateUser(c.Request.Context(), req.Name, req.Email)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer("failed to create user"))
		return
	}

	resp.WithStatusCode(c.Writer, http.StatusCreated, user)
}

// Get retrieves a user.
func (h *UserHandler) Get(c *gin.Context) {
	userID := c.Param("user_id")

	user, err := h.service.GetUser(c.Request.Context(), userID)
	if err != nil {
		resp.Fail(c.Writer, resp.NotFound("user not found"))
		return
	}

	resp.Success(c.Writer, user)
}

// List lists all users.
func (h *UserHandler) List(c *gin.Context) {
	users, err := h.service.ListUsers(c.Request.Context())
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer("failed to list users"))
		return
	}

	resp.Success(c.Writer, users)
}

// Update updates a user.
func (h *UserHandler) Update(c *gin.Context) {
	userID := c.Param("user_id")

	var req struct {
		Name  string `json:"name" binding:"required"`
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	user, err := h.service.UpdateUser(c.Request.Context(), userID, req.Name, req.Email)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer("failed to update user"))
		return
	}

	resp.Success(c.Writer, user)
}

// Delete deletes a user.
func (h *UserHandler) Delete(c *gin.Context) {
	userID := c.Param("user_id")

	if err := h.service.DeleteUser(c.Request.Context(), userID); err != nil {
		resp.Fail(c.Writer, resp.InternalServer("failed to delete user"))
		return
	}

	resp.WithStatusCode(c.Writer, http.StatusNoContent)
}
