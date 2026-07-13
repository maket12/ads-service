package dto

import "github.com/google/uuid"

type SendVerificationInput struct {
	AccountID uuid.UUID
	Email     string
}

type SendVerificationOutput struct {
	Sent bool
}
