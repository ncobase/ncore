package connection

import (
	"context"
	"database/sql"
	"errors"
	"ncobase/common/config"
	"ncobase/common/elastic"
	"ncobase/common/log"
	"ncobase/common/meili"
	"sync"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/mongo"
)

// Connections struct to hold all database connections and clients
type Connections struct {
	DBM    *DBManager
	RC     *redis.Client
	MS     *meili.Client
	ES     *elastic.Client
	MG     *mongo.Client
	Neo    neo4j.DriverWithContext
	RMQ    *amqp.Connection
	KFK    *kafka.Conn
	closed bool
	mu     sync.Mutex
}

// New creates a new Connections
func New(conf *config.Data) (*Connections, error) {
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

	if conf.MongoDB != nil && conf.MongoDB.URI != "" {
		c.MG, err = newMongoClient(conf.MongoDB)
		if err != nil {
			return nil, err
		}
	}

	if conf.Neo4j != nil && conf.Neo4j.URI != "" {
		c.Neo, err = newNeo4jClient(conf.Neo4j)
		if err != nil {
			return nil, err
		}
	}

	if conf.RabbitMQ != nil && conf.RabbitMQ.URL != "" {
		c.RMQ, err = newRabbitMQConnection(conf.RabbitMQ)
		if err != nil {
			return nil, err
		}
	}

	if conf.Kafka != nil && conf.Kafka.Brokers != nil && len(conf.Kafka.Brokers) > 0 {
		c.KFK, err = newKafkaConnection(conf.Kafka)
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}

// Close closes all data connections
func (d *Connections) Close() (errs []error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Check if already closed
	if d.closed {
		return nil
	}

	// Close Redis client if connected
	if d.RC != nil {
		if err := d.pingRedis(context.Background()); err == nil {
			if err := d.RC.Close(); err != nil {
				errs = append(errs, errors.New("redis close error: "+err.Error()))
			}
		}
		d.RC = nil
	}

	// Close database connections if connected
	if d.DBM != nil {
		if err := d.DBM.Close(); err != nil {
			errs = append(errs, errors.New("database close error: "+err.Error()))
		}
		d.DBM = nil
	}

	// Disconnect MongoDB client if connected
	if d.MG != nil {
		if err := d.MG.Ping(context.Background(), nil); err == nil {
			if err := d.MG.Disconnect(context.Background()); err != nil {
				errs = append(errs, errors.New("mongodb close error: "+err.Error()))
			}
		}
		d.MG = nil
	}

	// Close Neo4j client if connected
	if d.Neo != nil {
		if err := d.Neo.Close(context.Background()); err != nil {
			errs = append(errs, errors.New("neo4j close error: "+err.Error()))
		}
		d.Neo = nil
	}

	// Close RabbitMQ client if connected
	if d.RMQ != nil {
		if !d.RMQ.IsClosed() {
			if err := d.RMQ.Close(); err != nil {
				errs = append(errs, errors.New("rabbitmq close error: "+err.Error()))
			}
		}
		d.RMQ = nil
	}

	// Close Kafka client if connected
	if d.KFK != nil {
		if err := d.pingKafka(); err == nil {
			if err := d.KFK.Close(); err != nil {
				errs = append(errs, errors.New("kafka close error: "+err.Error()))
			}
		}
		d.KFK = nil
	}

	d.closed = true

	return errs
}

// Ping checks all database connections
func (d *Connections) Ping(ctx context.Context) error {
	if d.DBM != nil {
		return d.DBM.Health(ctx)
	}
	return nil
}

// DB returns the master database connection for write operations
func (d *Connections) DB() *sql.DB {
	if d.DBM == nil {
		return nil
	}
	return d.DBM.Master()
}

// DBRead returns a slave database connection for read operations
func (d *Connections) DBRead() (*sql.DB, error) {
	if d.DBM == nil {
		return nil, errors.New("database manager is nil")
	}
	return d.DBM.Slave()
}

// GetMongoDatabase retrieves a specific MongoDB database
func (d *Connections) GetMongoDatabase(databaseName string) *mongo.Database {
	if d.MG == nil {
		log.Errorf(context.Background(), "MongoDB client is nil")
		return nil
	}
	return d.MG.Database(databaseName)
}

// pingRedis checks if Redis connection is alive
func (d *Connections) pingRedis(ctx context.Context) error {
	if d.RC == nil {
		return errors.New("redis client is nil")
	}
	return d.RC.Ping(ctx).Err()
}

// pingKafka checks if Kafka connection is alive
func (d *Connections) pingKafka() error {
	if d.KFK == nil {
		return errors.New("kafka connection is nil")
	}

	// Try to read connection properties as a connection check
	_, err := d.KFK.Controller()
	return err
}
