-- 000032_reading_challenges.up.sql

CREATE TABLE reading_challenges (
                                    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                    year       INT NOT NULL,
                                    goal       INT NOT NULL,
                                    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                    UNIQUE (user_id, year)
);

CREATE INDEX idx_reading_challenges_user
    ON reading_challenges(user_id);
