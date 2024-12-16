package connection

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"ncobase/common/data/config"
	"sync"
	"sync/atomic"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/mattn/go-sqlite3"
)

var (
	ErrNoAvailableSlaves = errors.New("no available slave databases")
	ErrInvalidStrategy   = errors.New("invalid load balance strategy")
)

// DBManager manages database connections for read-write splitting
type DBManager struct {
	master     *sql.DB
	slaves     []*sql.DB
	strategy   LoadBalancer
	mutex      sync.RWMutex
	maxRetry   int
	currentIdx uint64 // for round robin
}

// LoadBalancer LoadBalancer interface
type LoadBalancer interface {
	Next([]*sql.DB) (*sql.DB, error)
}

// RoundRobinBalancer Implement polling strategy
type RoundRobinBalancer struct {
	current *uint64
}

// NewRoundRobinBalancer Create new RoundRobinBalancer
func NewRoundRobinBalancer() *RoundRobinBalancer {
	var counter uint64
	return &RoundRobinBalancer{
		current: &counter,
	}
}

func (rb *RoundRobinBalancer) Next(slaves []*sql.DB) (*sql.DB, error) {
	if len(slaves) == 0 {
		return nil, ErrNoAvailableSlaves
	}

	next := atomic.AddUint64(rb.current, 1) % uint64(len(slaves))
	return slaves[next], nil
}

// RandomBalancer Implement random strategy
type RandomBalancer struct{}

func (rb *RandomBalancer) Next(slaves []*sql.DB) (*sql.DB, error) {
	if len(slaves) == 0 {
		return nil, ErrNoAvailableSlaves
	}

	idx := rand.Intn(len(slaves))
	return slaves[idx], nil
}

// WeightBalancer Implement weight strategy
type WeightBalancer struct {
	weights []int
	current *uint64
}

func NewWeightBalancer(slaves []*config.DBNode) *WeightBalancer {
	weights := make([]int, len(slaves))
	for i, slave := range slaves {
		weights[i] = slave.Weight
		if weights[i] <= 0 {
			weights[i] = 1 // default
		}
	}

	var counter uint64
	return &WeightBalancer{
		weights: weights,
		current: &counter,
	}
}

func (wb *WeightBalancer) Next(slaves []*sql.DB) (*sql.DB, error) {
	if len(slaves) == 0 {
		return nil, ErrNoAvailableSlaves
	}

	// calculate total weight, default 1
	totalWeight := 0
	for _, w := range wb.weights {
		totalWeight += w
	}

	// select weights
	next := atomic.AddUint64(wb.current, 1) % uint64(totalWeight)

	// find corresponding slave
	var accumulator int
	for i, w := range wb.weights {
		accumulator += w
		if uint64(accumulator) > next {
			return slaves[i], nil
		}
	}

	// should not reach here, but just in case, return the first slave
	return slaves[0], nil
}

// NewDBManager creates a new database manager with read-write splitting
func NewDBManager(conf *config.Database) (*DBManager, error) {
	if conf.Master == nil {
		return nil, fmt.Errorf("master database configuration is required")
	}
	// Initialize master database connection
	master, err := newDBClient(conf.Master)
	if err != nil {
		return nil, err
	}

	// Initialize slave database connections
	var slaves []*sql.DB
	for _, slaveCfg := range conf.Slaves {
		slave, err := newDBClient(slaveCfg)
		if err != nil {
			fmt.Printf("Failed to connect to slave DB: %v", err)
			continue
		}
		slaves = append(slaves, slave)
	}

	// if no slave database is available, use master
	if len(slaves) == 0 {
		slaves = append(slaves, master)
	}

	// set up load balancing strategy
	var strategy LoadBalancer
	switch conf.Strategy {
	case "round_robin", "":
		strategy = NewRoundRobinBalancer()
	case "random":
		strategy = &RandomBalancer{}
	case "weight":
		strategy = NewWeightBalancer(conf.Slaves)
	default:
		return nil, ErrInvalidStrategy
	}

	return &DBManager{
		master:   master,
		slaves:   slaves,
		strategy: strategy,
		maxRetry: conf.MaxRetry,
	}, nil
}

// newDBClient creates a new database client
func newDBClient(conf *config.DBNode) (*sql.DB, error) {
	var db *sql.DB
	var err error

	switch conf.Driver {
	case "postgres":
		db, err = sql.Open("pgx", conf.Source)
	case "mysql":
		db, err = sql.Open("mysql", conf.Source)
	case "sqlite3":
		db, err = sql.Open("sqlite3", conf.Source)
	default:
		return nil, fmt.Errorf("dialect %v not supported", conf.Driver)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	db.SetMaxIdleConns(conf.MaxIdleConn)
	db.SetMaxOpenConns(conf.MaxOpenConn)
	db.SetConnMaxLifetime(conf.ConnMaxLifeTime)

	return db, nil
}

// Master returns the master database connection
func (dm *DBManager) Master() *sql.DB {
	return dm.master
}

// Slave returns a slave database connection based on the load balancing strategy
func (dm *DBManager) Slave() (*sql.DB, error) {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	var lastErr error
	for i := 0; i <= dm.maxRetry; i++ {
		slave, err := dm.strategy.Next(dm.slaves)
		if err != nil {
			lastErr = err
			continue
		}

		// Test the slave database connection
		if err := slave.PingContext(context.Background()); err != nil {
			lastErr = err
			continue
		}

		return slave, nil
	}

	// all retry attempts failed, return the last error
	return nil, fmt.Errorf("all retry attempts failed: %v", lastErr)
}

// Close closes all database connections
func (dm *DBManager) Close() error {
	var errs []error

	// Close master database
	if err := dm.master.Close(); err != nil {
		errs = append(errs, fmt.Errorf("error closing master connection: %v", err))
	}

	// Close slaves database
	for i, slave := range dm.slaves {
		if slave != dm.master { // Avoid double closing the master connection
			if err := slave.Close(); err != nil {
				errs = append(errs, fmt.Errorf("error closing slave %d connection: %v", i, err))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing database connections: %v", errs)
	}
	return nil
}

// Health checks the health of all database connections
func (dm *DBManager) Health(ctx context.Context) error {
	// Check health of master database
	if err := dm.master.PingContext(ctx); err != nil {
		return fmt.Errorf("master database health check failed: %v", err)
	}

	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	// Check health of slave databases, and update the list of healthy slaves
	var healthySlaves []*sql.DB
	for _, slave := range dm.slaves {
		if err := slave.PingContext(ctx); err != nil {
			fmt.Printf("Slave database health check failed: %v\n", err)
			continue
		}
		healthySlaves = append(healthySlaves, slave)
	}

	// Update the list of healthy slaves
	dm.slaves = healthySlaves

	// if no slave database is available, use master
	if len(dm.slaves) == 0 {
		dm.slaves = append(dm.slaves, dm.master)
	}

	return nil
}
