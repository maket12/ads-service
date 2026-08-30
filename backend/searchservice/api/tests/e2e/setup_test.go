///go:build e2e

package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/google/uuid"
	"github.com/maket12/ads-service/backend/authservice/pkg/utils"
	"github.com/maket12/ads-service/backend/searchservice/api/proto/generated/search_v1"
	adapterelasticsearch "github.com/maket12/ads-service/backend/searchservice/internal/adapter/out/elasticsearch"
	"github.com/maket12/ads-service/backend/searchservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/backend/searchservice/internal/app/errs"
	pkgelasticsearch "github.com/maket12/ads-service/backend/searchservice/pkg/elasticsearch"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	pkgrabbitmq "github.com/maket12/ads-service/backend/authservice/pkg/rabbitmq"
	"github.com/maket12/ads-service/backend/searchservice/cmd/app/config"
	adaptergrpc "github.com/maket12/ads-service/backend/searchservice/internal/adapter/in/grpc"
	adapterrabbitmq "github.com/maket12/ads-service/backend/searchservice/internal/adapter/in/rabbitmq"
	"github.com/maket12/ads-service/backend/searchservice/internal/app/usecase"
	"github.com/maket12/ads-service/backend/searchservice/internal/domain/port"
)

const bufSize = 1024 * 1024

type testApp struct {
	client        search_v1.SearchServiceClient
	conn          *grpc.ClientConn
	es            *pkgelasticsearch.TestContainer
	rabbitC       *pkgrabbitmq.TestContainer
	rabbitClient  *pkgrabbitmq.Client
	adIdxRepo     port.AdIndexRepository
	deleteAdIdxUC *usecase.DeleteAdIndexUC
	cfg           *config.TestConfig
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
		es, err := pkgelasticsearch.StartTestContainer(ctx)
		require.NoError(t, err)

		rabbitC, err := pkgrabbitmq.StartTestContainer(ctx)
		require.NoError(t, err)

		cfg.ESAddresses = es.Config.Addresses
		cfg.ESUsername, cfg.ESPassword = es.Config.Username, es.Config.Password

		cfg.RabbitHost, cfg.RabbitPort = rabbitC.Config.Host, rabbitC.Config.Port
		cfg.RabbitUser, cfg.RabbitPassword = rabbitC.Config.User, rabbitC.Config.Password

		logger := newLogger()

		esClient, err := pkgelasticsearch.NewClient(ctx,
			pkgelasticsearch.NewConfig(
				es.Config.Addresses,
				es.Config.Username,
				es.Config.Password,
			),
		)
		require.NoError(t, err)

		rabbitClient, err := pkgrabbitmq.NewClient(pkgrabbitmq.NewConfig(
			cfg.RabbitHost, cfg.RabbitPort,
			cfg.RabbitUser, cfg.RabbitPassword,
			cfg.RabbitWaitTime, cfg.RabbitAttempts,
		))
		require.NoError(t, err)

		// repositories
		adIdxRepo := adapterelasticsearch.NewAdIndexRepository(
			esClient,
			cfg.ESIndexName,
		)

		// use-cases
		createAdIdxUC := usecase.NewCreateAdIndexUC(adIdxRepo)
		deleteAdIdxUC := usecase.NewDeleteAdIndexUC(adIdxRepo)
		searchAdsUC := usecase.NewSearchAdsUC(adIdxRepo)

		// rabbitmq subscriber, wired the same way as in main.go
		subConfig := adapterrabbitmq.NewSubscriberConfig(
			cfg.AdExchange, cfg.AdQueue,
			cfg.AdPublishedRoutingKey,
			cfg.AdUpdatedRoutingKey,
			cfg.AdRejectedRoutingKey,
			cfg.AdDeletedRoutingKey,
		)
		subscriber := adapterrabbitmq.NewAdSubscriber(
			subConfig, logger, rabbitClient,
			createAdIdxUC, deleteAdIdxUC,
		)
		go func() {
			_ = subscriber.Start(ctx)
		}()

		time.Sleep(1 * time.Second)

		handler := adaptergrpc.NewSearchHandler(logger, searchAdsUC)
		fmt.Println("handler")

		// --- in-memory gRPC server via bufconn ---
		lis := bufconn.Listen(bufSize)
		grpcServer := grpc.NewServer()
		search_v1.RegisterSearchServiceServer(grpcServer, handler)

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
			client:        search_v1.NewSearchServiceClient(conn),
			conn:          conn,
			es:            es,
			rabbitC:       rabbitC,
			rabbitClient:  rabbitClient,
			adIdxRepo:     adIdxRepo,
			deleteAdIdxUC: deleteAdIdxUC,
			cfg:           cfg,
		}
	})

	if appInstance == nil {
		t.Fatal("setupE2E: initialization failed on a previous test, appInstance is nil")
	}

	appInstance.cleanData(t, context.Background())

	return appInstance
}

