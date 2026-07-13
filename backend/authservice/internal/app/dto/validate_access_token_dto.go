package dto

import "github.com/google/uuid"

type ValidateAccessTokenInput struct {
	AccessToken string
}

type ValidateAccessTokenOutput struct {
	AccountID uuid.UUID
	Role      string
}
