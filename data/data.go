package data

import (
	"context"
	"database/sql"
	"io"
	"ncobase/common/config"
	"ncobase/common/elastic"
	"ncobase/common/log"
	"ncobase/common/meili"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	// sqlite3
	_ "github.com/mattn/go-sqlite3"
	// mysql
	_ "github.com/go-sql-driver/mysql"
	// postgres
	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	sharedInstance *Data
	err            error
)

// Data struct to hold all database connections and clients
type Data struct {
	DB *sql.DB
	RC *redis.Client
	MS *meili.Client
	ES *elastic.Client
	MG *mongo.Client
}

// Option function type for configuring Data
type Option func(*Data)

// New creates a new Data instance with the provided options
func New(conf *config.Data, createNewInstance ...bool) (*Data, func(name ...string), error) {
	var createNew bool
	if len(createNewInstance) > 0 {
		createNew = createNewInstance[0]
	}

	if !createNew && sharedInstance != nil {
		cleanup := func(name ...string) {
			log.Printf(context.Background(), "execute %s data cleanup.", name[0])
			if errs := sharedInstance.Close(); len(errs) > 0 {
				log.Fatalf(context.Background(), "cleanup errors: %v", errs)
			}
		}
		return sharedInstance, cleanup, nil
	}

	d := &Data{}

	var opts []Option
	if conf.Database != nil && conf.Database.Source != "" {
		db, err := newDBClient(conf.Database)
		if err != nil {
			return nil, nil, err
		}
		opts = append(opts, WithDB(db))
	}

	if conf.Redis != nil && conf.Redis.Addr != "" {
		rc := newRedis(conf.Redis)
		opts = append(opts, WithRedis(rc))
	}

	if conf.Meilisearch != nil && conf.Meilisearch.Host != "" {
		ms := newMeilisearch(conf.Meilisearch)
		opts = append(opts, WithMeilisearch(ms))
	}

	if conf.Elasticsearch != nil && len(conf.Elasticsearch.Addresses) > 0 {
		es := newElasticsearch(conf.Elasticsearch)
		opts = append(opts, WithElasticsearch(es))
	}

	if conf.MongoDB != nil && conf.MongoDB.URI != "" {
		mg, err := newMongoClient(conf.MongoDB)
		if err != nil {
			return nil, nil, err
		}
		opts = append(opts, WithMongo(mg))
	}

	for _, option := range opts {
		option(d)
	}

	if !createNew {
		sharedInstance = d
	}

	cleanup := func(name ...string) {
		log.Printf(context.Background(), "execute %s data cleanup.", name[0])
		if errs := d.Close(); len(errs) > 0 {
			log.Fatalf(context.Background(), "cleanup errors: %v", errs)
		}
	}

	return d, cleanup, nil
}

// WithDB sets the database client in Data
func WithDB(db *sql.DB) Option {
	return func(d *Data) {
		d.DB = db
	}
}

// WithRedis sets the Redis client in Data
func WithRedis(rc *redis.Client) Option {
	return func(d *Data) {
		d.RC = rc
	}
}

// WithMeilisearch sets the Meilisearch client in Data
func WithMeilisearch(ms *meili.Client) Option {
	return func(d *Data) {
		d.MS = ms
	}
}

// WithElasticsearch sets the Elasticsearch client in Data
func WithElasticsearch(es *elastic.Client) Option {
	return func(d *Data) {
		d.ES = es
	}
}

// WithMongo sets the MongoDB client in Data
func WithMongo(mg *mongo.Client) Option {
	return func(d *Data) {
		d.MG = mg
	}
}

// newDBClient creates a new database client
func newDBClient(conf *config.Database) (*sql.DB, error) {
	if conf == nil {
		log.Fatalf(context.Background(), "database configuration is nil")
		return nil, err
	}

	var db *sql.DB
	switch conf.Driver {
	case "postgres":
		db, err = sql.Open("pgx", conf.Source)
	case "mysql":
		db, err = sql.Open("mysql", conf.Source)
	case "sqlite3":
		db, err = sql.Open("sqlite3", conf.Source)
	default:
		log.Fatalf(context.Background(), "dialect %v not supported", conf.Driver)
		return nil, err
	}

	if err != nil {
		log.Fatalf(context.Background(), "failed to open database: %v", err)
		return nil, err
	}

	db.SetMaxIdleConns(conf.MaxIdleConn)
	db.SetMaxOpenConns(conf.MaxOpenConn)
	db.SetConnMaxLifetime(conf.ConnMaxLifeTime)

	log.Infof(context.Background(), "database %v connected", conf.Driver)

	return db, nil
}

