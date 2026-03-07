-- Libraries: named personal or cooperative book collections owned by a user
CREATE TABLE libraries (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name           VARCHAR(255) NOT NULL,
    description    TEXT,
    is_cooperative BOOLEAN NOT NULL DEFAULT FALSE,
    visibility     VARCHAR(20) NOT NULL DEFAULT 'private'
                       CHECK (visibility IN ('private', 'semi_public', 'public')),
    deleted_at     TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_libraries_owner    ON libraries(owner_id);
CREATE INDEX idx_libraries_visibility ON libraries(visibility) WHERE deleted_at IS NULL;

-- updated_at trigger
CREATE TRIGGER libraries_updated_at
    BEFORE UPDATE ON libraries
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
