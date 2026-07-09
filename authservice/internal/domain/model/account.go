package model

import (
	"errors"
	"net/mail"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	pkgerrs "github.com/maket12/ads-service/pkg/errs"
)

var (
	ErrCannotBlockAccount  = errors.New("account is already blocked or deleted")
	ErrCannotDeleteAccount = errors.New("account is already deleted")
	ErrCannotLogin         = errors.New("account is not active")
)

type AccountStatus string

const (
	AccountActive  AccountStatus = "active"
	AccountBlocked AccountStatus = "blocked"
	AccountDeleted AccountStatus = "deleted"
)

func (s AccountStatus) String() string { return string(s) }

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

	emailLen := utf8.RuneCountInString(email)
	if emailLen < 6 || emailLen > 254 {
		return nil, pkgerrs.NewValueInvalidError("email")
	}

	_, err := mail.ParseAddress(email)
	if err != nil {
		return nil, pkgerrs.NewValueInvalidErrorWithReason(
			"email", err,
		)
	}

	now := time.Now()

	return &Account{
		id:            uuid.New(),
		email:         email,
		passwordHash:  passwordHash,
		status:        AccountActive,
		emailVerified: false,
		createdAt:     now,
		updatedAt:     now,
		lastLoginAt:   nil,
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

// ================ Business Logic ================

func (a *Account) CanLogin() bool  { return a.status == AccountActive }
func (a *Account) IsBlocked() bool { return a.status == AccountBlocked }
func (a *Account) IsDeleted() bool { return a.status == AccountDeleted }

// ================ Mutation ================

func (a *Account) Block() error {
	if a.IsBlocked() || a.IsDeleted() {
		return ErrCannotBlockAccount
	}

	a.status = AccountBlocked
	a.updatedAt = time.Now()

	return nil
}

func (a *Account) Delete() error {
	if a.IsDeleted() {
		return ErrCannotDeleteAccount
	}

	a.status = AccountDeleted
	a.updatedAt = time.Now()

	return nil
}

func (a *Account) MarkLogin() error {
	if !a.CanLogin() {
		return ErrCannotLogin
	}

	now := time.Now()
	a.lastLoginAt = &now
	a.updatedAt = now

	return nil
}

func (a *Account) VerifyEmail() {
	if a.emailVerified {
		return
	}
	a.emailVerified = true
	a.updatedAt = time.Now()
}
