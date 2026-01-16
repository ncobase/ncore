package mongodb

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"

	"github.com/ncobase/ncore/data/config"
	"github.com/ncobase/ncore/data/connection"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoManager struct {
	master   *mongo.Client
	slaves   []*mongo.Client
	strategy MongoLoadBalancer
	mutex    sync.RWMutex
}

func NewMongoManager(conf *config.MongoDB) (*MongoManager, error) {
	if conf.Master == nil {
		return nil, errors.New("master mongodb configuration is required")
	}

	master, err := newMongoClient(conf.Master)
	if err != nil {
		return nil, err
	}

	var slaves []*mongo.Client
	for i, slaveCfg := range conf.Slaves {
		slave, err := newMongoClient(slaveCfg)
		if err != nil {
			fmt.Printf("Failed to connect to slave MongoDB %d: %v", i, err)
			continue
		}
		slaves = append(slaves, slave)
	}

	if len(slaves) == 0 {
		slaves = append(slaves, master)
	}

	var strategy MongoLoadBalancer
	switch conf.Strategy {
	case "round_robin", "":
		strategy = NewMongoRoundRobinBalancer()
	case "random":
		strategy = &MongoRandomBalancer{}
	case "weight":
		strategy = NewMongoWeightBalancer(conf.Slaves)
	default:
		return nil, connection.ErrInvalidStrategy
	}

	return &MongoManager{
		master:   master,
		slaves:   slaves,
		strategy: strategy,
	}, nil
}

type MongoLoadBalancer interface {
	Next([]*mongo.Client) (*mongo.Client, error)
}

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
		return nil, connection.ErrNoAvailableSlaves
	}

	next := atomic.AddUint64(rb.current, 1) % uint64(len(slaves))
	return slaves[next], nil
}

type MongoRandomBalancer struct{}

func (rb *MongoRandomBalancer) Next(slaves []*mongo.Client) (*mongo.Client, error) {
	if len(slaves) == 0 {
		return nil, connection.ErrNoAvailableSlaves
	}

	idx := rand.Intn(len(slaves))
	return slaves[idx], nil
}

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
		return nil, connection.ErrNoAvailableSlaves
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

func (m *MongoManager) Master() *mongo.Client {
	if m == nil {
		return nil
	}
	return m.master
}

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

func (m *MongoManager) Health(ctx context.Context) error {
	if err := m.master.Ping(ctx, nil); err != nil {
		return fmt.Errorf("master mongodb health check failed: %v", err)
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	var healthySlaves []*mongo.Client
	for _, slave := range m.slaves {
		if err := slave.Ping(ctx, nil); err != nil {
			fmt.Printf("Slave mongodb health check failed: %v\n", err)
			continue
		}
		healthySlaves = append(healthySlaves, slave)
	}

	m.slaves = healthySlaves

	if len(m.slaves) == 0 {
		fmt.Println("No healthy slave mongodb available, using master for reads")
		m.slaves = append(m.slaves, m.master)
	}

	return nil
}

func (m *MongoManager) Close(ctx context.Context) error {
	var errs []error

	if err := m.master.Disconnect(ctx); err != nil {
		errs = append(errs, fmt.Errorf("error closing master connection: %v", err))
	}

	for i, slave := range m.slaves {
		if slave != m.master {
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
