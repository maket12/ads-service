package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	// Database
	PgHost     string `env:"PG_HOST,required"`
	PgPort     int    `env:"PG_PORT" envDefault:"5432"`
	PgUser     string `env:"PG_USER,required"`
	PgPassword string `env:"PG_PASSWORD,required"`
	PgDBName   string `env:"PG_DB_NAME,required"`
	PgSSLMode  string `env:"DB_SSL_MODE" envDefault:"prefer"`

	PgOpenConn     int           `env:"PG_OPEN_CONNECTIONS" envDefault:"25"`
	PgIdleConn     int           `env:"PG_IDLE_CONNECTIONS" envDefault:"25"`
	PgConnLifeTime time.Duration `env:"DB_CONNECTION_LIFETIME" envDefault:"5m"`

	// JWT
	AccessSecret  string        `env:"ACCESS_SECRET,required"`
	RefreshSecret string        `env:"REFRESH_SECRET,required"`
	AccessTTL     time.Duration `env:"ACCESS_TTL" envDefault:"15m"`
	RefreshTTL    time.Duration `env:"REFRESH_TTL" envDefault:"720h"`

	// Password hasher
	PasswordCost int `env:"PASSWORD_COST" envDefault:"4"`

	// RabbitMQ
	RabbitHost     string `env:"RABBIT_HOST,required"`
	RabbitPort     int    `env:"RABBIT_PORT" envDefault:"5672"`
	RabbitUser     string `env:"RABBIT_USER,required"`
	RabbitPassword string `env:"RABBIT_PASSWORD,required"`

	RabbitWaitTime time.Duration `env:"RABBIT_WAIT_TIME" envDefault:"30s"`
	RabbitAttempts int           `env:"RABBIT_ATTEMPTS" envDefault:"5"`

	ExchangeName string `env:"EXCHANGE_NAME" envDefault:"account_topic"`
	RoutingKey   string `env:"ROUTING_KEY,required"`

	// Service
	GRPCPort    int    `env:"GRPC_PORT" envDefault:"50051"`
	LogLevel    string `env:"LOG_LEVEL" envDefault:"INFO"`
	Environment string `env:"ENVIRONMENT" envDefault:"development"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	fmt.Printf("Config loaded successfully\n")
	fmt.Printf("   Environment: %s\n", cfg.Environment)
	fmt.Printf("   Log Level: %s\n", cfg.LogLevel)
	fmt.Printf("   Postgres Host: %s\n", cfg.PgHost)
	fmt.Printf("   RabbitMQ Host: %s\n", cfg.RabbitHost)
	fmt.Printf("   gRPC Port: %d\n", cfg.GRPCPort)

	return cfg, nil
}
