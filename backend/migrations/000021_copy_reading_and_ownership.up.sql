-- Reading progress fields
ALTER TABLE book_copies
    ADD COLUMN IF NOT EXISTS current_page         INT,
    ADD COLUMN IF NOT EXISTS started_reading_at   TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS finished_reading_at  TIMESTAMPTZ;

-- Ownership / location fields
ALTER TABLE book_copies
    ADD COLUMN IF NOT EXISTS owned_by_user  BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS borrowed_from  UUID REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS location       VARCHAR(200);
