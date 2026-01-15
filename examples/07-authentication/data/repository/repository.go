// Package repository stores auth users and sessions in SQLite.
package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/ncobase/ncore/examples/07-authentication/structs"
)

type UserRepository interface {
	Create(ctx context.Context, user *structs.User) error
	FindByID(ctx context.Context, id string) (*structs.User, error)
	FindByEmail(ctx context.Context, email string) (*structs.User, error)
	List(ctx context.Context) ([]*structs.User, error)
	Delete(ctx context.Context, id string) error
}

type SessionRepository interface {
	Create(ctx context.Context, session *structs.Session) error
	FindByRefreshToken(ctx context.Context, refreshToken string) (*structs.Session, error)
	DeleteByRefreshToken(ctx context.Context, refreshToken string) error
	DeleteByUserID(ctx context.Context, userID string) error
}

type userRepository struct {
	db *sql.DB
}

type SQLiteSessionRepository struct {
	db *sql.DB
}

func NewSQLiteRepositories(db *sql.DB) (UserRepository, *SQLiteSessionRepository, error) {
	if err := initSchema(context.Background(), db); err != nil {
		return nil, nil, err
	}
	return &userRepository{db: db}, &SQLiteSessionRepository{db: db}, nil
}

func initSchema(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			role TEXT NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);
	`)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			refresh_token TEXT NOT NULL UNIQUE,
			expires_at TEXT NOT NULL,
			created_at TEXT NOT NULL
		);
	`)
	return err
}

func (r *userRepository) Create(ctx context.Context, user *structs.User) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO users (id, name, email, password_hash, role, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`,
		user.ID,
		user.Name,
		user.Email,
		user.PasswordHash,
		string(user.Role),
		user.CreatedAt.UTC().Format(time.RFC3339Nano),
		user.UpdatedAt.UTC().Format(time.RFC3339Nano),
	)
	return err
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*structs.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, name, email, password_hash, role, created_at, updated_at
		FROM users WHERE id = ?
	`, id)

	return scanUser(row)
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*structs.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, name, email, password_hash, role, created_at, updated_at
		FROM users WHERE email = ?
	`, email)

	return scanUser(row)
}

func (r *userRepository) List(ctx context.Context) ([]*structs.User, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, email, password_hash, role, created_at, updated_at
		FROM users ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*structs.User
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

func (r *userRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, id)
	return err
}

func (r *SQLiteSessionRepository) Create(ctx context.Context, session *structs.Session) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO sessions (id, user_id, refresh_token, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?)
	`,
		session.ID,
		session.UserID,
		session.RefreshToken,
		session.ExpiresAt.UTC().Format(time.RFC3339Nano),
		session.CreatedAt.UTC().Format(time.RFC3339Nano),
	)
	return err
}

func (r *SQLiteSessionRepository) FindByRefreshToken(ctx context.Context, refreshToken string) (*structs.Session, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, refresh_token, expires_at, created_at
		FROM sessions WHERE refresh_token = ?
	`, refreshToken)

	return scanSession(row)
}

func (r *SQLiteSessionRepository) DeleteByRefreshToken(ctx context.Context, refreshToken string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE refresh_token = ?`, refreshToken)
	return err
}

func (r *SQLiteSessionRepository) DeleteByUserID(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE user_id = ?`, userID)
	return err
}

func scanUser(scanner interface{ Scan(dest ...any) error }) (*structs.User, error) {
	var role string
	var createdAt string
	var updatedAt string

	user := &structs.User{}
	if err := scanner.Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &role, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return nil, err
	}

	user.Role = structs.Role(role)
	parsedCreatedAt, err := time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return nil, err
	}
	parsedUpdatedAt, err := time.Parse(time.RFC3339Nano, updatedAt)
	if err != nil {
		return nil, err
	}
	user.CreatedAt = parsedCreatedAt
	user.UpdatedAt = parsedUpdatedAt

	return user, nil
}

func scanSession(scanner interface{ Scan(dest ...any) error }) (*structs.Session, error) {
	var expiresAt string
	var createdAt string

	session := &structs.Session{}
	if err := scanner.Scan(&session.ID, &session.UserID, &session.RefreshToken, &expiresAt, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return nil, err
	}

	parsedExpiresAt, err := time.Parse(time.RFC3339Nano, expiresAt)
	if err != nil {
		return nil, err
	}
	parsedCreatedAt, err := time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return nil, err
	}

	session.ExpiresAt = parsedExpiresAt
	session.CreatedAt = parsedCreatedAt

	return session, nil
}
