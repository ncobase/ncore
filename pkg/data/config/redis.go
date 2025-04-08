package config

import (
	"time"

	"github.com/spf13/viper"
)

// Redis redis config struct
type Redis struct {
	Addr         string
	Username     string
	Password     string
	Db           int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	DialTimeout  time.Duration
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
