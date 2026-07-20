package dto

import (
	"time"

	"github.com/google/uuid"
)

type GetProfileInput struct {
	AccountID uuid.UUID
}

type GetProfileOutput struct {
	AccountID uuid.UUID
	FirstName *string
	LastName  *string
	Phone     *string
	AvatarURL *string
	Bio       *string
	UpdatedAt time.Time
}
