package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	Host     string
	Port     int
	Password string
	DB       int

	PoolSize     int
	MinIdleConn  int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func NewConfig(
	host string, port int,
	password string, db int,
	poolSize, minIdleConn int,
	dialTimeout, readTimeout, writeTimeout time.Duration,
) *Config {
	return &Config{
		Host:         host,
		Port:         port,
		Password:     password,
		DB:           db,
		PoolSize:     poolSize,
		MinIdleConn:  minIdleConn,
		DialTimeout:  dialTimeout,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}
}

// Addr returns host:port for redis.Options
func (rc *Config) Addr() string {
	return fmt.Sprintf("%s:%d", rc.Host, rc.Port)
}

type Client struct {
	*redis.Client
}

func NewClient(ctx context.Context, cfg *Config) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("redis config is not specified")
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr(),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConn,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		_ = rdb.Close()
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return &Client{Client: rdb}, nil
}

func (c *Client) Close() error {
	return c.Client.Close()
}
