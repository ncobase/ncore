package config

import (
	"context"
	"flag"
	"ncobase/common/email"
	"ncobase/common/storage"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

var (
	c            *viper.Viper
	globalConfig *Config
	confPath     string
)

// Config is a struct representing the application's configuration.
type Config struct {
	AppName    string
	RunMode    string
	Protocol   string
	Domain     string
	Host       string
	Port       int
	JWTSecret  string
	JWTExpTime int
	Frontend   Frontend
	Logger     Logger
	Data       Data
	OAuth      OAuth
	Storage    storage.Config
	Email      email.Email
}

func init() {
	flag.StringVar(&confPath, "conf", "", "e.g: bin ./config.yaml")
}

// Init initializes and loads the application configuration.
func Init() (*Config, error) {
	flag.Parse()
	conf, err := load(confPath)
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

func load(in string) (*Config, error) {
	c = viper.New()

	// Add the directory of the executable
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}

	c.SetConfigFile(in)
	// By default, read from config.{yaml,toml, yml,json}, etc. files
	c.AddConfigPath(in)
	c.AddConfigPath("/etc/ncobase")
	c.AddConfigPath("$HOME/.ncobase")
	c.AddConfigPath(".")
	c.AddConfigPath(filepath.Dir(ex))

	err = c.ReadInConfig()

	return &Config{
		AppName:    c.GetString("app_name"),
		RunMode:    c.GetString("run_mode"),
		Protocol:   c.GetString("server.protocol"),
		Domain:     c.GetString("server.domain"),
		Host:       c.GetString("server.host"),
		Port:       c.GetInt("server.port"),
		JWTSecret:  c.GetString("jwt.secret"),
		JWTExpTime: c.GetInt("jwt.exp_time"),
		Frontend:   getFrontendConfig(),
		Logger:     getLoggerConfig(),
		Data:       getDataConfig(),
		OAuth:      getOAuthConfig(),
		Storage:    getStorageConfig(),
		Email:      getEmailConfig(),
	}, err
}
