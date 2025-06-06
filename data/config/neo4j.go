package config

import "github.com/spf13/viper"

// Neo4j neo4j config struct
type Neo4j struct {
	URI      string `json:"uri" yaml:"uri"`
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
}

// getNeo4jConfigs reads Neo4j configurations
func getNeo4jConfigs(v *viper.Viper) *Neo4j {
	return &Neo4j{
		URI:      v.GetString("data.neo4j.uri"),
		Username: v.GetString("data.neo4j.username"),
		Password: v.GetString("data.neo4j.password"),
	}
}
