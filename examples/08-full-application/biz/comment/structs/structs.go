// Package structs defines comment domain models for the full app.
package structs

import "time"

type Comment struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	TaskID      string    `json:"task_id"`
	Content     string    `json:"content"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateCommentRequest struct {
	TaskID  string `json:"task_id" binding:"required"`
	Content string `json:"content" binding:"required,min=1,max=2000"`
}

type UpdateCommentRequest struct {
	Content string `json:"content" binding:"required,min=1,max=2000"`
}
