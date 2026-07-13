package mapper

import (
	"database/sql"

	"github.com/maket12/ads-service/adservice/internal/adapter/out/postgres/sqlc"
	sqlc2 "github.com/maket12/ads-service/backend/adservice/internal/adapter/out/postgres/sqlc"
	"github.com/maket12/ads-service/backend/adservice/internal/domain/model"

	"github.com/google/uuid"
)

func MapSQLCToAd(rawAd sqlc2.Ad) *model.Ad {
	var description *string
	if rawAd.Description.Valid {
		description = &rawAd.Description.String
	}

	return model.RestoreAd(
		rawAd.ID,
		rawAd.SellerID,
		rawAd.Title,
		description,
		rawAd.Price,
		model.AdStatus(rawAd.Status),
		nil,
		rawAd.CreatedAt,
		rawAd.UpdatedAt,
	)
}

func MapAdToSQLCCreate(ad *model.Ad) sqlc2.CreateAdParams {
	var description sql.NullString
	if ad.Description() != nil {
		description = sql.NullString{
			String: *ad.Description(),
			Valid:  true,
		}
	}

	return sqlc2.CreateAdParams{
		ID:          ad.ID(),
		SellerID:    ad.SellerID(),
		Title:       ad.Title(),
		Description: description,
		Price:       ad.Price(),
		Status:      sqlc.AdStatus(ad.Status()),
		CreatedAt:   ad.CreatedAt(),
		UpdatedAt:   ad.UpdatedAt(),
	}
}

func MapAdToSQLCUpdate(ad *model.Ad) sqlc2.UpdateAdParams {
	var description sql.NullString
	if ad.Description() != nil {
		description = sql.NullString{
			String: *ad.Description(),
			Valid:  true,
		}
	}

	return sqlc2.UpdateAdParams{
		ID:          ad.ID(),
		Title:       ad.Title(),
		Description: description,
		Price:       ad.Price(),
		UpdatedAt:   ad.UpdatedAt(),
	}
}

func MapAdToSQLCUpdateStatus(ad *model.Ad) sqlc2.UpdateAdStatusParams {
	return sqlc2.UpdateAdStatusParams{
		ID:     ad.ID(),
		Status: sqlc.AdStatus(ad.Status()),
	}
}

func MapToSQLCList(limit, offset int) sqlc2.ListAdsParams {
	return sqlc2.ListAdsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	}
}

func MapSQLCToAdsList(rawAds []sqlc2.Ad) []*model.Ad {
	ads := make([]*model.Ad, 0, len(rawAds))
	for _, rawAd := range rawAds {
		ad := MapSQLCToAd(rawAd)
		ads = append(ads, ad)
	}
	return ads
}

func MapToSQLCSellerList(sellerID uuid.UUID, limit, offset int) sqlc2.ListSellerAdsParams {
	return sqlc2.ListSellerAdsParams{
		SellerID: sellerID,
		Limit:    int32(limit),
		Offset:   int32(offset),
	}
}
