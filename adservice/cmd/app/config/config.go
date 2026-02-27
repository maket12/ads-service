package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	// Postgres
	PgHost     string `env:"AD_PG_HOST,required"`
	PgPort     int    `env:"AD_PG_PORT" envDefault:"5433"`
	PgUser     string `env:"AD_PG_USER,required"`
	PgPassword string `env:"AD_PG_PASSWORD,required"`
	PgDBName   string `env:"AD_PG_DB_NAME,required"`
	PgSSLMode  string `env:"AD_PG_SSL_MODE" envDefault:"prefer"`

	PgOpenConn     int           `env:"AD_PG_OPEN_CONNECTIONS" envDefault:"25"`
	PgIdleConn     int           `env:"AD_PG_IDLE_CONNECTIONS" envDefault:"25"`
	PgConnLifeTime time.Duration `env:"AD_PG_CONNECTION_LIFETIME" envDefault:"5m"`

	// Mongo
	MongoHost     string `env:"AD_MONGO_HOST,required"`
	MongoPort     int    `env:"AD_MONGO_PORT" envDefault:"27017"`
	MongoUser     string `env:"AD_MONGO_USER,required"`
	MongoPassword string `env:"AD_MONGO_PASSWORD,required"`
	MongoDBName   string `env:"AD_MONGO_DB_NAME,required"`

	MongoCollectionName string `env:"AD_MONGO_COLLECTION_NAME,required"`

	// Service
	GRPCPort    int    `env:"AD_GRPC_PORT" envDefault:"50051"`
	LogLevel    string `env:"AD_LOG_LEVEL" envDefault:"INFO"`
	Environment string `env:"AD_ENVIRONMENT" envDefault:"development"`
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
