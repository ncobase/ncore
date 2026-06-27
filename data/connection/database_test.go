package connection

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"sync"
	"testing"

	"github.com/ncobase/ncore/data/config"
)

var registerFakeSQLDriver sync.Once

type fakeSQLDriver struct{}

func (fakeSQLDriver) Open(string) (driver.Conn, error) {
	return fakeSQLConn{}, nil
}

type fakeSQLConn struct{}

func (fakeSQLConn) Prepare(string) (driver.Stmt, error) {
	return fakeSQLStmt{}, nil
}

func (fakeSQLConn) Close() error {
	return nil
}

func (fakeSQLConn) Begin() (driver.Tx, error) {
	return fakeSQLTx{}, nil
}

func (fakeSQLConn) Ping(context.Context) error {
	return nil
}

type fakeSQLStmt struct{}

func (fakeSQLStmt) Close() error {
	return nil
}

func (fakeSQLStmt) NumInput() int {
	return 0
}

func (fakeSQLStmt) Exec([]driver.Value) (driver.Result, error) {
	return driver.RowsAffected(0), nil
}

func (fakeSQLStmt) Query([]driver.Value) (driver.Rows, error) {
	return fakeSQLRows{}, nil
}

type fakeSQLRows struct{}

func (fakeSQLRows) Columns() []string {
	return nil
}

func (fakeSQLRows) Close() error {
	return nil
}

func (fakeSQLRows) Next([]driver.Value) error {
	return io.EOF
}

type fakeSQLTx struct{}

func (fakeSQLTx) Commit() error {
	return nil
}

func (fakeSQLTx) Rollback() error {
	return nil
}

func openFakeDB(t *testing.T) *sql.DB {
	t.Helper()

	registerFakeSQLDriver.Do(func() {
		sql.Register("ncore_fake_sql", fakeSQLDriver{})
	})

	db, err := sql.Open("ncore_fake_sql", "")
	if err != nil {
		t.Fatalf("open fake db: %v", err)
	}
	return db
}

func TestSlaveContextUsesPrimaryWhenNoReplicasConfigured(t *testing.T) {
	db := openFakeDB(t)
	defer db.Close()

	manager := &DBManager{
		master:                 db,
		strategy:               NewRoundRobinBalancer(),
		readReplicasConfigured: false,
	}

	readDB, err := manager.SlaveContext(context.Background())
	if err != nil {
		t.Fatalf("read connection: %v", err)
	}
	if readDB != db {
		t.Fatal("expected primary database for single-node read routing")
	}
}

func TestSlaveContextErrorsWhenConfiguredReplicasUnavailable(t *testing.T) {
	db := openFakeDB(t)
	defer db.Close()

	manager := &DBManager{
		master:                 db,
		strategy:               NewRoundRobinBalancer(),
		readReplicasConfigured: true,
	}

	if _, err := manager.SlaveContext(context.Background()); !errors.Is(err, ErrNoAvailableSlaves) {
		t.Fatalf("expected ErrNoAvailableSlaves, got %v", err)
	}
}

func TestWeightBalancerHandlesWeightReplicaMismatch(t *testing.T) {
	db := openFakeDB(t)
	defer db.Close()

	balancer := NewWeightBalancer([]*config.DBNode{{Weight: 5}, {Weight: 1}})
	selected, err := balancer.Next([]*sql.DB{db})
	if err != nil {
		t.Fatalf("select weighted replica: %v", err)
	}
	if selected != db {
		t.Fatal("expected available replica to be selected")
	}
}
