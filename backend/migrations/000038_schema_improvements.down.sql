-- 6. Remove description from books
ALTER TABLE public.books DROP COLUMN description;

-- 5. Restore one-session-per-day constraint
ALTER TABLE public.reading_sessions
    ADD CONSTRAINT reading_sessions_user_id_copy_id_logged_date_key
        UNIQUE (user_id, copy_id, logged_date);

-- 4. Remove moods approval workflow
DROP INDEX IF EXISTS public.idx_moods_status;
ALTER TABLE public.moods
    DROP CONSTRAINT moods_status_check,
    DROP COLUMN rejection_reason,
    DROP COLUMN status;

-- 3. Restore series columns on book_editions
ALTER TABLE public.book_editions
    ADD COLUMN series_name character varying(255),
    ADD COLUMN series_position numeric(5,1);

-- Restore data from series table back to book_editions
UPDATE public.book_editions be
SET
    series_name = s.name,
    series_position = b.series_position
FROM public.books b
         JOIN public.series s ON s.id = b.series_id
WHERE be.book_id = b.id
  AND b.series_id IS NOT NULL;

-- Remove series from books
DROP INDEX IF EXISTS public.idx_books_series;
ALTER TABLE public.books
    DROP CONSTRAINT books_series_id_fkey,
    DROP COLUMN series_position,
    DROP COLUMN series_id;

ALTER TABLE public.series DROP CONSTRAINT series_name_unique;

DROP TRIGGER IF EXISTS series_updated_at ON public.series;
DROP INDEX IF EXISTS public.idx_series_status;
DROP INDEX IF EXISTS public.idx_series_name;
DROP TABLE IF EXISTS public.series;

-- 2. Restore contributors unique name constraint
DROP INDEX IF EXISTS public.idx_contributors_name;
ALTER TABLE public.contributors
    ADD CONSTRAINT contributors_name_unique UNIQUE (name);

