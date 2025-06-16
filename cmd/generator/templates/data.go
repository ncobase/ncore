package templates

import "fmt"

func DataTemplate(name, extType string) string {
	return fmt.Sprintf(`package data

import (
	"github.com/ncobase/ncore/config"
	"github.com/ncobase/ncore/data"
)

// Data .
type Data struct {
	*data.Data
}

// New creates a new Database Connection.
func New(conf *config.Data, env ...string) (*Data, func(name ...string), error) {
	d, cleanup, err := data.New(conf)
	if err != nil {
		return nil, nil, err
	}

	return &Data{
		Data: d,
	}, cleanup, nil
}

// Close closes all the resources in Data and returns any errors encountered.
func (d *Data) Close() (errs []error) {
	if baseErrs := d.Data.Close(); len(baseErrs) > 0 {
		errs = append(errs, baseErrs...)
	}
	return errs
}

/* Example usage:

// Write operations
db := d.GetDB()
_, err = db.Exec("INSERT INTO users (name) VALUES (?)", "test")

// Read operations
dbRead, err := d.GetSlaveDB()
if err != nil {
    // handle error
}
rows, err := dbRead.Query("SELECT * FROM users")

// Write transaction
err := d.WithTx(ctx, func(ctx context.Context) error {
    tx, err := GetTx(ctx)
    if err != nil {
        return err
    }
    _, err = tx.Exec("INSERT INTO users (name) VALUES (?)", "test")
    return err
})

// Read-only transaction
err := d.WithTxRead(ctx, func(ctx context.Context) error {
    tx, err := GetTx(ctx)
    if err != nil {
        return err
    }
    rows, err := tx.Query("SELECT * FROM users")
    return err
})
*/
`)
}

