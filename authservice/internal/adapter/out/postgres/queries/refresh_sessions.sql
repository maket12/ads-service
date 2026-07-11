-- name: CreateRefreshSession :exec
INSERT INTO refresh_sessions (
  id,
  account_id,
  refresh_token_hash,
  created_at,
  expires_at,
  revoked_at,
  revoke_reason,
  rotated_from,
  ip,
  user_agent
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
);

-- name: GetRefreshSessionByID :one
SELECT
    id,
    account_id,
    refresh_token_hash,
    created_at,
    expires_at,
    revoked_at,
    revoke_reason,
    rotated_from,
    ip,
    user_agent
FROM refresh_sessions
WHERE id = $1 LIMIT 1;

-- name: GetRefreshSessionByHash :one
SELECT
    id,
    account_id,
    refresh_token_hash,
    created_at,
    expires_at,
    revoked_at,
    revoke_reason,
    rotated_from,
    ip,
    user_agent
FROM refresh_sessions
WHERE refresh_token_hash = $1;

-- name: UpdateRefreshSession :exec
UPDATE refresh_sessions
SET
    account_id = $2,
    refresh_token_hash = $3,
    created_at = $4,
    expires_at = $5,
    revoked_at = $6,
    revoke_reason = $7,
    rotated_from = $8,
    ip = $9,
    user_agent = $10
WHERE id = $1;

-- name: RevokeRefreshSessionDescendants :exec
WITH RECURSIVE chain(target_id) AS (
    SELECT rs_init.id
    FROM refresh_sessions rs_init
    WHERE rs_init.id = $1

    UNION ALL

    SELECT rs_rec.id
    FROM refresh_sessions rs_rec
             JOIN chain c ON rs_rec.rotated_from = c.target_id
)
UPDATE refresh_sessions
SET
    revoked_at = $2,
    revoke_reason = $3
FROM chain
WHERE refresh_sessions.id = chain.target_id
  AND refresh_sessions.revoked_at IS NULL;

-- name: ListAccountActiveRefreshSessions :many
SELECT
    id,
    account_id,
    refresh_token_hash,
    created_at,
    expires_at,
    revoked_at,
    revoke_reason,
    rotated_from,
    ip,
    user_agent
FROM refresh_sessions
WHERE account_id = $1
    AND revoked_at IS NULL
    AND expires_at > $2
ORDER BY created_at DESC;

-- name: DeleteExpiredRefreshSessions :exec
DELETE FROM refresh_sessions
WHERE expires_at <= $1;

-- name: RevokeAllAccountRefreshSessions :exec
UPDATE refresh_sessions
SET
    revoked_at = $2,
    revoke_reason = $3
WHERE account_id = $1
    AND revoked_at IS NULL;
