package mapper

import (
	"database/sql"

	"github.com/maket12/ads-service/adservice/internal/adapter/out/postgres/sqlc"
	"github.com/maket12/ads-service/adservice/internal/domain/model"

	"github.com/google/uuid"
)

func MapSQLCToAd(rawAd sqlc.Ad) *model.Ad {
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

func MapAdToSQLCCreate(ad *model.Ad) sqlc.CreateAdParams {
	var description sql.NullString
	if ad.Description() != nil {
		description = sql.NullString{
			String: *ad.Description(),
			Valid:  true,
		}
	}

	return sqlc.CreateAdParams{
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

func MapAdToSQLCUpdate(ad *model.Ad) sqlc.UpdateAdParams {
	var description sql.NullString
	if ad.Description() != nil {
		description = sql.NullString{
			String: *ad.Description(),
			Valid:  true,
		}
	}

	return sqlc.UpdateAdParams{
		ID:          ad.ID(),
		Title:       ad.Title(),
		Description: description,
		Price:       ad.Price(),
		UpdatedAt:   ad.UpdatedAt(),
	}
}

func MapAdToSQLCUpdateStatus(ad *model.Ad) sqlc.UpdateAdStatusParams {
	return sqlc.UpdateAdStatusParams{
		ID:     ad.ID(),
		Status: sqlc.AdStatus(ad.Status()),
	}
}

func MapToSQLCList(limit, offset int) sqlc.ListAdsParams {
	return sqlc.ListAdsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	}
}

func MapSQLCToAdsList(rawAds []sqlc.Ad) []*model.Ad {
	ads := make([]*model.Ad, 0, len(rawAds))
	for _, rawAd := range rawAds {
		ad := MapSQLCToAd(rawAd)
		ads = append(ads, ad)
	}
	return ads
}

func MapToSQLCSellerList(sellerID uuid.UUID, limit, offset int) sqlc.ListSellerAdsParams {
	return sqlc.ListSellerAdsParams{
		SellerID: sellerID,
		Limit:    int32(limit),
		Offset:   int32(offset),
	}
}
