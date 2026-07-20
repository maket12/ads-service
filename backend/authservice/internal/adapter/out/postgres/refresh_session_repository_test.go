//go:build integration

package postgres_test

import (
	"testing"
	"time"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/google/uuid"
	adapterpg "github.com/maket12/ads-service/authservice/internal/adapter/out/postgres"
	"github.com/maket12/ads-service/authservice/internal/domain/model"
	pkgerrs "github.com/maket12/ads-service/authservice/pkg/errs"
	"github.com/stretchr/testify/suite"
)

type RefreshSessionsRepoSuite struct {
	BaseRepoSuite
	repo        *adapterpg.RefreshSessionRepository
	testSession *model.RefreshSession
}

func TestRefreshSessionRepoSuite(t *testing.T) {
	suite.Run(t, new(RefreshSessionsRepoSuite))
}

func (s *RefreshSessionsRepoSuite) SetupSuite() {
	s.SetupBase(2)
	s.repo = adapterpg.NewRefreshSessionsRepository(s.dbClient,
		trmpgx.DefaultCtxGetter,
	)
}

func (s *RefreshSessionsRepoSuite) SetupTest() {
	err := s.pgContainer.TruncateTables(s.ctx, "refresh_sessions")
	s.Require().NoError(err)

	s.seedData()
}

func (s *RefreshSessionsRepoSuite) seedData() {
	testAccount, _ := model.NewAccount("new@email.com", "hashed-secret-pass")

	accountsRepo := adapterpg.NewAccountsRepository(s.dbClient,
		trmpgx.DefaultCtxGetter,
	)
	err := accountsRepo.Create(s.ctx, testAccount)
	s.Require().NoError(err)

	s.testSession, _ = model.NewRefreshSession(
		uuid.New(),
		testAccount.ID(),
		"hashed-secret-token",
		nil,
		nil,
		nil,
		time.Minute,
	)
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

func (s *RefreshSessionsRepoSuite) TestUpdate() {
	// Create in advance
	_ = s.repo.Create(s.ctx, s.testSession)

	var revokedSession = *s.testSession
	_ = revokedSession.RevokeByLogout()

	// Update the session
	err := s.repo.Update(s.ctx, &revokedSession)
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
