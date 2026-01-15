// Package structs defines export job domain models.
package structs

import "time"

type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
)

type Job struct {
	ID          string     `json:"id"`
	WorkspaceID string     `json:"workspace_id"`
	UserID      string     `json:"user_id"`
	Type        string     `json:"type"`
	Format      string     `json:"format"`
	Status      JobStatus  `json:"status"`
	FilePath    string     `json:"file_path,omitempty"`
	Error       string     `json:"error,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type CreateExportRequest struct {
	Type   string `json:"type" binding:"required,oneof=tasks comments"`
	Format string `json:"format" binding:"required,oneof=csv json"`
}
