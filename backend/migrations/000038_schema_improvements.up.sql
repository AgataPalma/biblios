
-- 2. Drop contributors unique name constraint
ALTER TABLE public.contributors
    DROP CONSTRAINT contributors_name_unique;

-- Add index on name for fast lookup/deduplication in app layer
CREATE INDEX idx_contributors_name ON public.contributors USING btree (name);

-- 3. Series — create dedicated table
CREATE TABLE public.series (
                               id      UUID DEFAULT gen_random_uuid() NOT NULL,
                               name    character varying(255) NOT NULL,
                               description text,
                               status  character varying(10) DEFAULT 'pending' NOT NULL,
                               rejection_reason text,
                               deleted_at  timestamp with time zone,
                               created_at  timestamp with time zone DEFAULT now() NOT NULL,
                               updated_at  timestamp with time zone DEFAULT now() NOT NULL,
                               CONSTRAINT series_pkey PRIMARY KEY (id),
                               CONSTRAINT series_name_unique UNIQUE (name),
                               CONSTRAINT series_status_check CHECK (
                                   (status)::text = ANY (
                                       (ARRAY['pending'::character varying, 'approved'::character varying, 'rejected'::character varying])::text[]
                                       )
                                   )
);

CREATE INDEX idx_series_name ON public.series USING btree (name);
CREATE INDEX idx_series_status ON public.series USING btree (status) WHERE deleted_at IS NULL;

CREATE TRIGGER series_updated_at
    BEFORE UPDATE ON public.series
    FOR EACH ROW EXECUTE FUNCTION public.update_updated_at();

-- Move series info from book_editions to books
ALTER TABLE public.books
    ADD COLUMN series_id uuid,
    ADD COLUMN series_position numeric(5,1);

ALTER TABLE public.books
    ADD CONSTRAINT books_series_id_fkey
        FOREIGN KEY (series_id) REFERENCES public.series(id) ON DELETE SET NULL;

CREATE INDEX idx_books_series ON public.books USING btree (series_id) WHERE series_id IS NOT NULL;

-- Migrate existing series data from book_editions to books + series table
-- Creates series records and links books, using the most common series_name per book
WITH edition_series AS (
    SELECT DISTINCT ON (book_id)
        book_id,
        series_name,
        series_position
    FROM public.book_editions
    WHERE series_name IS NOT NULL
    ORDER BY book_id, series_name
),
     inserted_series AS (
         INSERT INTO public.series (name)
             SELECT DISTINCT series_name FROM edition_series
             ON CONFLICT DO NOTHING
             RETURNING id, name
     )
UPDATE public.books b
SET
    series_id = s.id,
    series_position = es.series_position
FROM edition_series es
         JOIN public.series s ON s.name = es.series_name
WHERE b.id = es.book_id;

-- Drop series columns from book_editions
ALTER TABLE public.book_editions
    DROP COLUMN series_name,
    DROP COLUMN series_position;

-- 4. Moods — add approval workflow to align with genres
ALTER TABLE public.moods
    ADD COLUMN status character varying(10) DEFAULT 'approved' NOT NULL,
    ADD COLUMN rejection_reason text;

-- Default existing moods to approved since they were already in use
ALTER TABLE public.moods
    ADD CONSTRAINT moods_status_check CHECK (
        (status)::text = ANY (
            (ARRAY['pending'::character varying, 'approved'::character varying, 'rejected'::character varying])::text[]
            )
        );

CREATE INDEX idx_moods_status ON public.moods USING btree (status);

-- 5. Reading sessions — drop one-session-per-day constraint
ALTER TABLE public.reading_sessions
    DROP CONSTRAINT reading_sessions_user_id_copy_id_logged_date_key;

-- 6. Books — add canonical description
ALTER TABLE public.books
    ADD COLUMN description text;
