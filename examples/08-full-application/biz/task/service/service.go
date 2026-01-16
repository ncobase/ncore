// Package service contains task business logic for the full app.
package service

import (
	"context"
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ncobase/ncore/examples/08-full-application/biz/task/data/repository"
	"github.com/ncobase/ncore/examples/08-full-application/biz/task/structs"
	"github.com/ncobase/ncore/examples/08-full-application/internal/event"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/resp"
)

var (
	ErrTaskNotFound     = errors.New("task not found")
	ErrWorkspaceInvalid = errors.New("workspace invalid")
	ErrAccessDenied     = errors.New("access denied")
)

type Service struct {
	repo   repository.TaskRepository
	bus    *event.Bus
	logger *logger.Logger
}

func NewService(logger *logger.Logger, bus *event.Bus) *Service {
	return &Service{
		bus:    bus,
		logger: logger,
	}
}

func (s *Service) SetRepository(repo repository.TaskRepository) {
	s.repo = repo
}

func (s *Service) CreateTask(ctx context.Context, workspaceID, createdBy string, req *structs.CreateTaskRequest) (*structs.Task, error) {
	task := &structs.Task{
		ID:          uuid.New().String(),
		WorkspaceID: workspaceID,
		Title:       req.Title,
		Description: req.Description,
		Status:      "pending",
		Priority:    req.Priority,
		AssignedTo:  req.AssignedTo,
		CreatedBy:   createdBy,
		DueDate:     req.DueDate,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if req.Status != "" {
		task.Status = req.Status
	}
	if req.Priority == "" {
		task.Priority = "medium"
	}

	if err := s.repo.Create(ctx, task); err != nil {
		s.logger.Error(ctx, "Failed to create task", "error", err, "workspace_id", workspaceID)
		return nil, err
	}

	if err := s.bus.Publish(ctx, &event.Event{
		Type:          event.EventTypeTaskCreated,
		AggregateID:   task.ID,
		AggregateName: "task",
		WorkspaceID:   task.WorkspaceID,
		UserID:        createdBy,
		Payload: map[string]any{
			"task_id":     task.ID,
			"title":       task.Title,
			"assigned_to": task.AssignedTo,
			"priority":    task.Priority,
			"status":      task.Status,
		},
	}); err != nil {
		s.logger.Error(ctx, "Failed to publish task created event", "error", err)
	}

	s.logger.Info(ctx, "Task created", "task_id", task.ID, "workspace_id", workspaceID, "created_by", createdBy)
	return task, nil
}

func (s *Service) GetTask(ctx context.Context, taskID string) (*structs.Task, error) {
	task, err := s.repo.FindByID(ctx, taskID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get task", "error", err, "task_id", taskID)
		return nil, ErrTaskNotFound
	}
	return task, nil
}

func (s *Service) ListTasks(ctx context.Context, workspaceID string, filter map[string]any, limit, offset int) ([]*structs.Task, error) {
	tasks, err := s.repo.List(ctx, workspaceID, filter, limit, offset)
	if err != nil {
		s.logger.Error(ctx, "Failed to list tasks", "error", err, "workspace_id", workspaceID)
		return nil, err
	}
	return tasks, nil
}

func (s *Service) UpdateTask(ctx context.Context, taskID, userID string, req *structs.UpdateTaskRequest) (*structs.Task, error) {
	task, err := s.repo.FindByID(ctx, taskID)
	if err != nil {
		return nil, ErrTaskNotFound
	}

	if req.Title != "" {
		task.Title = req.Title
	}
	if req.Description != "" {
		task.Description = req.Description
	}
	if req.Status != "" {
		task.Status = req.Status
	}
	if req.Priority != "" {
		task.Priority = req.Priority
	}
	if req.AssignedTo != "" {
		task.AssignedTo = req.AssignedTo
	}
	if req.DueDate != nil {
		task.DueDate = req.DueDate
	}
	task.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, task); err != nil {
		s.logger.Error(ctx, "Failed to update task", "error", err, "task_id", taskID)
		return nil, err
	}

	if err := s.bus.Publish(ctx, &event.Event{
		Type:          event.EventTypeTaskUpdated,
		AggregateID:   task.ID,
		AggregateName: "task",
		WorkspaceID:   task.WorkspaceID,
		UserID:        userID,
		Payload: map[string]any{
			"task_id":     task.ID,
			"title":       task.Title,
			"status":      task.Status,
			"assigned_to": task.AssignedTo,
		},
	}); err != nil {
		s.logger.Error(ctx, "Failed to publish task updated event", "error", err)
	}

	s.logger.Info(ctx, "Task updated", "task_id", task.ID, "user_id", userID)
	return task, nil
}

