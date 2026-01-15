// Package service contains post business logic for the multi-module example.
package service

import (
	"context"
	"fmt"

	"github.com/ncobase/ncore/examples/03-multi-module/core/post/data/repository"
	"github.com/ncobase/ncore/examples/03-multi-module/core/post/structs"
	"github.com/ncobase/ncore/examples/03-multi-module/core/post/wrapper"
	"github.com/ncobase/ncore/logging/logger"
)

type PostService struct {
	repo        repository.PostRepository
	userWrapper *wrapper.UserServiceWrapper
	logger      *logger.Logger
}

func NewPostService(repo repository.PostRepository, userWrapper *wrapper.UserServiceWrapper, logger *logger.Logger) *PostService {
	return &PostService{
		repo:        repo,
		userWrapper: userWrapper,
		logger:      logger,
	}
}

func (s *PostService) CreatePost(ctx context.Context, userID, title, content string) (*structs.Post, error) {
	if _, err := s.userWrapper.GetUser(ctx, userID); err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	post := &structs.Post{
		UserID:  userID,
		Title:   title,
		Content: content,
	}

	created, err := s.repo.Create(ctx, post)
	if err != nil {
		s.logger.Error(ctx, "failed to create post", "error", err)
		return nil, err
	}

	return created, nil
}

func (s *PostService) GetPost(ctx context.Context, id string) (*structs.Post, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *PostService) ListPosts(ctx context.Context) ([]*structs.Post, error) {
	return s.repo.List(ctx)
}

func (s *PostService) UpdatePost(ctx context.Context, id, title, content string) (*structs.Post, error) {
	return s.repo.Update(ctx, id, title, content)
}

func (s *PostService) DeletePost(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
