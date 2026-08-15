-- Keep invitation state truthful before enforcing one pending invitation per
-- library/email pair. The newest pending invitation remains usable.
UPDATE library_invitations
SET status = 'expired'
WHERE status = 'pending'
  AND expires_at <= NOW();

WITH ranked_pending AS (
    SELECT id,
           ROW_NUMBER() OVER (
               PARTITION BY library_id, LOWER(invited_email)
               ORDER BY created_at DESC, id DESC
           ) AS position
    FROM library_invitations
    WHERE status = 'pending'
)
UPDATE library_invitations invitation
SET status = 'revoked'
FROM ranked_pending ranked
WHERE invitation.id = ranked.id
  AND ranked.position > 1;

CREATE UNIQUE INDEX uq_library_invitations_pending_email
    ON library_invitations (library_id, LOWER(invited_email))
    WHERE status = 'pending';
