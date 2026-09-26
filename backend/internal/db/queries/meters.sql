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

-- name: ListMetersWithConsumption :many
SELECT
    m.meter_id,
    m.name,
    m.location,
    m.status,
    COALESCE(SUM(r.consumption_kwh), 0)::float8 AS total_consumption_kwh
FROM meters m
LEFT JOIN readings r ON r.tenant_id = m.tenant_id AND r.meter_id = m.meter_id
WHERE m.tenant_id = $1
GROUP BY m.meter_id, m.name, m.location, m.status
ORDER BY m.meter_id;

-- name: GetMeterDetail :one
SELECT
    m.meter_id,
    m.name,
    m.location,
    m.status,
    COALESCE(SUM(r.consumption_kwh), 0)::float8 AS total_consumption_kwh,
    COALESCE(AVG(r.voltage_v), 0)::float8 AS avg_voltage_v,
    COALESCE(AVG(r.current_a), 0)::float8 AS avg_current_a,
    COALESCE(AVG(r.power_factor), 0)::float8 AS avg_power_factor
FROM meters m
LEFT JOIN readings r ON r.tenant_id = m.tenant_id AND r.meter_id = m.meter_id
WHERE m.tenant_id = $1 AND m.meter_id = $2
GROUP BY m.meter_id, m.name, m.location, m.status;
