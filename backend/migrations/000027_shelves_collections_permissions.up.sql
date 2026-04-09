-- 000027_shelves_collections_permissions.up.sql

-- Collections: add collaborative flag
ALTER TABLE collections
    ADD COLUMN is_collaborative BOOLEAN NOT NULL DEFAULT false;

-- Shelves: personal lightweight tags
CREATE TABLE shelves (
                         id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                         user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                         name       VARCHAR(100) NOT NULL,
                         created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                         UNIQUE (user_id, name)
);

CREATE TABLE shelf_books (
                             shelf_id   UUID NOT NULL REFERENCES shelves(id) ON DELETE CASCADE,
                             copy_id    UUID NOT NULL REFERENCES book_copies(id) ON DELETE CASCADE,
                             added_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                             PRIMARY KEY (shelf_id, copy_id)
);

CREATE INDEX idx_shelves_user     ON shelves(user_id);
CREATE INDEX idx_shelf_books_copy ON shelf_books(copy_id);

-- Library members: replace role with granular permissions
ALTER TABLE library_members
    DROP CONSTRAINT library_members_role_check,
    DROP COLUMN role,
    ADD COLUMN is_owner            BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN can_view            BOOLEAN NOT NULL DEFAULT true,
    ADD COLUMN can_add             BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN can_remove          BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN can_edit            BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN can_invite          BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN can_manage_members  BOOLEAN NOT NULL DEFAULT false;

-- Owner must always have all permissions
ALTER TABLE library_members
    ADD CONSTRAINT owner_has_all_permissions CHECK (
        is_owner = false OR (
            can_view           = true AND
            can_add            = true AND
            can_remove         = true AND
            can_edit           = true AND
            can_invite         = true AND
            can_manage_members = true
            )
        );
