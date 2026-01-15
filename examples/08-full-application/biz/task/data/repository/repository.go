// Package repository stores tasks for the full application example.
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ncobase/ncore/examples/full-application/biz/task/structs"
)

type TaskRepository interface {
	Create(ctx context.Context, task *structs.Task) error
	FindByID(ctx context.Context, id string) (*structs.Task, error)
	Update(ctx context.Context, task *structs.Task) error
	Delete(ctx context.Context, id string) error
	FindByWorkspace(ctx context.Context, workspaceID string, limit, offset int) ([]*structs.Task, error)
	List(ctx context.Context, workspaceID string, filter map[string]any, limit, offset int) ([]*structs.Task, error)
	Assign(ctx context.Context, taskID, userID string) error
}

type taskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) (TaskRepository, error) {
	if db == nil {
		return nil, errors.New("database is nil")
	}

	repo := &taskRepository{db: db}
	if err := repo.initSchema(context.Background()); err != nil {
		return nil, err
	}

	return repo, nil
}

func (r *taskRepository) initSchema(ctx context.Context) error {
	if _, err := r.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS tasks (
			id TEXT PRIMARY KEY,
			workspace_id TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT NOT NULL,
			status TEXT NOT NULL,
			priority TEXT NOT NULL,
			assigned_to TEXT NOT NULL,
			created_by TEXT NOT NULL,
			due_date TIMESTAMPTZ NULL,
			created_at TIMESTAMPTZ NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL
		);
	`); err != nil {
		return err
	}

	if _, err := r.db.ExecContext(ctx, `
		CREATE INDEX IF NOT EXISTS idx_tasks_workspace_id ON tasks(workspace_id);
	`); err != nil {
		return err
	}

	if _, err := r.db.ExecContext(ctx, `
		CREATE INDEX IF NOT EXISTS idx_tasks_assigned_to ON tasks(assigned_to);
	`); err != nil {
		return err
	}

	return nil
}

func (r *taskRepository) Create(ctx context.Context, task *structs.Task) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO tasks (id, workspace_id, title, description, status, priority, assigned_to, created_by, due_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, task.ID, task.WorkspaceID, task.Title, task.Description, task.Status, task.Priority, task.AssignedTo, task.CreatedBy, task.DueDate, task.CreatedAt, task.UpdatedAt)
	return err
}

func (r *taskRepository) FindByID(ctx context.Context, id string) (*structs.Task, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, workspace_id, title, description, status, priority, assigned_to, created_by, due_date, created_at, updated_at
		FROM tasks WHERE id = $1
	`, id)
	return scanTask(row)
}

func (r *taskRepository) FindByWorkspace(ctx context.Context, workspaceID string, limit, offset int) ([]*structs.Task, error) {
	if limit <= 0 {
		limit = 20
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, workspace_id, title, description, status, priority, assigned_to, created_by, due_date, created_at, updated_at
		FROM tasks WHERE workspace_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, workspaceID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanTasks(rows)
}

func (r *taskRepository) FindByAssignee(ctx context.Context, assigneeID string, limit, offset int) ([]*structs.Task, error) {
	if limit <= 0 {
		limit = 20
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, workspace_id, title, description, status, priority, assigned_to, created_by, due_date, created_at, updated_at
		FROM tasks WHERE assigned_to = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, assigneeID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanTasks(rows)
}

func (r *taskRepository) Update(ctx context.Context, task *structs.Task) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE tasks
		SET title = $1, description = $2, status = $3, priority = $4, assigned_to = $5, created_by = $6, due_date = $7, updated_at = $8
		WHERE id = $9
	`, task.Title, task.Description, task.Status, task.Priority, task.AssignedTo, task.CreatedBy, task.DueDate, task.UpdatedAt, task.ID)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("task not found: %s", task.ID)
	}

	return nil
}

func (r *taskRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM tasks WHERE id = $1
	`, id)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("task not found: %s", id)
	}

	return nil
}

