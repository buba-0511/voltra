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
	Timestamp   time.Time
	Type        string // OPERATIONAL_CHANGE | SCHEDULED_OUTAGE | DATA_QUALITY | UNKNOWN
	Description string
}

// Evidence carries the structured facts a classification is based on -
// this is what the explanation layer (piece 5) turns into prose, and what
// the Investigation screen renders directly.
type Evidence struct {
	PeakZScore             float64
	AvgZScore              float64
	VoltageBaselineV       float64
	VoltageActualStartV    float64
	VoltageActualEndV      float64
	PowerFactorBaseline    float64
	PowerFactorActualStart float64
	PowerFactorActualEnd   float64
	DataQualityFlagCount   int
	DataQualitySampleSize  int
	RelatedEvent           *Event

	// Ongoing/Onset describe the temporal shape of the deviation - still
	// happening vs already resolved, abrupt step vs gradual ramp. Not set
	// for data-quality classifications (those aren't a single run).
	Ongoing    bool
	OnsetHours int
	Onset      analytics.OnsetPattern
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
