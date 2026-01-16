// Package repository stores workspaces and members for the full app.
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ncobase/ncore/data/cache"
	"github.com/ncobase/ncore/examples/08-full-application/core/workspace/structs"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/redis/go-redis/v9"
)

type WorkspaceRepository interface {
	Create(ctx context.Context, workspace *structs.Workspace) error
	FindByID(ctx context.Context, id string) (*structs.Workspace, error)
	FindByOwner(ctx context.Context, ownerID string) ([]*structs.Workspace, error)
	Update(ctx context.Context, workspace *structs.Workspace) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, userID string, limit, offset int) ([]*structs.Workspace, error)
}

type MemberRepository interface {
	Add(ctx context.Context, member *structs.Member) error
	Remove(ctx context.Context, workspaceID, userID string) error
	FindByWorkspace(ctx context.Context, workspaceID string) ([]*structs.Member, error)
	FindByUser(ctx context.Context, userID string) ([]*structs.Member, error)
	UpdateRole(ctx context.Context, workspaceID, userID, role string) error
	IsMember(ctx context.Context, workspaceID, userID string) (bool, error)
	GetRole(ctx context.Context, workspaceID, userID string) (string, error)
}

type workspaceRepository struct {
	db     *sql.DB
	logger *logger.Logger
	cache  *cache.Cache[structs.Workspace]
}

func NewWorkspaceRepository(db *sql.DB, logger *logger.Logger, rc *redis.Client) (WorkspaceRepository, error) {
	if db == nil {
		return nil, errors.New("database is nil")
	}

	repo := &workspaceRepository{db: db, logger: logger}
	if rc != nil {
		repo.cache = cache.NewCache[structs.Workspace](rc, "workspaces")
	}

	if err := repo.initSchema(context.Background()); err != nil {
		return nil, err
	}

	return repo, nil
}

func (r *workspaceRepository) initSchema(ctx context.Context) error {
	if _, err := r.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS workspaces (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT NOT NULL,
			owner_id TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL
		);
	`); err != nil {
		return err
	}

	if _, err := r.db.ExecContext(ctx, `
		CREATE INDEX IF NOT EXISTS idx_workspaces_owner_id ON workspaces(owner_id);
	`); err != nil {
		return err
	}

	return nil
}

func (r *workspaceRepository) Create(ctx context.Context, workspace *structs.Workspace) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO workspaces (id, name, description, owner_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, workspace.ID, workspace.Name, workspace.Description, workspace.OwnerID, workspace.CreatedAt, workspace.UpdatedAt)
	if err != nil {
		return err
	}

	if r.cache != nil {
		_ = r.cache.Set(ctx, workspace.ID, workspace, 10*time.Minute)
	}

	r.logger.Debug(ctx, "Workspace created in Postgres", "workspace_id", workspace.ID)
	return nil
}

func (r *workspaceRepository) FindByID(ctx context.Context, id string) (*structs.Workspace, error) {
	if r.cache != nil {
		cached, err := r.cache.Get(ctx, id)
		if err == nil && cached != nil {
			return cached, nil
		}
	}

	workspace := &structs.Workspace{}
	row := r.db.QueryRowContext(ctx, `
		SELECT id, name, description, owner_id, created_at, updated_at
		FROM workspaces WHERE id = $1
	`, id)
	if err := row.Scan(&workspace.ID, &workspace.Name, &workspace.Description, &workspace.OwnerID, &workspace.CreatedAt, &workspace.UpdatedAt); err != nil {
		return nil, err
	}

	if r.cache != nil {
		_ = r.cache.Set(ctx, workspace.ID, workspace, 10*time.Minute)
	}

	return workspace, nil
}

func (r *workspaceRepository) FindByOwner(ctx context.Context, ownerID string) ([]*structs.Workspace, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, description, owner_id, created_at, updated_at
		FROM workspaces
		WHERE owner_id = $1
		ORDER BY created_at DESC
	`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workspaces []*structs.Workspace
	for rows.Next() {
		workspace := &structs.Workspace{}
		if err := rows.Scan(&workspace.ID, &workspace.Name, &workspace.Description, &workspace.OwnerID, &workspace.CreatedAt, &workspace.UpdatedAt); err != nil {
			return nil, err
		}
		workspaces = append(workspaces, workspace)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return workspaces, nil
}

func (r *workspaceRepository) Update(ctx context.Context, workspace *structs.Workspace) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE workspaces
		SET name = $1, description = $2, owner_id = $3, updated_at = $4
		WHERE id = $5
	`, workspace.Name, workspace.Description, workspace.OwnerID, workspace.UpdatedAt, workspace.ID)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("workspace not found: %s", workspace.ID)
	}

	if r.cache != nil {
		_ = r.cache.Delete(ctx, workspace.ID)
	}

	r.logger.Debug(ctx, "Workspace updated in Postgres", "workspace_id", workspace.ID)
	return nil
}

func (r *workspaceRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM workspaces WHERE id = $1
	`, id)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("workspace not found: %s", id)
	}

	if r.cache != nil {
		_ = r.cache.Delete(ctx, id)
	}

	r.logger.Debug(ctx, "Workspace deleted from Postgres", "workspace_id", id)
	return nil
}

