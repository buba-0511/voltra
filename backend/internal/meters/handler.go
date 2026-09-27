package meters

import (
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"energy-platform/internal/analytics"
	"energy-platform/internal/auth"
	sqlcgen "energy-platform/internal/db/sqlc"
	"energy-platform/internal/httpx"
	"errors"
)

// maxReadingsPageSize caps how many raw readings a single page can ever
// return, regardless of what the caller asks for - the endpoint must stay
// bounded even if the underlying reading frequency grows far past the
// current hourly cadence.
const maxReadingsPageSize = 5000

const defaultReadingsPageSize = 1000

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
	BaselineKwh    float64 `json:"baseline_kwh"`
	VariationPct   float64 `json:"variation_pct"`
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

type readingsPage struct {
	Readings   []readingDTO `json:"readings"`
	NextCursor *string      `json:"next_cursor"`
}

type dailyPointDTO struct {
	Day            string  `json:"day"`
	ConsumptionKwh float64 `json:"consumption_kwh"`
	AvgVoltageV    float64 `json:"avg_voltage_v"`
	AvgPowerFactor float64 `json:"avg_power_factor"`
}

type meterDailyRowDTO struct {
	MeterID        string  `json:"meter_id"`
	Day            string  `json:"day"`
	ConsumptionKwh float64 `json:"consumption_kwh"`
}

// List returns every meter's summary, including a baseline/variación
// projected over its whole reading period (see analytics.MeterPeriodBaseline)
// so every meter reports a variación, not just the ones with a detected
// anomaly (challenge brief section 6). This used to fetch each meter's
// readings one at a time (N+1 round trips, worse every time a meter is
// added); it now loads every reading for the tenant in a single query and
// groups them in memory, so the round-trip count stays O(1) regardless of
// how many meters exist.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	rows, err := h.Queries.ListMetersWithConsumption(r.Context(), claims.TenantID)
	if err != nil {
		httpx.WriteServerError(w, err, "failed to list meters")
		return
	}

	allReadings, err := h.Queries.ListReadingsByTenant(r.Context(), claims.TenantID)
	if err != nil {
		httpx.WriteServerError(w, err, "failed to load readings")
		return
	}
	readingsByMeter := groupReadingsByMeter(allReadings)

	out := make([]meterSummary, 0, len(rows))
	for _, row := range rows {
		baselineKwh, variationPct := periodBaseline(readingsByMeter[row.MeterID], row.TotalConsumptionKwh)
		out = append(out, meterSummary{
			MeterID:        row.MeterID,
			Name:           row.Name,
			Location:       row.Location,
			Status:         row.Status,
			ConsumptionKwh: row.TotalConsumptionKwh,
			BaselineKwh:    baselineKwh,
			VariationPct:   variationPct,
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
		httpx.WriteServerError(w, err, "failed to get meter")
		return
	}

	readings, err := h.Queries.ListReadingsByMeter(r.Context(), sqlcgen.ListReadingsByMeterParams{
		TenantID: claims.TenantID,
		MeterID:  meterID,
	})
	if err != nil {
		httpx.WriteServerError(w, err, "failed to compute baseline")
		return
	}
	baselineKwh, variationPct := periodBaseline(toAnalyticsReadings(readings), row.TotalConsumptionKwh)

	httpx.WriteJSON(w, http.StatusOK, meterDetail{
		meterSummary: meterSummary{
			MeterID:        row.MeterID,
			Name:           row.Name,
			Location:       row.Location,
			Status:         row.Status,
			ConsumptionKwh: row.TotalConsumptionKwh,
			BaselineKwh:    baselineKwh,
			VariationPct:   variationPct,
		},
		AvgVoltageV:    row.AvgVoltageV,
		AvgCurrentA:    row.AvgCurrentA,
		AvgPowerFactor: row.AvgPowerFactor,
	})
}

// ListReadings is cursor-paginated (keyset on timestamp): a meter's
// reading count has no upper bound, so this endpoint must never be able
// to return an unbounded response no matter how fine-grained the
// underlying data collection frequency becomes. Pass the previous page's
// `next_cursor` back as `?cursor=` to continue; a nil `next_cursor` means
// there's nothing left.
func (h *Handler) ListReadings(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	meterID := r.PathValue("meterId")

	limit := int32(defaultReadingsPageSize)
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			httpx.WriteError(w, http.StatusBadRequest, "invalid limit")
			return
		}
		limit = int32(parsed)
	}
	if limit > maxReadingsPageSize {
		limit = maxReadingsPageSize
	}

	var cursor pgtype.Timestamptz
	if raw := r.URL.Query().Get("cursor"); raw != "" {
		parsed, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "invalid cursor")
			return
		}
		cursor = pgtype.Timestamptz{Time: parsed, Valid: true}
	}

	rows, err := h.Queries.ListReadingsByMeterPage(r.Context(), sqlcgen.ListReadingsByMeterPageParams{
		TenantID: claims.TenantID,
		MeterID:  meterID,
		Limit:    limit,
		Cursor:   cursor,
	})
	if err != nil {
		httpx.WriteServerError(w, err, "failed to list readings")
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

	var nextCursor *string
	if int32(len(rows)) == limit {
		last := httpx.FormatTime(rows[len(rows)-1].Timestamp.Time)
		nextCursor = &last
	}

	httpx.WriteJSON(w, http.StatusOK, readingsPage{Readings: out, NextCursor: nextCursor})
}