func DataTemplateWithEnt(name, extType string) string {
	return fmt.Sprintf(`package data

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/ncobase/ncore/config"
	"github.com/ncobase/ncore/data"
	"github.com/ncobase/ncore/logging/logger"
  "{{ .PackagePath }}/data/ent"
  "{{ .PackagePath }}/data/ent/migrate"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/schema"
)

// Data .
type Data struct {
	*data.Data
	EC     *ent.Client // master ent client
	ECRead *ent.Client // slave ent client for read operations
}

// New creates a new Database Connection.
func New(conf *config.Data, env ...string) (*Data, func(name ...string), error) {
	d, cleanup, err := data.New(conf)
	if err != nil {
		return nil, nil, err
	}

	ctx := context.Background()

	// get master connection
	masterDB := d.GetMasterDB()
	if masterDB == nil {
		return nil, cleanup, fmt.Errorf("master database connection is nil")
	}

	// create master ent client
	entClient, err := newEntClient(masterDB, conf.Database.Master, conf.Database.Migrate, env...)
	if err != nil {
		return nil, cleanup, fmt.Errorf("failed to create master ent client: %v", err)
	}

	// get read connection
	var entClientRead *ent.Client
	if readDB, err := d.GetSlaveDB(); err == nil && readDB != nil {
		if readDB != masterDB {
			entClientRead, err = newEntClient(readDB, conf.Database.Master, false, env...) // slave does not support migration
			if err != nil {
				logger.Warnf(ctx, "Failed to create read-only ent client, will use master for reads: %v", err)
				entClientRead = entClient // fallback to master
			}
		} else {
			// Read DB is the same as master (no slaves available)
			entClientRead = entClient
		}
	} else {
		// Failed to get read DB, use master
		entClientRead = entClient
	}

	// Log database configuration
	logDatabaseConfig(ctx, d)

	return &Data{
		Data:   d,
		EC:     entClient,
		ECRead: entClientRead,
	}, cleanup, nil
}

// logDatabaseConfig logs the current database configuration for debugging
func logDatabaseConfig(ctx context.Context, d *data.Data) {
	master, slaves, err := d.GetDatabaseNodes()
	if err != nil {
		logger.Warnf(ctx, "Failed to get database nodes info: %v", err)
		return
	}

	logger.Infof(ctx, "Database configuration - Master: %v, Slaves: %d", master != nil, len(slaves))

	if d.IsReadOnlyMode(ctx) {
		logger.Warnf(ctx, "System is in read-only mode")
	}
}

// newEntClient creates a new ent client.
func newEntClient(db *sql.DB, conf *config.DBNode, enableMigrate bool, env ...string) (*ent.Client, error) {
	client := ent.NewClient(ent.Driver(dialect.DebugWithContext(
		entsql.OpenDB(conf.Driver, db),
		func(ctx context.Context, i ...any) {
			if conf.Logging {
				logger.Infof(ctx, "%v", i)
			}
		},
	)))

	// Enable SQL logging
	if conf.Logging {
		client = client.Debug()
	}

	// Auto migrate (only for master)
	if enableMigrate {
		migrateOpts := []schema.MigrateOption{
			migrate.WithForeignKeys(false),
			// migrate.WithGlobalUniqueID(true),
		}
		// Production does not support drop index and drop column
		if len(env) == 0 || (len(env) > 0 && env[0] != "production") {
			migrateOpts = append(migrateOpts, migrate.WithDropIndex(true), migrate.WithDropColumn(true))
		}
		if err := client.Schema.Create(context.Background(), migrateOpts...); err != nil {
			return nil, fmt.Errorf("failed to migrate database schema: %v", err)
		}
	}

	return client, nil
}

// GetMasterEntClient get master ent client for write operations
func (d *Data) GetMasterEntClient() *ent.Client {
	return d.EC
}

// GetSlaveEntClient get slave ent client for read operations
func (d *Data) GetSlaveEntClient() *ent.Client {
	if d.ECRead != nil {
		return d.ECRead
	}
	return d.EC // Fallback to master
}

// GetEntClientWithFallback returns the appropriate ent client based on operation type
func (d *Data) GetEntClientWithFallback(ctx context.Context, readOnly ...bool) *ent.Client {
	isReadOnly := false
	if len(readOnly) > 0 {
		isReadOnly = readOnly[0]
	}

	if !isReadOnly {
		// For write operations, always use master
		return d.GetMasterEntClient()
	}

	// For read operations, try read client first
	if d.ECRead != nil && d.ECRead != d.EC {
		// We have a separate read client, use it
		return d.ECRead
	}

	// Check if system is in read-only mode
	if d.IsReadOnlyMode(ctx) {
		logger.Warnf(ctx, "System is in read-only mode, using available read connection")
	}

	// Fallback to master
	return d.EC
}

// Close closes all the resources in Data and returns any errors encountered.
func (d *Data) Close() (errs []error) {
	if d.EC != nil {
		if err := d.EC.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close master ent client: %v", err))
		}
	}

	if d.ECRead != nil && d.ECRead != d.EC {
		if err := d.ECRead.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close read ent client: %v", err))
		}
	}

	if baseErrs := d.Data.Close(); len(baseErrs) > 0 {
		errs = append(errs, baseErrs...)
	}

	return errs
}

// GetEntTx retrieves ent transaction from context
func (d *Data) GetEntTx(ctx context.Context) (*ent.Tx, error) {
	tx, ok := ctx.Value("entTx").(*ent.Tx)
	if !ok {
		return nil, fmt.Errorf("ent transaction not found in context")
	}
	return tx, nil
}

// WithEntTx wraps a function within an ent transaction for write operations
func (d *Data) WithEntTx(ctx context.Context, fn func(ctx context.Context, tx *ent.Tx) error) error {
	client := d.GetEntClientWithFallback(ctx)
	if client == nil {
		return fmt.Errorf("ent client is nil")
	}

	tx, err := d.EC.Tx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	err = fn(context.WithValue(ctx, "entTx", tx), tx)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

// WithEntTxRead wraps a function within an ent transaction for read-only operations
func (d *Data) WithEntTxRead(ctx context.Context, fn func(ctx context.Context, tx *ent.Tx) error) error {
	client := d.GetEntClientWithFallback(ctx, true)
	if client == nil {
		return fmt.Errorf("ent read client is nil")
	}

	tx, err := client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin read transaction: %v", err)
	}

	err = fn(context.WithValue(ctx, "entTx", tx), tx)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}



/* Example usage:

// Ent operations
// Write
err := d.WithEntTx(ctx, func(ctx context.Context, tx *ent.Tx) error {
    return tx.User.Create().
        SetName("test").
        Exec(ctx)
})

// Read
err := d.WithEntTxRead(ctx, func(ctx context.Context, tx *ent.Tx) error {
    users, err := tx.User.Query().
        Where(user.NameEQ("test")).
        All(ctx)
    return err
})

// Complex transaction
err := d.WithEntTx(ctx, func(ctx context.Context, tx *ent.Tx) error {
    // Create user
    u, err := tx.User.Create().
        SetName("test").
        Save(ctx)
    if err != nil {
        return err
    }

    // Create relationship config
    _, err = tx.Config.Create().
        SetUser(u).
        SetKey("theme").
        SetValue("dark").
        Save(ctx)
    return err
})
*/
`)
}

