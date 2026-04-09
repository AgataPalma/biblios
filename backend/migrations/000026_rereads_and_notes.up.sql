-- 000026_rereads_and_notes.up.sql

ALTER TABLE book_copies
    ADD COLUMN reread_count   INT NOT NULL DEFAULT 0,
    ADD COLUMN personal_notes TEXT;
