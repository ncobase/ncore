package templates

import "fmt"

// CmdMainTemplate generates the main.go file for the cmd directory
func CmdMainTemplate(name, extType, moduleName string) string {
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
	"github.com/ncobase/ncore/pkg/config"
	"github.com/ncobase/ncore/pkg/helper"
	"github.com/ncobase/ncore/pkg/logger"
)

const (
	shutdownTimeout = 3 * time.Second // service shutdown timeout
)

func main() {
	logger.SetVersion(helper.Version)
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

// CmdServerTemplate generates the server.go file for the provider directory
func CmdServerTemplate(name, extType, moduleName string) string {
	return fmt.Sprintf(`package provider

import (
	"context"
	nem "github.com/ncobase/ncore/ext/manager"
	"github.com/ncobase/ncore/pkg/config"
	"github.com/ncobase/ncore/pkg/logger"
	"net/http"
)

// NewServer creates a new server.
func NewServer(conf *config.Config) (http.Handler, func(), error) {
	// Initialize Extension Manager
	em, err := nem.NewManager(conf)
	if err != nil {
		logger.Fatalf(context.Background(), "Failed initializing extension manager: %%+v", err)
		return nil, nil, err
	}

	// Register extensions
	registerExtensions(em)
	if err := em.LoadPlugins(); err != nil {
		logger.Fatalf(context.Background(), "Failed loading plugins: %%+v", err)
	}

	// New server
	h, err := ginServer(conf, em)
	if err != nil {
		logger.Fatalf(context.Background(), "Failed initializing http: %%+v", err)
	}

	return h, func() {
		em.Cleanup()
	}, nil
}
`)
}

// CmdExtensionTemplate generates the extension.go file for the provider directory
func CmdExtensionTemplate(name, extType, moduleName string) string {
	return fmt.Sprintf(`package provider

import (
	"context"
	// Import your extensions here
	"{{ .PackagePath }}" // Import the current extension
	nec "github.com/ncobase/ncore/ext/core"
	"github.com/ncobase/ncore/pkg/logger"
	"strings"
)

// registerExtensions registers all extensions
func registerExtensions(em nec.ManagerInterface) {
	// All components
	fs := make([]nec.Interface, 0)

	// Add your extensions here
	fs = append(fs, %s.New()) // Register the current extension

	// Registered extensions
	registered := make([]nec.Interface, 0, len(fs))
	// Get extension names
	extensionNames := make([]string, 0, len(registered))

	for _, f := range fs {
		if err := em.Register(f); err != nil {
			logger.Errorf(context.Background(), "Failed to register extension %%s: %%v", f.Name(), err)
			continue // Skip this extension and try to register the next one
		}
		registered = append(registered, f)
		extensionNames = append(extensionNames, f.Name())
	}

	if len(registered) == 0 {
		logger.Warnf(context.Background(), "No extensions were registered.")
		return
	}

	logger.Debugf(context.Background(), "Successfully registered %%d extensions, [%%s]",
		len(registered),
		strings.Join(extensionNames, ", "))

	if err := em.InitExtensions(); err != nil {
		logger.Errorf(context.Background(), "Failed to initialize extensions: %%v", err)
		return
	}
}
`, name)
}

// CmdGinTemplate generates the gin.go file for the provider directory
func CmdGinTemplate(name, extType, moduleName string) string {
	return fmt.Sprintf(`package provider

import (
	nec "github.com/ncobase/ncore/ext/core"
	"github.com/ncobase/ncore/pkg/config"
	"github.com/ncobase/ncore/pkg/ecode"
	"github.com/ncobase/ncore/pkg/resp"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ginServer creates and initializes the server.
func ginServer(conf *config.Config, em nec.ManagerInterface) (*gin.Engine, error) {
	// Set gin mode
	if conf.RunMode == "" {
		conf.RunMode = gin.ReleaseMode
	}
	// Set mode before creating engine
	gin.SetMode(conf.RunMode)
	// Create gin engine
	engine := gin.New()

	// Add middleware here if needed
	// engine.Use(middleware.Logger)
	// engine.Use(middleware.CORSHandler)

	// Register REST
	registerRest(engine, conf)

	// Register extension / plugin routes
	em.RegisterRoutes(engine)

	// Register extension management routes
	if conf.Extension.HotReload {
		g := engine.Group("/sys")
		em.ManageRoutes(g)
	}

	// No route
	engine.NoRoute(func(c *gin.Context) {
		resp.Fail(c.Writer, resp.NotFound(ecode.Text(http.StatusNotFound)))
	})

	return engine, nil
}
`)
}

// CmdRestTemplate generates the rest.go file for the provider directory
func CmdRestTemplate(name, extType, moduleName string) string {
	return fmt.Sprintf(`package provider

import (
	"github.com/ncobase/ncore/pkg/helper"
	"net/http"

	"github.com/ncobase/ncore/pkg/config"

	"github.com/gin-gonic/gin"
)

// registerRest registers the REST routes.
func registerRest(e *gin.Engine, conf *config.Config) {
	// Root endpoint
	e.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Service is running.")
	})

	// Health check endpoint
	e.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"name":   conf.AppName,
			"version": helper.Version,
		})
	})

	// Add your API routes here
}
`)
}
