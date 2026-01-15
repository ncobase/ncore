// Package task defines task module wiring and routes.
package task

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/config"
	"github.com/ncobase/ncore/data"
	"github.com/ncobase/ncore/examples/full-application/biz/task/data/repository"
	"github.com/ncobase/ncore/examples/full-application/biz/task/service"
	"github.com/ncobase/ncore/examples/full-application/biz/task/structs"
	"github.com/ncobase/ncore/examples/full-application/internal/event"
	"github.com/ncobase/ncore/extension/registry"
	"github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
)

type Task = structs.Task
type CreateTaskRequest = structs.CreateTaskRequest
type UpdateTaskRequest = structs.UpdateTaskRequest

type TaskRepository = repository.TaskRepository

type Module struct {
	types.OptionalImpl

	service *service.Service
	logger  *logger.Logger
	repo    TaskRepository
	bus     *event.Bus
}

func init() {
	registry.RegisterToGroup(New(), "biz")
}

func New() types.Interface {
	return &Module{}
}

func (m *Module) Name() string {
	return "task"
}

func (m *Module) Version() string {
	return "1.0.0"
}

func (m *Module) Dependencies() []string {
	return []string{"workspace", "user"}
}

func (m *Module) GetMetadata() types.Metadata {
	return types.Metadata{
		Name:         m.Name(),
		Version:      m.Version(),
		Description:  "Task management module",
		Type:         "module",
		Group:        "biz",
		Dependencies: m.Dependencies(),
	}
}

func (m *Module) Init(conf *config.Config, em types.ManagerInterface) error {
	cleanup, err := logger.New(conf.Logger)
	if err != nil {
		return err
	}
	defer cleanup()
	m.logger = logger.StdLogger()

	busAny, err := em.GetCrossService("app", "EventBus")
	if err != nil {
		return err
	}
	bus, ok := busAny.(*event.Bus)
	if !ok {
		return fmt.Errorf("event bus type mismatch")
	}
	m.bus = bus

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

	repo, err := repository.NewTaskRepository(db)
	if err != nil {
		return err
	}

	m.service = service.NewService(m.logger, m.bus)
	m.repo = repo
	m.service.SetRepository(m.repo)
	em.RegisterCrossService("task.Repository", m.repo)

	m.logger.Info(context.Background(), "Task module initialized", "module", m.Name())
	return nil
}

func (m *Module) PostInit() error {
	m.logger.Info(context.Background(), "Task module post-initialized", "module", m.Name())
	return nil
}

func (m *Module) GetHandlers() types.Handler {
	return m.service
}

func (m *Module) GetServices() types.Service {
	return m.service
}

func (m *Module) RegisterRoutes(r *gin.RouterGroup) {
	workspaces := r.Group("/workspaces/:workspace_id/tasks")
	{
		workspaces.POST("", m.service.HandleCreate)
		workspaces.GET("", m.service.HandleList)
	}

	tasks := r.Group("/tasks")
	{
		tasks.GET("/:task_id", m.service.HandleGetByID)
		tasks.PUT("/:task_id", m.service.HandleUpdate)
		tasks.DELETE("/:task_id", m.service.HandleDelete)
		tasks.POST("/:task_id/assign", m.service.HandleAssign)
	}
}

func (m *Module) Service() *service.Service {
	return m.service
}
