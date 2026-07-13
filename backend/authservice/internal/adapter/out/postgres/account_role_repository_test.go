//go:build integration

package postgres_test

import (
	"testing"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/google/uuid"
	"github.com/maket12/ads-service/authservice/internal/domain/model"
	pkgerrs "github.com/maket12/ads-service/authservice/pkg/errs"
	"github.com/stretchr/testify/suite"
)

type AccountRolesRepoSuite struct {
	BaseRepoSuite
	repo     *AccountRoleRepository
	testRole *model.AccountRole
}

func TestAccountRolesRepoSuite(t *testing.T) {
	suite.Run(t, new(AccountRolesRepoSuite))
}

func (s *AccountRolesRepoSuite) SetupSuite() {
	s.SetupBase(3)
	s.repo = NewAccountRolesRepository(s.dbClient,
		trmpgx.DefaultCtxGetter,
	)
}

func (s *AccountRolesRepoSuite) SetupTest() {
	err := s.pgContainer.TruncateTables(s.ctx, "account_roles")
	s.Require().NoError(err)

	s.seedData()
}

func (s *AccountRolesRepoSuite) seedData() {
	testAccount, _ := model.NewAccount("new@email.com", "hashed-secret-pass")

	accountsRepo := NewAccountsRepository(s.dbClient,
		trmpgx.DefaultCtxGetter,
	)
	_ = accountsRepo.Create(s.ctx, testAccount)

	s.testRole, _ = model.NewAccountRole(testAccount.ID())
}

func (s *AccountRolesRepoSuite) TestCreateGet() {
	// Create at first
	err := s.repo.Create(s.ctx, s.testRole)
	s.Require().NoError(err)

	// Get by account id
	role, err := s.repo.Get(s.ctx, s.testRole.AccountID())
	s.Require().NoError(err)
	s.Require().Equal(s.testRole.Role(), role.Role())
}

func (s *AccountRolesRepoSuite) TestCreate_NonExistingAccount() {
	// Create an account role for unexisting account
	newRole, _ := model.NewAccountRole(uuid.New())
	err := s.repo.Create(s.ctx, newRole)
	s.Require().Error(err)
}

func (s *AccountRolesRepoSuite) TestGet_NotFound() {
	// Try to get non-existing account role
	_, err := s.repo.Get(s.ctx, s.testRole.AccountID())
	s.Require().Error(err)
	s.Require().ErrorIs(err, pkgerrs.ErrObjectNotFound)
}

func (s *AccountRolesRepoSuite) TestUpdate() {
	// Create at first
	_ = s.repo.Create(s.ctx, s.testRole)

	// Copy value and assigned to not change test data
	assignedRole := *s.testRole
	_ = assignedRole.Assign("admin")

	err := s.repo.Update(s.ctx, &assignedRole)
	s.Require().NoError(err)

	// Ensure update was correct
	acc, _ := s.repo.Get(s.ctx, s.testRole.AccountID())
	s.Require().Equal(model.RoleAdmin, acc.Role())
}

func (s *AccountRolesRepoSuite) TestDelete() {
	// Create at first
	_ = s.repo.Create(s.ctx, s.testRole)

	// Delete
	err := s.repo.Delete(s.ctx, s.testRole.AccountID())
	s.Require().NoError(err)

	// Ensure deletion was successful
	_, err = s.repo.Get(s.ctx, s.testRole.AccountID())
	s.Require().Error(err)
	s.Require().ErrorIs(err, pkgerrs.ErrObjectNotFound)
}
