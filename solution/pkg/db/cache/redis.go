package cache

import (
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Host string `env:"REDIS_HOST"`
	Port string `env:"REDIS_PORT"`
}

func New(cfg RedisConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		MaxRetries: 5,
		DB: 0,
	})
}
