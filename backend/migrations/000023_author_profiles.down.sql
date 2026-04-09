-- 000023_author_profiles.down.sql

ALTER TABLE authors
    DROP COLUMN IF EXISTS bio,
    DROP COLUMN IF EXISTS born_date,
    DROP COLUMN IF EXISTS died_date,
    DROP COLUMN IF EXISTS website,
    DROP COLUMN IF EXISTS photo_url,
    DROP COLUMN IF EXISTS nationality;