// newRedis creates a new Redis client
func newRedis(conf *config.Redis) *redis.Client {
	if conf == nil || conf.Addr == "" {
		log.Printf(context.Background(), "redis configuration is nil or empty")
		return nil
	}

	rc := redis.NewClient(&redis.Options{
		Addr:         conf.Addr,
		Username:     conf.Username,
		Password:     conf.Password,
		DB:           conf.Db,
		ReadTimeout:  conf.ReadTimeout,
		WriteTimeout: conf.WriteTimeout,
		DialTimeout:  conf.DialTimeout,
		PoolSize:     10,
	})

	timeout, cancelFunc := context.WithTimeout(context.Background(), conf.DialTimeout)
	defer cancelFunc()
	if err := rc.Ping(timeout).Err(); err != nil {
		log.Errorf(context.Background(), "redis connect error: %v", err)
	}

	log.Infof(context.Background(), "redis connected")

	return rc
}

// newMeilisearch creates a new Meilisearch client
func newMeilisearch(conf *config.Meilisearch) *meili.Client {
	if conf == nil || conf.Host == "" {
		log.Printf(context.Background(), "Meilisearch configuration is nil or empty")
		return nil
	}

	ms := meili.NewMeilisearch(conf.Host, conf.APIKey)

	if _, err := ms.GetClient().Health(); err != nil {
		log.Errorf(context.Background(), "Meilisearch connect error: %v", err)
		return nil
	}

	log.Infof(context.Background(), "Meilisearch connected")

	return ms
}

// newElasticsearch creates a new Elasticsearch client
func newElasticsearch(conf *config.Elasticsearch) *elastic.Client {
	if conf == nil || len(conf.Addresses) == 0 {
		log.Printf(context.Background(), "Elasticsearch configuration is nil or empty")
		return nil
	}

	es, err := elastic.NewClient(conf.Addresses, conf.Username, conf.Password)
	if err != nil {
		log.Errorf(context.Background(), "Elasticsearch client creation error: %v", err)
		return nil
	}

	res, err := es.GetClient().Info()
	if err != nil {
		log.Errorf(context.Background(), "Elasticsearch connect error: %v", err)
		return nil
	}
	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			log.Errorf(context.Background(), "Elasticsearch response body close error: %v", err)
		}
	}(res.Body)

	if res.IsError() {
		log.Errorf(context.Background(), "Elasticsearch info error: %s", res.Status())
		return nil
	}

	log.Infof(context.Background(), "Elasticsearch connected")

	return es
}

// newMongoClient creates a new MongoDB client
func newMongoClient(conf *config.MongoDB) (*mongo.Client, error) {
	if conf == nil || conf.URI == "" {
		log.Printf(context.Background(), "MongoDB configuration is nil or empty")
		return nil, nil
	}

	clientOptions := options.Client().ApplyURI(conf.URI)
	if conf.Username != "" && conf.Password != "" {
		clientOptions.SetAuth(options.Credential{
			Username: conf.Username,
			Password: conf.Password,
		})
	}

	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Errorf(context.Background(), "MongoDB connect error: %v", err)
		return nil, err
	}
	if err := client.Ping(context.Background(), nil); err != nil {
		log.Errorf(context.Background(), "MongoDB ping error: %v", err)
		return nil, err
	}

	log.Infof(context.Background(), "MongoDB connected")

	return client, nil
}

// Close closes all resources in Data and returns any errors encountered
func (d *Data) Close() (errs []error) {
	if d.RC != nil {
		if err := d.RC.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if d.DB != nil {
		if err := d.DB.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if d.MG != nil {
		if err := d.MG.Disconnect(context.Background()); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

// Ping checks the database connection
func (d *Data) Ping(ctx context.Context) error {
	if d.DB != nil {
		return d.DB.PingContext(ctx)
	}
	return nil
}

// GetMongoDatabase retrieves a specific MongoDB database
func (d *Data) GetMongoDatabase(databaseName string) *mongo.Database {
	if d.MG == nil {
		log.Errorf(context.Background(), "MongoDB client is nil")
		return nil
	}
	return d.MG.Database(databaseName)
}
