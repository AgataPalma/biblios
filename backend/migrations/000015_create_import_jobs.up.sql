-- Import jobs: tracks async Goodreads CSV import progress per user
CREATE TABLE import_jobs (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status        VARCHAR(20) NOT NULL DEFAULT 'pending'
                      CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    source        VARCHAR(50) NOT NULL DEFAULT 'goodreads',
    total_rows    INTEGER,
    processed     INTEGER NOT NULL DEFAULT 0,
    imported      INTEGER NOT NULL DEFAULT 0,
    skipped       INTEGER NOT NULL DEFAULT 0,
    failed_rows   INTEGER NOT NULL DEFAULT 0,
    error_message TEXT,
    started_at    TIMESTAMPTZ,
    completed_at  TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index to fetch a user's import history
CREATE INDEX idx_import_jobs_user   ON import_jobs(user_id);
CREATE INDEX idx_import_jobs_status ON import_jobs(status) WHERE status IN ('pending', 'processing');

-- updated_at trigger
CREATE TRIGGER import_jobs_updated_at
    BEFORE UPDATE ON import_jobs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
