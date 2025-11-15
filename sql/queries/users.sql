-- name: CreateUser :one
INSERT INTO users (email, hashed_password)
VALUES (
    $1,
    $2
)
RETURNING *;

-- name: ResetUsers :exec
TRUNCATE users CASCADE;

-- name: GetUserByMail :one
SELECT * FROM users WHERE email = $1;

-- name: UpdateUser :one
UPDATE users
SET email = $1, hashed_password = $2, updated_at = NOW()
WHERE id = $3
RETURNING *;

-- name: UpdateChirpyRed :exec
UPDATE users
SET is_chirpy_red = true
WHERE id = $1;