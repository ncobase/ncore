// Package service contains auth business logic for the full app.
package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/ncobase/ncore/config"
	authrepo "github.com/ncobase/ncore/examples/08-full-application/core/auth/data/repository"
	authstructs "github.com/ncobase/ncore/examples/08-full-application/core/auth/structs"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/resp"
	securityjwt "github.com/ncobase/ncore/security/jwt"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
)

// TokenPair represents access and refresh token pair.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// Service handles authentication operations.
type Service struct {
	repo         authrepo.UserRepository
	tokenManager *securityjwt.TokenManager
	logger       *logger.Logger
	config       Config
}

// Config holds auth service configuration.
type Config struct {
	JWTSecret             string
	AccessTokenTTL        time.Duration
	RefreshTokenTTL       time.Duration
	PasswordMinLength     int
	RequireUppercase      bool
	RequireNumber         bool
	MaxConcurrentSessions int
}

// NewService creates a new auth service.
func NewService(logger *logger.Logger, authConfig *config.Auth) *Service {
	secret := "your-secret-key-change-in-production-very-long-secret"
	accessTTL := securityjwt.DefaultAccessTokenExpire
	refreshTTL := securityjwt.DefaultRefreshTokenExpire
	if authConfig != nil && authConfig.JWT != nil {
		if authConfig.JWT.Secret != "" {
			secret = authConfig.JWT.Secret
		}
		if authConfig.JWT.Expiry > 0 {
			accessTTL = authConfig.JWT.Expiry
		}
	}

	tokenManager := securityjwt.NewTokenManager(secret, &securityjwt.TokenConfig{
		AccessTokenExpiry:  accessTTL,
		RefreshTokenExpiry: refreshTTL,
	})

	return &Service{
		tokenManager: tokenManager,
		logger:       logger,
		config: Config{
			JWTSecret:             secret,
			AccessTokenTTL:        accessTTL,
			RefreshTokenTTL:       refreshTTL,
			PasswordMinLength:     8,
			RequireUppercase:      true,
			RequireNumber:         true,
			MaxConcurrentSessions: 5,
		},
	}
}

// SetRepository sets the user repository.
func (s *Service) SetRepository(repo authrepo.UserRepository) {
	s.repo = repo
}

