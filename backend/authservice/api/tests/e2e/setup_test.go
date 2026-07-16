///go:build e2e

package e2e

import (
	"context"
	"log/slog"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/maket12/ads-service/authservice/internal/app/dto"
	"github.com/maket12/ads-service/authservice/internal/domain/model"
	"github.com/maket12/ads-service/authservice/internal/domain/port"
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
	client      auth_v1.AuthServiceClient
	conn        *grpc.ClientConn
	pg          *pkgpostgres.TestContainer
	redisC      *pkgredis.TestContainer
	rabbitC     *pkgrabbitmq.TestContainer
	email       *fakes.FakeMailSender
	accRepo     port.AccountRepository
	accRoleRepo port.AccountRoleRepository
	tokenRepo   port.VerificationTokenRepository
	dbClient    *pkgpostgres.Client
	cfg         *config.TestConfig
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
			client:      auth_v1.NewAuthServiceClient(conn),
			conn:        conn,
			pg:          pg,
			redisC:      redisC,
			rabbitC:     rabbitC,
			email:       smtpClient,
			accRepo:     accRepo,
			accRoleRepo: accRoleRepo,
			tokenRepo:   vTokenRepo,
			dbClient:    pgClient,
			cfg:         cfg,
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
// Returns the external account id and jwt tokens if they were requested.
func (a *testApp) createAccount(t *testing.T,
	email, password, ip, ua *string,
	needTokens bool,
) (string, string, string) {
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

	regResp, err := a.client.Register(context.Background(), &auth_v1.RegisterRequest{
		Email:    regEmail,
		Password: regPassword,
	})
	require.NoError(t, err)
	require.NotEmpty(t, regResp.GetAccountId())

	if !needTokens {
		return regResp.GetAccountId(), "", ""
	}

	var (
		loginIP = gofakeit.IPv4Address()
		loginUA = gofakeit.UserAgent()
	)

	if ip != nil {
		loginIP = *ip
	}
	if ua != nil {
		loginUA = *ua
	}

	loginResp, err := a.client.Login(context.Background(), &auth_v1.LoginRequest{
		Email:     regEmail,
		Password:  regPassword,
		Ip:        &loginIP,
		UserAgent: &loginUA,
	})
	require.NoError(t, err)
	require.NotEmpty(t, loginResp.GetAccessToken())
	require.NotEmpty(t, loginResp.GetRefreshToken())

	return regResp.GetAccountId(), loginResp.GetAccessToken(), loginResp.GetRefreshToken()
}

func (a *testApp) blockAccount(t *testing.T, accountID string) {
	uc := usecase.NewBlockAccountUC(a.accRepo, a.accRoleRepo)
	out, err := uc.Execute(context.Background(), dto.BlockAccountInput{AccountID: uuid.MustParse(accountID)})

	require.NoError(t, err)
	require.True(t, out.Blocked)
}

func (a *testApp) deleteAccount(t *testing.T, accountID string) {
	uc := usecase.NewDeleteAccountUC(a.accRepo, a.accRoleRepo)
	out, err := uc.Execute(context.Background(), dto.DeleteAccountInput{AccountID: uuid.MustParse(accountID)})

	require.NoError(t, err)
	require.True(t, out.Deleted)
}

func (a *testApp) logout(t *testing.T, refreshToken string) {
	resp, err := a.client.Logout(context.Background(), &auth_v1.LogoutRequest{
		RefreshToken: refreshToken,
	})
	require.NoError(t, err)
	require.True(t, resp.GetLogout())
}

// Helper for e2e tests.
// Creates a new verification token for specified account id.
// Set `shortLive` parameter if you need a short living token for tests
// such as testing of expiration.
//
// Returns the token.
func (a *testApp) sendToken(t *testing.T, accountID, email string, shortLive bool) string {
	if !shortLive {
		resp, err := a.client.SendVerification(context.Background(), &auth_v1.SendVerificationRequest{AccountId: accountID})
		require.NoError(t, err)
		require.True(t, resp.Sent)
	} else {
		vToken := model.RestoreVerificationToken(
			uuid.NewString(),
			uuid.MustParse(accountID),
			time.Minute,
			time.Now().Add(time.Microsecond),
		)

		err := a.tokenRepo.Save(context.Background(), vToken)
		require.NoError(t, err)

		err = a.email.SendVerificationEmail(context.Background(), email, vToken.Token())
		require.NoError(t, err)
	}

	token, ok := a.email.LastToken(email)
	require.NotEmpty(t, token)
	require.True(t, ok)

	return token
}
