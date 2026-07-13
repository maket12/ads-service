-- name: CreateAccountRole :exec
INSERT INTO account_roles (
    account_id,
    role
) VALUES (
    $1, $2
 );

-- name: GetAccountRole :one
SELECT
    account_id,
    role
FROM account_roles
WHERE account_id = $1;

-- name: UpdateAccountRole :exec
UPDATE account_roles
SET role = $2
WHERE account_id = $1;

-- name: DeleteAccountRole :exec
DELETE FROM account_roles
WHERE account_id = $1;
