-- 000029_review_likes.down.sql

DROP INDEX IF EXISTS idx_review_likes_review;
DROP TABLE IF EXISTS review_likes;

ALTER TABLE reviews
    DROP COLUMN IF EXISTS like_count;
