-- name: CreateChrip :one
INSERT INTO chirps (id, body, created_at, updated_at, user_id)
VALUES (gen_random_uuid(), $1, NOW(), NOW(), $2)
RETURNING *;

-- name: GetAllAsc :many
SELECT id, body, created_at, updated_at, user_id
FROM chirps
WHERE $1::uuid = '00000000-0000-0000-0000-000000000000' OR user_id = $1
ORDER BY created_at asc;

-- name: GetAllDesc :many
SELECT id, body, created_at, updated_at, user_id
FROM chirps
WHERE $1::uuid = '00000000-0000-0000-0000-000000000000' OR user_id = $1
ORDER BY created_at desc;

-- name: GetChirpByID :one
SELECT id, body, created_at, updated_at, user_id
FROM chirps
WHERE id = $1;

-- name: DeleteChirpByID :exec
DELETE FROM chirps
WHERE id = $1;