package comment

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/config"
	"github.com/ncobase/ncore/examples/03-multi-module/biz/comment/data/repository"
	"github.com/ncobase/ncore/examples/03-multi-module/biz/comment/handler"
	"github.com/ncobase/ncore/examples/03-multi-module/biz/comment/service"
	"github.com/ncobase/ncore/examples/03-multi-module/biz/comment/wrapper"
	"github.com/ncobase/ncore/extension/registry"
	"github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
	"go.mongodb.org/mongo-driver/mongo"
)

func init() {
	registry.RegisterToGroup(New(), "biz")
}

type Module struct {
	types.OptionalImpl

	logger  *logger.Logger
	em      types.ManagerInterface
	db      *mongo.Database
	service *service.CommentService
	handler *handler.CommentHandler
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
	return []string{"user", "post"}
}

func (m *Module) GetMetadata() types.Metadata {
	return types.Metadata{
		Name:         m.Name(),
		Version:      m.Version(),
		Description:  "Comment module",
		Type:         "module",
		Group:        "biz",
		Dependencies: m.Dependencies(),
	}
}

func (m *Module) Init(conf *config.Config, em types.ManagerInterface) error {
	m.em = em

	cleanup, err := logger.New(conf.Logger)
	if err != nil {
		return err
	}
	defer cleanup()
	m.logger = logger.StdLogger()

	if db, err := m.em.GetCrossService("app", "Database"); err == nil {
		if mongoDb, ok := db.(*mongo.Database); ok {
			m.db = mongoDb
		}
	}

	m.logger.Info(context.Background(), "Comment module initialized", "module", m.Name())
	return nil
}

func (m *Module) PostInit() error {
	repo := repository.NewCommentRepository(m.db)
	userWrapper := wrapper.NewUserServiceWrapper(m.em)
	postWrapper := wrapper.NewPostServiceWrapper(m.em)

	m.service = service.NewCommentService(repo, userWrapper, postWrapper, m.logger)
	m.handler = handler.NewCommentHandler(m.service, m.logger)

	m.em.RegisterCrossService(m.Name()+".CommentService", m.service)

	m.logger.Info(context.Background(), "Comment module post-initialized", "module", m.Name())
	return nil
}

func (m *Module) Cleanup() error {
	m.logger.Info(context.Background(), "Comment module cleanup", "module", m.Name())
	return nil
}

func (m *Module) GetHandlers() types.Handler {
	return m.handler
}

func (m *Module) GetServices() types.Service {
	return m.service
}

func (m *Module) RegisterRoutes(r *gin.RouterGroup) {
	posts := r.Group("/posts/:post_id/comments")
	{
		posts.POST("", m.handler.HandleCreate)
		posts.GET("", m.handler.HandleList)
	}

	comments := r.Group("/comments")
	{
		comments.DELETE("/:comment_id", m.handler.HandleDelete)
	}
}
