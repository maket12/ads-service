package dto

import "github.com/google/uuid"

type AssignRoleInput struct {
	AccountID uuid.UUID
	Role      string
}

type AssignRoleOutput struct {
	Assign bool
}
