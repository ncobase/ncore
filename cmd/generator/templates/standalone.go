package templates

import "fmt"

// StandaloneMainTemplate generates the main.go file for standalone applications
func StandaloneMainTemplate(name, moduleName string) string {
	return fmt.Sprintf(`package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"{{ .PackagePath }}/cmd/provider"
	"github.com/ncobase/ncore/config"
	"github.com/ncobase/ncore/helper"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/version"
)

const (
	shutdownTimeout = 3 * time.Second // service shutdown timeout
)

func main() {
	logger.SetVersion(version.GetVersionInfo().Version)
	// load config
	conf := loadConfig()

	// Application name
	appName := conf.AppName

	// Initialize logger
	cleanupLogger := initializeLogger(conf)
	defer cleanupLogger()

	logger.Infof(context.Background(), "Starting %%s", appName)

	if err := runServer(conf); err != nil {
		logger.Fatalf(context.Background(), "Server error: %%v", err)
	}
}

// runServer creates and runs HTTP server
func runServer(conf *config.Config) error {
	// create server
	handler, cleanup, err := provider.NewServer(conf)
	if err != nil {
		return fmt.Errorf("failed to create server: %%w", err)
	}
	defer cleanup()

	// create listener
	listener, err := createListener(conf)
	if err != nil {
		return fmt.Errorf("failed to create listener: %%w", err)
	}

	defer func(listener net.Listener) {
		_ = listener.Close()
	}(listener)

	// create server instance
	srv := &http.Server{
		Addr:    fmt.Sprintf("%%s:%%d", conf.Host, conf.Port),
		Handler: handler,
	}

	// create error channel
	errChan := make(chan error, 1)

	// start server
	go func() {
		logger.Infof(context.Background(), "Listening and serving HTTP on: %%s", srv.Addr)
		if err := srv.Serve(listener); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				errChan <- err
				logger.Errorf(context.Background(), "Listen error: %%s", err)
			} else {
				logger.Infof(context.Background(), "Server closed")
			}
		}
	}()

	return gracefulShutdown(srv, errChan)
}

// createListener creates network listener
func createListener(conf *config.Config) (net.Listener, error) {
	addr := fmt.Sprintf("%%s:%%d", conf.Host, conf.Port)
	if conf.Port == 0 {
		addr = fmt.Sprintf("%%s:0", conf.Host)
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("error starting server: %%w", err)
	}

	// update port if dynamically allocated
	if conf.Port == 0 {
		conf.Port = listener.Addr().(*net.TCPAddr).Port
	}

	return listener, nil
}

// loadConfig loads the application configuration
func loadConfig() *config.Config {
	conf, err := config.Init()
	if err != nil {
		logger.Fatalf(context.Background(), "[Config] Initialization error: %%+v", err)
	}
	return conf
}

// initializeLogger initializes the logger
func initializeLogger(conf *config.Config) func() {
	l, err := logger.New(conf.Logger)
	if err != nil {
		logger.Fatalf(context.Background(), "[Logger] Initialization error: %%+v", err)
	}
	return l
}

// gracefulShutdown gracefully shuts down the server
func gracefulShutdown(srv *http.Server, errChan chan error) error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		return fmt.Errorf("server error: %%w", err)

	case <-quit:
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		// Execute shutdown logic
		if err := srv.Shutdown(ctx); err != nil {
			logger.Errorf(context.Background(), "Shutdown error: %%v", err)
			return fmt.Errorf("shutdown error: %%w", err)
		}

		// wait for server to shutdown
		<-ctx.Done()
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			logger.Debugf(context.Background(), "Shutdown timed out after %%s", shutdownTimeout)
		} else {
			logger.Debugf(context.Background(), "Shutdown completed within %%s", shutdownTimeout)
		}

		return nil
	}
}
`)
}

// StandaloneServerTemplate generates the server.go file for standalone applications
func StandaloneServerTemplate(name, moduleName string) string {
	return fmt.Sprintf(`package provider

import (
	"context"
	"net/http"
	"github.com/ncobase/ncore/config"
	"github.com/ncobase/ncore/logging/logger"
	"{{ .PackagePath }}/handler"
	"{{ .PackagePath }}/service"
)

// NewServer creates a new server.
func NewServer(conf *config.Config) (http.Handler, func(), error) {
	// Initialize service layer
	svc, err := service.NewService(conf)
	if err != nil {
		logger.Fatalf(context.Background(), "Failed initializing service layer: %%+v", err)
		return nil, nil, err
	}

	// Initialize handler layer
	h, err := handler.NewHandler(svc)
	if err != nil {
		logger.Fatalf(context.Background(), "Failed initializing handler layer: %%+v", err)
		return nil, nil, err
	}

	// Initialize HTTP server
	router, err := ginServer(conf, h)
	if err != nil {
		logger.Fatalf(context.Background(), "Failed initializing http server: %%+v", err)
		return nil, nil, err
	}

	return router, func() {
		// Cleanup when server shuts down
		if err := svc.Close(); err != nil {
			logger.Errorf(context.Background(), "Error during service cleanup: %%v", err)
		}
	}, nil
}
`)
}

