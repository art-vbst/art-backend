-- name: CreateArtwork :one
INSERT INTO artworks (
        title,
        painting_number,
        painting_year,
        width_inches,
        height_inches,
        price_cents,
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
        $10
    )
RETURNING *;

-- name: GetArtwork :one
SELECT a.*,
    i.*
FROM artworks a
    LEFT JOIN LATERAL (
        SELECT id as image_id,
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
WHERE a.id = $1;

-- name: GetArtworkWithImages :many
SELECT a.*,
    i.id as image_id,
    i.is_main_image,
    i.image_url,
    i.image_width,
    i.image_height,
    i.created_at as image_created_at
FROM artworks a
    LEFT JOIN images i ON a.id = i.artwork_id
WHERE a.id = $1
ORDER BY i.created_at;

-- name: ListArtworks :many
SELECT a.*,
    i.image_id,
    COALESCE(i.image_url, '') as image_url,
    i.image_width,
    i.image_height,
    i.image_created_at
FROM artworks a
    LEFT JOIN LATERAL (
        SELECT id as image_id,
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

-- name: UpdateArtworksForOrder :many
UPDATE artworks
SET status = 'pending',
    order_id = $1,
    updated_at = current_timestamp
WHERE id = ANY($2::uuid [])
RETURNING *;

-- name: UpdateArtworkStatus :many
UPDATE artworks
SET status = $2,
    updated_at = current_timestamp
WHERE order_id = $1
RETURNING *;