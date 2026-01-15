// Package structs defines user domain models for the full app.
package structs

import "time"

type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateUserRequest struct {
	Name  string `json:"name" binding:"required,min=2"`
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"required,oneof=user admin"`
}

type UpdateUserRequest struct {
	Name  string `json:"name" binding:"omitempty,min=2"`
	Email string `json:"email" binding:"omitempty,email"`
	Role  string `json:"role" binding:"omitempty,oneof=user admin"`
}
