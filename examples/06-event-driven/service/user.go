// Package service contains application services for event-driven flows.
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/ncobase/ncore/examples/06-event-driven/event"
	"github.com/ncobase/ncore/logging/logger"
)

type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id string) (*User, error)
	List(ctx context.Context) ([]*User, error)
}

type UserService struct {
	eventBus *event.Bus
	logger   *logger.Logger
	repo     UserRepository
}

func NewUserService(eventBus *event.Bus, repo UserRepository, logger *logger.Logger) *UserService {
	return &UserService{
		eventBus: eventBus,
		logger:   logger,
		repo:     repo,
	}
}

func (s *UserService) RegisterUser(ctx context.Context, name, email string) (*User, error) {
	user := &User{
		ID:        fmt.Sprintf("user-%d", time.Now().UnixNano()),
		Name:      name,
		Email:     email,
		CreatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	evt := &event.Event{
		Type:          event.EventTypeUserRegistered,
		AggregateID:   user.ID,
		AggregateName: "user",
		Payload: map[string]any{
			"user_id": user.ID,
			"name":    user.Name,
			"email":   user.Email,
		},
		Version: 1,
	}

	if err := s.eventBus.Publish(ctx, evt); err != nil {
		s.logger.Error(ctx, "Failed to publish event", "error", err)
	}

	s.logger.Info(ctx, "User registered", "user_id", user.ID)
	return user, nil
}

func (s *UserService) UpdateUser(ctx context.Context, userID, name, email string) (*User, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	oldName := user.Name
	oldEmail := user.Email

	user.Name = name
	user.Email = email

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	evt := &event.Event{
		Type:          event.EventTypeUserUpdated,
		AggregateID:   user.ID,
		AggregateName: "user",
		Payload: map[string]any{
			"user_id":   user.ID,
			"old_name":  oldName,
			"new_name":  name,
			"old_email": oldEmail,
			"new_email": email,
		},
		Version: 1,
	}

	if err := s.eventBus.Publish(ctx, evt); err != nil {
		s.logger.Error(ctx, "Failed to publish event", "error", err)
	}

	s.logger.Info(ctx, "User updated", "user_id", user.ID)
	return user, nil
}

func (s *UserService) GetUser(userID string) (*User, error) {
	user, err := s.repo.FindByID(context.Background(), userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

func (s *UserService) ListUsers() []*User {
	users, err := s.repo.List(context.Background())
	if err != nil {
		s.logger.Error(context.Background(), "Failed to list users", "error", err)
		return []*User{}
	}
	return users
}