// RegisterRequest holds user registration data.
type RegisterRequest struct {
	Name     string `json:"name" binding:"required,min=2"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// LoginRequest holds user login data.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RefreshTokenRequest holds refresh token data.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Register registers a new user.
func (s *Service) Register(ctx context.Context, req *RegisterRequest) (*authstructs.User, error) {
	// Check if user exists
	existing, err := s.repo.FindByEmail(ctx, req.Email)
	if err == nil && existing != nil {
		return nil, ErrUserExists
	}

	// Validate password
	if err := s.validatePassword(req.Password); err != nil {
		return nil, err
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error(ctx, "Failed to hash password", "error", err)
		return nil, fmt.Errorf("failed to hash password")
	}

	// Create user
	user := &authstructs.User{
		ID:        uuid.New().String(),
		Name:      req.Name,
		Email:     req.Email,
		Password:  string(hashedPassword),
		Role:      "user",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		s.logger.Error(ctx, "Failed to create user", "error", err)
		return nil, fmt.Errorf("failed to create user")
	}

	s.logger.Info(ctx, "User registered", "user_id", user.ID, "email", user.Email)
	return user, nil
}

// Login authenticates a user and returns tokens.
func (s *Service) Login(ctx context.Context, req *LoginRequest) (*TokenPair, error) {
	// Find user by email
	user, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Generate tokens
	accessToken, err := s.tokenManager.GenerateAccessToken(user.ID, map[string]any{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
		"name":    user.Name,
	}, &securityjwt.TokenConfig{Expiry: s.config.AccessTokenTTL})
	if err != nil {
		s.logger.Error(ctx, "Failed to generate access token", "error", err)
		return nil, fmt.Errorf("failed to generate access token")
	}

	refreshToken, err := s.tokenManager.GenerateRefreshToken(user.ID, map[string]any{
		"user_id": user.ID,
	}, &securityjwt.TokenConfig{Expiry: s.config.RefreshTokenTTL})
	if err != nil {
		s.logger.Error(ctx, "Failed to generate refresh token", "error", err)
		return nil, fmt.Errorf("failed to generate refresh token")
	}

	s.logger.Info(ctx, "User logged in", "user_id", user.ID, "email", user.Email)

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.config.AccessTokenTTL.Seconds()),
		TokenType:    "Bearer",
	}, nil
}

// RefreshToken refreshes an access token using a refresh token.
func (s *Service) RefreshToken(ctx context.Context, req *RefreshTokenRequest) (*TokenPair, error) {
	token, err := s.tokenManager.ValidateToken(req.RefreshToken)
	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwtv5.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	if !securityjwt.IsRefreshToken(claims) {
		return nil, ErrInvalidToken
	}

	userID := securityjwt.GetPayloadString(claims, "user_id")
	if userID == "" {
		userID = securityjwt.GetTokenID(claims)
	}
	if userID == "" {
		return nil, ErrInvalidToken
	}

	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	accessToken, err := s.tokenManager.GenerateAccessToken(user.ID, map[string]any{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
		"name":    user.Name,
	}, &securityjwt.TokenConfig{Expiry: s.config.AccessTokenTTL})
	if err != nil {
		s.logger.Error(ctx, "Failed to generate access token", "error", err)
		return nil, fmt.Errorf("failed to generate access token")
	}

	newRefreshToken, err := s.tokenManager.GenerateRefreshToken(user.ID, map[string]any{
		"user_id": user.ID,
	}, &securityjwt.TokenConfig{Expiry: s.config.RefreshTokenTTL})
	if err != nil {
		s.logger.Error(ctx, "Failed to generate refresh token", "error", err)
		return nil, fmt.Errorf("failed to generate refresh token")
	}

	s.logger.Info(ctx, "Token refreshed", "user_id", user.ID)

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(s.config.AccessTokenTTL.Seconds()),
		TokenType:    "Bearer",
	}, nil
}

// ValidateToken validates a JWT token and returns claims.
func (s *Service) ValidateToken(tokenString string) (jwtv5.MapClaims, error) {
	token, err := s.tokenManager.ValidateToken(tokenString)
	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwtv5.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return nil, ErrTokenExpired
		}
	}

	return claims, nil
}

// validatePassword validates password strength.
func (s *Service) validatePassword(password string) error {
	if len(password) < s.config.PasswordMinLength {
		return fmt.Errorf("password must be at least %d characters", s.config.PasswordMinLength)
	}

	hasUpper := false
	hasNumber := false

	for _, char := range password {
		if char >= 'A' && char <= 'Z' {
			hasUpper = true
		}
		if char >= '0' && char <= '9' {
			hasNumber = true
		}
	}

	if s.config.RequireUppercase && !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}

	if s.config.RequireNumber && !hasNumber {
		return errors.New("password must contain at least one number")
	}

	return nil
}

// HandleRegister handles user registration HTTP request.
func (s *Service) HandleRegister(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	user, err := s.Register(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, ErrUserExists) {
			resp.Fail(c.Writer, resp.Conflict("user already exists"))
			return
		}
		resp.Fail(c.Writer, resp.InternalServer("failed to register user"))
		return
	}

	user.Password = ""
	resp.WithStatusCode(c.Writer, 201, user)
}

// HandleLogin handles user login HTTP request.
func (s *Service) HandleLogin(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	tokens, err := s.Login(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) || errors.Is(err, ErrUserNotFound) {
			resp.Fail(c.Writer, resp.UnAuthorized("invalid credentials"))
			return
		}
		resp.Fail(c.Writer, resp.InternalServer("login failed"))
		return
	}

	resp.Success(c.Writer, tokens)
}

// HandleRefreshToken handles token refresh HTTP request.
func (s *Service) HandleRefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	tokens, err := s.RefreshToken(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, ErrInvalidToken) || errors.Is(err, ErrTokenExpired) {
			resp.Fail(c.Writer, resp.UnAuthorized("invalid or expired token"))
			return
		}
		resp.Fail(c.Writer, resp.InternalServer("failed to refresh token"))
		return
	}

	resp.Success(c.Writer, tokens)
}

// HandleLogout handles logout HTTP request.
func (s *Service) HandleLogout(c *gin.Context) {
	resp.Success(c.Writer, map[string]string{"message": "logged out successfully"})
}

func (s *Service) TokenManager() *securityjwt.TokenManager {
	return s.tokenManager
}
