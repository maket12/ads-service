package mapper_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/maket12/ads-service/authservice/internal/adapter/out/postgres/mapper"
	"github.com/stretchr/testify/require"

	"github.com/maket12/ads-service/authservice/internal/adapter/out/postgres/sqlc"
	"github.com/maket12/ads-service/authservice/internal/domain/model"
)

func TestMapAccountToSQLCCreate(t *testing.T) {
	id := uuid.New()
	createdAt := time.Now()
	updatedAt := createdAt.Add(time.Hour)
	lastLoginAt := updatedAt.Add(time.Hour)

	acc := model.RestoreAccount(
		id,
		"shishi12377@weixin.cn",
		"hashed-password",
		model.AccountActive,
		true,
		createdAt,
		updatedAt,
		&lastLoginAt,
	)

	expected := sqlc.CreateAccountParams{
		ID: pgtype.UUID{
			Bytes: id,
			Valid: true,
		},
		Email:         "shishi12377@weixin.cn",
		PasswordHash:  "hashed-password",
		Status:        model.AccountActive.String(),
		EmailVerified: true,
		CreatedAt: pgtype.Timestamptz{
			Time:  createdAt,
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:  updatedAt,
			Valid: true,
		},
		LastLoginAt: pgtype.Timestamptz{
			Time:  lastLoginAt,
			Valid: true,
		},
	}

	actual := mapper.MapAccountToSQLCCreate(acc)

	require.Equal(t, expected, actual)
}

func TestMapAccountToSQLCCreate_NilLastLogin(t *testing.T) {
	id := uuid.New()
	createdAt := time.Now()
	updatedAt := createdAt.Add(time.Minute)

	acc := model.RestoreAccount(
		id,
		"noone@example.com",
		"hash",
		model.AccountBlocked,
		false,
		createdAt,
		updatedAt,
		nil,
	)

	expected := sqlc.CreateAccountParams{
		ID: pgtype.UUID{
			Bytes: id,
			Valid: true,
		},
		Email:         "noone@example.com",
		PasswordHash:  "hash",
		Status:        model.AccountBlocked.String(),
		EmailVerified: false,
		CreatedAt: pgtype.Timestamptz{
			Time:  createdAt,
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:  updatedAt,
			Valid: true,
		},
		LastLoginAt: pgtype.Timestamptz{},
	}

	actual := mapper.MapAccountToSQLCCreate(acc)

	require.Equal(t, expected, actual)
}

func TestMapAccountToSQLCUpdate(t *testing.T) {
	id := uuid.New()
	email := "shishi12377@weixin.cn"
	pass := "hashed-pass"
	createdAt := time.Now()
	updatedAt := time.Now().Add(time.Minute)
	lastLoginAt := updatedAt.Add(time.Minute)

	acc := model.RestoreAccount(
		id,
		email,
		pass,
		model.AccountActive,
		false,
		createdAt,
		updatedAt,
		&lastLoginAt,
	)

	expected := sqlc.UpdateAccountParams{
		ID: pgtype.UUID{
			Bytes: id,
			Valid: true,
		},
		Email:         email,
		PasswordHash:  pass,
		Status:        model.AccountActive.String(),
		EmailVerified: false,
		CreatedAt: pgtype.Timestamptz{
			Time:  createdAt,
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:  updatedAt,
			Valid: true,
		},
		LastLoginAt: pgtype.Timestamptz{
			Time:  lastLoginAt,
			Valid: true,
		},
	}

	actual := mapper.MapAccountToSQLCUpdate(acc)

	require.Equal(t, expected, actual)
}

func TestMapAccountToSQLCUpdate_NilLastLogin(t *testing.T) {
	id := uuid.New()
	email := "shishi12377@weixin.cn"
	pass := "hashed-pass"
	createdAt := time.Now()
	updatedAt := time.Now().Add(time.Minute)

	acc := model.RestoreAccount(
		id,
		email,
		pass,
		model.AccountActive,
		false,
		createdAt,
		updatedAt,
		nil,
	)

	expected := sqlc.UpdateAccountParams{
		ID: pgtype.UUID{
			Bytes: id,
			Valid: true,
		},
		Email:         email,
		PasswordHash:  pass,
		Status:        model.AccountActive.String(),
		EmailVerified: false,
		CreatedAt: pgtype.Timestamptz{
			Time:  createdAt,
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:  updatedAt,
			Valid: true,
		},
		LastLoginAt: pgtype.Timestamptz{},
	}

	actual := mapper.MapAccountToSQLCUpdate(acc)

	require.Equal(t, expected, actual)
}

func TestMapSQLCToAccount(t *testing.T) {
	id := uuid.New()
	createdAt := time.Now()
	updatedAt := createdAt.Add(time.Hour)
	lastLoginAt := updatedAt.Add(time.Hour)

	raw := sqlc.Account{
		ID: pgtype.UUID{
			Bytes: id,
			Valid: true,
		},
		Email:         "user@example.com",
		PasswordHash:  "hashed-password",
		Status:        model.AccountActive.String(),
		EmailVerified: true,
		CreatedAt: pgtype.Timestamptz{
			Time:  createdAt,
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:  updatedAt,
			Valid: true,
		},
		LastLoginAt: pgtype.Timestamptz{
			Time:  lastLoginAt,
			Valid: true,
		},
	}

	expected := model.RestoreAccount(
		id,
		"user@example.com",
		"hashed-password",
		model.AccountActive,
		true,
		createdAt,
		updatedAt,
		&lastLoginAt,
	)

	actual := mapper.MapSQLCToAccount(raw)

	require.Equal(t, expected, actual)
}

func TestMapSQLCToAccount_NilLastLogin(t *testing.T) {
	id := uuid.New()
	createdAt := time.Now()
	updatedAt := createdAt.Add(time.Hour)

	raw := sqlc.Account{
		ID: pgtype.UUID{
			Bytes: id,
			Valid: true,
		},
		Email:         "noone@example.com",
		PasswordHash:  "hash",
		Status:        model.AccountDeleted.String(),
		EmailVerified: false,
		CreatedAt: pgtype.Timestamptz{
			Time:  createdAt,
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:  updatedAt,
			Valid: true,
		},
		LastLoginAt: pgtype.Timestamptz{},
	}

	expected := model.RestoreAccount(
		id,
		"noone@example.com",
		"hash",
		model.AccountDeleted,
		false,
		createdAt,
		updatedAt,
		nil,
	)

	actual := mapper.MapSQLCToAccount(raw)

	require.Equal(t, expected, actual)
}
