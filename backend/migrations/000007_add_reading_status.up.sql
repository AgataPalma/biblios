
ALTER TABLE book_copies
    ADD COLUMN IF NOT EXISTS reading_status VARCHAR(20) NOT NULL DEFAULT 'want_to_read'
    CHECK (reading_status IN ('want_to_read', 'reading', 'read'));
