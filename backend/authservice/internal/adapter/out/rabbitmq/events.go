package rabbitmq

import "github.com/google/uuid"

type AccountCreatedEvent struct {
	AccountID uuid.UUID `json:"account_id"`
}

type AccountDeletedEvent struct {
	AccountID uuid.UUID `json:"account_id"`
}
