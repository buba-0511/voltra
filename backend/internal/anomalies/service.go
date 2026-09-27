package anomalies

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/sync/errgroup"

	"energy-platform/internal/ai"
	"energy-platform/internal/analytics"
	sqlcgen "energy-platform/internal/db/sqlc"
)

// maxConcurrentMeterAnalysis bounds how many meters get read+analyzed at
// once - unbounded fan-out would open one DB round trip per meter
// simultaneously, exhausting the connection pool once the tenant has more
// than a handful of meters.
const maxConcurrentMeterAnalysis = 8

type Service struct {
	Queries   *sqlcgen.Queries
	Explainer ai.Explainer
}

func NewService(q *sqlcgen.Queries, explainer ai.Explainer) *Service {
	return &Service{Queries: q, Explainer: explainer}
}

// RunAnalysis is the pipeline the challenge brief calls out: Lecturas →
// Baseline → Detección → Correlación → Eventos → Explicación →
// Recomendación. It runs once per meter (statistical detection +
// classification are fast; the explanation step is the slow part, so it
// runs concurrently across anomalies rather than one at a time) and
// persists an analysis_runs row plus one anomalies row per detected
// anomaly.
func (s *Service) RunAnalysis(ctx context.Context, tenantID int32) (sqlcgen.AnalysisRun, error) {
	run, err := s.Queries.CreateAnalysisRun(ctx, tenantID)
	if err != nil {
		return sqlcgen.AnalysisRun{}, fmt.Errorf("create analysis run: %w", err)
	}

	meters, err := s.Queries.ListMeters(ctx, tenantID)
	if err != nil {
		return s.fail(ctx, run.ID)
	}

	classifications, err := s.classifyAll(ctx, tenantID, meters)
	if err != nil {
		return s.fail(ctx, run.ID)
	}

	explanations := s.explainAll(ctx, classifications)

	highPriority := int32(0)
	for i, c := range classifications {
		if c.Severity == ai.SeverityHigh {
			highPriority++
		}
		if err := s.insertAnomaly(ctx, tenantID, run.ID, c, explanations[i]); err != nil {
			return s.fail(ctx, run.ID)
		}
	}

	return s.Queries.CompleteAnalysisRun(ctx, sqlcgen.CompleteAnalysisRunParams{
		ID:             run.ID,
		Status:         "DONE",
		Stage:          "DONE",
		MetersAnalyzed: int32(len(meters)),
		AnomaliesFound: int32(len(classifications)),
		HighPriority:   highPriority,
	})
}

// classifyAll reads and classifies every meter concurrently (bounded by
// maxConcurrentMeterAnalysis) instead of one at a time - previously a
// sequential loop, so a tenant with many meters made every "Run AI
// Analysis" click take proportionally longer.
func (s *Service) classifyAll(ctx context.Context, tenantID int32, meters []sqlcgen.Meter) ([]ai.Classification, error) {
	perMeter := make([][]ai.Classification, len(meters))
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(maxConcurrentMeterAnalysis)

	for i, m := range meters {
		g.Go(func() (err error) {
			// A panic here (nil pointer, index out of range, whatever)
			// would otherwise take down the entire server process, not
			// just this one meter's analysis - errgroup doesn't recover
			// goroutine panics on its own.
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("panic analyzing meter %s: %v", m.MeterID, r)
				}
			}()

			readings, err := s.Queries.ListReadingsByMeter(gctx, sqlcgen.ListReadingsByMeterParams{
				TenantID: tenantID, MeterID: m.MeterID,
			})
			if err != nil {
				return err
			}
			if len(readings) < 2 {
				return nil
			}
			events, err := s.Queries.ListEventsByMeter(gctx, sqlcgen.ListEventsByMeterParams{
				TenantID: tenantID, MeterID: m.MeterID,
			})
			if err != nil {
				return err
			}

			analysis := analytics.Analyze(m.MeterID, toAnalyticsReadings(readings))
			perMeter[i] = ai.Classify(m.MeterID, analysis, toAIEvents(events))
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}

	var classifications []ai.Classification
	for _, c := range perMeter {
		classifications = append(classifications, c...)
	}
	return classifications, nil
}

func (s *Service) fail(ctx context.Context, runID int64) (sqlcgen.AnalysisRun, error) {
	run, err := s.Queries.CompleteAnalysisRun(ctx, sqlcgen.CompleteAnalysisRunParams{
		ID: runID, Status: "FAILED", Stage: "FAILED",
	})
	if err != nil {
		return sqlcgen.AnalysisRun{}, err
	}
	return run, fmt.Errorf("analysis run %d failed", runID)
}

// explainAll fetches explanations concurrently - each one is an
// independent OpenAI call, and doing them serially would make a
// four-anomaly analysis take ~15s instead of ~3-4s.
func (s *Service) explainAll(ctx context.Context, classifications []ai.Classification) []ai.Explanation {
	explanations := make([]ai.Explanation, len(classifications))
	var wg sync.WaitGroup
	for i, c := range classifications {
		wg.Add(1)
		go func(i int, c ai.Classification) {
			defer wg.Done()
			// Same reasoning as classifyAll: an unrecovered panic in any
			// one of these would crash the whole process instead of just
			// failing this anomaly's explanation. Fall back to the
			// template explainer either way, same as a returned error.
			defer func() {
				if r := recover(); r != nil {
					log.Printf("panic explaining anomaly for %s: %v", c.MeterID, r)
					exp, _ := ai.TemplateExplainer{}.Explain(ctx, c)
					explanations[i] = exp
				}
			}()
			exp, err := s.Explainer.Explain(ctx, c)
			if err != nil {
				exp, _ = ai.TemplateExplainer{}.Explain(ctx, c)
			}
			explanations[i] = exp
		}(i, c)
	}
	wg.Wait()
	return explanations
}

func (s *Service) insertAnomaly(ctx context.Context, tenantID int32, runID int64, c ai.Classification, exp ai.Explanation) error {
	evidence, err := json.Marshal(c.Evidence)
	if err != nil {
		return fmt.Errorf("marshal evidence: %w", err)
	}
	_, err = s.Queries.InsertAnomaly(ctx, sqlcgen.InsertAnomalyParams{
		TenantID:          tenantID,
		AnalysisRunID:     runID,
		MeterID:           c.MeterID,
		WindowStart:       pgtype.Timestamptz{Time: c.WindowStart, Valid: true},
		WindowEnd:         pgtype.Timestamptz{Time: c.WindowEnd, Valid: true},
		Type:              string(c.Type),
		Severity:          string(c.Severity),
		Confidence:        c.Confidence,
		Reason:            exp.Reason,
		RecommendedAction: exp.RecommendedAction,
		BaselineKwh:       c.BaselineKWh,
		ActualKwh:         c.ActualKWh,
		VariationPct:      c.VariationPct,
		Evidence:          evidence,
	})
	return err
}

// toAnalyticsReadings/toAIEvents force UTC on every timestamp coming out
// of Postgres: pgx decodes timestamptz using time.Local, so without this
// a timestamp embedded in the evidence JSON (e.g. related_event.timestamp)
// would carry the host machine's offset instead of Z, inconsistent with
// everything that goes through httpx.FormatTime.
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

func toAIEvents(rows []sqlcgen.Event) []ai.Event {
	out := make([]ai.Event, len(rows))
	for i, e := range rows {
		out[i] = ai.Event{
			Timestamp:   e.Timestamp.Time.UTC(),
			Type:        e.Type,
			Description: e.Description,
		}
	}
	return out
}
