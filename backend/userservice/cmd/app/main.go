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

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/maket12/ads-service/userservice/cmd/app/config"
	adaptergrpc "github.com/maket12/ads-service/userservice/internal/adapter/in/grpc"
	adapterrabbitmq "github.com/maket12/ads-service/userservice/internal/adapter/in/rabbitmq"
	adapterpostgres "github.com/maket12/ads-service/userservice/internal/adapter/out/postgres"
	"github.com/maket12/ads-service/userservice/internal/app/usecase"
	adapterphone "github.com/maket12/ads-service/userservice/internal/infrastructure/phone"
	"github.com/maket12/ads-service/userservice/pkg/generated/user_v1"
	pkgpostgres "github.com/maket12/ads-service/userservice/pkg/postgres"
	pkgrabbitmq "github.com/maket12/ads-service/userservice/pkg/rabbitmq"

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

func newPostgresClient(ctx context.Context, cfg *config.Config) (*pkgpostgres.Client, error) {
	pgConfig := pkgpostgres.NewConfig(
		cfg.DbHost, cfg.DbPort, cfg.DbUser, cfg.DbPassword,
		cfg.DbName, cfg.DbSSLMode, cfg.DbMaxConn,
		cfg.DbMinConn, cfg.DbMaxConnLifeTime,
		cfg.DbMaxConnIdleTime,
	)

	pgClient, err := pkgpostgres.NewClient(ctx, pgConfig)
	if err != nil {
		return nil, err
	}

	return pgClient, nil
}

func closePostgresClient(
	ctx context.Context,
	logger *slog.Logger,
	client *pkgpostgres.Client,
) {
	logger.InfoContext(ctx, "closing postgres connection...")
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
	createProfileUC *usecase.CreateProfileUC,
	deleteProfileUC *usecase.DeleteProfileUC,
) *adapterrabbitmq.AccountSubscriber {
	subConfig := adapterrabbitmq.NewSubscriberConfig(
		cfg.ExchangeName,
		cfg.QueueName,
		cfg.RoutingKey,
	)

	sub := adapterrabbitmq.NewAccountSubscriber(
		subConfig,
		logger,
		rabbitClient,
		createProfileUC,
		deleteProfileUC,
	)

	return sub
}

func runServer(ctx context.Context, cfg *config.Config, logger *slog.Logger) error {
	// Postgres client
	pgClient, err := newPostgresClient(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to init postgres client: %w", err)
	}

	// Close Postgres
	defer closePostgresClient(ctx, logger, pgClient)

	// RabbitMQ client
	rabbitClient, err := newRabbitMQClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to init rabbitmq client: %w", err)
	}

	// Close RabbitMQ
	defer closeRabbitMQClient(ctx, logger, rabbitClient)

	// Repositories and infrastructure
	profileRepo := adapterpostgres.NewProfileRepository(pgClient, trmpgx.DefaultCtxGetter)
	phoneValidator := adapterphone.NewValidator(cfg.PhoneDefaultRegion)

	// Use-cases
	createProfileUC := usecase.NewCreateProfileUC(profileRepo)
	getProfileUC := usecase.NewGetProfileUC(profileRepo)
	updateProfileUC := usecase.NewUpdateProfileUC(profileRepo, phoneValidator)
	deleteProfileUC := usecase.NewDeleteProfileUC(profileRepo)

	// RabbitMQ Subscriber
	subscriber := newRabbitMQSubscriber(cfg, logger,
		rabbitClient,
		createProfileUC,
		deleteProfileUC,
	)

	// Handler
	userHandler := adaptergrpc.NewUserHandler(
		logger,
		getProfileUC,
		updateProfileUC,
	)

	// gRPC server
	gRPCServer := grpc.NewServer()
	user_v1.RegisterUserServiceServer(gRPCServer, userHandler)
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
			ctx, "userservice failed", slog.Any("error", err),
		)
		os.Exit(1)
	}

	logger.Info("userservice stopped successfully")
}
