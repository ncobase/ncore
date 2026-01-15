// Package data wires persistence for the basic REST API example.
package data

import (
	"context"
	"fmt"

	"github.com/ncobase/ncore/examples/01-basic-rest-api/data/ent"
	"github.com/ncobase/ncore/examples/01-basic-rest-api/data/repository"
	"github.com/ncobase/ncore/logging/logger"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// Data encapsulates all data layer dependencies.
type Data struct {
	db       *ent.Client
	TaskRepo repository.TaskRepository
}

// NewData creates a new Data instance with initialized repositories.
func NewData(db *ent.Client, logger *logger.Logger) (*Data, error) {
	// Run auto migration
	if err := db.Schema.Create(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	logger.Info(context.Background(), "Database schema created successfully")

	return &Data{
		db:       db,
		TaskRepo: repository.NewTaskRepository(db, logger),
	}, nil
}

// Close closes the database connection.
func (d *Data) Close() error {
	return d.db.Close()
}

// DB returns the Ent client for direct access if needed.
func (d *Data) DB() *ent.Client {
	return d.db
}
