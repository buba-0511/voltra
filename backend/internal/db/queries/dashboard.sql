-- name: DashboardMeterCount :one
SELECT count(*) FROM meters WHERE tenant_id = $1;

-- name: DashboardTotalConsumption :one
SELECT COALESCE(SUM(consumption_kwh), 0)::float8 FROM readings WHERE tenant_id = $1;

-- name: DashboardAnomalyCounts :one
SELECT
    count(*) AS total,
    count(*) FILTER (WHERE severity = 'HIGH') AS high_priority
FROM anomalies
WHERE tenant_id = $1;

-- name: DashboardLatestAnalysisRun :one
SELECT * FROM analysis_runs
WHERE tenant_id = $1
ORDER BY started_at DESC
LIMIT 1;
