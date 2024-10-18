package data

import (
	"context"
	"database/sql"
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
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	sharedInstance *Data
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
			log.Infof(context.Background(), "execute %s data cleanup.", name[0])
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
		log.Infof(context.Background(), "execute %s data cleanup.", name[0])
		if errs := d.Close(); len(errs) > 0 {
			log.Fatalf(context.Background(), "cleanup errors: %v", errs)
		}
	}

	return d, cleanup, nil
}

// WithDB sets the database client in Connections
func WithDB(db *sql.DB) Option {
	return func(d *Data) {
		d.Conn.DB = db
	}
}

// WithRedis sets the Redis client in Connections
func WithRedis(rc *redis.Client) Option {
	return func(d *Data) {
		d.Conn.RC = rc
	}
}

// WithMeilisearch sets the Meilisearch client in Connections
func WithMeilisearch(ms *meili.Client) Option {
	return func(d *Data) {
		d.Conn.MS = ms
	}
}

// WithElasticsearch sets the Elasticsearch client in Connections
func WithElasticsearch(es *elastic.Client) Option {
	return func(d *Data) {
		d.Conn.ES = es
	}
}

// WithMongo sets the MongoDB client in Connections
func WithMongo(mg *mongo.Client) Option {
	return func(d *Data) {
		d.Conn.MG = mg
	}
}

// WithNeo4j sets the Neo4j client in Connections
func WithNeo4j(neo neo4j.DriverWithContext) Option {
	return func(d *Data) {
		d.Conn.Neo = neo
	}
}

// WithRabbitMQ sets the RabbitMQ client in Connections
func WithRabbitMQ(rmq *amqp.Connection) Option {
	return func(d *Data) {
		d.Conn.RMQ = rmq
	}
}

// WithKafka sets the Kafka client in Connections
func WithKafka(kfk *kafka.Conn) Option {
	return func(d *Data) {
		d.Conn.KFK = kfk
	}
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

	if len(errs) > 0 {
		return errs
	}

	return nil
}

func (d *Data) Ping(ctx context.Context) error {
	if d.Conn.DB != nil {
		return d.Conn.DB.PingContext(ctx)
	}
	return nil
}

func (d *Data) GetMongoDatabase(databaseName string) interface{} {
	if d.Conn.MG == nil {
		log.Errorf(context.Background(), "MongoDB client is nil")
		return nil
	}
	return d.Conn.MG.Database(databaseName)
}
