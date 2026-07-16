package dto

import "github.com/google/uuid"

type DeleteAccountInput struct {
	AccountID uuid.UUID
}

type DeleteAccountOutput struct {
	Deleted bool
}
