package connection

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"

	"github.com/ncobase/ncore/data/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoManager represents a MongoDB connection manager
type MongoManager struct {
	master   *mongo.Client
	slaves   []*mongo.Client
	strategy MongoLoadBalancer
	mutex    sync.RWMutex
}

// NewMongoManager creates a new MongoDB connection manager
func NewMongoManager(conf *config.MongoDB) (*MongoManager, error) {
	if conf.Master == nil {
		return nil, errors.New("master mongodb configuration is required")
	}

	// connect to master
	master, err := newMongoClient(conf.Master)
	if err != nil {
		return nil, err
	}

	// connect to slaves
	var slaves []*mongo.Client
	for i, slaveCfg := range conf.Slaves {
		slave, err := newMongoClient(slaveCfg)
		if err != nil {
			fmt.Printf("Failed to connect to slave MongoDB %d: %v", i, err)
			continue
		}
		slaves = append(slaves, slave)
	}

	// if no slave available, use master
	if len(slaves) == 0 {
		slaves = append(slaves, master)
	}

	// set up load balancing strategy
	var strategy MongoLoadBalancer
	switch conf.Strategy {
	case "round_robin", "":
		strategy = NewMongoRoundRobinBalancer()
	case "random":
		strategy = &MongoRandomBalancer{}
	case "weight":
		strategy = NewMongoWeightBalancer(conf.Slaves)
	default:
		return nil, ErrInvalidStrategy
	}

	return &MongoManager{
		master:   master,
		slaves:   slaves,
		strategy: strategy,
	}, nil
}

// MongoLoadBalancer MongoDB load balancer
type MongoLoadBalancer interface {
	Next([]*mongo.Client) (*mongo.Client, error)
}

// MongoRoundRobinBalancer round-robin strategy
type MongoRoundRobinBalancer struct {
	current *uint64
}

func NewMongoRoundRobinBalancer() *MongoRoundRobinBalancer {
	var counter uint64
	return &MongoRoundRobinBalancer{
		current: &counter,
	}
}

func (rb *MongoRoundRobinBalancer) Next(slaves []*mongo.Client) (*mongo.Client, error) {
	if len(slaves) == 0 {
		return nil, ErrNoAvailableSlaves
	}

	next := atomic.AddUint64(rb.current, 1) % uint64(len(slaves))
	return slaves[next], nil
}

// MongoRandomBalancer random strategy
type MongoRandomBalancer struct{}

func (rb *MongoRandomBalancer) Next(slaves []*mongo.Client) (*mongo.Client, error) {
	if len(slaves) == 0 {
		return nil, ErrNoAvailableSlaves
	}

	idx := rand.Intn(len(slaves))
	return slaves[idx], nil
}

// MongoWeightBalancer weight strategy
type MongoWeightBalancer struct {
	weights []int
	current *uint64
}

func NewMongoWeightBalancer(nodes []*config.MongoNode) *MongoWeightBalancer {
	weights := make([]int, len(nodes))
	for i, node := range nodes {
		weights[i] = node.Weight
		if weights[i] <= 0 {
			weights[i] = 1
		}
	}

	var counter uint64
	return &MongoWeightBalancer{
		weights: weights,
		current: &counter,
	}
}

func (wb *MongoWeightBalancer) Next(slaves []*mongo.Client) (*mongo.Client, error) {
	if len(slaves) == 0 {
		return nil, ErrNoAvailableSlaves
	}

	totalWeight := 0
	for _, w := range wb.weights {
		totalWeight += w
	}

	next := atomic.AddUint64(wb.current, 1) % uint64(totalWeight)

	var accumulator int
	for i, w := range wb.weights {
		accumulator += w
		if uint64(accumulator) > next {
			return slaves[i], nil
		}
	}

	return slaves[0], nil
}

// Master returns the master client
func (m *MongoManager) Master() *mongo.Client {
	// check if manager is nil
	if m == nil {
		return nil
	}
	return m.master
}

// Slave returns a slave client based on the load balancing strategy
func (m *MongoManager) Slave() (*mongo.Client, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if len(m.slaves) == 0 {
		return m.master, nil
	}

	slave, err := m.strategy.Next(m.slaves)
	if err != nil {
		return m.master, nil
	}

	if err := slave.Ping(context.Background(), nil); err != nil {
		return m.master, nil
	}

	return slave, nil
}

// WithTransaction wraps a function within a transaction
func (m *MongoManager) WithTransaction(ctx context.Context, fn func(mongo.SessionContext) error, opts ...*options.TransactionOptions) error {
	session, err := m.master.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sctx mongo.SessionContext) (any, error) {
		return nil, fn(sctx)
	}, opts...)

	return err
}

// GetDatabase returns a database from master/slave client
func (m *MongoManager) GetDatabase(name string, readOnly bool) (*mongo.Database, error) {
	if readOnly {
		slave, err := m.Slave()
		if err != nil {
			return nil, err
		}
		return slave.Database(name), nil
	}
	return m.master.Database(name), nil
}

// GetCollection returns a collection from master/slave client
// dbName: database name
// collName: collection name
// readOnly: if true, returns collection from slave, otherwise from master
func (m *MongoManager) GetCollection(dbName, collName string, readOnly bool) (*mongo.Collection, error) {
	if readOnly {
		slave, err := m.Slave()
		if err != nil {
			return nil, err
		}
		return slave.Database(dbName).Collection(collName), nil
	}
	return m.master.Database(dbName).Collection(collName), nil
}

// Health checks the health of all MongoDB connections
func (m *MongoManager) Health(ctx context.Context) error {
	// Check master health
	if err := m.master.Ping(ctx, nil); err != nil {
		return fmt.Errorf("master mongodb health check failed: %v", err)
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check and clean up unhealthy slaves
	var healthySlaves []*mongo.Client
	for _, slave := range m.slaves {
		if err := slave.Ping(ctx, nil); err != nil {
			fmt.Printf("Slave mongodb health check failed: %v\n", err)
			continue
		}
		healthySlaves = append(healthySlaves, slave)
	}

	// Update healthy slaves
	m.slaves = healthySlaves

	// If no healthy slaves, use master
	if len(m.slaves) == 0 {
		fmt.Println("No healthy slave mongodb available, using master for reads")
		m.slaves = append(m.slaves, m.master)
	}

	return nil
}

// Close closes all MongoDB connections
func (m *MongoManager) Close(ctx context.Context) error {
	var errs []error

	// Close master
	if err := m.master.Disconnect(ctx); err != nil {
		errs = append(errs, fmt.Errorf("error closing master connection: %v", err))
	}

	// Close slaves
	for i, slave := range m.slaves {
		if slave != m.master { // Avoid double closing the master
			if err := slave.Disconnect(ctx); err != nil {
				errs = append(errs, fmt.Errorf("error closing slave %d connection: %v", i, err))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing mongodb connections: %v", errs)
	}

	return nil
}

// newMongoClient creates a new MongoDB client
func newMongoClient(conf *config.MongoNode) (*mongo.Client, error) {
	if conf == nil || conf.URI == "" {
		return nil, errors.New("mongodb configuration is nil or empty")
	}

	clientOptions := options.Client().ApplyURI(conf.URI)

	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, fmt.Errorf("MongoDB connect error: %v", err)
	}
	if err := client.Ping(context.Background(), nil); err != nil {
		return nil, fmt.Errorf("MongoDB ping error: %v", err)
	}

	return client, nil
}
