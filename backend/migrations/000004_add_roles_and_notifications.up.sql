-- Replace is_admin boolean with proper role system
ALTER TABLE users ADD COLUMN role VARCHAR(10) NOT NULL DEFAULT 'user'
    CHECK (role IN ('user', 'moderator', 'admin'));

-- Migrate existing admins to new role column
UPDATE users SET role = 'admin' WHERE is_admin = true;

-- Keep is_admin for now but sync it
ALTER TABLE users ADD CONSTRAINT users_admin_role_sync
    CHECK (is_admin = false OR role = 'admin');

-- Notifications table
CREATE TABLE notifications (
                               id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                               user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                               type       VARCHAR(50) NOT NULL,
                               title      VARCHAR(255) NOT NULL,
                               body       TEXT NOT NULL,
                               read_at    TIMESTAMPTZ,
                               data       JSONB,
                               created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_user ON notifications(user_id);
CREATE INDEX idx_notifications_read_at ON notifications(read_at);
CREATE INDEX idx_notifications_type ON notifications(type);

-- Email queue table for background email sending
CREATE TABLE email_queue (
                             id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                             user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                             to_email    VARCHAR(255) NOT NULL,
                             subject     VARCHAR(255) NOT NULL,
                             body        TEXT NOT NULL,
                             status      VARCHAR(10) NOT NULL DEFAULT 'pending'
                                 CHECK (status IN ('pending', 'sent', 'failed')),
                             attempts    INTEGER NOT NULL DEFAULT 0,
                             last_error  TEXT,
                             sent_at     TIMESTAMPTZ,
                             created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                             updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_email_queue_status ON email_queue(status);
CREATE INDEX idx_email_queue_user ON email_queue(user_id);

CREATE TRIGGER email_queue_updated_at
    BEFORE UPDATE ON email_queue
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();