// Package main boots the full application example.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ncobase/ncore/config"
	"github.com/ncobase/ncore/logging/logger"

	"github.com/ncobase/ncore/examples/08-full-application/internal/server"

	_ "github.com/ncobase/ncore/data/meilisearch"
	_ "github.com/ncobase/ncore/data/mongodb"
	_ "github.com/ncobase/ncore/data/postgres"
	_ "github.com/ncobase/ncore/data/redis"
)

func main() {
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	cleanup, err := logger.New(cfg.Logger)
	if err != nil {
		fmt.Printf("Failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer cleanup()
	log := logger.StdLogger()

	srv, err := server.NewServer(cfg, log)
	if err != nil {
		log.Error(context.Background(), "Failed to create server", "error", err)
		os.Exit(1)
	}

	router := srv.SetupRouter()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	srv.StartEventBus(ctx, 5)

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	httpServer := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info(context.Background(), "Starting server", "addr", addr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error(context.Background(), "Server failed", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info(context.Background(), "Shutting down server")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Error(context.Background(), "Server forced to shutdown", "error", err)
	}

	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cleanupCancel()
	srv.Cleanup(cleanupCtx)

	log.Info(context.Background(), "Server exited")
}
