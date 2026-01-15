// Package structs defines workspace domain models for the full app.
package structs

import "time"

type Workspace struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	OwnerID     string    `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Member struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	UserID      string    `json:"user_id"`
	Role        string    `json:"role"`
	CreatedAt   time.Time `json:"created_at"`
}

type CreateWorkspaceRequest struct {
	Name        string `json:"name" binding:"required,min=2"`
	Description string `json:"description"`
}

type UpdateWorkspaceRequest struct {
	Name        string `json:"name" binding:"omitempty,min=2"`
	Description string `json:"description"`
}

type AddMemberRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"required,oneof=admin member"`
}
