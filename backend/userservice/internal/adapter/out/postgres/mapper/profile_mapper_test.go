package mapper_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/maket12/ads-service/userservice/internal/adapter/out/postgres/mapper"
	"github.com/maket12/ads-service/userservice/internal/adapter/out/postgres/sqlc"
	"github.com/maket12/ads-service/userservice/internal/domain/model"
	"github.com/maket12/ads-service/userservice/pkg/utils"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMapProfileToSQLCCreate(t *testing.T) {
	accountID := uuid.New()
	profile, _ := model.NewProfile(accountID)

	expected := sqlc.CreateProfileParams{
		AccountID: pgtype.UUID{
			Bytes: accountID,
			Valid: true,
		},
		FirstName: pgtype.Text{},
		LastName:  pgtype.Text{},
		Phone:     pgtype.Text{},
		AvatarUrl: pgtype.Text{},
		Bio:       pgtype.Text{},
		UpdatedAt: pgtype.Timestamptz{
			Time:  profile.UpdatedAt(),
			Valid: true,
		},
	}

	mapped := mapper.MapProfileToSQLCCreate(profile)

	if !reflect.DeepEqual(expected, mapped) {
		t.Fatalf("MapProfileToSQLCCreate mismatch:\nexpected: %+v\ngot:      %+v", expected, mapped)
	}
}

func TestMapSQLCToProfile(t *testing.T) {
	accountID := uuid.New()
	firstName := gofakeit.FirstName()
	lastName := gofakeit.LastName()
	updatedAt := time.Now()

	raw := sqlc.Profile{
		AccountID: pgtype.UUID{
			Bytes: accountID,
			Valid: true,
		},
		FirstName: pgtype.Text{
			String: firstName,
			Valid:  true,
		},
		LastName: pgtype.Text{
			String: lastName,
			Valid:  true,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:  updatedAt,
			Valid: true,
		},
	}

	expected := model.RestoreProfile(
		accountID,
		&firstName,
		&lastName,
		nil,
		nil,
		nil,
		updatedAt,
	)

	mapped := mapper.MapSQLCToProfile(raw)

	if !reflect.DeepEqual(expected, mapped) {
		t.Fatalf("MapSQLCToProfile mismatch:\nexpected: %+v\ngot:      %+v", expected, mapped)
	}
}

func TestMapProfileToSQLCUpdate(t *testing.T) {
	accountID := uuid.New()
	profile, _ := model.NewProfile(accountID)
	err := profile.Update(
		utils.VPtr("Vladimir"), nil, nil, nil, nil,
	)
	assert.NoError(t, err)

	expected := sqlc.UpdateProfileParams{
		AccountID: pgtype.UUID{
			Bytes: accountID,
			Valid: true,
		},
		FirstName: pgtype.Text{
			String: "Vladimir",
			Valid:  true,
		},
		LastName:  pgtype.Text{},
		Phone:     pgtype.Text{},
		AvatarUrl: pgtype.Text{},
		Bio:       pgtype.Text{},
		UpdatedAt: pgtype.Timestamptz{
			Time:  profile.UpdatedAt(),
			Valid: true,
		},
	}

	mapped := mapper.MapProfileToSQLCUpdate(profile)

	if !reflect.DeepEqual(expected, mapped) {
		t.Fatalf("MapProfileToSQLCUpdate mismatch:\nexpected: %+v\ngot:      %+v", expected, mapped)
	}
}

func TestMapToSQLCList(t *testing.T) {
	var testLimit, testOffset = 10, 10

	expected := sqlc.ListProfilesParams{
		Limit:  int32(testLimit),
		Offset: int32(testOffset),
	}

	mapped := mapper.MapToSQLCList(testLimit, testOffset)

	if !reflect.DeepEqual(expected, mapped) {
		t.Fatalf("MapToSQLCList mismatch:\nexpected: %+v\ngot:      %+v", expected, mapped)
	}
}

func TestMapSQLCToProfilesList(t *testing.T) {
	firstAccountID := uuid.New()
	secondAccountID := uuid.New()
	firstName1 := "Vladimir"
	firstName2 := "ShiShi"

	rawProfiles := []sqlc.Profile{
		{
			AccountID: pgtype.UUID{
				Bytes: firstAccountID,
				Valid: true,
			},
			FirstName: pgtype.Text{
				String: firstName1,
				Valid:  true,
			},
		},
		{
			AccountID: pgtype.UUID{
				Bytes: secondAccountID,
				Valid: true,
			},
			FirstName: pgtype.Text{
				String: firstName2,
				Valid:  true,
			},
		},
	}

	expected := []*model.Profile{
		model.RestoreProfile(firstAccountID, &firstName1, nil, nil, nil, nil, time.Time{}),
		model.RestoreProfile(secondAccountID, &firstName2, nil, nil, nil, nil, time.Time{}),
	}

	mapped := mapper.MapSQLCToProfilesList(rawProfiles)

	if !reflect.DeepEqual(expected, mapped) {
		t.Fatalf("MapSQLCToProfilesList mismatch:\nexpected: %+v\ngot:      %+v", expected, mapped)
	}
}
