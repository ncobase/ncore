package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"ncobase/common/data/config"
	"ncobase/common/data/connection"
	"ncobase/common/data/elastic"
	"ncobase/common/data/kafka"
	"ncobase/common/data/meili"
	"ncobase/common/data/rabbitmq"

	"github.com/redis/go-redis/v9"
)

type ContextKey string

const (
	// ContextKeyTransaction is context key
	ContextKeyTransaction ContextKey = "tx"
)

var (
	// sharedInstance is shared instance
	sharedInstance *Data
)

// Data represents the data layer implementation
type Data struct {
	Conn     *connection.Connections
	RabbitMQ *rabbitmq.RabbitMQ
	Kafka    *kafka.Kafka
}

// Option function type for configuring Connections
type Option func(*Data)

// New creates new data layer
func New(cfg *config.Config, createNewInstance ...bool) (*Data, func(name ...string), error) {
	var createNew bool
	if len(createNewInstance) > 0 {
		createNew = createNewInstance[0]
	}

	if !createNew && sharedInstance != nil {
		cleanup := func(name ...string) {
			if errs := sharedInstance.Close(); len(errs) > 0 {
				fmt.Printf("cleanup errors: %v\n", errs)
			}
		}
		return sharedInstance, cleanup, nil
	}

	conn, err := connection.New(cfg)
	if err != nil {
		return nil, nil, err
	}

	d := &Data{
		Conn:     conn,
		RabbitMQ: rabbitmq.NewRabbitMQ(conn.RMQ),
		Kafka:    kafka.New(conn.KFK),
	}

	if !createNew {
		sharedInstance = d
	}

	cleanup := func(name ...string) {
		if errs := d.Close(); len(errs) > 0 {
			fmt.Printf("cleanup errors: %v\n", errs)
		}
	}

	return d, cleanup, nil
}

// GetTx retrieves transaction from context
func GetTx(ctx context.Context) (*sql.Tx, error) {
	tx, ok := ctx.Value(ContextKeyTransaction).(*sql.Tx)
	if !ok {
		return nil, errors.New("transaction not found in context")
	}
	return tx, nil
}

// WithTx wraps function within transaction
func (d *Data) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	db := d.DB()
	if db == nil {
		return errors.New("database connection is nil")
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	err = fn(context.WithValue(ctx, ContextKeyTransaction, tx))
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rollback err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

// WithTxRead wraps function within read-only transaction
func (d *Data) WithTxRead(ctx context.Context, fn func(ctx context.Context) error) error {
	dbRead, err := d.DBRead()
	if err != nil {
		return err
	}

	tx, err := dbRead.BeginTx(ctx, &sql.TxOptions{
		ReadOnly: true,
	})
	if err != nil {
		return err
	}

	if err = fn(context.WithValue(ctx, ContextKeyTransaction, tx)); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rollback err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

// GetDBManager get database manager
func (d *Data) GetDBManager() *connection.DBManager {
	if d.Conn != nil {
		return d.Conn.DBM
	}
	return nil
}

// DB returns the master database connection for write operations
func (d *Data) DB() *sql.DB {
	if d.Conn != nil {
		return d.Conn.DB()
	}
	return nil
}

// DBRead returns slave database connection for read operations
func (d *Data) DBRead() (*sql.DB, error) {
	if d.Conn != nil {
		return d.Conn.DBRead()
	}
	return nil, errors.New("no database connection available")
}

// GetRedis get redis
func (d *Data) GetRedis() *redis.Client {
	return d.Conn.RC
}

// GetMeilisearch get meilisearch
func (d *Data) GetMeilisearch() *meili.Client {
	return d.Conn.MS
}

// GetElasticsearch get elasticsearch
func (d *Data) GetElasticsearch() *elastic.Client {
	return d.Conn.ES
}

// GetMongoManager get mongo manager
func (d *Data) GetMongoManager() *connection.MongoManager {
	return d.Conn.MGM
}

// Ping checks all database connections
func (d *Data) Ping(ctx context.Context) error {
	if d.Conn != nil {
		return d.Conn.Ping(ctx)
	}
	return errors.New("no connection manager available")
}

// Close closes all data connections
func (d *Data) Close() []error {
	var errs []error

	// Close connections
	if d.Conn != nil {
		if connErrs := d.Conn.Close(); len(connErrs) > 0 {
			errs = append(errs, connErrs...)
		}
	}

	return errs
}
