package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	openAIURL          = "https://api.openai.com/v1/chat/completions"
	defaultOpenAIModel = "gpt-4o-mini"
	requestTimeout     = 15 * time.Second
)

// OpenAIExplainer asks a real LLM to narrate a Classification's evidence
// into Reason/RecommendedAction. It never lets the model decide type,
// severity or confidence - those come from Classify() and are only handed
// to the model as already-established facts to write about. If the
// request fails for any reason (no API key, network error, bad response),
// it falls back to Fallback so a demo never breaks over a network hiccup.
type OpenAIExplainer struct {
	APIKey   string
	Model    string
	BaseURL  string // defaults to openAIURL; overridable in tests
	Fallback Explainer
	client   *http.Client
}

func NewOpenAIExplainer(apiKey string) *OpenAIExplainer {
	return &OpenAIExplainer{
		APIKey:   apiKey,
		Model:    defaultOpenAIModel,
		BaseURL:  openAIURL,
		Fallback: TemplateExplainer{},
		client:   &http.Client{Timeout: requestTimeout},
	}
}

func (e *OpenAIExplainer) Explain(ctx context.Context, c Classification) (Explanation, error) {
	if e.APIKey == "" {
		return e.Fallback.Explain(ctx, c)
	}

	exp, err := e.callOpenAI(ctx, c)
	if err != nil {
		return e.Fallback.Explain(ctx, c)
	}
	return exp, nil
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model          string            `json:"model"`
	Messages       []chatMessage     `json:"messages"`
	ResponseFormat map[string]string `json:"response_format"`
	Temperature    float64           `json:"temperature"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

type llmOutput struct {
	Reason            string `json:"reason"`
	RecommendedAction string `json:"recommended_action"`
}

func (e *OpenAIExplainer) callOpenAI(ctx context.Context, c Classification) (Explanation, error) {
	if e.client == nil {
		e.client = &http.Client{Timeout: requestTimeout}
	}
	if e.BaseURL == "" {
		e.BaseURL = openAIURL
	}
	reqBody := chatRequest{
		Model: e.Model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: buildEvidencePrompt(c)},
		},
		ResponseFormat: map[string]string{"type": "json_object"},
		Temperature:    0.3,
	}
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return Explanation{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.BaseURL, bytes.NewReader(payload))
	if err != nil {
		return Explanation{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+e.APIKey)

	resp, err := e.client.Do(req)
	if err != nil {
		return Explanation{}, err
	}
	defer resp.Body.Close()

	var chatResp chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return Explanation{}, err
	}
	if chatResp.Error != nil {
		return Explanation{}, fmt.Errorf("openai: %s", chatResp.Error.Message)
	}
	if len(chatResp.Choices) == 0 {
		return Explanation{}, fmt.Errorf("openai: empty response")
	}

	var out llmOutput
	if err := json.Unmarshal([]byte(chatResp.Choices[0].Message.Content), &out); err != nil {
		return Explanation{}, fmt.Errorf("openai: parse content: %w", err)
	}
	if out.Reason == "" || out.RecommendedAction == "" {
		return Explanation{}, fmt.Errorf("openai: incomplete response")
	}

	return Explanation{Reason: out.Reason, RecommendedAction: out.RecommendedAction}, nil
}

const systemPrompt = `Sos un asistente de operaciones de una plataforma de gestión energética. Tu única tarea es redactar, en español, la explicación de una anomalía que YA fue detectada y clasificada por un motor estadístico - no diagnostiques, no cambies la clasificación, no inventes datos que no te dieron. Basate únicamente en la evidencia que te pasan.

Devolvé exclusivamente un objeto JSON con dos campos:
- "reason": 2-3 oraciones explicando por qué se clasificó así, citando los números concretos de la evidencia (variación %, kWh, voltaje, factor de potencia, fechas).
- "recommended_action": una acción concreta y accionable, acorde a la severidad y el tipo.`

func buildEvidencePrompt(c Classification) string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "meter_id: %s\n", c.MeterID)
	fmt.Fprintf(&b, "type: %s\n", c.Type)
	fmt.Fprintf(&b, "severity: %s\n", c.Severity)
	fmt.Fprintf(&b, "confidence: %.2f\n", c.Confidence)
	fmt.Fprintf(&b, "window: %s a %s\n", c.WindowStart.Format(time.RFC3339), c.WindowEnd.Format(time.RFC3339))
	fmt.Fprintf(&b, "baseline_kwh: %.1f\n", c.BaselineKWh)
	fmt.Fprintf(&b, "actual_kwh: %.1f\n", c.ActualKWh)
	fmt.Fprintf(&b, "variation_pct: %.1f\n", c.VariationPct)
	fmt.Fprintf(&b, "peak_zscore: %.1f\n", c.Evidence.PeakZScore)
	fmt.Fprintf(&b, "ongoing: %v\n", c.Evidence.Ongoing)
	fmt.Fprintf(&b, "onset_pattern: %s\n", c.Evidence.Onset)
	if c.Type == TypeDataQuality {
		fmt.Fprintf(&b, "voltage_baseline_v: %.1f\n", c.Evidence.VoltageBaselineV)
		fmt.Fprintf(&b, "power_factor_baseline: %.2f\n", c.Evidence.PowerFactorBaseline)
		fmt.Fprintf(&b, "data_quality_flag_count: %d de %d\n", c.Evidence.DataQualityFlagCount, c.Evidence.DataQualitySampleSize)
	}
	fmt.Fprintf(&b, "voltage_v: %.1f -> %.1f\n", c.Evidence.VoltageActualStartV, c.Evidence.VoltageActualEndV)
	fmt.Fprintf(&b, "power_factor: %.2f -> %.2f\n", c.Evidence.PowerFactorActualStart, c.Evidence.PowerFactorActualEnd)
	if c.Evidence.RelatedEvent != nil {
		fmt.Fprintf(&b, "related_event: %s (%s) el %s\n", c.Evidence.RelatedEvent.Description, c.Evidence.RelatedEvent.Type, c.Evidence.RelatedEvent.Timestamp.Format(time.RFC3339))
	} else {
		fmt.Fprintf(&b, "related_event: ninguno\n")
	}
	return b.String()
}
