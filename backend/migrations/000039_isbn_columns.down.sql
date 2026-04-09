-- Drop constraints and indexes
ALTER TABLE public.book_editions DROP CONSTRAINT IF EXISTS book_editions_isbn13_unique;
ALTER TABLE public.book_editions DROP CONSTRAINT IF EXISTS book_editions_isbn13_format_check;
DROP INDEX IF EXISTS idx_book_editions_isbn10;
DROP INDEX IF EXISTS idx_book_editions_isbn13;

-- Drop isbn13 and rename isbn10 back
ALTER TABLE public.book_editions DROP COLUMN isbn13;
ALTER TABLE public.book_editions RENAME COLUMN isbn10 TO isbn;
