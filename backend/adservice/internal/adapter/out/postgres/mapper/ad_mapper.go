package mapper

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/maket12/ads-service/adservice/internal/adapter/out/postgres/sqlc"
	"github.com/maket12/ads-service/adservice/internal/domain/model"

	"github.com/google/uuid"
)

func MapSQLCToAd(rawAd sqlc.Ad) *model.Ad {
	var (
		description *string
		updAt       *time.Time
	)

	if rawAd.Description.Valid {
		description = &rawAd.Description.String
	}

	if rawAd.UpdatedAt.Valid {
		updAt = &rawAd.UpdatedAt.Time
	}

	return model.RestoreAd(
		rawAd.ID.Bytes,
		rawAd.SellerID.Bytes,
		rawAd.Title,
		description,
		rawAd.Price,
		model.AdStatus(rawAd.Status),
		nil,
		rawAd.CreatedAt.Time,
		updAt,
	)
}

func MapAdToSQLCCreate(ad *model.Ad) sqlc.CreateAdParams {
	var (
		description pgtype.Text
		updAt       pgtype.Timestamptz
	)

	if ad.Description() != nil {
		description = pgtype.Text{
			String: *ad.Description(),
			Valid:  true,
		}
	}

	if ad.UpdatedAt() != nil {
		updAt = pgtype.Timestamptz{
			Time:  *ad.UpdatedAt(),
			Valid: true,
		}
	}

	return sqlc.CreateAdParams{
		ID: pgtype.UUID{
			Bytes: ad.ID(),
			Valid: true,
		},
		SellerID: pgtype.UUID{
			Bytes: ad.SellerID(),
			Valid: true,
		},
		Title:       ad.Title(),
		Description: description,
		Price:       ad.Price(),
		Status:      ad.Status().String(),
		CreatedAt: pgtype.Timestamptz{
			Time:  ad.CreatedAt(),
			Valid: true,
		},
		UpdatedAt: updAt,
	}
}

func MapAdToSQLCUpdate(ad *model.Ad) sqlc.UpdateAdParams {
	var (
		description pgtype.Text
		updAt       pgtype.Timestamptz
	)

	if ad.Description() != nil {
		description = pgtype.Text{
			String: *ad.Description(),
			Valid:  true,
		}
	}

	if ad.UpdatedAt() != nil {
		updAt = pgtype.Timestamptz{
			Time:  *ad.UpdatedAt(),
			Valid: true,
		}
	}

	return sqlc.UpdateAdParams{
		ID: pgtype.UUID{
			Bytes: ad.ID(),
			Valid: true,
		},
		SellerID: pgtype.UUID{
			Bytes: ad.SellerID(),
			Valid: true,
		},
		Title:       ad.Title(),
		Description: description,
		Price:       ad.Price(),
		Status:      ad.Status().String(),
		CreatedAt: pgtype.Timestamptz{
			Time:  ad.CreatedAt(),
			Valid: true,
		},
		UpdatedAt: updAt,
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
		SellerID: pgtype.UUID{
			Bytes: sellerID,
			Valid: true,
		},
		Limit:  int32(limit),
		Offset: int32(offset),
	}
}
