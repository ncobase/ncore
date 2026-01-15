// Package main boots the realtime WebSocket example.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/config"
	"github.com/ncobase/ncore/examples/04-realtime-websocket/websocket"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/resp"
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

	// Create WebSocket hub
	hub := websocket.NewHub(log)
	go hub.Run()

	// Create WebSocket handler
	wsHandler := websocket.NewHandler(hub, log)

	// Setup router
	if cfg.Environment != "" {
		gin.SetMode(cfg.Environment)
	}

	r := gin.Default()

	// Serve static files
	r.Static("/static", "./web")
	r.LoadHTMLGlob("web/*.html")

	// Routes
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	r.GET("/ws", wsHandler.HandleConnection)
	r.GET("/stats", wsHandler.HandleStats)

	r.GET("/health", func(c *gin.Context) {
		resp.Success(c.Writer, map[string]string{"status": "healthy"})
	})

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	go func() {
		log.Info(context.Background(), "Starting server", "addr", addr)
		if err := r.Run(addr); err != nil {
			log.Error(context.Background(), "Server failed", "error", err)
		}
	}()

	// Wait for interrupt
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info(context.Background(), "Server exited")
}
