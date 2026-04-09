-- 000029_review_likes.up.sql

CREATE TABLE review_likes (
                              review_id UUID NOT NULL REFERENCES reviews(id) ON DELETE CASCADE,
                              user_id   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                              liked_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                              PRIMARY KEY (review_id, user_id)
);

ALTER TABLE reviews
    ADD COLUMN like_count INT NOT NULL DEFAULT 0;

CREATE INDEX idx_review_likes_review ON review_likes(review_id);