func (r *workspaceRepository) List(ctx context.Context, userID string, limit, offset int) ([]*structs.Workspace, error) {
	if limit <= 0 {
		limit = 20
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT DISTINCT w.id, w.name, w.description, w.owner_id, w.created_at, w.updated_at
		FROM workspaces w
		LEFT JOIN workspace_members m ON w.id = m.workspace_id
		WHERE w.owner_id = $1 OR m.user_id = $1
		ORDER BY w.created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workspaces []*structs.Workspace
	for rows.Next() {
		workspace := &structs.Workspace{}
		if err := rows.Scan(&workspace.ID, &workspace.Name, &workspace.Description, &workspace.OwnerID, &workspace.CreatedAt, &workspace.UpdatedAt); err != nil {
			return nil, err
		}
		workspaces = append(workspaces, workspace)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return workspaces, nil
}

type PostgresMemberRepository struct {
	db     *sql.DB
	logger *logger.Logger
}

func NewPostgresMemberRepository(db *sql.DB, logger *logger.Logger) (MemberRepository, error) {
	if db == nil {
		return nil, errors.New("database is nil")
	}

	repo := &PostgresMemberRepository{db: db, logger: logger}
	if err := repo.initSchema(context.Background()); err != nil {
		return nil, err
	}

	return repo, nil
}

func (r *PostgresMemberRepository) initSchema(ctx context.Context) error {
	if _, err := r.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS workspace_members (
			id TEXT PRIMARY KEY,
			workspace_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			role TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL
		);
	`); err != nil {
		return err
	}

	if _, err := r.db.ExecContext(ctx, `
		CREATE UNIQUE INDEX IF NOT EXISTS idx_workspace_members_unique ON workspace_members(workspace_id, user_id);
	`); err != nil {
		return err
	}

	if _, err := r.db.ExecContext(ctx, `
		CREATE INDEX IF NOT EXISTS idx_workspace_members_user_id ON workspace_members(user_id);
	`); err != nil {
		return err
	}

	return nil
}

func (r *PostgresMemberRepository) Add(ctx context.Context, member *structs.Member) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO workspace_members (id, workspace_id, user_id, role, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, member.ID, member.WorkspaceID, member.UserID, member.Role, member.CreatedAt)
	if err != nil {
		return err
	}

	r.logger.Debug(ctx, "Member added in Postgres", "workspace_id", member.WorkspaceID, "user_id", member.UserID)
	return nil
}

func (r *PostgresMemberRepository) Remove(ctx context.Context, workspaceID, userID string) error {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM workspace_members WHERE workspace_id = $1 AND user_id = $2
	`, workspaceID, userID)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("member not found: %s in workspace %s", userID, workspaceID)
	}

	r.logger.Debug(ctx, "Member removed in Postgres", "workspace_id", workspaceID, "user_id", userID)
	return nil
}

