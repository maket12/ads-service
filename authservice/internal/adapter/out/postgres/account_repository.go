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
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type AccountRepository struct {
	q *sqlc.Queries
}

func NewAccountsRepository(pgClient *pkgpostgres.Client) *AccountRepository {
	queries := sqlc.New(pgClient.DB)
	return &AccountRepository{q: queries}
}

func (r *AccountRepository) Create(ctx context.Context, account *model.Account) error {
	params := mapper.MapAccountToSQLCCreate(account)
	err := r.q.CreateAccount(ctx, params)

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
	rawAcc, err := r.q.GetAccountByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, pkgerrs.NewObjectNotFoundError("account", email)
		}
		return nil, err
	}

	account := mapper.MapSQLCToAccount(rawAcc)

	return account, nil
}

func (r *AccountRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Account, error) {
	rawAcc, err := r.q.GetAccountByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, pkgerrs.NewObjectNotFoundError("account", id)
		}
		return nil, err
	}

	account := mapper.MapSQLCToAccount(rawAcc)

	return account, nil
}

func (r *AccountRepository) MarkLogin(ctx context.Context, account *model.Account) error {
	var lastLoginTime time.Time
	if account.LastLoginAt() != nil {
		lastLoginTime = *account.LastLoginAt()
	}

	var params = sqlc.MarkAccountLoginParams{
		ID: account.ID(),
		LastLoginAt: sql.NullTime{
			Time:  lastLoginTime,
			Valid: true,
		},
		UpdatedAt: account.UpdatedAt(),
	}

	if err := r.q.MarkAccountLogin(ctx, params); err != nil {
		return err
	}

	return nil
}

func (r *AccountRepository) VerifyEmail(ctx context.Context, account *model.Account) error {
	params := sqlc.VerifyAccountEmailParams{
		ID:        account.ID(),
		UpdatedAt: account.UpdatedAt(),
	}
	return r.q.VerifyAccountEmail(ctx, params)
}
