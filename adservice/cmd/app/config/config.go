package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	// Postgres
	PgHost     string `env:"PG_HOST,required"`
	PgPort     int    `env:"PG_PORT" envDefault:"5433"`
	PgUser     string `env:"PG_USER,required"`
	PgPassword string `env:"PG_PASSWORD,required"`
	PgDBName   string `env:"PG_DB_NAME,required"`
	PgSSLMode  string `env:"DB_SSL_MODE" envDefault:"prefer"`

	PgOpenConn     int           `env:"PG_OPEN_CONNECTIONS" envDefault:"25"`
	PgIdleConn     int           `env:"PG_IDLE_CONNECTIONS" envDefault:"25"`
	PgConnLifeTime time.Duration `env:"PG_CONNECTION_LIFETIME" envDefault:"5m"`

	// Mongo
	MongoHost     string `env:"MONGO_HOST,required"`
	MongoPort     int    `env:"MONGO_PORT" envDefault:"27017"`
	MongoUser     string `env:"MONGO_USER,required"`
	MongoPassword string `env:"MONGO_PASSWORD,required"`
	MongoDBName   string `env:"MONGO_DB_NAME,required"`

	MongoCollectionName string `env:"MONGO_COLLECTION_NAME,required"`

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
	fmt.Printf("   Mongo Host: %s\n", cfg.MongoHost)
	fmt.Printf("   gRPC Port: %d\n", cfg.GRPCPort)

	return cfg, nil
}
