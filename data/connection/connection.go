package connection

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"

	"github.com/ncobase/ncore/data/config"
)

type Connections struct {
	DBM    *DBManager
	RC     any
	MS     any
	ES     any
	OS     any
	MGM    any
	Neo    any
	RMQ    any
	KFK    any
	closed bool
	mu     sync.Mutex
}

func New(conf *config.Config) (*Connections, error) {
	c := &Connections{}
	var err error

	if conf.Database != nil && conf.Database.Master != nil && conf.Database.Master.Source != "" {
		c.DBM, err = NewDBManager(conf.Database)
		if err != nil {
			return nil, err
		}
	}

	if conf.Redis != nil && conf.Redis.Addr != "" {
		c.RC, err = newRedisClient(conf.Redis)
		if err != nil {
			return nil, err
		}
	}

	if conf.Meilisearch != nil && conf.Meilisearch.Host != "" {
		c.MS, err = newMeilisearchClient(conf.Meilisearch)
		if err != nil {
			return nil, err
		}
	}

	if conf.Elasticsearch != nil && len(conf.Elasticsearch.Addresses) > 0 {
		c.ES, err = newElasticsearchClient(conf.Elasticsearch)
		if err != nil {
			return nil, err
		}
	}

	if conf.OpenSearch != nil && len(conf.OpenSearch.Addresses) > 0 {
		c.OS, err = newOpenSearchClient(conf.OpenSearch)
		if err != nil {
			return nil, err
		}
	}

	if conf.MongoDB != nil && conf.MongoDB.Master.URI != "" {
		if driverRegistry == nil {
			return nil, errors.New("driver registry not initialized, ensure drivers are imported")
		}
		driver, err := driverRegistry.GetDatabaseDriver("mongodb")
		if err != nil {
			return nil, err
		}
		conn, err := driver.Connect(context.Background(), conf.MongoDB)
		if err != nil {
			return nil, err
		}
		c.MGM = conn
	}

	if conf.Neo4j != nil && conf.Neo4j.URI != "" {
		c.Neo, err = newNeo4jClient(conf.Neo4j)
		if err != nil {
			return nil, err
		}
	}

	if conf.Messaging != nil && conf.Messaging.IsEnabled() {
		if conf.RabbitMQ != nil && conf.RabbitMQ.URL != "" {
			c.RMQ, err = newRabbitMQConnection(conf.RabbitMQ)
			if err != nil {
				// RabbitMQ connection is optional - log warning but don't fail
				fmt.Printf("[WARN] RabbitMQ connection failed (optional): %v\n", err)
			}
		}

		if conf.Kafka != nil && conf.Kafka.Brokers != nil && len(conf.Kafka.Brokers) > 0 {
			c.KFK, err = newKafkaConnection(conf.Kafka)
			if err != nil {
				// Kafka connection is optional - log warning but don't fail
				fmt.Printf("[WARN] Kafka connection failed (optional): %v\n", err)
			}
		}
	}

	return c, nil
}

func (d *Connections) Close() (errs []error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return nil
	}

	if d.RC != nil {
		if err := d.pingRedis(context.Background()); err == nil {
			if closer, ok := d.RC.(interface{ Close() error }); ok {
				if err := closer.Close(); err != nil {
					errs = append(errs, errors.New("redis close error: "+err.Error()))
				}
			}
		}
		d.RC = nil
	}

	if d.DBM != nil {
		if err := d.DBM.Close(); err != nil {
			errs = append(errs, errors.New("database close error: "+err.Error()))
		}
		d.DBM = nil
	}

	if d.MGM != nil {
		if closer, ok := d.MGM.(interface{ Close(context.Context) error }); ok {
			if err := closer.Close(context.Background()); err != nil {
				errs = append(errs, errors.New("mongodb close error: "+err.Error()))
			}
		}
		d.MGM = nil
	}

	if d.Neo != nil {
		if closer, ok := d.Neo.(interface{ Close(context.Context) error }); ok {
			if err := closer.Close(context.Background()); err != nil {
				errs = append(errs, errors.New("neo4j close error: "+err.Error()))
			}
		}
		d.Neo = nil
	}

	if d.RMQ != nil {
		if conn, ok := d.RMQ.(interface {
			IsClosed() bool
			Close() error
		}); ok {
			if !conn.IsClosed() {
				if err := conn.Close(); err != nil {
					errs = append(errs, errors.New("rabbitmq close error: "+err.Error()))
				}
			}
		}
		d.RMQ = nil
	}

	if d.KFK != nil {
		if err := d.pingKafka(); err == nil {
			if closer, ok := d.KFK.(interface{ Close() error }); ok {
				if err := closer.Close(); err != nil {
					errs = append(errs, errors.New("kafka close error: "+err.Error()))
				}
			}
		}
		d.KFK = nil
	}

	d.MS = nil
	d.ES = nil
	d.OS = nil

	d.closed = true

	return errs
}

func (d *Connections) Ping(ctx context.Context) error {
	if d.DBM != nil {
		return d.DBM.Health(ctx)
	}
	return nil
}

func (d *Connections) DB() *sql.DB {
	if d.DBM == nil {
		return nil
	}
	return d.DBM.Master()
}

func (d *Connections) ReadDB() (*sql.DB, error) {
	if d.DBM == nil {
		return nil, errors.New("database manager is nil")
	}
	return d.DBM.Slave()
}

func (d *Connections) DBRead() (*sql.DB, error) {
	return d.ReadDB()
}

func (d *Connections) pingRedis(ctx context.Context) error {
	if d.RC == nil {
		return errors.New("redis client is nil")
	}
	if pinger, ok := d.RC.(interface {
		Ping(context.Context) error
	}); ok {
		return pinger.Ping(ctx)
	}
	if pinger, ok := d.RC.(interface {
		Ping(context.Context) interface{ Err() error }
	}); ok {
		return pinger.Ping(ctx).Err()
	}
	return errors.New("redis client ping not supported")
}

func (d *Connections) pingKafka() error {
	if d.KFK == nil {
		return errors.New("kafka connection is nil")
	}

	if controller, ok := d.KFK.(interface{ Controller() (any, error) }); ok {
		_, err := controller.Controller()
		return err
	}
	return errors.New("kafka connection controller not supported")
}
