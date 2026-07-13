-- name: CreateProfile :exec
INSERT INTO profiles (
    account_id,
    first_name,
    last_name,
    phone,
    avatar_url,
    bio,
    updated_at
) VALUES (
    $1, $2, $3,
    $4, $5, $6, $7
);

-- name: GetProfile :one
SELECT
    account_id,
    first_name,
    last_name,
    phone,
    avatar_url,
    bio,
    updated_at
FROM profiles
WHERE account_id = $1;

-- name: UpdateProfile :exec
UPDATE profiles
SET
    first_name = $2,
    last_name = $3,
    phone = $4,
    avatar_url = $5,
    bio = $6,
    updated_at = $7
WHERE account_id = $1;

-- name: DeleteProfile :exec
DELETE FROM profiles
WHERE account_id = $1;

-- name: ListProfiles :many
SELECT
    account_id,
    first_name,
    last_name,
    phone,
    avatar_url,
    bio,
    updated_at
FROM profiles
LIMIT $1
OFFSET $2;