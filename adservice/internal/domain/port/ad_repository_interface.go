package port

import (
	"context"
	"github.com/maket12/ads-service/adservice/internal/domain/model"

	"github.com/google/uuid"
)

type AdRepository interface {
	Create(ctx context.Context, ad *model.Ad) error
	Get(ctx context.Context, id uuid.UUID) (*model.Ad, error)
	Update(ctx context.Context, ad *model.Ad) error
	UpdateStatus(ctx context.Context, ad *model.Ad) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteAll(ctx context.Context, sellerID uuid.UUID) error
	ListAds(ctx context.Context, limit, offset int) ([]*model.Ad, error)
}
