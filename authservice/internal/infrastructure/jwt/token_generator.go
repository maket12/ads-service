package jwt

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/maket12/ads-service/authservice/internal/domain/port"
	pkgerrs "github.com/maket12/ads-service/pkg/errs"
)

const leewayVal = 30 * time.Second

type customClaims struct {
	jwt.RegisteredClaims
	Role string `json:"role,omitempty"`
	Type string `json:"type"`
}

type Generator struct {
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

func NewGenerator(
	accessSecret, refreshSecret string,
	accessTTL, refreshTTL time.Duration) *Generator {
	return &Generator{
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
	}
}

func (gen *Generator) generateToken(_ context.Context,
	tokenType port.TokenType, accountID uuid.UUID,
	role *string, sessionID *uuid.UUID,
) (string, error) {
	var (
		claims     customClaims
		signingKey []byte
	)

	if tokenType == port.AccessToken {
		if role == nil {
			return "", pkgerrs.NewValueRequiredError("role")
		}

		claims = customClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   accountID.String(),
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(gen.accessTTL)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
			Role: *role,
			Type: "access",
		}
		signingKey = gen.accessSecret
	} else if tokenType == port.RefreshToken {
		if sessionID == nil {
			return "", pkgerrs.NewValueRequiredError("session_id")
		}

		claims = customClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   accountID.String(),
				ID:        sessionID.String(),
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(gen.refreshTTL)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
			Type: tokenType.String(),
		}
		signingKey = gen.refreshSecret
	} else {
		return "", fmt.Errorf("unknown token type: %s", tokenType)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenStr, err := token.SignedString(signingKey)
	if err != nil {
		return "", err
	}

	return tokenStr, nil
}

func (gen *Generator) GeneratePair(
	ctx context.Context, accountID uuid.UUID,
	role string, sessionID uuid.UUID,
) (*port.TokensPair, error) {
	accessToken, err := gen.generateToken(
		ctx, port.AccessToken,
		accountID, &role, nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := gen.generateToken(
		ctx, port.RefreshToken,
		accountID, nil, &sessionID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &port.TokensPair{
		Access:  accessToken,
		Refresh: refreshToken,
	}, nil
}

func (gen *Generator) parseToken(_ context.Context,
	token string,
	tokenType port.TokenType,
) (*customClaims, error) {
	var (
		claims customClaims
		secret []byte
	)

	switch tokenType {
	case port.AccessToken:
		secret = gen.accessSecret
	case port.RefreshToken:
		secret = gen.refreshSecret
	default:
		return nil, fmt.Errorf("unknown token type: %s", tokenType)
	}

	parsedToken, err := jwt.ParseWithClaims(
		token,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			// Check the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf(
					"unexpected signing method: %v",
					token.Header["alg"],
				)
			}
			return secret, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
		jwt.WithLeeway(leewayVal),
	)
	if err != nil {
		return nil, err
	}

	if !parsedToken.Valid {
		return nil, fmt.Errorf("token is invalid")
	}

	if claims.Type != tokenType.String() {
		return nil, fmt.Errorf("token type mismatch: expected %s, got %s", tokenType, claims.Type)
	}

	return &claims, nil
}

func (gen *Generator) ValidateAccessToken(
	ctx context.Context,
	token string,
) (uuid.UUID, string, error) {
	claims, err := gen.parseToken(ctx, token, port.AccessToken)
	if err != nil {
		return uuid.Nil, "", fmt.Errorf(
			"failed to parse access token: %w", err,
		)
	}

	sub, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, "", fmt.Errorf(
			"failed to get account_id: %w", err,
		)
	}

	if claims.Role == "" {
		return uuid.Nil, "", fmt.Errorf("failed to get account role")
	}

	return sub, claims.Role, nil
}

func (gen *Generator) ValidateRefreshToken(ctx context.Context, token string) (uuid.UUID, uuid.UUID, error) {
	claims, err := gen.parseToken(ctx, token, port.RefreshToken)
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf(
			"failed to parse refresh token: %w", err,
		)
	}

	sub, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("failed to get account_id: %w", err)
	}

	sessionID, err := uuid.Parse(claims.ID)
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("failed to get session id: %w", err)
	}

	return sub, sessionID, nil
}
