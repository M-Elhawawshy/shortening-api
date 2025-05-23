-- name: RevokeJwt :one
INSERT INTO revoked_tokens (jti, user_id, expires_at)
VALUES ($1, $2, $3)
returning *;

-- name: IsTokenRevoked :one
SELECT EXISTS(
  SELECT TRUE FROM revoked_tokens WHERE jti = $1
);