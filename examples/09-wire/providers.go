package main

import (
	"github.com/ncobase/ncore/concurrency/worker"
	"github.com/ncobase/ncore/config"
	"github.com/ncobase/ncore/security/jwt"
)

type CustomWorkerConfig struct {
	MaxWorkers int
	QueueSize  int
}

func ProvideJWTConfig(auth *config.Auth) *jwt.Config {
	if auth == nil || auth.JWT == nil {
		return &jwt.Config{}
	}
	return &jwt.Config{
		Secret: auth.JWT.Secret,
	}
}

func ProvideDefaultWorkerConfig() *worker.Config {
	return worker.DefaultConfig()
}

func ProvideWorkerConfigFromCustom(custom *CustomWorkerConfig) *worker.Config {
	if custom == nil {
		return worker.DefaultConfig()
	}
	cfg := worker.DefaultConfig()
	if custom.MaxWorkers > 0 {
		cfg.MaxWorkers = custom.MaxWorkers
	}
	if custom.QueueSize > 0 {
		cfg.QueueSize = custom.QueueSize
	}
	return cfg
}
