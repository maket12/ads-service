///go:build e2e

package e2e

import (
	"context"
	"log/slog"
	"net"
	"os"
	"sync"
	"testing"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"github.com/maket12/ads-service/authservice/cmd/app/config"
	adaptergrpc "github.com/maket12/ads-service/authservice/internal/adapter/in/grpc"
	adapterpg "github.com/maket12/ads-service/authservice/internal/adapter/out/postgres"
	adaptermq "github.com/maket12/ads-service/authservice/internal/adapter/out/rabbitmq"
	adapterredis "github.com/maket12/ads-service/authservice/internal/adapter/out/redis"
	adapteryamail "github.com/maket12/ads-service/authservice/internal/adapter/out/yandexmail"
	"github.com/maket12/ads-service/authservice/internal/app/usecase"
	infrajwt "github.com/maket12/ads-service/authservice/internal/infrastructure/jwt"
	infrapassw "github.com/maket12/ads-service/authservice/internal/infrastructure/password"
	"github.com/maket12/ads-service/authservice/internal/testhelpers"
	"github.com/maket12/ads-service/authservice/pkg/generated/auth_v1"
	pkgpostgres "github.com/maket12/ads-service/authservice/pkg/postgres"
	pkgrabbitmq "github.com/maket12/ads-service/authservice/pkg/rabbitmq"
	pkgredis "github.com/maket12/ads-service/authservice/pkg/redis"
)

const bufSize = 1024 * 1024

type testApp struct {
	client   auth_v1.AuthServiceClient
	conn     *grpc.ClientConn
	pg       *testhelpers.PostgresContainer
	redisC   *testhelpers.RedisContainer
	rabbitC  *testhelpers.RabbitMQContainer
	mailpitC *testhelpers.MailpitContainer
	dbClient *pkgpostgres.Client
	cfg      *config.Config
}

var (
	appInstance *testApp
	once        sync.Once
)

func newLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
}

