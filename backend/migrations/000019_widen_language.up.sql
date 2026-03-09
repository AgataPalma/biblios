-- Widen language column to accommodate codes like 'pt-BR', 'other', etc.
ALTER TABLE book_editions ALTER COLUMN language TYPE VARCHAR(10);
