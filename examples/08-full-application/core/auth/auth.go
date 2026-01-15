// Package auth provides authentication and authorization modules.
package auth

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/config"
	"github.com/ncobase/ncore/examples/full-application/core/auth/data/repository"
	"github.com/ncobase/ncore/examples/full-application/core/auth/middleware"
	"github.com/ncobase/ncore/examples/full-application/core/auth/service"
	"github.com/ncobase/ncore/extension/registry"
	"github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
)

func init() {
	registry.RegisterToGroup(New(), "core")
}

// Module implements the extension interface for authentication.
type Module struct {
	types.OptionalImpl

	service    *service.Service
	middleware *middleware.Middleware
	logger     *logger.Logger
}

// New creates a new auth module.
func New() types.Interface {
	return &Module{}
}

func (m *Module) Name() string {
	return "auth"
}

func (m *Module) Version() string {
	return "1.0.0"
}

func (m *Module) Dependencies() []string {
	return []string{"user"}
}

func (m *Module) GetMetadata() types.Metadata {
	return types.Metadata{
		Name:         m.Name(),
		Version:      m.Version(),
		Description:  "Authentication module with JWT",
		Type:         "module",
		Group:        "core",
		Dependencies: m.Dependencies(),
	}
}

// Init initializes the auth module.
func (m *Module) Init(conf *config.Config, em types.ManagerInterface) error {
	cleanup, err := logger.New(conf.Logger)
	if err != nil {
		return err
	}
	defer cleanup()
	m.logger = logger.StdLogger()

	m.service = service.NewService(m.logger, conf.Auth)
	if repoAny, err := em.GetCrossService("user", "Repository"); err == nil {
		if repo, ok := repoAny.(repository.UserRepository); ok {
			m.service.SetRepository(repo)
		}
	} else {
		m.logger.Warn(context.Background(), "User repository not available for auth module", "error", err)
	}
	m.middleware = middleware.NewMiddleware(m.service.TokenManager(), m.logger)

	m.logger.Info(context.Background(), "Auth module initialized", "module", m.Name())
	return nil
}

// PostInit performs post-initialization.
func (m *Module) PostInit() error {
	m.logger.Info(context.Background(), "Auth module post-initialized", "module", m.Name())
	return nil
}

// GetHandlers returns the module's HTTP handlers.
func (m *Module) GetHandlers() types.Handler {
	return m.service
}

// GetServices returns the module's services.
func (m *Module) GetServices() types.Service {
	return m.service
}

// RegisterRoutes registers HTTP routes for the module.
func (m *Module) RegisterRoutes(r *gin.RouterGroup) {
	auth := r.Group("/auth")
	{
		auth.POST("/register", m.service.HandleRegister)
		auth.POST("/login", m.service.HandleLogin)
		auth.POST("/refresh", m.service.HandleRefreshToken)
		auth.POST("/logout", m.service.HandleLogout)
	}
}

// Service returns the auth service.
func (m *Module) Service() *service.Service {
	return m.service
}

// Middleware returns the auth middleware.
func (m *Module) Middleware() *middleware.Middleware {
	return m.middleware
}
