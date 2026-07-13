package port

import (
	"context"

	"github.com/google/uuid"
)

type AccountPublisher interface {
	PublishAccountCreate(ctx context.Context, accountID uuid.UUID) error
}
