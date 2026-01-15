package main

import (
	"github.com/ncobase/ncore/config"
	"github.com/ncobase/ncore/data"
	"github.com/ncobase/ncore/extension/manager"
	"github.com/ncobase/ncore/logging/logger"
)

type App struct {
	Config  *config.Config
	Logger  *logger.Logger
	Data    *data.Data
	Manager *manager.Manager
}

func NewApp(cfg *config.Config, log *logger.Logger, d *data.Data, mgr *manager.Manager) *App {
	return &App{
		Config:  cfg,
		Logger:  log,
		Data:    d,
		Manager: mgr,
	}
}
