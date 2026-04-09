-- 000024_unified_contributors.down.sql

CREATE TABLE authors (
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
                         CONSTRAINT authors_name_unique UNIQUE (name)
);

CREATE TABLE narrators (
                           id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                           name             VARCHAR(255) NOT NULL,
                           status           VARCHAR(10) NOT NULL DEFAULT 'pending'
                               CHECK (status IN ('pending', 'approved', 'rejected')),
                           rejection_reason TEXT,
                           deleted_at       TIMESTAMPTZ,
                           created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                           updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                           CONSTRAINT narrators_name_unique UNIQUE (name)
);

CREATE TABLE translators (
                             id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                             name             VARCHAR(255) NOT NULL,
                             status           VARCHAR(10) NOT NULL DEFAULT 'pending'
                                 CHECK (status IN ('pending', 'approved', 'rejected')),
                             rejection_reason TEXT,
                             deleted_at       TIMESTAMPTZ,
                             created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                             updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                             CONSTRAINT translators_name_unique UNIQUE (name)
);

CREATE TRIGGER authors_updated_at
    BEFORE UPDATE ON authors
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER narrators_updated_at
    BEFORE UPDATE ON narrators
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER translators_updated_at
    BEFORE UPDATE ON translators
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

INSERT INTO authors (id, name, bio, born_date, died_date, photo_url,
                     website, nationality, status, rejection_reason,
                     deleted_at, created_at, updated_at)
SELECT id, name, bio, born_date, died_date, photo_url,
       website, nationality, status, rejection_reason,
       deleted_at, created_at, updated_at
FROM contributors
WHERE id IN (SELECT contributor_id FROM book_contributors);

INSERT INTO narrators (id, name, status, rejection_reason,
                       deleted_at, created_at, updated_at)
SELECT id, name, status, rejection_reason,
       deleted_at, created_at, updated_at
FROM contributors
WHERE id IN (
    SELECT contributor_id FROM edition_contributors WHERE role = 'narrator'
);

INSERT INTO translators (id, name, status, rejection_reason,
                         deleted_at, created_at, updated_at)
SELECT id, name, status, rejection_reason,
       deleted_at, created_at, updated_at
FROM contributors
WHERE id IN (
    SELECT contributor_id FROM edition_contributors WHERE role = 'translator'
);

CREATE TABLE book_authors (
                              book_id   UUID NOT NULL REFERENCES books(id) ON DELETE CASCADE,
                              author_id UUID NOT NULL REFERENCES authors(id) ON DELETE CASCADE,
                              PRIMARY KEY (book_id, author_id)
);

CREATE TABLE book_edition_narrators (
                                        edition_id  UUID NOT NULL REFERENCES book_editions(id) ON DELETE CASCADE,
                                        narrator_id UUID NOT NULL REFERENCES narrators(id) ON DELETE CASCADE,
                                        PRIMARY KEY (edition_id, narrator_id)
);

CREATE TABLE book_edition_translators (
                                          edition_id    UUID NOT NULL REFERENCES book_editions(id) ON DELETE CASCADE,
                                          translator_id UUID NOT NULL REFERENCES translators(id) ON DELETE CASCADE,
                                          PRIMARY KEY (edition_id, translator_id)
);

INSERT INTO book_authors (book_id, author_id)
SELECT book_id, contributor_id FROM book_contributors;

INSERT INTO book_edition_narrators (edition_id, narrator_id)
SELECT edition_id, contributor_id
FROM edition_contributors WHERE role = 'narrator';

INSERT INTO book_edition_translators (edition_id, translator_id)
SELECT edition_id, contributor_id
FROM edition_contributors WHERE role = 'translator';

DROP INDEX IF EXISTS idx_edition_contributors_contributor;
DROP INDEX IF EXISTS idx_edition_contributors_edition;
DROP INDEX IF EXISTS idx_book_contributors_contributor;
DROP INDEX IF EXISTS idx_book_contributors_book;
DROP INDEX IF EXISTS idx_contributors_name_fts;
DROP TABLE IF EXISTS edition_contributors;
DROP TABLE IF EXISTS book_contributors;
DROP TABLE IF EXISTS contributors;
