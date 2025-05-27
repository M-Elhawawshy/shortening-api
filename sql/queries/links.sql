-- name: InsertLink :one
INSERT INTO links(hash, user_id, link)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetLink :one
SELECT * FROM links
WHERE hash = $1 LIMIT 1;