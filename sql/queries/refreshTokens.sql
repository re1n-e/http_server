-- name: CreateRefreshToken :exec
INSERT INTO refresh_tokens (token, user_id, expires_at, revoked_at)
VALUES ($1, $2, $3, $4);

-- name: GetRefreshTokenByRefToken :one
SELECT *
FROM refresh_tokens
WHERE token = $1;

-- name: GetRefreshToken :one
SELECT token 
FROM refresh_tokens 
WHERE user_id = $1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET updated_at = NOW(), revoked_at = NOW()
WHERE token = $1
RETURNING *;
