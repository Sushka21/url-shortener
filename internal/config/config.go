package config

import (
	"fmt"
	"net"
	"time"

	"github.com/caarlos0/env/v10"
)

const (
	ShutdownTimeout   = 5 * time.Second
	HealthCheckPeriod = 30 * time.Second
	MaxConnIdleTime   = 5 * time.Minute
)

type (
	Config struct {
		HTTP struct {
			Host string `env:"HTTP_HOST" envDefault:"localhost"`
			Port string `env:"HTTP_PORT" envDefault:"8080"`
		}

		StorageType string `env:"STORAGE_TYPE" envDefault:"inmemory"`

		PG struct {
			Host     string `env:"POSTGRES_HOST" envDefault:"localhost"`
			Port     string `env:"POSTGRES_PORT" envDefault:"5432"`
			DB       string `env:"POSTGRES_DB" envDefault:"urlshortener"`
			User     string `env:"POSTGRES_USER" envDefault:"urlshortener_user"`
			Password string `env:"POSTGRES_PASSWORD" envDefault:"12345"`
		}
	}
)

func (c *Config) ConstructPostgresURL() string {
	hostPort := net.JoinHostPort(c.PG.Host, c.PG.Port)

	return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		c.PG.User,
		c.PG.Password,
		hostPort,
		c.PG.DB,
	)
}

func (c *Config) ConstructBaseURL() string {
	return fmt.Sprintf("http://%s:%s", c.HTTP.Host, c.HTTP.Port)
}

func New() (*Config, error) {
	var cfg Config
	err := env.Parse(&cfg)
	return &cfg, err
}
