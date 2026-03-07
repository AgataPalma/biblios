-- Library invitations: secure token-based invites to semi-public or cooperative libraries
CREATE TABLE library_invitations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    library_id      UUID NOT NULL REFERENCES libraries(id) ON DELETE CASCADE,
    invited_by      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    -- invited_user_id is set if inviting an existing user; NULL for email-only invites
    invited_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    invited_email   VARCHAR(255) NOT NULL,
    token           VARCHAR(128) NOT NULL UNIQUE,
    status          VARCHAR(20) NOT NULL DEFAULT 'pending'
                        CHECK (status IN ('pending', 'accepted', 'declined', 'expired', 'revoked')),
    accepted_at     TIMESTAMPTZ,
    expires_at      TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_library_invitations_library ON library_invitations(library_id);
CREATE INDEX idx_library_invitations_token   ON library_invitations(token);
CREATE INDEX idx_library_invitations_email   ON library_invitations(invited_email);
