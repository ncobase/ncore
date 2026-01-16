package data

import (
	"context"
	"errors"
	"time"
)

func (d *Data) GetMongoDatabase(name string, readOnly bool) (any, error) {
	start := time.Now()
	mgm, ok := any(d.GetMongoManager()).(interface {
		GetDatabase(name string, readOnly bool) (any, error)
	})
	if !ok || mgm == nil {
		err := errors.New("mongodb manager not available")
		d.collector.MongoOperation("get_database", err)
		return nil, err
	}

	db, err := mgm.GetDatabase(name, readOnly)
	duration := time.Since(start)

	d.collector.MongoOperation("get_database", err)
	if duration > time.Second {
		d.collector.MongoOperation("slow_get_database", errors.New("slow_operation"))
	}

	return db, err
}

func (d *Data) GetMongoCollection(dbName, collName string, readOnly bool) (any, error) {
	start := time.Now()
	mgm, ok := any(d.GetMongoManager()).(interface {
		GetCollection(dbName, collName string, readOnly bool) (any, error)
	})
	if !ok || mgm == nil {
		err := errors.New("mongodb manager not available")
		d.collector.MongoOperation("get_collection", err)
		return nil, err
	}

	coll, err := mgm.GetCollection(dbName, collName, readOnly)
	duration := time.Since(start)

	d.collector.MongoOperation("get_collection", err)
	if duration > time.Second {
		d.collector.MongoOperation("slow_get_collection", errors.New("slow_operation"))
	}

	return coll, err
}

func (d *Data) WithMongoTransaction(ctx context.Context, fn func(any) error) error {
	start := time.Now()
	mgm, ok := any(d.GetMongoManager()).(interface {
		WithTransaction(context.Context, func(any) error, ...any) error
	})
	if !ok || mgm == nil {
		err := errors.New("mongodb manager not available")
		d.collector.MongoOperation("transaction", err)
		return err
	}

	err := mgm.WithTransaction(ctx, fn)
	duration := time.Since(start)

	d.collector.MongoOperation("transaction", err)
	if duration > 5*time.Second {
		d.collector.MongoOperation("slow_transaction", errors.New("slow_transaction"))
	}

	return err
}

func (d *Data) MongoHealthCheck(ctx context.Context) error {
	start := time.Now()
	mgm, ok := any(d.GetMongoManager()).(interface {
		Health(context.Context) error
	})
	if !ok || mgm == nil {
		err := errors.New("mongodb manager not available")
		d.collector.HealthCheck("mongodb", false)
		return err
	}

	err := mgm.Health(ctx)
	healthy := err == nil

	duration := time.Since(start)
	d.collector.HealthCheck("mongodb", healthy)
	d.collector.MongoOperation("health_check", err)

	if duration > 3*time.Second {
		d.collector.MongoOperation("slow_health_check", errors.New("slow_health_check"))
	}

	return err
}
