package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/maket12/ads-service/authservice/internal/adapter/out/postgres/mapper"
	"github.com/maket12/ads-service/authservice/internal/adapter/out/postgres/sqlc"
	"github.com/maket12/ads-service/authservice/internal/domain/model"
	pkgerrs "github.com/maket12/ads-service/authservice/pkg/errs"
	pkgpostgres "github.com/maket12/ads-service/authservice/pkg/postgres"

	"github.com/google/uuid"
)

type RefreshSessionRepository struct{ BaseRepository }

func NewRefreshSessionsRepository(
	pgClient *pkgpostgres.Client,
	getter *trmpgx.CtxGetter,
) *RefreshSessionRepository {
	return &RefreshSessionRepository{
		BaseRepository: NewBaseRepository(pgClient, getter),
	}
}

func (r *RefreshSessionRepository) Create(ctx context.Context, session *model.RefreshSession) error {
	params := mapper.MapRefreshSessionToSQLCCreate(session)
	return r.q.CreateRefreshSession(ctx, r.db(ctx), params)
}

func (r *RefreshSessionRepository) GetByHash(ctx context.Context, tokenHash string) (*model.RefreshSession, error) {
	rawSession, err := r.q.GetRefreshSessionByHash(ctx, r.db(ctx), tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, pkgerrs.NewObjectNotFoundError("refresh_session", tokenHash)
		}
		return nil, err
	}
	refreshSession := mapper.MapSQLCToRefreshSession(rawSession)
	return refreshSession, nil
}

func (r *RefreshSessionRepository) GetByID(ctx context.Context, tokenID uuid.UUID) (*model.RefreshSession, error) {
	rawSession, err := r.q.GetRefreshSessionByID(
		ctx, r.db(ctx),
		pgtype.UUID{Bytes: tokenID, Valid: true},
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, pkgerrs.NewObjectNotFoundError("refresh_session", tokenID)
		}
		return nil, err
	}
	return mapper.MapSQLCToRefreshSession(rawSession), nil
}

func (r *RefreshSessionRepository) Update(ctx context.Context, session *model.RefreshSession) error {
	params := mapper.MapRefreshSessionToSQLCUpdate(session)
	return r.q.UpdateRefreshSession(ctx, r.db(ctx), params)
}

func (r *RefreshSessionRepository) RevokeAllForAccount(ctx context.Context, accountID uuid.UUID, reason *string) error {
	params := mapper.MapToSQLCRevokeAllForAccount(accountID, reason)
	return r.q.RevokeAllAccountRefreshSessions(ctx, r.db(ctx), params)
}

func (r *RefreshSessionRepository) RevokeDescendants(ctx context.Context, sessionID uuid.UUID, reason *string) error {
	params := mapper.MapToSQLCRevokeDescendants(sessionID, reason)
	return r.q.RevokeRefreshSessionDescendants(ctx, r.db(ctx), params)
}

func (r *RefreshSessionRepository) DeleteExpired(ctx context.Context, expiresAt time.Time) error {
	return r.q.DeleteExpiredRefreshSessions(ctx, r.db(ctx),
		pgtype.Timestamptz{Time: expiresAt, Valid: true},
	)
}

func (r *RefreshSessionRepository) ListActiveForAccount(ctx context.Context, accountID uuid.UUID) ([]*model.RefreshSession, error) {
	rawList, err := r.q.ListAccountActiveRefreshSessions(
		ctx, r.db(ctx),
		sqlc.ListAccountActiveRefreshSessionsParams{
			AccountID: pgtype.UUID{Bytes: accountID, Valid: true},
			ExpiresAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		},
	)
	if err != nil {
		return nil, err
	}
	return mapper.MapSQLCToListRefreshSession(rawList), nil
}
