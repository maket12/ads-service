package mapper_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/maket12/ads-service/adservice/internal/adapter/out/postgres/mapper"
	"github.com/maket12/ads-service/adservice/internal/adapter/out/postgres/sqlc"
	"github.com/maket12/ads-service/adservice/internal/domain/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapSQLCToAd(t *testing.T) {
	t.Parallel()

	raw := sqlc.Ad{
		ID:       uuid.New(),
		SellerID: uuid.New(),
		Title:    "Sell a penthouse",
		Description: sql.NullString{
			String: "was built in 1983",
			Valid:  true,
		},
		Price:     800000,
		Status:    "published",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	ad := mapper.MapSQLCToAd(raw)

	require.NotNil(t, ad)
	require.NotNil(t, ad.Description())

	assert.Equal(t, raw.ID, ad.ID())
	assert.Equal(t, raw.SellerID, ad.SellerID())
	assert.Equal(t, raw.Title, ad.Title())
	assert.Equal(t, raw.Description.String, *ad.Description())
	assert.Equal(t, string(raw.Status), string(ad.Status()))
	assert.Equal(t, raw.Price, ad.Price())
	assert.Equal(t, raw.CreatedAt, ad.CreatedAt())
	assert.Equal(t, raw.UpdatedAt, ad.UpdatedAt())
}

func TestMapAdToSQLCCreate(t *testing.T) {
	t.Parallel()

	testDesc := "was built in 1983"

	ad, _ := model.NewAd(
		uuid.New(),
		"Sell penthouse",
		&testDesc,
		780000,
		nil,
	)

	mapped := mapper.MapAdToSQLCCreate(ad)

	require.NotNil(t, mapped)
	require.True(t, mapped.Description.Valid)

	assert.Equal(t, ad.ID(), mapped.ID)
	assert.Equal(t, ad.SellerID(), mapped.SellerID)
	assert.Equal(t, ad.Title(), mapped.Title)
	assert.Equal(t, testDesc, mapped.Description.String)
	assert.Equal(t, ad.Price(), mapped.Price)
	assert.Equal(t, string(ad.Status()), string(mapped.Status))
	assert.Equal(t, ad.CreatedAt(), mapped.CreatedAt)
	assert.Equal(t, ad.UpdatedAt(), mapped.UpdatedAt)
}

func TestMapAdToSQLCUpdate(t *testing.T) {
	t.Parallel()

	testDesc := "was built in 1983"

	ad, _ := model.NewAd(
		uuid.New(),
		"Sell penthouse",
		&testDesc,
		780000,
		nil,
	)

	mapped := mapper.MapAdToSQLCUpdate(ad)

	require.NotNil(t, mapped)
	require.True(t, mapped.Description.Valid)

	assert.Equal(t, ad.ID(), mapped.ID)
	assert.Equal(t, ad.Title(), mapped.Title)
	assert.Equal(t, testDesc, mapped.Description.String)
	assert.Equal(t, ad.Price(), mapped.Price)
	assert.Equal(t, ad.UpdatedAt(), mapped.UpdatedAt)
}

func TestMapAdToSQLCUpdateStatus(t *testing.T) {
	t.Parallel()

	ad, _ := model.NewAd(
		uuid.New(),
		"Sell penthouse",
		nil,
		780000,
		nil,
	)
	_ = ad.Reject()

	mapped := mapper.MapAdToSQLCUpdateStatus(ad)

	require.NotNil(t, mapped)

	assert.Equal(t, ad.ID(), mapped.ID)
	assert.Equal(t, string(ad.Status()), string(mapped.Status))
}

func TestMapToSQLCList(t *testing.T) {
	t.Parallel()

	testLimit := 10
	testOffset := 10

	mapped := mapper.MapToSQLCList(testLimit, testOffset)

	require.NotNil(t, mapped)

	assert.Equal(t, testLimit, int(mapped.Limit))
	assert.Equal(t, testOffset, int(mapped.Offset))
}

func TestMapSQLCToAdsList(t *testing.T) {
	t.Parallel()

	rawAds := []sqlc.Ad{
		{
			ID:          uuid.New(),
			SellerID:    uuid.New(),
			Title:       "A new product",
			Description: sql.NullString{},
			Price:       10000,
			Status:      sqlc.AdStatusDeleted,
			CreatedAt:   time.Time{},
			UpdatedAt:   time.Time{},
		},
		{
			ID:          uuid.New(),
			SellerID:    uuid.New(),
			Title:       "A new product",
			Description: sql.NullString{},
			Price:       10000,
			Status:      sqlc.AdStatusRejected,
			CreatedAt:   time.Time{},
			UpdatedAt:   time.Time{},
		},
	}

	mapped := mapper.MapSQLCToAdsList(rawAds)

	require.NotNil(t, mapped)
	require.NotEmpty(t, mapped)
	require.Len(t, mapped, len(rawAds))

	for i := 0; i < len(rawAds); i++ {
		require.NotNil(t, mapped[i])
		assert.Equal(t, rawAds[i].ID, mapped[i].ID())
		assert.Equal(t, rawAds[i].Title, mapped[i].Title())
		// ...
		assert.Equal(t, rawAds[i].UpdatedAt, mapped[i].UpdatedAt())
	}
}
