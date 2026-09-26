-- name: InsertAnomaly :one
INSERT INTO anomalies (
    tenant_id, analysis_run_id, meter_id, window_start, window_end,
    type, severity, confidence, reason, recommended_action, status,
    baseline_kwh, actual_kwh, variation_pct, evidence
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 'OPEN', $11, $12, $13, $14
)
RETURNING *;

-- name: ListAnomaliesForLatestRun :many
SELECT a.* FROM anomalies a
WHERE a.tenant_id = $1
  AND a.analysis_run_id = (
      SELECT id FROM analysis_runs
      WHERE tenant_id = $1 AND status = 'DONE'
      ORDER BY started_at DESC
      LIMIT 1
  )
ORDER BY CASE a.severity WHEN 'HIGH' THEN 0 WHEN 'MEDIUM' THEN 1 ELSE 2 END, a.confidence DESC;

-- name: GetAnomalyByID :one
SELECT * FROM anomalies WHERE tenant_id = $1 AND id = $2;
