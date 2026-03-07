-- Revert: restore NOT NULL and original FK behaviour.
-- Note: this will fail if any reviews have already been anonymised (user_id = NULL).

ALTER TABLE reviews
    DROP CONSTRAINT IF EXISTS reviews_user_id_fkey;

ALTER TABLE reviews
    ADD CONSTRAINT reviews_user_id_fkey
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE;

ALTER TABLE reviews
    ALTER COLUMN user_id SET NOT NULL;
