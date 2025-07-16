-- name: CreateChrip :one
INSERT INTO chirps (id, body, created_at, updated_at, user_id)
VALUES (gen_random_uuid(), $1, NOW(), NOW(), $2)
RETURNING *;

-- name: GetAll :many
SELECT id, body, created_at, updated_at, user_id
FROM chirps
ORDER BY created_at ASC;

-- name: GetChirpByID :one
SELECT id, body, created_at, updated_at, user_id
FROM chirps
WHERE id = $1;

-- name: DeleteChirpByID :exec
DELETE FROM chirps
WHERE id = $1;