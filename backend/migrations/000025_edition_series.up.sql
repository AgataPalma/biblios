-- 000025_edition_series.up.sql

ALTER TABLE book_editions
    ADD COLUMN series_name     VARCHAR(255),
    ADD COLUMN series_position NUMERIC(5,1);

ALTER TABLE book_editions
    DROP CONSTRAINT book_editions_format_check,
    ADD CONSTRAINT book_editions_format_check
        CHECK (format IN (
                          'hardcover', 'paperback', 'ebook', 'audiobook', 'graphic_novel'
            ));
