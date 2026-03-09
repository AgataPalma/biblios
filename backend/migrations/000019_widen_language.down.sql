-- Revert language column width (note: data truncation may occur if values exceed 2 chars)
ALTER TABLE book_editions ALTER COLUMN language TYPE CHAR(2);
