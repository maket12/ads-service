package postgres_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	adapterpostgres "github.com/maket12/ads-service/authservice/internal/adapter/out/postgres"
	"github.com/maket12/ads-service/authservice/internal/domain/model"
	"github.com/maket12/ads-service/authservice/migrations"
	pkgerrs "github.com/maket12/ads-service/pkg/errs"
	pkgpostgres "github.com/maket12/ads-service/pkg/postgres"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/stretchr/testify/suite"
)

type AccountsRepoSuite struct {
	suite.Suite
	dbClient    *pkgpostgres.Client
	repo        *adapterpostgres.AccountRepository
	ctx         context.Context
	migrate     *migrate.Migrate
	testAccount *model.Account
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
	s.repo = adapterpostgres.NewAccountsRepository(s.dbClient)
	s.ctx = context.Background()
	s.testAccount, _ = model.NewAccount("new@email.com", "hashed-secret-pass")
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
	_, err := s.dbClient.DB.Exec("TRUNCATE TABLE accounts CASCADE")
	s.Require().NoError(err)
}

func (s *AccountsRepoSuite) TestCreateGetByID() {
	// Check create first
	err := s.repo.Create(s.ctx, s.testAccount)
	s.Require().NoError(err)

	// And then get
	acc, err := s.repo.GetByID(s.ctx, s.testAccount.ID())
	s.Require().NoError(err)
	s.Require().Exactly(s.testAccount.Email(), acc.Email())
	s.Require().Exactly(s.testAccount.PasswordHash(), acc.PasswordHash())
}

func (s *AccountsRepoSuite) TestCreate_DuplicateEmail() {
	// Create an account at first
	_ = s.repo.Create(s.ctx, s.testAccount)

	// Trying to create an account with the same email
	newAcc, _ := model.NewAccount(s.testAccount.Email(), "hashed-pass")
	err := s.repo.Create(s.ctx, newAcc)
	s.Require().Error(err)
}

func (s *AccountsRepoSuite) TestGetByEmail() {
	// Create an account in advance
	_ = s.repo.Create(s.ctx, s.testAccount)

	// Get by email
	acc, err := s.repo.GetByEmail(s.ctx, s.testAccount.Email())
	s.Require().NoError(err)
	s.Require().Exactly(s.testAccount.ID(), acc.ID())
	s.Require().Exactly(s.testAccount.PasswordHash(), acc.PasswordHash())
}

func (s *AccountsRepoSuite) TestGetByEmail_CaseInsensitive() {
	// Create an account in advance
	_ = s.repo.Create(s.ctx, s.testAccount)

	// Trying to get by the same email, but in upper case
	var upperEmail = strings.ToUpper(s.testAccount.Email())
	acc, err := s.repo.GetByEmail(s.ctx, upperEmail)

	s.Require().NoError(err)
	s.Require().Equal(s.testAccount.ID(), acc.ID())
}

func (s *AccountsRepoSuite) TestGetByEmail_NotFound() {
	// Trying to get non-existing account
	var unexistingEmail = "unexist@gmail.com"
	_, err := s.repo.GetByEmail(s.ctx, unexistingEmail)

	s.Require().Error(err)
	s.Require().ErrorIs(err, pkgerrs.ErrObjectNotFound)
}

func (s *AccountsRepoSuite) TestMarkLogin() {
	// Create account at first
	_ = s.repo.Create(s.ctx, s.testAccount)

	// Mark as logged in
	err := s.repo.MarkLogin(s.ctx, s.testAccount)
	s.Require().NoError(err)

	// Check if the account is marked
	acc, _ := s.repo.GetByEmail(s.ctx, s.testAccount.Email())
	s.Require().NotNil(acc.LastLoginAt())

	// Check update time
	s.Require().NotEqual(s.testAccount.UpdatedAt(), acc.UpdatedAt(),
		"expected new update time")
}

func (s *AccountsRepoSuite) TestVerifyEmail() {
	// Create account at first
	_ = s.repo.Create(s.ctx, s.testAccount)

	// Verify its email
	err := s.repo.VerifyEmail(s.ctx, s.testAccount)
	s.Require().NoError(err)

	// Check if the account is marked
	acc, _ := s.repo.GetByID(s.ctx, s.testAccount.ID())
	s.Require().True(acc.EmailVerified())

	// Check update time
	s.Require().NotEqual(s.testAccount.UpdatedAt(), acc.UpdatedAt(),
		"expected new update time")
}
