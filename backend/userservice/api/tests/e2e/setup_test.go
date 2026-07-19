///go:build e2e

package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/google/uuid"
	"github.com/maket12/ads-service/userservice/internal/app/dto"
	"github.com/maket12/ads-service/userservice/pkg/utils"
	amqp "github.com/rabbitmq/amqp091-go"
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
	"github.com/maket12/ads-service/userservice/migrations"
	"github.com/maket12/ads-service/userservice/pkg/generated/user_v1"
	pkgpostgres "github.com/maket12/ads-service/userservice/pkg/postgres"
	pkgrabbitmq "github.com/maket12/ads-service/userservice/pkg/rabbitmq"
)

const bufSize = 1024 * 1024

type testApp struct {
	client       user_v1.UserServiceClient
	conn         *grpc.ClientConn
	pg           *pkgpostgres.TestContainer
	rabbitC      *pkgrabbitmq.TestContainer
	rabbitClient *pkgrabbitmq.Client
	profileRepo  port.ProfileRepository
	dbClient     *pkgpostgres.Client
	cfg          *config.TestConfig
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
		require.NoError(t, pg.MigrateUp(migrations.FS, 1))

		rabbitC, err := pkgrabbitmq.StartTestContainer(ctx)
		require.NoError(t, err)

		cfg.DbHost, cfg.DbPort = pg.Config.Host, pg.Config.Port
		cfg.DbUser, cfg.DbPassword, cfg.DbName = pg.Config.User, pg.Config.Password, pg.Config.Name

		cfg.RabbitHost, cfg.RabbitPort = rabbitC.Config.Host, rabbitC.Config.Port
		cfg.RabbitUser, cfg.RabbitPassword = rabbitC.Config.User, rabbitC.Config.Password

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
		fmt.Println("handler")

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

		fmt.Println("app instance")

		appInstance = &testApp{
			client:       user_v1.NewUserServiceClient(conn),
			conn:         conn,
			pg:           pg,
			rabbitC:      rabbitC,
			profileRepo:  profileRepo,
			rabbitClient: rabbitClient,
			dbClient:     pgClient,
			cfg:          cfg,
		}
	})

	if appInstance == nil {
		t.Fatal("setupE2E: initialization failed on a previous test, appInstance is nil")
	}

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
// If account id is not specified, then it uses random value instead.
//
// Returns the created profile's account id.
func (a *testApp) createProfile(t *testing.T, accountID *string) string {
	var id = uuid.NewString()
	if accountID != nil {
		id = *accountID
	}

	uc := usecase.NewCreateProfileUC(a.profileRepo)
	err := uc.Execute(context.Background(), dto.CreateProfileInput{AccountID: uuid.MustParse(id)})
	require.NoError(t, err)

	return id
}

// Helper for e2e tests.
// Returns the profile with specified account id.
//
// Make sure the profile with this account id was created before.
func (a *testApp) getProfile(t *testing.T, accountID string) *user_v1.GetProfileResponse {
	ctx := utils.PackAccountIDForGRPC(context.Background(), accountID)
	resp, err := a.client.GetProfile(ctx, &user_v1.GetProfileRequest{})

	require.NoError(t, err)
	require.Equal(t, accountID, resp.GetAccountId())

	return resp
}

// Helper for e2e tests.
// Publishes an account.created event to RabbitMQ the same way authservice's
// AccountPublisher does, so the AccountSubscriber -> CreateProfileUC path is
// exercised for real instead of being bypassed.
func (a *testApp) publishAccountCreated(t *testing.T, accountID string) {
	ch, err := a.rabbitClient.Conn.Channel()
	require.NoError(t, err)
	defer ch.Close()

	err = ch.ExchangeDeclare(
		a.cfg.ExchangeName,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	require.NoError(t, err)

	event := pkgrabbitmq.AccountCreatedEvent{AccountID: uuid.MustParse(accountID)}
	body, err := json.Marshal(event)
	require.NoError(t, err)

	err = ch.PublishWithContext(
		context.Background(),
		a.cfg.ExchangeName,
		a.cfg.RoutingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	require.NoError(t, err)
}

// Helper for e2e tests.
// Polls GetProfile until the profile appears (or the timeout elapses),
// since consumption of the published event is asynchronous.
func (a *testApp) waitForProfile(t *testing.T, accountID string, timeout time.Duration) *user_v1.GetProfileResponse {
	deadline := time.Now().Add(timeout)
	ctx := utils.PackAccountIDForGRPC(context.Background(), accountID)

	for time.Now().Before(deadline) {
		resp, err := a.client.GetProfile(ctx, &user_v1.GetProfileRequest{})
		if err == nil {
			return resp
		}
		time.Sleep(100 * time.Millisecond)
	}

	t.Fatalf("timed out waiting for profile %s to be created", accountID)
	return nil
}
