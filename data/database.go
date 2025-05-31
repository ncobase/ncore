package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// GetDatabaseNodes returns information about all database nodes (master and slaves)
func (d *Data) GetDatabaseNodes() (master *sql.DB, slaves []*sql.DB, err error) {
	if d.Conn == nil || d.Conn.DBM == nil {
		return nil, nil, errors.New("no database manager available")
	}

	master = d.Conn.DBM.Master()

	// Get all slave connections by repeatedly calling Slave() method
	slavesMap := make(map[*sql.DB]bool)
	for i := 0; i < 10; i++ { // Try up to 10 times to get different slaves
		slave, err := d.Conn.DBM.Slave()
		if err != nil {
			break
		}
		if slave != master { // Only add if it's not the master
			slavesMap[slave] = true
		}
	}

	for slave := range slavesMap {
		slaves = append(slaves, slave)
	}

	return master, slaves, nil
}

// IsReadOnlyMode checks if the system is in read-only mode (only slaves available)
func (d *Data) IsReadOnlyMode(ctx context.Context) bool {
	if d.Conn == nil || d.Conn.DBM == nil {
		return false
	}

	master := d.Conn.DBM.Master()
	if master == nil {
		return true
	}

	// Check if master is healthy
	if err := master.PingContext(ctx); err != nil {
		return true
	}

	return false
}

// Ping checks all database connections
func (d *Data) Ping(ctx context.Context) error {
	start := time.Now()
	var err error

	if d.Conn != nil {
		err = d.Conn.Ping(ctx)
	} else {
		err = errors.New("no connection manager available")
	}

	duration := time.Since(start)
	d.collector.DBQuery(duration, err)
	d.collector.HealthCheck("database", err == nil)

	return err
}
