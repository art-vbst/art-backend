-- name: CreateImage :one
INSERT INTO images (
        artwork_id,
        image_url,
        is_main_image,
        image_width,
        image_height
    )
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateImage :one
UPDATE images
SET is_main_image = $2
WHERE id = $1
RETURNING *;

-- name: DeleteImage :exec
DELETE FROM images
WHERE id = $1;