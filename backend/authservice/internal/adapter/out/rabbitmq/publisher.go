package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	pkgrabbitmq "github.com/maket12/ads-service/authservice/pkg/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
)

type PublisherConfig struct {
	Exchange string
}

func NewPublisherConfig(exchange string) *PublisherConfig {
	return &PublisherConfig{Exchange: exchange}
}

const (
	RoutingKeyAccountCreated = "account.created"
	RoutingKeyAccountDeleted = "account.deleted"
)

type AccountPublisher struct {
	cfg     *PublisherConfig
	client  *pkgrabbitmq.Client
	channel *amqp.Channel
}

func NewAccountPublisher(
	cfg *PublisherConfig,
	client *pkgrabbitmq.Client,
) (*AccountPublisher, error) {
	ch, err := client.Conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Exchange
	if err = ch.ExchangeDeclare(
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

func (p *AccountPublisher) publish(ctx context.Context, routingKey string, body []byte) error {
	return p.channel.PublishWithContext(
		ctx,
		p.cfg.Exchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

func (p *AccountPublisher) PublishAccountCreate(ctx context.Context, accountID uuid.UUID) error {
	event := pkgrabbitmq.AccountCreatedEvent{AccountID: accountID}

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	return p.publish(ctx, RoutingKeyAccountCreated, body)
}

func (p *AccountPublisher) PublishAccountDelete(ctx context.Context, accountID uuid.UUID) error {
	event := pkgrabbitmq.AccountDeletedEvent{AccountID: accountID}

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	return p.publish(ctx, RoutingKeyAccountDeleted, body)
}

func (p *AccountPublisher) Close() error {
	if p.channel != nil {
		if err := p.channel.Close(); err != nil {
			return fmt.Errorf("failed to close rabbitmq channel: %w", err)
		}
	}
	return nil
}
