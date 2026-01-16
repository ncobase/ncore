// Package service contains user business logic for the full app.
package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/ncobase/ncore/examples/08-full-application/core/user/data/repository"
	"github.com/ncobase/ncore/examples/08-full-application/core/user/structs"
	"github.com/ncobase/ncore/logging/logger"
)

var ErrUserNotFound = errors.New("user not found")

type Service struct {
	repo   repository.UserRepository
	logger *logger.Logger
}

func New(logger *logger.Logger, repo repository.UserRepository) *Service {
	return &Service{logger: logger, repo: repo}
}

func (s *Service) Create(ctx context.Context, req *structs.CreateUserRequest) (*structs.User, error) {
	user := &structs.User{
		ID:        uuid.New().String(),
		Name:      req.Name,
		Email:     req.Email,
		Role:      req.Role,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		s.logger.Error(ctx, "Failed to create user", "error", err)
		return nil, err
	}

	s.logger.Info(ctx, "User created", "user_id", user.ID, "email", user.Email)
	return user, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*structs.User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		s.logger.Error(ctx, "Failed to get user", "error", err, "user_id", id)
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *Service) GetByEmail(ctx context.Context, email string) (*structs.User, error) {
	return s.repo.FindByEmail(ctx, email)
}

func (s *Service) Update(ctx context.Context, id string, req *structs.UpdateUserRequest) (*structs.User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Role != "" {
		user.Role = req.Role
	}
	user.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, user); err != nil {
		s.logger.Error(ctx, "Failed to update user", "error", err, "user_id", id)
		return nil, err
	}

	s.logger.Info(ctx, "User updated", "user_id", user.ID)
	return user, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error(ctx, "Failed to delete user", "error", err, "user_id", id)
		return err
	}

	s.logger.Info(ctx, "User deleted", "user_id", id)
	return nil
}

func (s *Service) List(ctx context.Context, limit, offset int) ([]*structs.User, error) {
	users, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		s.logger.Error(ctx, "Failed to list users", "error", err)
		return nil, err
	}
	return users, nil
}

func (s *Service) Repository() repository.UserRepository {
	return s.repo
}
