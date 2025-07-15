-- name: CreateChrip :one
INSERT INTO chirps (id, body, created_at, updated_at, user_id)
VALUES (gen_random_uuid(), $1, NOW(), NOW(), $2)
RETURNING *;