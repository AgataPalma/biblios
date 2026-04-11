ALTER TABLE public.shelves
    ADD COLUMN IF NOT EXISTS scope VARCHAR(20) NOT NULL DEFAULT 'personal',
    ADD COLUMN IF NOT EXISTS library_id UUID,
    ADD COLUMN IF NOT EXISTS collection_id UUID;

ALTER TABLE public.shelves
    DROP CONSTRAINT IF EXISTS shelves_scope_check;

ALTER TABLE public.shelves
    ADD CONSTRAINT shelves_scope_check
        CHECK (scope IN ('personal', 'library', 'collection'));

ALTER TABLE public.shelves
    DROP CONSTRAINT IF EXISTS shelves_scope_target_check;

ALTER TABLE public.shelves
    ADD CONSTRAINT shelves_scope_target_check CHECK (
        (scope = 'personal' AND library_id IS NULL AND collection_id IS NULL) OR
        (scope = 'library' AND library_id IS NOT NULL AND collection_id IS NULL) OR
        (scope = 'collection' AND collection_id IS NOT NULL)
        );

ALTER TABLE public.shelves
    ADD CONSTRAINT shelves_library_id_fkey
        FOREIGN KEY (library_id) REFERENCES public.libraries(id) ON DELETE CASCADE;

ALTER TABLE public.shelves
    ADD CONSTRAINT shelves_collection_id_fkey
        FOREIGN KEY (collection_id) REFERENCES public.collections(id) ON DELETE CASCADE;

ALTER TABLE public.shelves
    DROP CONSTRAINT IF EXISTS shelves_user_id_name_key;

CREATE UNIQUE INDEX IF NOT EXISTS uq_shelves_personal_name
    ON public.shelves (user_id, lower(name))
    WHERE scope = 'personal';

CREATE UNIQUE INDEX IF NOT EXISTS uq_shelves_library_name
    ON public.shelves (library_id, lower(name))
    WHERE scope = 'library';

CREATE UNIQUE INDEX IF NOT EXISTS uq_shelves_collection_name
    ON public.shelves (collection_id, lower(name))
    WHERE scope = 'collection';

CREATE INDEX IF NOT EXISTS idx_shelves_library
    ON public.shelves (library_id)
    WHERE scope = 'library';

CREATE INDEX IF NOT EXISTS idx_shelves_collection
    ON public.shelves (collection_id)
    WHERE scope = 'collection';

-- Gap 2: explicit collection scope
ALTER TABLE public.collections
    ADD COLUMN IF NOT EXISTS scope VARCHAR(20) NOT NULL DEFAULT 'personal';

UPDATE public.collections c
SET scope = CASE
                WHEN l.is_cooperative = TRUE OR c.is_collaborative = TRUE THEN 'cooperative'
                ELSE 'personal'
    END
FROM public.libraries l
WHERE c.library_id = l.id;

ALTER TABLE public.collections
    DROP CONSTRAINT IF EXISTS collections_scope_check;

ALTER TABLE public.collections
    ADD CONSTRAINT collections_scope_check
        CHECK (scope IN ('personal', 'cooperative'));

CREATE INDEX IF NOT EXISTS idx_collections_scope
    ON public.collections (scope)
    WHERE deleted_at IS NULL;

-- Gap 4: strict public catalog contract (approved only)
CREATE OR REPLACE VIEW public.approved_catalog_books AS
SELECT b.*
FROM public.books b
WHERE b.deleted_at IS NULL
  AND b.status = 'approved'
  AND EXISTS (
    SELECT 1
    FROM public.book_editions be
    WHERE be.book_id = b.id
      AND be.deleted_at IS NULL
      AND be.status = 'approved'
);
