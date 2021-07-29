package conf

import (
	"time"
)

// DBConfig data configs object
type DBConfig struct {
	Dialect         string
	DSN             string
	AutoMigrate     bool
	TablePrefix     string
	MaxIdleConn     int
	MaxOpenConn     int
	ConnMaxLifeTime time.Duration
}

func getDBConfig() *DBConfig {
	return &DBConfig{
		Dialect:         c.GetString("db.dialect"),
		DSN:             c.GetString("db.dsn"),
		AutoMigrate:     c.GetBool("db.auto_migrate"),
		TablePrefix:     c.GetString("db.table_prefix"),
		MaxIdleConn:     c.GetInt("db.max_idle_conn"),
		MaxOpenConn:     c.GetInt("db.max_open_conn"),
		ConnMaxLifeTime: c.GetDuration("db.max_life_time"),
	}
}
