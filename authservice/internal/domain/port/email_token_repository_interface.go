package port

import "context"

type EmailTokenRepository interface {
	Save(ctx context.Context, token model)
}
