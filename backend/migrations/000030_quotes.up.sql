-- 000030_quotes.up.sql

CREATE TABLE quotes (
                        id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                        book_id     UUID NOT NULL REFERENCES books(id) ON DELETE CASCADE,
                        user_id     UUID REFERENCES users(id) ON DELETE SET NULL,
                        body        TEXT NOT NULL,
                        page_number INT,
                        created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                        deleted_at  TIMESTAMPTZ
);

CREATE INDEX idx_quotes_book
    ON quotes(book_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_quotes_user
    ON quotes(user_id) WHERE deleted_at IS NULL;
