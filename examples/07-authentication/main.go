// Package main boots the authentication example.
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
	"github.com/ncobase/ncore/examples/07-authentication/data"
	"github.com/ncobase/ncore/examples/07-authentication/data/repository"
	"github.com/ncobase/ncore/examples/07-authentication/handler"
	"github.com/ncobase/ncore/examples/07-authentication/middleware"
	auth "github.com/ncobase/ncore/examples/07-authentication/service"
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

	userRepo, sessionRepo, err := repository.NewSQLiteRepositories(dataLayer.DB())
	if err != nil {
		log.Error(context.Background(), "Failed to initialize auth repository", "error", err)
		os.Exit(1)
	}

	// Create auth service
	accessTTL := time.Duration(900) * time.Second
	refreshTTL := time.Duration(604800) * time.Second
	authService := auth.NewService(userRepo, sessionRepo, cfg.Auth.JWT.Secret, accessTTL, refreshTTL, log)

	// Create handlers
	authHandler := handler.NewAuthHandler(authService, log)
	adminHandler := handler.NewAdminHandler(authService, log)

	// Setup router
	if cfg.Environment != "" {
		gin.SetMode(cfg.Environment)
	}

	r := gin.Default()

	// Public routes
	authRoutes := r.Group("/auth")
	{
		authRoutes.POST("/register", authHandler.Register)
		authRoutes.POST("/login", authHandler.Login)
		authRoutes.POST("/refresh", authHandler.RefreshToken)
		authRoutes.POST("/logout", authHandler.Logout)
	}

	// Protected routes (require authentication)
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware(authService, log))
	{
		api.GET("/profile", authHandler.GetProfile)

		// User-level protected endpoint
		api.GET("/user-data", func(c *gin.Context) {
			userID, _ := middleware.GetCurrentUserID(c)
			resp.Success(c.Writer, map[string]any{
				"message": "This is user-level protected data",
				"user_id": userID,
			})
		})
	}

	// Admin-only routes
	admin := r.Group("/api/admin")
	admin.Use(middleware.AuthMiddleware(authService, log))
	admin.Use(middleware.RequireRole(authService, auth.RoleAdmin))
	{
		admin.GET("/users", adminHandler.ListUsers)
		admin.DELETE("/users/:user_id", adminHandler.DeleteUser)
		admin.GET("/stats", func(c *gin.Context) {
			resp.Success(c.Writer, map[string]string{"message": "Admin-only statistics"})
		})
	}

	// Moderator routes (admin or moderator)
	moderator := r.Group("/api/moderator")
	moderator.Use(middleware.AuthMiddleware(authService, log))
	moderator.Use(middleware.RequireRole(authService, auth.RoleAdmin, auth.RoleModerator))
	{
		moderator.GET("/pending", func(c *gin.Context) {
			resp.Success(c.Writer, map[string]string{"message": "Moderator-only pending items"})
		})
	}

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
