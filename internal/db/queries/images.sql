-- name: CreateImage :one
INSERT INTO images (
        artwork_id,
        image_url,
        image_width,
        image_height
    )
VALUES ($1, $2, $3, $4)
RETURNING *;


-- -- name: GetImage :one
-- SELECT *
-- FROM images
-- WHERE id = $1;
-- -- name: ListImagesByArtwork :many
-- SELECT *
-- FROM images
-- WHERE artwork_id = $1
-- ORDER BY created_at;
-- -- name: UpdateImage :one
-- UPDATE images
-- SET image_url = COALESCE(sqlc.narg('image_url'), image_url),
--     image_width = COALESCE(sqlc.narg('image_width'), image_width),
--     image_height = COALESCE(sqlc.narg('image_height'), image_height),
--     updated_at = current_timestamp
-- WHERE id = sqlc.arg('id')
-- RETURNING *;
-- -- name: DeleteImage :exec
-- DELETE FROM images
-- WHERE id = $1;
-- -- name: DeleteImagesByArtwork :exec
-- DELETE FROM images
-- WHERE artwork_id = $1;
-- -- name: CountImagesByArtwork :one
-- SELECT COUNT(*)
-- FROM images
-- WHERE artwork_id = $1;
-- -- name: GetFirstImageByArtwork :one
-- SELECT *
-- FROM images
-- WHERE artwork_id = $1
-- ORDER BY created_at
-- LIMIT 1;