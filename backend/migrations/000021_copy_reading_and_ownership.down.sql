ALTER TABLE book_copies
    DROP COLUMN IF EXISTS current_page,
    DROP COLUMN IF EXISTS started_reading_at,
    DROP COLUMN IF EXISTS finished_reading_at,
    DROP COLUMN IF EXISTS owned_by_user,
    DROP COLUMN IF EXISTS borrowed_from,
    DROP COLUMN IF EXISTS location;
