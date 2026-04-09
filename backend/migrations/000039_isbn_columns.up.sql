-- Rename existing isbn to isbn10
ALTER TABLE public.book_editions RENAME COLUMN isbn TO isbn10;

-- Add isbn13 column
ALTER TABLE public.book_editions ADD COLUMN isbn13 character varying(17);

-- Unique constraint on isbn13 only
ALTER TABLE public.book_editions
    ADD CONSTRAINT book_editions_isbn13_unique UNIQUE (isbn13);

-- Indexes for both
CREATE INDEX idx_book_editions_isbn10 ON public.book_editions USING btree (isbn10);
CREATE INDEX idx_book_editions_isbn13 ON public.book_editions USING btree (isbn13);

-- CHECK constraint for isbn13 format (13 digits, optional dashes/spaces)
ALTER TABLE public.book_editions
    ADD CONSTRAINT book_editions_isbn13_format_check
        CHECK (
            isbn13 IS NULL
                OR isbn13 ~ '^[0-9]{13}$|^[0-9]{1,5}-?[0-9]{1,7}-?[0-9]{1,6}-?[0-9]$'
            );
