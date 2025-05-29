package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Database database config struct
type Database struct {
	Master   *DBNode   `json:"master" yaml:"master"`
	Slaves   []*DBNode `json:"slaves" yaml:"slaves"`
	Migrate  bool      `json:"migrate" yaml:"migrate"`
	Strategy string    `json:"strategy" yaml:"strategy"`
	MaxRetry int       `json:"max_retry" yaml:"max_retry"`
}

// DBNode represents a single database node configuration
type DBNode struct {
	Driver          string        `json:"driver" yaml:"driver"`
	Source          string        `json:"source" yaml:"source"`
	Logging         bool          `json:"logging" yaml:"logging"`
	MaxIdleConn     int           `json:"max_idle_conn" yaml:"max_idle_conn"`
	MaxOpenConn     int           `json:"max_open_conn" yaml:"max_open_conn"`
	ConnMaxLifeTime time.Duration `json:"conn_max_life_time" yaml:"conn_max_life_time"`
	Weight          int           `json:"weight" yaml:"weight"`
}

// getDatabaseConfig reads database configurations
func getDatabaseConfig(v *viper.Viper) *Database {
	return &Database{
		Master:   getMasterConfig(v),
		Slaves:   getSlaveConfigs(v),
		Migrate:  v.GetBool("data.database.migrate"),
		Strategy: v.GetString("data.database.strategy"),
		MaxRetry: v.GetInt("data.database.max_retry"),
	}
}

// getMasterConfig reads master database configurations
func getMasterConfig(v *viper.Viper) *DBNode {
	return &DBNode{
		Driver:          v.GetString("data.database.master.driver"),
		Source:          v.GetString("data.database.master.source"),
		Logging:         v.GetBool("data.database.master.logging"),
		MaxIdleConn:     v.GetInt("data.database.master.max_idle_conn"),
		MaxOpenConn:     v.GetInt("data.database.master.max_open_conn"),
		ConnMaxLifeTime: v.GetDuration("data.database.master.max_life_time"),
		Weight:          v.GetInt("data.database.master.weight"),
	}
}

// getSlaveConfigs reads slave database configurations
func getSlaveConfigs(v *viper.Viper) []*DBNode {
	var slaves []*DBNode

	slavesConfig := v.Get("data.database.slaves")
	if slavesConfig == nil {
		return slaves
	}

	slavesList, ok := slavesConfig.([]any)
	if !ok {
		return slaves
	}

	slavesCount := len(slavesList)
	for i := 0; i < slavesCount; i++ {
		slave := &DBNode{
			Driver:          v.GetString(fmt.Sprintf("data.database.slaves.%d.driver", i)),
			Source:          v.GetString(fmt.Sprintf("data.database.slaves.%d.source", i)),
			Logging:         v.GetBool(fmt.Sprintf("data.database.slaves.%d.logging", i)),
			MaxIdleConn:     v.GetInt(fmt.Sprintf("data.database.slaves.%d.max_idle_conn", i)),
			MaxOpenConn:     v.GetInt(fmt.Sprintf("data.database.slaves.%d.max_open_conn", i)),
			ConnMaxLifeTime: v.GetDuration(fmt.Sprintf("data.database.slaves.%d.max_life_time", i)),
			Weight:          v.GetInt(fmt.Sprintf("data.database.slaves.%d.weight", i)),
		}
		slaves = append(slaves, slave)
	}
	return slaves
}
