CREATE OR REPLACE FUNCTION public.books_search_vector_trigger()
    RETURNS trigger
    LANGUAGE plpgsql
AS $$
DECLARE
    v_authors TEXT;
BEGIN
    SELECT string_agg(c.name, ' ')
    INTO v_authors
    FROM book_contributors bc
             JOIN contributors c ON c.id = bc.contributor_id
    WHERE bc.book_id = NEW.id;

    NEW.search_vector := public.books_search_vector(NEW.title, NEW.description, v_authors);
    RETURN NEW;
END;
$$;

UPDATE public.books b
SET search_vector = public.books_search_vector(
        b.title,
        NULL,
        (
            SELECT string_agg(c.name, ' ')
            FROM book_contributors bc
                     JOIN contributors c ON c.id = bc.contributor_id
            WHERE bc.book_id = b.id
        )
                    )
WHERE b.deleted_at IS NULL;

CREATE INDEX idx_book_copies_reading_status
    ON public.book_copies(reading_status)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_books_status
    ON public.books(status)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_book_editions_status
    ON public.book_editions(status)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_notifications_user_read
    ON public.notifications(user_id, read_at);

CREATE INDEX idx_book_copies_owner_status
    ON public.book_copies(owner_id, reading_status)
    WHERE deleted_at IS NULL;
