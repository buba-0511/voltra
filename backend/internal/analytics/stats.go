package analytics

import "math"

func mean(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	sum := 0.0
	for _, x := range xs {
		sum += x
	}
	return sum / float64(len(xs))
}

func stddev(xs []float64, m float64) float64 {
	if len(xs) < 2 {
		return 0
	}
	sumSq := 0.0
	for _, x := range xs {
		d := x - m
		sumSq += d * d
	}
	return math.Sqrt(sumSq / float64(len(xs)-1))
}

// zscore returns 0 when std is too small to be meaningful, avoiding
// division-by-near-zero blowing up otherwise-normal readings into
// false anomalies.
func zscore(x, m, std float64) float64 {
	if std < 1e-6 {
		std = 1e-6
	}
	return (x - m) / std
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
