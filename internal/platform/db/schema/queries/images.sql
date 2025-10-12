-- name: CreateImage :one
INSERT INTO images (
        artwork_id,
        image_url,
        image_width,
        image_height
    )
VALUES ($1, $2, $3, $4)
RETURNING *;