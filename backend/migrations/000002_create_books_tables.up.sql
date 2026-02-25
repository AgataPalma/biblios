-- Authors
CREATE TABLE authors (
                         id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                         name       VARCHAR(255) NOT NULL,
                         deleted_at TIMESTAMPTZ,
                         created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                         updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Narrators
CREATE TABLE narrators (
                           id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                           name       VARCHAR(255) NOT NULL,
                           deleted_at TIMESTAMPTZ,
                           created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                           updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Translators
CREATE TABLE translators (
                             id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                             name       VARCHAR(255) NOT NULL,
                             deleted_at TIMESTAMPTZ,
                             created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                             updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Genres
CREATE TABLE genres (
                        id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                        name       VARCHAR(100) NOT NULL UNIQUE,
                        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Books (universal, edition-independent)
CREATE TABLE books (
                       id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                       title        VARCHAR(500) NOT NULL,
                       description  TEXT,
                       cover_url    TEXT,
                       deleted_at   TIMESTAMPTZ,
                       created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                       updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Book Editions (a specific published version of a book)
CREATE TABLE book_editions (
                               id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                               book_id          UUID NOT NULL REFERENCES books(id) ON DELETE CASCADE,
                               format           VARCHAR(20) NOT NULL CHECK (format IN ('hardcover', 'paperback', 'ebook', 'audiobook')),
                               isbn             VARCHAR(20),
                               asin             VARCHAR(20),
                               language         CHAR(2) NOT NULL,
                               publisher        VARCHAR(255),
                               edition          VARCHAR(50),
                               published_at     DATE,
    -- Physical/ebook
                               page_count       INTEGER,
    -- Ebook specific
                               file_format      VARCHAR(10) CHECK (file_format IN ('EPUB', 'PDF', 'MOBI', 'AZW3')),
    -- Audiobook specific
                               duration_minutes INTEGER,
                               audio_format     VARCHAR(10) CHECK (audio_format IN ('MP3', 'AAC', 'WMA', 'FLAC')),
                               deleted_at       TIMESTAMPTZ,
                               created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                               updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Book Copies (a physical or digital copy owned by a user)
CREATE TABLE book_copies (
                             id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                             edition_id UUID NOT NULL REFERENCES book_editions(id) ON DELETE CASCADE,
                             owner_id   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                             condition  VARCHAR(10) CHECK (condition IN ('new', 'good', 'fair', 'poor')),
                             deleted_at TIMESTAMPTZ,
                             created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                             updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Join: books <-> authors
CREATE TABLE book_authors (
                              book_id   UUID NOT NULL REFERENCES books(id) ON DELETE CASCADE,
                              author_id UUID NOT NULL REFERENCES authors(id) ON DELETE CASCADE,
                              PRIMARY KEY (book_id, author_id)
);

-- Join: books <-> genres
CREATE TABLE book_genres (
                             book_id  UUID NOT NULL REFERENCES books(id) ON DELETE CASCADE,
                             genre_id UUID NOT NULL REFERENCES genres(id) ON DELETE CASCADE,
                             PRIMARY KEY (book_id, genre_id)
);

-- Join: book_editions <-> translators
CREATE TABLE book_edition_translators (
                                          edition_id    UUID NOT NULL REFERENCES book_editions(id) ON DELETE CASCADE,
                                          translator_id UUID NOT NULL REFERENCES translators(id) ON DELETE CASCADE,
                                          PRIMARY KEY (edition_id, translator_id)
);

-- Join: book_editions <-> narrators
CREATE TABLE book_edition_narrators (
                                        edition_id  UUID NOT NULL REFERENCES book_editions(id) ON DELETE CASCADE,
                                        narrator_id UUID NOT NULL REFERENCES narrators(id) ON DELETE CASCADE,
                                        PRIMARY KEY (edition_id, narrator_id)
);

-- Indexes
CREATE INDEX idx_books_title ON books(title);
CREATE INDEX idx_book_editions_book ON book_editions(book_id);
CREATE INDEX idx_book_editions_isbn ON book_editions(isbn);
CREATE INDEX idx_book_editions_format ON book_editions(format);
CREATE INDEX idx_book_copies_edition ON book_copies(edition_id);
CREATE INDEX idx_book_copies_owner ON book_copies(owner_id);
CREATE INDEX idx_book_authors_book ON book_authors(book_id);
CREATE INDEX idx_book_authors_author ON book_authors(author_id);
CREATE INDEX idx_book_genres_book ON book_genres(book_id);
CREATE INDEX idx_book_genres_genre ON book_genres(genre_id);
CREATE INDEX idx_book_edition_translators_edition ON book_edition_translators(edition_id);
CREATE INDEX idx_book_edition_narrators_edition ON book_edition_narrators(edition_id);

-- updated_at triggers
CREATE TRIGGER authors_updated_at
    BEFORE UPDATE ON authors
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER narrators_updated_at
    BEFORE UPDATE ON narrators
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER translators_updated_at
    BEFORE UPDATE ON translators
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER books_updated_at
    BEFORE UPDATE ON books
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER book_editions_updated_at
    BEFORE UPDATE ON book_editions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER book_copies_updated_at
    BEFORE UPDATE ON book_copies
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();