package ai

import (
	"encoding/csv"
	"os"
	"sort"
	"strconv"
	"testing"
	"time"

	"energy-platform/internal/analytics"
)

const (
	readingsCSVPath = "../../../data/readings.csv"
	eventsCSVPath   = "../../../data/events.csv"
)

func loadReadingsByMeter(t *testing.T) map[string][]analytics.Reading {
	t.Helper()

	f, err := os.Open(readingsCSVPath)
	if err != nil {
		t.Fatalf("open readings.csv: %v", err)
	}
	defer f.Close()

	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatalf("read readings.csv: %v", err)
	}

	byMeter := map[string][]analytics.Reading{}
	for _, row := range rows[1:] {
		ts, err := time.Parse("2006-01-02 15:04:05", row[1])
		if err != nil {
			t.Fatalf("parse timestamp %q: %v", row[1], err)
		}
		consumption, _ := strconv.ParseFloat(row[2], 64)
		voltage, _ := strconv.ParseFloat(row[3], 64)
		current, _ := strconv.ParseFloat(row[4], 64)
		pf, _ := strconv.ParseFloat(row[5], 64)

		byMeter[row[0]] = append(byMeter[row[0]], analytics.Reading{
			Timestamp:      ts,
			ConsumptionKWh: consumption,
			VoltageV:       voltage,
			CurrentA:       current,
			PowerFactor:    pf,
		})
	}
	for meterID, readings := range byMeter {
		sort.Slice(readings, func(i, j int) bool { return readings[i].Timestamp.Before(readings[j].Timestamp) })
		byMeter[meterID] = readings
	}
	return byMeter
}

func loadEventsByMeter(t *testing.T) map[string][]Event {
	t.Helper()

	f, err := os.Open(eventsCSVPath)
	if err != nil {
		t.Fatalf("open events.csv: %v", err)
	}
	defer f.Close()

	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatalf("read events.csv: %v", err)
	}

	byMeter := map[string][]Event{}
	for _, row := range rows[1:] {
		ts, err := time.Parse("2006-01-02 15:04:05", row[1])
		if err != nil {
			t.Fatalf("parse event timestamp %q: %v", row[1], err)
		}
		byMeter[row[0]] = append(byMeter[row[0]], Event{
			Timestamp:   ts,
			Type:        row[2],
			Description: row[3],
		})
	}
	return byMeter
}

// classifyMeter is the small end-to-end pipeline this test exercises:
// analytics.Analyze -> Classify, exactly what piece 6 will do per meter.
func classifyMeter(t *testing.T, meterID string) []Classification {
	t.Helper()
	readings := loadReadingsByMeter(t)
	events := loadEventsByMeter(t)
	analysis := analytics.Analyze(meterID, readings[meterID])
	return Classify(meterID, analysis, events[meterID])
}

func TestClassify_M109_RealAnomalyHigh(t *testing.T) {
	results := classifyMeter(t, "M-109")
	if len(results) != 1 {
		t.Fatalf("expected exactly 1 classification for M-109, got %d: %+v", len(results), results)
	}
	c := results[0]
	if c.Type != TypeRealAnomaly {
		t.Errorf("expected REAL_ANOMALY for M-109, got %s", c.Type)
	}
	if c.Severity != SeverityHigh {
		t.Errorf("expected HIGH severity for M-109, got %s", c.Severity)
	}
	if c.Confidence < 0.8 {
		t.Errorf("expected high confidence for M-109's obvious spike, got %.2f", c.Confidence)
	}
}

func TestClassify_M106_FalsePositiveLow(t *testing.T) {
	results := classifyMeter(t, "M-106")
	if len(results) != 1 {
		t.Fatalf("expected exactly 1 classification for M-106, got %d: %+v", len(results), results)
	}
	c := results[0]
	if c.Type != TypeFalsePositive {
		t.Errorf("expected FALSE_POSITIVE for M-106 (matches its SCHEDULED_OUTAGE event), got %s", c.Type)
	}
	if c.Severity != SeverityLow {
		t.Errorf("expected LOW severity for M-106, got %s", c.Severity)
	}
}

func TestClassify_M104_ExplainableMedium(t *testing.T) {
	results := classifyMeter(t, "M-104")
	if len(results) != 1 {
		t.Fatalf("expected exactly 1 classification for M-104, got %d: %+v", len(results), results)
	}
	c := results[0]
	if c.Type != TypeExplainable {
		t.Errorf("expected EXPLAINABLE_ANOMALY for M-104 (matches its OPERATIONAL_CHANGE event), got %s", c.Type)
	}
	if c.Severity != SeverityMedium {
		t.Errorf("expected MEDIUM severity for M-104, got %s", c.Severity)
	}
}

func TestClassify_M112_DataQualityHigh(t *testing.T) {
	results := classifyMeter(t, "M-112")
	if len(results) != 1 {
		t.Fatalf("expected exactly 1 classification for M-112, got %d: %+v", len(results), results)
	}
	c := results[0]
	if c.Type != TypeDataQuality {
		t.Errorf("expected DATA_QUALITY for M-112, got %s", c.Type)
	}
	if c.Severity != SeverityHigh {
		t.Errorf("expected HIGH severity for M-112, got %s", c.Severity)
	}
}

func TestClassify_QuietMetersProduceNothing(t *testing.T) {
	quietMeters := []string{"M-101", "M-102", "M-103", "M-105", "M-107", "M-108", "M-110", "M-111"}
	for _, meterID := range quietMeters {
		results := classifyMeter(t, meterID)
		if len(results) != 0 {
			t.Errorf("expected no classifications for quiet meter %s, got %+v", meterID, results)
		}
	}
}
