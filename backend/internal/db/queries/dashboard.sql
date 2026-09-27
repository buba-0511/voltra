-- name: DashboardMeterCount :one
SELECT count(*) FROM meters WHERE tenant_id = $1;

-- name: DashboardTotalConsumption :one
SELECT COALESCE(SUM(consumption_kwh), 0)::float8 FROM readings WHERE tenant_id = $1;

-- name: DashboardAnomalyCounts :one
-- Scoped to the latest completed run, same as ListAnomaliesForLatestRun -
-- otherwise this grows every time "Run AI Analysis" is clicked instead of
-- reflecting the current state.
SELECT
    count(*) AS total,
    count(*) FILTER (WHERE anomalies.severity = 'HIGH') AS high_priority
FROM anomalies
WHERE anomalies.tenant_id = $1
  AND anomalies.analysis_run_id = (
      SELECT ar.id FROM analysis_runs ar
      WHERE ar.tenant_id = $1 AND ar.status = 'DONE'
      ORDER BY ar.started_at DESC
      LIMIT 1
  );

-- name: DashboardLatestAnalysisRun :one
SELECT * FROM analysis_runs
WHERE tenant_id = $1
ORDER BY started_at DESC
LIMIT 1;
