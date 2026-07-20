package main

import (
	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/maket12/ads-service/authservice/cmd/app/config"
	adaptergrpc "github.com/maket12/ads-service/authservice/internal/adapter/in/grpc"
	adapterpg "github.com/maket12/ads-service/authservice/internal/adapter/out/postgres"
	adaptermq "github.com/maket12/ads-service/authservice/internal/adapter/out/rabbitmq"
	adapterredis "github.com/maket12/ads-service/authservice/internal/adapter/out/redis"
	adapteryamail "github.com/maket12/ads-service/authservice/internal/adapter/out/yandexmail"
	"github.com/maket12/ads-service/authservice/internal/app/usecase"
	infrajwt "github.com/maket12/ads-service/authservice/internal/infrastructure/jwt"
	infrapassw "github.com/maket12/ads-service/authservice/internal/infrastructure/password"
	"github.com/maket12/ads-service/authservice/pkg/generated/auth_v1"

	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	pkgpostgres "github.com/maket12/ads-service/authservice/pkg/postgres"
	pkgrabbitmq "github.com/maket12/ads-service/authservice/pkg/rabbitmq"
	pkgredis "github.com/maket12/ads-service/authservice/pkg/redis"

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

func newRedisClient(ctx context.Context, cfg *config.Config) (*pkgredis.Client, error) {
	redisConfig := pkgredis.NewConfig(
		cfg.RedisHost, cfg.RedisPort,
		cfg.RedisPassword, cfg.RedisDBNumber,
		cfg.RedisPoolSize, cfg.RedisMinIdleConn,
		cfg.RedisDialTimeout, cfg.RedisReadTimeout,
		cfg.RedisWriteTimeout,
	)

	redisClient, err := pkgredis.NewClient(ctx, redisConfig)
	if err != nil {
		return nil, err
	}

	return redisClient, nil
}

func closeRedisClient(
	ctx context.Context,
	logger *slog.Logger,
	client *pkgredis.Client,
) {
	logger.InfoContext(ctx, "closing redis connection...")
	if err := client.Close(); err != nil {
		logger.ErrorContext(ctx, "failed to close redis",
			slog.Any("error", err),
		)
	}
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

func newAccountPublisher(cfg *config.Config, rabbitClient *pkgrabbitmq.Client) (*adaptermq.AccountPublisher, error) {
	publisherConfig := adaptermq.NewPublisherConfig(cfg.ExchangeName)

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

	// Redis client
	redisClient, err := newRedisClient(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to init redis client: %w", err)
	}

	// Close Redis
	defer closeRedisClient(ctx, logger, redisClient)

	// Transaction manager
	trManager := manager.Must(trmpgx.NewDefaultFactory(pgClient.Pool))

	// Repositories
	accRepo := adapterpg.NewAccountsRepository(pgClient, trmpgx.DefaultCtxGetter)
	accRoleRepo := adapterpg.NewAccountRolesRepository(pgClient, trmpgx.DefaultCtxGetter)
	rSessRepo := adapterpg.NewRefreshSessionsRepository(pgClient, trmpgx.DefaultCtxGetter)
	vTokenRepo := adapterredis.NewVerificationTokenRepository(redisClient)

	// Email Sender
	smtpClient := adapteryamail.NewSmtpClient(
		cfg.SMTPHost, cfg.SMTPPort,
		cfg.SMTPEmail, cfg.SMTPPassword,
		cfg.VerificationBaseURL,
	)

	// Infrastructure
	passwordHasher := infrapassw.NewHasher(cfg.PasswordCost)
	tokenGenerator := infrajwt.NewGenerator(
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
		trManager, accRepo,
		accRoleRepo, passwordHasher,
		accountPublisher,
	)
	loginUC := usecase.NewLoginUC(
		trManager, accRepo, accRoleRepo,
		rSessRepo, passwordHasher,
		tokenGenerator, cfg.RefreshTTL,
	)
	logoutUC := usecase.NewLogoutUC(rSessRepo, tokenGenerator)
	refreshSessionUC := usecase.NewRefreshSessionUC(
		accRoleRepo, rSessRepo,
		tokenGenerator, cfg.RefreshTTL,
	)
	validateAccessUC := usecase.NewValidateAccessTokenUC(accRepo, tokenGenerator)
	assignRoleUC := usecase.NewAssignRoleUC(accRoleRepo, rSessRepo)
	sendVerificationUC := usecase.NewSendVerificationUC(
		accRepo, vTokenRepo,
		smtpClient, cfg.VerificationTTL,
	)
	verifyEmailUC := usecase.NewVerifyEmailUC(accRepo, vTokenRepo, smtpClient)

	// Handler
	handler := adaptergrpc.NewAuthHandler(
		logger, registerUC, loginUC,
		logoutUC, refreshSessionUC,
		validateAccessUC, assignRoleUC,
		sendVerificationUC, verifyEmailUC,
	)

	// gRPC server
	gRPCServer := grpc.NewServer()
	auth_v1.RegisterAuthServiceServer(gRPCServer, handler)
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
		if err = gRPCServer.Serve(lis); err != nil {
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

	if err = runServer(ctx, cfg, logger); err != nil {
		logger.ErrorContext(
			ctx, "authservice failed", slog.Any("error", err),
		)
		os.Exit(1)
	}

	logger.Info("authservice stopped successfully")
}
