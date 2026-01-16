// Package service contains comment business logic for the full app.
package service

import (
	"context"
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ncobase/ncore/examples/08-full-application/biz/comment/data/repository"
	"github.com/ncobase/ncore/examples/08-full-application/biz/comment/structs"
	"github.com/ncobase/ncore/examples/08-full-application/internal/event"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/resp"
)

var (
	ErrCommentNotFound = errors.New("comment not found")
)

type Service struct {
	repo   repository.CommentRepository
	bus    *event.Bus
	logger *logger.Logger
}

func NewService(logger *logger.Logger, bus *event.Bus) *Service {
	return &Service{
		bus:    bus,
		logger: logger,
	}
}

func (s *Service) SetRepository(repo repository.CommentRepository) {
	s.repo = repo
}

func (s *Service) CreateComment(ctx context.Context, workspaceID, createdBy string, req *structs.CreateCommentRequest) (*structs.Comment, error) {
	comment := &structs.Comment{
		ID:          uuid.New().String(),
		WorkspaceID: workspaceID,
		TaskID:      req.TaskID,
		Content:     req.Content,
		CreatedBy:   createdBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.Create(ctx, comment); err != nil {
		s.logger.Error(ctx, "Failed to create comment", "error", err, "task_id", req.TaskID)
		return nil, err
	}

	if err := s.bus.Publish(ctx, &event.Event{
		Type:          event.EventTypeCommentCreated,
		AggregateID:   comment.ID,
		AggregateName: "comment",
		WorkspaceID:   comment.WorkspaceID,
		UserID:        createdBy,
		Payload: map[string]any{
			"comment_id": comment.ID,
			"task_id":    comment.TaskID,
			"content":    comment.Content,
		},
	}); err != nil {
		s.logger.Error(ctx, "Failed to publish comment created event", "error", err)
	}

	s.logger.Info(ctx, "Comment created", "comment_id", comment.ID, "task_id", req.TaskID, "user_id", createdBy)
	return comment, nil
}

func (s *Service) GetComment(ctx context.Context, commentID string) (*structs.Comment, error) {
	comment, err := s.repo.FindByID(ctx, commentID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get comment", "error", err, "comment_id", commentID)
		return nil, ErrCommentNotFound
	}
	return comment, nil
}

func (s *Service) ListComments(ctx context.Context, taskID string, limit, offset int) ([]*structs.Comment, error) {
	comments, err := s.repo.FindByTask(ctx, taskID, limit, offset)
	if err != nil {
		s.logger.Error(ctx, "Failed to list comments", "error", err, "task_id", taskID)
		return nil, err
	}
	return comments, nil
}

func (s *Service) UpdateComment(ctx context.Context, commentID, userID string, req *structs.UpdateCommentRequest) (*structs.Comment, error) {
	comment, err := s.repo.FindByID(ctx, commentID)
	if err != nil {
		return nil, ErrCommentNotFound
	}

	if comment.CreatedBy != userID {
		return nil, errors.New("not authorized to update this comment")
	}

	comment.Content = req.Content
	comment.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, comment); err != nil {
		s.logger.Error(ctx, "Failed to update comment", "error", err, "comment_id", commentID)
		return nil, err
	}

	if err := s.bus.Publish(ctx, &event.Event{
		Type:          event.EventTypeCommentUpdated,
		AggregateID:   comment.ID,
		AggregateName: "comment",
		WorkspaceID:   comment.WorkspaceID,
		UserID:        userID,
		Payload: map[string]any{
			"comment_id": comment.ID,
			"task_id":    comment.TaskID,
		},
	}); err != nil {
		s.logger.Error(ctx, "Failed to publish comment updated event", "error", err)
	}

	s.logger.Info(ctx, "Comment updated", "comment_id", commentID, "user_id", userID)
	return comment, nil
}

