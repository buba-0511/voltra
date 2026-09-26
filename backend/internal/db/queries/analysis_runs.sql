-- name: CreateAnalysisRun :one
INSERT INTO analysis_runs (tenant_id, status, stage)
VALUES ($1, 'RUNNING', 'READINGS')
RETURNING *;

-- name: CompleteAnalysisRun :one
UPDATE analysis_runs
SET status = $2,
    stage = $3,
    finished_at = now(),
    meters_analyzed = $4,
    anomalies_found = $5,
    high_priority = $6
WHERE id = $1
RETURNING *;

-- name: GetAnalysisRun :one
SELECT * FROM analysis_runs WHERE tenant_id = $1 AND id = $2;
