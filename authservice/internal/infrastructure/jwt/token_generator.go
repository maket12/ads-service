package jwt

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type CustomClaims struct {
	jwt.RegisteredClaims
	Role string `json:"role,omitempty"`
	Type string `json:"type"`
}

type TokenGenerator struct {
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

func NewTokenGenerator(
	accessSecret, refreshSecret string,
	accessTTL, refreshTTL time.Duration) *TokenGenerator {
	return &TokenGenerator{
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
	}
}

func (gen *TokenGenerator) GenerateAccessToken(_ context.Context, accountID uuid.UUID, role string) (string, error) {
	accessClaims := CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   accountID.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(gen.accessTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Role: role,
		Type: "access",
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessStr, err := accessToken.SignedString(gen.accessSecret)

	if err != nil {
		return "", err
	}

	return accessStr, nil
}

func (gen *TokenGenerator) GenerateRefreshToken(_ context.Context, accountID, sessionID uuid.UUID) (string, error) {
	refreshClaims := CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   accountID.String(),
			ID:        sessionID.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(gen.refreshTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Type: "refresh",
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshStr, err := refreshToken.SignedString(gen.refreshSecret)

	if err != nil {
		return "", err
	}

	return refreshStr, nil
}

func (gen *TokenGenerator) parseAccessToken(_ context.Context, token string) (*CustomClaims, error) {
	accessClaims := &CustomClaims{}

	parsedToken, err := jwt.ParseWithClaims(token, accessClaims, func(token *jwt.Token) (interface{}, error) {
		// Check the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return gen.accessSecret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}), jwt.WithLeeway(30*time.Second))

	if err != nil || !parsedToken.Valid {
		return nil, fmt.Errorf("failed to parse access token: %w", err)
	}

	return accessClaims, nil
}

func (gen *TokenGenerator) parseRefreshToken(_ context.Context, token string) (*CustomClaims, error) {
	refreshClaims := &CustomClaims{}

	parsedToken, err := jwt.ParseWithClaims(token, refreshClaims, func(token *jwt.Token) (interface{}, error) {
		// Check the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return gen.refreshSecret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}), jwt.WithLeeway(30*time.Second))

	if err != nil || !parsedToken.Valid {
		return nil, fmt.Errorf("failed to parse refresh token: %w", err)
	}

	return refreshClaims, nil
}

func (gen *TokenGenerator) ValidateAccessToken(ctx context.Context, token string) (uuid.UUID, string, error) {
	claims, err := gen.parseAccessToken(ctx, token)
	if err != nil {
		return uuid.Nil, "", err
	}

	if claims.Type != "access" {
		return uuid.Nil, "", fmt.Errorf("invalid token type")
	}

	sub, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("failed to get account_id: %w", err)
	}
	role := claims.Role
	if role == "" {
		return uuid.Nil, "", fmt.Errorf("failed to get account role")
	}

	return sub, role, nil
}

func (gen *TokenGenerator) ValidateRefreshToken(ctx context.Context, token string) (uuid.UUID, uuid.UUID, error) {
	claims, err := gen.parseRefreshToken(ctx, token)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}

	if claims.Type != "refresh" {
		return uuid.Nil, uuid.Nil, fmt.Errorf("invalid token type")
	}

	sub, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("failed to get account_id: %w", err)
	}
	id, err := uuid.Parse(claims.ID)
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("failed to get session id: %w", err)
	}

	return sub, id, nil
}
