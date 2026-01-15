package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/examples/03-multi-module/biz/comment/service"
	"github.com/ncobase/ncore/examples/03-multi-module/biz/comment/structs"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/resp"
)

type CommentHandler struct {
	service *service.CommentService
	logger  *logger.Logger
}

func NewCommentHandler(svc *service.CommentService, logger *logger.Logger) *CommentHandler {
	return &CommentHandler{
		service: svc,
		logger:  logger,
	}
}

func (h *CommentHandler) HandleCreate(c *gin.Context) {
	postID := c.Param("post_id")

	var req structs.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	comment, err := h.service.CreateComment(c.Request.Context(), postID, req.UserID, req.Content)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer("failed to create comment"))
		return
	}

	resp.WithStatusCode(c.Writer, http.StatusCreated, comment)
}

func (h *CommentHandler) HandleList(c *gin.Context) {
	postID := c.Param("post_id")

	comments, err := h.service.ListComments(c.Request.Context(), postID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer("failed to list comments"))
		return
	}

	resp.Success(c.Writer, comments)
}

func (h *CommentHandler) HandleDelete(c *gin.Context) {
	commentID := c.Param("comment_id")

	if err := h.service.DeleteComment(c.Request.Context(), commentID); err != nil {
		resp.Fail(c.Writer, resp.InternalServer("failed to delete comment"))
		return
	}

	resp.WithStatusCode(c.Writer, http.StatusNoContent)
}
