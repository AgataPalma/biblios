-- 000023_author_profiles.up.sql

ALTER TABLE authors
    ADD COLUMN bio          TEXT,
    ADD COLUMN born_date    DATE,
    ADD COLUMN died_date    DATE,
    ADD COLUMN website      TEXT,
    ADD COLUMN photo_url    TEXT,
    ADD COLUMN nationality  VARCHAR(100);
