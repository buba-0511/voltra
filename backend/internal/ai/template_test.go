package ai

import (
	"context"
	"strings"
	"testing"
	"time"

	"energy-platform/internal/analytics"
)

func sampleClassification(typ AnomalyType) Classification {
	c := Classification{
		MeterID:      "M-109",
		Type:         typ,
		Severity:     SeverityHigh,
		Confidence:   0.95,
		WindowStart:  time.Date(2026, 9, 12, 14, 0, 0, 0, time.UTC),
		WindowEnd:    time.Date(2026, 9, 14, 23, 0, 0, 0, time.UTC),
		BaselineKWh:  2557.7,
		ActualKWh:    5380.8,
		VariationPct: 110.4,
		Evidence: Evidence{
			PeakZScore:             64.3,
			VoltageActualStartV:    218.2,
			VoltageActualEndV:      217.5,
			PowerFactorActualStart: 0.74,
			PowerFactorActualEnd:   0.76,
			Ongoing:                true,
			Onset:                  analytics.OnsetStep,
		},
	}
	if typ == TypeExplainable || typ == TypeFalsePositive {
		c.Evidence.RelatedEvent = &Event{
			Timestamp:   time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC),
			Type:        "OPERATIONAL_CHANGE",
			Description: "New production line activated",
		}
	}
	if typ == TypeDataQuality {
		c.Evidence.DataQualityFlagCount = 16
		c.Evidence.DataQualitySampleSize = 168
		c.Evidence.VoltageBaselineV = 220.0
		c.Evidence.PowerFactorBaseline = 0.94
	}
	return c
}

func TestTemplateExplainer_AllTypesProduceGroundedText(t *testing.T) {
	explainer := TemplateExplainer{}
	for _, typ := range []AnomalyType{TypeRealAnomaly, TypeExplainable, TypeFalsePositive, TypeDataQuality} {
		c := sampleClassification(typ)
		exp, err := explainer.Explain(context.Background(), c)
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", typ, err)
		}
		if exp.Reason == "" || exp.RecommendedAction == "" {
			t.Fatalf("%s: expected non-empty reason and action, got %+v", typ, exp)
		}
		if !strings.Contains(exp.Reason, "110.4") {
			t.Errorf("%s: expected reason to cite the actual variation %%, got %q", typ, exp.Reason)
		}
	}
}

func TestTemplateExplainer_ExplainableCitesEvent(t *testing.T) {
	c := sampleClassification(TypeExplainable)
	exp, _ := TemplateExplainer{}.Explain(context.Background(), c)
	if !strings.Contains(exp.Reason, "New production line activated") {
		t.Errorf("expected reason to cite the related event description, got %q", exp.Reason)
	}
}

func TestTemplateExplainer_DataQualityCitesFlagCount(t *testing.T) {
	c := sampleClassification(TypeDataQuality)
	exp, _ := TemplateExplainer{}.Explain(context.Background(), c)
	if !strings.Contains(exp.Reason, "16") || !strings.Contains(exp.Reason, "168") {
		t.Errorf("expected reason to cite the flag count (16 of 168), got %q", exp.Reason)
	}
}
