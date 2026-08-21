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

// TODO: Accept and return specified structs with metadata instead of raw urls

type MediaRepository interface {
	Save(ctx context.Context, adID uuid.UUID, images []string) error
	Get(ctx context.Context, adID uuid.UUID) ([]string, error)
	Delete(ctx context.Context, adID uuid.UUID) error
}
