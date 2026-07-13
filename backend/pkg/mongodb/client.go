package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

type Config struct {
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
}

func NewConfig(
	dbHost string,
	dbPort int,
	dbUser string,
	dbPassword string,
	dbName string,
) *Config {
	return &Config{
		DBHost:     dbHost,
		DBPort:     dbPort,
		DBUser:     dbUser,
		DBPassword: dbPassword,
		DBName:     dbName,
	}
}

func (c *Config) uri() string {
	return fmt.Sprintf("mongodb://%s:%s@%s:%d",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort,
	)
}

type Client struct {
	Database *mongo.Database
	client   *mongo.Client
}

func NewClient(ctx context.Context, mongoCfg *Config) (*Client, error) {
	opts := options.Client().ApplyURI(mongoCfg.uri())

	client, err := mongo.Connect(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mongodb: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("failed to ping mongodb: %w", err)
	}

	return &Client{
		Database: client.Database(mongoCfg.DBName),
		client:   client,
	}, nil
}

func (c *Client) Close(ctx context.Context) error {
	return c.client.Disconnect(ctx)
}
