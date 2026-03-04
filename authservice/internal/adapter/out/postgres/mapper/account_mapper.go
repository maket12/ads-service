package mapper

import (
	"database/sql"
	"github.com/maket12/ads-service/authservice/internal/adapter/out/postgres/sqlc"
	"github.com/maket12/ads-service/authservice/internal/domain/model"
	"time"
)

func MapAccountToSQLCCreate(account *model.Account) sqlc.CreateAccountParams {
	var lastLogin sql.NullTime
	if account.LastLoginAt() != nil {
		lastLogin = sql.NullTime{Time: *account.LastLoginAt(), Valid: true}
	}

	return sqlc.CreateAccountParams{
		ID:            account.ID(),
		Email:         account.Email(),
		PasswordHash:  account.PasswordHash(),
		Status:        sqlc.AccountStatus(account.Status()),
		EmailVerified: account.EmailVerified(),
		CreatedAt:     account.CreatedAt(),
		UpdatedAt:     account.UpdatedAt(),
		LastLoginAt:   lastLogin,
	}
}

func MapSQLCToAccount(rawAccount sqlc.Account) *model.Account {
	var lastLogin *time.Time
	if rawAccount.LastLoginAt.Valid {
		// Create a local copy to obtain a stable pointer for the domain model
		t := rawAccount.LastLoginAt.Time
		lastLogin = &t
	}

	status := model.AccountStatus(rawAccount.Status)

	account := model.RestoreAccount(
		rawAccount.ID,
		rawAccount.Email,
		rawAccount.PasswordHash,
		status,
		rawAccount.EmailVerified,
		rawAccount.CreatedAt,
		rawAccount.UpdatedAt,
		lastLogin,
	)

	return account
}
