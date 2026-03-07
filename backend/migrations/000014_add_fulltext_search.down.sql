DROP TRIGGER IF EXISTS books_search_vector_update ON books;
DROP FUNCTION IF EXISTS books_search_vector_trigger();
DROP FUNCTION IF EXISTS books_search_vector(TEXT, TEXT, TEXT);
DROP INDEX IF EXISTS idx_books_search_vector;
DROP INDEX IF EXISTS idx_authors_name_fts;
ALTER TABLE books DROP COLUMN IF EXISTS search_vector;
