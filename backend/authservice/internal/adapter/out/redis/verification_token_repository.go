package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/maket12/ads-service/authservice/internal/domain/model"
	pkgerrs "github.com/maket12/ads-service/authservice/pkg/errs"
	pkgredis "github.com/maket12/ads-service/authservice/pkg/redis"
	"github.com/redis/go-redis/v9"
)

type VerificationTokenRepository struct {
	client *pkgredis.Client
}

func NewVerificationTokenRepository(redisClient *pkgredis.Client) *VerificationTokenRepository {
	return &VerificationTokenRepository{client: redisClient}
}

func (r *VerificationTokenRepository) redisKey(token string) string {
	return "verification_token" + token
}

func (r *VerificationTokenRepository) Save(
	ctx context.Context,
	vToken *model.VerificationToken,
) error {
	jsonData, err := json.Marshal(mapVerificationTokenToRedisDTO(vToken))
	if err != nil {
		return fmt.Errorf("failed to marshal token to json: %w", err)
	}

	key := r.redisKey(vToken.Token())
	err = r.client.Set(ctx, key, jsonData, vToken.TTL()).Err()
	if err != nil {
		return err
	}

	return nil
}

func (r *VerificationTokenRepository) Get(
	ctx context.Context,
	token string,
) (*model.VerificationToken, error) {
	key := r.redisKey(token)
	jsonData, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, pkgerrs.NewObjectNotFoundError(
				"verification_token", fmt.Sprintf("%s...", token),
			)
		}
		return nil, err
	}

	var dto redisTokenDTO
	if err = json.Unmarshal([]byte(jsonData), &dto); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token json: %w", err)
	}

	return mapRedisDTOToVerificationToken(dto), nil
}

func (r *VerificationTokenRepository) Delete(ctx context.Context, token string) error {
	key := r.redisKey(token)
	return r.client.Del(ctx, key).Err()
}
