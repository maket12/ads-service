package model

import (
	"strings"

	pkgerrs "github.com/maket12/ads-service/authservice/pkg/errs"

	"github.com/google/uuid"
)

type Role string

func (r Role) String() string { return string(r) }

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

// ================ Rich model for account's Role ================

type AccountRole struct {
	accountID uuid.UUID
	role      Role
}

func NewAccountRole(accountID uuid.UUID) (*AccountRole, error) {
	if accountID == uuid.Nil {
		return nil, pkgerrs.NewValueInvalidError("account_id")
	}
	return &AccountRole{
		accountID: accountID,
		role:      RoleUser,
	}, nil
}

func RestoreAccountRole(accountID uuid.UUID, role Role) *AccountRole {
	return &AccountRole{
		accountID: accountID,
		role:      role,
	}
}

// ================ Read-Only ================

func (a *AccountRole) AccountID() uuid.UUID { return a.accountID }
func (a *AccountRole) Role() Role           { return a.role }

// ================ Business Logic ================

func (a *AccountRole) IsUser() bool  { return a.role == RoleUser }
func (a *AccountRole) IsAdmin() bool { return a.role == RoleAdmin }

// ================ Mutation ================

func (a *AccountRole) Assign(rawRole string) error {
	lowerRawRole := strings.ToLower(rawRole)
	if lowerRawRole != RoleUser.String() && lowerRawRole != RoleAdmin.String() {
		return pkgerrs.NewValueInvalidError("role")
	}
	a.role = Role(lowerRawRole)
	return nil
}
