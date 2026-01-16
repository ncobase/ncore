// Package main boots the event-driven example.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/config"
	"github.com/ncobase/ncore/examples/06-event-driven/data"
	datarepo "github.com/ncobase/ncore/examples/06-event-driven/data/repository"
	"github.com/ncobase/ncore/examples/06-event-driven/event"
	"github.com/ncobase/ncore/examples/06-event-driven/handler"
	"github.com/ncobase/ncore/examples/06-event-driven/service"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/resp"

	_ "github.com/ncobase/ncore/data/sqlite"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Create logger
	cleanup, err := logger.New(cfg.Logger)
	if err != nil {
		fmt.Printf("Failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer cleanup()
	log := logger.StdLogger()

	dataLayer, err := data.New(cfg.Data.Database.Master.Driver, cfg.Data.Database.Master.Source, log)
	if err != nil {
		log.Error(context.Background(), "Failed to connect database", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := dataLayer.Close(); err != nil {
			log.Error(context.Background(), "Failed to close database", "error", err)
		}
	}()

	eventStore, err := event.NewSQLiteStore(dataLayer.DB(), log)
	if err != nil {
		log.Error(context.Background(), "Failed to initialize event store", "error", err)
		os.Exit(1)
	}

	// Create event bus
	eventBus := event.NewBus(1000, log, eventStore)

	// Create context for event bus workers
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start event bus workers
	eventBus.Start(ctx, 10)

	userRepo, err := datarepo.NewUserRepository(dataLayer.DB())
	if err != nil {
		log.Error(context.Background(), "Failed to initialize user repository", "error", err)
		os.Exit(1)
	}

	// Create services
	userService := service.NewUserService(eventBus, userRepo, log)
	notificationService := service.NewNotificationService(log)
	analyticsService := service.NewAnalyticsService(log)
	auditService := service.NewAuditService(log)

	// Subscribe event handlers
	eventBus.Subscribe(event.EventTypeUserRegistered, notificationService.HandleUserRegistered)
	eventBus.Subscribe(event.EventTypeUserUpdated, notificationService.HandleUserUpdated)
	eventBus.Subscribe(event.EventTypeUserRegistered, analyticsService.HandleEvent)
	eventBus.Subscribe(event.EventTypeUserUpdated, analyticsService.HandleEvent)
	eventBus.Subscribe(event.EventTypeUserRegistered, auditService.HandleEvent)
	eventBus.Subscribe(event.EventTypeUserUpdated, auditService.HandleEvent)

	// Create HTTP handler
	h := handler.NewHandler(userService, eventBus, eventStore, analyticsService, log)

	// Setup router
	if cfg.Environment != "" {
		gin.SetMode(cfg.Environment)
	}

	r := gin.Default()

	// User routes
	users := r.Group("/users")
	{
		users.POST("", h.RegisterUser)
		users.GET("", h.ListUsers)
		users.GET("/:user_id", h.GetUser)
		users.PUT("/:user_id", h.UpdateUser)
	}

	// Event routes
	events := r.Group("/events")
	{
		events.POST("", h.PublishEvent)
		events.GET("", h.GetEvents)
	}

	r.GET("/stats", h.GetStats)
	r.GET("/health", func(c *gin.Context) {
		resp.Success(c.Writer, map[string]string{"status": "healthy"})
	})

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		log.Info(context.Background(), "Starting server", "addr", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error(context.Background(), "Server failed", "error", err)
		}
	}()

	// Wait for interrupt
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info(context.Background(), "Shutting down server...")

	// Cancel event bus context
	cancel()

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error(context.Background(), "Server forced to shutdown", "error", err)
	}

	log.Info(context.Background(), "Server exited")
}
