package anomalies

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"

	"energy-platform/internal/auth"
	sqlcgen "energy-platform/internal/db/sqlc"
	"energy-platform/internal/httpx"
)

type Handler struct {
	Service *Service
	Queries *sqlcgen.Queries
}

func NewHandler(service *Service, q *sqlcgen.Queries) *Handler {
	return &Handler{Service: service, Queries: q}
}

type runDTO struct {
	ID             int64   `json:"id"`
	Status         string  `json:"status"`
	Stage          string  `json:"stage"`
	StartedAt      string  `json:"started_at"`
	FinishedAt     *string `json:"finished_at"`
	MetersAnalyzed int32   `json:"meters_analyzed"`
	AnomaliesFound int32   `json:"anomalies_found"`
	HighPriority   int32   `json:"high_priority"`
}

func toRunDTO(run sqlcgen.AnalysisRun) runDTO {
	dto := runDTO{
		ID:             run.ID,
		Status:         run.Status,
		Stage:          run.Stage,
		StartedAt:      httpx.FormatTime(run.StartedAt.Time),
		MetersAnalyzed: run.MetersAnalyzed,
		AnomaliesFound: run.AnomaliesFound,
		HighPriority:   run.HighPriority,
	}
	if run.FinishedAt.Valid {
		finished := httpx.FormatTime(run.FinishedAt.Time)
		dto.FinishedAt = &finished
	}
	return dto
}

// Analyze runs the full pipeline synchronously and returns the finished
// run's summary. For this dataset's size (12 meters, at most a handful of
// anomalies) it completes in a few seconds - not worth building polling
// infrastructure for.
func (h *Handler) Analyze(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	run, err := h.Service.RunAnalysis(r.Context(), claims.TenantID)
	if err != nil {
		httpx.WriteServerError(w, err, "analysis failed")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toRunDTO(run))
}

func (h *Handler) GetAnalysisRun(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	run, err := h.Queries.GetAnalysisRun(r.Context(), sqlcgen.GetAnalysisRunParams{
		TenantID: claims.TenantID, ID: id,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteError(w, http.StatusNotFound, "analysis run not found")
		return
	}
	if err != nil {
		httpx.WriteServerError(w, err, "failed to load analysis run")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toRunDTO(run))
}

type anomalyDTO struct {
	ID                int64          `json:"id"`
	MeterID           string         `json:"meter_id"`
	DetectedAt        string         `json:"detected_at"`
	WindowStart       string         `json:"window_start"`
	WindowEnd         string         `json:"window_end"`
	Type              string         `json:"type"`
	Severity          string         `json:"severity"`
	Confidence        float64        `json:"confidence"`
	Reason            string         `json:"reason"`
	RecommendedAction string         `json:"recommended_action"`
	Status            string         `json:"status"`
	BaselineKWh       float64        `json:"baseline_kwh"`
	ActualKWh         float64        `json:"actual_kwh"`
	VariationPct      float64        `json:"variation_pct"`
	Evidence          map[string]any `json:"evidence"`
}

func toAnomalyDTO(a sqlcgen.Anomaly) anomalyDTO {
	var evidence map[string]any
	_ = json.Unmarshal(a.Evidence, &evidence)
	return anomalyDTO{
		ID:                a.ID,
		MeterID:           a.MeterID,
		DetectedAt:        httpx.FormatTime(a.DetectedAt.Time),
		WindowStart:       httpx.FormatTime(a.WindowStart.Time),
		WindowEnd:         httpx.FormatTime(a.WindowEnd.Time),
		Type:              a.Type,
		Severity:          a.Severity,
		Confidence:        a.Confidence,
		Reason:            a.Reason,
		RecommendedAction: a.RecommendedAction,
		Status:            a.Status,
		BaselineKWh:       a.BaselineKwh,
		ActualKWh:         a.ActualKwh,
		VariationPct:      a.VariationPct,
		Evidence:          evidence,
	}
}

// List returns the anomalies from the most recent completed analysis run
// only - re-running "Run AI Analysis" replaces what's shown, it doesn't
// pile up duplicates from every previous click.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	rows, err := h.Queries.ListAnomaliesForLatestRun(r.Context(), claims.TenantID)
	if err != nil {
		httpx.WriteServerError(w, err, "failed to list anomalies")
		return
	}

	out := make([]anomalyDTO, 0, len(rows))
	for _, a := range rows {
		out = append(out, toAnomalyDTO(a))
	}
	httpx.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	a, err := h.Queries.GetAnomalyByID(r.Context(), sqlcgen.GetAnomalyByIDParams{
		TenantID: claims.TenantID, ID: id,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteError(w, http.StatusNotFound, "anomaly not found")
		return
	}
	if err != nil {
		httpx.WriteServerError(w, err, "failed to load anomaly")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toAnomalyDTO(a))
}

var validAnomalyStatuses = map[string]bool{
	"OPEN": true, "IN_REVIEW": true, "RESOLVED": true, "DISMISSED": true,
}

type updateStatusRequest struct {
	Status string `json:"status"`
}

func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req updateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || !validAnomalyStatuses[req.Status] {
		httpx.WriteError(w, http.StatusBadRequest, "invalid status")
		return
	}

	a, err := h.Queries.UpdateAnomalyStatus(r.Context(), sqlcgen.UpdateAnomalyStatusParams{
		TenantID: claims.TenantID, ID: id, Status: req.Status,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteError(w, http.StatusNotFound, "anomaly not found")
		return
	}
	if err != nil {
		httpx.WriteServerError(w, err, "failed to update anomaly")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toAnomalyDTO(a))
}