func DataTemplateWithGorm(name, extType string) string {
	return fmt.Sprintf(`package data

import (
    "context"
    "database/sql"
    "fmt"
    "github.com/ncobase/ncore/config"
    "github.com/ncobase/ncore/data"
    "github.com/ncobase/ncore/logging/logger"

    "gorm.io/driver/mysql"
    "gorm.io/driver/postgres"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

// Data .
type Data struct {
    *data.Data
    GormClient *gorm.DB    // master gorm client
    GormRead   *gorm.DB    // slave gorm client for read operations
}

// New creates a new Database Connection.
func New(conf *config.Data, env ...string) (*Data, func(name ...string), error) {
    d, cleanup, err := data.New(conf)
    if err != nil {
        return nil, nil, err
    }

    // get master connection
    masterDB := d.DB()
    if masterDB == nil {
        return nil, nil, err
    }

    // create gorm master client
    gormClient, err := newGormClient(masterDB, conf.Database.Master)
    if err != nil {
        return nil, nil, err
    }

    // create gorm read client
    var gormRead *gorm.DB
    if readDB, err := d.GetSlaveDB(); err == nil && readDB != nil {
			if readDB != masterDB {
				gormRead, err = newGormClient(readDB, conf.Database.Master)
				if err != nil {
					logger.Warnf(ctx, "Failed to create read-only gorm client, will use master for reads: %v", err)
					gormRead = gormClient // fallback to master
				}
			} else {
				// Read DB is the same as master (no slaves available)
				gormRead = gormClient
			}
		} else {
			// Failed to get read DB, use master
			gormRead = gormClient
		}

    return &Data{
        Data:       d,
        GormClient: gormClient,
        GormRead:   gormRead,
    }, cleanup, nil
}

// newGormClient creates a new GORM client.
func newGormClient(db *sql.DB, conf *config.DBNode) (*gorm.DB, error) {
    var dialector gorm.Dialector
    switch conf.Driver {
    case "postgres":
        dialector = postgres.New(postgres.Config{
            Conn: db,
        })
    case "mysql":
        dialector = mysql.New(mysql.Config{
            Conn: db,
        })
    case "sqlite3":
        dialector = sqlite.Open(conf.Source)
    default:
        return nil, fmt.Errorf("unsupported database driver: %%s", conf.Driver)
    }

    gormConfig := &gorm.Config{
        Logger: logger.Default.LogMode(logger.Silent),
    }

    if conf.Logging {
        gormConfig.Logger = logger.Default.LogMode(logger.Info)
    }

    return gorm.Open(dialector, gormConfig)
}

// GetGormClient returns the master GORM client for write operations
func (d *Data) GetGormClient() *gorm.DB {
    return d.GormClient
}

// GetGormClientRead returns the slave GORM client for read operations
func (d *Data) GetGormClientRead() *gorm.DB {
    if d.GormRead != nil {
        return d.GormRead
    }
    return d.GormClient // Downgrade, use master
}

// GetGormTx retrieves gorm transaction from context
func GetGormTx(ctx context.Context) (*gorm.DB, error) {
    tx, ok := ctx.Value("gormTx").(*gorm.DB)
    if !ok {
        return nil, fmt.Errorf("gorm transaction not found in context")
    }
    return tx, nil
}

// WithGormTx wraps a function within a gorm transaction
func (d *Data) WithGormTx(ctx context.Context, fn func(ctx context.Context, tx *gorm.DB) error) error {
    if d.GormClient == nil {
        return fmt.Errorf("gorm client is nil")
    }

    return d.GormClient.Transaction(func(tx *gorm.DB) error {
        return fn(context.WithValue(ctx, "gormTx", tx), tx)
    })
}

// WithGormTxRead wraps a function within a transaction using read replica
func (d *Data) WithGormTxRead(ctx context.Context, fn func(ctx context.Context, tx *gorm.DB) error) error {
    client := d.GetGormClientRead()
    if client == nil {
        return fmt.Errorf("gorm read client is nil")
    }

    sqlDB, err := client.DB()
    if err != nil {
        return err
    }

    sqlTx, err := sqlDB.BeginTx(ctx, &sql.TxOptions{
        ReadOnly: true,
    })
    if err != nil {
        return err
    }

    tx := client.Session(&gorm.Session{
        SkipHooks: true,
    }).WithContext(ctx)
    tx.Statement.ConnPool = sqlTx

    err = fn(context.WithValue(ctx, "gormTx", tx), tx)
    if err != nil {
        if rbErr := sqlTx.Rollback(); rbErr != nil {
            return fmt.Errorf("tx err: %%v, rb err: %%v", err, rbErr)
        }
        return err
    }

    return sqlTx.Commit()
}

// Close closes all the resources in Data
func (d *Data) Close() (errs []error) {
    // Close gorm clients
    if d.GormClient != nil {
        if db, err := d.GormClient.DB(); err == nil {
            if err := db.Close(); err != nil {
                errs = append(errs, err)
            }
        }
    }
    if d.GormRead != nil && d.GormRead != d.GormClient {
        if db, err := d.GormRead.DB(); err == nil {
            if err := db.Close(); err != nil {
                errs = append(errs, err)
            }
        }
    }

    // Close base resources
    if baseErrs := d.Data.Close(); len(baseErrs) > 0 {
        errs = append(errs, baseErrs...)
    }

    return errs
}

/* Example usage:

// Write operations with GORM
db := d.GetGormClient()
result := db.Create(&User{Name: "test"})

// Read operations with GORM
dbRead := d.GetGormClientRead()
var users []User
result := dbRead.Where("name = ?", "test").Find(&users)

// Write transaction with GORM
err := d.WithGormTx(ctx, func(ctx context.Context, tx *gorm.DB) error {
    // Create user
    if err := tx.Create(&User{Name: "test"}).Error; err != nil {
        return err
    }

    // Create user config
    if err := tx.Create(&Config{
        UserID: user.ID,
        Key: "theme",
        Value: "dark",
    }).Error; err != nil {
        return err
    }

    return nil
})

// Read-only transaction with GORM
err := d.WithGormTxRead(ctx, func(ctx context.Context, tx *gorm.DB) error {
    var users []User
    if err := tx.Where("status = ?", "active").Find(&users).Error; err != nil {
        return err
    }
    return nil
})
*/
`)
}

