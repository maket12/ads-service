-- name: CreateAccount :exec
INSERT INTO accounts (
    id,
    email,
    password_hash,
    status,
    email_verified,
    created_at,
    updated_at,
    last_login_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
);

-- name: GetAccountByEmail :one
SELECT
    id,
    email,
    password_hash,
    status,
    email_verified,
    created_at,
    updated_at,
    last_login_at
FROM accounts
WHERE email = $1;

-- name: GetAccountByID :one
SELECT
    id,
    email,
    password_hash,
    status,
    email_verified,
    created_at,
    updated_at,
    last_login_at
FROM accounts
WHERE id = $1;

-- name: UpdateAccount :exec
UPDATE accounts
SET
    email = $2,
    password_hash = $3,
    status = $4,
    email_verified = $5,
    created_at = $6,
    updated_at = $7,
    last_login_at = $8
WHERE id = $1;