// DailyConsumption aggregates a meter's readings into daily totals in
// Postgres (GROUP BY, not a client-side loop over raw rows) - the
// response size is bounded by the number of days in the period, not by
// how many readings exist within each day, so it stays cheap regardless
// of reading frequency. This is what the charts on MeterDetail and
// Investigación actually need, instead of the raw /readings endpoint.
func (h *Handler) DailyConsumption(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	meterID := r.PathValue("meterId")
	rows, err := h.Queries.DailyConsumptionByMeter(r.Context(), sqlcgen.DailyConsumptionByMeterParams{
		TenantID: claims.TenantID,
		MeterID:  meterID,
	})
	if err != nil {
		httpx.WriteServerError(w, err, "failed to load daily consumption")
		return
	}

	out := make([]dailyPointDTO, 0, len(rows))
	for _, row := range rows {
		out = append(out, dailyPointDTO{
			Day:            row.Day,
			ConsumptionKwh: row.TotalKwh,
			AvgVoltageV:    row.AvgVoltageV,
			AvgPowerFactor: row.AvgPowerFactor,
		})
	}

	httpx.WriteJSON(w, http.StatusOK, out)
}

// AllMetersDailyConsumption returns every meter's daily totals in one
// query, so the frontend's all-meters timeline no longer needs to fire
// one raw-readings request per meter (previously N parallel unbounded
// fetches on the Dashboard/Medidores pages - the biggest scalability risk
// in the app as meter count grows).
func (h *Handler) AllMetersDailyConsumption(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	rows, err := h.Queries.DailyConsumptionAllMeters(r.Context(), claims.TenantID)
	if err != nil {
		httpx.WriteServerError(w, err, "failed to load daily consumption")
		return
	}

	out := make([]meterDailyRowDTO, 0, len(rows))
	for _, row := range rows {
		out = append(out, meterDailyRowDTO{
			MeterID:        row.MeterID,
			Day:            row.Day,
			ConsumptionKwh: row.TotalKwh,
		})
	}

	httpx.WriteJSON(w, http.StatusOK, out)
}

func periodBaseline(readings []analytics.Reading, actualKwh float64) (baselineKwh, variationPct float64) {
	baselineKwh = analytics.MeterPeriodBaseline(readings)
	if baselineKwh > 0 {
		variationPct = (actualKwh - baselineKwh) / baselineKwh * 100
	}
	return baselineKwh, variationPct
}

func groupReadingsByMeter(rows []sqlcgen.Reading) map[string][]analytics.Reading {
	out := make(map[string][]analytics.Reading)
	for _, r := range rows {
		out[r.MeterID] = append(out[r.MeterID], analytics.Reading{
			Timestamp:      r.Timestamp.Time.UTC(),
			ConsumptionKWh: r.ConsumptionKwh,
			VoltageV:       r.VoltageV,
			CurrentA:       r.CurrentA,
			PowerFactor:    r.PowerFactor,
		})
	}
	return out
}

func toAnalyticsReadings(rows []sqlcgen.Reading) []analytics.Reading {
	out := make([]analytics.Reading, len(rows))
	for i, r := range rows {
		out[i] = analytics.Reading{
			Timestamp:      r.Timestamp.Time.UTC(),
			ConsumptionKWh: r.ConsumptionKwh,
			VoltageV:       r.VoltageV,
			CurrentA:       r.CurrentA,
			PowerFactor:    r.PowerFactor,
		}
	}
	return out
}
