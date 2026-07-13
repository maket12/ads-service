package redis

import (
	"time"

	"github.com/google/uuid"
)

type redisTokenDTO struct {
	Token     string    `json:"token"`
	AccountID uuid.UUID `json:"account_id"`
	TTL       int64     `json:"ttl_ns"` // nanoseconds
	ExpiresAt time.Time `json:"expires_at"`
}
