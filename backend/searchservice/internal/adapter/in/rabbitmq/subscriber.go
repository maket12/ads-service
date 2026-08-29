package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	pkgrabbitmq "github.com/maket12/ads-service/backend/authservice/pkg/rabbitmq"
	"github.com/maket12/ads-service/backend/searchservice/internal/app/dto"
	"github.com/maket12/ads-service/backend/searchservice/internal/app/usecase"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SubscriberConfig struct {
	Exchange        string
	Queue           string
	AdPublishedRKey string
	AdUpdatedRKey   string
	AdRejectedRKey  string
	AdDeletedRKey   string
}

func NewSubscriberConfig(
	exchange, queue,
	adPublishedRK, adUpdatedRK,
	adRejectedRKey, adDeletedRK string,
) *SubscriberConfig {
	return &SubscriberConfig{
		Exchange:        exchange,
		Queue:           queue,
		AdPublishedRKey: adPublishedRK,
		AdUpdatedRKey:   adUpdatedRK,
		AdRejectedRKey:  adRejectedRKey,
		AdDeletedRKey:   adDeletedRK,
	}
}

type AdSubscriber struct {
	cfg      *SubscriberConfig
	log      *slog.Logger
	client   *pkgrabbitmq.Client
	createUC *usecase.CreateAdIndexUC
	deleteUC *usecase.DeleteAdIndexUC
}

func NewAdSubscriber(
	cfg *SubscriberConfig,
	log *slog.Logger,
	client *pkgrabbitmq.Client,
	createUC *usecase.CreateAdIndexUC,
	deleteUC *usecase.DeleteAdIndexUC,
) *AdSubscriber {
	return &AdSubscriber{
		cfg:      cfg,
		log:      log,
		client:   client,
		createUC: createUC,
		deleteUC: deleteUC,
	}
}

func (s *AdSubscriber) Start(ctx context.Context) error {
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
	rkeys := []string{
		s.cfg.AdPublishedRKey, s.cfg.AdUpdatedRKey,
		s.cfg.AdRejectedRKey, s.cfg.AdDeletedRKey,
	}
	for _, rk := range rkeys {
		if err = ch.QueueBind(q.Name, rk,
			s.cfg.Exchange,
			false,
			nil,
		); err != nil {
			return fmt.Errorf("failed to bind queue to %s: %w", rk, err)
		}
	}

	// Define consumer
	messages, err := ch.ConsumeWithContext(
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
		for d := range messages {
			s.handleMessage(ctx, &d)
		}
	}()

	return nil
}

func (s *AdSubscriber) handleMessage(ctx context.Context, d *amqp.Delivery) {
	switch d.RoutingKey {
	case s.cfg.AdPublishedRKey:
		s.handleAdPublished(ctx, d)
	case s.cfg.AdUpdatedRKey:
		s.handleAdUpdated(ctx, d)
	case s.cfg.AdRejectedRKey:
		s.handleAdRejected(ctx, d)
	case s.cfg.AdDeletedRKey:
		s.handleAdDeleted(ctx, d)
	default:
		s.log.WarnContext(ctx, "unknown routing key",
			slog.String("routing_key", d.RoutingKey),
		)
		_ = d.Nack(false, false)
	}
}

func (s *AdSubscriber) handleAdPublished(ctx context.Context, d *amqp.Delivery) {
	var event AdPublishedEvent
	if err := json.Unmarshal(d.Body, &event); err != nil {
		s.log.ErrorContext(ctx, "failed to unmarshal ad event",
			slog.String("body", string(d.Body)),
			slog.Any("reason", err),
		)
		_ = d.Nack(false, false)
		return
	}

	if err := s.createUC.Execute(ctx, MapAdPublishedEventToInput(event)); err != nil {
		s.log.ErrorContext(ctx, "failed to create ad index (caused by rabbitmq)",
			slog.String("ad_id", event.ID),
			slog.Any("reason", err),
		)
		_ = d.Nack(false, true)
		return
	}

	s.log.InfoContext(ctx, "created ad index (caused by rabbitmq)",
		slog.String("ad_id", event.ID),
		slog.String("title", event.Title),
		slog.String("description", event.Description),
		slog.Int64("price", event.Price),
		slog.String("category", event.Category),
		slog.String("main_image", event.MainImage),
	)
	_ = d.Ack(false)
}

func (s *AdSubscriber) handleAdUpdated(ctx context.Context, d *amqp.Delivery) {
	var event AdUpdatedEvent
	if err := json.Unmarshal(d.Body, &event); err != nil {
		s.log.ErrorContext(ctx, "failed to unmarshal ad event",
			slog.String("body", string(d.Body)),
			slog.Any("reason", err),
		)
		_ = d.Nack(false, false)
		return
	}

	if err := s.deleteUC.Execute(ctx,
		dto.DeleteAdIndexInput{ID: event.ID},
	); err != nil {
		s.log.ErrorContext(ctx, "failed to delete ad index (caused by rabbitmq)",
			slog.String("ad_id", event.ID),
			slog.Any("reason", err),
		)
		_ = d.Nack(false, true)
		return
	}

	s.log.InfoContext(ctx, "deleted ad index (caused by rabbitmq)",
		slog.String("ad_id", event.ID),
	)
	_ = d.Ack(false)
}

func (s *AdSubscriber) handleAdRejected(ctx context.Context, d *amqp.Delivery) {
	var event AdRejectedEvent
	if err := json.Unmarshal(d.Body, &event); err != nil {
		s.log.ErrorContext(ctx, "failed to unmarshal ad event",
			slog.String("body", string(d.Body)),
			slog.Any("reason", err),
		)
		_ = d.Nack(false, false)
		return
	}

	if err := s.deleteUC.Execute(ctx,
		dto.DeleteAdIndexInput{ID: event.ID},
	); err != nil {
		s.log.ErrorContext(ctx, "failed to delete ad index (caused by rabbitmq)",
			slog.String("ad_id", event.ID),
			slog.Any("reason", err),
		)
		_ = d.Nack(false, true)
		return
	}

	s.log.InfoContext(ctx, "deleted ad index (caused by rabbitmq)",
		slog.String("ad_id", event.ID),
	)
	_ = d.Ack(false)
}

func (s *AdSubscriber) handleAdDeleted(ctx context.Context, d *amqp.Delivery) {
	var event AdDeletedEvent
	if err := json.Unmarshal(d.Body, &event); err != nil {
		s.log.ErrorContext(ctx, "failed to unmarshal ad event",
			slog.String("body", string(d.Body)),
			slog.Any("reason", err),
		)
		_ = d.Nack(false, false)
		return
	}

	if err := s.deleteUC.Execute(ctx,
		dto.DeleteAdIndexInput{ID: event.ID},
	); err != nil {
		s.log.ErrorContext(ctx, "failed to delete ad index (caused by rabbitmq)",
			slog.String("ad_id", event.ID),
			slog.Any("reason", err),
		)
		_ = d.Nack(false, true)
		return
	}

	s.log.InfoContext(ctx, "deleted ad index (caused by rabbitmq)",
		slog.String("ad_id", event.ID),
	)
	_ = d.Ack(false)
}
