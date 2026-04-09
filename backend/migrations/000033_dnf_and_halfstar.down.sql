-- 000033_dnf_and_halfstar.down.sql

ALTER TABLE reviews
    DROP CONSTRAINT reviews_rating_check,
    ALTER COLUMN rating TYPE SMALLINT USING ROUND(rating)::SMALLINT,
    ADD CONSTRAINT reviews_rating_check
        CHECK (rating >= 1 AND rating <= 5);

ALTER TABLE book_copies
    DROP CONSTRAINT book_copies_reading_status_check,
    ADD CONSTRAINT book_copies_reading_status_check
        CHECK (reading_status IN ('want_to_read', 'reading', 'read'));
