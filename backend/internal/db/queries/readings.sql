-- name: InsertReading :exec
INSERT INTO readings (tenant_id, meter_id, timestamp, consumption_kwh, voltage_v, current_a, power_factor, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (tenant_id, meter_id, timestamp) DO NOTHING;

-- name: ListReadingsByMeter :many
SELECT * FROM readings
WHERE tenant_id = $1 AND meter_id = $2
ORDER BY timestamp;