func (r *PostgresMemberRepository) FindByWorkspace(ctx context.Context, workspaceID string) ([]*structs.Member, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, workspace_id, user_id, role, created_at
		FROM workspace_members
		WHERE workspace_id = $1
		ORDER BY created_at DESC
	`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*structs.Member
	for rows.Next() {
		member := &structs.Member{}
		if err := rows.Scan(&member.ID, &member.WorkspaceID, &member.UserID, &member.Role, &member.CreatedAt); err != nil {
			return nil, err
		}
		members = append(members, member)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return members, nil
}

func (r *PostgresMemberRepository) FindByUser(ctx context.Context, userID string) ([]*structs.Member, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, workspace_id, user_id, role, created_at
		FROM workspace_members
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*structs.Member
	for rows.Next() {
		member := &structs.Member{}
		if err := rows.Scan(&member.ID, &member.WorkspaceID, &member.UserID, &member.Role, &member.CreatedAt); err != nil {
			return nil, err
		}
		members = append(members, member)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return members, nil
}

func (r *PostgresMemberRepository) UpdateRole(ctx context.Context, workspaceID, userID, role string) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE workspace_members SET role = $1 WHERE workspace_id = $2 AND user_id = $3
	`, role, workspaceID, userID)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("member not found: %s in workspace %s", userID, workspaceID)
	}

	r.logger.Debug(ctx, "Member role updated in Postgres", "workspace_id", workspaceID, "user_id", userID, "role", role)
	return nil
}

