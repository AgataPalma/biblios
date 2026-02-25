DROP TRIGGER IF EXISTS email_queue_updated_at ON email_queue;
DROP TABLE IF EXISTS email_queue;
DROP TABLE IF EXISTS notifications;

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_admin_role_sync;
ALTER TABLE users DROP COLUMN IF EXISTS role;