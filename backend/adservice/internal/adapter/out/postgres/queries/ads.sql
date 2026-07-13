-- name: CreateAd :exec
INSERT INTO ads (
    id,
    seller_id,
    title,
    description,
    price,
    status,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
);

-- name: GetAd :one
SELECT
    id,
    seller_id,
    title,
    description,
    price,
    status,
    created_at,
    updated_at
FROM ads
WHERE id = $1;

-- name: UpdateAd :exec
UPDATE ads
SET
    title = $2,
    description = $3,
    price = $4,
    updated_at = $5
WHERE id = $1;

-- name: UpdateAdStatus :exec
UPDATE ads
SET status = $2
WHERE id = $1;

-- name: DeleteAd :exec
DELETE FROM ads
WHERE id = $1;

-- name: DeleteAllAds :exec
DELETE FROM ads
WHERE seller_id = $1;

-- name: ListAds :many
SELECT
    id,
    seller_id,
    title,
    description,
    price,
    status,
    created_at,
    updated_at
FROM ads
LIMIT $1
OFFSET $2;

-- name: ListSellerAds :many
SELECT
    id,
    seller_id,
    title,
    description,
    price,
    status,
    created_at,
    updated_at
FROM ads
WHERE seller_id = $1
LIMIT $2
OFFSET $3;