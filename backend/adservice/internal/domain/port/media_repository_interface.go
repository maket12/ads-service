package port

import (
	"context"

	"github.com/google/uuid"
)

type MediaRepository interface {
	Save(ctx context.Context, adID uuid.UUID, images []string) error
	Get(ctx context.Context, adID uuid.UUID) ([]string, error)
	Delete(ctx context.Context, adID uuid.UUID) error
}
