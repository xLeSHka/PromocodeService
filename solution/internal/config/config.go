package config

import (
	"solution/pkg/db/cache"
	"solution/pkg/db/postgres"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	ServerAddress    string `env:"SERVER_ADDRESS"`
	AntifraudAddress string `env:"ANTIFRAUD_ADDRESS"`
	RandomSecret     string `env:"RANDOM_SECRET"`
	postgres.PostgresConfig
	cache.RedisConfig
}

func Read() (*Config, error) {
	cfg := Config{}
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
