// Package server wires the full application server and extensions.
package server

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/config"
	"github.com/ncobase/ncore/data"
	"github.com/ncobase/ncore/extension/manager"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/oss"
	"go.mongodb.org/mongo-driver/mongo"

	_ "github.com/ncobase/ncore/examples/08-full-application/biz/comment"
	"github.com/ncobase/ncore/examples/08-full-application/biz/realtime"
	_ "github.com/ncobase/ncore/examples/08-full-application/biz/task"
	_ "github.com/ncobase/ncore/examples/08-full-application/core/auth"
	_ "github.com/ncobase/ncore/examples/08-full-application/core/user"
	_ "github.com/ncobase/ncore/examples/08-full-application/core/workspace"
	"github.com/ncobase/ncore/examples/08-full-application/internal/event"
	_ "github.com/ncobase/ncore/examples/08-full-application/plugin/export"
	_ "github.com/ncobase/ncore/examples/08-full-application/plugin/notification"
)

type Server struct {
	config   *config.Config
	logger   *logger.Logger
	manager  *manager.Manager
	eventBus *event.Bus
	engine   *gin.Engine
}

func NewServer(cfg *config.Config, log *logger.Logger) (*Server, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if log == nil {
		return nil, fmt.Errorf("logger is nil")
	}

	mgr, err := manager.NewManager(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create extension manager: %w", err)
	}

	dataLayer := mgr.GetData()
	if dataLayer == nil {
		return nil, fmt.Errorf("data layer not initialized")
	}

	dbName := cfg.AppName
	if dbName == "" {
		dbName = "fullappdb"
	}

	collectionAny, err := dataLayer.GetMongoCollection(dbName, "events", false)
	if err != nil {
		return nil, fmt.Errorf("failed to get event collection: %w", err)
	}

	collection, ok := collectionAny.(*mongo.Collection)
	if !ok {
		return nil, fmt.Errorf("event collection type mismatch")
	}

	store, err := event.NewMongoStore(collection, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create event store: %w", err)
	}

	eventBus := event.NewBus(1000, log, store)
	mgr.RegisterCrossService("app.EventBus", eventBus)
	mgr.RegisterCrossService("app.Data", dataLayer)

	// Initialize Search
	searchClient := data.NewSearchClient(dataLayer)
	mgr.RegisterCrossService("app.Search", searchClient)

	// Initialize OSS
	if cfg.Storage != nil {
		ossClient, err := oss.NewStorage(cfg.Storage)
		if err != nil {
			log.Warn(context.Background(), "Failed to initialize OSS", "error", err)
		} else {
			mgr.RegisterCrossService("app.OSS", ossClient)
		}
	}

	if err := mgr.InitExtensions(); err != nil {
		return nil, fmt.Errorf("failed to initialize extensions: %w", err)
	}

	log.Info(context.Background(), "All extensions initialized successfully")

	return &Server{
		config:   cfg,
		logger:   log,
		manager:  mgr,
		eventBus: eventBus,
	}, nil
}

func (s *Server) SetupRouter() *gin.Engine {
	if s.config.Environment != "" {
		gin.SetMode(s.config.Environment)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(s.loggerMiddleware())

	r.GET("/health", func(c *gin.Context) {
		resp.Success(c.Writer, map[string]string{"status": "healthy"})
	})

	r.GET("/events/stats", s.handleEventStats)
	r.GET("/realtime/stats", s.handleRealtimeStats)

	s.manager.RegisterRoutes(r)

	s.engine = r
	return r
}

func (s *Server) StartEventBus(ctx context.Context, workers int) {
	if workers <= 0 {
		workers = 5
	}
	s.eventBus.Start(ctx, workers)
}

func (s *Server) Cleanup(ctx context.Context) {
	s.manager.Cleanup()
	if err := s.eventBus.Shutdown(ctx); err != nil {
		s.logger.Warn(ctx, "Failed to shutdown event bus", "error", err)
	}
}

func (s *Server) loggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()

		s.logger.Info(c.Request.Context(), "HTTP request",
			"method", method,
			"path", path,
			"status", status,
			"duration", duration.String(),
		)
	}
}

func (s *Server) handleEventStats(c *gin.Context) {
	stats := s.eventBus.GetStats()
	resp.Success(c.Writer, stats)
}

func (s *Server) handleRealtimeStats(c *gin.Context) {
	ext, err := s.manager.GetExtensionByName("realtime")
	if err != nil {
		resp.Fail(c.Writer, resp.NotFound("realtime module not found"))
		return
	}

	rt, ok := ext.(*realtime.Module)
	if !ok {
		resp.Fail(c.Writer, resp.InternalServer("realtime module type mismatch"))
		return
	}

	stats := rt.Hub().GetStats()
	resp.Success(c.Writer, stats)
}
