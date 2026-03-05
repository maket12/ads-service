package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	adapterpostgres "github.com/maket12/ads-service/adservice/internal/adapter/out/postgres"
	"github.com/maket12/ads-service/adservice/internal/domain/model"
	"github.com/maket12/ads-service/adservice/migrations"
	pkgerrs "github.com/maket12/ads-service/pkg/errs"
	pkgpostgres "github.com/maket12/ads-service/pkg/postgres"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type AdRepoSuite struct {
	suite.Suite
	dbClient *pkgpostgres.Client
	repo     *adapterpostgres.AdRepository
	ctx      context.Context
	migrate  *migrate.Migrate
	testAd   *model.Ad
}

func TestAdRepoSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	suite.Run(t, new(AdRepoSuite))
}

func (s *AdRepoSuite) setupDatabase() {
	const targetVersion = 1

	dbConfig := pkgpostgres.NewConfig(
		"localhost", 5432,
		"test", "test", "testdb",
		"disable", 25, 25, time.Minute*5,
	)
	dsn := "postgres://test:test@localhost:5432/testdb?sslmode=disable"

	dbClient, err := pkgpostgres.NewClient(dbConfig)
	s.Require().NoError(err)
	s.dbClient = dbClient

	sourceDriver, err := iofs.New(migrations.FS, ".")
	s.Require().NoError(err, "failed to create iofs driver")

	m, err := migrate.NewWithSourceInstance(
		"iofs",
		sourceDriver,
		dsn,
	)
	s.Require().NoError(err, "failed to create migration instance")

	s.migrate = m

	err = m.Migrate(targetVersion)

	// If migration is correct - setup has done
	if err == nil || errors.Is(err, migrate.ErrNoChange) {
		return
	}

	// Except dirty db as a normal scenario
	var dirtyErr migrate.ErrDirty
	if !errors.As(err, &dirtyErr) {
		s.FailNowf("failed to migrate up", "unexpected error: %v", err)
	}

	// ================ Restore dirty database ================
	_ = m.Force(dirtyErr.Version)

	err = m.Down()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		s.Require().NoError(err, "failed to migrate down during recovery")
	}

	err = m.Migrate(targetVersion)
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		s.Require().NoError(err, "failed to migrate up after recovery")
	}
}

func (s *AdRepoSuite) SetupSuite() {
	s.setupDatabase()
	s.repo = adapterpostgres.NewAdRepository(s.dbClient)
	s.ctx = context.Background()
	s.testAd, _ = model.NewAd(
		uuid.New(),
		"Lamborghini X5",
		nil,
		int64(1000000),
		[]string{"overview.png", "salon.png", "circles.jpeg"},
	)
}

func (s *AdRepoSuite) TearDownSuite() {
	if s.migrate != nil {
		if err := s.migrate.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			s.Require().NoError(err, "failed to migrate down")
		}
	}
	err := s.dbClient.Close()
	s.Require().NoError(err, "failed to close db connection")
}

func (s *AdRepoSuite) SetupTest() {
	_, err := s.dbClient.DB.Exec("TRUNCATE TABLE ads CASCADE")
	s.Require().NoError(err)
}

func (s *AdRepoSuite) TestCreateGet() {
	// Create at first
	err := s.repo.Create(s.ctx, s.testAd)
	s.Require().NoError(err)

	// And then get
	ad, err := s.repo.Get(s.ctx, s.testAd.ID())
	s.Require().NoError(err)
	s.Require().NotNil(ad)
	s.Require().Exactly(s.testAd.ID(), ad.ID())
	s.Require().Exactly(s.testAd.SellerID(), ad.SellerID())
	s.Require().Exactly(s.testAd.Title(), ad.Title())
}

func (s *AdRepoSuite) TestGet_NotFound() {
	// Trying to get non-existing ad
	var unexistingAdID = uuid.New()
	ad, err := s.repo.Get(s.ctx, unexistingAdID)

	s.Require().Error(err)
	s.Require().ErrorIs(err, pkgerrs.ErrObjectNotFound)
	s.Require().Nil(ad)
}

func (s *AdRepoSuite) TestUpdate() {
	// Create an ad in advance
	_ = s.repo.Create(s.ctx, s.testAd)

	var (
		updatedAd = *s.testAd
		testTitle = "Ferrari F40"
		testPrice = int64(2000000)
	)

	_ = updatedAd.Update(
		&testTitle, nil, &testPrice, nil,
	)

	// Update
	err := s.repo.Update(s.ctx, &updatedAd)
	s.Require().NoError(err)

	// Ensure update was successful
	ad, _ := s.repo.Get(s.ctx, s.testAd.ID())
	s.Require().NotNil(ad)
	s.Require().Exactly(testTitle, ad.Title())
	s.Require().Exactly(testPrice, ad.Price())
}