func setupE2E(t *testing.T) *testApp {
	once.Do(func() {
		ctx := context.Background()

		cfg, err := config.LoadTest()
		require.NoError(t, err)

		// --- containers ---
		pg, err := testhelpers.StartPostgresContainer(ctx)
		require.NoError(t, err)
		require.NoError(t, pg.MigrateUp())

		redisC, err := testhelpers.StartRedisContainer(ctx)
		require.NoError(t, err)

		rabbitC, err := testhelpers.StartRabbitMQContainer(ctx)
		require.NoError(t, err)

		mailpitC, err := testhelpers.StartMailpitContainer(ctx)
		require.NoError(t, err)

		cfg.DbHost, cfg.DbPort = pg.Config.Host, pg.Config.Port
		cfg.DbUser, cfg.DbPassword, cfg.DbName = pg.Config.User, pg.Config.Password, pg.Config.Name

		cfg.RedisHost, cfg.RedisPort = redisC.Host, redisC.Port
		cfg.RabbitHost, cfg.RabbitPort = rabbitC.Host, rabbitC.Port
		cfg.SMTPHost, cfg.SMTPPort = mailpitC.SMTPHost, mailpitC.SMTPPort

		logger := newLogger()

		pgClient, err := pkgpostgres.NewClient(ctx, pkgpostgres.NewConfig(
			cfg.DbHost, cfg.DbPort, cfg.DbUser, cfg.DbPassword,
			cfg.DbName, cfg.DbSSLMode, cfg.DbMaxConn,
			cfg.DbMinConn, cfg.DbMaxConnLifeTime, cfg.DbMaxConnIdleTime,
		))
		require.NoError(t, err)

		redisClient, err := pkgredis.NewClient(ctx, pkgredis.NewConfig(
			cfg.RedisHost, cfg.RedisPort, cfg.RedisPassword, cfg.RedisDBNumber,
			cfg.RedisPoolSize, cfg.RedisMinIdleConn,
			cfg.RedisDialTimeout, cfg.RedisReadTimeout, cfg.RedisWriteTimeout,
		))
		require.NoError(t, err)

		rabbitClient, err := pkgrabbitmq.NewClient(pkgrabbitmq.NewConfig(
			cfg.RabbitHost, cfg.RabbitPort, cfg.RabbitUser, cfg.RabbitPassword,
			cfg.RabbitWaitTime, cfg.RabbitAttempts,
		))
		require.NoError(t, err)

		trManager := manager.Must(trmpgx.NewDefaultFactory(pgClient.Pool))

		// repositories
		accRepo := adapterpg.NewAccountsRepository(pgClient, trmpgx.DefaultCtxGetter)
		accRoleRepo := adapterpg.NewAccountRolesRepository(pgClient, trmpgx.DefaultCtxGetter)
		rSessRepo := adapterpg.NewRefreshSessionsRepository(pgClient, trmpgx.DefaultCtxGetter)
		vTokenRepo := adapterredis.NewVerificationTokenRepository(redisClient)

		smtpClient := adapteryamail.NewSmtpClient(
			cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPEmail, cfg.SMTPPassword, cfg.VerificationBaseURL,
		)

		passwordHasher := infrapassw.NewHasher(cfg.PasswordCost)
		tokenGenerator := infrajwt.NewGenerator(
			cfg.AccessSecret, cfg.RefreshSecret, cfg.AccessTTL, cfg.RefreshTTL,
		)

		accountPublisher, err := adaptermq.NewAccountPublisher(
			adaptermq.NewPublisherConfig(cfg.ExchangeName, cfg.RoutingKey), rabbitClient,
		)
		require.NoError(t, err)

		// use-cases
		registerUC := usecase.NewRegisterUC(trManager, accRepo, accRoleRepo, passwordHasher, accountPublisher)
		loginUC := usecase.NewLoginUC(accRepo, accRoleRepo, rSessRepo, passwordHasher, tokenGenerator, cfg.RefreshTTL)
		logoutUC := usecase.NewLogoutUC(rSessRepo, tokenGenerator)
		refreshSessionUC := usecase.NewRefreshSessionUC(accRoleRepo, rSessRepo, tokenGenerator, cfg.RefreshTTL)
		validateAccessUC := usecase.NewValidateAccessTokenUC(accRepo, tokenGenerator)
		assignRoleUC := usecase.NewAssignRoleUC(accRoleRepo, rSessRepo)
		sendVerificationUC := usecase.NewSendVerificationUC(accRepo, vTokenRepo, smtpClient, cfg.VerificationTTL)
		verifyEmailUC := usecase.NewVerifyEmailUC(accRepo, vTokenRepo, smtpClient)

		handler := adaptergrpc.NewAuthHandler(
			logger, registerUC, loginUC, logoutUC, refreshSessionUC,
			validateAccessUC, assignRoleUC, sendVerificationUC, verifyEmailUC,
		)

		// --- in-memory gRPC server via bufconn ---
		lis := bufconn.Listen(bufSize)
		grpcServer := grpc.NewServer()
		auth_v1.RegisterAuthServiceServer(grpcServer, handler)

		go func() {
			_ = grpcServer.Serve(lis)
		}()

		dialer := func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}

		conn, err := grpc.NewClient(
			"passthrough:///bufnet",
			grpc.WithContextDialer(dialer),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		require.NoError(t, err)

		appInstance = &testApp{
			client:   auth_v1.NewAuthServiceClient(conn),
			conn:     conn,
			pg:       pg,
			redisC:   redisC,
			rabbitC:  rabbitC,
			mailpitC: mailpitC,
			dbClient: pgClient,
			cfg:      cfg,
		}
	})

	appInstance.cleanData(t, context.Background())

	return appInstance
}

func (a *testApp) cleanData(t *testing.T, ctx context.Context) {
	err := a.pg.TruncateTables(ctx)
	require.NoError(t, err, "failed to truncate tables")
}
