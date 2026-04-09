-- 000031_user_follows.down.sql

DROP INDEX IF EXISTS idx_user_follows_following;
DROP INDEX IF EXISTS idx_user_follows_follower;
DROP TABLE IF EXISTS user_follows;
