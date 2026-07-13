package port

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var ErrTokenExpired = errors.New("token has expired")

type TokensPair struct {
	Access  string
	Refresh string
}

type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

func (t TokenType) String() string { return string(t) }

type TokenGenerator interface {
	GeneratePair(
		ctx context.Context,
		accountID uuid.UUID,
		role string,
		sessionID uuid.UUID,
	) (*TokensPair, error)

	ValidateAccessToken(
		ctx context.Context,
		token string,
	) (accountID uuid.UUID, role string, err error)

	ValidateRefreshToken(
		ctx context.Context,
		token string,
	) (accountID uuid.UUID, sessionID uuid.UUID, err error)
}
