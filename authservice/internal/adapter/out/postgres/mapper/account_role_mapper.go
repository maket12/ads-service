package mapper

import (
	"github.com/maket12/ads-service/authservice/internal/adapter/out/postgres/sqlc"
	"github.com/maket12/ads-service/authservice/internal/domain/model"
)

func MapAccountRoleToSQLCCreate(accountRole *model.AccountRole) sqlc.CreateAccountRoleParams {
	return sqlc.CreateAccountRoleParams{
		AccountID: accountRole.AccountID(),
		Role:      sqlc.RoleType(accountRole.Role()),
	}
}

func MapSQLCToAccountRole(rawAccountRole sqlc.AccountRole) *model.AccountRole {
	role := model.Role(rawAccountRole.Role)
	accountRole := model.RestoreAccountRole(
		rawAccountRole.AccountID,
		role,
	)
	return accountRole
}
