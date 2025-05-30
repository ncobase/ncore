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
	db := d.GetMasterDB()
	if db == nil {
		err := errors.New("database connection is nil")
		d.collector.DBTransaction(err)
		return err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		duration := time.Since(start)
		d.collector.DBQuery(duration, err)
		d.collector.DBTransaction(err)
		return err
	}

	err = fn(context.WithValue(ctx, ContextKeyTransaction, tx))
	duration := time.Since(start)

	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			d.collector.DBQuery(duration, rbErr)
			d.collector.DBTransaction(rbErr)
			return fmt.Errorf("tx err: %v, rollback err: %v", err, rbErr)
		}
		d.collector.DBQuery(duration, err)
		d.collector.DBTransaction(err)
		return err
	}

	commitErr := tx.Commit()
	d.collector.DBQuery(duration, commitErr)
	d.collector.DBTransaction(commitErr)
	return commitErr
}

// WithTxRead wraps function within read-only transaction
func (d *Data) WithTxRead(ctx context.Context, fn func(ctx context.Context) error) error {
	start := time.Now()
	dbRead, err := d.GetSlaveDB()
	if err != nil {
		d.collector.DBTransaction(err)
		return err
	}

	tx, err := dbRead.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		duration := time.Since(start)
		d.collector.DBQuery(duration, err)
		d.collector.DBTransaction(err)
		return err
	}

	err = fn(context.WithValue(ctx, ContextKeyTransaction, tx))
	duration := time.Since(start)

	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			d.collector.DBQuery(duration, rbErr)
			d.collector.DBTransaction(rbErr)
			return fmt.Errorf("tx err: %v, rollback err: %v", err, rbErr)
		}
		d.collector.DBQuery(duration, err)
		d.collector.DBTransaction(err)
		return err
	}

	commitErr := tx.Commit()
	d.collector.DBQuery(duration, commitErr)
	d.collector.DBTransaction(commitErr)
	return commitErr
}
