package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/maket12/ads-service/backend/adservice/internal/domain/model"
	pkgrabbitmq "github.com/maket12/ads-service/backend/authservice/pkg/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
)

type PublisherConfig struct {
	Exchange        string
	AdPublishedRKey string
	AdUpdatedRKey   string
	AdRejectedRKey  string
	AdDeletedRKey   string
}

func NewPublisherConfig(exchange,
	adPublishedRK, adUpdatedRK,
	adRejectedRKey, adDeletedRK string,
) *PublisherConfig {
	return &PublisherConfig{
		Exchange:        exchange,
		AdPublishedRKey: adPublishedRK,
		AdUpdatedRKey:   adUpdatedRK,
		AdRejectedRKey:  adRejectedRKey,
		AdDeletedRKey:   adDeletedRK,
	}
}

type AdPublisher struct {
	cfg     *PublisherConfig
	client  *pkgrabbitmq.Client
	channel *amqp.Channel
}

func NewAdPublisher(
	cfg *PublisherConfig,
	client *pkgrabbitmq.Client,
) (*AdPublisher, error) {
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

	return &AdPublisher{
		cfg:     cfg,
		client:  client,
		channel: ch,
	}, nil
}

func (p *AdPublisher) publish(ctx context.Context, routingKey string, body []byte) error {
	return p.channel.PublishWithContext(ctx,
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

func (p *AdPublisher) PublishAdPublished(ctx context.Context, ad *model.Ad) error {
	body, err := json.Marshal(MapAdToAdPublishedEvent(ad))
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	return p.publish(ctx, p.cfg.AdPublishedRKey, body)
}

func (p *AdPublisher) PublishAdUpdated(ctx context.Context, ad *model.Ad) error {
	body, err := json.Marshal(MapAdToAdUpdatedEvent(ad))
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	return p.publish(ctx, p.cfg.AdUpdatedRKey, body)
}

func (p *AdPublisher) PublishAdRejected(ctx context.Context, adID uuid.UUID) error {
	body, err := json.Marshal(AdRejectedEvent{ID: adID.String()})
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	return p.publish(ctx, p.cfg.AdRejectedRKey, body)
}

func (p *AdPublisher) PublishAdDeleted(ctx context.Context, adID uuid.UUID) error {
	body, err := json.Marshal(AdDeletedEvent{ID: adID.String()})
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	return p.publish(ctx, p.cfg.AdDeletedRKey, body)
}

func (p *AdPublisher) Close() error {
	if p.channel != nil {
		if err := p.channel.Close(); err != nil {
			return fmt.Errorf("failed to close rabbitmq channel: %w", err)
		}
	}
	return nil
}
