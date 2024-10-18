package connection

import (
	"context"
	"errors"
	"ncobase/common/config"
	"ncobase/common/log"

	"github.com/redis/go-redis/v9"
)

// newRedisClient creates a new Redis client
func newRedisClient(conf *config.Redis) (*redis.Client, error) {
	if conf == nil || conf.Addr == "" {
		log.Infof(context.Background(), "Redis configuration is nil or empty")
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
		log.Errorf(context.Background(), "Redis connect error: %v", err)
		return nil, err
	}

	log.Infof(context.Background(), "Redis connected")

	return rc, nil
}
