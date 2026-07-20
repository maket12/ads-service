package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	// Database
	DbHost     string `env:"DB_HOST,required"`
	DbPort     int    `env:"DB_PORT" envDefault:"5432"`
	DbUser     string `env:"DB_USER,required"`
	DbPassword string `env:"DB_PASSWORD,required"`
	DbName     string `env:"DB_NAME,required"`
	DbSSLMode  string `env:"DB_SSL_MODE" envDefault:"prefer"`

	DbMaxConn         int           `env:"DB_MAX_CONNECTIONS" envDefault:"30"`
	DbMinConn         int           `env:"DB_MIN_CONNECTIONS" envDefault:"10"`
	DbMaxConnLifeTime time.Duration `env:"DB_MAX_CONNECTION_LIFETIME" envDefault:"10m"`
	DbMaxConnIdleTime time.Duration `env:"DB_MAX_CONNECTION_IDLETIME" envDefault:"5m"`

	// RabbitMQ
	RabbitHost     string `env:"RABBIT_HOST,required"`
	RabbitPort     int    `env:"RABBIT_PORT" envDefault:"5672"`
	RabbitUser     string `env:"RABBIT_USER,required"`
	RabbitPassword string `env:"RABBIT_PASSWORD,required"`

	RabbitWaitTime time.Duration `env:"RABBIT_WAIT_TIME" envDefault:"30s"`
	RabbitAttempts int           `env:"RABBIT_ATTEMPTS" envDefault:"5"`

	ExchangeName string `env:"ACCOUNT_EXCHANGE" envDefault:"account_topic"`
	QueueName    string `env:"USER_QUEUE" envDefault:"account_create"`
	RoutingKey   string `env:"ACCOUNT_ROUTING_KEY,required"`

	// Phone validator
	PhoneDefaultRegion string `env:"PHONE_DEFAULT_REGION"`

	// Service
	GRPCPort    int    `env:"GRPC_PORT" envDefault:"50052"`
	LogLevel    string `env:"LOG_LEVEL" envDefault:"INFO"`
	Environment string `env:"ENVIRONMENT" envDefault:"production"`
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

	// RabbitMQ
	RabbitHost     string `env:"TEST_RABBIT_HOST" envDefault:"localhost"`
	RabbitPort     int    `env:"TEST_RABBIT_PORT" envDefault:"5672"`
	RabbitUser     string `env:"TEST_RABBIT_USER" envDefault:"user"`
	RabbitPassword string `env:"TEST_RABBIT_PASSWORD" envDefault:"pass"`

	RabbitWaitTime time.Duration `env:"TEST_RABBIT_WAIT_TIME" envDefault:"30s"`
	RabbitAttempts int           `env:"TEST_RABBIT_ATTEMPTS" envDefault:"5"`

	ExchangeName string `env:"TEST_ACCOUNT_EXCHANGE" envDefault:"account_topic"`
	QueueName    string `env:"TEST_USER_QUEUE" envDefault:"account_create"`
	RoutingKey   string `env:"TEST_ACCOUNT_ROUTING_KEY" envDefault:"account.created"`

	// Phone validator
	PhoneDefaultRegion string `env:"TEST_PHONE_DEFAULT_REGION"`

	// Service
	GRPCPort    int    `env:"TEST_GRPC_PORT" envDefault:"50052"`
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
