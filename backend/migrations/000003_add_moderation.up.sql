-- Add status to all content tables
ALTER TABLE authors ADD COLUMN status VARCHAR(10) NOT NULL DEFAULT 'pending'
    CHECK (status IN ('pending', 'approved', 'rejected'));
ALTER TABLE authors ADD COLUMN rejection_reason TEXT;

ALTER TABLE narrators ADD COLUMN status VARCHAR(10) NOT NULL DEFAULT 'pending'
    CHECK (status IN ('pending', 'approved', 'rejected'));
ALTER TABLE narrators ADD COLUMN rejection_reason TEXT;

ALTER TABLE translators ADD COLUMN status VARCHAR(10) NOT NULL DEFAULT 'pending'
    CHECK (status IN ('pending', 'approved', 'rejected'));
ALTER TABLE translators ADD COLUMN rejection_reason TEXT;

ALTER TABLE genres ADD COLUMN status VARCHAR(10) NOT NULL DEFAULT 'pending'
    CHECK (status IN ('pending', 'approved', 'rejected'));
ALTER TABLE genres ADD COLUMN rejection_reason TEXT;

ALTER TABLE books ADD COLUMN status VARCHAR(10) NOT NULL DEFAULT 'pending'
    CHECK (status IN ('pending', 'approved', 'rejected'));
ALTER TABLE books ADD COLUMN rejection_reason TEXT;

ALTER TABLE book_editions ADD COLUMN status VARCHAR(10) NOT NULL DEFAULT 'pending'
    CHECK (status IN ('pending', 'approved', 'rejected'));
ALTER TABLE book_editions ADD COLUMN rejection_reason TEXT;

-- Submissions table — one per user book submission
CREATE TABLE submissions (
                             id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                             submitted_by     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                             status           VARCHAR(10) NOT NULL DEFAULT 'pending'
                                 CHECK (status IN ('pending', 'approved', 'rejected')),
                             rejection_reason TEXT,
                             reviewed_by      UUID REFERENCES users(id) ON DELETE SET NULL,
                             reviewed_at      TIMESTAMPTZ,
                             book_id          UUID REFERENCES books(id) ON DELETE SET NULL,
                             edition_id       UUID REFERENCES book_editions(id) ON DELETE SET NULL,
                             copy_id          UUID REFERENCES book_copies(id) ON DELETE SET NULL,
                             deleted_at       TIMESTAMPTZ,
                             created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                             updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_submissions_submitted_by ON submissions(submitted_by);
CREATE INDEX idx_submissions_status ON submissions(status);
CREATE INDEX idx_submissions_reviewed_by ON submissions(reviewed_by);

CREATE TRIGGER submissions_updated_at
    BEFORE UPDATE ON submissions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- Moderation log — full audit trail
CREATE TABLE moderation_log (
                                id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                moderator_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                entity_type VARCHAR(50) NOT NULL,
                                entity_id   UUID NOT NULL,
                                action      VARCHAR(20) NOT NULL CHECK (action IN ('approved', 'rejected', 'edited')),
    before      JSONB,
    after       JSONB,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_moderation_log_moderator ON moderation_log(moderator_id);
CREATE INDEX idx_moderation_log_entity ON moderation_log(entity_type, entity_id);
CREATE INDEX idx_moderation_log_action ON moderation_log(action);