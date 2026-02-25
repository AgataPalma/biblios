DROP TABLE IF EXISTS moderation_log;
DROP TABLE IF EXISTS submissions;

ALTER TABLE book_editions DROP COLUMN IF EXISTS rejection_reason;
ALTER TABLE book_editions DROP COLUMN IF EXISTS status;

ALTER TABLE books DROP COLUMN IF EXISTS rejection_reason;
ALTER TABLE books DROP COLUMN IF EXISTS status;

ALTER TABLE genres DROP COLUMN IF EXISTS rejection_reason;
ALTER TABLE genres DROP COLUMN IF EXISTS status;

ALTER TABLE translators DROP COLUMN IF EXISTS rejection_reason;
ALTER TABLE translators DROP COLUMN IF EXISTS status;

ALTER TABLE narrators DROP COLUMN IF EXISTS rejection_reason;
ALTER TABLE narrators DROP COLUMN IF EXISTS status;

ALTER TABLE authors DROP COLUMN IF EXISTS rejection_reason;
ALTER TABLE authors DROP COLUMN IF EXISTS status;