package analytics

const (
	consumptionZThreshold = 2.5
	minRunHours           = 3
	// minVariationPct filters out runs that cross the z-score threshold by
	// chance but aren't practically meaningful (e.g. a 3-hour blip that's
	// only 1-2% off baseline) - statistical significance isn't the same as
	// a change worth reporting.
	minVariationPct = 5.0

	dqZThreshold = 5.0
	// dqSustainedRunHours: voltage/PF readings that deviate from baseline
	// but do so in one long contiguous block are a *real, sustained*
	// electrical change (e.g. power factor genuinely dropping under a
	// heavier load) - that's evidence for a REAL_ANOMALY, not a sensor
	// fault. A data-quality issue looks the opposite: isolated, erratic
	// blips scattered between otherwise-normal readings. This threshold is
	// what separates the two.
	dqSustainedRunHours = 6
)

// BuildBaseline computes the meter's expected hourly consumption profile
// and stable voltage/power-factor range from a reference period (the
// caller decides what counts as "baseline" - typically the first half of
// a time-sorted reading history).
func BuildBaseline(baselineReadings []Reading) MeterBaseline {
	var byHour [24][]float64
	var voltages, pfs []float64

	for _, r := range baselineReadings {
		h := r.Timestamp.Hour()
		byHour[h] = append(byHour[h], r.ConsumptionKWh)
		voltages = append(voltages, r.VoltageV)
		pfs = append(pfs, r.PowerFactor)
	}

	var b MeterBaseline
	for h := 0; h < 24; h++ {
		m := mean(byHour[h])
		b.Hourly.Mean[h] = m
		b.Hourly.Std[h] = stddev(byHour[h], m)
	}
	b.VoltageMean = mean(voltages)
	b.VoltageStd = stddev(voltages, b.VoltageMean)
	b.PFMean = mean(pfs)
	b.PFStd = stddev(pfs, b.PFMean)
	return b
}

// baselineTotalFor projects the baseline's hourly profile onto a set of
// readings, summing the expected consumption for each reading's hour of
// day - the same accumulation a ConsumptionRun's BaselineTotal uses, but
// reusable for any reading window.
func baselineTotalFor(baseline MeterBaseline, readings []Reading) float64 {
	total := 0.0
	for _, r := range readings {
		total += baseline.Hourly.Mean[r.Timestamp.Hour()]
	}
	return total
}

// MeterPeriodBaseline projects a meter's expected total consumption for
// its whole reading period, using the first half (already time-sorted) to
// build the hourly baseline profile - the same baseline Analyze scores
// anomalies against, just totaled over the full period instead of a
// single flagged run. This is what lets a meter with no detected anomaly
// still report a baseline/variation (challenge brief section 6's example
// table shows a variación for every meter, not just flagged ones).
func MeterPeriodBaseline(readings []Reading) float64 {
	if len(readings) == 0 {
		return 0
	}
	baseline := BuildBaseline(readings[:len(readings)/2])
	return baselineTotalFor(baseline, readings)
}

// Analyze splits a meter's full, time-sorted reading history in half: the
// first half establishes the baseline, the second half is scored against
// it for consumption-spike runs and electrical data-quality
// inconsistencies.
func Analyze(meterID string, readings []Reading) MeterAnalysis {
	n := len(readings)
	half := n / 2
	baselineReadings := readings[:half]
	evalReadings := readings[half:]

	baseline := BuildBaseline(baselineReadings)

	consumptionFlagged := make([]bool, len(evalReadings))
	consumptionZ := make([]float64, len(evalReadings))
	dqFlaggedRaw := make([]bool, len(evalReadings))
	voltageZ := make([]float64, len(evalReadings))
	pfZ := make([]float64, len(evalReadings))

	for i, r := range evalReadings {
		h := r.Timestamp.Hour()
		cz := zscore(r.ConsumptionKWh, baseline.Hourly.Mean[h], baseline.Hourly.Std[h])
		consumptionZ[i] = cz
		if abs(cz) > consumptionZThreshold {
			consumptionFlagged[i] = true
		}

		vz := zscore(r.VoltageV, baseline.VoltageMean, baseline.VoltageStd)
		pz := zscore(r.PowerFactor, baseline.PFMean, baseline.PFStd)
		voltageZ[i] = vz
		pfZ[i] = pz
		if abs(vz) > dqZThreshold || abs(pz) > dqZThreshold {
			dqFlaggedRaw[i] = true
		}
	}

	runs := buildConsumptionRuns(evalReadings, consumptionZ, consumptionFlagged, baseline)
	dqFlags := buildDataQualityFlags(evalReadings, voltageZ, pfZ, dqFlaggedRaw)

	return MeterAnalysis{
		MeterID:        meterID,
		Baseline:       baseline,
		Runs:           runs,
		DQFlags:        dqFlags,
		EvalSampleSize: len(evalReadings),
	}
}