func DataTemplateWithMongo(name, extType string) string {
	return fmt.Sprintf(`package data

import (
    "context"
    "fmt"
    "github.com/ncobase/ncore/config"
    "github.com/ncobase/ncore/data"
    "github.com/ncobase/ncore/logging/logger"

    "go.mongodb.org/mongo-driver/mongo"
)

// Data .
type Data struct {
    *data.Data
    MC     *mongo.Client // master mongo client
    MCRead *mongo.Client // slave mongo client for read operations
}

// New creates a new Database Connection.
func New(conf *config.Data, env ...string) (*Data, func(name ...string), error) {
    d, cleanup, err := data.New(conf)
    if err != nil {
        return nil, nil, err
    }

    // get mongo master connection
    mongoMaster := d.Conn.MGM.Master()
    if mongoMaster == nil {
        return nil, nil, fmt.Errorf("mongo master client is nil")
    }

    // get mongo slave connection
    mongoSlave, err := d.Conn.MGM.Slave()
    if err != nil {
        logger.Warnf(context.Background(), "Failed to get read-only mongo client: %%v", err)
    }

    // no slave, use master
    if mongoSlave == nil {
        mongoSlave = mongoMaster
    }

    return &Data{
        Data:   d,
        MC:     mongoMaster,
        MCRead: mongoSlave,
    }, cleanup, nil
}

// GetMongoClient get master mongo client for write operations
func (d *Data) GetMongoClient() *mongo.Client {
    return d.MC
}

// GetMongoClientRead get slave mongo client for read operations
func (d *Data) GetMongoClientRead() *mongo.Client {
    if d.MCRead != nil {
        return d.MCRead
    }
    return d.MC // Downgrade, use master
}

// GetMongoCollection returns a collection from master/slave client
func (d *Data) GetMongoCollection(dbName, collName string, readOnly bool) *mongo.Collection {
    if readOnly {
        return d.MCRead.Database(dbName).Collection(collName)
    }
    return d.MC.Database(dbName).Collection(collName)
}

// GetMongoTx retrieves mongo transaction from context
func GetMongoTx(ctx context.Context) (mongo.SessionContext, error) {
    session, ok := ctx.Value("mongoTx").(mongo.SessionContext)
    if !ok {
        return nil, fmt.Errorf("mongo session not found in context")
    }
    return session, nil
}

// WithMongoTx wraps a function within a mongo transaction
func (d *Data) WithMongoTx(ctx context.Context, fn func(mongo.SessionContext) error) error {
    if d.MC == nil {
        return fmt.Errorf("mongo client is nil")
    }

    session, err := d.MC.StartSession()
    if err != nil {
        return err
    }
    defer session.EndSession(ctx)

    _, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
        return nil, fn(sessCtx)
    })
    return err
}

// WithMongoTxRead wraps a function within a read-only mongo transaction
func (d *Data) WithMongoTxRead(ctx context.Context, fn func(mongo.SessionContext) error) error {
    client := d.GetMongoClientRead()
    if client == nil {
        return fmt.Errorf("mongo read client is nil")
    }

    session, err := client.StartSession()
    if err != nil {
        return err
    }
    defer session.EndSession(ctx)

    // MongoDB does not support read-only transaction, so we downgrade to read-write
    _, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
        return nil, fn(sessCtx)
    })
    return err
}

// Close closes all the resources in Data
func (d *Data) Close() (errs []error) {
	// Close base resources
	if baseErrs := d.Data.Close(); len(baseErrs) > 0 {
		errs = append(errs, baseErrs...)
	}

	return errs
}

/* Example usage:

// Write operations with MongoDB
coll := d.GetMongoCollection("dbName", "users", false)
_, err := coll.InsertOne(ctx, bson.M{"name": "test"})

// Read operations with MongoDB
coll := d.GetMongoCollection("dbName", "users", true)
result := coll.FindOne(ctx, bson.M{"name": "test"})

// Write transaction with MongoDB
err := d.WithMongoTx(ctx, func(sessCtx mongo.SessionContext) error {
    coll := d.GetMongoCollection("dbName", "users", false)

    // Insert user
    _, err := coll.InsertOne(sessCtx, bson.M{
        "name": "test",
        "created_at": time.Now(),
    })
    if err != nil {
        return err
    }

    // Insert config
    configColl := d.GetMongoCollection("dbName", "configs", false)
    _, err = configColl.InsertOne(sessCtx, bson.M{
        "user_id": userID,
        "key": "theme",
        "value": "dark",
    })
    return err
})

// Read operations with MongoDB transaction
err := d.WithMongoTxRead(ctx, func(sessCtx mongo.SessionContext) error {
    coll := d.GetMongoCollection("dbName", "users", true)

    cursor, err := coll.Find(sessCtx, bson.M{"status": "active"})
    if err != nil {
        return err
    }
    defer cursor.Close(sessCtx)

    var users []User
    return cursor.All(sessCtx, &users)
})
*/
`)
}
