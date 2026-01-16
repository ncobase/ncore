// Package handler provides user HTTP handlers for the full app.
package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/examples/08-full-application/core/user/service"
	"github.com/ncobase/ncore/examples/08-full-application/core/user/structs"
	"github.com/ncobase/ncore/net/resp"
)

type Handler struct {
	service *service.Service
}

func New(svc *service.Service) *Handler {
	return &Handler{service: svc}
}

func (h *Handler) HandleCreate(c *gin.Context) {
	var req structs.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	user, err := h.service.Create(c.Request.Context(), &req)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer("failed to create user"))
		return
	}

	resp.WithStatusCode(c.Writer, 201, user)
}

func (h *Handler) HandleGetByID(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		resp.Fail(c.Writer, resp.BadRequest("user id is required"))
		return
	}

	user, err := h.service.GetByID(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			resp.Fail(c.Writer, resp.NotFound("user not found"))
			return
		}
		resp.Fail(c.Writer, resp.InternalServer("failed to get user"))
		return
	}

	resp.Success(c.Writer, user)
}

func (h *Handler) HandleUpdate(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		resp.Fail(c.Writer, resp.BadRequest("user id is required"))
		return
	}

	var req structs.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	user, err := h.service.Update(c.Request.Context(), userID, &req)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			resp.Fail(c.Writer, resp.NotFound("user not found"))
			return
		}
		resp.Fail(c.Writer, resp.InternalServer("failed to update user"))
		return
	}

	resp.Success(c.Writer, user)
}

func (h *Handler) HandleDelete(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		resp.Fail(c.Writer, resp.BadRequest("user id is required"))
		return
	}

	if err := h.service.Delete(c.Request.Context(), userID); err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			resp.Fail(c.Writer, resp.NotFound("user not found"))
			return
		}
		resp.Fail(c.Writer, resp.InternalServer("failed to delete user"))
		return
	}

	resp.WithStatusCode(c.Writer, 204)
}

func (h *Handler) HandleList(c *gin.Context) {
	limit := 20
	offset := 0

	users, err := h.service.List(c.Request.Context(), limit, offset)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer("failed to list users"))
		return
	}

	resp.Success(c.Writer, map[string]any{
		"users":  users,
		"limit":  limit,
		"offset": offset,
		"count":  len(users),
	})
}
