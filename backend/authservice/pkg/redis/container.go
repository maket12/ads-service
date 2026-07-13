package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	container "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TestContainer struct {
	Container *container.RedisContainer
	Config    *Config
}

// StartTestContainer Creates and launches a Redis container for tests.
// Do not use it as a database because the data will be lost once the program stops.
// You can either specify a configuration for the container (db, timeouts, etc.)
// or omit it to get the default configuration.
func StartTestContainer(ctx context.Context, cfg *Config) (*TestContainer, error) {
	redisContainer, err := container.Run(ctx,
		"redis:8.8.0-alpine",
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections tcp").
				WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	host, err := redisContainer.Host(ctx)
	if err != nil || host == "" {
		return nil, fmt.Errorf("failed to get host: %w", err)
	}

	port, err := redisContainer.MappedPort(ctx, "6379")
	if err != nil {
		return nil, fmt.Errorf("failed to get port: %w", err)
	}

	var (
		poolSize     = 10
		minIdleConn  = 5
		dialTimeout  = 5 * time.Second
		readTimeout  = 3 * time.Second
		writeTimeout = 3 * time.Second
		db           = 0
	)

	if cfg != nil {
		poolSize = cfg.PoolSize
		minIdleConn = cfg.MinIdleConn
		dialTimeout = cfg.DialTimeout
		readTimeout = cfg.ReadTimeout
		writeTimeout = cfg.WriteTimeout
		db = cfg.DB
	}

	newCfg := NewConfig(host, int(port.Num()), "", db,
		poolSize, minIdleConn, dialTimeout, readTimeout, writeTimeout,
	)

	return &TestContainer{
		Container: redisContainer,
		Config:    newCfg,
	}, nil
}

// Close Terminates the test Redis container
func (tc *TestContainer) Close(ctx context.Context) error {
	return tc.Container.Terminate(ctx)
}

// FlushAll deletes all data from the redis instance.
func (tc *TestContainer) FlushAll(ctx context.Context) error {
	rdb := redis.NewClient(&redis.Options{
		Addr:     tc.Config.Addr(),
		Password: tc.Config.Password,
		DB:       tc.Config.DB,
	})
	defer func() { _ = rdb.Close() }()

	if err := rdb.FlushAll(ctx).Err(); err != nil {
		return fmt.Errorf("failed to flush redis: %w", err)
	}

	return nil
}
