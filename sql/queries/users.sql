-- name: GetUser :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (id, email, password_hash)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateUserURLCounter :one
UPDATE users
SET total_url_shortened = total_url_shortened + 1
WHERE id = $1
RETURNING *;
