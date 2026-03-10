-- Reverse: move description and cover_url back from book_editions to books
ALTER TABLE books ADD COLUMN IF NOT EXISTS description TEXT;
ALTER TABLE books ADD COLUMN IF NOT EXISTS cover_url   TEXT;

-- Copy the most recently updated edition's data back to the book
UPDATE books b
SET    description = (
    SELECT be.description
    FROM   book_editions be
    WHERE  be.book_id = b.id
      AND  be.description IS NOT NULL
      AND  be.deleted_at IS NULL
    ORDER  BY be.updated_at DESC
    LIMIT 1
),
       cover_url = (
           SELECT be.cover_url
           FROM   book_editions be
           WHERE  be.book_id = b.id
             AND  be.cover_url IS NOT NULL
             AND  be.deleted_at IS NULL
           ORDER  BY be.updated_at DESC
           LIMIT 1
       );

ALTER TABLE book_editions DROP COLUMN IF EXISTS description;
ALTER TABLE book_editions DROP COLUMN IF EXISTS cover_url;

-- Recreate the FTS trigger correctly
CREATE TRIGGER books_search_vector_update
    BEFORE INSERT OR UPDATE OF title, description ON books
    FOR EACH ROW EXECUTE FUNCTION books_search_vector_trigger();

-- Rebuild search vectors after restoring description
UPDATE books SET title = title;