package mapper_test

import (
	"reflect"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/maket12/ads-service/adservice/internal/adapter/out/postgres/mapper"
	"github.com/maket12/ads-service/adservice/internal/adapter/out/postgres/sqlc"
	"github.com/maket12/ads-service/adservice/internal/domain/model"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapSQLCToAd(t *testing.T) {
	id := uuid.New()
	sellerID := uuid.New()
	description := gofakeit.Bio()
	price := int64(gofakeit.Price(1000, 1000000))
	createdAt := gofakeit.Date()
	updatedAt := gofakeit.Date()

	raw := sqlc.Ad{
		ID:       pgtype.UUID{Bytes: id, Valid: true},
		SellerID: pgtype.UUID{Bytes: sellerID, Valid: true},
		Title:    gofakeit.ProductName(),
		Description: pgtype.Text{
			String: description,
			Valid:  true,
		},
		Price:  price,
		Status: model.AdPublished.String(),
		CreatedAt: pgtype.Timestamptz{
			Time:  createdAt,
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:  updatedAt,
			Valid: true,
		},
	}

	expected := model.RestoreAd(
		id,
		sellerID,
		raw.Title,
		&description,
		price,
		model.AdPublished,
		nil,
		createdAt,
		&updatedAt,
	)

	ad := mapper.MapSQLCToAd(raw)

	require.NotNil(t, ad)
	assert.True(t, reflect.DeepEqual(expected, ad))
}

func TestMapSQLCToAd_NilDescriptionAndUpdatedAt(t *testing.T) {
	id := uuid.New()
	sellerID := uuid.New()
	title := gofakeit.ProductName()
	price := int64(gofakeit.Price(1000, 1000000))
	createdAt := gofakeit.Date()

	raw := sqlc.Ad{
		ID:          pgtype.UUID{Bytes: id, Valid: true},
		SellerID:    pgtype.UUID{Bytes: sellerID, Valid: true},
		Title:       title,
		Description: pgtype.Text{Valid: false},
		Price:       price,
		Status:      model.AdOnModeration.String(),
		CreatedAt: pgtype.Timestamptz{
			Time:  createdAt,
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamptz{Valid: false},
	}

	expected := model.RestoreAd(
		id,
		sellerID,
		title,
		nil,
		price,
		model.AdOnModeration,
		nil,
		createdAt,
		nil,
	)

	ad := mapper.MapSQLCToAd(raw)

	require.NotNil(t, ad)
	assert.True(t, reflect.DeepEqual(expected, ad))
}

func TestMapAdToSQLCCreate(t *testing.T) {
	testDesc := gofakeit.Bio()
	testPrice := int64(gofakeit.Price(1000, 1000000))

	ad, err := model.NewAd(
		uuid.New(),
		gofakeit.ProductName(),
		&testDesc,
		testPrice,
		nil,
	)
	require.NoError(t, err)

	expected := sqlc.CreateAdParams{
		ID: pgtype.UUID{
			Bytes: ad.ID(),
			Valid: true,
		},
		SellerID: pgtype.UUID{
			Bytes: ad.SellerID(),
			Valid: true,
		},
		Title: ad.Title(),
		Description: pgtype.Text{
			String: testDesc,
			Valid:  true,
		},
		Price:  ad.Price(),
		Status: string(ad.Status()),
		CreatedAt: pgtype.Timestamptz{
			Time:  ad.CreatedAt(),
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamptz{},
	}

	mapped := mapper.MapAdToSQLCCreate(ad)

	assert.True(t, reflect.DeepEqual(expected, mapped))
}

func TestMapAdToSQLCUpdate(t *testing.T) {
	testDesc := gofakeit.Bio()
	testPrice := int64(gofakeit.Price(1000, 1000000))

	ad, err := model.NewAd(
		uuid.New(),
		gofakeit.ProductName(),
		&testDesc,
		testPrice,
		nil,
	)
	require.NoError(t, err)

	expected := sqlc.UpdateAdParams{
		ID: pgtype.UUID{
			Bytes: ad.ID(),
			Valid: true,
		},
		SellerID: pgtype.UUID{
			Bytes: ad.SellerID(),
			Valid: true,
		},
		Title: ad.Title(),
		Description: pgtype.Text{
			String: testDesc,
			Valid:  true,
		},
		Price:  ad.Price(),
		Status: string(ad.Status()),
		CreatedAt: pgtype.Timestamptz{
			Time:  ad.CreatedAt(),
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamptz{},
	}

	mapped := mapper.MapAdToSQLCUpdate(ad)

	assert.True(t, reflect.DeepEqual(expected, mapped))
}

func TestMapToSQLCList(t *testing.T) {
	testLimit := gofakeit.Number(1, 100)
	testOffset := gofakeit.Number(0, 100)

	expected := sqlc.ListAdsParams{
		Limit:  int32(testLimit),
		Offset: int32(testOffset),
	}

	mapped := mapper.MapToSQLCList(testLimit, testOffset)

	assert.True(t, reflect.DeepEqual(expected, mapped))
}

func TestMapToSQLCSellerList(t *testing.T) {
	sellerID := uuid.New()
	testLimit := gofakeit.Number(1, 100)
	testOffset := gofakeit.Number(0, 100)

	expected := sqlc.ListSellerAdsParams{
		SellerID: pgtype.UUID{
			Bytes: sellerID,
			Valid: true,
		},
		Limit:  int32(testLimit),
		Offset: int32(testOffset),
	}

	mapped := mapper.MapToSQLCSellerList(sellerID, testLimit, testOffset)

	assert.True(t, reflect.DeepEqual(expected, mapped))
}

func TestMapSQLCToAdsList(t *testing.T) {
	rawAds := []sqlc.Ad{
		{
			ID:          pgtype.UUID{Bytes: uuid.New(), Valid: true},
			SellerID:    pgtype.UUID{Bytes: uuid.New(), Valid: true},
			Title:       gofakeit.ProductName(),
			Description: pgtype.Text{},
			Price:       int64(gofakeit.Price(1000, 1000000)),
			Status:      "deleted",
			CreatedAt:   pgtype.Timestamptz{Time: gofakeit.Date(), Valid: true},
			UpdatedAt:   pgtype.Timestamptz{},
		},
		{
			ID:          pgtype.UUID{Bytes: uuid.New(), Valid: true},
			SellerID:    pgtype.UUID{Bytes: uuid.New(), Valid: true},
			Title:       gofakeit.ProductName(),
			Description: pgtype.Text{},
			Price:       int64(gofakeit.Price(1000, 1000000)),
			Status:      "rejected",
			CreatedAt:   pgtype.Timestamptz{Time: gofakeit.Date(), Valid: true},
			UpdatedAt:   pgtype.Timestamptz{},
		},
	}

	expected := make([]*model.Ad, 0, len(rawAds))
	for _, raw := range rawAds {
		expected = append(expected, model.RestoreAd(
			raw.ID.Bytes,
			raw.SellerID.Bytes,
			raw.Title,
			nil,
			raw.Price,
			model.AdStatus(raw.Status),
			nil,
			raw.CreatedAt.Time,
			nil,
		))
	}

	mapped := mapper.MapSQLCToAdsList(rawAds)

	require.Len(t, mapped, len(rawAds))
	assert.True(t, reflect.DeepEqual(expected, mapped))
}
