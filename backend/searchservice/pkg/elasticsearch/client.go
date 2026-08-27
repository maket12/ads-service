package elasticsearch

import (
	"context"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8"
)

type Config struct {
	Addresses []string
	Username  string
	Password  string
}

func NewConfig(addresses []string, username, password string) *Config {
	return &Config{
		Addresses: addresses,
		Username:  username,
		Password:  password,
	}
}

func (c *Config) URL() string {
	if len(c.Addresses) > 0 {
		return c.Addresses[0]
	}
	return ""
}

type Client struct {
	*elasticsearch.Client
}

func NewClient(ctx context.Context, cfg *Config) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("elasticsearch config is not specified")
	}

	esCfg := elasticsearch.Config{
		Addresses: cfg.Addresses,
		Username:  cfg.Username,
		Password:  cfg.Password,
	}

	client, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create elasticsearch client: %w", err)
	}

	res, err := client.Info(client.Info.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to ping elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("failed to connect to elasticsearch, status: %s", res.Status())
	}

	return &Client{Client: client}, nil
}

func (c *Client) Close() {}
