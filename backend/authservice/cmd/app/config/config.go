package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	// Database
	DbHost     string `env:"DB_HOST,required"`
	DbPort     int    `env:"DB_PORT" envDefault:"5431"`
	DbUser     string `env:"DB_USER,required"`
	DbPassword string `env:"DB_PASSWORD,required"`
	DbName     string `env:"DB_NAME,required"`
	DbSSLMode  string `env:"DB_SSL_MODE" envDefault:"prefer"`

	DbMaxConn         int           `env:"DB_MAX_CONNECTIONS" envDefault:"30"`
	DbMinConn         int           `env:"DB_MIN_CONNECTIONS" envDefault:"10"`
	DbMaxConnLifeTime time.Duration `env:"DB_MAX_CONNECTION_LIFETIME" envDefault:"10m"`
	DbMaxConnIdleTime time.Duration `env:"DB_MAX_CONNECTION_IDLETIME" envDefault:"5m"`

	// Redis
	RedisHost     string `env:"REDIS_HOST,required"`
	RedisPort     int    `env:"REDIS_PORT" envDefault:"6379"`
	RedisPassword string `env:"REDIS_PASSWORD,required"`
	RedisDBNumber int    `env:"REDIS_DB_NUM" envDefault:"0"`

	RedisPoolSize    int `env:"REDIS_POOL_SIZE" envDefault:"10"`
	RedisMinIdleConn int `env:"REDIS_MIN_IDLE_CONN" envDefault:"2"`

	RedisDialTimeout  time.Duration `env:"REDIS_DIAL_TIMEOUT" envDefault:"5s"`
	RedisReadTimeout  time.Duration `env:"REDIS_READ_TIMEOUT" envDefault:"3s"`
	RedisWriteTimeout time.Duration `env:"REDIS_WRITE_TIMEOUT" envDefault:"3s"`

	// JWT
	AccessSecret  string        `env:"ACCESS_SECRET,required"`
	RefreshSecret string        `env:"REFRESH_SECRET,required"`
	AccessTTL     time.Duration `env:"ACCESS_TTL" envDefault:"15m"`
	RefreshTTL    time.Duration `env:"REFRESH_TTL" envDefault:"720h"`

	// Password Hasher
	PasswordCost int `env:"PASSWORD_COST" envDefault:"4"`

	// RabbitMQ
	RabbitHost     string `env:"RABBIT_HOST,required"`
	RabbitPort     int    `env:"RABBIT_PORT" envDefault:"5672"`
	RabbitUser     string `env:"RABBIT_USER,required"`
	RabbitPassword string `env:"RABBIT_PASSWORD,required"`

	RabbitWaitTime time.Duration `env:"RABBIT_WAIT_TIME" envDefault:"30s"`
	RabbitAttempts int           `env:"RABBIT_ATTEMPTS" envDefault:"5"`

	ExchangeName string `env:"ACCOUNT_EXCHANGE" envDefault:"account_topic"`
	RoutingKey   string `env:"ACCOUNT_ROUTING_KEY,required"`

	// SMTP Client
	SMTPHost     string `env:"SMTP_HOST,required"`
	SMTPPort     string `env:"SMTP_PORT,required"`
	SMTPEmail    string `env:"SMTP_EMAIL,required"`
	SMTPPassword string `env:"SMTP_PASSWORD,required"`

	VerificationBaseURL string        `env:"VERIFICATION_BASE_URL,required"`
	VerificationTTL     time.Duration `env:"VERIFICATION_TTL" envDefault:"5m"`

	// Service
	GRPCPort    int    `env:"GRPC_PORT" envDefault:"50051"`
	LogLevel    string `env:"LOG_LEVEL" envDefault:"INFO"`
	Environment string `env:"ENVIRONMENT" envDefault:"production"`
}

type TestConfig struct {
	// Database
	DbHost     string `env:"TEST_DB_HOST"`
	DbPort     int    `env:"TEST_DB_PORT"`
	DbUser     string `env:"TEST_DB_USER"`
	DbPassword string `env:"TEST_DB_PASSWORD"`
	DbName     string `env:"TEST_DB_NAME"`
	DbSSLMode  string `env:"TEST_DB_SSL_MODE" envDefault:"disable"`

	DbMaxConn         int           `env:"TEST_DB_MAX_CONNECTIONS"`
	DbMinConn         int           `env:"TEST_DB_MIN_CONNECTIONS"`
	DbMaxConnLifeTime time.Duration `env:"TEST_DB_MAX_CONNECTION_LIFETIME"`
	DbMaxConnIdleTime time.Duration `env:"TEST_DB_MAX_CONNECTION_IDLETIME"`

	// Redis
	RedisHost     string `env:"TEST_REDIS_HOST"`
	RedisPort     int    `env:"TEST_REDIS_PORT"`
	RedisPassword string `env:"TEST_REDIS_PASSWORD"`
	RedisDBNumber int    `env:"TEST_REDIS_DB_NUM"`

	RedisPoolSize    int `env:"TEST_REDIS_POOL_SIZE"`
	RedisMinIdleConn int `env:"TEST_REDIS_MIN_IDLE_CONN"`

	RedisDialTimeout  time.Duration `env:"TEST_REDIS_DIAL_TIMEOUT"`
	RedisReadTimeout  time.Duration `env:"TEST_REDIS_READ_TIMEOUT"`
	RedisWriteTimeout time.Duration `env:"TEST_REDIS_WRITE_TIMEOUT"`

	// JWT
	AccessSecret  string        `env:"TEST_ACCESS_SECRET" envDefault:"test-access-token"`
	RefreshSecret string        `env:"TEST_REFRESH_SECRET" envDefault:"test-refresh-token"`
	AccessTTL     time.Duration `env:"TEST_ACCESS_TTL" envDefault:"45s"`
	RefreshTTL    time.Duration `env:"TEST_REFRESH_TTL" envDefault:"45s"`

	// Password Hasher
	PasswordCost int `env:"TEST_PASSWORD_COST" envDefault:"4"`

	// RabbitMQ
	RabbitHost     string `env:"TEST_RABBIT_HOST"`
	RabbitPort     int    `env:"TEST_RABBIT_PORT"`
	RabbitUser     string `env:"TEST_RABBIT_USER"`
	RabbitPassword string `env:"TEST_RABBIT_PASSWORD"`

	RabbitWaitTime time.Duration `env:"TEST_RABBIT_WAIT_TIME"`
	RabbitAttempts int           `env:"TEST_RABBIT_ATTEMPTS"`

	ExchangeName string `env:"TEST_ACCOUNT_EXCHANGE" envDefault:"account_topic"`
	RoutingKey   string `env:"TEST_ACCOUNT_ROUTING_KEY" envDefault:"account.created"`

	// SMTP Client
	SMTPHost     string `env:"TEST_SMTP_HOST"`
	SMTPPort     string `env:"TEST_SMTP_PORT"`
	SMTPEmail    string `env:"TEST_SMTP_EMAIL"`
	SMTPPassword string `env:"TEST_SMTP_PASSWORD"`

	VerificationBaseURL string        `env:"TEST_VERIFICATION_BASE_URL"`
	VerificationTTL     time.Duration `env:"TEST_VERIFICATION_TTL" envDefault:"1m"`

	// Service
	GRPCPort    int    `env:"TEST_GRPC_PORT" envDefault:"50050"`
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
	fmt.Printf("   Redis Host: %s\n", cfg.RedisHost)
	fmt.Printf("   RabbitMQ Host: %s\n", cfg.RabbitHost)
	fmt.Printf("   SMTP Host: %s\n", cfg.SMTPHost)
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
