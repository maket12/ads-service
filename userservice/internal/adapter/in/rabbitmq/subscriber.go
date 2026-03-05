package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/maket12/ads-service/pkg/rabbitmq"
	"github.com/maket12/ads-service/userservice/internal/app/dto"
	"github.com/maket12/ads-service/userservice/internal/app/usecase"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SubscriberConfig struct {
	Exchange   string
	Queue      string
	RoutingKey string
}

func NewSubscriberConfig(exchange, queue, routingKey string) *SubscriberConfig {
	return &SubscriberConfig{
		Exchange:   exchange,
		Queue:      queue,
		RoutingKey: routingKey,
	}
}

type AccountSubscriber struct {
	cfg      *SubscriberConfig
	log      *slog.Logger
	client   *rabbitmq.RabbitClient
	createUC *usecase.CreateProfileUC
}

func NewAccountSubscriber(
	cfg *SubscriberConfig,
	log *slog.Logger,
	client *rabbitmq.RabbitClient,
	createUC *usecase.CreateProfileUC,
) *AccountSubscriber {
	return &AccountSubscriber{
		cfg:      cfg,
		log:      log,
		client:   client,
		createUC: createUC,
	}
}

func (s *AccountSubscriber) Start(ctx context.Context) error {
	ch, err := s.client.Conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}

	// Exchange
	if err = ch.ExchangeDeclare(
		s.cfg.Exchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Queue
	q, err := ch.QueueDeclare(
		s.cfg.Queue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue
	if err = ch.QueueBind(
		q.Name,
		s.cfg.RoutingKey,
		s.cfg.Exchange,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	// Define consumer
	msgs, err := ch.ConsumeWithContext(
		ctx,
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to define consumer: %w", err)
	}

	// Listening to queue
	go func() {
		for d := range msgs {
			s.handleMessage(ctx, &d)
		}
	}()

	return nil
}

func (s *AccountSubscriber) handleMessage(ctx context.Context, d *amqp.Delivery) {
	// Deserialisation json to DTO
	var event rabbitmq.AccountCreatedEvent
	if err := json.Unmarshal(d.Body, &event); err != nil {
		s.log.ErrorContext(ctx, "failed to unmarshal account event",
			slog.String("body", string(d.Body)),
			slog.Any("reason", err),
		)
		_ = d.Nack(false, false)
		return
	}

	// Calling UC
	if err := s.createUC.Execute(
		ctx, dto.CreateProfileInput{AccountID: event.AccountID},
	); err != nil {
		s.log.ErrorContext(ctx, "failed to create profile from event",
			slog.String("account_id", event.AccountID.String()),
			slog.Any("reason", err),
		)
		_ = d.Nack(false, true)
		return
	}

	// Notify queue about success
	s.log.InfoContext(ctx, "created profile from event",
		slog.String("account_id", event.AccountID.String()),
	)
	_ = d.Ack(false)
}
