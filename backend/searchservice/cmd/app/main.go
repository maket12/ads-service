package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	pkgrabbitmq "github.com/maket12/ads-service/backend/authservice/pkg/rabbitmq"
	"github.com/maket12/ads-service/backend/searchservice/api/proto/generated/search_v1"
	"github.com/maket12/ads-service/backend/searchservice/cmd/app/config"
	adaptergrpc "github.com/maket12/ads-service/backend/searchservice/internal/adapter/in/grpc"
	adapterrabbitmq "github.com/maket12/ads-service/backend/searchservice/internal/adapter/in/rabbitmq"
	adapterelasticsearch "github.com/maket12/ads-service/backend/searchservice/internal/adapter/out/elasticsearch"
	"github.com/maket12/ads-service/backend/searchservice/internal/app/usecase"
	pkgelasticsearch "github.com/maket12/ads-service/backend/searchservice/pkg/elasticsearch"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func parseLogLevel(level string) slog.Level {
	switch level {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func newLogger(level string) *slog.Logger {
	logLevel := parseLogLevel(level)
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
}

func newESClient(ctx context.Context, cfg *config.Config) (*pkgelasticsearch.Client, error) {
	esConfig := pkgelasticsearch.NewConfig(cfg.ESAddresses, cfg.ESUsername, cfg.ESPassword)

	esClient, err := pkgelasticsearch.NewClient(ctx, esConfig)
	if err != nil {
		return nil, err
	}

	return esClient, nil
}

func closeESClient(
	ctx context.Context,
	logger *slog.Logger,
	client *pkgelasticsearch.Client,
) {
	logger.InfoContext(ctx, "closing elasticsearch connection...")
	client.Close()
}

func newRabbitMQClient(cfg *config.Config) (*pkgrabbitmq.Client, error) {
	rabbitConfig := pkgrabbitmq.NewConfig(
		cfg.RabbitHost,
		cfg.RabbitPort,
		cfg.RabbitUser,
		cfg.RabbitPassword,
		cfg.RabbitWaitTime,
		cfg.RabbitAttempts,
	)

	rabbitClient, err := pkgrabbitmq.NewClient(rabbitConfig)
	if err != nil {
		return nil, err
	}

	return rabbitClient, nil
}

func closeRabbitMQClient(
	ctx context.Context,
	logger *slog.Logger,
	rabbitClient *pkgrabbitmq.Client,
) {
	logger.InfoContext(ctx, "closing rabbitmq connection...")
	if err := rabbitClient.Close(); err != nil {
		logger.ErrorContext(ctx, "failed to close rabbitmq",
			slog.Any("error", err),
		)
	}
}

func newRabbitMQSubscriber(cfg *config.Config, logger *slog.Logger,
	rabbitClient *pkgrabbitmq.Client,
	createAdIndexUC *usecase.CreateAdIndexUC,
	deleteAdIndexUC *usecase.DeleteAdIndexUC,
) *adapterrabbitmq.AdSubscriber {
	subConfig := adapterrabbitmq.NewSubscriberConfig(
		cfg.AdExchange, cfg.AdQueue,
		cfg.AdPublishedRoutingKey,
		cfg.AdUpdatedRoutingKey,
		cfg.AdRejectedRoutingKey,
		cfg.AdDeletedRoutingKey,
	)

	return adapterrabbitmq.NewAdSubscriber(
		subConfig, logger, rabbitClient,
		createAdIndexUC, deleteAdIndexUC,
	)
}

func runServer(ctx context.Context, cfg *config.Config, logger *slog.Logger) error {
	// ElasticSearch client
	esClient, err := newESClient(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to init elasticsearch client: %w", err)
	}

	// Close ElasticSearch
	defer closeESClient(ctx, logger, esClient)

	// RabbitMQ client
	rabbitClient, err := newRabbitMQClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to init rabbitmq client: %w", err)
	}

	// Close RabbitMQ
	defer closeRabbitMQClient(ctx, logger, rabbitClient)

	// Repository
	adIndexRepo := adapterelasticsearch.NewAdIndexRepository(esClient, cfg.ESIndexName)

	// Use-cases
	createdAdIndexUC := usecase.NewCreateAdIndexUC(adIndexRepo)
	deleteAdIndexUC := usecase.NewDeleteAdIndexUC(adIndexRepo)
	searchAdsUC := usecase.NewSearchAdsUC(adIndexRepo)

	// RabbitMQ Subscriber
	subscriber := newRabbitMQSubscriber(cfg, logger,
		rabbitClient,
		createdAdIndexUC,
		deleteAdIndexUC,
	)

	// Handler
	searchHandler := adaptergrpc.NewSearchHandler(logger, searchAdsUC)

	// gRPC server
	gRPCServer := grpc.NewServer()
	search_v1.RegisterSearchServiceServer(gRPCServer, searchHandler)
	reflection.Register(gRPCServer)

	address := fmt.Sprintf(":%d", cfg.GRPCPort)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to listen port %d: %w",
			cfg.GRPCPort, err,
		)
	}

	errChan := make(chan error, 2)

	// Launch RabbitMQ Subscriber
	go func() {
		logger.InfoContext(ctx, "starting rabbitmq subscriber...")
		if err := subscriber.Start(ctx); err != nil {
			errChan <- fmt.Errorf("subscriber failure: %w", err)
		}
	}()

	// Launch gRPC server
	go func() {
		logger.InfoContext(ctx, "starting grpc server",
			slog.String("address", address),
		)
		if err := gRPCServer.Serve(lis); err != nil {
			errChan <- err
		}
	}()

	// Graceful shutdown
	select {
	case <-ctx.Done():
		logger.InfoContext(
			ctx, "received shutdown signal, stopping grpc server...",
		)
		gRPCServer.GracefulStop()
		return nil
	case err = <-errChan:
		return fmt.Errorf("grpc server/rabbitmq failed: %w", err)
	}
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logger := newLogger(cfg.LogLevel)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err = runServer(ctx, cfg, logger); err != nil {
		logger.ErrorContext(
			ctx, "searchservice failed", slog.Any("error", err),
		)
		os.Exit(1)
	}

	logger.Info("searchservice stopped successfully")
}
