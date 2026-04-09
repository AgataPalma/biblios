-- 000028_awards.up.sql

CREATE TABLE awards (
                        id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                        name        VARCHAR(255) NOT NULL UNIQUE,
                        description TEXT
);

CREATE TABLE book_awards (
                             book_id   UUID NOT NULL REFERENCES books(id) ON DELETE CASCADE,
                             award_id  UUID NOT NULL REFERENCES awards(id) ON DELETE CASCADE,
                             year      INT,
                             category  VARCHAR(255),
                             result    VARCHAR(20) CHECK (result IN ('winner', 'nominee')),
                             PRIMARY KEY (book_id, award_id, year)
);

CREATE INDEX idx_book_awards_book
    ON book_awards(book_id);

CREATE TABLE contributor_awards (
                                    contributor_id UUID NOT NULL REFERENCES contributors(id) ON DELETE CASCADE,
                                    award_id       UUID NOT NULL REFERENCES awards(id) ON DELETE CASCADE,
                                    year           INT,
                                    category       VARCHAR(255),
                                    result         VARCHAR(20) CHECK (result IN ('winner', 'nominee')),
                                    PRIMARY KEY (contributor_id, award_id, year)
);

CREATE INDEX idx_contributor_awards_contributor
    ON contributor_awards(contributor_id);
