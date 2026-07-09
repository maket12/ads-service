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
	"github.com/jackc/pgx/v5/pgconn"
)

type AccountRepository struct{ BaseRepository }

func NewAccountsRepository(
	pgClient *pkgpostgres.Client,
	getter *trmpgx.CtxGetter,
) *AccountRepository {
	return &AccountRepository{
		BaseRepository: NewBaseRepository(pgClient, getter),
	}
}

func (r *AccountRepository) Create(ctx context.Context, account *model.Account) error {
	params := mapper.MapAccountToSQLCCreate(account)

	err := r.q.CreateAccount(ctx, r.db(ctx), params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return pkgerrs.NewObjectAlreadyExistsErrorWithReason(
					"account", pgErr,
				)
			}
		}
		return err
	}

	return nil
}

func (r *AccountRepository) GetByEmail(ctx context.Context, email string) (*model.Account, error) {
	rawAcc, err := r.q.GetAccountByEmail(ctx, r.db(ctx), email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, pkgerrs.NewObjectNotFoundError("account", email)
		}
		return nil, err
	}

	return mapper.MapSQLCToAccount(rawAcc), nil
}

func (r *AccountRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Account, error) {
	rawAcc, err := r.q.GetAccountByID(ctx, r.db(ctx),
		pgtype.UUID{Bytes: id, Valid: true},
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, pkgerrs.NewObjectNotFoundError("account", id)
		}
		return nil, err
	}

	return mapper.MapSQLCToAccount(rawAcc), nil
}

func (r *AccountRepository) MarkLogin(ctx context.Context, account *model.Account) error {
	params := mapper.MapAccountToSQLCMarkLogin(account)
	return r.q.MarkAccountLogin(ctx, r.db(ctx), params)
}

func (r *AccountRepository) VerifyEmail(ctx context.Context, account *model.Account) error {
	params := mapper.MapAccountToSQLCVerifyEmail(account)
	return r.q.VerifyAccountEmail(ctx, r.db(ctx), params)
}
