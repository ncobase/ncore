// Package data manages SQLite persistence for authentication.
package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/ncobase/ncore/logging/logger"
)

type Data struct {
	db *sql.DB
}

func New(driver, source string, log *logger.Logger) (*Data, error) {
	db, err := sql.Open(driver, source)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	log.Info(ctx, "Database connected", "driver", driver, "source", source)
	return &Data{db: db}, nil
}

func (d *Data) DB() *sql.DB {
	return d.db
}

func (d *Data) Close() error {
	return d.db.Close()
}
