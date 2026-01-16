// Package mysql provides a MySQL driver for ncore/data.
//
// This driver uses the official MySQL driver (github.com/go-sql-driver/mysql)
// as the underlying database/sql driver. It registers itself automatically when imported:
//
//	import _ "github.com/ncobase/ncore/data/mysql"
//
// The driver supports standard sql.DB connection pooling and configuration options
// including max idle connections, max open connections, and connection lifetime.
package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ncobase/ncore/data"
	"github.com/ncobase/ncore/data/config"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
)

// driver implements data.DatabaseDriver for MySQL.
type driver struct{}

// Name returns the driver identifier used in configuration files.
func (d *driver) Name() string {
	return "mysql"
}

// Connect establishes a MySQL connection using the provided configuration.
//
// The configuration must be a *config.DBNode containing:
//   - Source: MySQL connection string (DSN)
//   - MaxIdleConn: Maximum number of idle connections
//   - MaxOpenConn: Maximum number of open connections
//   - ConnMaxLifetime: Maximum connection lifetime
//
// Example DSN format:
//
//	user:password@tcp(localhost:3306)/dbname?parseTime=true&charset=utf8mb4
//
// The connection is verified with a ping before being returned.
func (d *driver) Connect(ctx context.Context, cfg any) (any, error) {
	dbCfg, ok := cfg.(*config.DBNode)
	if !ok {
		return nil, fmt.Errorf("mysql: invalid configuration type, expected *config.DBNode")
	}

	if dbCfg.Source == "" {
		return nil, fmt.Errorf("mysql: connection source is empty")
	}

	// Open connection using mysql driver
	db, err := sql.Open("mysql", dbCfg.Source)
	if err != nil {
		return nil, fmt.Errorf("mysql: failed to open connection: %w", err)
	}

	// Apply connection pool configuration
	if dbCfg.MaxIdleConn > 0 {
		db.SetMaxIdleConns(dbCfg.MaxIdleConn)
	}
	if dbCfg.MaxOpenConn > 0 {
		db.SetMaxOpenConns(dbCfg.MaxOpenConn)
	}
	if dbCfg.ConnMaxLifeTime > 0 {
		db.SetConnMaxLifetime(dbCfg.ConnMaxLifeTime)
	}

	// Verify the connection works
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("mysql: failed to ping database: %w", err)
	}

	return db, nil
}

// Close terminates the MySQL connection and releases resources.
func (d *driver) Close(conn any) error {
	db, ok := conn.(*sql.DB)
	if !ok {
		return fmt.Errorf("mysql: invalid connection type, expected *sql.DB")
	}

	if err := db.Close(); err != nil {
		return fmt.Errorf("mysql: failed to close connection: %w", err)
	}

	return nil
}

// Ping verifies the MySQL connection is alive and functional.
func (d *driver) Ping(ctx context.Context, conn any) error {
	db, ok := conn.(*sql.DB)
	if !ok {
		return fmt.Errorf("mysql: invalid connection type, expected *sql.DB")
	}

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("mysql: ping failed: %w", err)
	}

	return nil
}

// init registers the MySQL driver with the data package.
// This function is called automatically when the package is imported.
func init() {
	data.RegisterDatabaseDriver(&driver{})
}
