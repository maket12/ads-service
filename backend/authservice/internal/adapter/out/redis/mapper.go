package redis

import (
	"time"

	"github.com/maket12/ads-service/backend/authservice/v2/internal/domain/model"
)

func mapVerificationTokenToRedisDTO(vToken *model.VerificationToken) redisTokenDTO {
	return redisTokenDTO{
		Token:     vToken.Token(),
		AccountID: vToken.AccountID(),
		TTL:       vToken.TTL().Nanoseconds(),
		ExpiresAt: vToken.ExpiresAt(),
	}
}

func mapRedisDTOToVerificationToken(dto redisTokenDTO) *model.VerificationToken {
	return model.RestoreVerificationToken(
		dto.Token,
		dto.AccountID,
		time.Duration(dto.TTL),
		dto.ExpiresAt,
	)
}
