-- name: UpsertMeter :one
INSERT INTO meters (tenant_id, meter_id, name, location, status)
VALUES ($1, $2, $3, $4, 'OK')
ON CONFLICT (tenant_id, meter_id) DO UPDATE SET name = EXCLUDED.name
RETURNING *;

-- name: ListMeters :many
SELECT * FROM meters WHERE tenant_id = $1 ORDER BY meter_id;

-- name: GetMeterByMeterID :one
SELECT * FROM meters WHERE tenant_id = $1 AND meter_id = $2;

-- name: CountReadings :one
SELECT count(*) FROM readings WHERE tenant_id = $1;
