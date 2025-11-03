-- name: GetUserByID :one
SELECT id,
    email,
    password_hash,
    created_at
FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT id,
    email,
    password_hash,
    created_at
FROM users
WHERE email = $1;

-- name: CreateUser :one
INSERT INTO users (email, password_hash)
VALUES ($1, $2)
RETURNING id,
    email,
    password_hash,
    created_at;

-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (user_id, token_hash, jti, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetRefreshTokenByJTI :one
SELECT *
FROM refresh_tokens
WHERE jti = $1
    AND revoked = FALSE;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked = TRUE
WHERE id = $1;

-- name: RevokeAllUserRefreshTokens :exec
UPDATE refresh_tokens
SET revoked = TRUE
WHERE user_id = $1;

-- name: DeleteExpiredRefreshTokens :exec
DELETE FROM refresh_tokens
WHERE expires_at < now()
    OR revoked = TRUE;