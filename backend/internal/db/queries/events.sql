-- name: InsertEvent :exec
INSERT INTO events (tenant_id, meter_id, timestamp, type, description)
VALUES ($1, $2, $3, $4, $5);

-- name: ListEventsByMeter :many
SELECT * FROM events
WHERE tenant_id = $1 AND meter_id = $2
ORDER BY timestamp;
