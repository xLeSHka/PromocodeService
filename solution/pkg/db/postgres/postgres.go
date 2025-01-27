package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type PostgresConfig struct {
	UserName string `env:"POSTGRES_USERNAME"`
	Password string `env:"POSTGRES_PASSWORD"`
	Host     string `env:"POSTGRES_HOST"`
	Port     string `env:"POSTGRES_PORT"`
	DbName   string `env:"POSTGRES_DATABASE"`
}
type DB struct {
	Db *sqlx.DB
}

func New(cfg PostgresConfig) (*DB, error) {
	dsn := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable host=%s port=%s",
		cfg.UserName, cfg.Password, cfg.DbName, cfg.Host, cfg.Port)
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if _, err := db.Conn(context.Background()); err != nil {
		return nil, err
	}
	return &DB{Db: db}, nil
}
