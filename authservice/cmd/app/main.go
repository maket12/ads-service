package main

import (
	"github.com/maket12/ads-service/authservice/cmd/app/config"
	adaptergrpc "github.com/maket12/ads-service/authservice/internal/adapter/in/grpc"
	adapterph "github.com/maket12/ads-service/authservice/internal/adapter/out/hasher"
	adaptertg "github.com/maket12/ads-service/authservice/internal/adapter/out/jwt"
	adapterdb "github.com/maket12/ads-service/authservice/internal/adapter/out/postgres"
	adaptermq "github.com/maket12/ads-service/authservice/internal/adapter/out/rabbitmq"
	"github.com/maket12/ads-service/authservice/internal/app/usecase"
	"github.com/maket12/ads-service/pkg/generated/auth_v1"

	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	pkgpostgres "github.com/maket12/ads-service/pkg/postgres"
	pkgrabbitmq "github.com/maket12/ads-service/pkg/rabbitmq"

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

func closePostgresClient(
	ctx context.Context,
	logger *slog.Logger,
	pgClient *pkgpostgres.Client,
) {
	logger.InfoContext(ctx, "closing postgres connection...")
	if err := pgClient.Close(); err != nil {
		logger.ErrorContext(ctx, "failed to close postgres",
			slog.Any("error", err),
		)
	}
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

func closeRabbitMQClient(
	ctx context.Context,
	logger *slog.Logger,
	rabbitClient *pkgrabbitmq.RabbitClient,
) {
	logger.InfoContext(ctx, "closing rabbitmq connection...")
	if err := rabbitClient.Close(); err != nil {
		logger.ErrorContext(ctx, "failed to close rabbitmq",
			slog.Any("error", err),
		)
	}
}

func newAccountPublisher(
	cfg *config.Config, rabbitClient *pkgrabbitmq.RabbitClient,
) (*adaptermq.AccountPublisher, error) {
	publisherConfig := adaptermq.NewPublisherConfig(
		cfg.ExchangeName, cfg.RoutingKey,
	)

	pub, err := adaptermq.NewAccountPublisher(publisherConfig, rabbitClient)
	if err != nil {
		return nil, err
	}

	return pub, nil
}

func closeAccountPublisher(
	ctx context.Context,
	logger *slog.Logger,
	accountPublisher *adaptermq.AccountPublisher,
) {
	logger.InfoContext(ctx, "closing account publisher...")
	if err := accountPublisher.Close(); err != nil {
		logger.ErrorContext(ctx, "failed to close account publisher",
			slog.Any("error", err),
		)
	}
}

func runServer(ctx context.Context, cfg *config.Config, logger *slog.Logger) error {
	// Postgres client
	pgClient, err := newPostgresClient(cfg)
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

	// Repositories
	accountRepo := adapterdb.NewAccountsRepository(pgClient)
	accountRoleRepo := adapterdb.NewAccountRolesRepository(pgClient)
	refreshSessionRepo := adapterdb.NewRefreshSessionsRepository(pgClient)
	passwordHasher := adapterph.NewBcryptHasher(cfg.PasswordCost)
	tokenGenerator := adaptertg.NewTokenGenerator(
		cfg.AccessSecret, cfg.RefreshSecret,
		cfg.AccessTTL, cfg.RefreshTTL,
	)

	// RabbitMQ Publisher
	accountPublisher, err := newAccountPublisher(cfg, rabbitClient)
	if err != nil {
		return fmt.Errorf("failed to init event publisher: %w", err)
	}
	defer closeAccountPublisher(ctx, logger, accountPublisher)

	// Use-cases
	registerUC := usecase.NewRegisterUC(
		accountRepo, accountRoleRepo, passwordHasher, accountPublisher,
	)
	loginUC := usecase.NewLoginUC(
		accountRepo, accountRoleRepo, refreshSessionRepo,
		passwordHasher, tokenGenerator, cfg.RefreshTTL,
	)
	logoutUC := usecase.NewLogoutUC(refreshSessionRepo, tokenGenerator)
	refreshSessionUC := usecase.NewRefreshSessionUC(
		accountRoleRepo, refreshSessionRepo,
		tokenGenerator, cfg.RefreshTTL,
	)
	validateAccessUC := usecase.NewValidateAccessTokenUC(
		accountRepo, tokenGenerator,
	)
	assignRoleUC := usecase.NewAssignRoleUC(accountRoleRepo)

	// Handler
	authHandler := adaptergrpc.NewAuthHandler(
		logger,
		registerUC,
		loginUC,
		logoutUC,
		refreshSessionUC,
		validateAccessUC,
		assignRoleUC,
	)

	// gRPC server
	gRPCServer := grpc.NewServer()
	auth_v1.RegisterAuthServiceServer(gRPCServer, authHandler)
	reflection.Register(gRPCServer)

	address := fmt.Sprintf(":%d", cfg.GRPCPort)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to listen port %d: %w",
			cfg.GRPCPort, err,
		)
	}

	// Launch gRPC server
	errChan := make(chan error, 1)
	go func() {
		logger.InfoContext(
			ctx, "starting grpc server", slog.String("address", address))
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
	case err := <-errChan:
		return fmt.Errorf("grpc server failed: %w", err)
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
			ctx, "authservice failed", slog.Any("error", err),
		)
		os.Exit(1)
	}

	logger.Info("authservice stopped successfully")
}
