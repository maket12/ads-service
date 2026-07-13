//go:build integration

package redis_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/maket12/ads-service/authservice/internal/domain/model"
	pkgerrs "github.com/maket12/ads-service/authservice/pkg/errs"
	"github.com/stretchr/testify/suite"
)

type VerificationTokenRepoSuite struct {
	BaseRepoSuite
	repo      *VerificationTokenRepository
	testToken *model.VerificationToken
}

func TestVerificationTokenRepoSuite(t *testing.T) { suite.Run(t, new(VerificationTokenRepoSuite)) }

func (s *VerificationTokenRepoSuite) SetupSuite() {
	s.SetupBase()
	s.repo = NewVerificationTokenRepository(s.redisClient)
}

func (s *VerificationTokenRepoSuite) SetupTest() {
	err := s.redisContainer.FlushAll(s.ctx)
	s.Require().NoError(err)

	s.testToken, _ = model.NewVerificationToken(uuid.New(), time.Minute*15)
}

func (s *VerificationTokenRepoSuite) TestSaveGet() {
	// Check save first
	err := s.repo.Save(s.ctx, s.testToken)
	s.Require().NoError(err)

	// And then get
	token, err := s.repo.Get(s.ctx, s.testToken.Token())
	s.Require().NoError(err)
	s.Require().Exactly(s.testToken.Token(), token.Token())
	s.Require().Exactly(s.testToken.AccountID(), token.AccountID())
}

func (s *VerificationTokenRepoSuite) TestGet_NotFound() {
	_, err := s.repo.Get(s.ctx, "non-existing-token")

	s.Require().Error(err)
	s.Require().ErrorIs(err, pkgerrs.ErrObjectNotFound)
}

func (s *VerificationTokenRepoSuite) TestDelete() {
	// Save token first
	_ = s.repo.Save(s.ctx, s.testToken)

	// Delete it
	err := s.repo.Delete(s.ctx, s.testToken.Token())
	s.Require().NoError(err)

	// Check it's gone
	_, err = s.repo.Get(s.ctx, s.testToken.Token())
	s.Require().Error(err)
	s.Require().ErrorIs(err, pkgerrs.ErrObjectNotFound)
}

func (s *VerificationTokenRepoSuite) TestSave_TTLExpiration() {
	// Save a token with a very short TTL
	shortToken := model.RestoreVerificationToken(
		"token",
		uuid.New(),
		time.Millisecond*100,
		time.Now().Add(time.Millisecond*100),
	)
	err := s.repo.Save(s.ctx, shortToken)
	s.Require().NoError(err)

	// Wait for it to expire
	time.Sleep(time.Millisecond * 300)

	_, err = s.repo.Get(s.ctx, shortToken.Token())
	s.Require().Error(err)
	s.Require().ErrorIs(err, pkgerrs.ErrObjectNotFound)
}
