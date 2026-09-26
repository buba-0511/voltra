package dashboard

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"

	"energy-platform/internal/auth"
	sqlcgen "energy-platform/internal/db/sqlc"
	"energy-platform/internal/httpx"
)

type Handler struct {
	Queries *sqlcgen.Queries
}

func NewHandler(q *sqlcgen.Queries) *Handler {
	return &Handler{Queries: q}
}

type lastAnalysis struct {
	Status     string  `json:"status"`
	Stage      string  `json:"stage"`
	StartedAt  string  `json:"started_at"`
	FinishedAt *string `json:"finished_at"`
}

type summary struct {
	MetersCount         int64         `json:"meters_count"`
	TotalConsumptionKwh float64       `json:"total_consumption_kwh"`
	AnomaliesDetected   int64         `json:"anomalies_detected"`
	HighPriority        int64         `json:"high_priority"`
	LastAnalysis        *lastAnalysis `json:"last_analysis"`
}

func (h *Handler) Summary(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	ctx := r.Context()

	metersCount, err := h.Queries.DashboardMeterCount(ctx, claims.TenantID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to count meters")
		return
	}

	totalConsumption, err := h.Queries.DashboardTotalConsumption(ctx, claims.TenantID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to sum consumption")
		return
	}

	anomalyCounts, err := h.Queries.DashboardAnomalyCounts(ctx, claims.TenantID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to count anomalies")
		return
	}

	var last *lastAnalysis
	run, err := h.Queries.DashboardLatestAnalysisRun(ctx, claims.TenantID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to load last analysis")
		return
	}
	if err == nil {
		last = &lastAnalysis{
			Status:    run.Status,
			Stage:     run.Stage,
			StartedAt: httpx.FormatTime(run.StartedAt.Time),
		}
		if run.FinishedAt.Valid {
			finished := httpx.FormatTime(run.FinishedAt.Time)
			last.FinishedAt = &finished
		}
	}

	httpx.WriteJSON(w, http.StatusOK, summary{
		MetersCount:         metersCount,
		TotalConsumptionKwh: totalConsumption,
		AnomaliesDetected:   anomalyCounts.Total,
		HighPriority:        anomalyCounts.HighPriority,
		LastAnalysis:        last,
	})
}
