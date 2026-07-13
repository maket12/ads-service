DROP TABLE IF EXISTS refresh_sessions;

DROP INDEX IF EXISTS idx_refresh_sessions_account;
DROP INDEX IF EXISTS idx_refresh_sessions_expires;
DROP INDEX IF EXISTS idx_refresh_session_revoked;