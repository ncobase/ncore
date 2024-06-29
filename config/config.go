package config

import (
	"context"
	"flag"
	"os"
	"path/filepath"

	"ncobase/common/email"
	"ncobase/common/storage"

	"github.com/spf13/viper"
)

var (
	globalConfig *Config
	confPath     string
)

// Config is a struct representing the application's configuration.
type Config struct {
	AppName  string
	RunMode  string
	Protocol string
	Domain   string
	Host     string
	Port     int
	Plugin   Plugin
	Frontend Frontend
	Logger   Logger
	Data     Data
	Auth     Auth
	Storage  storage.Config
	OAuth    OAuth
	Email    email.Email
}

func init() {
	flag.StringVar(&confPath, "conf", "", "e.g: bin ./config.yaml")
}

// Init initializes and loads the application configuration.
func Init() (*Config, error) {
	flag.Parse()
	conf, err := loadConfig(confPath)
	if err == nil {
		globalConfig = conf
	}
	return conf, err
}

// GetConfig returns the application configuration.
func GetConfig() *Config {
	return globalConfig
}

// BindConfigToContext binds the application configuration to the context.
func BindConfigToContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, "config", globalConfig)
}

func loadConfig(configPath string) (*Config, error) {
	v := viper.New()

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
		Plugin:   getPluginConfig(v),
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
