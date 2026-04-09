-- 000026_rereads_and_notes.down.sql

ALTER TABLE book_copies
    DROP COLUMN IF EXISTS reread_count,
    DROP COLUMN IF EXISTS personal_notes;
