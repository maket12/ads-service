package port

import (
	"context"

	"github.com/maket12/ads-service/authservice/internal/domain/model"

	"github.com/google/uuid"
)

type AccountRoleRepository interface {
	Create(ctx context.Context, accountRole *model.AccountRole) error
	Get(ctx context.Context, accountID uuid.UUID) (*model.AccountRole, error)
	Update(ctx context.Context, accountRole *model.AccountRole) error
	Delete(ctx context.Context, accountID uuid.UUID) error
}
