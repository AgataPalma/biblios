-- 000036_fix_search_vector_and_gaps.up.sql

-- 1. Fix the search vector function — no description on books table
CREATE OR REPLACE FUNCTION public.books_search_vector_trigger()
    RETURNS trigger
    LANGUAGE plpgsql
AS $$
DECLARE
    v_contributors TEXT;
BEGIN
    SELECT string_agg(c.name, ' ')
    INTO v_contributors
    FROM book_contributors bc
             JOIN contributors c ON c.id = bc.contributor_id
    WHERE bc.book_id = NEW.id
      AND bc.role IN ('author', 'co_author');

    NEW.search_vector := books_search_vector(
            NEW.title,
            NULL,           -- no description on books, lives on editions
            v_contributors
                         );
    RETURN NEW;
END;
$$;

-- 2. Attach the trigger to the books table
CREATE TRIGGER books_search_vector_update
    BEFORE INSERT OR UPDATE ON books
    FOR EACH ROW EXECUTE FUNCTION books_search_vector_trigger();

-- 3. Backfill search_vector for all existing books
UPDATE books b SET search_vector = (
    SELECT books_search_vector(
                   b.title,
                   NULL,
                   string_agg(c.name, ' ')
           )
    FROM book_contributors bc
             JOIN contributors c ON c.id = bc.contributor_id
    WHERE bc.book_id = b.id
      AND bc.role IN ('author', 'co_author')
);

-- 4. Add contributor_id to submissions
ALTER TABLE submissions
    ADD COLUMN contributor_id UUID REFERENCES contributors(id) ON DELETE SET NULL;

CREATE INDEX idx_submissions_contributor
    ON submissions(contributor_id);

-- 5. Add missing indexes
CREATE INDEX idx_review_likes_user
    ON review_likes(user_id);

CREATE INDEX idx_contributor_awards_award
    ON contributor_awards(award_id);

CREATE INDEX idx_book_awards_award
    ON book_awards(award_id);
