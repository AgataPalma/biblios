package tokenstore

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Store struct {
	rdb *redis.Client
}

func NewStore(rdb *redis.Client) *Store {
	return &Store{rdb: rdb}
}

func (s *Store) RevokeToken(ctx context.Context, tokenID string, expiry time.Duration) error {
	var err error = s.rdb.Set(ctx, "revoked:"+tokenID, "1", expiry).Err()
	if err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}
	return nil
}

func (s *Store) IsRevoked(ctx context.Context, tokenID string) (bool, error) {
	var val string
	var err error

	val, err = s.rdb.Get(ctx, "revoked:"+tokenID).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check token: %w", err)
	}

	return val == "1", nil
}
