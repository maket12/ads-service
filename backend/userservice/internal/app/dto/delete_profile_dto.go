package dto

import "github.com/google/uuid"

type DeleteProfileInput struct {
	AccountID uuid.UUID
}

type DeleteProfileOutput struct {
	Deleted bool
}
