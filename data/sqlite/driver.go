// Package sqlite provides a SQLite driver for ncore/data.
//
// This driver uses mattn/go-sqlite3 (github.com/mattn/go-sqlite3) as the underlying
// database/sql driver with CGO. It registers itself automatically when imported:
//
//	import _ "github.com/ncobase/ncore/data/sqlite"
//
// The driver supports standard sql.DB connection pooling and configuration options
// including max idle connections, max open connections, and connection lifetime.
//
// Example usage:
//
//	driver, err := data.GetDatabaseDriver("sqlite")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	cfg := &config.DBNode{
//	    Source: "file:test.db?cache=shared&mode=rwc",
//	    MaxIdleConn: 10,
//	    MaxOpenConn: 1, // SQLite typically uses 1 for write safety
//	}
//
//	conn, err := driver.Connect(ctx, cfg)
//	db := conn.(*sql.DB)
package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ncobase/ncore/data"
	"github.com/ncobase/ncore/data/config"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// driver implements data.DatabaseDriver for SQLite.
type driver struct{}

// Name returns the driver identifier used in configuration files.
func (d *driver) Name() string {
	return "sqlite"
}

// Connect establishes a SQLite connection using the provided configuration.
//
// The configuration must be a *config.DBNode containing:
//   - Source: SQLite connection string (file path or URI)
//   - MaxIdleConn: Maximum number of idle connections (default: 2)
//   - MaxOpenConn: Maximum number of open connections (recommended: 1 for SQLite)
//   - ConnMaxLifetime: Maximum connection lifetime
//
// Example connection strings:
//
//	"file:test.db?cache=shared&mode=rwc"        // URI format with options
//	"test.db"                                     // Simple file path
//	":memory:"                                    // In-memory database
//	"file::memory:?cache=shared"                  // Shared in-memory database
//
// Common URI parameters:
//   - cache: shared, private (default: private)
//   - mode: ro, rw, rwc, memory (default: rwc)
//   - _journal_mode: DELETE, TRUNCATE, PERSIST, MEMORY, WAL, OFF
//   - _timeout: milliseconds (default: 5000)
//
// The connection is verified with a ping before being returned.
func (d *driver) Connect(ctx context.Context, cfg any) (any, error) {
	dbCfg, ok := cfg.(*config.DBNode)
	if !ok {
		return nil, fmt.Errorf("sqlite: invalid configuration type, expected *config.DBNode")
	}

	if dbCfg.Source == "" {
		return nil, fmt.Errorf("sqlite: connection source is empty")
	}

	// Open connection using sqlite3 driver
	db, err := sql.Open("sqlite3", dbCfg.Source)
	if err != nil {
		return nil, fmt.Errorf("sqlite: failed to open connection: %w", err)
	}

	// Apply connection pool configuration
	// Note: SQLite typically works best with MaxOpenConn=1 for write safety
	if dbCfg.MaxIdleConn > 0 {
		db.SetMaxIdleConns(dbCfg.MaxIdleConn)
	} else {
		db.SetMaxIdleConns(2) // SQLite default
	}

	if dbCfg.MaxOpenConn > 0 {
		db.SetMaxOpenConns(dbCfg.MaxOpenConn)
	} else {
		db.SetMaxOpenConns(1) // SQLite recommended for write safety
	}

	if dbCfg.ConnMaxLifeTime > 0 {
		db.SetConnMaxLifetime(dbCfg.ConnMaxLifeTime)
	}

	// Verify the connection works
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("sqlite: failed to ping database: %w", err)
	}

	return db, nil
}

// Close terminates the SQLite connection and releases resources.
func (d *driver) Close(conn any) error {
	db, ok := conn.(*sql.DB)
	if !ok {
		return fmt.Errorf("sqlite: invalid connection type, expected *sql.DB")
	}

	if err := db.Close(); err != nil {
		return fmt.Errorf("sqlite: failed to close connection: %w", err)
	}

	return nil
}

// Ping verifies the SQLite connection is alive and functional.
func (d *driver) Ping(ctx context.Context, conn any) error {
	db, ok := conn.(*sql.DB)
	if !ok {
		return fmt.Errorf("sqlite: invalid connection type, expected *sql.DB")
	}

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("sqlite: ping failed: %w", err)
	}

	return nil
}

// init registers the SQLite driver with the data package.
// This function is called automatically when the package is imported.
func init() {
	data.RegisterDatabaseDriver(&driver{})
}
