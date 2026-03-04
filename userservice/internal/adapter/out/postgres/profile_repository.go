package postgres

import (
	"context"
	"database/sql"
	"errors"
	pkgerrs "github.com/maket12/ads-service/pkg/errs"
	pkgpostgres "github.com/maket12/ads-service/pkg/postgres"
	"github.com/maket12/ads-service/userservice/internal/adapter/out/postgres/mapper"
	"github.com/maket12/ads-service/userservice/internal/adapter/out/postgres/sqlc"
	"github.com/maket12/ads-service/userservice/internal/domain/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type ProfileRepository struct {
	q *sqlc.Queries
}

func NewProfileRepository(pgClient *pkgpostgres.Client) *ProfileRepository {
	queries := sqlc.New(pgClient.DB)
	return &ProfileRepository{q: queries}
}

func (r *ProfileRepository) Create(ctx context.Context, profile *model.Profile) error {
	params := mapper.MapProfileToSQLCCreate(profile)

	err := r.q.CreateProfile(ctx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return pkgerrs.NewObjectAlreadyExistsErrorWithReason(
					"profile", pgErr,
				)
			}
		}
		return err
	}

	return nil
}

func (r *ProfileRepository) Get(ctx context.Context, accountID uuid.UUID) (*model.Profile, error) {
	rawProfile, err := r.q.GetProfile(ctx, accountID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, pkgerrs.NewObjectNotFoundError("profile", accountID)
		}
		return nil, err
	}
	return mapper.MapSQLCToProfile(rawProfile), nil
}

func (r *ProfileRepository) Update(ctx context.Context, profile *model.Profile) error {
	params := mapper.MapProfileToSQLCUpdate(profile)
	return r.q.UpdateProfile(ctx, params)
}

func (r *ProfileRepository) Delete(ctx context.Context, accountID uuid.UUID) error {
	return r.q.DeleteProfile(ctx, accountID)
}

func (r *ProfileRepository) ListProfiles(ctx context.Context, limit, offset int) ([]*model.Profile, error) {
	params := mapper.MapToSQLCList(limit, offset)

	rawProfiles, err := r.q.ListProfiles(ctx, params)
	if err != nil {
		return nil, err
	}

	return mapper.MapSQLCToProfilesList(rawProfiles), nil
}
