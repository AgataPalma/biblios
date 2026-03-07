-- Library members: users who belong to a library (owner, admin, or member)
CREATE TABLE library_members (
    library_id UUID NOT NULL REFERENCES libraries(id) ON DELETE CASCADE,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role       VARCHAR(20) NOT NULL DEFAULT 'member'
                   CHECK (role IN ('owner', 'admin', 'member')),
    joined_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (library_id, user_id)
);

-- Indexes
CREATE INDEX idx_library_members_user    ON library_members(user_id);
CREATE INDEX idx_library_members_library ON library_members(library_id);
