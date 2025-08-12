package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"gostartkit/internal/application/apperr"
	"gostartkit/internal/application/ports"

	"github.com/redis/go-redis/v9"
)

// RedisRefreshStore implements RefreshTokenStore using Redis.
// Keys:
//   - refresh:<token> => userID (TTL=token TTL)
//   - refresh_user:<userID>:<token> => 1 (TTL=token TTL) to support revocation per user if needed
type RedisRefreshStore struct{ client *redis.Client }

func NewRedisRefreshStore(addr, password string, db int) *RedisRefreshStore {
	return &RedisRefreshStore{client: redis.NewClient(&redis.Options{Addr: addr, Password: password, DB: db})}
}

func (s *RedisRefreshStore) Issue(ctx context.Context, userID string, ttlSeconds int) (string, error) {
	token, err := secureRandomToken(32)
	if err != nil {
		return "", err
	}
	ttl := time.Duration(ttlSeconds) * time.Second
	pipe := s.client.TxPipeline()
	pipe.Set(ctx, refreshKey(token), userID, ttl)
	pipe.Set(ctx, userTokenKey(userID, token), 1, ttl)
	_, err = pipe.Exec(ctx)
	return token, err
}

func (s *RedisRefreshStore) Rotate(ctx context.Context, oldToken string, ttlSeconds int) (string, string, error) {
	userID, err := s.Validate(ctx, oldToken)
	if err != nil {
		return "", "", err
	}
	if err := s.Revoke(ctx, oldToken); err != nil {
		return "", "", err
	}
	newTok, err := s.Issue(ctx, userID, ttlSeconds)
	if err != nil {
		return "", "", err
	}
	return newTok, userID, nil
}

func (s *RedisRefreshStore) Revoke(ctx context.Context, token string) error {
	uid, err := s.client.Get(ctx, refreshKey(token)).Result()
	if err != nil {
		if err == redis.Nil {
			return apperr.ErrInvalidRefreshToken
		}
		return err
	}
	pipe := s.client.TxPipeline()
	pipe.Del(ctx, refreshKey(token))
	pipe.Del(ctx, userTokenKey(uid, token))
	_, err = pipe.Exec(ctx)
	return err
}

func (s *RedisRefreshStore) Validate(ctx context.Context, token string) (string, error) {
	uid, err := s.client.Get(ctx, refreshKey(token)).Result()
	if err != nil {
		if err == redis.Nil {
			return "", apperr.ErrInvalidRefreshToken
		}
		return "", err
	}
	if uid == "" {
		return "", apperr.ErrInvalidRefreshToken
	}
	return uid, nil
}

func refreshKey(token string) string           { return "refresh:" + token }
func userTokenKey(userID, token string) string { return "refresh_user:" + userID + ":" + token }

func secureRandomToken(numBytes int) (string, error) {
	b := make([]byte, numBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

var _ ports.RefreshTokenStore = (*RedisRefreshStore)(nil)
