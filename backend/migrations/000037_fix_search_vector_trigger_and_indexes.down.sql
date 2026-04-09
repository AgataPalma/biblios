CREATE OR REPLACE FUNCTION public.books_search_vector_trigger()
    RETURNS trigger
    LANGUAGE plpgsql
AS $$
DECLARE
    v_authors TEXT;
BEGIN
    SELECT string_agg(a.name, ' ')
    INTO v_authors
    FROM book_authors ba
             JOIN authors a ON a.id = ba.author_id
    WHERE ba.book_id = NEW.id;

    NEW.search_vector := public.books_search_vector(NEW.title, NEW.description, v_authors);
    RETURN NEW;
END;
$$;

UPDATE public.books SET search_vector = NULL WHERE deleted_at IS NULL;

DROP INDEX IF EXISTS public.idx_book_copies_reading_status;
DROP INDEX IF EXISTS public.idx_books_status;
DROP INDEX IF EXISTS public.idx_book_editions_status;
DROP INDEX IF EXISTS public.idx_notifications_user_read;
DROP INDEX IF EXISTS public.idx_book_copies_owner_status;