func (s *Service) DeleteComment(ctx context.Context, commentID, userID string) error {
	comment, err := s.repo.FindByID(ctx, commentID)
	if err != nil {
		return ErrCommentNotFound
	}

	if comment.CreatedBy != userID {
		return errors.New("not authorized to delete this comment")
	}

	if err := s.repo.Delete(ctx, commentID); err != nil {
		s.logger.Error(ctx, "Failed to delete comment", "error", err, "comment_id", commentID)
		return err
	}

	if err := s.bus.Publish(ctx, &event.Event{
		Type:          event.EventTypeCommentDeleted,
		AggregateID:   comment.ID,
		AggregateName: "comment",
		WorkspaceID:   comment.WorkspaceID,
		UserID:        userID,
		Payload: map[string]any{
			"comment_id": comment.ID,
			"task_id":    comment.TaskID,
		},
	}); err != nil {
		s.logger.Error(ctx, "Failed to publish comment deleted event", "error", err)
	}

	s.logger.Info(ctx, "Comment deleted", "comment_id", commentID, "user_id", userID)
	return nil
}

func (s *Service) HandleCreate(c *gin.Context) {
	workspaceID := c.Param("workspace_id")
	userID, _ := c.Get("user_id")

	var req structs.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	comment, err := s.CreateComment(c.Request.Context(), workspaceID, userID.(string), &req)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer("failed to create comment"))
		return
	}

	resp.WithStatusCode(c.Writer, 201, comment)
}

func (s *Service) HandleGetByID(c *gin.Context) {
	commentID := c.Param("comment_id")
	if commentID == "" {
		resp.Fail(c.Writer, resp.BadRequest("comment id is required"))
		return
	}

	comment, err := s.GetComment(c.Request.Context(), commentID)
	if err != nil {
		if errors.Is(err, ErrCommentNotFound) {
			resp.Fail(c.Writer, resp.NotFound("comment not found"))
			return
		}
		resp.Fail(c.Writer, resp.InternalServer("failed to get comment"))
		return
	}

	resp.Success(c.Writer, comment)
}

func (s *Service) HandleList(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		resp.Fail(c.Writer, resp.BadRequest("task id is required"))
		return
	}

	limit := 50
	offset := 0

	comments, err := s.ListComments(c.Request.Context(), taskID, limit, offset)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer("failed to list comments"))
		return
	}

	resp.Success(c.Writer, map[string]any{
		"comments": comments,
		"limit":    limit,
		"offset":   offset,
		"count":    len(comments),
	})
}

func (s *Service) HandleUpdate(c *gin.Context) {
	commentID := c.Param("comment_id")
	userID, _ := c.Get("user_id")

	var req structs.UpdateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	comment, err := s.UpdateComment(c.Request.Context(), commentID, userID.(string), &req)
	if err != nil {
		if errors.Is(err, ErrCommentNotFound) {
			resp.Fail(c.Writer, resp.NotFound("comment not found"))
			return
		}
		if err.Error() == "not authorized to update this comment" {
			resp.Fail(c.Writer, resp.Forbidden("not authorized"))
			return
		}
		resp.Fail(c.Writer, resp.InternalServer("failed to update comment"))
		return
	}

	resp.Success(c.Writer, comment)
}

func (s *Service) HandleDelete(c *gin.Context) {
	commentID := c.Param("comment_id")
	userID, _ := c.Get("user_id")

	if err := s.DeleteComment(c.Request.Context(), commentID, userID.(string)); err != nil {
		if errors.Is(err, ErrCommentNotFound) {
			resp.Fail(c.Writer, resp.NotFound("comment not found"))
			return
		}
		if err.Error() == "not authorized to delete this comment" {
			resp.Fail(c.Writer, resp.Forbidden("not authorized"))
			return
		}
		resp.Fail(c.Writer, resp.InternalServer("failed to delete comment"))
		return
	}

	resp.Success(c.Writer, map[string]string{"message": "comment deleted"})
}
