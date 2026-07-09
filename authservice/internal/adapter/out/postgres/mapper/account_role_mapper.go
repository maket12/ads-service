package mapper

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/maket12/ads-service/authservice/internal/adapter/out/postgres/sqlc"
	"github.com/maket12/ads-service/authservice/internal/domain/model"
)

func MapAccountRoleToSQLCCreate(accRole *model.AccountRole) sqlc.CreateAccountRoleParams {
	return sqlc.CreateAccountRoleParams{
		AccountID: pgtype.UUID{
			Bytes: accRole.AccountID(),
			Valid: true,
		},
		Role: accRole.Role().String(),
	}
}

func MapAccountRoleToSQLCUpdate(accRole *model.AccountRole) sqlc.UpdateAccountRoleParams {
	return sqlc.UpdateAccountRoleParams{
		AccountID: pgtype.UUID{
			Bytes: accRole.AccountID(),
			Valid: true,
		},
		Role: accRole.Role().String(),
	}
}

func MapSQLCToAccountRole(rawAccRole sqlc.AccountRole) *model.AccountRole {
	accountRole := model.RestoreAccountRole(
		rawAccRole.AccountID.Bytes,
		model.Role(rawAccRole.Role),
	)
	return accountRole
}
