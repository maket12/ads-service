//go:build e2e

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
	"github.com/google/uuid"
	"github.com/maket12/ads-service/adservice/pkg/utils"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"

	"github.com/maket12/ads-service/adservice/cmd/app/config"
	adaptergrpc "github.com/maket12/ads-service/adservice/internal/adapter/in/grpc"
	adaptermongo "github.com/maket12/ads-service/adservice/internal/adapter/out/mongodb"
	adapterpg "github.com/maket12/ads-service/adservice/internal/adapter/out/postgres"
	"github.com/maket12/ads-service/adservice/internal/app/usecase"
	"github.com/maket12/ads-service/adservice/internal/domain/port"
	"github.com/maket12/ads-service/adservice/migrations"
	"github.com/maket12/ads-service/adservice/pkg/generated/ad_v1"
	pkgmongodb "github.com/maket12/ads-service/adservice/pkg/mongodb"
	pkgpostgres "github.com/maket12/ads-service/adservice/pkg/postgres"
)

const bufSize = 1024 * 1024

type testApp struct {
	client    ad_v1.AdServiceClient
	conn      *grpc.ClientConn
	pg        *pkgpostgres.TestContainer
	mongoC    *pkgmongodb.TestContainer
	adRepo    port.AdRepository
	mediaRepo port.MediaRepository
	dbClient  *pkgpostgres.Client
	mongoCl   *pkgmongodb.Client
	cfg       *config.TestConfig
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

		mongoC, err := pkgmongodb.StartTestContainer(ctx)
		require.NoError(t, err)

		cfg.DbHost, cfg.DbPort = pg.Config.Host, pg.Config.Port
		cfg.DbUser, cfg.DbPassword, cfg.DbName = pg.Config.User, pg.Config.Password, pg.Config.Name

		cfg.MongoHost, cfg.MongoPort = mongoC.Config.Host, mongoC.Config.Port
		cfg.MongoUser, cfg.MongoPassword, cfg.MongoDBName = mongoC.Config.User, mongoC.Config.Password, mongoC.Config.DBName

		logger := newLogger()

		pgClient, err := pkgpostgres.NewClient(ctx, pkgpostgres.NewConfig(
			pg.Config.Host, pg.Config.Port, pg.Config.User,
			pg.Config.Password, pg.Config.Name, pg.Config.SSLMode,
			pg.Config.MaxConn, pg.Config.MinConn,
			pg.Config.MaxConnLifeTime, pg.Config.MaxConnIdleTime,
		))
		require.NoError(t, err)

		mongoClient, err := pkgmongodb.NewClient(ctx, pkgmongodb.NewConfig(
			mongoC.Config.Host, mongoC.Config.Port,
			mongoC.Config.User, mongoC.Config.Password,
			mongoC.Config.DBName,
		))
		require.NoError(t, err)

		// media repository config
		mediaRepoCfg := adaptermongo.NewMediaRepositoryConfig(
			mongoClient,
			cfg.MongoCollectionName,
		)

		// transaction manager
		trManager := manager.Must(trmpgx.NewDefaultFactory(pgClient.Pool))

		// repositories
		adRepo := adapterpg.NewAdRepository(pgClient, trmpgx.DefaultCtxGetter)
		mediaRepo := adaptermongo.NewMediaRepository(mediaRepoCfg)

		// use-cases
		createAdUC := usecase.NewCreateAdUC(trManager, adRepo, mediaRepo)
		getAdUC := usecase.NewGetAdUC(adRepo, mediaRepo)
		updateAdUC := usecase.NewUpdateAdUC(trManager, adRepo, mediaRepo)
		publishAdUC := usecase.NewPublishAdUC(adRepo)
		rejectAdUC := usecase.NewRejectAdUC(adRepo)
		deleteAdUC := usecase.NewDeleteAdUC(trManager, adRepo, mediaRepo)
		deleteAllAdsUC := usecase.NewDeleteAllAdsUC(trManager, adRepo, mediaRepo)
		listSellerAdsUC := usecase.NewListSellerAdsUC(adRepo)

		handler := adaptergrpc.NewAdHandler(
			logger,
			createAdUC,
			getAdUC,
			updateAdUC,
			publishAdUC,
			rejectAdUC,
			deleteAdUC,
			deleteAllAdsUC,
			listSellerAdsUC,
		)

		// --- in-memory gRPC server via bufconn ---
		lis := bufconn.Listen(bufSize)
		grpcServer := grpc.NewServer()
		ad_v1.RegisterAdServiceServer(grpcServer, handler)

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
			client:    ad_v1.NewAdServiceClient(conn),
			conn:      conn,
			pg:        pg,
			mongoC:    mongoC,
			adRepo:    adRepo,
			mediaRepo: mediaRepo,
			dbClient:  pgClient,
			mongoCl:   mongoClient,
			cfg:       cfg,
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
	require.NoError(t, err, "failed to truncate pg tables")

	err = a.mongoC.ClearCollections(ctx)
	require.NoError(t, err, "failed to clear mongo collections")
}

// Helper for e2e tests.
// Creates a new ad via request to “CreateAd“ grpc method.
//
// If seller id or payload are not specified, then it uses random values instead.
//
// Returns ad id and its seller id.
func (a *testApp) createAd(t *testing.T, sellerID *string, payload *ad_v1.CreateAdRequest) (string, string) {
	var accID = uuid.NewString()
	if sellerID != nil {
		accID = *sellerID
	}

	var req *ad_v1.CreateAdRequest
	if payload != nil {
		req = payload
	} else {
		req = &ad_v1.CreateAdRequest{
			Title:       gofakeit.ProductName(),
			Description: utils.VPtr(gofakeit.ProductDescription()),
			Price:       int64(gofakeit.Price(10000, 50000)),
			Images:      nil,
		}
	}

	ctx := utils.PackAccountIDForGRPC(context.Background(), accID)

	resp, err := a.client.CreateAd(ctx, req)
	require.NoError(t, err)
	require.NotEmpty(t, resp.GetAdId())

	return resp.GetAdId(), accID
}

// Helper for e2e tests.
// Publish ad by calling “PublishAd“ grpc method.
//
// Make sure you created the ad before and the owner has specified “seller id“.
func (a *testApp) publishAd(t *testing.T, adID, sellerID string) {
	ctx := utils.PackAccountIDForGRPC(context.Background(), sellerID)
	resp, err := a.client.PublishAd(ctx, &ad_v1.PublishAdRequest{AdId: adID})
	require.NoError(t, err)
	require.True(t, resp.GetSuccess())
}

// Helper for e2e tests.
// Reject ad by calling “RejectAd“ grpc method.
//
// Make sure the ad is available and the owner has specified “seller id“.
func (a *testApp) rejectAd(t *testing.T, adID, sellerID string) {
	ctx := utils.PackAccountIDForGRPC(context.Background(), sellerID)
	resp, err := a.client.RejectAd(ctx, &ad_v1.RejectAdRequest{AdId: adID})
	require.NoError(t, err)
	require.True(t, resp.GetSuccess())
}

// Helper for e2e tests.
// Reject ad by calling “DeleteAd“ grpc method.
//
// Make sure the ad isn't deleted and the owner has specified “seller id“.
func (a *testApp) deleteAd(t *testing.T, adID, sellerID string) {
	ctx := utils.PackAccountIDForGRPC(context.Background(), sellerID)
	resp, err := a.client.DeleteAd(ctx, &ad_v1.DeleteAdRequest{AdId: adID})
	require.NoError(t, err)
	require.True(t, resp.GetSuccess())
}
