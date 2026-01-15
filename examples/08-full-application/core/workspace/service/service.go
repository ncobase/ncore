// Package service contains workspace business logic for the full app.
package service

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ncobase/ncore/examples/full-application/core/workspace/data/repository"
	"github.com/ncobase/ncore/examples/full-application/core/workspace/structs"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/resp"
)

var (
	ErrWorkspaceNotFound = errors.New("workspace not found")
	ErrMemberExists      = errors.New("member already exists")
	ErrMemberNotFound    = errors.New("member not found")
	ErrNotOwner          = errors.New("not the workspace owner")
)

type Service struct {
	workspaceRepo repository.WorkspaceRepository
	memberRepo    repository.MemberRepository
	logger        *logger.Logger
}

func NewService(logger *logger.Logger) *Service {
	return &Service{
		logger: logger,
	}
}

func (s *Service) SetRepositories(workspaceRepo repository.WorkspaceRepository, memberRepo repository.MemberRepository) {
	s.workspaceRepo = workspaceRepo
	s.memberRepo = memberRepo
}

func (s *Service) CreateWorkspace(ctx context.Context, ownerID string, req *structs.CreateWorkspaceRequest) (*structs.Workspace, error) {
	workspace := &structs.Workspace{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		OwnerID:     ownerID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.workspaceRepo.Create(ctx, workspace); err != nil {
		s.logger.Error(ctx, "Failed to create workspace", "error", err)
		return nil, err
	}

	member := &structs.Member{
		ID:          uuid.New().String(),
		WorkspaceID: workspace.ID,
		UserID:      ownerID,
		Role:        "owner",
		CreatedAt:   time.Now(),
	}

	if err := s.memberRepo.Add(ctx, member); err != nil {
		s.logger.Error(ctx, "Failed to add owner as member", "error", err)
		return nil, err
	}

	s.logger.Info(ctx, "Workspace created", "workspace_id", workspace.ID, "owner_id", ownerID)
	return workspace, nil
}

func (s *Service) GetWorkspace(ctx context.Context, workspaceID string) (*structs.Workspace, error) {
	workspace, err := s.workspaceRepo.FindByID(ctx, workspaceID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get workspace", "error", err, "workspace_id", workspaceID)
		return nil, ErrWorkspaceNotFound
	}
	return workspace, nil
}

func (s *Service) ListWorkspaces(ctx context.Context, userID string, limit, offset int) ([]*structs.Workspace, error) {
	workspaces, err := s.workspaceRepo.List(ctx, userID, limit, offset)
	if err != nil {
		s.logger.Error(ctx, "Failed to list workspaces", "error", err)
		return nil, err
	}
	return workspaces, nil
}

func (s *Service) AddMember(ctx context.Context, workspaceID, userID string, req *structs.AddMemberRequest) error {
	isMember, err := s.memberRepo.IsMember(ctx, workspaceID, userID)
	if err != nil {
		return err
	}

	role, err := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if err != nil {
		return err
	}

	if !isMember || (role != "owner" && role != "admin") {
		return ErrNotOwner
	}

	exists, err := s.memberRepo.IsMember(ctx, workspaceID, req.UserID)
	if err != nil {
		return err
	}
	if exists {
		return ErrMemberExists
	}

	member := &structs.Member{
		ID:          uuid.New().String(),
		WorkspaceID: workspaceID,
		UserID:      req.UserID,
		Role:        req.Role,
		CreatedAt:   time.Now(),
	}

	if err := s.memberRepo.Add(ctx, member); err != nil {
		s.logger.Error(ctx, "Failed to add member", "error", err)
		return err
	}

	s.logger.Info(ctx, "Member added", "workspace_id", workspaceID, "user_id", req.UserID)
	return nil
}

func (s *Service) RemoveMember(ctx context.Context, workspaceID, requestorID, memberID string) error {
	role, err := s.memberRepo.GetRole(ctx, workspaceID, requestorID)
	if err != nil {
		return err
	}

	memberRole, err := s.memberRepo.GetRole(ctx, workspaceID, memberID)
	if err != nil {
		return err
	}
	if memberRole == "owner" {
		return errors.New("cannot remove workspace owner")
	}

	if role != "owner" && role != "admin" {
		return ErrNotOwner
	}

	if role == "admin" && memberRole == "admin" {
		return errors.New("admins cannot remove other admins")
	}

	if err := s.memberRepo.Remove(ctx, workspaceID, memberID); err != nil {
		s.logger.Error(ctx, "Failed to remove member", "error", err)
		return err
	}

	s.logger.Info(ctx, "Member removed", "workspace_id", workspaceID, "user_id", memberID)
	return nil
}

func (s *Service) ListMembers(ctx context.Context, workspaceID string) ([]*structs.Member, error) {
	members, err := s.memberRepo.FindByWorkspace(ctx, workspaceID)
	if err != nil {
		s.logger.Error(ctx, "Failed to list members", "error", err)
		return nil, err
	}
	return members, nil
}

func (s *Service) DeleteWorkspace(ctx context.Context, workspaceID, userID string) error {
	workspace, err := s.workspaceRepo.FindByID(ctx, workspaceID)
	if err != nil {
		return ErrWorkspaceNotFound
	}

	if workspace.OwnerID != userID {
		return ErrNotOwner
	}

	if err := s.workspaceRepo.Delete(ctx, workspaceID); err != nil {
		s.logger.Error(ctx, "Failed to delete workspace", "error", err)
		return err
	}

	s.logger.Info(ctx, "Workspace deleted", "workspace_id", workspaceID, "user_id", userID)
	return nil
}

func (s *Service) HandleCreate(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		resp.Fail(c.Writer, resp.UnAuthorized("not authenticated"))
		return
	}

	var req structs.CreateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	workspace, err := s.CreateWorkspace(c.Request.Context(), userID.(string), &req)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer("failed to create workspace"))
		return
	}

	resp.WithStatusCode(c.Writer, 201, workspace)
}

