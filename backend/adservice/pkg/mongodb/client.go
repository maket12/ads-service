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
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

func NewConfig(
	host string,
	port int,
	user string,
	password string,
	dbName string,
) *Config {
	return &Config{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DBName:   dbName,
	}
}

func (c *Config) uri() string {
	return fmt.Sprintf("mongodb://%s:%s@%s:%d",
		c.User, c.Password, c.Host, c.Port,
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
