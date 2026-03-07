-- Full-text search: trigger-maintained tsvector column on books
-- Combines title, description, and all associated author names into one searchable vector

-- Add the search_vector column to books
ALTER TABLE books ADD COLUMN search_vector TSVECTOR;

-- Function to rebuild the search vector for a single book
-- Weights: A = title, B = author names, C = description
CREATE OR REPLACE FUNCTION books_search_vector(
    p_title       TEXT,
    p_description TEXT,
    p_authors     TEXT   -- space-separated author names, pre-joined by the trigger
) RETURNS TSVECTOR AS $$
BEGIN
    RETURN (
        setweight(to_tsvector('english', coalesce(p_title, '')), 'A') ||
        setweight(to_tsvector('english', coalesce(p_authors, '')), 'B') ||
        setweight(to_tsvector('english', coalesce(p_description, '')), 'C')
    );
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Trigger function: fires on INSERT/UPDATE of books, joins authors, rebuilds vector
CREATE OR REPLACE FUNCTION books_search_vector_trigger() RETURNS TRIGGER AS $$
DECLARE
    v_authors TEXT;
BEGIN
    SELECT string_agg(a.name, ' ')
    INTO v_authors
    FROM book_authors ba
    JOIN authors a ON a.id = ba.author_id
    WHERE ba.book_id = NEW.id;

    NEW.search_vector := books_search_vector(NEW.title, NEW.description, v_authors);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER books_search_vector_update
    BEFORE INSERT OR UPDATE OF title, description ON books
    FOR EACH ROW EXECUTE FUNCTION books_search_vector_trigger();

-- GIN index on the computed vector (fast full-text queries)
CREATE INDEX idx_books_search_vector ON books USING GIN(search_vector);

-- GIN index on authors.name for standalone author search
CREATE INDEX idx_authors_name_fts ON authors USING GIN(to_tsvector('english', name));

-- Backfill existing rows
UPDATE books SET title = title;
