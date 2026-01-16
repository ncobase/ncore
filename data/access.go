package data

import (
	"database/sql"
	"errors"

	"github.com/ncobase/ncore/data/connection"
)

func (d *Data) GetDBManager() *connection.DBManager {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.closed || d.Conn == nil {
		return nil
	}
	return d.Conn.DBM
}

func (d *Data) GetMasterDB() *sql.DB {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.closed || d.Conn == nil {
		return nil
	}
	return d.Conn.DB()
}

func (d *Data) GetSlaveDB() (*sql.DB, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.closed {
		return nil, errors.New("data layer is closed")
	}
	if d.Conn == nil {
		return nil, errors.New("no database connection available")
	}
	return d.Conn.ReadDB()
}

func (d *Data) GetRedis() any {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.closed || d.Conn == nil {
		return nil
	}

	return d.Conn.RC
}

func (d *Data) GetMongoManager() any {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.closed || d.Conn == nil {
		return nil
	}
	return d.Conn.MGM
}

func (d *Data) GetElasticsearch() any {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.closed || d.Conn == nil {
		return nil
	}
	return d.Conn.ES
}

func (d *Data) GetOpenSearch() any {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.closed || d.Conn == nil {
		return nil
	}
	return d.Conn.OS
}

func (d *Data) GetMeilisearch() any {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.closed || d.Conn == nil {
		return nil
	}
	return d.Conn.MS
}

func (d *Data) DB() *sql.DB {
	return d.GetMasterDB()
}

func (d *Data) DBRead() (*sql.DB, error) {
	return d.GetSlaveDB()
}
