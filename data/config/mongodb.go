package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// MongoDB mongodb config struct
type MongoDB struct {
	Master   *MongoNode   `json:"master"`
	Slaves   []*MongoNode `json:"slaves"`
	Strategy string       `json:"strategy"`
	MaxRetry int          `json:"max_retry"`
}

// MongoNode mongodb node config
type MongoNode struct {
	URI     string `json:"uri"`
	Logging bool   `json:"logging"`
	Weight  int    `json:"weight"`
}

// getMongoDBConfigs reads MongoDB configurations
func getMongoDBConfigs(v *viper.Viper) *MongoDB {
	return &MongoDB{
		Master: &MongoNode{
			URI:     v.GetString("data.mongodb.master.uri"),
			Logging: v.GetBool("data.mongodb.master.logging"),
		},
		Slaves:   getMongoSlaveConfigs(v),
		Strategy: v.GetString("data.mongodb.strategy"),
		MaxRetry: v.GetInt("data.mongodb.max_retry"),
	}
}

// getMongoSlaveConfigs reads MongoDB slave configurations
func getMongoSlaveConfigs(v *viper.Viper) []*MongoNode {
	var slaves []*MongoNode

	// get mongodb slaves
	slavesConfig := v.Get("data.mongodb.slaves")
	if slavesConfig == nil {
		return slaves
	}

	// check if the slaves config is a slice
	slavesInterface, ok := slavesConfig.([]any)
	if !ok {
		fmt.Println("Invalid mongodb slaves configuration format")
		return slaves
	}

	// parse each slave
	for i := 0; i < len(slavesInterface); i++ {
		slave := &MongoNode{
			URI:     v.GetString(fmt.Sprintf("data.mongodb.slaves.%d.uri", i)),
			Logging: v.GetBool(fmt.Sprintf("data.mongodb.slaves.%d.logging", i)),
			Weight:  v.GetInt(fmt.Sprintf("data.mongodb.slaves.%d.weight", i)),
		}

		// check if the slave is valid
		if slave.URI != "" {
			// set default values
			if slave.Weight <= 0 {
				slave.Weight = 1
			}
			slaves = append(slaves, slave)
		}
	}

	return slaves
}
