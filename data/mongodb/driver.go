// Package mongodb provides a MongoDB driver for ncore/data.
//
// This driver uses mongo-driver (go.mongodb.org/mongo-driver) as the underlying client.
// It registers itself automatically when imported:
//
//	import _ "github.com/ncobase/ncore/data/mongodb"
//
// The driver supports MongoDB connection configuration including master/slave setup,
// load balancing strategies (round-robin, random, weight), and transaction support.
//
// Example usage:
//
//	driver, err := data.GetDatabaseDriver("mongodb")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	cfg := &config.MongoDB{
//	    Master: &config.MongoNode{
//	        URI: "mongodb://localhost:27017/mydb",
//	    },
//	    Slaves: []*config.MongoNode{
//	        {URI: "mongodb://slave1:27017/mydb", Weight: 2},
//	    },
//	    Strategy: "round_robin",
//	}
//
//	conn, err := driver.Connect(ctx, cfg)
//	manager := conn.(*MongoManager)
package mongodb

import (
	"context"
	"errors"
	"fmt"

	"github.com/ncobase/ncore/data"
	"github.com/ncobase/ncore/data/config"
)

// driver implements data.DatabaseDriver for MongoDB.
type driver struct{}

// Name returns the driver identifier used in configuration files.
func (d *driver) Name() string {
	return "mongodb"
}

// Connect establishes a MongoDB connection using the provided configuration.
//
// The configuration must be a *config.MongoDB containing:
//   - Master: Required master node configuration with URI
//   - Slaves: Optional slave nodes for read operations
//   - Strategy: Load balancing strategy ("round_robin", "random", "weight")
//   - MaxRetry: Maximum retry attempts for operations
//
// Returns a *MongoManager that handles master/slave operations,
// transaction support, and automatic health checking.
func (d *driver) Connect(ctx context.Context, cfg interface{}) (interface{}, error) {
	mongoCfg, ok := cfg.(*config.MongoDB)
	if !ok {
		return nil, fmt.Errorf("mongodb: invalid configuration type, expected *config.MongoDB")
	}

	if mongoCfg.Master == nil {
		return nil, errors.New("mongodb: master configuration is required")
	}

	if mongoCfg.Master.URI == "" {
		return nil, errors.New("mongodb: master URI is empty")
	}

	manager, err := NewMongoManager(mongoCfg)
	if err != nil {
		return nil, fmt.Errorf("mongodb: failed to create manager: %w", err)
	}

	return manager, nil
}

// Close terminates the MongoDB connection and releases resources.
// This disconnects both master and slave connections if present.
func (d *driver) Close(conn interface{}) error {
	manager, ok := conn.(*MongoManager)
	if !ok {
		return fmt.Errorf("mongodb: invalid connection type, expected *mongodb.MongoManager")
	}

	ctx := context.Background()
	if err := manager.Close(ctx); err != nil {
		return fmt.Errorf("mongodb: failed to disconnect: %w", err)
	}

	return nil
}

// Ping verifies the MongoDB connection is alive and functional.
// This performs health checks on both master and slave connections.
func (d *driver) Ping(ctx context.Context, conn interface{}) error {
	manager, ok := conn.(*MongoManager)
	if !ok {
		return fmt.Errorf("mongodb: invalid connection type, expected *mongodb.MongoManager")
	}

	if err := manager.Health(ctx); err != nil {
		return fmt.Errorf("mongodb: ping failed: %w", err)
	}

	return nil
}

// init registers the MongoDB driver with the data package.
// This function is called automatically when the package is imported.
func init() {
	data.RegisterDatabaseDriver(&driver{})
}
