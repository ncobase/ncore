// Package repository stores comments for the full application example.
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"

	"github.com/ncobase/ncore/examples/08-full-application/biz/comment/structs"
)

type CommentRepository interface {
	Create(ctx context.Context, comment *structs.Comment) error
	FindByID(ctx context.Context, id string) (*structs.Comment, error)
	FindByTask(ctx context.Context, taskID string, limit, offset int) ([]*structs.Comment, error)
	Update(ctx context.Context, comment *structs.Comment) error
	Delete(ctx context.Context, id string) error
}

type commentRepository struct {
	db *sql.DB
}

func NewCommentRepository(db *sql.DB) (CommentRepository, error) {
	if db == nil {
		return nil, errors.New("database is nil")
	}

	repo := &commentRepository{db: db}
	if err := repo.initSchema(context.Background()); err != nil {
		return nil, err
	}

	return repo, nil
}

func (r *commentRepository) initSchema(ctx context.Context) error {
	if _, err := r.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS comments (
			id TEXT PRIMARY KEY,
			workspace_id TEXT NOT NULL,
			task_id TEXT NOT NULL,
			content TEXT NOT NULL,
			created_by TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL
		);
	`); err != nil {
		return err
	}

	if _, err := r.db.ExecContext(ctx, `
		CREATE INDEX IF NOT EXISTS idx_comments_task_id ON comments(task_id);
	`); err != nil {
		return err
	}

	return nil
}

func (r *commentRepository) Create(ctx context.Context, comment *structs.Comment) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO comments (id, workspace_id, task_id, content, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, comment.ID, comment.WorkspaceID, comment.TaskID, comment.Content, comment.CreatedBy, comment.CreatedAt, comment.UpdatedAt)
	return err
}

func (r *commentRepository) FindByID(ctx context.Context, id string) (*structs.Comment, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, workspace_id, task_id, content, created_by, created_at, updated_at
		FROM comments WHERE id = $1
	`, id)
	comment := &structs.Comment{}
	if err := row.Scan(&comment.ID, &comment.WorkspaceID, &comment.TaskID, &comment.Content, &comment.CreatedBy, &comment.CreatedAt, &comment.UpdatedAt); err != nil {
		return nil, err
	}
	return comment, nil
}

func (r *commentRepository) FindByTask(ctx context.Context, taskID string, limit, offset int) ([]*structs.Comment, error) {
	if limit <= 0 {
		limit = 20
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, workspace_id, task_id, content, created_by, created_at, updated_at
		FROM comments
		WHERE task_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, taskID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*structs.Comment
	for rows.Next() {
		comment := &structs.Comment{}
		if err := rows.Scan(&comment.ID, &comment.WorkspaceID, &comment.TaskID, &comment.Content, &comment.CreatedBy, &comment.CreatedAt, &comment.UpdatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

func (r *commentRepository) Update(ctx context.Context, comment *structs.Comment) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE comments SET content = $1, updated_at = $2 WHERE id = $3
	`, comment.Content, comment.UpdatedAt, comment.ID)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("comment not found: %s", comment.ID)
	}

	return nil
}

func (r *commentRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM comments WHERE id = $1
	`, id)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("comment not found: %s", id)
	}

	return nil
}

type MemoryCommentRepository struct {
	comments map[string]*structs.Comment
	mu       sync.RWMutex
}

func NewMemoryCommentRepository() CommentRepository {
	return &MemoryCommentRepository{
		comments: make(map[string]*structs.Comment),
	}
}

func (r *MemoryCommentRepository) Create(ctx context.Context, comment *structs.Comment) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.comments[comment.ID]; exists {
		return fmt.Errorf("comment already exists: %s", comment.ID)
	}

	r.comments[comment.ID] = comment
	return nil
}

func (r *MemoryCommentRepository) FindByID(ctx context.Context, id string) (*structs.Comment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	comment, exists := r.comments[id]
	if !exists {
		return nil, fmt.Errorf("comment not found: %s", id)
	}

	commentCopy := *comment
	return &commentCopy, nil
}

func (r *MemoryCommentRepository) FindByTask(ctx context.Context, taskID string, limit, offset int) ([]*structs.Comment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var comments []*structs.Comment
	for _, comment := range r.comments {
		if comment.TaskID == taskID {
			commentCopy := *comment
			comments = append(comments, &commentCopy)
		}
	}

	if offset >= len(comments) {
		return []*structs.Comment{}, nil
	}

	end := offset + limit
	if end > len(comments) {
		end = len(comments)
	}

	return comments[offset:end], nil
}

func (r *MemoryCommentRepository) Update(ctx context.Context, comment *structs.Comment) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.comments[comment.ID]; !exists {
		return fmt.Errorf("comment not found: %s", comment.ID)
	}

	r.comments[comment.ID] = comment
	return nil
}

func (r *MemoryCommentRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.comments[id]; !exists {
		return fmt.Errorf("comment not found: %s", id)
	}

	delete(r.comments, id)
	return nil
}
