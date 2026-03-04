package main

import (
	"context"
	"fmt"
	"github.com/maket12/ads-service/pkg/generated/user_v1"
	pkgpostgres "github.com/maket12/ads-service/pkg/postgres"
	pkgrabbitmq "github.com/maket12/ads-service/pkg/rabbitmq"
	"github.com/maket12/ads-service/userservice/cmd/app/config"
	adaptergrpc "github.com/maket12/ads-service/userservice/internal/adapter/in/grpc"
	adapterrabbitmq "github.com/maket12/ads-service/userservice/internal/adapter/in/rabbitmq"
	adapterpostgres "github.com/maket12/ads-service/userservice/internal/adapter/out/postgres"
	adapterphone "github.com/maket12/ads-service/userservice/internal/adapter/out/validator"
	"github.com/maket12/ads-service/userservice/internal/app/usecase"
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

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

func newPostgresClient(cfg *config.Config) (*pkgpostgres.Client, error) {
	pgConfig := pkgpostgres.NewConfig(
		cfg.PgHost, cfg.PgPort, cfg.PgUser, cfg.PgPassword,
		cfg.PgDBName, cfg.PgSSLMode, cfg.PgOpenConn,
		cfg.PgIdleConn, cfg.PgConnLifeTime,
	)

	pgClient, err := pkgpostgres.NewClient(pgConfig)
	if err != nil {
		return nil, err
	}

	return pgClient, nil
}

func newRabbitMQClient(cfg *config.Config) (*pkgrabbitmq.RabbitClient, error) {
	rabbitConfig := pkgrabbitmq.NewRabbitConfig(
		cfg.RabbitHost,
		cfg.RabbitPort,
		cfg.RabbitUser,
		cfg.RabbitPassword,
		cfg.RabbitWaitTime,
		cfg.RabbitAttempts,
	)

	rabbitClient, err := pkgrabbitmq.NewRabbitClient(rabbitConfig)
	if err != nil {
		return nil, err
	}

	return rabbitClient, nil
}

func newRabbitMQSubscriber(
	cfg *config.Config,
	logger *slog.Logger,
	rabbitClient *pkgrabbitmq.RabbitClient,
	createProfileUC *usecase.CreateProfileUC,
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
	)

	return sub
}

func runServer(ctx context.Context, cfg *config.Config, logger *slog.Logger) error {
	// Postgres client
	pgClient, err := newPostgresClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to init postgres client: %w", err)
	}

	// Close Postgres
	defer func() {
		logger.InfoContext(ctx, "closing postgres connection...")
		if err := pgClient.Close(); err != nil {
			logger.ErrorContext(ctx, "failed to close postgres",
				slog.Any("error", err),
			)
		}
	}()

	// RabbitMQ client
	rabbitClient, err := newRabbitMQClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to init rabbitmq client: %w", err)
	}

	// Close RabbitMQ
	defer func() {
		logger.InfoContext(ctx, "closing rabbitmq connection...")
		if err := rabbitClient.Close(); err != nil {
			logger.ErrorContext(ctx, "failed to close rabbitmq",
				slog.Any("error", err),
			)
		}
	}()

	// Repositories
	profileRepo := adapterpostgres.NewProfileRepository(pgClient)
	phoneValidator := adapterphone.NewPhoneValidator(cfg.PhoneDefaultRegion)

	// Use-cases
	createProfileUC := usecase.NewCreateProfileUC(profileRepo)
	getProfileUC := usecase.NewGetProfileUC(profileRepo)
	updateProfileUC := usecase.NewUpdateProfileUC(profileRepo, phoneValidator)

	// RabbitMQ Subscriber
	subscriber := newRabbitMQSubscriber(cfg, logger, rabbitClient, createProfileUC)

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

	if err := runServer(ctx, cfg, logger); err != nil {
		logger.ErrorContext(
			ctx, "userservice failed", slog.Any("error", err),
		)
		os.Exit(1)
	}

	logger.Info("userservice stopped successfully")
}
