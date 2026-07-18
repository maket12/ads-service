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
	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"

	"github.com/maket12/ads-service/userservice/cmd/app/config"
	adaptergrpc "github.com/maket12/ads-service/userservice/internal/adapter/in/grpc"
	adapterrabbitmq "github.com/maket12/ads-service/userservice/internal/adapter/in/rabbitmq"
	adapterpostgres "github.com/maket12/ads-service/userservice/internal/adapter/out/postgres"
	"github.com/maket12/ads-service/userservice/internal/app/usecase"
	"github.com/maket12/ads-service/userservice/internal/domain/port"
	adapterphone "github.com/maket12/ads-service/userservice/internal/infrastructure/phone"
	// TODO: adjust to the actual migrations package path for userservice, mirroring
	// authservice's `internal/.../migrations` package (embedded FS with .sql files).
	"github.com/maket12/ads-service/userservice/migrations"
	"github.com/maket12/ads-service/userservice/pkg/generated/user_v1"
	pkgpostgres "github.com/maket12/ads-service/userservice/pkg/postgres"
	pkgrabbitmq "github.com/maket12/ads-service/userservice/pkg/rabbitmq"
)

const bufSize = 1024 * 1024

type testApp struct {
	client      user_v1.UserServiceClient
	conn        *grpc.ClientConn
	pg          *pkgpostgres.TestContainer
	rabbitC     *pkgrabbitmq.TestContainer
	profileRepo port.ProfileRepository
	dbClient    *pkgpostgres.Client
	cfg         *config.TestConfig // TODO: confirm userservice has a config.TestConfig / config.LoadTest, like authservice does
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
		require.NoError(t, pg.MigrateUp(migrations.FS, 3)) // TODO: confirm migration version count for userservice

		rabbitC, err := pkgrabbitmq.StartTestContainer(ctx)
		require.NoError(t, err)

		cfg.DbHost, cfg.DbPort = pg.Config.Host, pg.Config.Port
		cfg.DbUser, cfg.DbPassword, cfg.DbName = pg.Config.User, pg.Config.Password, pg.Config.Name

		cfg.RabbitHost, cfg.RabbitPort = rabbitC.Config.Host, rabbitC.Config.Port

		logger := newLogger()

		pgClient, err := pkgpostgres.NewClient(ctx, pkgpostgres.NewConfig(
			pg.Config.Host, pg.Config.Port, pg.Config.User,
			pg.Config.Password, pg.Config.Name, pg.Config.SSLMode,
			pg.Config.MaxConn, pg.Config.MinConn,
			pg.Config.MaxConnLifeTime, pg.Config.MaxConnIdleTime,
		))
		require.NoError(t, err)

		rabbitClient, err := pkgrabbitmq.NewClient(pkgrabbitmq.NewConfig(
			cfg.RabbitHost, cfg.RabbitPort,
			cfg.RabbitUser, cfg.RabbitPassword,
			cfg.RabbitWaitTime, cfg.RabbitAttempts,
		))
		require.NoError(t, err)

		// repositories
		profileRepo := adapterpostgres.NewProfileRepository(pgClient, trmpgx.DefaultCtxGetter)

		// infrastructure
		phoneValidator := adapterphone.NewValidator(cfg.PhoneDefaultRegion)

		// use-cases
		createProfileUC := usecase.NewCreateProfileUC(profileRepo)
		getProfileUC := usecase.NewGetProfileUC(profileRepo)
		updateProfileUC := usecase.NewUpdateProfileUC(profileRepo, phoneValidator)

		// rabbitmq subscriber, wired the same way as in main.go
		subConfig := adapterrabbitmq.NewSubscriberConfig(
			cfg.ExchangeName, cfg.QueueName, cfg.RoutingKey,
		)
		subscriber := adapterrabbitmq.NewAccountSubscriber(
			subConfig, logger, rabbitClient, createProfileUC,
		)
		go func() {
			_ = subscriber.Start(ctx)
		}()

		handler := adaptergrpc.NewUserHandler(
			logger, getProfileUC, updateProfileUC,
		)

		// --- in-memory gRPC server via bufconn ---
		lis := bufconn.Listen(bufSize)
		grpcServer := grpc.NewServer()
		user_v1.RegisterUserServiceServer(grpcServer, handler)

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
			client:      user_v1.NewUserServiceClient(conn),
			conn:        conn,
			pg:          pg,
			rabbitC:     rabbitC,
			profileRepo: profileRepo,
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
}

// Helper for e2e tests.
// Creates a new profile directly through the repository, since profile creation
// in userservice is triggered by a RabbitMQ event (AccountSubscriber -> CreateProfileUC)
// rather than by a gRPC call, unlike authservice's Register.
//
// If parameters are not specified, then it uses random values instead.
//
// Returns the created profile's account id.
func (a *testApp) createProfile(t *testing.T, accountID *string, firstName, lastName *string) string {
	var (
		id    = uuid.NewString()
		fName = gofakeit.FirstName()
		lName = gofakeit.LastName()
	)

	if accountID != nil {
		id = *accountID
	}
	if firstName != nil {
		fName = *firstName
	}
	if lastName != nil {
		lName = *lastName
	}

	uc := usecase.NewCreateProfileUC(a.profileRepo)
	// TODO: replace dto.CreateProfileInput with the actual input struct/fields for CreateProfileUC.
	_, err := uc.Execute(context.Background(), usecase.CreateProfileInput{
		AccountID: uuid.MustParse(id),
		FirstName: fName,
		LastName:  lName,
	})
	require.NoError(t, err)

	return id
}

// Helper for e2e tests.
// Fetches a profile via the gRPC client.
func (a *testApp) getProfile(t *testing.T, accountID string) *user_v1.GetProfileResponse {
	resp, err := a.client.GetProfile(context.Background(), &user_v1.GetProfileRequest{
		AccountId: accountID,
	})
	require.NoError(t, err)

	return resp
}

// Helper for e2e tests.
// Updates a profile via the gRPC client.
func (a *testApp) updateProfile(t *testing.T, accountID string, firstName, lastName, phone *string) *user_v1.UpdateProfileResponse {
	req := &user_v1.UpdateProfileRequest{
		AccountId: accountID,
		FirstName: firstName,
		LastName:  lastName,
		Phone:     phone,
	}

	resp, err := a.client.UpdateProfile(context.Background(), req)
	require.NoError(t, err)

	return resp
}
