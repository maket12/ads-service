package port

import "context"

type PhoneValidator interface {
	Validate(ctx context.Context, phone string) (string, error)
}
