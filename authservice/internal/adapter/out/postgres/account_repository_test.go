//go:build integration

package postgres_test

import (
	"strings"
	"testing"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	adapterpostgres "github.com/maket12/ads-service/authservice/internal/adapter/out/postgres"
	"github.com/maket12/ads-service/authservice/internal/domain/model"
	pkgerrs "github.com/maket12/ads-service/authservice/pkg/errs"
	"github.com/stretchr/testify/suite"
)

type AccountsRepoSuite struct {
	BaseRepoSuite
	repo        *adapterpostgres.AccountRepository
	testAccount *model.Account
}

func TestAccountsRepoSuite(t *testing.T) { suite.Run(t, new(AccountsRepoSuite)) }

func (s *AccountsRepoSuite) SetupSuite() {
	s.SetupBase(1)
	s.repo = adapterpostgres.NewAccountsRepository(s.dbClient,
		trmpgx.DefaultCtxGetter,
	)
	s.testAccount, _ = model.NewAccount("new@email.com", "hashed-secret-pass")
}

func (s *AccountsRepoSuite) SetupTest() {
	err := s.pgContainer.TruncateTables(s.ctx, "accounts")
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

func (s *AccountsRepoSuite) TestUpdate() {
	// Create account at first
	_ = s.repo.Create(s.ctx, s.testAccount)

	// Mutate it
	_ = s.testAccount.MarkLogin()
	s.testAccount.VerifyEmail()
	_ = s.testAccount.Block()

	// Update its state in database
	err := s.repo.Update(s.ctx, s.testAccount)
	s.Require().NoError(err)

	// Check if the account was updated
	acc, _ := s.repo.GetByEmail(s.ctx, s.testAccount.Email())
	s.Require().NotNil(acc.LastLoginAt())
	s.Require().True(acc.EmailVerified())
	s.Require().True(acc.Status() == model.AccountBlocked)
}
