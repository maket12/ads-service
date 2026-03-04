package postgres

import (
	"context"
	"database/sql"
	"errors"
	"github.com/maket12/ads-service/authservice/internal/adapter/out/postgres/mapper"
	"github.com/maket12/ads-service/authservice/internal/adapter/out/postgres/sqlc"
	"github.com/maket12/ads-service/authservice/internal/domain/model"
	pkgerrs "github.com/maket12/ads-service/pkg/errs"
	pkgpostgres "github.com/maket12/ads-service/pkg/postgres"

	"github.com/google/uuid"
)

type AccountRoleRepository struct {
	q *sqlc.Queries
}

func NewAccountRolesRepository(pgClient *pkgpostgres.Client) *AccountRoleRepository {
	queries := sqlc.New(pgClient.DB)
	return &AccountRoleRepository{q: queries}
}

func (r *AccountRoleRepository) Create(ctx context.Context, accountRole *model.AccountRole) error {
	params := mapper.MapAccountRoleToSQLCCreate(accountRole)
	return r.q.CreateAccountRole(ctx, params)
}

func (r *AccountRoleRepository) Get(ctx context.Context, accountID uuid.UUID) (*model.AccountRole, error) {
	rawAccRole, err := r.q.GetAccountRole(ctx, accountID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, pkgerrs.NewObjectNotFoundError("account_role", accountID)
		}
		return nil, err
	}

	accountRole := mapper.MapSQLCToAccountRole(rawAccRole)

	return accountRole, nil
}

func (r *AccountRoleRepository) Update(ctx context.Context, accountRole *model.AccountRole) error {
	var params = sqlc.UpdateAccountRoleParams{
		AccountID: accountRole.AccountID(),
		Role:      sqlc.RoleType(accountRole.Role()),
	}

	if err := r.q.UpdateAccountRole(ctx, params); err != nil {
		return err
	}

	return nil
}

func (r *AccountRoleRepository) Delete(ctx context.Context, accountID uuid.UUID) error {
	return r.q.DeleteAccountRole(ctx, accountID)
}
