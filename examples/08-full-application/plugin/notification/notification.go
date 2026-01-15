// Package notification implements email notifications for the full app.
package notification

import (
	"context"
	"time"

	"github.com/ncobase/ncore/config"
	"github.com/ncobase/ncore/examples/full-application/internal/event"
	"github.com/ncobase/ncore/examples/full-application/plugin/notification/service"
	"github.com/ncobase/ncore/extension/registry"
	"github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/messaging/email"
)

type Module struct {
	types.OptionalImpl

	service *service.Service
	logger  *logger.Logger
	bus     *event.Bus
}

func init() {
	registry.RegisterToGroup(New(), "plugin")
}

func New() types.Interface {
	return &Module{}
}

func (m *Module) Name() string {
	return "notification"
}

func (m *Module) Version() string {
	return "1.0.0"
}

func (m *Module) Dependencies() []string {
	return []string{"task", "comment"}
}

func (m *Module) GetMetadata() types.Metadata {
	return types.Metadata{
		Name:         m.Name(),
		Version:      m.Version(),
		Description:  "Email notification plugin",
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

	if busAny, err := em.GetCrossService("app", "EventBus"); err == nil {
		if bus, ok := busAny.(*event.Bus); ok {
			m.bus = bus
		}
	}

	var sender email.Sender
	if conf.Email != nil && conf.Email.SMTP != nil {
		sender, err = email.NewSender(conf.Email.SMTP)
		if err != nil {
			m.logger.Warn(context.Background(), "Failed to create email sender", "error", err)
		}
	}

	m.service = service.NewService(m.logger, sender)

	if m.bus != nil {
		m.bus.Subscribe(event.EventTypeTaskCreated, m.service.HandleTaskCreated)
		m.bus.Subscribe(event.EventTypeTaskAssigned, m.service.HandleTaskAssigned)
		m.bus.Subscribe(event.EventTypeCommentCreated, m.service.HandleCommentCreated)
	}

	m.logger.Info(context.Background(), "Notification plugin initialized", "plugin", m.Name())
	return nil
}

func (m *Module) PostInit() error {
	m.logger.Info(context.Background(), "Notification plugin post-initialization", "plugin", m.Name())
	return nil
}

func (m *Module) GetHandlers() types.Handler {
	return nil
}

func (m *Module) GetServices() types.Service {
	return m.service
}

func (m *Module) Cleanup() error {
	m.logger.Info(context.Background(), "Notification plugin cleanup", "plugin", m.Name())
	time.Sleep(5 * time.Second)
	return nil
}

func (m *Module) Service() *service.Service {
	return m.service
}
