package dto

import "github.com/google/uuid"

type SendVerificationInput struct {
	AccountID uuid.UUID
}

type SendVerificationOutput struct {
	Sent bool
}
