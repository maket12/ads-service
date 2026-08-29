package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/maket12/ads-service/backend/adservice/internal/domain/model"
)

type AdPublisher interface {
	PublishAdPublished(ctx context.Context, ad *model.Ad) error
	PublishAdUpdated(ctx context.Context, ad *model.Ad) error
	PublishAdRejected(ctx context.Context, adID uuid.UUID) error
	PublishAdDeleted(ctx context.Context, adID uuid.UUID) error
}
