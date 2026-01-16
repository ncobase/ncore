//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/ncobase/ncore/config"
	"github.com/ncobase/ncore/examples/01-basic-rest-api/data"
	"github.com/ncobase/ncore/examples/01-basic-rest-api/data/ent"
	"github.com/ncobase/ncore/examples/01-basic-rest-api/handler"
	"github.com/ncobase/ncore/examples/01-basic-rest-api/service"
	"github.com/ncobase/ncore/logging/logger"

	_ "github.com/ncobase/ncore/data/postgres"
)

// InitializeApp wires up the entire application with all dependencies.
func InitializeApp() (*App, func(), error) {
	panic(wire.Build(
		// Config provider
		config.ProviderSet,

		// Logger provider
		logger.ProviderSet,

		// Database client provider
		ProvideEntClient,

		// Data layer provider
		data.NewData,

		// Service layer provider
		service.NewService,

		// Handler layer provider
		handler.NewHandler,

		// Application constructor
		NewApp,
	))
}

// ProvideEntClient creates an Ent database client from configuration.
func ProvideEntClient(dataCfg *config.Data) (*ent.Client, error) {
	if dataCfg == nil || dataCfg.Database == nil || dataCfg.Database.Master == nil {
		return nil, ErrInvalidConfig
	}

	client, err := ent.Open(
		dataCfg.Database.Master.Driver,
		dataCfg.Database.Master.Source,
	)
	if err != nil {
		return nil, err
	}

	return client, nil
}
