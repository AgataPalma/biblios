-- Remove duplicates first before adding constraints
DELETE FROM authors a
    USING authors b
WHERE a.id > b.id AND a.name = b.name;

DELETE FROM narrators a
    USING narrators b
WHERE a.id > b.id AND a.name = b.name;

DELETE FROM translators a
    USING translators b
WHERE a.id > b.id AND a.name = b.name;

-- Add unique constraints
ALTER TABLE authors ADD CONSTRAINT authors_name_unique UNIQUE (name);
ALTER TABLE narrators ADD CONSTRAINT narrators_name_unique UNIQUE (name);
ALTER TABLE translators ADD CONSTRAINT translators_name_unique UNIQUE (name);