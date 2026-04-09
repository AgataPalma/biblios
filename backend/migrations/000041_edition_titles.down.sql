-- Drop indexes
DROP INDEX IF EXISTS public.idx_book_editions_titles_dual_fts;
DROP INDEX IF EXISTS public.idx_book_editions_title;
DROP INDEX IF EXISTS public.idx_book_editions_original_title;
DROP INDEX IF EXISTS public.idx_book_editions_dedupe;

-- Drop columns
ALTER TABLE public.book_editions
    DROP COLUMN IF EXISTS title,
    DROP COLUMN IF EXISTS original_title;