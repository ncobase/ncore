// Package user defines user module routes and wiring.
package user

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/config"
	"github.com/ncobase/ncore/data"
	"github.com/ncobase/ncore/examples/full-application/core/user/data/repository"
	"github.com/ncobase/ncore/examples/full-application/core/user/handler"
	"github.com/ncobase/ncore/examples/full-application/core/user/service"
	"github.com/ncobase/ncore/extension/registry"
	"github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
)

type Module struct {
	types.OptionalImpl

	service *service.Service
	handler *handler.Handler
	repo    repository.UserRepository
	logger  *logger.Logger
}

func init() {
	registry.RegisterToGroup(New(), "core")
}

func New() types.Interface {
	return &Module{}
}

func (m *Module) Name() string {
	return "user"
}

func (m *Module) Version() string {
	return "1.0.0"
}

func (m *Module) Dependencies() []string {
	return []string{}
}

func (m *Module) GetMetadata() types.Metadata {
	return types.Metadata{
		Name:        m.Name(),
		Version:     m.Version(),
		Description: "User management module",
		Type:        "module",
		Group:       "core",
	}
}

func (m *Module) Init(conf *config.Config, em types.ManagerInterface) error {
	cleanup, err := logger.New(conf.Logger)
	if err != nil {
		return err
	}
	defer cleanup()
	m.logger = logger.StdLogger()

	dataAny, err := em.GetCrossService("app", "Data")
	if err != nil {
		return err
	}
	dataLayer, ok := dataAny.(*data.Data)
	if !ok {
		return fmt.Errorf("app data type mismatch")
	}

	db := dataLayer.GetMasterDB()
	if db == nil {
		return fmt.Errorf("master database not configured")
	}

	repo, err := repository.NewUserRepository(db, m.logger, dataLayer.GetRedis())
	if err != nil {
		return err
	}
	m.repo = repo
	m.service = service.New(m.logger, m.repo)
	m.handler = handler.New(m.service)

	em.RegisterCrossService("user.Repository", m.repo)

	m.logger.Info(context.Background(), "User module initialized", "module", m.Name())
	return nil
}

func (m *Module) PostInit() error {
	m.logger.Info(context.Background(), "User module post-initialized", "module", m.Name())
	return nil
}

func (m *Module) GetHandlers() types.Handler {
	return m.handler
}

func (m *Module) GetServices() types.Service {
	return m.service
}

func (m *Module) RegisterRoutes(r *gin.RouterGroup) {
	users := r.Group("/users")
	{
		users.POST("", m.handler.HandleCreate)
		users.GET("", m.handler.HandleList)
		users.GET("/:user_id", m.handler.HandleGetByID)
		users.PUT("/:user_id", m.handler.HandleUpdate)
		users.DELETE("/:user_id", m.handler.HandleDelete)
	}
}

func (m *Module) Service() *service.Service {
	return m.service
}
