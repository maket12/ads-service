package port

import (
	"context"

	"github.com/maket12/ads-service/backend/searchservice/internal/domain/model"
)

type AdIndexRepository interface {
	Index(ctx context.Context, adIndex *model.AdIndex) error
	Delete(ctx context.Context, id string) error
	Search(ctx context.Context, query *model.SearchQuery) ([]*model.AdIndex, int64, error)
}
