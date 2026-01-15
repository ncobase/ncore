package service

import (
	"context"
	"fmt"

	"github.com/ncobase/ncore/examples/03-multi-module/biz/comment/data/repository"
	"github.com/ncobase/ncore/examples/03-multi-module/biz/comment/structs"
	"github.com/ncobase/ncore/examples/03-multi-module/biz/comment/wrapper"
	"github.com/ncobase/ncore/logging/logger"
)

type CommentService struct {
	repo        repository.CommentRepository
	userWrapper *wrapper.UserServiceWrapper
	postWrapper *wrapper.PostServiceWrapper
	logger      *logger.Logger
}

func NewCommentService(repo repository.CommentRepository, userWrapper *wrapper.UserServiceWrapper, postWrapper *wrapper.PostServiceWrapper, logger *logger.Logger) *CommentService {
	return &CommentService{
		repo:        repo,
		userWrapper: userWrapper,
		postWrapper: postWrapper,
		logger:      logger,
	}
}

func (s *CommentService) CreateComment(ctx context.Context, postID, userID, content string) (*structs.Comment, error) {
	user, err := s.userWrapper.GetUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	post, err := s.postWrapper.GetPost(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("post not found: %w", err)
	}

	comment := &structs.Comment{
		PostID:  post.ID,
		UserID:  user.ID,
		Content: content,
	}

	created, err := s.repo.Create(ctx, comment)
	if err != nil {
		s.logger.Error(ctx, "failed to create comment", "error", err)
		return nil, err
	}

	return created, nil
}

func (s *CommentService) ListComments(ctx context.Context, postID string) ([]*structs.Comment, error) {
	if _, err := s.postWrapper.GetPost(ctx, postID); err != nil {
		return nil, fmt.Errorf("post not found: %w", err)
	}

	return s.repo.ListByPost(ctx, postID)
}

func (s *CommentService) DeleteComment(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *CommentService) GetComment(ctx context.Context, id string) (*structs.Comment, error) {
	return s.repo.FindByID(ctx, id)
}