func (s *Service) DeleteTask(ctx context.Context, taskID, userID string) error {
	task, err := s.repo.FindByID(ctx, taskID)
	if err != nil {
		return ErrTaskNotFound
	}

	if err := s.repo.Delete(ctx, taskID); err != nil {
		s.logger.Error(ctx, "Failed to delete task", "error", err, "task_id", taskID)
		return err
	}

	if err := s.bus.Publish(ctx, &event.Event{
		Type:          event.EventTypeTaskDeleted,
		AggregateID:   task.ID,
		AggregateName: "task",
		WorkspaceID:   task.WorkspaceID,
		UserID:        userID,
		Payload: map[string]any{
			"task_id": task.ID,
		},
	}); err != nil {
		s.logger.Error(ctx, "Failed to publish task deleted event", "error", err)
	}

	s.logger.Info(ctx, "Task deleted", "task_id", taskID, "user_id", userID)
	return nil
}

func (s *Service) AssignTask(ctx context.Context, taskID, userID, assigneeID string) (*structs.Task, error) {
	task, err := s.repo.FindByID(ctx, taskID)
	if err != nil {
		return nil, ErrTaskNotFound
	}

	task.AssignedTo = assigneeID
	task.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, task); err != nil {
		s.logger.Error(ctx, "Failed to assign task", "error", err, "task_id", taskID)
		return nil, err
	}

	if err := s.bus.Publish(ctx, &event.Event{
		Type:          event.EventTypeTaskAssigned,
		AggregateID:   task.ID,
		AggregateName: "task",
		WorkspaceID:   task.WorkspaceID,
		UserID:        userID,
		Payload: map[string]any{
			"task_id":     task.ID,
			"title":       task.Title,
			"assigned_to": assigneeID,
		},
	}); err != nil {
		s.logger.Error(ctx, "Failed to publish task assigned event", "error", err)
	}

	s.logger.Info(ctx, "Task assigned", "task_id", taskID, "assignee_id", assigneeID, "assigned_by", userID)
	return task, nil
}

func (s *Service) HandleCreate(c *gin.Context) {
	workspaceID := c.Param("workspace_id")
	userID, _ := c.Get("user_id")

	var req structs.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	task, err := s.CreateTask(c.Request.Context(), workspaceID, userID.(string), &req)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer("failed to create task"))
		return
	}

	resp.WithStatusCode(c.Writer, 201, task)
}

func (s *Service) HandleGetByID(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		resp.Fail(c.Writer, resp.BadRequest("task id is required"))
		return
	}

	task, err := s.GetTask(c.Request.Context(), taskID)
	if err != nil {
		if errors.Is(err, ErrTaskNotFound) {
			resp.Fail(c.Writer, resp.NotFound("task not found"))
			return
		}
		resp.Fail(c.Writer, resp.InternalServer("failed to get task"))
		return
	}

	resp.Success(c.Writer, task)
}

func (s *Service) HandleList(c *gin.Context) {
	workspaceID := c.Param("workspace_id")

	limit := 20
	offset := 0
	filter := make(map[string]any)

	if status := c.Query("status"); status != "" {
		filter["status"] = status
	}
	if priority := c.Query("priority"); priority != "" {
		filter["priority"] = priority
	}
	if assignedTo := c.Query("assigned_to"); assignedTo != "" {
		filter["assigned_to"] = assignedTo
	}

	tasks, err := s.ListTasks(c.Request.Context(), workspaceID, filter, limit, offset)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer("failed to list tasks"))
		return
	}

	resp.Success(c.Writer, map[string]any{
		"tasks":  tasks,
		"limit":  limit,
		"offset": offset,
		"count":  len(tasks),
	})
}

func (s *Service) HandleUpdate(c *gin.Context) {
	taskID := c.Param("task_id")
	userID, _ := c.Get("user_id")

	var req structs.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	task, err := s.UpdateTask(c.Request.Context(), taskID, userID.(string), &req)
	if err != nil {
		if errors.Is(err, ErrTaskNotFound) {
			resp.Fail(c.Writer, resp.NotFound("task not found"))
			return
		}
		resp.Fail(c.Writer, resp.InternalServer("failed to update task"))
		return
	}

	resp.Success(c.Writer, task)
}

func (s *Service) HandleDelete(c *gin.Context) {
	taskID := c.Param("task_id")
	userID, _ := c.Get("user_id")

	if err := s.DeleteTask(c.Request.Context(), taskID, userID.(string)); err != nil {
		if errors.Is(err, ErrTaskNotFound) {
			resp.Fail(c.Writer, resp.NotFound("task not found"))
			return
		}
		resp.Fail(c.Writer, resp.InternalServer("failed to delete task"))
		return
	}

	resp.Success(c.Writer, map[string]string{"message": "task deleted"})
}

func (s *Service) HandleAssign(c *gin.Context) {
	taskID := c.Param("task_id")
	userID, _ := c.Get("user_id")

	var req struct {
		AssigneeID string `json:"assignee_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	task, err := s.AssignTask(c.Request.Context(), taskID, userID.(string), req.AssigneeID)
	if err != nil {
		if errors.Is(err, ErrTaskNotFound) {
			resp.Fail(c.Writer, resp.NotFound("task not found"))
			return
		}
		resp.Fail(c.Writer, resp.InternalServer("failed to assign task"))
		return
	}

	resp.Success(c.Writer, task)
}
