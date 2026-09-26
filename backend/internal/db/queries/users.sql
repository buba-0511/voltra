-- name: GetUserByEmail :one
SELECT * FROM users WHERE tenant_id = $1 AND email = $2;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: CreateUser :one
INSERT INTO users (tenant_id, email, password_hash, name)
VALUES ($1, $2, $3, $4)
ON CONFLICT (tenant_id, email) DO UPDATE SET password_hash = EXCLUDED.password_hash
RETURNING *;
