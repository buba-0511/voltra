-- name: ListReadingsByTenant :many
-- Used to compute every meter's baseline in one round trip instead of one
-- query per meter (see meters.Handler.List) - still O(readings) in Go,
-- but O(1) in database round trips regardless of meter count.
SELECT * FROM readings
WHERE tenant_id = $1
ORDER BY meter_id, timestamp;

-- name: InsertReading :exec
INSERT INTO readings (tenant_id, meter_id, timestamp, consumption_kwh, voltage_v, current_a, power_factor, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (tenant_id, meter_id, timestamp) DO NOTHING;

-- name: ListReadingsByMeter :many
SELECT * FROM readings
WHERE tenant_id = $1 AND meter_id = $2
ORDER BY timestamp;

-- name: ListReadingsByMeterPage :many
-- Cursor (keyset) pagination on timestamp: a meter's readings have no
-- upper bound on how many rows they can grow to, so this endpoint must
-- never be able to return an unbounded response.
SELECT * FROM readings
WHERE tenant_id = $1 AND meter_id = $2
  AND (sqlc.narg('cursor')::timestamptz IS NULL OR timestamp > sqlc.narg('cursor'))
ORDER BY timestamp
LIMIT $3;

-- name: DailyConsumptionByMeter :many
-- Aggregates in Postgres instead of pulling every raw reading to the app
-- to sum client-side - the response stays bounded by day count, not by
-- reading count, no matter how fine-grained the underlying data gets.
SELECT
    to_char(timestamp AT TIME ZONE 'UTC', 'YYYY-MM-DD') AS day,
    COALESCE(SUM(consumption_kwh), 0)::float8 AS total_kwh,
    COALESCE(AVG(voltage_v), 0)::float8 AS avg_voltage_v,
    COALESCE(AVG(power_factor), 0)::float8 AS avg_power_factor
FROM readings
WHERE tenant_id = $1 AND meter_id = $2
GROUP BY day
ORDER BY day;

-- name: DailyConsumptionAllMeters :many
SELECT
    meter_id,
    to_char(timestamp AT TIME ZONE 'UTC', 'YYYY-MM-DD') AS day,
    COALESCE(SUM(consumption_kwh), 0)::float8 AS total_kwh
FROM readings
WHERE tenant_id = $1
GROUP BY meter_id, day
ORDER BY meter_id, day;
