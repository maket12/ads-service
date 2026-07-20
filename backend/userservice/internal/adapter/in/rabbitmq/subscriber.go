package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/maket12/ads-service/userservice/internal/app/dto"
	"github.com/maket12/ads-service/userservice/internal/app/usecase"
	pkgrabbitmq "github.com/maket12/ads-service/userservice/pkg/rabbitmq"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SubscriberConfig struct {
	Exchange    string
	Queue       string
	RoutingKeys []string
}

func NewSubscriberConfig(exchange, queue string, routingKey ...string) *SubscriberConfig {
	return &SubscriberConfig{
		Exchange:    exchange,
		Queue:       queue,
		RoutingKeys: routingKey,
	}
}

const (
	RoutingKeyAccountCreated = "account.created"
	RoutingKeyAccountDeleted = "account.deleted"
)

type AccountSubscriber struct {
	cfg      *SubscriberConfig
	log      *slog.Logger
	client   *pkgrabbitmq.Client
	createUC *usecase.CreateProfileUC
	deleteUC *usecase.DeleteProfileUC
}

func NewAccountSubscriber(
	cfg *SubscriberConfig,
	log *slog.Logger,
	client *pkgrabbitmq.Client,
	createUC *usecase.CreateProfileUC,
	deleteUC *usecase.DeleteProfileUC,
) *AccountSubscriber {
	return &AccountSubscriber{
		cfg:      cfg,
		log:      log,
		client:   client,
		createUC: createUC,
		deleteUC: deleteUC,
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

	// Bind queues
	for _, rk := range s.cfg.RoutingKeys {
		if err = ch.QueueBind(q.Name, rk,
			s.cfg.Exchange,
			false,
			nil,
		); err != nil {
			return fmt.Errorf("failed to bind queue to %s: %w", rk, err)
		}
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
	switch d.RoutingKey {
	case RoutingKeyAccountCreated:
		s.handleAccountCreated(ctx, d)
	case RoutingKeyAccountDeleted:
		s.handleAccountDeleted(ctx, d)
	default:
		s.log.WarnContext(ctx, "unknown routing key",
			slog.String("routing_key", d.RoutingKey),
		)
		_ = d.Nack(false, false)
	}
}

func (s *AccountSubscriber) handleAccountCreated(ctx context.Context, d *amqp.Delivery) {
	// Deserialisation json to DTO
	var event pkgrabbitmq.AccountCreatedEvent
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

func (s *AccountSubscriber) handleAccountDeleted(ctx context.Context, d *amqp.Delivery) {
	var event pkgrabbitmq.AccountDeletedEvent
	if err := json.Unmarshal(d.Body, &event); err != nil {
		s.log.ErrorContext(ctx, "failed to unmarshal account deleted event",
			slog.String("body", string(d.Body)), slog.Any("reason", err))
		_ = d.Nack(false, false)
		return
	}

	if _, err := s.deleteUC.Execute(ctx,
		dto.DeleteProfileInput{AccountID: event.AccountID},
	); err != nil {
		s.log.ErrorContext(ctx, "failed to delete profile from event",
			slog.String("account_id", event.AccountID.String()),
			slog.Any("reason", err),
		)
		_ = d.Nack(false, true)
		return
	}

	s.log.InfoContext(ctx, "deleted profile from event", slog.String("account_id", event.AccountID.String()))
	_ = d.Ack(false)
}
