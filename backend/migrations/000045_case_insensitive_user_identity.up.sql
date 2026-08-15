-- Enforce the identity semantics used by registration. These partial indexes
-- protect active accounts while allowing soft-deleted identities to be
-- anonymized/released by the account-deletion workflow.
CREATE UNIQUE INDEX users_email_active_lower_key
    ON users (LOWER(BTRIM(email)))
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX users_username_active_lower_key
    ON users (LOWER(BTRIM(username)))
    WHERE deleted_at IS NULL;
