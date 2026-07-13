package mapper_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"

	"github.com/maket12/ads-service/authservice/internal/adapter/out/postgres/sqlc"
	"github.com/maket12/ads-service/authservice/internal/domain/model"
)

func TestMapAccountRoleToSQLCCreate(t *testing.T) {
	accountID := uuid.New()

	accRole := model.RestoreAccountRole(
		accountID,
		model.RoleAdmin,
	)

	expected := sqlc.CreateAccountRoleParams{
		AccountID: pgtype.UUID{
			Bytes: accountID,
			Valid: true,
		},
		Role: model.RoleAdmin.String(),
	}

	actual := MapAccountRoleToSQLCCreate(accRole)

	require.Equal(t, expected, actual)
}

func TestMapAccountRoleToSQLCUpdate(t *testing.T) {
	accountID := uuid.New()

	accRole := model.RestoreAccountRole(
		accountID,
		model.RoleUser,
	)

	expected := sqlc.UpdateAccountRoleParams{
		AccountID: pgtype.UUID{
			Bytes: accountID,
			Valid: true,
		},
		Role: model.RoleUser.String(),
	}

	actual := MapAccountRoleToSQLCUpdate(accRole)

	require.Equal(t, expected, actual)
}

func TestMapSQLCToAccountRole(t *testing.T) {
	accountID := uuid.New()

	raw := sqlc.AccountRole{
		AccountID: pgtype.UUID{
			Bytes: accountID,
			Valid: true,
		},
		Role: model.RoleAdmin.String(),
	}

	expected := model.RestoreAccountRole(
		accountID,
		model.RoleAdmin,
	)

	actual := MapSQLCToAccountRole(raw)

	require.Equal(t, expected, actual)
}
