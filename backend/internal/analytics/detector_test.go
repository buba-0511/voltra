package analytics

import (
	"encoding/csv"
	"os"
	"sort"
	"strconv"
	"testing"
	"time"
)

const readingsCSVPath = "../../../data/readings.csv"

func loadReadingsByMeter(t *testing.T) map[string][]Reading {
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

	byMeter := map[string][]Reading{}
	for _, row := range rows[1:] {
		ts, err := time.Parse("2006-01-02 15:04:05", row[1])
		if err != nil {
			t.Fatalf("parse timestamp %q: %v", row[1], err)
		}
		consumption, _ := strconv.ParseFloat(row[2], 64)
		voltage, _ := strconv.ParseFloat(row[3], 64)
		current, _ := strconv.ParseFloat(row[4], 64)
		pf, _ := strconv.ParseFloat(row[5], 64)

		byMeter[row[0]] = append(byMeter[row[0]], Reading{
			Timestamp:      ts,
			ConsumptionKWh: consumption,
			VoltageV:       voltage,
			CurrentA:       current,
			PowerFactor:    pf,
		})
	}

	for meterID, readings := range byMeter {
		sort.Slice(readings, func(i, j int) bool {
			return readings[i].Timestamp.Before(readings[j].Timestamp)
		})
		byMeter[meterID] = readings
	}

	return byMeter
}

func totalAbsVariation(runs []ConsumptionRun) float64 {
	max := 0.0
	for _, r := range runs {
		v := abs(r.VariationPct)
		if v > max {
			max = v
		}
	}
	return max
}

func TestAnalyze_M109_RealAnomaly(t *testing.T) {
	byMeter := loadReadingsByMeter(t)
	analysis := Analyze("M-109", byMeter["M-109"])

	if len(analysis.Runs) == 0 {
		t.Fatalf("expected at least one consumption run for M-109, got none")
	}
	peak := totalAbsVariation(analysis.Runs)
	if peak < 50 {
		t.Errorf("expected M-109's peak variation to be well above 50%%, got %.1f%%", peak)
	}

	var maxZ float64
	for _, r := range analysis.Runs {
		if r.PeakAbsZ > maxZ {
			maxZ = r.PeakAbsZ
		}
	}
	if maxZ < 5 {
		t.Errorf("expected a strong z-score signal (>5) for M-109's spike, got %.1f", maxZ)
	}

	// M-109's spike starts abruptly (jumps in a single hour) and never
	// recovers before the dataset ends.
	main := analysis.Runs[len(analysis.Runs)-1]
	if !main.Ongoing {
		t.Errorf("expected M-109's spike to still be ongoing at the end of the dataset")
	}
	if main.Onset != OnsetStep {
		t.Errorf("expected M-109's spike to be a STEP onset, got %s (onset hour %d)", main.Onset, main.OnsetHours)
	}
}

func TestAnalyze_M106_DipDetected(t *testing.T) {
	// The detector should still flag M-106's drop as a run - it's the
	// classifier's job (not analytics) to later recognize the matching
	// SCHEDULED_OUTAGE event and reclassify it as a false positive.
	byMeter := loadReadingsByMeter(t)
	analysis := Analyze("M-106", byMeter["M-106"])

	if len(analysis.Runs) == 0 {
		t.Fatalf("expected at least one run for M-106 (the outage dip), got none")
	}
	found := false
	for _, r := range analysis.Runs {
		if r.VariationPct < -30 {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a run with a strong negative variation for M-106, got runs: %+v", analysis.Runs)
	}

	// The outage resolves itself well before the dataset ends.
	for _, r := range analysis.Runs {
		if r.Ongoing {
			t.Errorf("expected M-106's outage dip to have resolved (not ongoing), got %+v", r)
		}
	}
}

func TestAnalyze_M104_IncreaseDetected(t *testing.T) {
	byMeter := loadReadingsByMeter(t)
	analysis := Analyze("M-104", byMeter["M-104"])

	if len(analysis.Runs) == 0 {
		t.Fatalf("expected at least one run for M-104 (the production-line increase), got none")
	}
	found := false
	for _, r := range analysis.Runs {
		if r.VariationPct > 20 {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a run with a positive variation for M-104, got runs: %+v", analysis.Runs)
	}
}

func TestAnalyze_M112_DataQualityNotConsumptionSpike(t *testing.T) {
	byMeter := loadReadingsByMeter(t)
	analysis := Analyze("M-112", byMeter["M-112"])

	if len(analysis.DQFlags) == 0 {
		t.Fatalf("expected data-quality flags for M-112 (erratic voltage/PF), got none")
	}
	// Consumption itself stays close to baseline for M-112 - the point of
	// this case is that it's a sensor fault, not a real load change.
	peak := totalAbsVariation(analysis.Runs)
	if peak > 30 {
		t.Errorf("expected M-112's consumption variation to stay modest (sensor fault, not a load change), got %.1f%%", peak)
	}
}

func TestAnalyze_QuietMetersHaveNoRuns(t *testing.T) {
	byMeter := loadReadingsByMeter(t)
	quietMeters := []string{"M-101", "M-102", "M-103", "M-105", "M-107", "M-108", "M-110", "M-111"}

	for _, meterID := range quietMeters {
		analysis := Analyze(meterID, byMeter[meterID])
		if len(analysis.Runs) > 0 {
			t.Errorf("expected no consumption runs for quiet meter %s, got %+v", meterID, analysis.Runs)
		}
		if len(analysis.DQFlags) > 0 {
			t.Errorf("expected no data-quality flags for quiet meter %s, got %d flags", meterID, len(analysis.DQFlags))
		}
	}
}
