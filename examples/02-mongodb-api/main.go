// Package main boots the MongoDB API example.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/config"
	"github.com/ncobase/ncore/examples/02-mongodb-api/data"
	"github.com/ncobase/ncore/examples/02-mongodb-api/handler"
	"github.com/ncobase/ncore/examples/02-mongodb-api/service"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/resp"

	_ "github.com/ncobase/ncore/data/mongodb"
)

// App represents the main application.
type App struct {
	config  *config.Config
	logger  *logger.Logger
	data    *data.Data
	handler *handler.Handler
	server  *http.Server
}

// NewApp creates a new application instance with manual dependency injection.
func NewApp() (*App, func(), error) {
	// Load configuration
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Create logger
	cleanup1, err := logger.New(cfg.Logger)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create logger: %w", err)
	}
	log := logger.StdLogger()

	// Create data layer
	mongoURI := cfg.Data.Database.Master.Source
	dbName := "userdb" // Extract from config or use default
	dataLayer, err := data.New(mongoURI, dbName, log)
	if err != nil {
		cleanup1()
		return nil, nil, fmt.Errorf("failed to create data layer: %w", err)
	}

	// Create service layer
	svc := service.NewService(dataLayer, log)

	// Create handler layer
	h := handler.NewHandler(svc, log)

	// Create app
	app := &App{
		config:  cfg,
		logger:  log,
		data:    dataLayer,
		handler: h,
	}

	// Combined cleanup function
	cleanup := func() {
		if err := dataLayer.Close(); err != nil {
			log.Error(context.Background(), "failed to close data layer", "error", err)
		}
		cleanup1()
	}

	return app, cleanup, nil
}

// Run starts the application server.
func (a *App) Run() error {
	// Set Gin mode
	if a.config.Environment != "" {
		if a.config.IsProd() {
			gin.SetMode(gin.ReleaseMode)
		} else {
			gin.SetMode(gin.DebugMode)
		}
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Setup router
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(a.loggerMiddleware())

	// Register routes
	a.handler.RegisterRoutes(router)

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		resp.Success(c.Writer, map[string]string{"status": "healthy"})
	})

	// Configure server
	addr := fmt.Sprintf("%s:%d", a.config.Host, a.config.Port)
	a.server = &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		a.logger.Info(context.Background(), "Starting server", "addr", addr)
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.logger.Error(context.Background(), "Server failed", "error", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	a.logger.Info(context.Background(), "Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		a.logger.Error(ctx, "Server forced to shutdown", "error", err)
		return err
	}

	a.logger.Info(context.Background(), "Server exited")
	return nil
}

// loggerMiddleware creates a Gin middleware for request logging.
func (a *App) loggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()

		a.logger.Info(c.Request.Context(), "HTTP request",
			"method", method,
			"path", path,
			"status", status,
			"duration", duration.String(),
			"ip", c.ClientIP(),
		)
	}
}

func main() {
	// Initialize application with manual DI
	app, cleanup, err := NewApp()
	if err != nil {
		fmt.Printf("Failed to initialize app: %v\n", err)
		os.Exit(1)
	}
	defer cleanup()

	// Run application
	if err := app.Run(); err != nil {
		fmt.Printf("Failed to run app: %v\n", err)
		os.Exit(1)
	}
}
