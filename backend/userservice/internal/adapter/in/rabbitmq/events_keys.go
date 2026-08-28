package rabbitmq

import "github.com/google/uuid"

const (
	RoutingKeyAccountCreated = "account.created"
	RoutingKeyAccountDeleted = "account.deleted"
)

type AccountCreatedEvent struct {
	AccountID uuid.UUID `json:"account_id"`
}

type AccountDeletedEvent struct {
	AccountID uuid.UUID `json:"account_id"`
}
