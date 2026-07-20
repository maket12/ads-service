package rabbitmq

import (
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	WaitTime time.Duration // how much time to wait between retries of connecting
	Attempts int           // amount of retries to connect
}

func NewConfig(
	host string, port int,
	user, password string,
	waitTime time.Duration,
	attempts int,
) *Config {
	return &Config{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		WaitTime: waitTime,
		Attempts: attempts,
	}
}

// Builds connection url
func (c *Config) url() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%d/",
		c.User, c.Password, c.Host, c.Port,
	)
}

type Client struct {
	Conn *amqp.Connection
}

func NewClient(config *Config) (*Client, error) {
	var (
		conn *amqp.Connection
		err  error
	)

	// Connection with retries
	for i := 1; i <= config.Attempts; i++ {
		conn, err = amqp.Dial(config.url())
		if err == nil {
			return &Client{Conn: conn}, nil
		}

		time.Sleep(config.WaitTime)
	}

	return nil, fmt.Errorf("failed to connect to rabbitmq after %d attemptsL %w",
		config.Attempts, err,
	)
}

func (c *Client) Close() error {
	return c.Conn.Close()
}
