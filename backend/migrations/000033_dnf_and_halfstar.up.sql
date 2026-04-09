-- 000033_dnf_and_halfstar.up.sql

ALTER TABLE book_copies
    DROP CONSTRAINT book_copies_reading_status_check,
    ADD CONSTRAINT book_copies_reading_status_check
        CHECK (reading_status IN (
                                  'want_to_read', 'reading', 'read', 'did_not_finish'
            ));

ALTER TABLE reviews
    DROP CONSTRAINT reviews_rating_check,
    ALTER COLUMN rating TYPE NUMERIC(2,1),
    ADD CONSTRAINT reviews_rating_check
        CHECK (
            rating >= 0.5 AND rating <= 5.0
                AND rating * 2 = FLOOR(rating * 2)
            );
