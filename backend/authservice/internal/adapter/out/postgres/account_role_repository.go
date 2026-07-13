package postgres

import (
	"context"
	"database/sql"
	"errors"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/maket12/ads-service/authservice/internal/adapter/out/postgres/mapper"
	"github.com/maket12/ads-service/authservice/internal/domain/model"
	pkgerrs "github.com/maket12/ads-service/authservice/pkg/errs"
	pkgpostgres "github.com/maket12/ads-service/authservice/pkg/postgres"

	"github.com/google/uuid"
)

type AccountRoleRepository struct{ BaseRepository }

func NewAccountRolesRepository(
	pgClient *pkgpostgres.Client,
	getter *trmpgx.CtxGetter,
) *AccountRoleRepository {
	return &AccountRoleRepository{
		BaseRepository: NewBaseRepository(pgClient, getter),
	}
}

func (r *AccountRoleRepository) Create(ctx context.Context, accountRole *model.AccountRole) error {
	params := mapper.MapAccountRoleToSQLCCreate(accountRole)
	return r.q.CreateAccountRole(ctx, r.db(ctx), params)
}

func (r *AccountRoleRepository) Get(ctx context.Context, accountID uuid.UUID) (*model.AccountRole, error) {
	rawAccRole, err := r.q.GetAccountRole(ctx, r.db(ctx),
		pgtype.UUID{Bytes: accountID, Valid: true},
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, pkgerrs.NewObjectNotFoundError("account_role", accountID)
		}
		return nil, err
	}
	return mapper.MapSQLCToAccountRole(rawAccRole), nil
}

func (r *AccountRoleRepository) Update(ctx context.Context, accountRole *model.AccountRole) error {
	params := mapper.MapAccountRoleToSQLCUpdate(accountRole)
	return r.q.UpdateAccountRole(ctx, r.db(ctx), params)
}

func (r *AccountRoleRepository) Delete(ctx context.Context, accountID uuid.UUID) error {
	return r.q.DeleteAccountRole(ctx, r.db(ctx),
		pgtype.UUID{Bytes: accountID, Valid: true},
	)
}
