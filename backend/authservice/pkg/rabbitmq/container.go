package rabbitmq

import (
	"context"
	"fmt"
	"time"

	container "github.com/testcontainers/testcontainers-go/modules/rabbitmq"
)

type TestContainer struct {
	Container *container.RabbitMQContainer
	Config    *Config
}

// StartTestContainer Creates and launches a RabbitMQ container for tests.
// Do not use it as a real queue because the data will be lost once the program stops.
// You can either specify a configuration for the container
func StartTestContainer(ctx context.Context) (*TestContainer, error) {
	var (
		user     = "user"
		password = "password"
		waitTime = time.Second * 10
		attempts = 5
	)

	rabbitMQContainer, err := container.Run(ctx,
		"rabbitmq:4.2.2-management-alpine",
		container.WithAdminUsername(user),
		container.WithAdminPassword(password),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	host, err := rabbitMQContainer.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get host: %w", err)
	}

	port, err := rabbitMQContainer.MappedPort(ctx, "5672")
	if err != nil {
		return nil, fmt.Errorf("failed to get port: %w", err)
	}

	cfg := NewConfig(host, int(port.Num()), user,
		password, waitTime, attempts,
	)

	return &TestContainer{
		Container: rabbitMQContainer,
		Config:    cfg,
	}, nil
}

// Close Terminates the test RabbitMQ container
func (tc *TestContainer) Close(ctx context.Context) error {
	return tc.Container.Terminate(ctx)
}
