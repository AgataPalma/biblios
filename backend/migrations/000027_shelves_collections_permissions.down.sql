-- 000027_shelves_collections_permissions.down.sql

ALTER TABLE library_members
    DROP CONSTRAINT IF EXISTS owner_has_all_permissions,
    DROP COLUMN IF EXISTS can_manage_members,
    DROP COLUMN IF EXISTS can_invite,
    DROP COLUMN IF EXISTS can_edit,
    DROP COLUMN IF EXISTS can_remove,
    DROP COLUMN IF EXISTS can_add,
    DROP COLUMN IF EXISTS can_view,
    DROP COLUMN IF EXISTS is_owner,
    ADD COLUMN role VARCHAR(20) NOT NULL DEFAULT 'member'
        CHECK (role IN ('owner', 'admin', 'member'));

DROP INDEX IF EXISTS idx_shelf_books_copy;
DROP INDEX IF EXISTS idx_shelves_user;
DROP TABLE IF EXISTS shelf_books;
DROP TABLE IF EXISTS shelves;

ALTER TABLE collections
    DROP COLUMN IF EXISTS is_collaborative;
