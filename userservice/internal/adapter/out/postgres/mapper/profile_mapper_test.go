package mapper_test

import (
	"database/sql"
	"github.com/maket12/ads-service/userservice/internal/adapter/out/postgres/mapper"
	"github.com/maket12/ads-service/userservice/internal/adapter/out/postgres/sqlc"
	"github.com/maket12/ads-service/userservice/internal/domain/model"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMapProfileToSQLCCreate(t *testing.T) {
	t.Parallel()

	profile, _ := model.NewProfile(uuid.New())
	mapped := mapper.MapProfileToSQLCCreate(profile)
	assert.Equal(t, profile.AccountID(), mapped.AccountID)
}

func TestMapSQLCToProfile(t *testing.T) {
	t.Parallel()

	raw := sqlc.Profile{
		AccountID: uuid.New(),
		FirstName: sql.NullString{
			String: "Vladimir",
			Valid:  true,
		},
		LastName: sql.NullString{
			String: "Ziabkin",
			Valid:  true,
		},
		UpdatedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
	}
	profile := mapper.MapSQLCToProfile(raw)

	assert.Equal(t, raw.AccountID, profile.AccountID())
	assert.Equal(t, raw.FirstName.String, *profile.FirstName())
	assert.Equal(t, raw.LastName.String, *profile.LastName())
	assert.Equal(t, raw.UpdatedAt.Time, profile.UpdatedAt())
}

func TestMapProfileToSQLCUpdate(t *testing.T) {
	t.Parallel()

	var testFirstName = "Vladimir"

	profile, _ := model.NewProfile(uuid.New())
	_ = profile.Update(
		&testFirstName, nil, nil, nil, nil,
	)
	mapped := mapper.MapProfileToSQLCUpdate(profile)

	assert.Equal(t, profile.AccountID(), mapped.AccountID)
	assert.Equal(t, *profile.FirstName(), mapped.FirstName.String)
}

func TestMapToSQLCList(t *testing.T) {
	t.Parallel()

	var testLimit, testOffset = 10, 10
	mapped := mapper.MapToSQLCList(testLimit, testOffset)

	assert.Equal(t, testLimit, int(mapped.Limit))
	assert.Equal(t, testOffset, int(mapped.Offset))
}

func TestMapMapSQLCToProfilesList(t *testing.T) {
	t.Parallel()

	rawProfiles := []sqlc.Profile{
		{
			AccountID: uuid.New(),
			FirstName: sql.NullString{
				String: "Vladimir",
				Valid:  true,
			},
		},
		{
			AccountID: uuid.New(),
			FirstName: sql.NullString{
				String: "ShiShi",
				Valid:  true,
			},
		},
	}
	mapped := mapper.MapSQLCToProfilesList(rawProfiles)

	assert.Equal(t, rawProfiles[0].AccountID, mapped[0].AccountID())
	assert.Equal(t, rawProfiles[0].FirstName.String, *mapped[0].FirstName())
	assert.Equal(t, rawProfiles[1].AccountID, mapped[1].AccountID())
	assert.Equal(t, rawProfiles[1].FirstName.String, *mapped[1].FirstName())
}
