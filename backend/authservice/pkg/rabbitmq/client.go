package rabbitmq

import (
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	WaitTime time.Duration // how much time to wait between retries of connecting
	Attempts int           // amount of retries to connect
}

func NewRabbitConfig(
	host string,
	port int,
	user string,
	password string,
	waitTime time.Duration,
	attempts int,
) *RabbitConfig {
	return &RabbitConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		WaitTime: waitTime,
		Attempts: attempts,
	}
}

// Builds connection url
func (c *RabbitConfig) url() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%d/",
		c.User, c.Password, c.Host, c.Port,
	)
}

type RabbitClient struct {
	Conn *amqp.Connection
}

func NewRabbitClient(config *RabbitConfig) (*RabbitClient, error) {
	var (
		conn *amqp.Connection
		err  error
	)

	// Connection with retries
	for i := 1; i <= config.Attempts; i++ {
		conn, err = amqp.Dial(config.url())
		if err == nil {
			return &RabbitClient{Conn: conn}, nil
		}

		time.Sleep(config.WaitTime)
	}

	return nil, fmt.Errorf("failed to connect to rabbitmq after %d attemptsL %w",
		config.Attempts, err,
	)
}

func (c *RabbitClient) Close() error {
	return c.Conn.Close()
}
