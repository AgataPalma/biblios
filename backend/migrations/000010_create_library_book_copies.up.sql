-- Library book copies: assigns a user's book copy to a specific library
-- A copy belongs to exactly one library at a time
CREATE TABLE library_book_copies (
    library_id   UUID NOT NULL REFERENCES libraries(id) ON DELETE CASCADE,
    book_copy_id UUID NOT NULL REFERENCES book_copies(id) ON DELETE CASCADE,
    added_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (library_id, book_copy_id)
);

-- Index to quickly find which library a copy is in
CREATE INDEX idx_library_book_copies_copy    ON library_book_copies(book_copy_id);
CREATE INDEX idx_library_book_copies_library ON library_book_copies(library_id);
