package ai

import (
	"time"

	"energy-platform/internal/analytics"
)

type AnomalyType string

const (
	TypeRealAnomaly   AnomalyType = "REAL_ANOMALY"
	TypeExplainable   AnomalyType = "EXPLAINABLE_ANOMALY"
	TypeFalsePositive AnomalyType = "FALSE_POSITIVE"
	TypeDataQuality   AnomalyType = "DATA_QUALITY"
)

type Severity string

const (
	SeverityHigh   Severity = "HIGH"
	SeverityMedium Severity = "MEDIUM"
	SeverityLow    Severity = "LOW"
)

// Event mirrors an events.csv row for a single meter - kept independent of
// the DB layer so this package stays unit-testable without Postgres.
type Event struct {
	Timestamp   time.Time `json:"timestamp"`
	Type        string    `json:"type"` // OPERATIONAL_CHANGE | SCHEDULED_OUTAGE | DATA_QUALITY | UNKNOWN
	Description string    `json:"description"`
}

// Evidence carries the structured facts a classification is based on -
// this is what the explanation layer (piece 5) turns into prose, and what
// the Investigation screen renders directly.
type Evidence struct {
	PeakZScore             float64 `json:"peak_zscore"`
	AvgZScore              float64 `json:"avg_zscore"`
	VoltageBaselineV       float64 `json:"voltage_baseline_v"`
	VoltageActualStartV    float64 `json:"voltage_actual_start_v"`
	VoltageActualEndV      float64 `json:"voltage_actual_end_v"`
	PowerFactorBaseline    float64 `json:"power_factor_baseline"`
	PowerFactorActualStart float64 `json:"power_factor_actual_start"`
	PowerFactorActualEnd   float64 `json:"power_factor_actual_end"`
	DataQualityFlagCount   int     `json:"data_quality_flag_count"`
	DataQualitySampleSize  int     `json:"data_quality_sample_size"`
	RelatedEvent           *Event  `json:"related_event,omitempty"`

	// Ongoing/Onset describe the temporal shape of the deviation - still
	// happening vs already resolved, abrupt step vs gradual ramp. Not set
	// for data-quality classifications (those aren't a single run).
	Ongoing    bool                   `json:"ongoing"`
	OnsetHours int                    `json:"onset_hours"`
	Onset      analytics.OnsetPattern `json:"onset,omitempty"`
}

type Classification struct {
	MeterID      string
	Type         AnomalyType
	Severity     Severity
	Confidence   float64
	WindowStart  time.Time
	WindowEnd    time.Time
	BaselineKWh  float64
	ActualKWh    float64
	VariationPct float64
	Evidence     Evidence
}
