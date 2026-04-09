-- 000034_reading_sessions.down.sql

DROP INDEX IF EXISTS idx_reading_sessions_copy;
DROP INDEX IF EXISTS idx_reading_sessions_user_date;
DROP TABLE IF EXISTS reading_sessions;