func (r *PostgresMemberRepository) IsMember(ctx context.Context, workspaceID, userID string) (bool, error) {
	var exists bool
	row := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM workspace_members WHERE workspace_id = $1 AND user_id = $2
		)
	`, workspaceID, userID)
	if err := row.Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (r *PostgresMemberRepository) GetRole(ctx context.Context, workspaceID, userID string) (string, error) {
	var role string
	row := r.db.QueryRowContext(ctx, `
		SELECT role FROM workspace_members WHERE workspace_id = $1 AND user_id = $2
	`, workspaceID, userID)
	if err := row.Scan(&role); err != nil {
		return "", err
	}
	return role, nil
}

type memoryWorkspaceRepository struct {
	workspaces map[string]*structs.Workspace
	mu         sync.RWMutex
	logger     *logger.Logger
}

func NewMemoryWorkspaceRepository(logger *logger.Logger) WorkspaceRepository {
	return &memoryWorkspaceRepository{
		workspaces: make(map[string]*structs.Workspace),
		logger:     logger,
	}
}

func (r *memoryWorkspaceRepository) Create(ctx context.Context, workspace *structs.Workspace) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.workspaces[workspace.ID]; exists {
		return fmt.Errorf("workspace already exists: %s", workspace.ID)
	}

	r.workspaces[workspace.ID] = workspace
	r.logger.Debug(ctx, "Workspace created in memory", "workspace_id", workspace.ID)
	return nil
}

func (r *memoryWorkspaceRepository) FindByID(ctx context.Context, id string) (*structs.Workspace, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	workspace, exists := r.workspaces[id]
	if !exists {
		return nil, fmt.Errorf("workspace not found: %s", id)
	}

	workspaceCopy := *workspace
	return &workspaceCopy, nil
}

func (r *memoryWorkspaceRepository) FindByOwner(ctx context.Context, ownerID string) ([]*structs.Workspace, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var workspaces []*structs.Workspace
	for _, ws := range r.workspaces {
		if ws.OwnerID == ownerID {
			wsCopy := *ws
			workspaces = append(workspaces, &wsCopy)
		}
	}
	return workspaces, nil
}

func (r *memoryWorkspaceRepository) Update(ctx context.Context, workspace *structs.Workspace) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.workspaces[workspace.ID]; !exists {
		return fmt.Errorf("workspace not found: %s", workspace.ID)
	}

	r.workspaces[workspace.ID] = workspace
	r.logger.Debug(ctx, "Workspace updated in memory", "workspace_id", workspace.ID)
	return nil
}

func (r *memoryWorkspaceRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.workspaces[id]; !exists {
		return fmt.Errorf("workspace not found: %s", id)
	}

	delete(r.workspaces, id)
	r.logger.Debug(ctx, "Workspace deleted from memory", "workspace_id", id)
	return nil
}

func (r *memoryWorkspaceRepository) List(ctx context.Context, userID string, limit, offset int) ([]*structs.Workspace, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var workspaces []*structs.Workspace
	for _, ws := range r.workspaces {
		if ws.OwnerID == userID {
			wsCopy := *ws
			workspaces = append(workspaces, &wsCopy)
		}
	}

	if offset >= len(workspaces) {
		return []*structs.Workspace{}, nil
	}

	end := offset + limit
	if end > len(workspaces) {
		end = len(workspaces)
	}

	return workspaces[offset:end], nil
}

type memoryMemberRepository struct {
	members map[string]map[string]*structs.Member
	mu      sync.RWMutex
	logger  *logger.Logger
}

func NewMemoryMemberRepository(logger *logger.Logger) MemberRepository {
	return &memoryMemberRepository{
		members: make(map[string]map[string]*structs.Member),
		logger:  logger,
	}
}

func (r *memoryMemberRepository) Add(ctx context.Context, member *structs.Member) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.members[member.WorkspaceID]; !exists {
		r.members[member.WorkspaceID] = make(map[string]*structs.Member)
	}

	if _, exists := r.members[member.WorkspaceID][member.UserID]; exists {
		return fmt.Errorf("member already exists: %s in workspace %s", member.UserID, member.WorkspaceID)
	}

	r.members[member.WorkspaceID][member.UserID] = member
	r.logger.Debug(ctx, "Member added to memory", "workspace_id", member.WorkspaceID, "user_id", member.UserID)
	return nil
}

func (r *memoryMemberRepository) Remove(ctx context.Context, workspaceID, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	wsMembers, exists := r.members[workspaceID]
	if !exists {
		return fmt.Errorf("workspace not found: %s", workspaceID)
	}

	if _, exists := wsMembers[userID]; !exists {
		return fmt.Errorf("member not found: %s in workspace %s", userID, workspaceID)
	}

	delete(wsMembers, userID)
	r.logger.Debug(ctx, "Member removed from memory", "workspace_id", workspaceID, "user_id", userID)
	return nil
}

func (r *memoryMemberRepository) FindByWorkspace(ctx context.Context, workspaceID string) ([]*structs.Member, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	wsMembers, exists := r.members[workspaceID]
	if !exists {
		return []*structs.Member{}, nil
	}

	var members []*structs.Member
	for _, member := range wsMembers {
		memberCopy := *member
		members = append(members, &memberCopy)
	}
	return members, nil
}

func (r *memoryMemberRepository) FindByUser(ctx context.Context, userID string) ([]*structs.Member, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var members []*structs.Member
	for _, wsMembers := range r.members {
		for _, member := range wsMembers {
			if member.UserID == userID {
				memberCopy := *member
				members = append(members, &memberCopy)
			}
		}
	}
	return members, nil
}

func (r *memoryMemberRepository) UpdateRole(ctx context.Context, workspaceID, userID, role string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	wsMembers, exists := r.members[workspaceID]
	if !exists {
		return fmt.Errorf("workspace not found: %s", workspaceID)
	}

	member, exists := wsMembers[userID]
	if !exists {
		return fmt.Errorf("member not found: %s in workspace %s", userID, workspaceID)
	}

	member.Role = role
	r.logger.Debug(ctx, "Member role updated", "workspace_id", workspaceID, "user_id", userID, "role", role)
	return nil
}

func (r *memoryMemberRepository) IsMember(ctx context.Context, workspaceID, userID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	wsMembers, exists := r.members[workspaceID]
	if !exists {
		return false, nil
	}

	_, exists = wsMembers[userID]
	return exists, nil
}

func (r *memoryMemberRepository) GetRole(ctx context.Context, workspaceID, userID string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	wsMembers, exists := r.members[workspaceID]
	if !exists {
		return "", fmt.Errorf("workspace not found: %s", workspaceID)
	}

	member, exists := wsMembers[userID]
	if !exists {
		return "", fmt.Errorf("member not found: %s in workspace %s", userID, workspaceID)
	}

	return member.Role, nil
}
