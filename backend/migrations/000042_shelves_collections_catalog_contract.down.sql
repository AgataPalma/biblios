DROP VIEW IF EXISTS public.approved_catalog_books;

DROP INDEX IF EXISTS public.idx_collections_scope;
ALTER TABLE public.collections DROP CONSTRAINT IF EXISTS collections_scope_check;
ALTER TABLE public.collections DROP COLUMN IF EXISTS scope;

DROP INDEX IF EXISTS public.idx_shelves_collection;
DROP INDEX IF EXISTS public.idx_shelves_library;
DROP INDEX IF EXISTS public.uq_shelves_collection_name;
DROP INDEX IF EXISTS public.uq_shelves_library_name;
DROP INDEX IF EXISTS public.uq_shelves_personal_name;

ALTER TABLE public.shelves
    ADD CONSTRAINT shelves_user_id_name_key UNIQUE (user_id, name);

ALTER TABLE public.shelves DROP CONSTRAINT IF EXISTS shelves_collection_id_fkey;
ALTER TABLE public.shelves DROP CONSTRAINT IF EXISTS shelves_library_id_fkey;
ALTER TABLE public.shelves DROP CONSTRAINT IF EXISTS shelves_scope_target_check;
ALTER TABLE public.shelves DROP CONSTRAINT IF EXISTS shelves_scope_check;

ALTER TABLE public.shelves
    DROP COLUMN IF EXISTS collection_id,
    DROP COLUMN IF EXISTS library_id,
    DROP COLUMN IF EXISTS scope;