// Package service contains user business logic for the multi-module example.
package service

import (
	"context"

	"github.com/ncobase/ncore/examples/03-multi-module/core/user/data/repository"
	"github.com/ncobase/ncore/examples/03-multi-module/core/user/structs"
	"github.com/ncobase/ncore/logging/logger"
)

type UserService struct {
	repo   repository.UserRepository
	logger *logger.Logger
}

func NewUserService(repo repository.UserRepository, logger *logger.Logger) *UserService {
	return &UserService{
		repo:   repo,
		logger: logger,
	}
}

func (s *UserService) CreateUser(ctx context.Context, name, email string) (*structs.User, error) {
	user := &structs.User{
		Name:  name,
		Email: email,
	}

	created, err := s.repo.Create(ctx, user)
	if err != nil {
		s.logger.Error(ctx, "failed to create user", "error", err)
		return nil, err
	}

	return created, nil
}

func (s *UserService) GetUser(ctx context.Context, id string) (*structs.User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) ListUsers(ctx context.Context) ([]*structs.User, error) {
	users, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (s *UserService) UpdateUser(ctx context.Context, id, name, email string) (*structs.User, error) {
	user, err := s.repo.Update(ctx, id, name, email)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
