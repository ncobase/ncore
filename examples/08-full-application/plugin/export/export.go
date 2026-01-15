// Package export provides background export jobs for the full app.
package export

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/concurrency/worker"
	"github.com/ncobase/ncore/config"
	"github.com/ncobase/ncore/data"
	taskrepo "github.com/ncobase/ncore/examples/full-application/biz/task/data/repository"
	"github.com/ncobase/ncore/examples/full-application/internal/event"
	exportrepo "github.com/ncobase/ncore/examples/full-application/plugin/export/data/repository"
	"github.com/ncobase/ncore/examples/full-application/plugin/export/service"
	"github.com/ncobase/ncore/extension/registry"
	"github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
)

type Module struct {
	types.OptionalImpl

	service  *service.Service
	logger   *logger.Logger
	bus      *event.Bus
	pool     *worker.Pool
	jobRepo  exportrepo.JobRepository
	taskRepo taskrepo.TaskRepository
}

func init() {
	registry.RegisterToGroup(New(), "plugin")
}

func New() types.Interface {
	return &Module{}
}

func (m *Module) Name() string {
	return "export"
}

func (m *Module) Version() string {
	return "1.0.0"
}

func (m *Module) Dependencies() []string {
	return []string{"task"}
}

func (m *Module) GetMetadata() types.Metadata {
	return types.Metadata{
		Name:         m.Name(),
		Version:      m.Version(),
		Description:  "Background export plugin",
		Type:         "plugin",
		Group:        "plugin",
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

	dbName := mongoDatabaseName(conf)
	collection, err := dataLayer.GetMongoCollection(dbName, "export_jobs", false)
	if err != nil {
		return err
	}

	jobRepo, err := exportrepo.NewJobRepository(collection)
	if err != nil {
		return err
	}
	m.jobRepo = jobRepo
	em.RegisterCrossService("export.Repository", m.jobRepo)

	repoAny, err := em.GetCrossService("task", "Repository")
	if err != nil {
		return err
	}
	repo, ok := repoAny.(taskrepo.TaskRepository)
	if !ok {
		return fmt.Errorf("task repository type mismatch")
	}
	m.taskRepo = repo

	m.pool = worker.NewPool(worker.DefaultConfig())
	m.pool.Start()

	m.service = service.NewService(m.logger, m.bus, m.pool)
	m.service.SetRepositories(m.jobRepo, m.taskRepo)

	m.logger.Info(context.Background(), "Export plugin initialized", "plugin", m.Name())
	return nil
}

func (m *Module) PostInit() error {
	m.logger.Info(context.Background(), "Export plugin post-initialization", "plugin", m.Name())
	return nil
}

func (m *Module) GetHandlers() types.Handler {
	return m.service
}

func (m *Module) GetServices() types.Service {
	return m.service
}

func (m *Module) RegisterRoutes(r *gin.RouterGroup) {
	exports := r.Group("/workspaces/:workspace_id/exports")
	{
		exports.POST("", m.service.HandleCreate)
		exports.GET("", m.service.HandleList)
	}

	r.GET("/exports/:job_id", m.service.HandleGetByID)
}

func mongoDatabaseName(conf *config.Config) string {
	if conf != nil && conf.AppName != "" {
		return conf.AppName
	}
	return "fullappdb"
}

func (m *Module) Cleanup() error {
	m.logger.Info(context.Background(), "Export plugin cleanup", "plugin", m.Name())
	if m.pool != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		m.pool.Stop(ctx)
	}
	return nil
}

func (m *Module) Service() *service.Service {
	return m.service
}
