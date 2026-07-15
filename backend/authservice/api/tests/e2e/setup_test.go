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
	"github.com/brianvoe/gofakeit/v7"
	"github.com/maket12/ads-service/authservice/internal/fakes"
	"github.com/maket12/ads-service/authservice/migrations"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"

	"github.com/maket12/ads-service/authservice/cmd/app/config"
	adaptergrpc "github.com/maket12/ads-service/authservice/internal/adapter/in/grpc"
	adapterpg "github.com/maket12/ads-service/authservice/internal/adapter/out/postgres"
	adapterredis "github.com/maket12/ads-service/authservice/internal/adapter/out/redis"
	"github.com/maket12/ads-service/authservice/internal/app/usecase"
	infrajwt "github.com/maket12/ads-service/authservice/internal/infrastructure/jwt"
	infrapassw "github.com/maket12/ads-service/authservice/internal/infrastructure/password"
	"github.com/maket12/ads-service/authservice/pkg/generated/auth_v1"
	pkgpostgres "github.com/maket12/ads-service/authservice/pkg/postgres"
	pkgrabbitmq "github.com/maket12/ads-service/authservice/pkg/rabbitmq"
	pkgredis "github.com/maket12/ads-service/authservice/pkg/redis"
)

const bufSize = 1024 * 1024

type testApp struct {
	client   auth_v1.AuthServiceClient
	conn     *grpc.ClientConn
	pg       *pkgpostgres.TestContainer
	redisC   *pkgredis.TestContainer
	rabbitC  *pkgrabbitmq.TestContainer
	email    *fakes.FakeMailSender
	dbClient *pkgpostgres.Client
	cfg      *config.TestConfig
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
		pg, err := pkgpostgres.StartTestContainer(ctx)
		require.NoError(t, err)
		require.NoError(t, pg.MigrateUp(migrations.FS, 3))

		redisC, err := pkgredis.StartTestContainer(ctx)
		require.NoError(t, err)

		rabbitC, err := pkgrabbitmq.StartTestContainer(ctx)
		require.NoError(t, err)

		cfg.DbHost, cfg.DbPort = pg.Config.Host, pg.Config.Port
		cfg.DbUser, cfg.DbPassword, cfg.DbName = pg.Config.User, pg.Config.Password, pg.Config.Name

		cfg.RedisHost, cfg.RedisPort = redisC.Config.Host, redisC.Config.Port
		cfg.RabbitHost, cfg.RabbitPort = rabbitC.Config.Host, rabbitC.Config.Port

		logger := newLogger()

		pgClient, err := pkgpostgres.NewClient(ctx, pkgpostgres.NewConfig(
			pg.Config.Host, pg.Config.Port, pg.Config.User,
			pg.Config.Password, pg.Config.Name, pg.Config.SSLMode,
			pg.Config.MaxConn, pg.Config.MinConn,
			pg.Config.MaxConnLifeTime, pg.Config.MaxConnIdleTime,
		))
		require.NoError(t, err)

		redisClient, err := pkgredis.NewClient(ctx, pkgredis.NewConfig(
			redisC.Config.Host, redisC.Config.Port,
			redisC.Config.Password, redisC.Config.DB,
			redisC.Config.PoolSize, redisC.Config.MinIdleConn,
			redisC.Config.DialTimeout, redisC.Config.ReadTimeout,
			redisC.Config.WriteTimeout,
		))
		require.NoError(t, err)

		trManager := manager.Must(trmpgx.NewDefaultFactory(pgClient.Pool))

		// repositories
		accRepo := adapterpg.NewAccountsRepository(pgClient, trmpgx.DefaultCtxGetter)
		accRoleRepo := adapterpg.NewAccountRolesRepository(pgClient, trmpgx.DefaultCtxGetter)
		rSessRepo := adapterpg.NewRefreshSessionsRepository(pgClient, trmpgx.DefaultCtxGetter)
		vTokenRepo := adapterredis.NewVerificationTokenRepository(redisClient)

		// smtp client
		smtpClient := fakes.NewFakeMailSender()

		// infrastructure
		passwordHasher := infrapassw.NewHasher(cfg.PasswordCost)
		tokenGenerator := infrajwt.NewGenerator(
			cfg.AccessSecret, cfg.RefreshSecret, cfg.AccessTTL, cfg.RefreshTTL,
		)

		// event publisher
		accountPublisher := fakes.NewFakePublisher()

		// use-cases
		registerUC := usecase.NewRegisterUC(trManager, accRepo, accRoleRepo, passwordHasher, accountPublisher)
		loginUC := usecase.NewLoginUC(trManager, accRepo, accRoleRepo, rSessRepo, passwordHasher, tokenGenerator, cfg.RefreshTTL)
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
			email:    smtpClient,
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

	err = a.redisC.FlushAll(ctx)
	require.NoError(t, err, "failed to flush all")

	a.email.Reset()
}

// Helper for e2e tests.
// Creates a new user with specified parameters by calling `Register` method.
//
// If parameters are not specified, then it uses random values instead.
//
// Returns the external account id.
func (a *testApp) createAccount(t *testing.T, email, password *string) string {
	var (
		regEmail    = gofakeit.Email()
		regPassword = gofakeit.Password(
			true, true, true,
			false, false, 12,
		)
	)

	if email != nil {
		regEmail = *email
	}
	if password != nil {
		regPassword = *password
	}

	resp, err := a.client.Register(context.Background(), &auth_v1.RegisterRequest{
		Email:    regEmail,
		Password: regPassword,
	})

	require.NoError(t, err)
	require.NotEmpty(t, resp.GetAccountId())

	return resp.GetAccountId()
}
