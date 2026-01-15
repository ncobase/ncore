// Package structs defines task domain models for the full app.
package structs

import "time"

type Task struct {
	ID          string     `json:"id"`
	WorkspaceID string     `json:"workspace_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority"`
	AssignedTo  string     `json:"assigned_to"`
	CreatedBy   string     `json:"created_by"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type CreateTaskRequest struct {
	Title       string     `json:"title" binding:"required,min=2"`
	Description string     `json:"description" binding:"omitempty,max=2000"`
	Status      string     `json:"status" binding:"omitempty,oneof=pending in_progress completed"`
	Priority    string     `json:"priority" binding:"omitempty,oneof=low medium high"`
	AssignedTo  string     `json:"assigned_to" binding:"omitempty"`
	DueDate     *time.Time `json:"due_date" binding:"omitempty"`
}

type UpdateTaskRequest struct {
	Title       string     `json:"title" binding:"omitempty,min=2"`
	Description string     `json:"description" binding:"omitempty,max=2000"`
	Status      string     `json:"status" binding:"omitempty,oneof=pending in_progress completed"`
	Priority    string     `json:"priority" binding:"omitempty,oneof=low medium high"`
	AssignedTo  string     `json:"assigned_to" binding:"omitempty"`
	DueDate     *time.Time `json:"due_date" binding:"omitempty"`
}
