package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// GetTx retrieves transaction from context
func GetTx(ctx context.Context) (*sql.Tx, error) {
	tx, ok := ctx.Value(ContextKeyTransaction).(*sql.Tx)
	if !ok {
		return nil, errors.New("transaction not found in context")
	}
	return tx, nil
}

// WithTx wraps function within transaction
func (d *Data) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	start := time.Now()

	d.mu.RLock()
	closed := d.closed
	collector := d.collector
	d.mu.RUnlock()

	if closed {
		err := errors.New("data layer is closed")
		collector.DBTransaction(err)
		return err
	}

	db := d.GetMasterDB()
	if db == nil {
		err := errors.New("database connection is nil")
		collector.DBTransaction(err)
		return err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		duration := time.Since(start)
		collector.DBQuery(duration, err)
		collector.DBTransaction(err)
		return err
	}

	err = fn(context.WithValue(ctx, ContextKeyTransaction, tx))
	duration := time.Since(start)

	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			collector.DBQuery(duration, rbErr)
			collector.DBTransaction(rbErr)
			return fmt.Errorf("tx err: %v, rollback err: %v", err, rbErr)
		}
		collector.DBQuery(duration, err)
		collector.DBTransaction(err)
		return err
	}

	commitErr := tx.Commit()
	collector.DBQuery(duration, commitErr)
	collector.DBTransaction(commitErr)
	return commitErr
}

// WithTxRead wraps function within read-only transaction
func (d *Data) WithTxRead(ctx context.Context, fn func(ctx context.Context) error) error {
	start := time.Now()

	d.mu.RLock()
	closed := d.closed
	collector := d.collector
	d.mu.RUnlock()

	if closed {
		err := errors.New("data layer is closed")
		collector.DBTransaction(err)
		return err
	}

	dbRead, err := d.GetSlaveDB()
	if err != nil {
		collector.DBTransaction(err)
		return err
	}

	tx, err := dbRead.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		duration := time.Since(start)
		collector.DBQuery(duration, err)
		collector.DBTransaction(err)
		return err
	}

	err = fn(context.WithValue(ctx, ContextKeyTransaction, tx))
	duration := time.Since(start)

	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			collector.DBQuery(duration, rbErr)
			collector.DBTransaction(rbErr)
			return fmt.Errorf("tx err: %v, rollback err: %v", err, rbErr)
		}
		collector.DBQuery(duration, err)
		collector.DBTransaction(err)
		return err
	}

	commitErr := tx.Commit()
	collector.DBQuery(duration, commitErr)
	collector.DBTransaction(commitErr)
	return commitErr
}
