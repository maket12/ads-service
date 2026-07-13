package mapper

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/maket12/ads-service/authservice/internal/adapter/out/postgres/sqlc"
	"github.com/maket12/ads-service/authservice/internal/domain/model"
)

func MapAccountToSQLCCreate(acc *model.Account) sqlc.CreateAccountParams {
	var lastLogin pgtype.Timestamptz
	if acc.LastLoginAt() != nil {
		lastLogin = pgtype.Timestamptz{
			Time:  *acc.LastLoginAt(),
			Valid: true,
		}
	}

	return sqlc.CreateAccountParams{
		ID: pgtype.UUID{
			Bytes: acc.ID(),
			Valid: true,
		},
		Email:         acc.Email(),
		PasswordHash:  acc.PasswordHash(),
		Status:        acc.Status().String(),
		EmailVerified: acc.EmailVerified(),
		CreatedAt: pgtype.Timestamptz{
			Time:  acc.CreatedAt(),
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:  acc.UpdatedAt(),
			Valid: true,
		},
		LastLoginAt: lastLogin,
	}
}

func MapAccountToSQLCUpdate(acc *model.Account) sqlc.UpdateAccountParams {
	var lastLoginTime pgtype.Timestamptz
	if acc.LastLoginAt() != nil {
		lastLoginTime = pgtype.Timestamptz{
			Time:  *acc.LastLoginAt(),
			Valid: true,
		}
	}

	return sqlc.UpdateAccountParams{
		ID: pgtype.UUID{
			Bytes: acc.ID(),
			Valid: true,
		},
		Email:         acc.Email(),
		PasswordHash:  acc.PasswordHash(),
		Status:        acc.Status().String(),
		EmailVerified: acc.EmailVerified(),
		CreatedAt: pgtype.Timestamptz{
			Time:  acc.CreatedAt(),
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:  acc.UpdatedAt(),
			Valid: true,
		},
		LastLoginAt: lastLoginTime,
	}
}

func MapSQLCToAccount(rawAccount sqlc.Account) *model.Account {
	var lastLogin *time.Time
	if rawAccount.LastLoginAt.Valid {
		// Create a local copy to obtain a stable pointer for the domain model
		t := rawAccount.LastLoginAt.Time
		lastLogin = &t
	}

	account := model.RestoreAccount(
		rawAccount.ID.Bytes,
		rawAccount.Email,
		rawAccount.PasswordHash,
		model.AccountStatus(rawAccount.Status),
		rawAccount.EmailVerified,
		rawAccount.CreatedAt.Time,
		rawAccount.UpdatedAt.Time,
		lastLogin,
	)

	return account
}
