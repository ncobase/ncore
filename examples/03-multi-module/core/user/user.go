// Package user defines the user module for the multi-module example.
package user

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/config"
	"github.com/ncobase/ncore/examples/03-multi-module/core/user/data/repository"
	"github.com/ncobase/ncore/examples/03-multi-module/core/user/handler"
	"github.com/ncobase/ncore/examples/03-multi-module/core/user/service"
	"github.com/ncobase/ncore/extension/registry"
	"github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
	"go.mongodb.org/mongo-driver/mongo"
)

func init() {
	registry.RegisterToGroup(New(), "core")
}

// Module represents the user module.
type Module struct {
	types.OptionalImpl

	logger  *logger.Logger
	em      types.ManagerInterface
	db      *mongo.Database
	service *service.UserService
	handler *handler.UserHandler
}

// New creates a new user module instance.
func New() types.Interface {
	return &Module{}
}

// Name returns the module name.
func (m *Module) Name() string {
	return "user"
}

// Version returns the module version.
func (m *Module) Version() string {
	return "1.0.0"
}

// Dependencies returns the module's hard dependencies.
func (m *Module) Dependencies() []string {
	return []string{}
}

// GetMetadata returns module metadata.
func (m *Module) GetMetadata() types.Metadata {
	return types.Metadata{
		Name:        m.Name(),
		Version:     m.Version(),
		Description: "User management module",
		Type:        "module",
		Group:       "core",
	}
}

// Init initializes the module with configuration.
func (m *Module) Init(conf *config.Config, em types.ManagerInterface) error {
	m.em = em

	cleanup, err := logger.New(conf.Logger)
	if err != nil {
		return err
	}
	defer cleanup()
	m.logger = logger.StdLogger()

	// Get database from cross-service (assuming main app registered it)
	if db, err := m.em.GetCrossService("app", "Database"); err == nil {
		if mongoDb, ok := db.(*mongo.Database); ok {
			m.db = mongoDb
		}
	}

	m.logger.Info(context.Background(), "User module initialized", "module", m.Name())
	return nil
}

// PostInit performs post-initialization setup.
func (m *Module) PostInit() error {
	// Create service
	repo := repository.NewUserRepository(m.db)
	m.service = service.NewUserService(repo, m.logger)

	// Create handler
	m.handler = handler.NewUserHandler(m.service, m.logger)

	// Register service for cross-module access
	m.em.RegisterCrossService(m.Name()+".UserService", m.service)

	m.logger.Info(context.Background(), "User module post-initialized", "module", m.Name())
	return nil
}

// Cleanup performs cleanup when module is unloaded.
func (m *Module) Cleanup() error {
	m.logger.Info(context.Background(), "User module cleanup", "module", m.Name())
	return nil
}

// GetHandlers returns the module's HTTP handlers.
func (m *Module) GetHandlers() types.Handler {
	return m.handler
}

// GetServices returns the module's services.
func (m *Module) GetServices() types.Service {
	return m.service
}

// RegisterRoutes registers HTTP routes.
func (m *Module) RegisterRoutes(r *gin.RouterGroup) {
	users := r.Group("/users")
	{
		users.POST("", m.handler.Create)
		users.GET("", m.handler.List)
		users.GET("/:user_id", m.handler.Get)
		users.PUT("/:user_id", m.handler.Update)
		users.DELETE("/:user_id", m.handler.Delete)
	}
}
