// Package repository stores users for the event-driven example.
package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/ncobase/ncore/examples/06-event-driven/service"
)

type UserRepository interface {
	Create(ctx context.Context, user *service.User) error
	Update(ctx context.Context, user *service.User) error
	FindByID(ctx context.Context, id string) (*service.User, error)
	List(ctx context.Context) ([]*service.User, error)
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) (UserRepository, error) {
	repo := &userRepository{db: db}
	if err := repo.initSchema(context.Background()); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *userRepository) initSchema(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT NOT NULL,
			created_at TEXT NOT NULL
		);
	`)
	return err
}

func (r *userRepository) Create(ctx context.Context, user *service.User) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO users (id, name, email, created_at)
		VALUES (?, ?, ?, ?)
	`,
		user.ID,
		user.Name,
		user.Email,
		user.CreatedAt.UTC().Format(time.RFC3339Nano),
	)
	return err
}

func (r *userRepository) Update(ctx context.Context, user *service.User) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users SET name = ?, email = ? WHERE id = ?
	`, user.Name, user.Email, user.ID)
	return err
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*service.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, name, email, created_at
		FROM users WHERE id = ?
	`, id)

	return scanUser(row)
}

func (r *userRepository) List(ctx context.Context) ([]*service.User, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, email, created_at
		FROM users ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*service.User
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func scanUser(scanner interface{ Scan(dest ...any) error }) (*service.User, error) {
	var createdAt string
	user := &service.User{}
	if err := scanner.Scan(&user.ID, &user.Name, &user.Email, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return nil, err
	}

	parsedCreatedAt, err := time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return nil, err
	}
	user.CreatedAt = parsedCreatedAt
	return user, nil
}
