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


-- name: GetArtworkRaw :one
SELECT *
FROM artworks
WHERE id = $1
LIMIT 1;


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


-- name: ListArtworksRaw :many
SELECT *
FROM artworks
ORDER BY sort_order,
    created_at DESC;


-- name: ListArtworks :many
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
ORDER BY a.sort_order,
    a.created_at DESC;