package mapper_test

import (
	"net/netip"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/maket12/ads-service/authservice/internal/adapter/out/postgres/mapper"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"

	"github.com/maket12/ads-service/authservice/internal/adapter/out/postgres/sqlc"
	"github.com/maket12/ads-service/authservice/internal/domain/model"
)

func TestMapRefreshSessionToSQLCCreate(t *testing.T) {
	id := uuid.New()
	accountID := uuid.New()
	rotatedFrom := uuid.New()

	createdAt := time.Now()
	expiresAt := createdAt.Add(time.Minute)
	revokedAt := expiresAt.Add(time.Minute)

	revokeReason := model.ReasonTokenRotation
	ip, _ := netip.ParseAddr(gofakeit.IPv4Address())
	ipStr := ip.String()
	userAgent := gofakeit.UserAgent()

	session := model.RestoreRefreshSession(
		id,
		accountID,
		"refresh-token-hash",
		createdAt,
		expiresAt,
		&revokedAt,
		&revokeReason,
		&rotatedFrom,
		&ipStr,
		&userAgent,
	)

	expected := sqlc.CreateSessionParams{
		ID: pgtype.UUID{
			Bytes: id,
			Valid: true,
		},
		AccountID: pgtype.UUID{
			Bytes: accountID,
			Valid: true,
		},
		RefreshTokenHash: "refresh-token-hash",
		CreatedAt: pgtype.Timestamptz{
			Time:  createdAt,
			Valid: true,
		},
		ExpiresAt: pgtype.Timestamptz{
			Time:  expiresAt,
			Valid: true,
		},
		RevokedAt: pgtype.Timestamptz{
			Time:  revokedAt,
			Valid: true,
		},
		RevokeReason: pgtype.Text{
			String: revokeReason.String(),
			Valid:  true,
		},
		RotatedFrom: pgtype.UUID{
			Bytes: rotatedFrom,
			Valid: true,
		},
		Ip: &ip,
		UserAgent: pgtype.Text{
			String: userAgent,
			Valid:  true,
		},
	}

	actual := mapper.MapRefreshSessionToSQLCCreate(session)

	require.Equal(t, expected, actual)
}

