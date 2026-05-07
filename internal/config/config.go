package config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	LogLevel string `envconfig:"LOG_LEVEL" default:"info"`
}

type DatabaseConfig struct {
	URL             string        `envconfig:"DATABASE_URL" default:"postgres://auction:auctionpass@localhost:5432/auction?sslmode=disable"`
	MaxConns        int           `envconfig:"DATABASE_MAX_CONNS" default:"10"`
	MinConns        int           `envconfig:"DATABASE_MIN_CONNS" default:"2"`
	MaxConnIdleTime time.Duration `envconfig:"DATABASE_MAX_CONN_IDLE_TIME" default:"15m"`
	MaxConnLifetime time.Duration `envconfig:"DATABASE_MAX_CONN_LIFETIME" default:"1h"`
	HealthTimeout   time.Duration `envconfig:"DATABASE_HEALTH_TIMEOUT" default:"5s"`
}

type ServerConfig struct {
	Port            int           `envconfig:"SERVER_PORT" default:"8080"`
	ReadTimeout     time.Duration `envconfig:"SERVER_READ_TIMEOUT" default:"5s"`
	WriteTimeout    time.Duration `envconfig:"SERVER_WRITE_TIMEOUT" default:"10s"`
	ShutdownTimeout time.Duration `envconfig:"SERVER_SHUTDOWN_TIMEOUT" default:"15s"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return &cfg, nil
}
