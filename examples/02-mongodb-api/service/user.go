package service

import (
	"context"
	"fmt"

	"github.com/ncobase/ncore/examples/02-mongodb-api/data"
	"github.com/ncobase/ncore/examples/02-mongodb-api/data/repository"
	"github.com/ncobase/ncore/logging/logger"
)

// UserService handles user-related business logic.
type UserService struct {
	data   *data.Data
	logger *logger.Logger
}

// NewUserService creates a new user service.
func NewUserService(d *data.Data, logger *logger.Logger) *UserService {
	return &UserService{
		data:   d,
		logger: logger,
	}
}

// CreateUserRequest represents the request to create a user.
type CreateUserRequest struct {
	Name  string `json:"name" binding:"required,min=1,max=100"`
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"omitempty,oneof=admin user moderator"`
}

// UpdateUserRequest represents the request to update a user.
type UpdateUserRequest struct {
	Name  string `json:"name" binding:"required,min=1,max=100"`
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"required,oneof=admin user moderator"`
}

// CreateUser creates a new user.
func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) (*repository.User, error) {
	// Set default role if not provided
	role := req.Role
	if role == "" {
		role = "user"
	}

	// Validate role
	validRoles := map[string]bool{
		"admin":     true,
		"user":      true,
		"moderator": true,
	}
	if !validRoles[role] {
		return nil, fmt.Errorf("invalid role: %s", role)
	}

	user := &repository.User{
		Name:  req.Name,
		Email: req.Email,
		Role:  role,
	}

	return s.data.UserRepo.Create(ctx, user)
}

// GetUser retrieves a user by ID.
func (s *UserService) GetUser(ctx context.Context, id string) (*repository.User, error) {
	if id == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	return s.data.UserRepo.FindByID(ctx, id)
}

// ListUsers retrieves a paginated list of users.
func (s *UserService) ListUsers(ctx context.Context, page, pageSize int) ([]*repository.User, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	skip := int64((page - 1) * pageSize)
	limit := int64(pageSize)

	users, err := s.data.UserRepo.List(ctx, skip, limit)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.data.UserRepo.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// UpdateUser updates an existing user.
func (s *UserService) UpdateUser(ctx context.Context, id string, req *UpdateUserRequest) (*repository.User, error) {
	if id == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	// Check if user exists
	existing, err := s.data.UserRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields
	existing.Name = req.Name
	existing.Email = req.Email
	existing.Role = req.Role

	return s.data.UserRepo.Update(ctx, existing)
}

// DeleteUser deletes a user by ID.
func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("user ID is required")
	}

	return s.data.UserRepo.Delete(ctx, id)
}
