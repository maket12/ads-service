package port

import "context"

type EmailSender interface {
	SendVerificationEmail(ctx context.Context, toEmail, token string) error
}
