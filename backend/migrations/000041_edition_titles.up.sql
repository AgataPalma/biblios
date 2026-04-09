-- Add title and original_title to book_editions
ALTER TABLE public.book_editions
    ADD COLUMN title character varying(500),
    ADD COLUMN original_title character varying(500);

-- Backfill title before NOT NULL
UPDATE public.book_editions be
SET title = b.title
FROM public.books b
WHERE b.id = be.book_id
  AND be.title IS NULL;

-- NOT NULL on title
ALTER TABLE public.book_editions
    ALTER COLUMN title SET NOT NULL;

-- Regular indexes
CREATE INDEX idx_book_editions_title
    ON public.book_editions USING btree (lower(title));

CREATE INDEX idx_book_editions_original_title
    ON public.book_editions USING btree (lower(original_title));

CREATE INDEX idx_book_editions_dedupe
    ON public.book_editions USING btree (lower(original_title), lower(language));

-- DUAL-LANGUAGE GIN full-text search index
CREATE INDEX idx_book_editions_titles_dual_fts
    ON public.book_editions
        USING gin ((
                       setweight(
                               to_tsvector('portuguese', coalesce(title, '') || ' ' || coalesce(original_title, '')),
                               'A'
                       ) ||
                       setweight(
                               to_tsvector('english', coalesce(title, '') || ' ' || coalesce(original_title, '')),
                               'B'
                       )
                       ));