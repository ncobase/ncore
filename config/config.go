package config

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var (
	config *Config
	path   string
	once   sync.Once
	mu     sync.Mutex
	v      *viper.Viper
)

// Config represents the configuration implementation.
type Config struct {
	AppName   string
	RunMode   string
	Protocol  string
	Domain    string
	Host      string
	Port      int
	Consul    *Consul
	Observes  *Observes
	Extension *Extension
	Frontend  *Frontend
	Logger    *Logger
	Data      *Data
	Auth      *Auth
	Storage   *Storage
	OAuth     *OAuth
	Email     *Email
	Viper     *viper.Viper
}

func init() {
	flag.StringVar(&path, "conf", "", "e.g: bin ./config.yaml")
	v = viper.New()
}

// Init initializes and loads the configuration.
func Init() (cfg *Config, err error) {
	once.Do(func() {
		cfg, err = loadConfiguration()
	})
	return cfg, err
}

// GetConfig returns the configuration.
// It does not handle errors internally; instead, it returns the error for the caller to handle.
func GetConfig() (*Config, error) {
	if config == nil {
		var err error
		config, err = Init()
		if err != nil {
			return nil, fmt.Errorf("failed to initialize config: %w", err)
		}
	}
	return config, nil
}

// BindConfigToContext binds the configuration to the context.
func BindConfigToContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, "config", config)
}

// loadConfiguration loads the configuration from the file and sets it globally.
func loadConfiguration() (*Config, error) {
	flag.Parse()
	cfg, err := LoadConfig(path)
	if err != nil {
		return nil, fmt.Errorf("error loading config: %w", err)
	}
	config = cfg
	return cfg, nil
}

// LoadConfig loads the configuration from the file.
func LoadConfig(configPath string) (*Config, error) {
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		ex, err := os.Executable()
		if err != nil {
			return nil, fmt.Errorf("failed to get executable path: %w", err)
		}
		v.SetConfigName("config")
		v.AddConfigPath("/etc/ncobase")
		v.AddConfigPath("$HOME/.ncobase")
		v.AddConfigPath(".")
		v.AddConfigPath(filepath.Dir(ex))
	}

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := &Config{
		AppName:   v.GetString("app_name"),
		RunMode:   v.GetString("run_mode"),
		Protocol:  v.GetString("server.protocol"),
		Domain:    v.GetString("server.domain"),
		Host:      v.GetString("server.host"),
		Port:      v.GetInt("server.port"),
		Consul:    getConsulConfig(v),
		Observes:  getObservesConfig(v),
		Extension: getExtensionConfig(v),
		Auth:      getAuth(v),
		Frontend:  getFrontendConfig(v),
		Logger:    getLoggerConfig(v),
		Data:      getDataConfig(v),
		Storage:   getStorageConfig(v),
		OAuth:     getOAuthConfig(v),
		Email:     getEmailConfig(v),
		Viper:     v,
	}

	return cfg, nil
}

// Reload reloads the configuration from the file.
func Reload() error {
	mu.Lock()
	defer mu.Unlock()

	newConfig, err := LoadConfig(path)
	if err != nil {
		return fmt.Errorf("failed to reload config: %w", err)
	}

	config = newConfig
	return nil
}

// Watch watches the configuration file and reloads it when it changes.
func Watch(callback func(*Config)) {
	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		if err := Reload(); err != nil {
			fmt.Printf("Error reloading config: %v\n", err)
			return
		}
		callback(config)
	})
}
