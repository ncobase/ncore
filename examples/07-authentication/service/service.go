// Package service implements authentication services and JWT flows.
package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ncobase/ncore/examples/07-authentication/data/repository"
	"github.com/ncobase/ncore/examples/07-authentication/structs"
	"github.com/ncobase/ncore/logging/logger"
	securityjwt "github.com/ncobase/ncore/security/jwt"
	"golang.org/x/crypto/bcrypt"
)

type Role = structs.Role

type User = structs.User

type Session = structs.Session

type TokenPair = structs.TokenPair

const (
	RoleAdmin     = structs.RoleAdmin
	RoleUser      = structs.RoleUser
	RoleModerator = structs.RoleModerator
)

type Service struct {
	userRepo     repository.UserRepository
	sessionRepo  repository.SessionRepository
	tokenManager *securityjwt.TokenManager
	accessTTL    time.Duration
	refreshTTL   time.Duration
	logger       *logger.Logger
}

func NewService(userRepo repository.UserRepository, sessionRepo repository.SessionRepository, jwtSecret string, accessTTL, refreshTTL time.Duration, logger *logger.Logger) *Service {
	tokenManager := securityjwt.NewTokenManager(jwtSecret, &securityjwt.TokenConfig{
		AccessTokenExpiry:  accessTTL,
		RefreshTokenExpiry: refreshTTL,
	})

	return &Service{
		userRepo:     userRepo,
		sessionRepo:  sessionRepo,
		tokenManager: tokenManager,
		accessTTL:    accessTTL,
		refreshTTL:   refreshTTL,
		logger:       logger,
	}
}

func (s *Service) Register(ctx context.Context, name, email, password string, role Role) (*User, error) {
	if _, err := s.userRepo.FindByEmail(ctx, email); err == nil {
		return nil, fmt.Errorf("user with email %s already exists", email)
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	if len(password) < 8 {
		return nil, fmt.Errorf("password must be at least 8 characters")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &User{
		ID:           uuid.New().String(),
		Name:         name,
		Email:        email,
		PasswordHash: string(hashedPassword),
		Role:         role,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	s.logger.Info(ctx, "User registered", "user_id", user.ID, "email", email)
	return user, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (*TokenPair, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	tokens, err := s.GenerateTokens(user.ID, user.Role)
	if err != nil {
		return nil, err
	}

	session := &Session{
		ID:           uuid.New().String(),
		UserID:       user.ID,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    time.Now().Add(s.refreshTTL),
		CreatedAt:    time.Now(),
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	s.logger.Info(ctx, "User logged in", "user_id", user.ID, "email", email)
	return tokens, nil
}

func (s *Service) GenerateTokens(userID string, role Role) (*TokenPair, error) {
	payload := map[string]any{
		"user_id": userID,
		"role":    string(role),
	}

	accessToken, err := s.tokenManager.GenerateAccessToken(userID, payload, &securityjwt.TokenConfig{Expiry: s.accessTTL})
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.tokenManager.GenerateRefreshToken(userID, payload, &securityjwt.TokenConfig{Expiry: s.refreshTTL})
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.accessTTL.Seconds()),
	}, nil
}

func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	claims, err := s.tokenManager.DecodeToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	if !securityjwt.IsRefreshToken(claims) {
		return nil, fmt.Errorf("invalid refresh token")
	}

	session, err := s.sessionRepo.FindByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	if time.Now().After(session.ExpiresAt) {
		_ = s.sessionRepo.DeleteByRefreshToken(ctx, refreshToken)
		return nil, fmt.Errorf("refresh token expired")
	}

	user, err := s.userRepo.FindByID(ctx, session.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	tokens, err := s.GenerateTokens(user.ID, user.Role)
	if err != nil {
		return nil, err
	}

	newSession := &Session{
		ID:           uuid.New().String(),
		UserID:       user.ID,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    time.Now().Add(s.refreshTTL),
		CreatedAt:    time.Now(),
	}

	if err := s.sessionRepo.Create(ctx, newSession); err != nil {
		return nil, err
	}

	_ = s.sessionRepo.DeleteByRefreshToken(ctx, refreshToken)

	s.logger.Info(ctx, "Token refreshed", "user_id", user.ID)
	return tokens, nil
}

func (s *Service) GetUserByID(userID string) (*User, error) {
	return s.userRepo.FindByID(context.Background(), userID)
}

func (s *Service) ValidateToken(token string) (map[string]any, error) {
	claims, err := s.tokenManager.DecodeToken(token)
	if err != nil {
		return nil, err
	}

	if !securityjwt.IsAccessToken(claims) {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	if err := s.sessionRepo.DeleteByRefreshToken(ctx, refreshToken); err != nil {
		return fmt.Errorf("session not found")
	}
	return nil
}

func (s *Service) ListUsers(ctx context.Context) ([]*User, error) {
	return s.userRepo.List(ctx)
}

func (s *Service) DeleteUser(ctx context.Context, userID string) error {
	if err := s.sessionRepo.DeleteByUserID(ctx, userID); err != nil {
		return err
	}
	return s.userRepo.Delete(ctx, userID)
}