// runSpan is a contiguous (gap-bridged) block of flagged indices.
type runSpan struct{ start, end int }

// findRunSpans merges flagged indices into contiguous blocks, bridging a
// single non-flagged gap (a lone normal reading inside an otherwise
// flagged stretch), so brief noise doesn't split what's really one
// sustained event into several.
func findRunSpans(flagged []bool) []runSpan {
	var spans []runSpan
	i := 0
	for i < len(flagged) {
		if !flagged[i] {
			i++
			continue
		}
		start := i
		end := i
		for end+1 < len(flagged) {
			if flagged[end+1] {
				end++
				continue
			}
			if end+2 < len(flagged) && flagged[end+2] {
				end += 2 // bridge a single-reading gap
				continue
			}
			break
		}
		spans = append(spans, runSpan{start, end})
		i = end + 1
	}
	return spans
}

func buildConsumptionRuns(readings []Reading, zscores []float64, flagged []bool, baseline MeterBaseline) []ConsumptionRun {
	var runs []ConsumptionRun
	for _, span := range findRunSpans(flagged) {
		if span.end-span.start+1 < minRunHours {
			continue
		}
		ongoing := span.end == len(readings)-1
		run := buildRun(readings[span.start:span.end+1], zscores[span.start:span.end+1], baseline, ongoing)
		if abs(run.VariationPct) < minVariationPct {
			continue
		}
		runs = append(runs, run)
	}
	return runs
}

func buildRun(readings []Reading, zscores []float64, baseline MeterBaseline, ongoing bool) ConsumptionRun {
	actualTotal := 0.0
	sumAbsZ := 0.0
	peakAbsZ := 0.0
	peakIdx := 0
	for i, r := range readings {
		actualTotal += r.ConsumptionKWh
		az := abs(zscores[i])
		sumAbsZ += az
		if az > peakAbsZ {
			peakAbsZ = az
			peakIdx = i
		}
	}
	baselineTotal := baselineTotalFor(baseline, readings)
	variation := 0.0
	if baselineTotal > 0 {
		variation = (actualTotal - baselineTotal) / baselineTotal * 100
	}
	// Onset shape: compare the very first flagged reading's deviation to
	// the run's peak. If it already arrives most of the way to the peak
	// (rather than climbing there over several hours), the change was a
	// step, not a ramp - peakIdx alone isn't reliable here because the
	// z-score keeps fluctuating with noise even while already anomalous.
	onset := OnsetGradual
	firstAbsZ := abs(zscores[0])
	if peakAbsZ > 0 && firstAbsZ/peakAbsZ >= 0.5 {
		onset = OnsetStep
	}
	return ConsumptionRun{
		Start:         readings[0].Timestamp,
		End:           readings[len(readings)-1].Timestamp,
		AvgAbsZ:       sumAbsZ / float64(len(readings)),
		PeakAbsZ:      peakAbsZ,
		Ongoing:       ongoing,
		OnsetHours:    peakIdx,
		Onset:         onset,
		ActualTotal:   actualTotal,
		BaselineTotal: baselineTotal,
		VariationPct:  variation,
		VoltageStart:  readings[0].VoltageV,
		VoltageEnd:    readings[len(readings)-1].VoltageV,
		PFStart:       readings[0].PowerFactor,
		PFEnd:         readings[len(readings)-1].PowerFactor,
	}
}

// buildDataQualityFlags keeps only the flagged readings that belong to a
// short/isolated span (see dqSustainedRunHours) - a long contiguous span
// is treated as evidence of a real, sustained electrical change instead
// and is surfaced via ConsumptionRun's voltage/PF fields, not here.
func buildDataQualityFlags(readings []Reading, voltageZ, pfZ []float64, flagged []bool) []DataQualityFlag {
	var flags []DataQualityFlag
	for _, span := range findRunSpans(flagged) {
		if span.end-span.start+1 >= dqSustainedRunHours {
			continue
		}
		for i := span.start; i <= span.end; i++ {
			r := readings[i]
			flags = append(flags, DataQualityFlag{
				Timestamp:      r.Timestamp,
				ConsumptionKWh: r.ConsumptionKWh,
				VoltageV:       r.VoltageV,
				PF:             r.PowerFactor,
				VoltageZ:       voltageZ[i],
				PFZ:            pfZ[i],
			})
		}
	}
	return flags
}
