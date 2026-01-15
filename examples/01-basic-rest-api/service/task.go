package service

import (
	"context"
	"fmt"

	"github.com/ncobase/ncore/examples/01-basic-rest-api/data"
	"github.com/ncobase/ncore/examples/01-basic-rest-api/data/ent"
	"github.com/ncobase/ncore/examples/01-basic-rest-api/data/ent/task"
	"github.com/ncobase/ncore/logging/logger"
)

// TaskService handles task-related business logic.
type TaskService struct {
	data   *data.Data
	logger *logger.Logger
}

// NewTaskService creates a new task service.
func NewTaskService(d *data.Data, logger *logger.Logger) *TaskService {
	return &TaskService{
		data:   d,
		logger: logger,
	}
}

// CreateTaskRequest represents the request to create a task.
type CreateTaskRequest struct {
	Title       string `json:"title" binding:"required,min=1,max=255"`
	Description string `json:"description"`
	Status      string `json:"status" binding:"omitempty,oneof=pending in_progress completed cancelled"`
}

// UpdateTaskRequest represents the request to update a task.
type UpdateTaskRequest struct {
	Title       string `json:"title" binding:"required,min=1,max=255"`
	Description string `json:"description"`
	Status      string `json:"status" binding:"required,oneof=pending in_progress completed cancelled"`
}

// CreateTask creates a new task.
func (s *TaskService) CreateTask(ctx context.Context, req *CreateTaskRequest) (*ent.Task, error) {
	// Set default status if not provided
	status := req.Status
	if status == "" {
		status = "pending"
	}

	// Validate status
	validStatus := map[string]bool{
		"pending":     true,
		"in_progress": true,
		"completed":   true,
		"cancelled":   true,
	}
	if !validStatus[status] {
		return nil, fmt.Errorf("invalid status: %s", status)
	}

	t := &ent.Task{
		Title:       req.Title,
		Description: req.Description,
		Status:      task.Status(status),
	}

	return s.data.TaskRepo.Create(ctx, t)
}

// GetTask retrieves a task by ID.
func (s *TaskService) GetTask(ctx context.Context, id int) (*ent.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid task ID: %d", id)
	}

	return s.data.TaskRepo.GetByID(ctx, id)
}

// ListTasks retrieves a paginated list of tasks.
func (s *TaskService) ListTasks(ctx context.Context, page, pageSize int) ([]*ent.Task, int, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	tasks, err := s.data.TaskRepo.List(ctx, offset, pageSize)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.data.TaskRepo.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// UpdateTask updates an existing task.
func (s *TaskService) UpdateTask(ctx context.Context, id int, req *UpdateTaskRequest) (*ent.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid task ID: %d", id)
	}

	// Check if task exists
	existing, err := s.data.TaskRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields
	existing.Title = req.Title
	existing.Description = req.Description
	existing.Status = task.Status(req.Status)

	return s.data.TaskRepo.Update(ctx, existing)
}

// DeleteTask deletes a task by ID.
func (s *TaskService) DeleteTask(ctx context.Context, id int) error {
	if id <= 0 {
		return fmt.Errorf("invalid task ID: %d", id)
	}

	return s.data.TaskRepo.Delete(ctx, id)
}
