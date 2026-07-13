package port

import (
	"context"

	"github.com/maket12/ads-service/userservice/internal/domain/model"

	"github.com/google/uuid"
)

type ProfileRepository interface {
	Create(ctx context.Context, profile *model.Profile) error
	Get(ctx context.Context, accountID uuid.UUID) (*model.Profile, error)
	Update(ctx context.Context, profile *model.Profile) error
	Delete(ctx context.Context, accountID uuid.UUID) error
	ListProfiles(ctx context.Context, limit, offset int) ([]*model.Profile, error)
}
