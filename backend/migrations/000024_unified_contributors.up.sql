-- 000024_unified_contributors.up.sql

CREATE TABLE contributors (
                              id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                              name             VARCHAR(255) NOT NULL,
                              bio              TEXT,
                              born_date        DATE,
                              died_date        DATE,
                              photo_url        TEXT,
                              website          TEXT,
                              nationality      VARCHAR(100),
                              status           VARCHAR(10) NOT NULL DEFAULT 'pending'
                                  CHECK (status IN ('pending', 'approved', 'rejected')),
                              rejection_reason TEXT,
                              deleted_at       TIMESTAMPTZ,
                              created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                              updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                              CONSTRAINT contributors_name_unique UNIQUE (name)
);

CREATE TRIGGER contributors_updated_at
    BEFORE UPDATE ON contributors
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE INDEX idx_contributors_name_fts
    ON contributors USING gin(to_tsvector('english', name));

INSERT INTO contributors (id, name, bio, born_date, died_date, photo_url,
                          website, nationality, status, rejection_reason,
                          deleted_at, created_at, updated_at)
SELECT id, name, bio, born_date, died_date, photo_url,
       website, nationality, status, rejection_reason,
       deleted_at, created_at, updated_at
FROM authors
ON CONFLICT (name) DO NOTHING;

INSERT INTO contributors (id, name, status, rejection_reason,
                          deleted_at, created_at, updated_at)
SELECT id, name, status, rejection_reason,
       deleted_at, created_at, updated_at
FROM narrators
ON CONFLICT (name) DO NOTHING;

INSERT INTO contributors (id, name, status, rejection_reason,
                          deleted_at, created_at, updated_at)
SELECT id, name, status, rejection_reason,
       deleted_at, created_at, updated_at
FROM translators
ON CONFLICT (name) DO NOTHING;

CREATE TABLE book_contributors (
                                   book_id        UUID NOT NULL REFERENCES books(id) ON DELETE CASCADE,
                                   contributor_id UUID NOT NULL REFERENCES contributors(id) ON DELETE CASCADE,
                                   role           VARCHAR(20) NOT NULL
                                       CHECK (role IN ('author', 'co_author')),
                                   PRIMARY KEY (book_id, contributor_id, role)
);

CREATE INDEX idx_book_contributors_book
    ON book_contributors(book_id);
CREATE INDEX idx_book_contributors_contributor
    ON book_contributors(contributor_id);

INSERT INTO book_contributors (book_id, contributor_id, role)
SELECT book_id, author_id, 'author'
FROM book_authors;

CREATE TABLE edition_contributors (
                                      edition_id     UUID NOT NULL REFERENCES book_editions(id) ON DELETE CASCADE,
                                      contributor_id UUID NOT NULL REFERENCES contributors(id) ON DELETE CASCADE,
                                      role           VARCHAR(20) NOT NULL
                                          CHECK (role IN ('narrator', 'translator', 'illustrator', 'editor')),
                                      PRIMARY KEY (edition_id, contributor_id, role)
);

CREATE INDEX idx_edition_contributors_edition
    ON edition_contributors(edition_id);
CREATE INDEX idx_edition_contributors_contributor
    ON edition_contributors(contributor_id);

INSERT INTO edition_contributors (edition_id, contributor_id, role)
SELECT edition_id, narrator_id, 'narrator'
FROM book_edition_narrators;

INSERT INTO edition_contributors (edition_id, contributor_id, role)
SELECT edition_id, translator_id, 'translator'
FROM book_edition_translators;

DROP TABLE book_edition_narrators;
DROP TABLE book_edition_translators;
DROP TABLE book_authors;
DROP TABLE narrators;
DROP TABLE translators;
DROP TABLE authors;
