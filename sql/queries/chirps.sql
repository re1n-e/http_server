-- name: CreateChirp :one
INSERT INTO chirp (body, user_id) 
VALUES (
    $1,
    $2
)
RETURNING *;

-- name: GetChirps :many
SELECT * FROM chirp ORDER BY created_at;

-- name: GetChirpById :one
SELECT * FROM chirp WHERE id = $1;

-- name: DeleteChirp :exec
DELETE FROM chirp WHERE id = $1;

-- name: GetChirpOwnerId :one
SELECT user_id FROM chirp WHERE id = $1;

-- name: GetChirpsByOwnerId :many
SELECT * FROM chirp WHERE user_id = $1 ORDER BY created_at ASC;