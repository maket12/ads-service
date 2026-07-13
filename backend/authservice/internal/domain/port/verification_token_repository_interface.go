package port

import (
	"context"

	"github.com/maket12/ads-service/authservice/internal/domain/model"
)

type VerificationTokenRepository interface {
	Save(ctx context.Context, token *model.VerificationToken) error
	Get(ctx context.Context, token string) (*model.VerificationToken, error)
	Delete(ctx context.Context, token string) error
}