func (s *AdRepoSuite) TestUpdateStatus() {
	// Create an ad in advance
	_ = s.repo.Create(s.ctx, s.testAd)

	// Change status (for example, publish)
	_ = s.testAd.Publish()
	err := s.repo.UpdateStatus(s.ctx, s.testAd)
	s.Require().NoError(err)

	// Ensure update was successful
	ad, _ := s.repo.Get(s.ctx, s.testAd.ID())
	s.Require().NotNil(ad)
	s.Require().Exactly(model.AdPublished, ad.Status())
}

func (s *AdRepoSuite) TestDelete() {
	// Create an ad in advance
	_ = s.repo.Create(s.ctx, s.testAd)

	// Then delete
	err := s.repo.Delete(s.ctx, s.testAd.ID())
	s.Require().NoError(err)

	// Ensure delete was successful
	_, err = s.repo.Get(s.ctx, s.testAd.ID())

	s.Require().Error(err)
	s.Require().ErrorIs(err, pkgerrs.ErrObjectNotFound)
}

func (s *AdRepoSuite) TestDeleteAllAds() {
	// Create ads in advance with same seller id
	anotherAd, _ := model.NewAd(
		s.testAd.SellerID(),
		"New car",
		nil,
		int64(300000),
		nil,
	)

	_ = s.repo.Create(s.ctx, s.testAd)
	_ = s.repo.Create(s.ctx, anotherAd)

	// Then delete
	err := s.repo.DeleteAll(s.ctx, s.testAd.SellerID())
	s.Require().NoError(err)

	// Ensure delete was successful
	_, err = s.repo.Get(s.ctx, s.testAd.ID())
	s.Require().Error(err)
	s.Require().ErrorIs(err, pkgerrs.ErrObjectNotFound)

	_, err = s.repo.Get(s.ctx, s.testAd.ID())
	s.Require().Error(err)
	s.Require().ErrorIs(err, pkgerrs.ErrObjectNotFound)
}

func (s *AdRepoSuite) TestListAds() {
	// Create ads in advance
	anotherAd, _ := model.NewAd(
		uuid.New(),
		"New car",
		nil,
		int64(300000),
		nil,
	)

	_ = s.repo.Create(s.ctx, s.testAd)
	_ = s.repo.Create(s.ctx, anotherAd)

	// ################ Test limit ################
	var (
		testLimit  = 1
		testOffset = 0
	)

	ads, err := s.repo.ListAds(s.ctx, testLimit, testOffset)
	s.Require().NoError(err)
	s.Require().NotNil(ads)
	s.Require().Len(ads, 1)

	var fstFound bool
	for i := range ads {
		value := *ads[i]
		if value.ID() == s.testAd.ID() {
			fstFound = true
		}
	}

	s.Require().Truef(fstFound, "expected account with id %v\n in %v",
		s.testAd.ID(), ads)

	// ################ Test offset ################
	testLimit = 10
	testOffset = 1

	ads, err = s.repo.ListAds(s.ctx, testLimit, testOffset)
	s.Require().NoError(err)
	s.Require().NotNil(ads)
	s.Require().Len(ads, 1)
}

func (s *AdRepoSuite) TestListSellerAds() {
	var testSellerID = uuid.New()
	// Create ads in advance
	anotherAd1, _ := model.NewAd(
		testSellerID,
		"New car",
		nil,
		int64(300000),
		nil,
	)
	anotherAd2, _ := model.NewAd(
		testSellerID,
		"New car",
		nil,
		int64(300000),
		nil,
	)

	_ = s.repo.Create(s.ctx, s.testAd)
	_ = s.repo.Create(s.ctx, anotherAd1)
	_ = s.repo.Create(s.ctx, anotherAd2)

	var (
		testLimit  = 10
		testOffset = 0
	)

	ads, err := s.repo.ListSellerAds(
		s.ctx, testSellerID, testLimit, testOffset,
	)
	s.Require().NoError(err)
	s.Require().NotNil(ads)
	s.Require().Len(ads, 2)

	var fstFound, sndFound bool
	for i := range ads {
		value := *ads[i]
		s.Require().Equal(testSellerID, value.SellerID())
		if value.ID() == anotherAd1.ID() {
			fstFound = true
		} else if value.ID() == anotherAd2.ID() {
			sndFound = true
		}
	}

	s.Require().Truef(fstFound, "expected account with id %v\n in %v",
		anotherAd1.ID(), ads,
	)
	s.Require().Truef(sndFound, "expected account with id %v\n in %v",
		anotherAd2.ID(), ads,
	)
}
