package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const leewayVal = 3 * time.Second

type CustomClaims struct {
	jwt.RegisteredClaims
	Role string `json:"role,omitempty"`
}

type TokenGenerator struct {
	secret []byte
	ttl    time.Duration
}

func NewTokenGenerator(secret string, ttl time.Duration) *TokenGenerator {
	return &TokenGenerator{
		secret: []byte(secret),
		ttl:    ttl,
	}
}

func (gen *TokenGenerator) Generate(userID uuid.UUID, role string) (string, error) {
	accessClaims := CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(gen.ttl).UTC()),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
		Role: role,
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessStr, err := accessToken.SignedString(gen.secret)

	if err != nil {
		return "", err
	}

	return accessStr, nil
}

func (gen *TokenGenerator) parse(token string) (*CustomClaims, error) {
	claims := &CustomClaims{}

	parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		// Check the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return gen.secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}), jwt.WithLeeway(leewayVal))

	if err != nil || !parsedToken.Valid {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	return claims, nil
}

func (gen *TokenGenerator) Validate(token string) (uuid.UUID, string, error) {
	claims, err := gen.parse(token)
	if err != nil {
		return uuid.Nil, "", err
	}

	sub, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("failed to get user id: %w", err)
	}
	role := claims.Role
	if role == "" {
		return uuid.Nil, "", fmt.Errorf("failed to get user role")
	}

	return sub, role, nil
}
