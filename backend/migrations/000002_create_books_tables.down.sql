DROP TRIGGER IF EXISTS book_copies_updated_at ON book_copies;
DROP TRIGGER IF EXISTS book_editions_updated_at ON book_editions;
DROP TRIGGER IF EXISTS books_updated_at ON books;
DROP TRIGGER IF EXISTS translators_updated_at ON translators;
DROP TRIGGER IF EXISTS narrators_updated_at ON narrators;
DROP TRIGGER IF EXISTS authors_updated_at ON authors;

DROP TABLE IF EXISTS book_edition_narrators;
DROP TABLE IF EXISTS book_edition_translators;
DROP TABLE IF EXISTS book_genres;
DROP TABLE IF EXISTS book_authors;
DROP TABLE IF EXISTS book_copies;
DROP TABLE IF EXISTS book_editions;
DROP TABLE IF EXISTS books;
DROP TABLE IF EXISTS genres;
DROP TABLE IF EXISTS translators;
DROP TABLE IF EXISTS narrators;
DROP TABLE IF EXISTS authors;