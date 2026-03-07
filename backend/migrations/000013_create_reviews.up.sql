-- Reviews: one review per user per book, with optional body text
CREATE TABLE reviews (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    book_id    UUID NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    rating     SMALLINT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    body       TEXT CHECK (char_length(body) <= 5000),
    is_public  BOOLEAN NOT NULL DEFAULT TRUE,
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (book_id, user_id)
);

-- Indexes
CREATE INDEX idx_reviews_book    ON reviews(book_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_reviews_user    ON reviews(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_reviews_public  ON reviews(book_id, rating) WHERE is_public = TRUE AND deleted_at IS NULL;

-- updated_at trigger
CREATE TRIGGER reviews_updated_at
    BEFORE UPDATE ON reviews
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
