package postgres

import (
	"context"
	"database/sql"
	"errors"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/maket12/ads-service/adservice/internal/adapter/out/postgres/mapper"
	"github.com/maket12/ads-service/adservice/internal/domain/model"
	pkgerrs "github.com/maket12/ads-service/adservice/pkg/errs"
	pkgpostgres "github.com/maket12/ads-service/adservice/pkg/postgres"

	"github.com/google/uuid"
)

type AdRepository struct{ BaseRepository }

func NewAdRepository(
	pgClient *pkgpostgres.Client,
	getter *trmpgx.CtxGetter,
) *AdRepository {
	return &AdRepository{
		BaseRepository: NewBaseRepository(pgClient, getter),
	}
}

func (r *AdRepository) Create(ctx context.Context, ad *model.Ad) error {
	params := mapper.MapAdToSQLCCreate(ad)
	return r.q.CreateAd(ctx, r.db(ctx), params)
}

func (r *AdRepository) Get(ctx context.Context, id uuid.UUID) (*model.Ad, error) {
	rawAd, err := r.q.GetAd(ctx, r.db(ctx),
		pgtype.UUID{
			Bytes: id,
			Valid: true,
		},
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, pkgerrs.NewObjectNotFoundError("ad", id)
		}
		return nil, err
	}

	return mapper.MapSQLCToAd(rawAd), nil
}

func (r *AdRepository) Update(ctx context.Context, ad *model.Ad) error {
	params := mapper.MapAdToSQLCUpdate(ad)
	return r.q.UpdateAd(ctx, r.db(ctx), params)
}

func (r *AdRepository) UpdateStatus(ctx context.Context, ad *model.Ad) error {
	params := mapper.MapAdToSQLCUpdateStatus(ad)
	return r.q.UpdateAdStatus(ctx, r.db(ctx), params)
}

func (r *AdRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteAd(ctx, r.db(ctx),
		pgtype.UUID{
			Bytes: id,
			Valid: true,
		},
	)
}

func (r *AdRepository) DeleteAll(ctx context.Context, sellerID uuid.UUID) error {
	return r.q.DeleteAllAds(ctx, r.db(ctx),
		pgtype.UUID{
			Bytes: sellerID,
			Valid: true,
		},
	)
}

func (r *AdRepository) ListAds(ctx context.Context, limit, offset int) ([]*model.Ad, error) {
	params := mapper.MapToSQLCList(limit, offset)

	rawAds, err := r.q.ListAds(ctx, r.db(ctx), params)
	if err != nil {
		return nil, err
	}

	return mapper.MapSQLCToAdsList(rawAds), nil
}

func (r *AdRepository) ListSellerAds(ctx context.Context, sellerID uuid.UUID, limit, offset int) ([]*model.Ad, error) {
	params := mapper.MapToSQLCSellerList(sellerID, limit, offset)

	rawAds, err := r.q.ListSellerAds(ctx, r.db(ctx), params)
	if err != nil {
		return nil, err
	}

	return mapper.MapSQLCToAdsList(rawAds), nil
}
