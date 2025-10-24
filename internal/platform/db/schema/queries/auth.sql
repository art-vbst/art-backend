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
INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
VALUES ($1, $2, $3)
RETURNING id,
    user_id,
    token_hash,
    created_at,
    expires_at,
    revoked;

-- name: GetRefreshTokenByHash :one
SELECT id,
    user_id,
    token_hash,
    created_at,
    expires_at,
    revoked
FROM refresh_tokens
WHERE token_hash = $1
    AND revoked = FALSE
    AND expires_at > now();

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked = TRUE
WHERE token_hash = $1;

-- name: RevokeAllUserRefreshTokens :exec
UPDATE refresh_tokens
SET revoked = TRUE
WHERE user_id = $1;

-- name: DeleteExpiredRefreshTokens :exec
DELETE FROM refresh_tokens
WHERE expires_at < now()
    OR revoked = TRUE;