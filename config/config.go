package config

import (
	"context"
	"flag"
	"log"
	"os"
	"path/filepath"
	"sync"

	"ncobase/common/email"
	"ncobase/common/storage"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var (
	globalConfig *Config
	confPath     string
	once         sync.Once
	mu           sync.Mutex
	v            *viper.Viper
)

// Config is a struct representing the application's configuration.
type Config struct {
	AppName  string
	RunMode  string
	Protocol string
	Domain   string
	Host     string
	Port     int
	Observes *Observes
	Feature  *Feature
	Frontend *Frontend
	Logger   *Logger
	Data     *Data
	Auth     *Auth
	Storage  *storage.Config
	OAuth    *OAuth
	Email    *email.Email
}

func init() {
	flag.StringVar(&confPath, "conf", "", "e.g: bin ./config.yaml")
	v = viper.New()
}

// Init initializes and loads the application configuration.
func Init() (*Config, error) {
	var err error
	once.Do(func() {
		flag.Parse()
		globalConfig, err = loadConfig(confPath)
		if err != nil {
			log.Fatalf("Error loading config: %v", err)
		}
	})
	return globalConfig, err
}

// GetConfig returns the application configuration.
func GetConfig() *Config {
	if globalConfig == nil {
		if _, err := Init(); err != nil {
			log.Fatalf("Error initializing config: %v", err)
		}
	}
	return globalConfig
}

// BindConfigToContext binds the application configuration to the context.
func BindConfigToContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, "config", globalConfig)
}

func loadConfig(configPath string) (*Config, error) {
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		// Add the directory of the executable
		ex, err := os.Executable()
		if err != nil {
			return nil, err
		}

		// Set default config file name
		v.SetConfigName("config")
		// Add default config paths
		v.AddConfigPath("/etc/ncobase")
		v.AddConfigPath("$HOME/.ncobase")
		v.AddConfigPath(".")
		v.AddConfigPath(filepath.Dir(ex))
	}

	// Attempt to read the config file
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	config := &Config{
		AppName:  v.GetString("app_name"),
		RunMode:  v.GetString("run_mode"),
		Protocol: v.GetString("server.protocol"),
		Domain:   v.GetString("server.domain"),
		Host:     v.GetString("server.host"),
		Port:     v.GetInt("server.port"),
		Observes: getObservesConfig(v),
		Feature:  getFeatureConfig(v),
		Auth:     getAuth(v),
		Frontend: getFrontendConfig(v),
		Logger:   getLoggerConfig(v),
		Data:     getDataConfig(v),
		Storage:  getStorageConfig(v),
		OAuth:    getOAuthConfig(v),
		Email:    getEmailConfig(v),
	}

	return config, nil
}

// Reload reloads the configuration from the file
func Reload() error {
	mu.Lock()
	defer mu.Unlock()

	newConfig, err := loadConfig(confPath)
	if err != nil {
		log.Printf("Error reloading config: %v", err)
		return err
	}

	globalConfig = newConfig
	return nil
}

// Watch watches the configuration file and reloads it when it changes
func Watch(callback func(*Config)) {
	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		if err := Reload(); err != nil {
			log.Printf("Error reloading config: %v", err)
			return
		}
		callback(globalConfig)
	})
}