// StandaloneGinTemplate generates the gin.go file for standalone applications
func StandaloneGinTemplate(name, moduleName string) string {
	return fmt.Sprintf(`package provider

import (
	"github.com/ncobase/ncore/config"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/net/resp"
	"net/http"
	"{{ .PackagePath }}/handler"

	"github.com/gin-gonic/gin"
)

// ginServer creates and initializes the server.
func ginServer(conf *config.Config, h *handler.Handler) (*gin.Engine, error) {
	// Set gin mode
	if conf.RunMode == "" {
		conf.RunMode = gin.ReleaseMode
	}
	gin.SetMode(conf.RunMode)

	// Create gin engine
	engine := gin.New()

	// Add middleware here if needed
	// engine.Use(middleware.Logger)
	// engine.Use(middleware.CORSHandler)

	// Register API routes
	registerRest(engine, conf, h)

	// No route
	engine.NoRoute(func(c *gin.Context) {
		resp.Fail(c.Writer, resp.NotFound(ecode.Text(http.StatusNotFound)))
	})

	return engine, nil
}
`)
}

// StandaloneRestTemplate generates the rest.go file for standalone applications
func StandaloneRestTemplate(name, moduleName string) string {
	return fmt.Sprintf(`package provider

import (
	"github.com/ncobase/ncore/helper"
	"net/http"
	"github.com/ncobase/ncore/config"
	"{{ .PackagePath }}/handler"

	"github.com/gin-gonic/gin"
)

// registerRest registers the REST routes.
func registerRest(e *gin.Engine, conf *config.Config, h *handler.Handler) {
	// Root endpoint
	e.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Service is running.")
	})

	// Health check endpoint
	e.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"name":   conf.AppName,
			"version": version.Version,
		})
	})

	// API v1 routes
	v1 := e.Group("/api/v1")
	{
		// Add your API routes here
		v1.GET("/example", h.GetExample)
	}
}
`)
}

// StandaloneConfigTemplate generates the config.go file for standalone applications
func StandaloneConfigTemplate(name, moduleName string) string {
	return fmt.Sprintf(`package config

import (
	"github.com/ncobase/ncore/config"
)

// GetAppConfig returns the application-specific configuration
func GetAppConfig(cfg *config.Config) *config.Config {
	// Check if app configuration exists in custom section
	if cfg == nil {
		return &config.Config{}
	}

	return cfg
}
`)
}

// StandaloneHandlerTemplate generates the handler.go file for standalone applications
func StandaloneHandlerTemplate(name, moduleName string) string {
	return fmt.Sprintf(`package handler

import (
	"github.com/ncobase/ncore/net/resp"
	"{{ .PackagePath }}/service"

	"github.com/gin-gonic/gin"
)

// Handler represents the HTTP handler layer
type Handler struct {
	svc *service.Service
}

// NewHandler creates a new handler
func NewHandler(svc *service.Service) (*Handler, error) {
	return &Handler{
		svc: svc,
	}, nil
}

// GetExample is an example handler
func (h *Handler) GetExample(c *gin.Context) {
	result, err := h.svc.GetExample(c)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}
`)
}

// StandaloneModelTemplate generates the model.go file for standalone applications
func StandaloneModelTemplate(name, moduleName string) string {
	return fmt.Sprintf(`package model

import (
	"time"
)

// Example represents an example model
type Example struct {
	ID        string     ` + "`" + `json:"id"` + "`" + `
	Name      string     ` + "`" + `json:"name"` + "`" + `
	CreatedAt time.Time  ` + "`" + `json:"created_at"` + "`" + `
	UpdatedAt time.Time  ` + "`" + `json:"updated_at"` + "`" + `
	DeletedAt *time.Time ` + "`" + `json:"deleted_at,omitempty"` + "`" + `
}

// ExampleResponse represents an example response
type ExampleResponse struct {
	ID   string ` + "`" + `json:"id"` + "`" + `
	Name string ` + "`" + `json:"name"` + "`" + `
}
`)
}

// StandaloneServiceTemplate generates the service.go file for standalone applications
func StandaloneServiceTemplate(name, moduleName string) string {
	return fmt.Sprintf(`package service

import (
	"context"
	"github.com/ncobase/ncore/config"
	"{{ .PackagePath }}/model"
	"time"

	"github.com/google/uuid"
)

// Service represents the business logic layer
type Service struct {
	cfg *config.Config
}

// NewService creates a new service
func NewService(cfg *config.Config) (*Service, error) {
	return &Service{
		cfg: cfg,
	}, nil
}

// Close performs cleanup operations for the service
func (s *Service) Close() error {
	return nil
}

// GetExample is an example service method
func (s *Service) GetExample(ctx context.Context) (*model.ExampleResponse, error) {
	// This is a mock example
	example := &model.Example{
		ID:        uuid.New().String(),
		Name:      "Example Model",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return &model.ExampleResponse{
		ID:   example.ID,
		Name: example.Name,
	}, nil
}
`)
}

