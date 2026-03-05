package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/maket12/ads-service/pkg/rabbitmq"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type PublisherConfig struct {
	Exchange   string
	RoutingKey string
}

func NewPublisherConfig(exchange, routingKey string) *PublisherConfig {
	return &PublisherConfig{
		Exchange:   exchange,
		RoutingKey: routingKey,
	}
}

type AccountPublisher struct {
	cfg     *PublisherConfig
	client  *rabbitmq.RabbitClient
	channel *amqp.Channel
}

func NewAccountPublisher(
	cfg *PublisherConfig,
	client *rabbitmq.RabbitClient,
) (*AccountPublisher, error) {
	ch, err := client.Conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Exchange
	if err := ch.ExchangeDeclare(
		cfg.Exchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	return &AccountPublisher{
		cfg:     cfg,
		client:  client,
		channel: ch,
	}, nil
}

func (p *AccountPublisher) PublishAccountCreate(ctx context.Context, accountID uuid.UUID) error {
	event := rabbitmq.AccountCreatedEvent{AccountID: accountID}

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	if err = p.channel.PublishWithContext(
		ctx,
		p.cfg.Exchange,
		p.cfg.RoutingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	); err != nil {
		return err
	}

	return nil
}

func (p *AccountPublisher) Close() error {
	if p.channel != nil {
		if err := p.channel.Close(); err != nil {
			return fmt.Errorf("failed to close rabbitmq channel: %w", err)
		}
	}
	return nil
}
