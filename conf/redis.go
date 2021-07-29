package conf

// RedisConfig redis object
type RedisConfig struct {
	Host     string
	Password string
	Port     int
	DB       int
}

func getRedisConfig() *RedisConfig {
	return &RedisConfig{
		Host:     c.GetString("redis.host"),
		Password: c.GetString("redis.password"),
		Port:     c.GetInt("redis.port"),
		DB:       c.GetInt("redis.db"),
	}
}
