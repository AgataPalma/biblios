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

func sessionKey(userID, jti string) string {
	return "session:" + userID + ":" + jti
}

func (s *Store) StoreToken(ctx context.Context, userID, jti string, ttl time.Duration) error {
	if err := s.rdb.Set(ctx, sessionKey(userID, jti), "1", ttl).Err(); err != nil {
		return fmt.Errorf("store token: %w", err)
	}
	return nil
}

func (s *Store) SessionExists(ctx context.Context, userID, jti string) (bool, error) {
	val, err := s.rdb.Get(ctx, sessionKey(userID, jti)).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("check session: %w", err)
	}
	return val == "1", nil
}

func (s *Store) DeleteToken(ctx context.Context, userID, jti string) error {
	if err := s.rdb.Del(ctx, sessionKey(userID, jti)).Err(); err != nil {
		return fmt.Errorf("delete token: %w", err)
	}
	return nil
}

func (s *Store) DeleteAllUserSessions(ctx context.Context, userID string) error {
	pattern := "session:" + userID + ":*"
	var cursor uint64
	for {
		keys, next, err := s.rdb.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return fmt.Errorf("scan sessions: %w", err)
		}
		if len(keys) > 0 {
			if err := s.rdb.Del(ctx, keys...).Err(); err != nil {
				return fmt.Errorf("delete sessions: %w", err)
			}
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
	return nil
}
