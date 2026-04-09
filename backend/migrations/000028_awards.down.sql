-- 000028_awards.down.sql

DROP INDEX IF EXISTS idx_contributor_awards_contributor;
DROP INDEX IF EXISTS idx_book_awards_book;
DROP TABLE IF EXISTS contributor_awards;
DROP TABLE IF EXISTS book_awards;
DROP TABLE IF EXISTS awards;
