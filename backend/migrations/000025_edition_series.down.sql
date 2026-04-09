-- 000025_edition_series.down.sql

ALTER TABLE book_editions
    DROP COLUMN IF EXISTS series_name,
    DROP COLUMN IF EXISTS series_position;

ALTER TABLE book_editions
    DROP CONSTRAINT book_editions_format_check,
    ADD CONSTRAINT book_editions_format_check
        CHECK (format IN (
                          'hardcover', 'paperback', 'ebook', 'audiobook'
            ));
