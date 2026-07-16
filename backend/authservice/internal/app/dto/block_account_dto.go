package dto

import "github.com/google/uuid"

type BlockAccountInput struct {
	AccountID uuid.UUID
}

type BlockAccountOutput struct {
	Blocked bool
}
