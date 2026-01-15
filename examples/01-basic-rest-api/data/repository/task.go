// Package repository provides task persistence for the basic REST API example.
package repository

import (
	"context"
	"fmt"

	"github.com/ncobase/ncore/examples/01-basic-rest-api/data/ent"
	"github.com/ncobase/ncore/examples/01-basic-rest-api/data/ent/task"
	"github.com/ncobase/ncore/logging/logger"
)

// TaskRepository defines the interface for task data operations.
type TaskRepository interface {
	Create(ctx context.Context, t *ent.Task) (*ent.Task, error)
	GetByID(ctx context.Context, id int) (*ent.Task, error)
	List(ctx context.Context, offset, limit int) ([]*ent.Task, error)
	Update(ctx context.Context, t *ent.Task) (*ent.Task, error)
	Delete(ctx context.Context, id int) error
	Count(ctx context.Context) (int, error)
}

type taskRepository struct {
	db     *ent.Client
	logger *logger.Logger
}

// NewTaskRepository creates a new task repository instance.
func NewTaskRepository(db *ent.Client, logger *logger.Logger) TaskRepository {
	return &taskRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new task.
func (r *taskRepository) Create(ctx context.Context, t *ent.Task) (*ent.Task, error) {
	created, err := r.db.Task.
		Create().
		SetTitle(t.Title).
		SetNillableDescription(nilString(t.Description)).
		SetStatus(t.Status).
		Save(ctx)

	if err != nil {
		r.logger.Error(ctx, "failed to create task", "error", err)
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	r.logger.Info(ctx, "task created", "id", created.ID)
	return created, nil
}

// GetByID retrieves a task by ID.
func (r *taskRepository) GetByID(ctx context.Context, id int) (*ent.Task, error) {
	t, err := r.db.Task.
		Query().
		Where(task.ID(id)).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("task not found")
		}
		r.logger.Error(ctx, "failed to get task", "id", id, "error", err)
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return t, nil
}

// List retrieves a list of tasks with pagination.
func (r *taskRepository) List(ctx context.Context, offset, limit int) ([]*ent.Task, error) {
	tasks, err := r.db.Task.
		Query().
		Order(ent.Desc(task.FieldCreatedAt)).
		Offset(offset).
		Limit(limit).
		All(ctx)

	if err != nil {
		r.logger.Error(ctx, "failed to list tasks", "error", err)
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	return tasks, nil
}

// Update updates an existing task.
func (r *taskRepository) Update(ctx context.Context, t *ent.Task) (*ent.Task, error) {
	updated, err := r.db.Task.
		UpdateOneID(t.ID).
		SetTitle(t.Title).
		SetNillableDescription(nilString(t.Description)).
		SetStatus(t.Status).
		Save(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("task not found")
		}
		r.logger.Error(ctx, "failed to update task", "id", t.ID, "error", err)
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	r.logger.Info(ctx, "task updated", "id", updated.ID)
	return updated, nil
}

// Delete deletes a task by ID.
func (r *taskRepository) Delete(ctx context.Context, id int) error {
	err := r.db.Task.
		DeleteOneID(id).
		Exec(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("task not found")
		}
		r.logger.Error(ctx, "failed to delete task", "id", id, "error", err)
		return fmt.Errorf("failed to delete task: %w", err)
	}

	r.logger.Info(ctx, "task deleted", "id", id)
	return nil
}

// Count returns the total number of tasks.
func (r *taskRepository) Count(ctx context.Context) (int, error) {
	count, err := r.db.Task.Query().Count(ctx)
	if err != nil {
		r.logger.Error(ctx, "failed to count tasks", "error", err)
		return 0, fmt.Errorf("failed to count tasks: %w", err)
	}
	return count, nil
}

// nilString converts a string to *string, returning nil if empty.
func nilString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
