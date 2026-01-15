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
	"github.com/ncobase/ncore/examples/01-basic-rest-api/handler"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/resp"
)

var (
	// ErrInvalidConfig is returned when configuration is invalid.
	ErrInvalidConfig = errors.New("invalid configuration")
)

// App represents the main application.
type App struct {
	config  *config.Config
	logger  *logger.Logger
	handler *handler.Handler
	server  *http.Server
}

// NewApp creates a new application instance.
func NewApp(
	cfg *config.Config,
	logger *logger.Logger,
	h *handler.Handler,
) *App {
	// Set Gin mode
	if cfg.Environment != "" {
		gin.SetMode(cfg.Environment)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	return &App{
		config:  cfg,
		logger:  logger,
		handler: h,
	}
}

// Run starts the application server.
func (a *App) Run() error {
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
		a.logger.Error(context.Background(), "Server forced to shutdown", "error", err)
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

		a.logger.Info(context.Background(), "HTTP request",
			"method", method,
			"path", path,
			"status", status,
			"duration", duration.String(),
			"ip", c.ClientIP(),
		)
	}
}
