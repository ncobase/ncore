// Package main boots the multi-module extension example.
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

	"github.com/ncobase/ncore/config"
	"github.com/ncobase/ncore/examples/03-multi-module/internal/server"
	"github.com/ncobase/ncore/logging/logger"
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

	// Create server
	srv, err := server.NewServer(cfg, log)
	if err != nil {
		log.Error(context.Background(), "Failed to create server", "error", err)
		os.Exit(1)
	}
	defer srv.Cleanup()

	// Setup router
	router := srv.SetupRouter()

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	httpServer := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start HTTP server
	go func() {
		log.Info(context.Background(), "Starting server", "addr", addr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error(context.Background(), "Server failed", "error", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info(context.Background(), "Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Error(ctx, "Server forced to shutdown", "error", err)
	}

	log.Info(context.Background(), "Server exited")
}
