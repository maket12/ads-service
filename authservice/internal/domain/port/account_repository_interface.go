package port

import (
	"context"

	"github.com/maket12/ads-service/authservice/internal/domain/model"

	"github.com/google/uuid"
)

type AccountRepository interface {
	Create(ctx context.Context, account *model.Account) error
	GetByEmail(ctx context.Context, email string) (*model.Account, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Account, error)
	Update(ctx context.Context, account *model.Account) error
}
