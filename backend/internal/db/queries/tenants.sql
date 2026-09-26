-- name: GetTenantBySlug :one
SELECT * FROM tenants WHERE slug = $1;

-- name: CreateTenant :one
INSERT INTO tenants (slug, name)
VALUES ($1, $2)
ON CONFLICT (slug) DO UPDATE SET slug = EXCLUDED.slug
RETURNING *;
