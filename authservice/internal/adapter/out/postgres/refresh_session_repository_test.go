package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	adapterpostgres "github.com/maket12/ads-service/authservice/internal/adapter/out/postgres"
	"github.com/maket12/ads-service/authservice/internal/domain/model"
	"github.com/maket12/ads-service/authservice/migrations"
	pkgerrs "github.com/maket12/ads-service/pkg/errs"
	pkgpostgres "github.com/maket12/ads-service/pkg/postgres"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type RefreshSessionsRepoSuite struct {
	suite.Suite
	dbClient    *pkgpostgres.Client
	repo        *adapterpostgres.RefreshSessionRepository
	ctx         context.Context
	migrate     *migrate.Migrate
	testSession *model.RefreshSession
}

func TestRefreshSessionRepoSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	suite.Run(t, new(RefreshSessionsRepoSuite))
}

func (s *RefreshSessionsRepoSuite) setupDatabase() {
	const targetVersion = 3

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

func (s *RefreshSessionsRepoSuite) SetupSuite() {
	s.setupDatabase()
	s.repo = adapterpostgres.NewRefreshSessionsRepository(s.dbClient)
	s.ctx = context.Background()

	var testAccount, _ = model.NewAccount("new@email.com", "hashed-secret-pass")
	s.testSession, _ = model.NewRefreshSession(
		uuid.New(),
		testAccount.ID(),
		"hashed-secret-token",
		nil,
		nil,
		nil,
		time.Minute,
	)

	// Create an account in the main table
	accountsRepo := adapterpostgres.NewAccountsRepository(s.dbClient)
	_ = accountsRepo.Create(s.ctx, testAccount)
}

func (s *RefreshSessionsRepoSuite) TearDownSuite() {
	if s.migrate != nil {
		if err := s.migrate.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			s.Require().NoError(err, "failed to migrate down")
		}
	}
	err := s.dbClient.Close()
	s.Require().NoError(err, "failed to close db connection")
}

func (s *RefreshSessionsRepoSuite) SetupTest() {
	_, err := s.dbClient.DB.Exec("TRUNCATE TABLE refresh_sessions CASCADE")
	s.Require().NoError(err)
}

func (s *RefreshSessionsRepoSuite) TestCreateGetByID() {
	// Create at first
	err := s.repo.Create(s.ctx, s.testSession)
	s.Require().NoError(err)

	// Get by id
	session, err := s.repo.GetByID(s.ctx, s.testSession.ID())
	s.Require().NoError(err)
	s.Require().Equal(s.testSession.AccountID(), session.AccountID())
	s.Require().Equal(s.testSession.RefreshTokenHash(), session.RefreshTokenHash())
}

func (s *RefreshSessionsRepoSuite) TestCreate_NonExistingAccount() {
	// Trying to create a session for non-existing account
	var anotherSession, _ = model.NewRefreshSession(
		uuid.New(),
		uuid.New(),
		"hashed-token",
		nil,
		nil,
		nil,
		time.Minute,
	)
	err := s.repo.Create(s.ctx, anotherSession)
	s.Require().Error(err)
}

func (s *RefreshSessionsRepoSuite) TestCreate_DuplicateHash() {
	// Create a session
	_ = s.repo.Create(s.ctx, s.testSession)

	// Trying to create a session with the same token hash
	var anotherSession, _ = model.NewRefreshSession(
		uuid.New(),
		s.testSession.AccountID(),
		s.testSession.RefreshTokenHash(),
		nil,
		nil,
		nil,
		time.Minute,
	)
	err := s.repo.Create(s.ctx, anotherSession)
	s.Require().Error(err)
}

func (s *RefreshSessionsRepoSuite) TestGetByHash() {
	// Create in advance
	_ = s.repo.Create(s.ctx, s.testSession)

	// Get by hash
	session, err := s.repo.GetByHash(s.ctx, s.testSession.RefreshTokenHash())
	s.Require().NoError(err)
	s.Require().Equal(s.testSession.ID(), session.ID())
	s.Require().Equal(s.testSession.AccountID(), session.AccountID())
}

func (s *RefreshSessionsRepoSuite) TestGetByHash_NotFound() {
	// Get a non-existing session
	_, err := s.repo.GetByHash(s.ctx, s.testSession.RefreshTokenHash())
	s.Require().Error(err)
	s.Require().ErrorIs(err, pkgerrs.ErrObjectNotFound)
}

