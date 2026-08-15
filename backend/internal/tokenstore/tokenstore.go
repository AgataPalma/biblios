package tokenstore

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrUserDisabled = errors.New("user sessions are disabled")

var storeTokenScript = redis.NewScript(`
if redis.call("EXISTS", KEYS[1]) == 1 then
    return 0
end
redis.call("SET", KEYS[2], "1", "PX", ARGV[1])
return 1
`)

var sessionExistsScript = redis.NewScript(`
if redis.call("EXISTS", KEYS[1]) == 1 then
    return 0
end
if redis.call("GET", KEYS[2]) == "1" then
    return 1
end
return 0
`)

type Store struct {
	rdb *redis.Client
}

func NewStore(rdb *redis.Client) *Store {
	return &Store{rdb: rdb}
}

func sessionKey(userID, jti string) string {
	return "session:" + userID + ":" + jti
}

func disabledUserKey(userID string) string {
	return "session-disabled:" + userID
}

func (s *Store) StoreToken(ctx context.Context, userID, jti string, ttl time.Duration) error {
	stored, err := storeTokenScript.Run(
		ctx,
		s.rdb,
		[]string{disabledUserKey(userID), sessionKey(userID, jti)},
		ttl.Milliseconds(),
	).Int()
	if err != nil {
		return fmt.Errorf("store token: %w", err)
	}
	if stored == 0 {
		return ErrUserDisabled
	}
	return nil
}

func (s *Store) SessionExists(ctx context.Context, userID, jti string) (bool, error) {
	exists, err := sessionExistsScript.Run(
		ctx,
		s.rdb,
		[]string{disabledUserKey(userID), sessionKey(userID, jti)},
	).Int()
	if err != nil {
		return false, fmt.Errorf("check session: %w", err)
	}
	return exists == 1, nil
}

func (s *Store) DeleteToken(ctx context.Context, userID, jti string) error {
	if err := s.rdb.Del(ctx, sessionKey(userID, jti)).Err(); err != nil {
		return fmt.Errorf("delete token: %w", err)
	}
	return nil
}

// DisableUser prevents both existing and concurrently-created sessions from
// authenticating. The marker intentionally has no TTL because deleted account
// IDs must remain disabled permanently.
func (s *Store) DisableUser(ctx context.Context, userID string) error {
	if err := s.rdb.Set(ctx, disabledUserKey(userID), "1", 0).Err(); err != nil {
		return fmt.Errorf("disable user sessions: %w", err)
	}
	return nil
}

// EnableUser rolls back a disable operation when the database deletion fails.
func (s *Store) EnableUser(ctx context.Context, userID string) error {
	if err := s.rdb.Del(ctx, disabledUserKey(userID)).Err(); err != nil {
		return fmt.Errorf("enable user sessions: %w", err)
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
