package port

import (
	"context"
	"time"

	"github.com/maket12/ads-service/authservice/internal/domain/model"

	"github.com/google/uuid"
)

type RefreshSessionRepository interface {
	Create(ctx context.Context, session *model.RefreshSession) error
	GetByHash(ctx context.Context, tokenHash string) (*model.RefreshSession, error)
	GetByID(ctx context.Context, tokenID uuid.UUID) (*model.RefreshSession, error)
	Revoke(ctx context.Context, session *model.RefreshSession) error
	RevokeAllForAccount(ctx context.Context, accountID uuid.UUID, reason *string) error
	RevokeDescendants(ctx context.Context, sessionID uuid.UUID, reason *string) error
	DeleteExpired(ctx context.Context, expiresAt time.Time) error
	ListActiveForAccount(ctx context.Context, accountID uuid.UUID) ([]*model.RefreshSession, error)
}