func (s *RefreshSessionsRepoSuite) TestRevoke() {
	// Create in advance
	_ = s.repo.Create(s.ctx, s.testSession)

	var (
		revokedSession = *s.testSession
		reason         = "account is blocked"
	)
	_ = revokedSession.Revoke(&reason)

	// Revoke the session
	err := s.repo.Revoke(s.ctx, &revokedSession)
	s.Require().NoError(err)

	// Ensure the session has been revoked
	session, _ := s.repo.GetByID(s.ctx, s.testSession.ID())
	s.Require().Equal(revokedSession.RevokeReason(), session.RevokeReason())
}

func (s *RefreshSessionsRepoSuite) TestRevokeAllForAccount() {
	var anotherSession, _ = model.NewRefreshSession(
		uuid.New(),
		s.testSession.AccountID(),
		"hashed",
		nil,
		nil,
		nil,
		time.Minute,
	)

	// Create some sessions for the same account
	_ = s.repo.Create(s.ctx, s.testSession)
	_ = s.repo.Create(s.ctx, anotherSession)

	var reason = "tests"

	err := s.repo.RevokeAllForAccount(s.ctx, s.testSession.AccountID(), &reason)
	s.Require().NoError(err)

	// Ensure all sessions have been revoked
	sess, _ := s.repo.GetByID(s.ctx, s.testSession.ID())
	s.Require().Equal(reason, *sess.RevokeReason())

	sess, _ = s.repo.GetByID(s.ctx, anotherSession.ID())
	s.Require().Equal(reason, *sess.RevokeReason())
}

func (s *RefreshSessionsRepoSuite) TestRevokeDescendants() {
	// Create sessions - one is the descendant of the second
	var (
		rotatedID         = s.testSession.ID()
		anotherSession, _ = model.NewRefreshSession(
			uuid.New(),
			s.testSession.AccountID(),
			"hashed",
			&rotatedID,
			nil,
			nil,
			time.Minute,
		)
		reason = "test revoke"
	)
	_ = s.repo.Create(s.ctx, s.testSession)
	_ = s.repo.Create(s.ctx, anotherSession)

	err := s.repo.RevokeDescendants(s.ctx, s.testSession.ID(), &reason)
	s.Require().NoError(err)

	// Ensure the session has been revoked
	session, _ := s.repo.GetByHash(s.ctx, anotherSession.RefreshTokenHash())
	s.Require().Equal(reason, *session.RevokeReason())
}

func (s *RefreshSessionsRepoSuite) TestDeleteExpired() {
	// Create a session
	_ = s.repo.Create(s.ctx, s.testSession)

	// Delete expired (set time what is much later)
	var expiresAt = time.Now().Add(time.Hour)
	err := s.repo.DeleteExpired(s.ctx, expiresAt)
	s.Require().NoError(err)

	// Ensure it was deleted
	_, err = s.repo.GetByID(s.ctx, s.testSession.ID())
	s.Require().Error(err)
	s.Require().ErrorIs(err, pkgerrs.ErrObjectNotFound)
}

func (s *RefreshSessionsRepoSuite) TestListActiveForAccount() {
	const sessionsAmount = 2

	var anotherSession, _ = model.NewRefreshSession(
		uuid.New(),
		s.testSession.AccountID(),
		"hashed",
		nil,
		nil,
		nil,
		time.Minute,
	)

	// Create sessions
	_ = s.repo.Create(s.ctx, s.testSession)
	_ = s.repo.Create(s.ctx, anotherSession)

	// List of active
	sessions, err := s.repo.ListActiveForAccount(s.ctx, s.testSession.AccountID())
	s.Require().NoError(err)
	s.Require().Len(sessions, sessionsAmount)

	var fstFound, sndFound bool
	for i := range sessions {
		value := *sessions[i]
		if value.ID() == s.testSession.ID() {
			fstFound = true
		}
		if value.ID() == anotherSession.ID() {
			sndFound = true
		}
	}

	s.Require().Truef(fstFound, "expected %v\n in %v",
		s.testSession, sessions)
	s.Require().Truef(sndFound, "expected %v\n in %v",
		anotherSession, sessions)
}
