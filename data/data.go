package data

import (
	"context"
	"database/sql"
	"fmt"
	"ncobase/common/config"
	"ncobase/common/data/connection"
	"ncobase/common/data/service"
	"ncobase/common/elastic"
	"ncobase/common/log"
	"ncobase/common/meili"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

var (
	sharedInstance    *Data
	errExecuteCleanup = "execute %s data cleanup."
)

type Data struct {
	Conn *connection.Connections
	Svc  *service.Services
}

// Option function type for configuring Connections
type Option func(*Data)

func New(conf *config.Data, createNewInstance ...bool) (*Data, func(name ...string), error) {
	var createNew bool
	if len(createNewInstance) > 0 {
		createNew = createNewInstance[0]
	}

	if !createNew && sharedInstance != nil {
		cleanup := func(name ...string) {
			// log.Infof(context.Background(), errExecuteCleanup, name[0])
			if errs := sharedInstance.Close(); len(errs) > 0 {
				log.Fatalf(context.Background(), "cleanup errors: %v", errs)
			}
		}
		return sharedInstance, cleanup, nil
	}

	conn, err := connection.New(conf)
	if err != nil {
		return nil, nil, err
	}

	d := &Data{
		Conn: conn,
		Svc:  service.New(conn),
	}

	if !createNew {
		sharedInstance = d
	}

	cleanup := func(name ...string) {
		// log.Infof(context.Background(), errExecuteCleanup, name[0])
		if errs := d.Close(); len(errs) > 0 {
			log.Fatalf(context.Background(), "cleanup errors: %v", errs)
		}
	}

	return d, cleanup, nil
}

// GetTx retrieves transaction from context
func GetTx(ctx context.Context) (*sql.Tx, error) {
	tx, ok := ctx.Value("tx").(*sql.Tx)
	if !ok {
		return nil, fmt.Errorf("transaction not found in context")
	}
	return tx, nil
}

// WithTx wraps a function within a transaction
func (d *Data) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	db := d.DB()
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	err = fn(context.WithValue(ctx, "tx", tx))
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

// WithTxRead wraps a function within a read-only transaction
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

	err = fn(context.WithValue(ctx, "tx", tx))
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
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

// DBRead returns a slave database connection for read operations
func (d *Data) DBRead() (*sql.DB, error) {
	if d.Conn != nil {
		return d.Conn.DBRead()
	}
	return nil, nil
}

// GetRedis get redis
func (d *Data) GetRedis() *redis.Client {
	return d.Conn.RC
}

// GetMeilisearch get meilisearch
func (d *Data) GetMeilisearch() *meili.Client {
	return d.Conn.MS
}

// GetElasticsearch get meilisearch
func (d *Data) GetElasticsearch() *elastic.Client {
	return d.Conn.ES
}

// GetMongoManager get mongo manager
func (d *Data) GetMongoManager() *connection.MongoManager {
	return d.Conn.MGM
}

// GetNeo4j get neo4j
func (d *Data) GetNeo4j() neo4j.DriverWithContext {
	return d.Conn.Neo
}

// GetRabbitMQ get rabbitMQ
func (d *Data) GetRabbitMQ() *amqp.Connection {
	return d.Conn.RMQ
}

// GetKafka get kafka
func (d *Data) GetKafka() *kafka.Conn {
	return d.Conn.KFK
}

// Ping checks all database connections
func (d *Data) Ping(ctx context.Context) error {
	if d.Conn != nil {
		return d.Conn.Ping(ctx)
	}
	return nil
}

func (d *Data) Close() (errs []error) {
	// Close connections
	if connErrs := d.Conn.Close(); len(connErrs) > 0 {
		errs = append(errs, connErrs...)
	}

	// Close services
	if svcEerrs := d.Svc.Close(); len(svcEerrs) > 0 {
		errs = append(errs, svcEerrs...)
	}

	return errs
}
