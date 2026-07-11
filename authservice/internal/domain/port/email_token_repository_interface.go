package port

import (
	"context"

	"github.com/maket12/ads-service/authservice/internal/domain/model"
)

type EmailTokenRepository interface {
	Save(ctx context.Context, token model.VerificationToken)
}
