package meters

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

type meterSummary struct {
	MeterID        string  `json:"meter_id"`
	Name           string  `json:"name"`
	Location       string  `json:"location"`
	Status         string  `json:"status"`
	ConsumptionKwh float64 `json:"consumption_kwh"`
}

type meterDetail struct {
	meterSummary
	AvgVoltageV    float64 `json:"avg_voltage_v"`
	AvgCurrentA    float64 `json:"avg_current_a"`
	AvgPowerFactor float64 `json:"avg_power_factor"`
}

type readingDTO struct {
	Timestamp      string  `json:"timestamp"`
	ConsumptionKwh float64 `json:"consumption_kwh"`
	VoltageV       float64 `json:"voltage_v"`
	CurrentA       float64 `json:"current_a"`
	PowerFactor    float64 `json:"power_factor"`
	Status         string  `json:"status"`
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	rows, err := h.Queries.ListMetersWithConsumption(r.Context(), claims.TenantID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to list meters")
		return
	}

	out := make([]meterSummary, 0, len(rows))
	for _, row := range rows {
		out = append(out, meterSummary{
			MeterID:        row.MeterID,
			Name:           row.Name,
			Location:       row.Location,
			Status:         row.Status,
			ConsumptionKwh: row.TotalConsumptionKwh,
		})
	}

	httpx.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	meterID := r.PathValue("meterId")
	row, err := h.Queries.GetMeterDetail(r.Context(), sqlcgen.GetMeterDetailParams{
		TenantID: claims.TenantID,
		MeterID:  meterID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteError(w, http.StatusNotFound, "meter not found")
		return
	}
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to get meter")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, meterDetail{
		meterSummary: meterSummary{
			MeterID:        row.MeterID,
			Name:           row.Name,
			Location:       row.Location,
			Status:         row.Status,
			ConsumptionKwh: row.TotalConsumptionKwh,
		},
		AvgVoltageV:    row.AvgVoltageV,
		AvgCurrentA:    row.AvgCurrentA,
		AvgPowerFactor: row.AvgPowerFactor,
	})
}

func (h *Handler) ListReadings(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	meterID := r.PathValue("meterId")
	rows, err := h.Queries.ListReadingsByMeter(r.Context(), sqlcgen.ListReadingsByMeterParams{
		TenantID: claims.TenantID,
		MeterID:  meterID,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to list readings")
		return
	}

	out := make([]readingDTO, 0, len(rows))
	for _, row := range rows {
		out = append(out, readingDTO{
			Timestamp:      httpx.FormatTime(row.Timestamp.Time),
			ConsumptionKwh: row.ConsumptionKwh,
			VoltageV:       row.VoltageV,
			CurrentA:       row.CurrentA,
			PowerFactor:    row.PowerFactor,
			Status:         row.Status,
		})
	}

	httpx.WriteJSON(w, http.StatusOK, out)
}
