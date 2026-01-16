// Package comment wires comment module routes and services.
package comment

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/config"
	"github.com/ncobase/ncore/data"
	"github.com/ncobase/ncore/examples/08-full-application/biz/comment/data/repository"
	"github.com/ncobase/ncore/examples/08-full-application/biz/comment/service"
	"github.com/ncobase/ncore/examples/08-full-application/biz/comment/structs"
	"github.com/ncobase/ncore/examples/08-full-application/internal/event"
	"github.com/ncobase/ncore/extension/registry"
	"github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
)

type Comment = structs.Comment
type CreateCommentRequest = structs.CreateCommentRequest
type UpdateCommentRequest = structs.UpdateCommentRequest

type CommentRepository = repository.CommentRepository

type Module struct {
	types.OptionalImpl

	service *service.Service
	logger  *logger.Logger
	repo    CommentRepository
	bus     *event.Bus
}

func init() {
	registry.RegisterToGroup(New(), "biz")
}

func New() types.Interface {
	return &Module{}
}

func (m *Module) Name() string {
	return "comment"
}

func (m *Module) Version() string {
	return "1.0.0"
}

func (m *Module) Dependencies() []string {
	return []string{"task", "user", "workspace"}
}

func (m *Module) GetMetadata() types.Metadata {
	return types.Metadata{
		Name:         m.Name(),
		Version:      m.Version(),
		Description:  "Comment module with task association",
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

	repo, err := repository.NewCommentRepository(db)
	if err != nil {
		return err
	}

	m.service = service.NewService(m.logger, m.bus)
	m.repo = repo
	m.service.SetRepository(m.repo)
	em.RegisterCrossService("comment.Repository", m.repo)

	m.logger.Info(context.Background(), "Comment module initialized", "module", m.Name())
	return nil
}

func (m *Module) PostInit() error {
	m.logger.Info(context.Background(), "Comment module post-initialized", "module", m.Name())
	return nil
}

func (m *Module) GetHandlers() types.Handler {
	return m.service
}

func (m *Module) GetServices() types.Service {
	return m.service
}

func (m *Module) RegisterRoutes(r *gin.RouterGroup) {
	workspaces := r.Group("/workspaces/:workspace_id/comments")
	{
		workspaces.POST("", m.service.HandleCreate)
	}

	tasks := r.Group("/tasks/:task_id/comments")
	{
		tasks.GET("", m.service.HandleList)
	}

	comments := r.Group("/comments")
	{
		comments.GET("/:comment_id", m.service.HandleGetByID)
		comments.PUT("/:comment_id", m.service.HandleUpdate)
		comments.DELETE("/:comment_id", m.service.HandleDelete)
	}
}

func (m *Module) Service() *service.Service {
	return m.service
}
