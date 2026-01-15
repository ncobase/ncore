// Package post defines the post module for the multi-module example.
package post

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/config"
	"github.com/ncobase/ncore/examples/03-multi-module/core/post/data/repository"
	"github.com/ncobase/ncore/examples/03-multi-module/core/post/handler"
	"github.com/ncobase/ncore/examples/03-multi-module/core/post/service"
	"github.com/ncobase/ncore/examples/03-multi-module/core/post/wrapper"
	"github.com/ncobase/ncore/extension/registry"
	"github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
	"go.mongodb.org/mongo-driver/mongo"
)

func init() {
	registry.RegisterToGroupWithWeakDeps(New(), "core", []string{"user"})
}

// Module represents the post module.
type Module struct {
	types.OptionalImpl

	logger  *logger.Logger
	em      types.ManagerInterface
	db      *mongo.Database
	service *service.PostService
	handler *handler.PostHandler
}

// New creates a new post module instance.
func New() types.Interface {
	return &Module{}
}

// Name returns the module name.
func (m *Module) Name() string {
	return "post"
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
		Description: "Post management module",
		Type:        "module",
		Group:       "core",
	}
}

// Init initializes the module.
func (m *Module) Init(conf *config.Config, em types.ManagerInterface) error {
	m.em = em

	cleanup, err := logger.New(conf.Logger)
	if err != nil {
		return err
	}
	defer cleanup()
	m.logger = logger.StdLogger()

	// Get database from cross-service
	if db, err := m.em.GetCrossService("app", "Database"); err == nil {
		if mongoDb, ok := db.(*mongo.Database); ok {
			m.db = mongoDb
		}
	}

	m.logger.Info(context.Background(), "Post module initialized", "module", m.Name())
	return nil
}

// PostInit performs post-initialization.
func (m *Module) PostInit() error {
	// Create user service wrapper for cross-module calls
	userWrapper := wrapper.NewUserServiceWrapper(m.em)

	// Create service with user wrapper
	repo := repository.NewPostRepository(m.db)
	m.service = service.NewPostService(repo, userWrapper, m.logger)

	// Create handler
	m.handler = handler.NewPostHandler(m.service, m.logger)

	// Register service for cross-module access
	m.em.RegisterCrossService(m.Name()+".PostService", m.service)

	m.logger.Info(context.Background(), "Post module post-initialized", "module", m.Name())
	return nil
}

// Cleanup performs cleanup.
func (m *Module) Cleanup() error {
	m.logger.Info(context.Background(), "Post module cleanup", "module", m.Name())
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
	posts := r.Group("/posts")
	{
		posts.POST("", m.handler.Create)
		posts.GET("", m.handler.List)
		posts.GET("/:id", m.handler.Get)
		posts.PUT("/:id", m.handler.Update)
		posts.DELETE("/:id", m.handler.Delete)
	}
}
