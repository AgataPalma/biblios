-- 000034_reading_sessions.up.sql

CREATE TABLE reading_sessions (
                                  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                  user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                  copy_id      UUID NOT NULL REFERENCES book_copies(id) ON DELETE CASCADE,
                                  logged_date  DATE NOT NULL,
                                  pages_read   INT,
                                  progress_pct NUMERIC(5,2)
                                      CHECK (progress_pct >= 0 AND progress_pct <= 100),
                                  note         TEXT,
                                  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                  UNIQUE (user_id, copy_id, logged_date)
);

CREATE INDEX idx_reading_sessions_user_date
    ON reading_sessions(user_id, logged_date DESC);
CREATE INDEX idx_reading_sessions_copy
    ON reading_sessions(copy_id);
