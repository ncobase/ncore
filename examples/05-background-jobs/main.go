// Package main boots the background jobs example.
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
	"github.com/ncobase/ncore/concurrency/worker"
	"github.com/ncobase/ncore/config"
	"github.com/ncobase/ncore/examples/05-background-jobs/job"
	"github.com/ncobase/ncore/examples/05-background-jobs/job/data"
	jobRepo "github.com/ncobase/ncore/examples/05-background-jobs/job/data/repository"
	"github.com/ncobase/ncore/examples/05-background-jobs/job/handler"
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
	logCleanup, err := logger.New(cfg.Logger)
	if err != nil {
		fmt.Printf("Failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer logCleanup()
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

	jobRepo, err := jobRepo.NewJobRepository(dataLayer.DB())
	if err != nil {
		log.Error(context.Background(), "Failed to initialize job repository", "error", err)
		os.Exit(1)
	}

	// Create job manager
	workerCfg := &worker.Config{
		MaxWorkers: 10,
		QueueSize:  100,
	}
	mgr, jobCleanup, err := job.NewManager(workerCfg, jobRepo, log)
	if err != nil {
		log.Error(context.Background(), "Failed to create job manager", "error", err)
		os.Exit(1)
	}
	defer jobCleanup()

	// Register built-in handlers
	job.RegisterBuiltInHandlers(mgr)

	// Create HTTP handler
	jobHandler := handler.NewJobHandler(mgr, log)

	// Setup router
	if cfg.Environment != "" {
		gin.SetMode(cfg.Environment)
	}

	r := gin.Default()

	// Routes
	r.POST("/jobs", jobHandler.CreateJob)
	r.GET("/jobs", jobHandler.ListJobs)
	r.GET("/jobs/:id", jobHandler.GetJob)
	r.GET("/stats", jobHandler.GetStats)

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

	log.Info(context.Background(), "Server exited")
}
