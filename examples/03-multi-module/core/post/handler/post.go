// Package handler provides post HTTP handlers for the multi-module example.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/examples/03-multi-module/core/post/service"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/resp"
)

// PostHandler handles post HTTP requests.
type PostHandler struct {
	service *service.PostService
	logger  *logger.Logger
}

// NewPostHandler creates a new post handler.
func NewPostHandler(svc *service.PostService, logger *logger.Logger) *PostHandler {
	return &PostHandler{
		service: svc,
		logger:  logger,
	}
}

// Create handles post creation.
func (h *PostHandler) Create(c *gin.Context) {
	var req struct {
		UserID  string `json:"user_id" binding:"required"`
		Title   string `json:"title" binding:"required"`
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	post, err := h.service.CreatePost(c.Request.Context(), req.UserID, req.Title, req.Content)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer("failed to create post"))
		return
	}

	resp.WithStatusCode(c.Writer, http.StatusCreated, post)
}

// Get retrieves a post.
func (h *PostHandler) Get(c *gin.Context) {
	id := c.Param("id")

	post, err := h.service.GetPost(c.Request.Context(), id)
	if err != nil {
		resp.Fail(c.Writer, resp.NotFound("post not found"))
		return
	}

	resp.Success(c.Writer, post)
}

// List lists all posts.
func (h *PostHandler) List(c *gin.Context) {
	posts, err := h.service.ListPosts(c.Request.Context())
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer("failed to list posts"))
		return
	}

	resp.Success(c.Writer, posts)
}

// Update updates a post.
func (h *PostHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Title   string `json:"title" binding:"required"`
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	post, err := h.service.UpdatePost(c.Request.Context(), id, req.Title, req.Content)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer("failed to update post"))
		return
	}

	resp.Success(c.Writer, post)
}

// Delete deletes a post.
func (h *PostHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeletePost(c.Request.Context(), id); err != nil {
		resp.Fail(c.Writer, resp.InternalServer("failed to delete post"))
		return
	}

	resp.WithStatusCode(c.Writer, http.StatusNoContent)
}
