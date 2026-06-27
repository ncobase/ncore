package connection

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ncobase/ncore/data/config"
)

// driverRegistry provides access to the driver registration without import cycle.
// This is set by the parent data package during initialization.
var driverRegistry DriverRegistry

// DriverRegistry defines the driver lookups needed from the parent data package.
type DriverRegistry interface {
	GetDatabaseDriver(name string) (DatabaseDriver, error)
	GetCacheDriver(name string) (CacheDriver, error)
	GetSearchDriver(name string) (SearchDriver, error)
	GetMessageDriver(name string) (MessageDriver, error)
}

// DatabaseDriver mirrors the interface from data package to avoid import cycle.
type DatabaseDriver interface {
	Name() string
	Connect(ctx context.Context, cfg any) (any, error)
	Close(conn any) error
	Ping(ctx context.Context, conn any) error
}

type CacheDriver interface {
	Name() string
	Connect(ctx context.Context, cfg any) (any, error)
	Close(conn any) error
	Ping(ctx context.Context, conn any) error
}

type SearchDriver interface {
	Name() string
	Connect(ctx context.Context, cfg any) (any, error)
	Close(conn any) error
}

type MessageDriver interface {
	Name() string
	Connect(ctx context.Context, cfg any) (any, error)
	Close(conn any) error
}

// SetDriverRegistry is called by the data package to inject the registry.
func SetDriverRegistry(registry DriverRegistry) {
	driverRegistry = registry
}

var (
	ErrNoAvailableSlaves = errors.New("no available read replica databases")
	ErrInvalidStrategy   = errors.New("invalid load balance strategy")
)

// DBManager manages database connections for read-write splitting
type DBManager struct {
	master                 *sql.DB
	slaves                 []*sql.DB
	readReplicasConfigured bool
	strategy               LoadBalancer
	mutex                  sync.RWMutex
	maxRetry               int
	currentIdx             uint64 // for round robin
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
	if len(wb.weights) != len(slaves) {
		wb.weights = make([]int, len(slaves))
		for i := range wb.weights {
			wb.weights[i] = 1
		}
	}

	// calculate total weight, default 1
	totalWeight := 0
	for _, w := range wb.weights {
		if w <= 0 {
			w = 1
		}
		totalWeight += w
	}
	if totalWeight <= 0 {
		return slaves[0], nil
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
	var connectedSlaveConfigs []*config.DBNode
	for _, slaveCfg := range conf.Slaves {
		slave, err := newDBClient(slaveCfg)
		if err != nil {
			fmt.Printf("Failed to connect to slave DB: %v", err)
			continue
		}
		slaves = append(slaves, slave)
		connectedSlaveConfigs = append(connectedSlaveConfigs, slaveCfg)
	}

	// set up load balancing strategy
	var strategy LoadBalancer
	switch conf.Strategy {
	case "round_robin", "":
		strategy = NewRoundRobinBalancer()
	case "random":
		strategy = &RandomBalancer{}
	case "weight":
		strategy = NewWeightBalancer(connectedSlaveConfigs)
	default:
		return nil, ErrInvalidStrategy
	}

	return &DBManager{
		master:                 master,
		slaves:                 slaves,
		readReplicasConfigured: len(conf.Slaves) > 0,
		strategy:               strategy,
		maxRetry:               conf.MaxRetry,
	}, nil
}

func newDBClient(conf *config.DBNode) (*sql.DB, error) {
	if driverRegistry == nil {
		return nil, fmt.Errorf("driver registry not initialized, ensure drivers are imported")
	}

	driver, err := driverRegistry.GetDatabaseDriver(conf.Driver)
	if err != nil {
		return nil, fmt.Errorf("failed to get database driver: %w", err)
	}

	conn, err := driver.Connect(context.Background(), conf)
	if err != nil {
		return nil, fmt.Errorf("failed to connect using %s driver: %w", conf.Driver, err)
	}

	db, ok := conn.(*sql.DB)
	if !ok {
		return nil, fmt.Errorf("driver %s returned invalid connection type, expected *sql.DB", conf.Driver)
	}

	return db, nil
}

// Master returns the master database connection
func (dm *DBManager) Master() *sql.DB {
	return dm.master
}

// Slave returns a slave database connection based on the load balancing strategy
func (dm *DBManager) Slave() (*sql.DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return dm.SlaveContext(ctx)
}

// SlaveContext returns a read database connection. If no read replicas are
// configured, the primary connection is returned. If replicas are configured but
// unavailable, an error is returned instead of silently routing reads to primary.
func (dm *DBManager) SlaveContext(ctx context.Context) (*sql.DB, error) {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	if len(dm.slaves) == 0 {
		if dm.readReplicasConfigured {
			return nil, ErrNoAvailableSlaves
		}
		if err := dm.master.PingContext(ctx); err != nil {
			return nil, fmt.Errorf("primary database health check failed: %w", err)
		}
		return dm.master, nil
	}

	var lastErr error
	for i := 0; i <= dm.maxRetry; i++ {
		slave, err := dm.strategy.Next(dm.slaves)
		if err != nil {
			lastErr = err
			continue
		}

		// Test the slave database connection
		if err := slave.PingContext(ctx); err != nil {
			lastErr = err
			continue
		}

		return slave, nil
	}

	// all retry attempts failed, return the last error
	if lastErr == nil {
		lastErr = ErrNoAvailableSlaves
	}
	return nil, fmt.Errorf("all read replica attempts failed: %w", lastErr)
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

	var replicaErrors []error
	for i, slave := range dm.slaves {
		if err := slave.PingContext(ctx); err != nil {
			replicaErrors = append(replicaErrors, fmt.Errorf("read replica %d health check failed: %w", i, err))
		}
	}
	if dm.readReplicasConfigured && len(dm.slaves) == 0 {
		replicaErrors = append(replicaErrors, ErrNoAvailableSlaves)
	}
	if len(replicaErrors) > 0 {
		return fmt.Errorf("read replica health check failed: %v", replicaErrors)
	}

	return nil
}
