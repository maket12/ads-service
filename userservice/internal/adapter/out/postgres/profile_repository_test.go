package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	pkgerrs "github.com/maket12/ads-service/pkg/errs"
	pkgpostgres "github.com/maket12/ads-service/pkg/postgres"
	adapterpostgres "github.com/maket12/ads-service/userservice/internal/adapter/out/postgres"
	"github.com/maket12/ads-service/userservice/internal/domain/model"
	"github.com/maket12/ads-service/userservice/migrations"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type AccountsRepoSuite struct {
	suite.Suite
	dbClient    *pkgpostgres.Client
	repo        *adapterpostgres.ProfileRepository
	ctx         context.Context
	migrate     *migrate.Migrate
	testProfile *model.Profile
}

func TestAccountsRepoSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	suite.Run(t, new(AccountsRepoSuite))
}

func (s *AccountsRepoSuite) setupDatabase() {
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

func (s *AccountsRepoSuite) SetupSuite() {
	s.setupDatabase()
	s.repo = adapterpostgres.NewProfileRepository(s.dbClient)
	s.ctx = context.Background()
	s.testProfile, _ = model.NewProfile(uuid.New())
}

func (s *AccountsRepoSuite) TearDownSuite() {
	if s.migrate != nil {
		if err := s.migrate.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			s.Require().NoError(err, "failed to migrate down")
		}
	}
	err := s.dbClient.Close()
	s.Require().NoError(err, "failed to close db connection")
}

func (s *AccountsRepoSuite) SetupTest() {
	_, err := s.dbClient.DB.Exec("TRUNCATE TABLE profiles CASCADE")
	s.Require().NoError(err)
}

func (s *AccountsRepoSuite) TestCreateGet() {
	// Create at first
	err := s.repo.Create(s.ctx, s.testProfile)
	s.Require().NoError(err)

	// And then get
	profile, err := s.repo.Get(s.ctx, s.testProfile.AccountID())
	s.Require().NoError(err)
	s.Require().Exactly(s.testProfile.AccountID(), profile.AccountID())
	s.Require().WithinDuration(s.testProfile.UpdatedAt(), profile.UpdatedAt(), time.Microsecond)
}

func (s *AccountsRepoSuite) TestCreate_Duplicate() {
	// Create a profile
	_ = s.repo.Create(s.ctx, s.testProfile)

	// Trying to create the same profile again (same account id)
	err := s.repo.Create(s.ctx, s.testProfile)
	s.Require().Error(err)
	s.Require().ErrorIs(err, pkgerrs.ErrObjectAlreadyExists)
}

func (s *AccountsRepoSuite) TestGet_NotFound() {
	// Trying to get non-existing profile
	var unexistingAccountID = uuid.New()
	_, err := s.repo.Get(s.ctx, unexistingAccountID)

	s.Require().Error(err)
	s.Require().ErrorIs(err, pkgerrs.ErrObjectNotFound)
}

func (s *AccountsRepoSuite) TestUpdate() {
	// Create a profile in advance
	_ = s.repo.Create(s.ctx, s.testProfile)

	var (
		updatedProfile = *s.testProfile
		firstName      = "Vladimir"
		bio            = "digital nomad, programmer"
	)

	_ = updatedProfile.Update(
		&firstName,
		nil,
		nil,
		nil,
		&bio,
	)

	// Update
	err := s.repo.Update(s.ctx, &updatedProfile)
	s.Require().NoError(err)

	// Ensure update was successful
	profile, _ := s.repo.Get(s.ctx, updatedProfile.AccountID())
	s.Require().NotNil(profile.FirstName())
	s.Require().Equal(firstName, *profile.FirstName())
	s.Require().NotNil(profile.Bio())
	s.Require().Equal(bio, *profile.Bio())
}

func (s *AccountsRepoSuite) TestDelete() {
	// Create a profile in advance
	_ = s.repo.Create(s.ctx, s.testProfile)

	// Then delete
	err := s.repo.Delete(s.ctx, s.testProfile.AccountID())
	s.Require().NoError(err)

	// Ensure it was deleted correctly
	_, err = s.repo.Get(s.ctx, s.testProfile.AccountID())

	s.Require().Error(err)
	s.Require().ErrorIs(err, pkgerrs.ErrObjectNotFound)
}

func (s *AccountsRepoSuite) TestListProfiles() {
	// Create profiles
	ids := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	for _, id := range ids {
		p, _ := model.NewProfile(id)
		_ = s.repo.Create(s.ctx, p)
	}

	// Testing limit
	var (
		testLimit  = 2
		testOffset = 0
	)

	profiles, err := s.repo.ListProfiles(s.ctx, testLimit, testOffset)
	s.Require().NoError(err)
	s.Require().Len(profiles, 2, "limit should restrict result count")

	var fstFound, sndFound bool
	for i := range profiles {
		value := *profiles[i]
		if value.AccountID() == ids[0] {
			fstFound = true
		}
		if value.AccountID() == ids[1] {
			sndFound = true
		}
	}

	s.Require().Truef(fstFound, "expected account with id %v\n in %v",
		ids[0], profiles)
	s.Require().Truef(sndFound, "expected account with id %v\n in %v",
		ids[1], profiles)

	// Testing offset
	testLimit = 10
	testOffset = 2

	profiles, err = s.repo.ListProfiles(s.ctx, testLimit, testOffset)
	s.Require().NoError(err)
	s.Require().Len(profiles, 1, "offset should skip records")
	s.Require().Equal(ids[2], profiles[0].AccountID(),
		"expected account with id %v\n in %v", ids[2], profiles)
}
