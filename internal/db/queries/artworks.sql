-- name: CreateArtwork :one
INSERT INTO artworks (
        title,
        painting_number,
        painting_year,
        width_inches,
        height_inches,
        price_cents,
        paper,
        sort_order,
        sold_at,
        status,
        medium,
        category
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8,
        $9,
        $10,
        $11,
        $12
    )
RETURNING *;


-- name: GetArtwork :one
SELECT *
FROM artworks
WHERE id = $1
LIMIT 1;


-- name: GetArtworkWithImages :many
SELECT a.*,
    i.id as image_id,
    i.image_url,
    i.image_width,
    i.image_height,
    i.created_at as image_created_at
FROM artworks a
    LEFT JOIN images i ON a.id = i.artwork_id
WHERE a.id = $1
ORDER BY i.created_at;


-- name: ListArtworks :many
SELECT *
FROM artworks
ORDER BY sort_order,
    created_at DESC;


-- name: ListArtworksByStatus :many
SELECT *
FROM artworks
WHERE status = $1
ORDER BY sort_order,
    created_at DESC;


-- name: ListAvailableArtworks :many
SELECT *
FROM artworks
WHERE status = 'available'
ORDER BY sort_order,
    created_at DESC
LIMIT $1 OFFSET $2;


-- name: ListArtworksByCategory :many
SELECT *
FROM artworks
WHERE category = $1
ORDER BY sort_order,
    created_at DESC;


-- name: ListArtworksByStatusAndCategory :many
SELECT *
FROM artworks
WHERE status = $1
    AND category = $2
ORDER BY sort_order,
    created_at DESC;


-- name: SearchArtworksByTitle :many
SELECT *
FROM artworks
WHERE title ILIKE '%' || $1 || '%'
ORDER BY sort_order,
    created_at DESC;


-- name: UpdateArtwork :one
UPDATE artworks
SET title = COALESCE(sqlc.narg('title'), title),
    painting_number = COALESCE(sqlc.narg('painting_number'), painting_number),
    painting_year = COALESCE(sqlc.narg('painting_year'), painting_year),
    width_inches = COALESCE(sqlc.narg('width_inches'), width_inches),
    height_inches = COALESCE(sqlc.narg('height_inches'), height_inches),
    price_cents = COALESCE(sqlc.narg('price_cents'), price_cents),
    paper = COALESCE(sqlc.narg('paper'), paper),
    sort_order = COALESCE(sqlc.narg('sort_order'), sort_order),
    sold_at = COALESCE(sqlc.narg('sold_at'), sold_at),
    status = COALESCE(sqlc.narg('status'), status),
    medium = COALESCE(sqlc.narg('medium'), medium),
    category = COALESCE(sqlc.narg('category'), category),
    updated_at = current_timestamp
WHERE id = sqlc.arg('id')
RETURNING *;


-- name: UpdateArtworkStatus :one
UPDATE artworks
SET status = $2,
    sold_at = CASE
        WHEN $2 = 'sold' THEN current_timestamp
        ELSE sold_at
    END,
    updated_at = current_timestamp
WHERE id = $1
RETURNING *;


-- name: UpdateArtworkSortOrder :exec
UPDATE artworks
SET sort_order = $2,
    updated_at = current_timestamp
WHERE id = $1;


-- name: DeleteArtwork :exec
DELETE FROM artworks
WHERE id = $1;


-- name: CountArtworksByStatus :one
SELECT COUNT(*)
FROM artworks
WHERE status = $1;


-- name: GetRecentArtworks :many
SELECT *
FROM artworks
ORDER BY created_at DESC
LIMIT $1;


-- name: GetArtworksByPriceRange :many
SELECT *
FROM artworks
WHERE status = 'available'
    AND price_cents >= $1
    AND price_cents <= $2
ORDER BY price_cents;