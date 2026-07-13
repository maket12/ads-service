CREATE TABLE IF NOT EXISTS refresh_sessions (
    id uuid PRIMARY KEY,
    account_id uuid NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    refresh_token_hash text NOT NULL UNIQUE,
    created_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz NOT NULL,
    revoked_at timestamptz,
    revoke_reason text,
    rotated_from uuid REFERENCES refresh_sessions(id),
    ip inet,
    user_agent text
);

CREATE INDEX IF NOT EXISTS idx_refresh_sessions_account ON refresh_sessions(account_id);
CREATE INDEX IF NOT EXISTS idx_refresh_sessions_expires ON refresh_sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_refresh_session_revoked ON refresh_sessions(revoked_at);