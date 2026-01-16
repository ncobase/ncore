// Package neo4j provides a Neo4j driver for ncore/data.
//
// This driver uses neo4j-go-driver (github.com/neo4j/neo4j-go-driver/v5) as the
// underlying graph database client. It registers itself automatically when imported:
//
//	import _ "github.com/ncobase/ncore/data/neo4j"
//
// The driver supports Neo4j connection configuration including authentication,
// connection pooling, and health checking.
//
// Example usage:
//
//	driver, err := data.GetDatabaseDriver("neo4j")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	cfg := &config.Neo4j{
//	    URI:      "neo4j://localhost:7687",
//	    Username: "neo4j",
//	    Password: "password",
//	}
//
//	conn, err := driver.Connect(ctx, cfg)
//	neo4jDriver := conn.(neo4j.DriverWithContext)
package neo4j

import (
	"context"
	"fmt"

	"github.com/ncobase/ncore/data"
	"github.com/ncobase/ncore/data/config"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// driver implements data.DatabaseDriver for Neo4j.
type driver struct{}

// Name returns the driver identifier used in configuration files.
func (d *driver) Name() string {
	return "neo4j"
}

// Connect establishes a Neo4j connection using the provided configuration.
//
// The configuration must be a *config.Neo4j containing:
//   - URI: Neo4j connection URI (e.g., "neo4j://localhost:7687")
//   - Username: Authentication username
//   - Password: Authentication password
//
// Example URIs:
//
//	"neo4j://localhost:7687"               // Bolt protocol
//	"neo4j+s://dbhash.databases.neo4j.io"  // Bolt with TLS
//	"bolt://localhost:7687"                // Direct Bolt protocol
//
// Returns a neo4j.DriverWithContext that can be used for database operations.
func (d *driver) Connect(ctx context.Context, cfg any) (any, error) {
	neo4jCfg, ok := cfg.(*config.Neo4j)
	if !ok {
		return nil, fmt.Errorf("neo4j: invalid configuration type, expected *config.Neo4j")
	}

	if neo4jCfg.URI == "" {
		return nil, fmt.Errorf("neo4j: URI is empty")
	}

	auth := neo4j.BasicAuth(neo4jCfg.Username, neo4jCfg.Password, "")

	neoDriver, err := neo4j.NewDriverWithContext(neo4jCfg.URI, auth)
	if err != nil {
		return nil, fmt.Errorf("neo4j: failed to create driver: %w", err)
	}

	if err := neoDriver.VerifyConnectivity(ctx); err != nil {
		neoDriver.Close(ctx)
		return nil, fmt.Errorf("neo4j: connectivity verification failed: %w", err)
	}

	return neoDriver, nil
}

// Close terminates the Neo4j connection and releases resources.
func (d *driver) Close(conn any) error {
	neoDriver, ok := conn.(neo4j.DriverWithContext)
	if !ok {
		return fmt.Errorf("neo4j: invalid connection type, expected neo4j.DriverWithContext")
	}

	ctx := context.Background()
	if err := neoDriver.Close(ctx); err != nil {
		return fmt.Errorf("neo4j: failed to close connection: %w", err)
	}

	return nil
}

// Ping verifies the Neo4j connection is alive and functional.
func (d *driver) Ping(ctx context.Context, conn any) error {
	neoDriver, ok := conn.(neo4j.DriverWithContext)
	if !ok {
		return fmt.Errorf("neo4j: invalid connection type, expected neo4j.DriverWithContext")
	}

	if err := neoDriver.VerifyConnectivity(ctx); err != nil {
		return fmt.Errorf("neo4j: connectivity verification failed: %w", err)
	}

	return nil
}

// init registers the Neo4j driver with the data package.
// This function is called automatically when the package is imported.
func init() {
	data.RegisterDatabaseDriver(&driver{})
}
