-- Collections: named subsets of books within a library (e.g. "Favourites", "Summer Reading")
CREATE TABLE collections (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    library_id   UUID NOT NULL REFERENCES libraries(id) ON DELETE CASCADE,
    created_by   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name         VARCHAR(255) NOT NULL,
    description  TEXT,
    cover_colour VARCHAR(7),  -- hex colour e.g. #3a86ff
    is_public    BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at   TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- A book copy can belong to multiple collections
CREATE TABLE collection_books (
    collection_id UUID NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
    book_copy_id  UUID NOT NULL REFERENCES book_copies(id) ON DELETE CASCADE,
    added_by      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    added_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (collection_id, book_copy_id)
);

-- Indexes
CREATE INDEX idx_collections_library        ON collections(library_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_collections_created_by     ON collections(created_by);
CREATE INDEX idx_collection_books_copy      ON collection_books(book_copy_id);
CREATE INDEX idx_collection_books_collection ON collection_books(collection_id);

-- updated_at trigger
CREATE TRIGGER collections_updated_at
    BEFORE UPDATE ON collections
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
