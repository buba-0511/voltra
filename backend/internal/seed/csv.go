package seed

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const timeLayout = "2006-01-02 15:04:05"

type rawReading struct {
	MeterID        string
	Timestamp      time.Time
	ConsumptionKWh float64
	VoltageV       float64
	CurrentA       float64
	PowerFactor    float64
	Status         string
}

type rawEvent struct {
	MeterID     string
	Timestamp   time.Time
	Type        string
	Description string
}

func parseReadings(dataDir string) ([]rawReading, []string, error) {
	f, err := os.Open(filepath.Join(dataDir, "readings.csv"))
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, nil, err
	}
	if len(rows) < 2 {
		return nil, nil, fmt.Errorf("readings.csv has no data rows")
	}

	var out []rawReading
	seen := map[string]bool{}
	var meterIDs []string
	for _, row := range rows[1:] {
		if len(row) < 7 {
			continue
		}
		ts, err := time.Parse(timeLayout, row[1])
		if err != nil {
			return nil, nil, fmt.Errorf("parse timestamp %q: %w", row[1], err)
		}
		consumption, _ := strconv.ParseFloat(row[2], 64)
		voltage, _ := strconv.ParseFloat(row[3], 64)
		current, _ := strconv.ParseFloat(row[4], 64)
		pf, _ := strconv.ParseFloat(row[5], 64)

		if !seen[row[0]] {
			seen[row[0]] = true
			meterIDs = append(meterIDs, row[0])
		}
		out = append(out, rawReading{
			MeterID:        row[0],
			Timestamp:      ts,
			ConsumptionKWh: consumption,
			VoltageV:       voltage,
			CurrentA:       current,
			PowerFactor:    pf,
			Status:         row[6],
		})
	}
	return out, meterIDs, nil
}

func parseEvents(dataDir string) ([]rawEvent, error) {
	f, err := os.Open(filepath.Join(dataDir, "events.csv"))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) < 2 {
		return nil, nil
	}

	var out []rawEvent
	for _, row := range rows[1:] {
		if len(row) < 4 {
			continue
		}
		ts, err := time.Parse(timeLayout, row[1])
		if err != nil {
			return nil, fmt.Errorf("parse event timestamp %q: %w", row[1], err)
		}
		out = append(out, rawEvent{
			MeterID:     row[0],
			Timestamp:   ts,
			Type:        row[2],
			Description: row[3],
		})
	}
	return out, nil
}
