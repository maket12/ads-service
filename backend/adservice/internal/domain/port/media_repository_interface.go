package port

import (
	"context"

	"github.com/google/uuid"
)

type ImageInput struct {
	ID        string
	URL       string
	Width     int
	Height    int
	SizeBytes int64
	Format    string
}

type ImageRef struct {
	ID     string
	URL    string
	Width  int
	Height int
}

type MediaRepository interface {
	Save(ctx context.Context, adID uuid.UUID, images []ImageInput) error
	Get(ctx context.Context, adID uuid.UUID) ([]ImageRef, error)
	Delete(ctx context.Context, adID uuid.UUID) error
}