func (r *taskRepository) Assign(ctx context.Context, taskID, userID string) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE tasks SET assigned_to = $1, updated_at = $2 WHERE id = $3
	`, userID, time.Now(), taskID)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("task not found: %s", taskID)
	}

	return nil
}

func (r *taskRepository) List(ctx context.Context, workspaceID string, filter map[string]any, limit, offset int) ([]*structs.Task, error) {
	if limit <= 0 {
		limit = 20
	}

	query := `
		SELECT id, workspace_id, title, description, status, priority, assigned_to, created_by, due_date, created_at, updated_at
		FROM tasks
		WHERE workspace_id = $1`
	args := []any{workspaceID}
	argPos := 2

	if filter != nil {
		if status, ok := filter["status"].(string); ok && status != "" {
			query += fmt.Sprintf(" AND status = $%d", argPos)
			args = append(args, status)
			argPos++
		}
		if priority, ok := filter["priority"].(string); ok && priority != "" {
			query += fmt.Sprintf(" AND priority = $%d", argPos)
			args = append(args, priority)
			argPos++
		}
		if assignedTo, ok := filter["assigned_to"].(string); ok && assignedTo != "" {
			query += fmt.Sprintf(" AND assigned_to = $%d", argPos)
			args = append(args, assignedTo)
			argPos++
		}
	}

	query = strings.TrimSpace(query) + fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argPos, argPos+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanTasks(rows)
}

func scanTask(scanner interface{ Scan(dest ...any) error }) (*structs.Task, error) {
	var dueDate sql.NullTime
	item := &structs.Task{}
	if err := scanner.Scan(
		&item.ID,
		&item.WorkspaceID,
		&item.Title,
		&item.Description,
		&item.Status,
		&item.Priority,
		&item.AssignedTo,
		&item.CreatedBy,
		&dueDate,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return nil, err
	}

	if dueDate.Valid {
		item.DueDate = &dueDate.Time
	}

	return item, nil
}

func scanTasks(rows *sql.Rows) ([]*structs.Task, error) {
	var tasks []*structs.Task
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

type MemoryTaskRepository struct {
	tasks map[string]*structs.Task
	mu    sync.RWMutex
}

func NewMemoryTaskRepository() TaskRepository {
	return &MemoryTaskRepository{
		tasks: make(map[string]*structs.Task),
	}
}

func (r *MemoryTaskRepository) Create(ctx context.Context, task *structs.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tasks[task.ID]; exists {
		return fmt.Errorf("task already exists: %s", task.ID)
	}

	r.tasks[task.ID] = task
	return nil
}

func (r *MemoryTaskRepository) FindByID(ctx context.Context, id string) (*structs.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	task, exists := r.tasks[id]
	if !exists {
		return nil, fmt.Errorf("task not found: %s", id)
	}

	taskCopy := *task
	return &taskCopy, nil
}

func (r *MemoryTaskRepository) FindByWorkspace(ctx context.Context, workspaceID string, limit, offset int) ([]*structs.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var tasks []*structs.Task
	for _, task := range r.tasks {
		if task.WorkspaceID == workspaceID {
			taskCopy := *task
			tasks = append(tasks, &taskCopy)
		}
	}

	if offset >= len(tasks) {
		return []*structs.Task{}, nil
	}

	end := offset + limit
	if end > len(tasks) {
		end = len(tasks)
	}

	return tasks[offset:end], nil
}

func (r *MemoryTaskRepository) FindByAssignee(ctx context.Context, assigneeID string, limit, offset int) ([]*structs.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var tasks []*structs.Task
	for _, task := range r.tasks {
		if task.AssignedTo == assigneeID {
			taskCopy := *task
			tasks = append(tasks, &taskCopy)
		}
	}

	if offset >= len(tasks) {
		return []*structs.Task{}, nil
	}

	end := offset + limit
	if end > len(tasks) {
		end = len(tasks)
	}

	return tasks[offset:end], nil
}

func (r *MemoryTaskRepository) Update(ctx context.Context, task *structs.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tasks[task.ID]; !exists {
		return fmt.Errorf("task not found: %s", task.ID)
	}

	r.tasks[task.ID] = task
	return nil
}

func (r *MemoryTaskRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tasks[id]; !exists {
		return fmt.Errorf("task not found: %s", id)
	}

	delete(r.tasks, id)
	return nil
}

func (r *MemoryTaskRepository) Assign(ctx context.Context, taskID, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	task, exists := r.tasks[taskID]
	if !exists {
		return fmt.Errorf("task not found: %s", taskID)
	}

	task.AssignedTo = userID
	r.tasks[taskID] = task
	return nil
}

func (r *MemoryTaskRepository) List(ctx context.Context, workspaceID string, filter map[string]any, limit, offset int) ([]*structs.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var tasks []*structs.Task
	for _, task := range r.tasks {
		if task.WorkspaceID != workspaceID {
			continue
		}

		match := true
		for key, value := range filter {
			switch key {
			case "status":
				if task.Status != value.(string) {
					match = false
				}
			case "priority":
				if task.Priority != value.(string) {
					match = false
				}
			case "assigned_to":
				if task.AssignedTo != value.(string) {
					match = false
				}
			}
			if !match {
				break
			}
		}

		if match {
			taskCopy := *task
			tasks = append(tasks, &taskCopy)
		}
	}

	if offset >= len(tasks) {
		return []*structs.Task{}, nil
	}

	end := offset + limit
	if end > len(tasks) {
		end = len(tasks)
	}

	return tasks[offset:end], nil
}
