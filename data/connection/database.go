package connection

import (
	"context"
	"database/sql"
	"ncobase/common/config"
	"ncobase/common/log"

	// sqlite3
	_ "github.com/mattn/go-sqlite3"
	// mysql
	_ "github.com/go-sql-driver/mysql"
	// postgres
	_ "github.com/jackc/pgx/v5/stdlib"
)

// newDBClient creates a new database client
func newDBClient(conf *config.Database) (*sql.DB, error) {
	var db *sql.DB
	var err error

	switch conf.Driver {
	case "postgres":
		db, err = sql.Open("pgx", conf.Source)
	case "mysql":
		db, err = sql.Open("mysql", conf.Source)
	case "sqlite3":
		db, err = sql.Open("sqlite3", conf.Source)
	default:
		log.Fatalf(context.Background(), "Dialect %v not supported", conf.Driver)
		return nil, err
	}

	if err != nil {
		log.Fatalf(context.Background(), "Failed to open database: %v", err)
		return nil, err
	}

	db.SetMaxIdleConns(conf.MaxIdleConn)
	db.SetMaxOpenConns(conf.MaxOpenConn)
	db.SetConnMaxLifetime(conf.ConnMaxLifeTime)

	log.Infof(context.Background(), "Database %v connected", conf.Driver)

	return db, nil
}