// StandaloneRepositoryTemplate generates the repository.go file for standalone applications
func StandaloneRepositoryTemplate(name, moduleName string, useMongo, useEnt, useGorm bool) string {
	var imports string
	var repoStruct string
	var newFunc string
	var methods string

	// Basic imports
	imports = fmt.Sprintf(`import (
	"context"
	"github.com/ncobase/ncore/config"
	"{{ .PackagePath }}/model"
`, moduleName)

	// Add DB specific imports
	if useMongo {
		imports += `	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
`
	}
	if useEnt {
		imports += fmt.Sprintf(`	"%s/repository/ent"
	"entgo.io/ent/dialect"
`, moduleName)
	}
	if useGorm {
		imports += `	"gorm.io/gorm"
`
	}

	imports += ")"

	// Repository struct
	repoStruct = "// Repository handles data access layer\ntype Repository struct {\n"
	if useMongo {
		repoStruct += "	mongoDB *mongo.Client\n"
	}
	if useEnt {
		repoStruct += "	entClient *ent.Client\n"
	}
	if useGorm {
		repoStruct += "	gormDB *gorm.DB\n"
	}
	repoStruct += "	cfg *config.Config\n}\n"

	// Constructor
	newFunc = `// NewRepository creates a new repository
func NewRepository(cfg *config.Config) (*Repository, error) {
	repo := &Repository{
		cfg: cfg,
	}
`

	if useMongo {
		newFunc += `
	// Initialize MongoDB client
	// mongoOpts := options.Client().ApplyURI(cfg.MongoDB.URI)
	// mongoClient, err := mongo.Connect(context.Background(), mongoOpts)
	// if err != nil {
	//     return nil, err
	// }
	// repo.mongoDB = mongoClient
`
	}

	if useEnt {
		newFunc += `
	// Initialize Ent client
	// driver := dialect.DebugWithContext(
	//     entsql.OpenDB(cfg.Database.Master.Driver, db),
	//     func(ctx context.Context, i ...any) {
	//         if cfg.Database.Master.Logging {
	//             log.Printf("%v", i)
	//         }
	//     },
	// )
	// entClient := ent.NewClient(ent.Driver(driver))
	// repo.entClient = entClient
`
	}

	if useGorm {
		newFunc += `
	// Initialize GORM
	// dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
	//     cfg.Database.Master.Host, cfg.Database.Master.User, cfg.Database.Master.Password,
	//     cfg.Database.Master.DBName, cfg.Database.Master.Port)
	// gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	// if err != nil {
	//     return nil, err
	// }
	// repo.gormDB = gormDB
`
	}

	newFunc += `
	return repo, nil
}

// Close closes all connections
func (r *Repository) Close() error {
`

	if useMongo {
		newFunc += `	// if r.mongoDB != nil {
	//     if err := r.mongoDB.Disconnect(context.Background()); err != nil {
	//         return err
	//     }
	// }
`
	}

	newFunc += `	return nil
}
`

	// Sample method
	methods = `
// GetExampleByID retrieves an example by ID
func (r *Repository) GetExampleByID(ctx context.Context, id string) (*model.Example, error) {
	// This is a mock implementation
	return &model.Example{
		ID:   id,
		Name: "Example from repository",
	}, nil
}
`

	// Combine all parts
	return fmt.Sprintf(`package repository

%s

%s

%s

%s
`, imports, repoStruct, newFunc, methods)
}

// StandaloneHandlerTestTemplate generates the handler_test.go file for standalone applications
func StandaloneHandlerTestTemplate(name, moduleName string) string {
	return fmt.Sprintf(`package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"{{ .PackagePath }}/handler"
	"{{ .PackagePath }}/model"
	"{{ .PackagePath }}/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHandlerGetExample(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockSvc := &service.Service{}

	// Create handler with mock service
	h, err := handler.NewHandler(mockSvc)
	assert.NoError(t, err)

	// Setup router with test handler
	router := gin.New()
	router.GET("/example", h.GetExample)

	// Create request
	req := httptest.NewRequest("GET", "/example", nil)
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Code int               ` + "`" + `json:"code"` + "`" + `
		Data model.ExampleResponse ` + "`" + `json:"data"` + "`" + `
	}

	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response.Data.ID)
	assert.NotEmpty(t, response.Data.Name)
}
`)
}

// StandaloneServiceTestTemplate generates the service_test.go file for standalone applications
func StandaloneServiceTestTemplate(name, moduleName string) string {
	return fmt.Sprintf(`package tests

import (
	"context"
	"github.com/ncobase/ncore/config"
	"testing"
	"{{ .PackagePath }}/service"

	"github.com/stretchr/testify/assert"
)

func TestServiceGetExample(t *testing.T) {
	// Create config
	cfg := &config.Config{}

	// Create service
	svc, err := service.NewService(cfg)
	assert.NoError(t, err)

	// Call method
	resp, respErr := svc.GetExample(context.Background())

	// Assert results
	assert.Nil(t, respErr)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.ID)
	assert.NotEmpty(t, resp.Name)
}
`)
}
