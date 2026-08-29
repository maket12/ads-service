package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	// ElasticSearch
	ESAddresses []string `env:"ELASTICSEARCH_ADDRESSES,required" envSeparator:","`
	ESUsername  string   `env:"ELASTICSEARCH_USERNAME,required"`
	ESPassword  string   `env:"ELASTICSEARCH_PASSWORD,required"`
	ESIndexName string   `env:"ELASTICSEARCH_INDEX_NAME,required"`

	// RabbitMQ
	RabbitHost     string `env:"RABBIT_HOST,required"`
	RabbitPort     int    `env:"RABBIT_PORT" envDefault:"5672"`
	RabbitUser     string `env:"RABBIT_USER,required"`
	RabbitPassword string `env:"RABBIT_PASSWORD,required"`

	RabbitWaitTime time.Duration `env:"RABBIT_WAIT_TIME" envDefault:"30s"`
	RabbitAttempts int           `env:"RABBIT_ATTEMPTS" envDefault:"5"`

	AdExchange string `env:"AD_EXCHANGE" envDefault:"ad_topic"`
	AdQueue    string `env:"AD_QUEUE" envDefault:"ad_events"`

	AdPublishedRoutingKey string `env:"AD_PUBLISHED_ROUTING_KEY" envDefault:"ad.published"`
	AdUpdatedRoutingKey   string `env:"AD_UPDATED_ROUTING_KEY" envDefault:"ad.updated"`
	AdRejectedRoutingKey  string `env:"AD_REJECTED_ROUTING_KEY" envDefault:"ad.rejected"`
	AdDeletedRoutingKey   string `env:"AD_DELETED_ROUTING_KEY" envDefault:"ad.deleted"`

	// Service
	GRPCPort    int    `env:"GRPC_PORT" envDefault:"50054"`
	LogLevel    string `env:"LOG_LEVEL" envDefault:"INFO"`
	Environment string `env:"ENVIRONMENT" envDefault:"production"`
}

type TestConfig struct {
	// ElasticSearch
	ESAddresses []string `env:"TEST_ELASTICSEARCH_ADDRESSES" envDefault:"http://localhost:9200" envSeparator:","`
	ESUsername  string   `env:"TEST_ELASTICSEARCH_USERNAME" envDefault:"user"`
	ESPassword  string   `env:"TEST_ELASTICSEARCH_PASSWORD" envDefault:"pass"`
	ESIndexName string   `env:"TEST_ELASTICSEARCH_INDEX_NAME" envDefault:"ads"`

	// RabbitMQ
	RabbitHost     string `env:"TEST_RABBIT_HOST" envDefault:"localhost"`
	RabbitPort     int    `env:"TEST_RABBIT_PORT" envDefault:"5672"`
	RabbitUser     string `env:"TEST_RABBIT_USER" envDefault:"user"`
	RabbitPassword string `env:"TEST_RABBIT_PASSWORD" envDefault:"pass"`

	RabbitWaitTime time.Duration `env:"TEST_RABBIT_WAIT_TIME" envDefault:"30s"`
	RabbitAttempts int           `env:"TEST_RABBIT_ATTEMPTS" envDefault:"5"`

	AdExchange string `env:"TEST_AD_EXCHANGE" envDefault:"ad_topic"`
	AdQueue    string `env:"TEST_AD_QUEUE" envDefault:"ad_events"`

	AdPublishedRoutingKey string `env:"TEST_AD_PUBLISHED_ROUTING_KEY" envDefault:"ad.published"`
	AdUpdatedRoutingKey   string `env:"TEST_AD_UPDATED_ROUTING_KEY" envDefault:"ad.updated"`
	AdRejectedRoutingKey  string `env:"TEST_AD_REJECTED_ROUTING_KEY" envDefault:"ad.rejected"`
	AdDeletedRoutingKey   string `env:"TEST_AD_DELETED_ROUTING_KEY" envDefault:"ad.deleted"`

	// Phone validator
	PhoneDefaultRegion string `env:"TEST_PHONE_DEFAULT_REGION"`

	// Service
	GRPCPort    int    `env:"TEST_GRPC_PORT" envDefault:"50054"`
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
	fmt.Printf("   ElasticSearch Hosts: %v\n", cfg.ESAddresses)
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
