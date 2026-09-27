-- DashboardLatestAnalysisRun and the "latest DONE run" subquery in
-- ListAnomaliesForLatestRun/DashboardAnomalyCounts both filter+sort by
-- exactly these columns, and analysis_runs grows by one row every time
-- "Run AI Analysis" is clicked - without this index that scan gets
-- slower as history accumulates, on some of the most-hit endpoints.
CREATE INDEX IF NOT EXISTS idx_analysis_runs_tenant_status_started
    ON analysis_runs (tenant_id, status, started_at DESC);
