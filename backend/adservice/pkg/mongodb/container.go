package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	container "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type TestContainer struct {
	Container *container.MongoDBContainer
	Config    *Config
}

// StartTestContainer Creates and launches a MongoDB container for tests.
// Do not use it as a database because the data will be lost once the program stops.
func StartTestContainer(ctx context.Context) (*TestContainer, error) {
	var (
		user     = "user"
		password = "password"
		dbName   = "testdb"
	)

	mongoContainer, err := container.Run(ctx,
		"mongo:7",
		container.WithUsername(user),
		container.WithPassword(password),
		testcontainers.WithWaitStrategy(
			wait.ForLog("Waiting for connections").
				WithOccurrence(1).WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	host, err := mongoContainer.Host(ctx)
	if err != nil || host == "" {
		return nil, fmt.Errorf("failed to get host: %w", err)
	}

	port, err := mongoContainer.MappedPort(ctx, "27017")
	if err != nil {
		return nil, fmt.Errorf("failed to get port: %w", err)
	}

	cfg := NewConfig(host, int(port.Num()), user, password, dbName)

	return &TestContainer{
		Container: mongoContainer,
		Config:    cfg,
	}, nil
}

// Close Terminates the test MongoDB container
func (tc *TestContainer) Close(ctx context.Context) error {
	return tc.Container.Terminate(ctx)
}

// ClearCollections deletes all documents from the specified collections.
// If no collections are specified, it clears every collection in the database.
func (tc *TestContainer) ClearCollections(ctx context.Context, collections ...string) error {
	client, err := NewClient(ctx, tc.Config)
	if err != nil {
		return fmt.Errorf("failed to connect to mongodb for clear: %w", err)
	}
	defer func() { _ = client.Close(ctx) }()

	if len(collections) == 0 {
		names, listErr := client.Database.ListCollectionNames(ctx, bson.D{})
		if listErr != nil {
			return fmt.Errorf("failed to fetch collection names for clear: %w", listErr)
		}
		collections = names

		if len(collections) == 0 {
			return nil
		}
	}

	for _, name := range collections {
		if _, err = client.Database.Collection(name).DeleteMany(ctx, bson.D{}); err != nil {
			return fmt.Errorf("failed to clear collection %q: %w", name, err)
		}
	}

	return nil
}

// DropCollections drops the specified collections entirely (including indexes).
// If no collections are specified, it drops every collection in the database.
func (tc *TestContainer) DropCollections(ctx context.Context, collections ...string) error {
	client, err := NewClient(ctx, tc.Config)
	if err != nil {
		return fmt.Errorf("failed to connect to mongodb for drop: %w", err)
	}
	defer func() { _ = client.Close(ctx) }()

	if len(collections) == 0 {
		names, listErr := client.Database.ListCollectionNames(ctx, bson.D{})
		if listErr != nil {
			return fmt.Errorf("failed to fetch collection names for drop: %w", listErr)
		}
		collections = names

		if len(collections) == 0 {
			return nil
		}
	}

	for _, name := range collections {
		if err = client.Database.Collection(name).Drop(ctx); err != nil {
			return fmt.Errorf("failed to drop collection %q: %w", name, err)
		}
	}

	return nil
}
