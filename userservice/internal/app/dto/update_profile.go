package dto

import (
	"github.com/google/uuid"
)

type UpdateProfileInput struct {
	AccountID uuid.UUID
	FirstName *string
	LastName  *string
	Phone     *string
	AvatarURL *string
	Bio       *string
}

type UpdateProfileOutput struct {
	Success bool
}
