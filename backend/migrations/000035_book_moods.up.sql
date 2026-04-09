-- 000035_book_moods.up.sql

CREATE TABLE moods (
                       id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                       name VARCHAR(50) NOT NULL UNIQUE
);

CREATE TABLE book_moods (
                            book_id UUID NOT NULL REFERENCES books(id) ON DELETE CASCADE,
                            mood_id UUID NOT NULL REFERENCES moods(id) ON DELETE CASCADE,
                            PRIMARY KEY (book_id, mood_id)
);

CREATE INDEX idx_book_moods_book ON book_moods(book_id);

INSERT INTO moods (name) VALUES
                             ('cozy'), ('dark'), ('atmospheric'), ('funny'),
                             ('slow_burn'), ('fast_paced'), ('emotional'),
                             ('thought_provoking'), ('romantic'), ('scary');
