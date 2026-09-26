package ai

import (
	"context"
	"fmt"

	"energy-platform/internal/analytics"
)

// TemplateExplainer builds Reason/RecommendedAction from the
// Classification's evidence with plain Go string formatting - no network
// call, no variability, always available. Used directly when there's no
// OPENAI_API_KEY, and as OpenAIExplainer's fallback when the API call
// fails.
type TemplateExplainer struct{}

func (TemplateExplainer) Explain(_ context.Context, c Classification) (Explanation, error) {
	switch c.Type {
	case TypeRealAnomaly:
		return explainRealAnomaly(c), nil
	case TypeExplainable:
		return explainExplainable(c), nil
	case TypeFalsePositive:
		return explainFalsePositive(c), nil
	case TypeDataQuality:
		return explainDataQuality(c), nil
	default:
		return Explanation{
			Reason:            fmt.Sprintf("Variación de %.1f%% respecto al baseline, sin clasificación específica.", c.VariationPct),
			RecommendedAction: "Revisar manualmente.",
		}, nil
	}
}

func explainRealAnomaly(c Classification) Explanation {
	onset := "de forma gradual, a lo largo de varias horas"
	if c.Evidence.Onset == analytics.OnsetStep {
		onset = "de forma abrupta, en cuestión de una hora"
	}
	status := "el consumo ya volvió a su nivel habitual"
	if c.Evidence.Ongoing {
		status = "el consumo sigue elevado al final del período analizado"
	}

	reason := fmt.Sprintf(
		"Consumo %.1f%% por encima del baseline (%.0f kWh vs %.0f kWh esperado) entre %s y %s, sin ningún evento operativo registrado que lo explique. El cambio ocurrió %s, y %s. Pico estadístico: z=%.1f (esperado: z<2.5).",
		c.VariationPct, c.ActualKWh, c.BaselineKWh,
		c.WindowStart.Format("02/01 15:04"), c.WindowEnd.Format("02/01 15:04"),
		onset, status, c.Evidence.PeakZScore,
	)
	return Explanation{
		Reason:            reason,
		RecommendedAction: fmt.Sprintf("Investigar el medidor %s y la instalación asociada en las próximas 24h.", c.MeterID),
	}
}

func explainExplainable(c Classification) Explanation {
	event := c.Evidence.RelatedEvent
	desc := "un evento operativo registrado"
	when := ""
	if event != nil {
		desc = fmt.Sprintf("'%s'", event.Description)
		when = " (" + event.Timestamp.Format("02/01 15:04") + ")"
	}
	reason := fmt.Sprintf(
		"Consumo %.1f%% por encima del baseline, coincidiendo con %s%s. La magnitud y el momento del cambio son consistentes con ese evento, no con una falla.",
		c.VariationPct, desc, when,
	)
	return Explanation{
		Reason:            reason,
		RecommendedAction: "Validar que el aumento corresponda al cambio operativo; no requiere escalamiento si el consumo se estabiliza en el nuevo nivel.",
	}
}

func explainFalsePositive(c Classification) Explanation {
	event := c.Evidence.RelatedEvent
	desc := "un evento operativo registrado"
	when := ""
	if event != nil {
		desc = fmt.Sprintf("'%s'", event.Description)
		when = " (" + event.Timestamp.Format("02/01 15:04") + ")"
	}
	reason := fmt.Sprintf(
		"Consumo %.1f%% por debajo del baseline entre %s y %s, explicado por %s%s. La caída está totalmente cubierta por ese evento planificado, no por una falla.",
		c.VariationPct, c.WindowStart.Format("02/01 15:04"), c.WindowEnd.Format("02/01 15:04"), desc, when,
	)
	return Explanation{
		Reason:            reason,
		RecommendedAction: "No requiere acción — se espera que el consumo vuelva al baseline una vez terminado el evento.",
	}
}

func explainDataQuality(c Classification) Explanation {
	reason := fmt.Sprintf(
		"%d de %d lecturas en el período analizado muestran voltaje y/o factor de potencia inconsistentes con el comportamiento estable del medidor (voltaje entre %.0fV y %.0fV vs ~%.0fV habitual; factor de potencia llegando a %.2f vs ~%.2f habitual), mientras el consumo se mantuvo dentro de %.1f%% del baseline. Las lecturas fuera de rango están dispersas, no forman un cambio sostenido — es consistente con una falla de sensor/telemetría, no con un cambio real de carga.",
		c.Evidence.DataQualityFlagCount, c.Evidence.DataQualitySampleSize,
		minFloat(c.Evidence.VoltageActualStartV, c.Evidence.VoltageActualEndV),
		maxFloat(c.Evidence.VoltageActualStartV, c.Evidence.VoltageActualEndV),
		c.Evidence.VoltageBaselineV,
		c.Evidence.PowerFactorActualStart, c.Evidence.PowerFactorBaseline,
		c.VariationPct,
	)
	return Explanation{
		Reason:            reason,
		RecommendedAction: fmt.Sprintf("Validar y recalibrar la telemetría (voltaje/factor de potencia) del medidor %s antes de confiar en próximas lecturas.", c.MeterID),
	}
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
