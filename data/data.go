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

	// sqlite3
	_ "github.com/mattn/go-sqlite3"
	// mysql
	_ "github.com/go-sql-driver/mysql"
	// postgres
	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	err error
)

// Data struct to hold all database connections and clients
type Data struct {
	DB *sql.DB
	RC *redis.Client
	MS *meili.Client
	ES *elastic.Client
}

// New creates a new Data instance with all necessary clients
func New(conf *config.Data) (*Data, func(), error) {
	db, err := newDBClient(conf.Database)
	if err != nil {
		return nil, nil, err
	}
	es := newElasticsearch(conf.Elasticsearch)
	d := &Data{
		DB: db,
		RC: newRedis(conf.Redis),
		MS: newMeilisearch(conf.Meilisearch),
		ES: es,
	}

	cleanup := func() {
		log.Printf(context.Background(), "execute data cleanup.")
		if errs := d.Close(); len(errs) > 0 {
			log.Fatalf(context.Background(), "cleanup errors: %v", errs)
		}
	}

	return d, cleanup, nil
}

// newDBClient creates a new database client
func newDBClient(conf *config.Database) (*sql.DB, error) {
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

	return es
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
	if len(errs) > 0 {
		return errs
	}
	return nil
}

// Ping checks the database connection
func (d *Data) Ping(ctx context.Context) error {
	return d.DB.PingContext(ctx)
}
