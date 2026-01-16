// Package workspace defines workspace module routing and wiring.
package workspace

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/config"
	"github.com/ncobase/ncore/data"
	"github.com/ncobase/ncore/examples/08-full-application/core/workspace/data/repository"
	"github.com/ncobase/ncore/examples/08-full-application/core/workspace/service"
	"github.com/ncobase/ncore/examples/08-full-application/core/workspace/structs"
	"github.com/ncobase/ncore/extension/registry"
	"github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/redis/go-redis/v9"
)

type Workspace = structs.Workspace
type Member = structs.Member
type CreateWorkspaceRequest = structs.CreateWorkspaceRequest
type UpdateWorkspaceRequest = structs.UpdateWorkspaceRequest
type AddMemberRequest = structs.AddMemberRequest

type WorkspaceRepository = repository.WorkspaceRepository
type MemberRepository = repository.MemberRepository

type Module struct {
	types.OptionalImpl

	service       *service.Service
	logger        *logger.Logger
	workspaceRepo WorkspaceRepository
	memberRepo    MemberRepository
}

func init() {
	registry.RegisterToGroup(New(), "core")
}

func New() types.Interface {
	return &Module{}
}

func (m *Module) Name() string {
	return "workspace"
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
		Description:  "Workspace management module",
		Type:         "module",
		Group:        "core",
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

	workspaceRepo, err := repository.NewWorkspaceRepository(db, m.logger, dataLayer.GetRedis().(*redis.Client))
	if err != nil {
		return err
	}
	memberRepo, err := repository.NewPostgresMemberRepository(db, m.logger)
	if err != nil {
		return err
	}

	m.service = service.NewService(m.logger)
	m.workspaceRepo = workspaceRepo
	m.memberRepo = memberRepo
	m.service.SetRepositories(m.workspaceRepo, m.memberRepo)
	em.RegisterCrossService("workspace.Repository", m.workspaceRepo)
	em.RegisterCrossService("workspace.MemberRepository", m.memberRepo)

	m.logger.Info(context.Background(), "Workspace module initialized", "module", m.Name())
	return nil
}

func (m *Module) PostInit() error {
	m.logger.Info(context.Background(), "Workspace module post-initialized", "module", m.Name())
	return nil
}

func (m *Module) GetHandlers() types.Handler {
	return m.service
}

func (m *Module) GetServices() types.Service {
	return m.service
}

func (m *Module) RegisterRoutes(r *gin.RouterGroup) {
	workspaces := r.Group("/workspaces")
	{
		workspaces.POST("", m.service.HandleCreate)
		workspaces.GET("", m.service.HandleList)
		workspaces.GET("/:workspace_id", m.service.HandleGetByID)
		workspaces.POST("/:workspace_id/members", m.service.HandleAddMember)
		workspaces.GET("/:workspace_id/members", m.service.HandleListMembers)
	}
}

func (m *Module) Service() *service.Service {
	return m.service
}