func TestMapRefreshSessionToSQLCCreate_NilOptionalFields(t *testing.T) {
	id := uuid.New()
	accountID := uuid.New()

	createdAt := time.Now()
	expiresAt := createdAt.Add(time.Minute)

	session := model.RestoreRefreshSession(
		id,
		accountID,
		"refresh-token-hash",
		createdAt,
		expiresAt,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	expected := sqlc.CreateSessionParams{
		ID: pgtype.UUID{
			Bytes: id,
			Valid: true,
		},
		AccountID: pgtype.UUID{
			Bytes: accountID,
			Valid: true,
		},
		RefreshTokenHash: "refresh-token-hash",
		CreatedAt: pgtype.Timestamptz{
			Time:  createdAt,
			Valid: true,
		},
		ExpiresAt: pgtype.Timestamptz{
			Time:  expiresAt,
			Valid: true,
		},
		RevokedAt:    pgtype.Timestamptz{},
		RevokeReason: pgtype.Text{},
		RotatedFrom:  pgtype.UUID{},
		Ip:           nil,
		UserAgent:    pgtype.Text{},
	}

	actual := mapper.MapRefreshSessionToSQLCCreate(session)

	require.Equal(t, expected, actual)
}

func TestMapSQLCToRefreshSession(t *testing.T) {
	id := uuid.New()
	accountID := uuid.New()
	rotatedFrom := uuid.New()

	revokeReason := model.ReasonLogout
	ip, _ := netip.ParseAddr(gofakeit.IPv4Address())
	ipStr := ip.String()
	userAgent := gofakeit.UserAgent()

	createdAt := time.Now()
	expiresAt := createdAt.Add(time.Minute)
	revokedAt := expiresAt.Add(time.Minute)

	raw := sqlc.RefreshSession{
		ID: pgtype.UUID{
			Bytes: id,
			Valid: true,
		},
		AccountID: pgtype.UUID{
			Bytes: accountID,
			Valid: true,
		},
		RefreshTokenHash: "refresh-token-hash",
		CreatedAt: pgtype.Timestamptz{
			Time:  createdAt,
			Valid: true,
		},
		ExpiresAt: pgtype.Timestamptz{
			Time:  expiresAt,
			Valid: true,
		},
		RevokedAt: pgtype.Timestamptz{
			Time:  revokedAt,
			Valid: true,
		},
		RevokeReason: pgtype.Text{
			String: revokeReason.String(),
			Valid:  true,
		},
		RotatedFrom: pgtype.UUID{
			Bytes: rotatedFrom,
			Valid: true,
		},
		Ip: &ip,
		UserAgent: pgtype.Text{
			String: userAgent,
			Valid:  true,
		},
	}

	expected := model.RestoreRefreshSession(
		id,
		accountID,
		"refresh-token-hash",
		createdAt,
		expiresAt,
		&revokedAt,
		&revokeReason,
		&rotatedFrom,
		&ipStr,
		&userAgent,
	)

	actual := mapper.MapSQLCToRefreshSession(raw)

	require.Equal(t, expected, actual)
}

func TestMapSQLCToRefreshSession_NilOptionalFields(t *testing.T) {
	id := uuid.New()
	accountID := uuid.New()

	createdAt := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	expiresAt := time.Date(2024, 1, 8, 10, 0, 0, 0, time.UTC)

	raw := sqlc.RefreshSession{
		ID: pgtype.UUID{
			Bytes: id,
			Valid: true,
		},
		AccountID: pgtype.UUID{
			Bytes: accountID,
			Valid: true,
		},
		RefreshTokenHash: "refresh-token-hash",
		CreatedAt: pgtype.Timestamptz{
			Time:  createdAt,
			Valid: true,
		},
		ExpiresAt: pgtype.Timestamptz{
			Time:  expiresAt,
			Valid: true,
		},
		RevokedAt:    pgtype.Timestamptz{},
		RevokeReason: pgtype.Text{},
		RotatedFrom:  pgtype.UUID{},
		Ip:           nil,
		UserAgent:    pgtype.Text{},
	}

	expected := model.RestoreRefreshSession(
		id,
		accountID,
		"refresh-token-hash",
		createdAt,
		expiresAt,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	actual := mapper.MapSQLCToRefreshSession(raw)

	require.Equal(t, expected, actual)
}

func TestMapRefreshSessionToSQLCRevoke(t *testing.T) {
	id := uuid.New()
	accountID := uuid.New()
	rotatedFrom := uuid.New()

	createdAt := time.Now()
	expiresAt := createdAt.Add(time.Second)
	revokedAt := expiresAt.Add(time.Second)

	refreshTokenHash := "refresh-token-hash"
	revokeReason := model.ReasonLogout
	ipStr := gofakeit.IPv4Address()
	ip, _ := netip.ParseAddr(ipStr)
	userAgent := gofakeit.UserAgent()

	session := model.RestoreRefreshSession(
		id,
		accountID,
		refreshTokenHash,
		createdAt,
		expiresAt,
		&revokedAt,
		&revokeReason,
		&rotatedFrom,
		&ipStr,
		&userAgent,
	)

	expected := sqlc.UpdateSessionParams{
		ID:               pgtype.UUID{Bytes: id, Valid: true},
		AccountID:        pgtype.UUID{Bytes: accountID, Valid: true},
		RefreshTokenHash: refreshTokenHash,
		CreatedAt:        pgtype.Timestamptz{Time: createdAt, Valid: true},
		ExpiresAt:        pgtype.Timestamptz{Time: expiresAt, Valid: true},
		RevokedAt:        pgtype.Timestamptz{Time: revokedAt, Valid: true},
		RevokeReason:     pgtype.Text{String: revokeReason.String(), Valid: true},
		RotatedFrom:      pgtype.UUID{Bytes: rotatedFrom, Valid: true},
		Ip:               &ip,
		UserAgent:        pgtype.Text{String: userAgent, Valid: true},
	}

	actual := mapper.MapRefreshSessionToSQLCUpdate(session)

	require.Equal(t, expected, actual)
}

func TestMapRefreshSessionToSQLCRevoke_NilOptionalFields(t *testing.T) {
	id := uuid.New()
	accountID := uuid.New()

	createdAt := time.Now()
	expiresAt := createdAt.Add(time.Minute)

	refreshTokenHash := "refresh-token-hash"

	session := model.RestoreRefreshSession(
		id,
		accountID,
		refreshTokenHash,
		createdAt,
		expiresAt,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	expected := sqlc.UpdateSessionParams{
		ID:               pgtype.UUID{Bytes: id, Valid: true},
		AccountID:        pgtype.UUID{Bytes: accountID, Valid: true},
		RefreshTokenHash: refreshTokenHash,
		CreatedAt:        pgtype.Timestamptz{Time: createdAt, Valid: true},
		ExpiresAt:        pgtype.Timestamptz{Time: expiresAt, Valid: true},
		RevokedAt:        pgtype.Timestamptz{},
		RevokeReason:     pgtype.Text{},
		RotatedFrom:      pgtype.UUID{},
		Ip:               nil,
		UserAgent:        pgtype.Text{},
	}

	actual := mapper.MapRefreshSessionToSQLCUpdate(session)

	require.Equal(t, expected, actual)
}

func TestMapToSQLCRevokeAllForAccount(t *testing.T) {
	accountID := uuid.New()
	reason := "security incident"

	before := time.Now()
	actual := mapper.MapToSQLCRevokeAllForAccount(accountID, &reason)
	after := time.Now()

	require.Equal(t, pgtype.UUID{Bytes: accountID, Valid: true}, actual.AccountID)
	require.Equal(t, pgtype.Text{String: reason, Valid: true}, actual.RevokeReason)
	require.True(t, actual.RevokedAt.Valid)
	require.WithinRange(t, actual.RevokedAt.Time, before, after)
}

func TestMapToSQLCRevokeAllForAccount_NilReason(t *testing.T) {
	accountID := uuid.New()

	before := time.Now()
	actual := mapper.MapToSQLCRevokeAllForAccount(accountID, nil)
	after := time.Now()

	require.Equal(t, pgtype.UUID{Bytes: accountID, Valid: true}, actual.AccountID)
	require.Equal(t, pgtype.Text{}, actual.RevokeReason)
	require.True(t, actual.RevokedAt.Valid)
	require.WithinRange(t, actual.RevokedAt.Time, before, after)
}

func TestMapToSQLCRevokeDescendants(t *testing.T) {
	sessionID := uuid.New()
	reason := "rotation"

	before := time.Now()
	actual := mapper.MapToSQLCRevokeDescendants(sessionID, &reason)
	after := time.Now()

	require.Equal(t, pgtype.UUID{Bytes: sessionID, Valid: true}, actual.ID)
	require.Equal(t, pgtype.Text{String: reason, Valid: true}, actual.RevokeReason)
	require.True(t, actual.RevokedAt.Valid)
	require.WithinRange(t, actual.RevokedAt.Time, before, after)
}

func TestMapToSQLCRevokeDescendants_NilReason(t *testing.T) {
	sessionID := uuid.New()

	before := time.Now()
	actual := mapper.MapToSQLCRevokeDescendants(sessionID, nil)
	after := time.Now()

	require.Equal(t, pgtype.UUID{Bytes: sessionID, Valid: true}, actual.ID)
	require.Equal(t, pgtype.Text{}, actual.RevokeReason)
	require.True(t, actual.RevokedAt.Valid)
	require.WithinRange(t, actual.RevokedAt.Time, before, after)
}

func TestMapSQLCToListRefreshSession(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()
	accountID := uuid.New()

	createdAt := time.Now()
	expiresAt := createdAt.Add(time.Minute)

	rawList := []sqlc.RefreshSession{
		{
			ID:               pgtype.UUID{Bytes: id1, Valid: true},
			AccountID:        pgtype.UUID{Bytes: accountID, Valid: true},
			RefreshTokenHash: "hash-1",
			CreatedAt:        pgtype.Timestamptz{Time: createdAt, Valid: true},
			ExpiresAt:        pgtype.Timestamptz{Time: expiresAt, Valid: true},
		},
		{
			ID:               pgtype.UUID{Bytes: id2, Valid: true},
			AccountID:        pgtype.UUID{Bytes: accountID, Valid: true},
			RefreshTokenHash: "hash-2",
			CreatedAt:        pgtype.Timestamptz{Time: createdAt, Valid: true},
			ExpiresAt:        pgtype.Timestamptz{Time: expiresAt, Valid: true},
		},
	}

	expected := []*model.RefreshSession{
		model.RestoreRefreshSession(
			id1,
			accountID,
			"hash-1",
			createdAt,
			expiresAt,
			nil,
			nil,
			nil,
			nil,
			nil,
		),
		model.RestoreRefreshSession(
			id2,
			accountID,
			"hash-2",
			createdAt,
			expiresAt,
			nil,
			nil,
			nil,
			nil,
			nil,
		),
	}

	actual := mapper.MapSQLCToListRefreshSession(rawList)

	require.Equal(t, expected, actual)
}

func TestMapSQLCToListRefreshSession_Empty(t *testing.T) {
	actual := mapper.MapSQLCToListRefreshSession([]sqlc.RefreshSession{})

	require.Empty(t, actual)
	require.NotNil(t, actual)
}
