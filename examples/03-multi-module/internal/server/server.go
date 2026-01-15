// Package server wires the extension manager and HTTP server.
package server

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/config"
	_ "github.com/ncobase/ncore/examples/03-multi-module/biz/comment"
	_ "github.com/ncobase/ncore/examples/03-multi-module/core/post"
	_ "github.com/ncobase/ncore/examples/03-multi-module/core/user"
	"github.com/ncobase/ncore/extension/manager"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/resp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Server represents the application server.
type Server struct {
	config  *config.Config
	logger  *logger.Logger
	manager *manager.Manager
	db      *mongo.Database
	engine  *gin.Engine
}

// NewServer creates a new server instance.
func NewServer(cfg *config.Config, log *logger.Logger) (*Server, error) {
	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.Data.Database.Master.Source))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	db := client.Database("moduledb")

	// Create extension manager
	mgr, err := manager.NewManager(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create extension manager: %w", err)
	}

	// Register database as cross-service for modules to access
	mgr.RegisterCrossService("app.Database", db)

	// Initialize all registered modules
	if err := mgr.InitExtensions(); err != nil {
		return nil, fmt.Errorf("failed to initialize modules: %w", err)
	}

	log.Info(context.Background(), "All modules initialized successfully")

	return &Server{
		config:  cfg,
		logger:  log,
		manager: mgr,
		db:      db,
	}, nil
}

// SetupRouter sets up the Gin router with all module routes.
func (s *Server) SetupRouter() *gin.Engine {
	if s.config.Environment != "" {
		gin.SetMode(s.config.Environment)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(s.loggerMiddleware())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		resp.Success(c.Writer, map[string]string{"status": "healthy"})
	})

	// API routes
	api := r.Group("/api/v1")

	// Register routes from all modules via manager
	extensions := s.manager.ListExtensions()
	for _, ext := range extensions {
		if ext.Instance != nil {
			ext.Instance.RegisterRoutes(api)
			s.logger.Info(context.Background(), "Registered routes for module", "module", ext.Instance.Name())
		}
	}

	s.engine = r
	return r
}

// loggerMiddleware creates request logging middleware.
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

// Cleanup performs cleanup.
func (s *Server) Cleanup() {
	s.manager.Cleanup()
}