func (a *testApp) cleanData(t *testing.T, ctx context.Context) {
	err := a.es.ClearIndices(ctx, a.cfg.ESIndexName)
	require.NoError(t, err, "failed to clear es indices")
}

// Helper for e2e tests.
// Creates a new ad index directly through the repository.
// Any parameter passed as nil will be generated randomly.
func (a *testApp) createAdIndex(t *testing.T,
	adID, title, description *string,
	price *int64,
	category, mainImage *string,
) string {
	id := uuid.NewString()
	if adID != nil {
		id = *adID
	}

	tTitle := gofakeit.ProductName()
	if title != nil {
		tTitle = *title
	}

	tDesc := gofakeit.ProductDescription()
	if description != nil {
		tDesc = *description
	}

	tPrice := int64(gofakeit.Price(100, 10000))
	if price != nil {
		tPrice = *price
	}

	tCategory := gofakeit.RandomString([]string{"food", "video_games", "electronics"})
	if category != nil {
		tCategory = *category
	}

	tImage := gofakeit.URL()
	if mainImage != nil {
		tImage = *mainImage
	}

	uc := usecase.NewCreateAdIndexUC(a.adIdxRepo)
	err := uc.Execute(context.Background(), dto.CreateAdIndexInput{
		ID:          id,
		Title:       tTitle,
		Description: tDesc,
		Price:       tPrice,
		Category:    tCategory,
		MainImage:   tImage,
	})
	require.NoError(t, err)

	return id
}

// Helper for e2e tests.
// Publishes an ad.published event to RabbitMQ the same way adservice's
// AdPublisher does, so the AdSubscriber -> CreateAdIndexUC path is
// exercised for real instead of being bypassed.
func (a *testApp) publishAdPublished(t *testing.T, adID string) {
	ch, err := a.rabbitClient.Conn.Channel()
	require.NoError(t, err)
	defer ch.Close()

	err = ch.ExchangeDeclare(
		a.cfg.AdExchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	require.NoError(t, err)

	event := adapterrabbitmq.AdPublishedEvent{
		ID:          adID,
		Title:       gofakeit.BookTitle(),
		Description: gofakeit.ProductDescription(),
		Price:       int64(gofakeit.Price(100, 10000)),
		Category:    gofakeit.RandomString([]string{"books", "home", "food"}),
		MainImage:   gofakeit.URL(),
	}
	body, err := json.Marshal(event)
	require.NoError(t, err)

	err = ch.PublishWithContext(
		context.Background(),
		a.cfg.AdExchange,
		a.cfg.AdPublishedRoutingKey,
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
// Polls DeleteAdIndex until the ad index disappears (or the timeout elapses),
// since consumption of the published event is asynchronous.
func (a *testApp) waitForAdIdxCreated(t *testing.T, adID string, timeout time.Duration) {
	ctx := context.Background()
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		err := a.deleteAdIdxUC.Execute(ctx, dto.DeleteAdIndexInput{ID: adID})
		if err == nil {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}

	t.Fatalf("timed out waiting for ad index %s to be created", adID)
	return
}

// Helper for e2e tests.
// Publishes an ad.published event to RabbitMQ the same way adservice's
// AdPublisher does, so the AdSubscriber -> CreateAdIndexUC path is
// exercised for real instead of being bypassed.
func (a *testApp) publishAdDeleted(t *testing.T, adID string) {
	ch, err := a.rabbitClient.Conn.Channel()
	require.NoError(t, err)
	defer ch.Close()

	err = ch.ExchangeDeclare(
		a.cfg.AdExchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	require.NoError(t, err)

	event := adapterrabbitmq.AdDeletedEvent{ID: adID}
	body, err := json.Marshal(event)
	require.NoError(t, err)

	err = ch.PublishWithContext(
		context.Background(),
		a.cfg.AdExchange,
		a.cfg.AdDeletedRoutingKey,
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
// Waits until rabbitmq deliver the delete ad index event.
func (a *testApp) waitForAdIdxDeleted(t *testing.T, adID string, timeout time.Duration) {
	// wait some time to let rabbitmq deliver the message
	time.Sleep(timeout)

	err := a.deleteAdIdxUC.Execute(context.Background(), dto.DeleteAdIndexInput{ID: adID})
	if err != nil && errors.Is(err, ucerrs.ErrAdIndexNotFound) {
		return
	}

	t.Fatalf("timed out waiting for ad index %s to be deleted", adID)
}

// Helper for e2e tests.
//
// Returns `context.Context` contains user auth data.
func (a *testApp) userCtx() context.Context {
	ctx := utils.PackAccountIDForGRPC(context.Background(), uuid.NewString())
	return utils.PackAccountRoleForGRPC(ctx, "user")
}
