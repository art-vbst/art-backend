-- name: CreateArtwork :one
INSERT INTO artworks (
        title,
        painting_number,
        painting_year,
        width_inches,
        height_inches,
        price_cents,
        description,
        paper,
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
        $11
    )
RETURNING *;

-- name: ListArtworks :many
SELECT a.*,
    i.image_id,
    COALESCE(i.object_name, '') as object_name,
    COALESCE(i.image_url, '') as image_url,
    i.image_width,
    i.image_height,
    i.image_created_at
FROM artworks a
    LEFT JOIN LATERAL (
        SELECT id as image_id,
            object_name,
            image_url,
            image_width,
            image_height,
            created_at as image_created_at
        FROM images
        WHERE artwork_id = a.id
        ORDER BY is_main_image DESC NULLS LAST,
            created_at
        LIMIT 1
    ) i ON true
WHERE $1::text [] IS NULL
    OR cardinality($1) = 0
    OR a.status = ANY($1::artwork_status [])
ORDER BY a.sort_order,
    a.created_at DESC;

-- name: ListArtworkStripeData :many
SELECT a.id,
    a.title,
    a.price_cents,
    a.status,
    i.*
FROM artworks a
    LEFT JOIN LATERAL (
        SELECT id as image_id,
            image_url
        FROM images
        WHERE artwork_id = a.id
        ORDER BY is_main_image DESC NULLS LAST,
            created_at
        LIMIT 1
    ) i ON true
WHERE a.id = ANY($1::uuid [])
    AND a.status = 'available';

-- name: GetArtworkWithImages :many
SELECT a.*,
    i.id as image_id,
    i.is_main_image,
    i.object_name,
    i.image_url,
    i.image_width,
    i.image_height,
    i.created_at as image_created_at
FROM artworks a
    LEFT JOIN images i ON a.id = i.artwork_id
WHERE a.id = $1
ORDER BY i.created_at;

-- name: UpdateArtwork :one
UPDATE artworks
SET title = $2,
    painting_number = $3,
    painting_year = $4,
    width_inches = $5,
    height_inches = $6,
    price_cents = $7,
    description = $8,
    paper = $9,
    sort_order = $10,
    status = $11,
    medium = $12,
    category = $13
WHERE id = $1
RETURNING *;

-- name: SelectArtworksForUpdate :many
SELECT *
FROM artworks
WHERE id = ANY($1::uuid [])
    AND status = 'available' FOR
UPDATE;

-- name: UpdateArtworksAsPurchased :many
UPDATE artworks
SET status = 'sold',
    sold_at = current_timestamp,
    order_id = $2
WHERE id = ANY($1::uuid [])
RETURNING *;

-- name: DeleteArtwork :exec
DELETE FROM artworks
WHERE id = $1;