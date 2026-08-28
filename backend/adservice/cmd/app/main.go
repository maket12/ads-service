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
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/maket12/ads-service/backend/adservice/api/proto/generated/ad_v1"
	"github.com/maket12/ads-service/backend/adservice/cmd/app/config"
	adaptergrpc "github.com/maket12/ads-service/backend/adservice/internal/adapter/in/grpc"
	adaptermongo "github.com/maket12/ads-service/backend/adservice/internal/adapter/out/mongodb"
	adapterpg "github.com/maket12/ads-service/backend/adservice/internal/adapter/out/postgres"
	"github.com/maket12/ads-service/backend/adservice/internal/app/usecase"
	pkgmongodb "github.com/maket12/ads-service/backend/adservice/pkg/mongodb"
	pkgpostgres "github.com/maket12/ads-service/backend/authservice/pkg/postgres"

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

func newMongoClient(ctx context.Context, cfg *config.Config) (*pkgmongodb.Client, error) {
	mongoConfig := pkgmongodb.NewConfig(
		cfg.MongoHost, cfg.MongoPort, cfg.MongoUser,
		cfg.MongoPassword, cfg.MongoDBName,
	)

	mongoClient, err := pkgmongodb.NewClient(ctx, mongoConfig)
	if err != nil {
		return nil, err
	}

	return mongoClient, nil
}

func closeMongoClient(
	ctx context.Context,
	logger *slog.Logger,
	mongoClient *pkgmongodb.Client,
) {
	logger.InfoContext(ctx, "closing mongo connection...")
	if err := mongoClient.Close(ctx); err != nil {
		logger.ErrorContext(ctx, "failed to close mongo",
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

	// Mongo client
	mongoClient, err := newMongoClient(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to init mongo client: %w", err)
	}

	// Close Mongo
	defer closeMongoClient(ctx, logger, mongoClient)

	// Media repository config
	mediaRepoCfg := adaptermongo.NewMediaRepositoryConfig(
		mongoClient,
		cfg.MongoCollectionName,
	)

	// Transaction manager
	trManager := manager.Must(trmpgx.NewDefaultFactory(pgClient.Pool))

	// Repositories
	adRepo := adapterpg.NewAdRepository(pgClient, trmpgx.DefaultCtxGetter)
	mediaRepo := adaptermongo.NewMediaRepository(mediaRepoCfg)

	// Use-cases
	createAdUC := usecase.NewCreateAdUC(trManager, adRepo, mediaRepo)
	getAdUC := usecase.NewGetAdUC(adRepo, mediaRepo)
	updateAdUC := usecase.NewUpdateAdUC(trManager, adRepo, mediaRepo)
	publishAdUC := usecase.NewPublishAdUC(adRepo)
	rejectAdUC := usecase.NewRejectAdUC(adRepo)
	deleteAdUC := usecase.NewDeleteAdUC(trManager, adRepo, mediaRepo)
	deleteAllAdsUC := usecase.NewDeleteAllAdsUC(trManager, adRepo, mediaRepo)
	listSellerAdsUC := usecase.NewListSellerAdsUC(adRepo)
	listAllAdsUC := usecase.NewListAllAdsUC(adRepo)

	// Handler
	adHandler := adaptergrpc.NewAdHandler(
		logger, createAdUC, getAdUC,
		updateAdUC, publishAdUC, rejectAdUC,
		deleteAdUC, deleteAllAdsUC,
		listSellerAdsUC, listAllAdsUC,
	)

	// gRPC server
	gRPCServer := grpc.NewServer()
	ad_v1.RegisterAdServiceServer(gRPCServer, adHandler)
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
			ctx, "adservice failed", slog.Any("error", err),
		)
		os.Exit(1)
	}

	logger.Info("adservice stopped successfully")
}
