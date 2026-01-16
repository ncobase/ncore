// Package repository stores users for the full application example.
// Package repository stores users for the full application example.
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/ncobase/ncore/data/cache"
	"github.com/ncobase/ncore/examples/08-full-application/core/user/data/ent"
	entuser "github.com/ncobase/ncore/examples/08-full-application/core/user/data/ent/user"
	"github.com/ncobase/ncore/examples/08-full-application/core/user/structs"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/redis/go-redis/v9"
)

type UserRepository interface {
	Create(ctx context.Context, user *structs.User) error
	FindByID(ctx context.Context, id string) (*structs.User, error)
	FindByEmail(ctx context.Context, email string) (*structs.User, error)
	Update(ctx context.Context, user *structs.User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*structs.User, error)
	Count(ctx context.Context) (int, error)
}

type userRepository struct {
	client *ent.Client
	logger *logger.Logger
	cache  *cache.Cache[structs.User]
}

func NewUserRepository(db *sql.DB, logger *logger.Logger, rc *redis.Client) (UserRepository, error) {
	if db == nil {
		return nil, errors.New("database is nil")
	}

	driver := entsql.OpenDB(dialect.Postgres, db)
	client := ent.NewClient(ent.Driver(driver))
	if err := client.Schema.Create(context.Background()); err != nil {
		return nil, err
	}

	repo := &userRepository{client: client, logger: logger}
	if rc != nil {
		repo.cache = cache.NewCache[structs.User](rc, "users")
	}

	return repo, nil
}

func (r *userRepository) Create(ctx context.Context, user *structs.User) error {
	_, err := r.client.User.Create().
		SetID(user.ID).
		SetName(user.Name).
		SetEmail(user.Email).
		SetRole(user.Role).
		SetCreatedAt(user.CreatedAt).
		SetUpdatedAt(user.UpdatedAt).
		Save(ctx)
	if err != nil {
		return err
	}

	if r.cache != nil {
		_ = r.cache.Set(ctx, user.ID, user, 10*time.Minute)
	}

	r.logger.Debug(ctx, "User created in Postgres", "user_id", user.ID)
	return nil
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*structs.User, error) {
	if r.cache != nil {
		cached, err := r.cache.Get(ctx, id)
		if err == nil && cached != nil {
			return cached, nil
		}
	}

	entUser, err := r.client.User.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	user := toStruct(entUser)
	if r.cache != nil {
		_ = r.cache.Set(ctx, user.ID, user, 10*time.Minute)
	}

	return user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*structs.User, error) {
	entUser, err := r.client.User.Query().Where(entuser.EmailEQ(email)).Only(ctx)
	if err != nil {
		return nil, err
	}

	result := toStruct(entUser)
	if r.cache != nil {
		_ = r.cache.Set(ctx, result.ID, result, 10*time.Minute)
	}

	return result, nil
}

func (r *userRepository) Update(ctx context.Context, user *structs.User) error {
	_, err := r.client.User.UpdateOneID(user.ID).
		SetName(user.Name).
		SetEmail(user.Email).
		SetRole(user.Role).
		SetUpdatedAt(user.UpdatedAt).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("user not found: %s", user.ID)
		}
		return err
	}

	if r.cache != nil {
		_ = r.cache.Delete(ctx, user.ID)
	}

	r.logger.Debug(ctx, "User updated in Postgres", "user_id", user.ID)
	return nil
}

func (r *userRepository) Delete(ctx context.Context, id string) error {
	if err := r.client.User.DeleteOneID(id).Exec(ctx); err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("user not found: %s", id)
		}
		return err
	}

	if r.cache != nil {
		_ = r.cache.Delete(ctx, id)
	}

	r.logger.Debug(ctx, "User deleted from Postgres", "user_id", id)
	return nil
}

func (r *userRepository) List(ctx context.Context, limit, offset int) ([]*structs.User, error) {
	if limit <= 0 {
		limit = 20
	}

	rows, err := r.client.User.Query().
		Order(ent.Desc(entuser.FieldCreatedAt)).
		Limit(limit).
		Offset(offset).
		All(ctx)
	if err != nil {
		return nil, err
	}

	users := make([]*structs.User, 0, len(rows))
	for _, row := range rows {
		users = append(users, toStruct(row))
	}

	return users, nil
}

func (r *userRepository) Count(ctx context.Context) (int, error) {
	return r.client.User.Query().Count(ctx)
}

func toStruct(user *ent.User) *structs.User {
	return &structs.User{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

type MemoryRepository struct {
	users  map[string]*structs.User
	mu     sync.RWMutex
	logger *logger.Logger
}

func NewMemoryRepository(logger *logger.Logger) UserRepository {
	return &MemoryRepository{
		users:  make(map[string]*structs.User),
		logger: logger,
	}
}

func (r *MemoryRepository) Create(ctx context.Context, user *structs.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.ID]; exists {
		return errors.New("user already exists")
	}

	for _, u := range r.users {
		if u.Email == user.Email {
			return errors.New("email already in use")
		}
	}

	r.users[user.ID] = user
	r.logger.Debug(ctx, "User created in memory", "user_id", user.ID)
	return nil
}

func (r *MemoryRepository) FindByID(ctx context.Context, id string) (*structs.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return nil, fmt.Errorf("user not found: %s", id)
	}

	userCopy := *user
	return &userCopy, nil
}

func (r *MemoryRepository) FindByEmail(ctx context.Context, email string) (*structs.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if user.Email == email {
			userCopy := *user
			return &userCopy, nil
		}
	}

	return nil, fmt.Errorf("user not found with email: %s", email)
}

func (r *MemoryRepository) Update(ctx context.Context, user *structs.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.ID]; !exists {
		return fmt.Errorf("user not found: %s", user.ID)
	}

	r.users[user.ID] = user
	r.logger.Debug(ctx, "User updated in memory", "user_id", user.ID)
	return nil
}

func (r *MemoryRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[id]; !exists {
		return fmt.Errorf("user not found: %s", id)
	}

	delete(r.users, id)
	r.logger.Debug(ctx, "User deleted from memory", "user_id", id)
	return nil
}

func (r *MemoryRepository) List(ctx context.Context, limit, offset int) ([]*structs.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users := make([]*structs.User, 0, len(r.users))
	for _, user := range r.users {
		userCopy := *user
		users = append(users, &userCopy)
	}

	if offset >= len(users) {
		return []*structs.User{}, nil
	}

	end := offset + limit
	if end > len(users) {
		end = len(users)
	}

	return users[offset:end], nil
}

func (r *MemoryRepository) Count(ctx context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.users), nil
}
