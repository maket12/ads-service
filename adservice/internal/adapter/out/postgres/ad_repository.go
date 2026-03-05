package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/maket12/ads-service/adservice/internal/adapter/out/postgres/mapper"
	"github.com/maket12/ads-service/adservice/internal/adapter/out/postgres/sqlc"
	"github.com/maket12/ads-service/adservice/internal/domain/model"
	pkgerrs "github.com/maket12/ads-service/pkg/errs"
	pkgpostgres "github.com/maket12/ads-service/pkg/postgres"

	"github.com/google/uuid"
)

type AdRepository struct {
	q *sqlc.Queries
}

func NewAdRepository(pgClient *pkgpostgres.Client) *AdRepository {
	queries := sqlc.New(pgClient.DB)
	return &AdRepository{q: queries}
}

func (r *AdRepository) Create(ctx context.Context, ad *model.Ad) error {
	params := mapper.MapAdToSQLCCreate(ad)
	return r.q.CreateAd(ctx, params)
}

func (r *AdRepository) Get(ctx context.Context, id uuid.UUID) (*model.Ad, error) {
	rawAd, err := r.q.GetAd(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, pkgerrs.NewObjectNotFoundError("ad", id)
		}
		return nil, err
	}

	ad := mapper.MapSQLCToAd(rawAd)

	return ad, nil
}

func (r *AdRepository) Update(ctx context.Context, ad *model.Ad) error {
	params := mapper.MapAdToSQLCUpdate(ad)
	return r.q.UpdateAd(ctx, params)
}

func (r *AdRepository) UpdateStatus(ctx context.Context, ad *model.Ad) error {
	params := mapper.MapAdToSQLCUpdateStatus(ad)
	return r.q.UpdateAdStatus(ctx, params)
}

func (r *AdRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteAd(ctx, id)
}

func (r *AdRepository) DeleteAll(ctx context.Context, sellerID uuid.UUID) error {
	return r.q.DeleteAllAds(ctx, sellerID)
}

func (r *AdRepository) ListAds(ctx context.Context, limit, offset int) ([]*model.Ad, error) {
	params := mapper.MapToSQLCList(limit, offset)

	rawAds, err := r.q.ListAds(ctx, params)
	if err != nil {
		return nil, err
	}

	return mapper.MapSQLCToAdsList(rawAds), nil
}

func (r *AdRepository) ListSellerAds(ctx context.Context, sellerID uuid.UUID, limit, offset int) ([]*model.Ad, error) {
	params := mapper.MapToSQLCSellerList(sellerID, limit, offset)

	rawAds, err := r.q.ListSellerAds(ctx, params)
	if err != nil {
		return nil, err
	}

	return mapper.MapSQLCToAdsList(rawAds), nil
}
