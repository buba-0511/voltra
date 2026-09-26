package analytics

import "time"

// Reading is analytics' own view of a meter reading, decoupled from the
// sqlc-generated DB types so this package has no dependency on Postgres
// and can be unit tested directly against the raw dataset.
type Reading struct {
	Timestamp      time.Time
	ConsumptionKWh float64
	VoltageV       float64
	CurrentA       float64
	PowerFactor    float64
}

type HourlyBaseline struct {
	Mean [24]float64
	Std  [24]float64
}

type MeterBaseline struct {
	Hourly      HourlyBaseline
	VoltageMean float64
	VoltageStd  float64
	PFMean      float64
	PFStd       float64
}

// OnsetPattern describes how abruptly a run's deviation set in, which
// distinguishes a "spike/cambio brusco" from a "cambio persistente" (the
// two categories the challenge brief calls out separately in its anomaly
// engine requirements).
type OnsetPattern string

const (
	OnsetStep    OnsetPattern = "STEP"    // reached its peak within the first couple of hours
	OnsetGradual OnsetPattern = "GRADUAL" // ramped up over several hours
)

// ConsumptionRun is a sustained block of readings that deviate from the
// meter's expected hourly baseline.
type ConsumptionRun struct {
	Start, End    time.Time
	AvgAbsZ       float64
	PeakAbsZ      float64
	ActualTotal   float64
	BaselineTotal float64
	VariationPct  float64
	VoltageStart  float64
	VoltageEnd    float64
	PFStart       float64
	PFEnd         float64

	// Ongoing is true when the run's window reaches all the way to the
	// last reading in the evaluated period - the anomaly hadn't resolved
	// itself by the time the data ends.
	Ongoing bool
	// OnsetHours is how many hours into the run its peak deviation was
	// reached; a small value means an abrupt step change.
	OnsetHours int
	Onset      OnsetPattern
}

type DataQualityFlag struct {
	Timestamp      time.Time
	ConsumptionKWh float64
	VoltageV       float64
	PF             float64
	VoltageZ       float64
	PFZ            float64
}

type MeterAnalysis struct {
	MeterID        string
	Baseline       MeterBaseline
	Runs           []ConsumptionRun
	DQFlags        []DataQualityFlag
	EvalSampleSize int
}
