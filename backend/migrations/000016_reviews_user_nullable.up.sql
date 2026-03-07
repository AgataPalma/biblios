-- Allow reviews to be anonymised when a user deletes their account.
-- Drops the NOT NULL constraint and changes the FK to SET NULL on user delete.

ALTER TABLE reviews
    ALTER COLUMN user_id DROP NOT NULL;

ALTER TABLE reviews
    DROP CONSTRAINT IF EXISTS reviews_user_id_fkey;

ALTER TABLE reviews
    ADD CONSTRAINT reviews_user_id_fkey
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE SET NULL;
