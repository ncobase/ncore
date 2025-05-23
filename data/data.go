package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ncobase/ncore/data/config"
	"github.com/ncobase/ncore/data/connection"
	"github.com/ncobase/ncore/data/messaging/kafka"
	"github.com/ncobase/ncore/data/messaging/rabbitmq"
	"github.com/ncobase/ncore/data/search/elastic"
	"github.com/ncobase/ncore/data/search/meili"
	"github.com/ncobase/ncore/data/search/opensearch"
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
	dbRead, err := d.GetSlaveDB()
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

// GetDBManager returns the database manager
func (d *Data) GetDBManager() *connection.DBManager {
	if d.Conn != nil {
		return d.Conn.DBM
	}
	return nil
}

// GetMasterDB returns the master database connection for write operations
func (d *Data) GetMasterDB() *sql.DB {
	if d.Conn != nil {
		return d.Conn.DB()
	}
	return nil
}

// GetSlaveDB returns slave database connection for read operations
func (d *Data) GetSlaveDB() (*sql.DB, error) {
	if d.Conn != nil {
		return d.Conn.ReadDB()
	}
	return nil, errors.New("no database connection available")
}

// DB returns the master database connection for write operations
// Deprecated: Use GetMasterDB() for better clarity
func (d *Data) DB() *sql.DB {
	return d.GetMasterDB()
}

// DBRead returns slave database connection for read operations
// Deprecated: Use GetSlaveDB() for better clarity
func (d *Data) DBRead() (*sql.DB, error) {
	return d.GetSlaveDB()
}

// GetDatabaseNodes returns information about all database nodes (master and slaves)
func (d *Data) GetDatabaseNodes() (master *sql.DB, slaves []*sql.DB, err error) {
	if d.Conn == nil || d.Conn.DBM == nil {
		return nil, nil, errors.New("no database manager available")
	}

	master = d.Conn.DBM.Master()

	// Get all slave connections by repeatedly calling Slave() method
	// This is a bit hacky but works with the current DBManager interface
	slavesMap := make(map[*sql.DB]bool)
	for i := 0; i < 10; i++ { // Try up to 10 times to get different slaves
		slave, err := d.Conn.DBM.Slave()
		if err != nil {
			break
		}
		if slave != master { // Only add if it's not the master
			slavesMap[slave] = true
		}
	}

	for slave := range slavesMap {
		slaves = append(slaves, slave)
	}

	return master, slaves, nil
}

// IsReadOnlyMode checks if the system is in read-only mode (only slaves available)
func (d *Data) IsReadOnlyMode(ctx context.Context) bool {
	if d.Conn == nil || d.Conn.DBM == nil {
		return false
	}

	master := d.Conn.DBM.Master()
	if master == nil {
		return true
	}

	// Check if master is healthy
	if err := master.PingContext(ctx); err != nil {
		return true
	}

	return false
}

// GetRedis returns the Redis client
func (d *Data) GetRedis() *redis.Client {
	if d.Conn != nil {
		return d.Conn.RC
	}
	return nil
}

// GetMeilisearch returns the Meilisearch client
func (d *Data) GetMeilisearch() *meili.Client {
	if d.Conn != nil {
		return d.Conn.MS
	}
	return nil
}

// GetElasticsearch returns the Elasticsearch client
func (d *Data) GetElasticsearch() *elastic.Client {
	if d.Conn != nil {
		return d.Conn.ES
	}
	return nil
}

// GetOpenSearch returns the OpenSearch client
func (d *Data) GetOpenSearch() *opensearch.Client {
	if d.Conn != nil {
		return d.Conn.OS
	}
	return nil
}

// GetMongoManager returns the MongoDB client
func (d *Data) GetMongoManager() *connection.MongoManager {
	if d.Conn != nil {
		return d.Conn.MGM
	}
	return nil
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
