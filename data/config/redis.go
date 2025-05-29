package config

import (
	"time"

	"github.com/spf13/viper"
)

// Redis redis config struct
type Redis struct {
	Addr         string        `json:"addr" yaml:"addr"`
	Username     string        `json:"username" yaml:"username"`
	Password     string        `json:"password" yaml:"password"`
	Db           int           `json:"db" yaml:"db"`
	ReadTimeout  time.Duration `json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout" yaml:"write_timeout"`
	DialTimeout  time.Duration `json:"dial_timeout" yaml:"dial_timeout"`
}

// getRedisConfigs reads Redis configurations
func getRedisConfigs(v *viper.Viper) *Redis {
	return &Redis{
		Addr:         v.GetString("data.redis.addr"),
		Username:     v.GetString("data.redis.username"),
		Password:     v.GetString("data.redis.password"),
		Db:           v.GetInt("data.redis.db"),
		ReadTimeout:  v.GetDuration("data.redis.read_timeout"),
		WriteTimeout: v.GetDuration("data.redis.write_timeout"),
		DialTimeout:  v.GetDuration("data.redis.dial_timeout"),
	}
}
