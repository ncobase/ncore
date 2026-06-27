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

// Space configures Space-level database isolation (multi-tenancy).
//
// When Strategy is set, each Space operates in its own database namespace.
// Projects that do not need multi-tenancy can omit this section entirely.
//
// YAML example (PostgreSQL schema-per-space):
//
//	data:
//	  space:
//	    strategy: schema
//	    global_schema: public
//	    space_prefix: space_
//	    max_cached_pools: 200
//
// YAML example (MySQL database-per-space):
//
//	data:
//	  space:
//	    strategy: database
//	    space_prefix: space_
//	    max_cached_pools: 100
type Space struct {
	// Strategy selects the isolation model.
	//   "schema"   – PostgreSQL schema-per-space via SET LOCAL search_path (recommended for PG)
	//   "database" – MySQL / PostgreSQL database-per-space with per-space connection pools
	//   ""         – disabled (default; all Space-aware methods fall back to ordinary transactions)
	Strategy string `json:"strategy" yaml:"strategy"`

	// GlobalSchema is the PostgreSQL schema that holds shared (cross-space) tables
	// such as users, auth tokens, and the spaces registry itself.
	// Only relevant when Strategy is "schema". Defaults to "public".
	GlobalSchema string `json:"global_schema" yaml:"global_schema"`

	// SpacePrefix is prepended to the normalized space identifier to form the
	// schema or database name (e.g. "space_" → schema "space_acme").
	// Defaults to "space_".
	SpacePrefix string `json:"space_prefix" yaml:"space_prefix"`

	// MaxCachedPools is the maximum number of per-space connection pools kept
	// open simultaneously. Excess pools are evicted using an LRU policy.
	// Only relevant when Strategy is "database". Defaults to 100.
	MaxCachedPools int `json:"max_cached_pools" yaml:"max_cached_pools"`
}

// IsEnabled reports whether Space isolation is active.
func (s *Space) IsEnabled() bool {
	return s != nil && s.Strategy != ""
}

// EffectiveGlobalSchema returns the global schema name, defaulting to "public".
func (s *Space) EffectiveGlobalSchema() string {
	if s == nil || s.GlobalSchema == "" {
		return "public"
	}
	return s.GlobalSchema
}

// EffectiveSpacePrefix returns the space prefix, defaulting to "space_".
func (s *Space) EffectiveSpacePrefix() string {
	if s == nil || s.SpacePrefix == "" {
		return "space_"
	}
	return s.SpacePrefix
}

// EffectiveMaxCachedPools returns the max cached pools count, defaulting to 100.
func (s *Space) EffectiveMaxCachedPools() int {
	if s == nil || s.MaxCachedPools <= 0 {
		return 100
	}
	return s.MaxCachedPools
}

// getSpaceConfig reads Space isolation configuration from Viper.
// Returns nil (not an empty struct) when the strategy key is absent, allowing
// callers to use a nil check instead of inspecting the Strategy field.
func getSpaceConfig(v *viper.Viper) *Space {
	strategy := v.GetString("data.space.strategy")
	if strategy == "" {
		return nil // Space isolation disabled; keep zero overhead
	}
	return &Space{
		Strategy:       strategy,
		GlobalSchema:   v.GetString("data.space.global_schema"),
		SpacePrefix:    v.GetString("data.space.space_prefix"),
		MaxCachedPools: v.GetInt("data.space.max_cached_pools"),
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
