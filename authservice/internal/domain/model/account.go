package model

import (
	"time"

	pkgerrs "github.com/maket12/ads-service/pkg/errs"

	"github.com/google/uuid"
)

type AccountStatus string

const (
	AccountActive  AccountStatus = "active"
	AccountBlocked AccountStatus = "blocked"
	AccountDeleted AccountStatus = "deleted"
)

// ================ Rich model for account ================

type Account struct {
	id            uuid.UUID
	email         string
	passwordHash  string
	status        AccountStatus
	emailVerified bool
	createdAt     time.Time
	updatedAt     time.Time
	lastLoginAt   *time.Time
}

func NewAccount(email, passwordHash string) (*Account, error) {
	if email == "" {
		return nil, pkgerrs.NewValueRequiredError("email")
	}
	if passwordHash == "" {
		return nil, pkgerrs.NewValueRequiredError("password_hash")
	}

	now := time.Now()

	return &Account{
		id:           uuid.New(),
		email:        email,
		passwordHash: passwordHash,
		status:       AccountActive,
		createdAt:    now,
		updatedAt:    now,
	}, nil
}

func RestoreAccount(
	id uuid.UUID, email, passwordHash string,
	status AccountStatus, emailVerified bool,
	createdAt, updatedAt time.Time, lastLoginAt *time.Time,
) *Account {
	return &Account{
		id:            id,
		email:         email,
		passwordHash:  passwordHash,
		status:        status,
		emailVerified: emailVerified,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
		lastLoginAt:   lastLoginAt,
	}
}

// ================ Read-Only ================

func (a *Account) ID() uuid.UUID           { return a.id }
func (a *Account) Email() string           { return a.email }
func (a *Account) PasswordHash() string    { return a.passwordHash }
func (a *Account) Status() AccountStatus   { return a.status }
func (a *Account) EmailVerified() bool     { return a.emailVerified }
func (a *Account) CreatedAt() time.Time    { return a.createdAt }
func (a *Account) UpdatedAt() time.Time    { return a.updatedAt }
func (a *Account) LastLoginAt() *time.Time { return a.lastLoginAt }

func (a *Account) CanLogin() bool  { return a.status == AccountActive }
func (a *Account) IsBlocked() bool { return a.status == AccountBlocked }
func (a *Account) IsDeleted() bool { return a.status == AccountDeleted }

// ================ Mutation ================

func (a *Account) Block() {
	a.status = AccountBlocked
	a.updatedAt = time.Now()
}

func (a *Account) Delete() {
	a.status = AccountDeleted
	a.updatedAt = time.Now()
}

func (a *Account) MarkLogin() {
	now := time.Now()
	a.lastLoginAt = &now
	a.updatedAt = now
}

func (a *Account) VerifyEmail() {
	a.emailVerified = true
	a.updatedAt = time.Now()
}
