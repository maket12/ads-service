package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	// Database
	DbHost     string `env:"DB_HOST,required"`
	DbPort     int    `env:"DB_PORT" envDefault:"5433"`
	DbUser     string `env:"DB_USER,required"`
	DbPassword string `env:"DB_PASSWORD,required"`
	DbName     string `env:"DB_NAME,required"`
	DbSSLMode  string `env:"DB_SSL_MODE" envDefault:"prefer"`

	DbMaxConn         int           `env:"DB_MAX_CONNECTIONS" envDefault:"30"`
	DbMinConn         int           `env:"DB_MIN_CONNECTIONS" envDefault:"10"`
	DbMaxConnLifeTime time.Duration `env:"DB_MAX_CONNECTION_LIFETIME" envDefault:"10m"`
	DbMaxConnIdleTime time.Duration `env:"DB_MAX_CONNECTION_IDLETIME" envDefault:"5m"`

	// Mongo
	MongoHost     string `env:"MONGO_HOST,required"`
	MongoPort     int    `env:"MONGO_PORT" envDefault:"27017"`
	MongoUser     string `env:"MONGO_USER,required"`
	MongoPassword string `env:"MONGO_PASSWORD,required"`
	MongoDBName   string `env:"MONGO_DB_NAME,required"`

	MongoCollectionName string `env:"MONGO_COLLECTION_NAME,required"`

	// Service
	GRPCPort    int    `env:"AD_GRPC_PORT" envDefault:"50053"`
	LogLevel    string `env:"AD_LOG_LEVEL" envDefault:"INFO"`
	Environment string `env:"AD_ENVIRONMENT" envDefault:"development"`
}

type TestConfig struct {
	// Database
	DbHost     string `env:"TEST_DB_HOST" envDefault:"localhost"`
	DbPort     int    `env:"TEST_DB_PORT" envDefault:"5432"`
	DbUser     string `env:"TEST_DB_USER" envDefault:"user"`
	DbPassword string `env:"TEST_DB_PASSWORD" envDefault:"pass"`
	DbName     string `env:"TEST_DB_NAME" envDefault:"user-db"`
	DbSSLMode  string `env:"TEST_DB_SSL_MODE" envDefault:"prefer"`

	DbMaxConn         int           `env:"TEST_DB_MAX_CONNECTIONS" envDefault:"30"`
	DbMinConn         int           `env:"TEST_DB_MIN_CONNECTIONS" envDefault:"10"`
	DbMaxConnLifeTime time.Duration `env:"TEST_DB_MAX_CONNECTION_LIFETIME" envDefault:"10m"`
	DbMaxConnIdleTime time.Duration `env:"TEST_DB_MAX_CONNECTION_IDLETIME" envDefault:"5m"`

	// Mongo
	MongoHost     string `env:"TEST_MONGO_HOST,required"`
	MongoPort     int    `env:"TEST_MONGO_PORT" envDefault:"27017"`
	MongoUser     string `env:"TEST_MONGO_USER,required"`
	MongoPassword string `env:"TEST_MONGO_PASSWORD,required"`
	MongoDBName   string `env:"TEST_MONGO_DB_NAME,required"`

	MongoCollectionName string `env:"TEST_MONGO_COLLECTION_NAME,required"`

	// Service
	GRPCPort    int    `env:"TEST_GRPC_PORT" envDefault:"50053"`
	LogLevel    string `env:"TEST_LOG_LEVEL" envDefault:"DEBUG"`
	Environment string `env:"TEST_ENVIRONMENT" envDefault:"test"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	fmt.Printf("Config loaded successfully\n")
	fmt.Printf("   Environment: %s\n", cfg.Environment)
	fmt.Printf("   Log Level: %s\n", cfg.LogLevel)
	fmt.Printf("   Postgres Host: %s\n", cfg.DbHost)
	fmt.Printf("   Mongo Host: %s\n", cfg.MongoHost)
	fmt.Printf("   gRPC Port: %d\n", cfg.GRPCPort)

	return cfg, nil
}

func LoadTest() (*TestConfig, error) {
	cfg := &TestConfig{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to load test config: %v", err)
	}
	return cfg, nil
}
