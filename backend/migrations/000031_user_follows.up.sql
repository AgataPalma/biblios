-- 000031_user_follows.up.sql

CREATE TABLE user_follows (
                              follower_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                              following_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                              followed_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                              PRIMARY KEY (follower_id, following_id),
                              CHECK (follower_id != following_id)
);

CREATE INDEX idx_user_follows_follower
    ON user_follows(follower_id);
CREATE INDEX idx_user_follows_following
    ON user_follows(following_id);
