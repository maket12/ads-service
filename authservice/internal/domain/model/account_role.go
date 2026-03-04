package model

import (
	pkgerrs "github.com/maket12/ads-service/pkg/errs"
	"strings"

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

// ================ Mutation ================

func (a *AccountRole) Assign(rawRole string) error {
	lowerRawRole := strings.ToLower(rawRole)
	if lowerRawRole != "user" && lowerRawRole != "admin" {
		return pkgerrs.NewValueInvalidError("role")
	}
	a.role = Role(lowerRawRole)
	return nil
}
