package connection

import (
	"context"
	"database/sql"
	"ncobase/common/config"
	"ncobase/common/elastic"
	"ncobase/common/log"
	"ncobase/common/meili"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/mongo"
)

// Connections struct to hold all database connections and clients
type Connections struct {
	DB  *sql.DB
	RC  *redis.Client
	MS  *meili.Client
	ES  *elastic.Client
	MG  *mongo.Client
	Neo neo4j.DriverWithContext
	RMQ *amqp.Connection
	KFK *kafka.Conn
}

// New creates a new Connections
func New(conf *config.Data) (*Connections, error) {
	c := &Connections{}
	var err error

	if conf.Database != nil && conf.Database.Source != "" {
		c.DB, err = newDBClient(conf.Database)
		if err != nil {
			return nil, err
		}
	}

	if conf.Redis != nil && conf.Redis.Addr != "" {
		c.RC, _ = newRedisClient(conf.Redis)
	}

	if conf.Meilisearch != nil && conf.Meilisearch.Host != "" {
		c.MS, _ = newMeilisearchClient(conf.Meilisearch)
	}

	if conf.Elasticsearch != nil && len(conf.Elasticsearch.Addresses) > 0 {
		c.ES, _ = newElasticsearchClient(conf.Elasticsearch)
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
	// Close Redis client if not already closed
	if d.RC != nil {
		if err := d.RC.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	// Close SQL database client if not already closed
	if d.DB != nil {
		if err := d.DB.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	// Disconnect MongoDB client if not already disconnected
	if d.MG != nil {
		if err := d.MG.Disconnect(context.Background()); err != nil {
			errs = append(errs, err)
		}
	}

	// Close Neo4j client if not already closed
	if d.Neo != nil {
		if err := d.Neo.Close(context.Background()); err != nil {
			errs = append(errs, err)
		}
	}

	// Close RabbitMQ client if not already closed
	if d.RMQ != nil {
		if err := d.RMQ.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	// Close Kafka client if not already closed
	if d.KFK != nil {
		if err := d.KFK.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

// Ping checks the database connection
func (d *Connections) Ping(ctx context.Context) error {
	if d.DB != nil {
		return d.DB.PingContext(ctx)
	}
	return nil
}

// GetMongoDatabase retrieves a specific MongoDB database
func (d *Connections) GetMongoDatabase(databaseName string) *mongo.Database {
	if d.MG == nil {
		log.Errorf(context.Background(), "MongoDB client is nil")
		return nil
	}
	return d.MG.Database(databaseName)
}
