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

	// RabbitMQ
	RabbitHost     string `env:"RABBIT_HOST,required"`
	RabbitPort     int    `env:"RABBIT_PORT" envDefault:"5672"`
	RabbitUser     string `env:"RABBIT_USER,required"`
	RabbitPassword string `env:"RABBIT_PASSWORD,required"`

	RabbitWaitTime time.Duration `env:"RABBIT_WAIT_TIME" envDefault:"30s"`
	RabbitAttempts int           `env:"RABBIT_ATTEMPTS" envDefault:"5"`

	AdExchange            string `env:"AD_EXCHANGE" envDefault:"ad_topic"`
	AdPublishedRoutingKey string `env:"AD_PUBLISHED_ROUTING_KEY" envDefault:"ad.published"`
	AdUpdatedRoutingKey   string `env:"AD_UPDATED_ROUTING_KEY" envDefault:"ad.updated"`
	AdRejectedRoutingKey  string `env:"AD_REJECTED_ROUTING_KEY" envDefault:"ad.rejected"`
	AdDeletedRoutingKey   string `env:"AD_DELETED_ROUTING_KEY" envDefault:"ad.deleted"`

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
	MongoHost     string `env:"TEST_MONGO_HOST" envDefault:"localhost"`
	MongoPort     int    `env:"TEST_MONGO_PORT" envDefault:"27017"`
	MongoUser     string `env:"TEST_MONGO_USER" envDefault:"user"`
	MongoPassword string `env:"TEST_MONGO_PASSWORD" envDefault:"pass"`
	MongoDBName   string `env:"TEST_MONGO_DB_NAME" envDefault:"test-db"`

	MongoCollectionName string `env:"TEST_MONGO_COLLECTION_NAME" envDefault:"test-img"`

	// RabbitMQ
	RabbitHost     string `env:"TEST_RABBIT_HOST"`
	RabbitPort     int    `env:"TEST_RABBIT_PORT"`
	RabbitUser     string `env:"TEST_RABBIT_USER"`
	RabbitPassword string `env:"TEST_RABBIT_PASSWORD"`

	RabbitWaitTime time.Duration `env:"TEST_RABBIT_WAIT_TIME"`
	RabbitAttempts int           `env:"TEST_RABBIT_ATTEMPTS"`

	AdExchange            string `env:"TEST_AD_EXCHANGE" envDefault:"ad_topic"`
	AdPublishedRoutingKey string `env:"TEST_AD_PUBLISHED_ROUTING_KEY" envDefault:"ad.published"`
	AdUpdatedRoutingKey   string `env:"TEST_AD_UPDATED_ROUTING_KEY" envDefault:"ad.updated"`
	AdRejectedRoutingKey  string `env:"TEST_AD_REJECTED_ROUTING_KEY" envDefault:"ad.rejected"`
	AdDeletedRoutingKey   string `env:"TEST_AD_DELETED_ROUTING_KEY" envDefault:"ad.deleted"`

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
	fmt.Printf("   MongoDB Host: %s\n", cfg.MongoHost)
	fmt.Printf("   RabbitMQ Host: %s\n", cfg.RabbitHost)
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
