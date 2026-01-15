// Package data manages MongoDB connections for the example API.
package data

import (
	"context"
	"fmt"
	"time"

	"github.com/ncobase/ncore/examples/02-mongodb-api/data/repository"
	"github.com/ncobase/ncore/logging/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Data encapsulates all data layer dependencies.
type Data struct {
	client   *mongo.Client
	db       *mongo.Database
	UserRepo repository.UserRepository
}

// New creates a new Data instance with MongoDB connection.
func New(mongoURI string, dbName string, logger *logger.Logger) (*Data, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	logger.Info(ctx, "Connected to MongoDB successfully", "uri", mongoURI, "database", dbName)

	db := client.Database(dbName)

	return &Data{
		client:   client,
		db:       db,
		UserRepo: repository.NewUserRepository(db, logger),
	}, nil
}

// Close closes the MongoDB connection.
func (d *Data) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return d.client.Disconnect(ctx)
}

// DB returns the MongoDB database instance.
func (d *Data) DB() *mongo.Database {
	return d.db
}
