package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/examples/02-mongodb-api/service"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/resp"
)

// UserHandler handles HTTP requests for users.
type UserHandler struct {
	svc    *service.UserService
	logger *logger.Logger
}

// NewUserHandler creates a new user handler.
func NewUserHandler(svc *service.UserService, logger *logger.Logger) *UserHandler {
	return &UserHandler{
		svc:    svc,
		logger: logger,
	}
}

// Create handles user creation.
func (h *UserHandler) Create(c *gin.Context) {
	var req service.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn(c.Request.Context(), "invalid request", "error", err)
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	user, err := h.svc.CreateUser(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error(c.Request.Context(), "failed to create user", "error", err)
		resp.Fail(c.Writer, resp.InternalServer("failed to create user"))
		return
	}

	resp.WithStatusCode(c.Writer, http.StatusCreated, user)
}

// Get handles user retrieval.
func (h *UserHandler) Get(c *gin.Context) {
	userID := c.Param("user_id")

	user, err := h.svc.GetUser(c.Request.Context(), userID)
	if err != nil {
		if err.Error() == "user not found" {
			resp.Fail(c.Writer, resp.NotFound("user not found"))
			return
		}
		h.logger.Error(c.Request.Context(), "failed to get user", "id", userID, "error", err)
		resp.Fail(c.Writer, resp.InternalServer("failed to get user"))
		return
	}

	resp.Success(c.Writer, user)
}

// List handles user listing with pagination.
func (h *UserHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	users, total, err := h.svc.ListUsers(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.Error(c.Request.Context(), "failed to list users", "error", err)
		resp.Fail(c.Writer, resp.InternalServer("failed to list users"))
		return
	}

	resp.Success(c.Writer, map[string]any{
		"data": users,
		"pagination": map[string]any{
			"page":      page,
			"page_size": pageSize,
			"total":     total,
		},
	})
}

// Update handles user updates.
func (h *UserHandler) Update(c *gin.Context) {
	userID := c.Param("user_id")

	var req service.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn(c.Request.Context(), "invalid request", "error", err)
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	user, err := h.svc.UpdateUser(c.Request.Context(), userID, &req)
	if err != nil {
		if err.Error() == "user not found" {
			resp.Fail(c.Writer, resp.NotFound("user not found"))
			return
		}
		h.logger.Error(c.Request.Context(), "failed to update user", "id", userID, "error", err)
		resp.Fail(c.Writer, resp.InternalServer("failed to update user"))
		return
	}

	resp.Success(c.Writer, user)
}

// Delete handles user deletion.
func (h *UserHandler) Delete(c *gin.Context) {
	userID := c.Param("user_id")

	if err := h.svc.DeleteUser(c.Request.Context(), userID); err != nil {
		if err.Error() == "user not found" {
			resp.Fail(c.Writer, resp.NotFound("user not found"))
			return
		}
		h.logger.Error(c.Request.Context(), "failed to delete user", "id", userID, "error", err)
		resp.Fail(c.Writer, resp.InternalServer("failed to delete user"))
		return
	}

	resp.WithStatusCode(c.Writer, http.StatusNoContent)
}
