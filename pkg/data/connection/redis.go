package connection

import (
	"context"
	"errors"
	"fmt"
	"github.com/ncobase/ncore/pkg/data/config"

	"github.com/redis/go-redis/v9"
)

// newRedisClient creates a new Redis client
func newRedisClient(conf *config.Redis) (*redis.Client, error) {
	if conf == nil || conf.Addr == "" {
		return nil, errors.New("redis configuration is nil or empty")
	}

	rc := redis.NewClient(&redis.Options{
		Addr:         conf.Addr,
		Username:     conf.Username,
		Password:     conf.Password,
		DB:           conf.Db,
		ReadTimeout:  conf.ReadTimeout,
		WriteTimeout: conf.WriteTimeout,
		DialTimeout:  conf.DialTimeout,
		PoolSize:     10,
	})

	timeout, cancelFunc := context.WithTimeout(context.Background(), conf.DialTimeout)
	defer cancelFunc()
	if err := rc.Ping(timeout).Err(); err != nil {
		return nil, fmt.Errorf("redis connect error: %v", err)
	}

	return rc, nil
}