func (s *Service) HandleGetByID(c *gin.Context) {
	workspaceID := c.Param("workspace_id")
	if workspaceID == "" {
		resp.Fail(c.Writer, resp.BadRequest("workspace id is required"))
		return
	}

	workspace, err := s.GetWorkspace(c.Request.Context(), workspaceID)
	if err != nil {
		if errors.Is(err, ErrWorkspaceNotFound) {
			resp.Fail(c.Writer, resp.NotFound("workspace not found"))
			return
		}
		resp.Fail(c.Writer, resp.InternalServer("failed to get workspace"))
		return
	}

	resp.Success(c.Writer, workspace)
}

func (s *Service) HandleList(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		resp.Fail(c.Writer, resp.UnAuthorized("not authenticated"))
		return
	}

	limit := 20
	offset := 0

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	workspaces, err := s.ListWorkspaces(c.Request.Context(), userID.(string), limit, offset)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer("failed to list workspaces"))
		return
	}

	resp.Success(c.Writer, map[string]any{
		"workspaces": workspaces,
		"limit":      limit,
		"offset":     offset,
		"count":      len(workspaces),
	})
}

func (s *Service) HandleAddMember(c *gin.Context) {
	workspaceID := c.Param("workspace_id")
	userID, _ := c.Get("user_id")

	var req structs.AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	if err := s.AddMember(c.Request.Context(), workspaceID, userID.(string), &req); err != nil {
		if errors.Is(err, ErrMemberExists) {
			resp.Fail(c.Writer, resp.Conflict("member already exists"))
			return
		}
		if errors.Is(err, ErrNotOwner) {
			resp.Fail(c.Writer, resp.Forbidden("insufficient permissions"))
			return
		}
		resp.Fail(c.Writer, resp.InternalServer("failed to add member"))
		return
	}

	resp.WithStatusCode(c.Writer, 201, map[string]string{"message": "member added"})
}

func (s *Service) HandleListMembers(c *gin.Context) {
	workspaceID := c.Param("workspace_id")

	members, err := s.ListMembers(c.Request.Context(), workspaceID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer("failed to list members"))
		return
	}

	resp.Success(c.Writer, map[string]any{
		"members": members,
		"count":   len(members),
	})
}
