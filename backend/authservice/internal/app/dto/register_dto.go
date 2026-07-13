package dto

import "github.com/google/uuid"

type RegisterInput struct {
	Email    string
	Password string
}

type RegisterOutput struct {
	AccountID uuid.UUID
}
