-- name: CreateUser :one
INSERT INTO users(id, created_at, updated_at, email, hashed_password)
VALUES (gen_random_uuid(), NOW(), NOW(), $1, $2)
RETURNING *;

-- name: DeleteUsers :exec
DELETE FROM users;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1;

-- name: UpdateUser :one
UPDATE users
SET hashed_password = $2, email = $3, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpgradeUserRedChirp :one
UPDATE users
SET is_chirpy_red = true WHERE id = $1
RETURNING *;
