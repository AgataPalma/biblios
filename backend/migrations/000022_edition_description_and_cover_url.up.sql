-- Move description and cover_url from books to book_editions.
-- Each edition now carries its own synopsis and cover image, preventing
-- different editions of the same book from overwriting each other's data.

-- 1. Add columns to book_editions
ALTER TABLE book_editions ADD COLUMN IF NOT EXISTS description TEXT;
ALTER TABLE book_editions ADD COLUMN IF NOT EXISTS cover_url   TEXT;

-- 2. Copy existing book data down to all their editions
UPDATE book_editions be
SET    description = b.description,
       cover_url   = b.cover_url
FROM   books b
WHERE  be.book_id = b.id
  AND  be.deleted_at IS NULL;

-- 3. Drop dependent trigger first, then drop columns from books
DROP TRIGGER IF EXISTS books_search_vector_update ON books;

ALTER TABLE books DROP COLUMN IF EXISTS description;
ALTER TABLE books DROP COLUMN IF EXISTS cover_url;