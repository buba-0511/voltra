package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOpenAIExplainer_NoAPIKey_UsesFallbackWithoutNetworkCall(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	defer srv.Close()

	e := NewOpenAIExplainer("")
	e.BaseURL = srv.URL

	c := sampleClassification(TypeRealAnomaly)
	exp, err := e.Explain(context.Background(), c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Errorf("expected no network call when APIKey is empty")
	}
	if exp.Reason == "" {
		t.Errorf("expected fallback explanation, got empty reason")
	}
}

func TestOpenAIExplainer_SuccessfulResponse_ParsesReasonAndAction(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Errorf("expected Authorization header, got %q", got)
		}
		content, _ := json.Marshal(llmOutput{
			Reason:            "El consumo subió 110% de forma abrupta y sigue así.",
			RecommendedAction: "Investigar el medidor.",
		})
		resp := chatResponse{}
		resp.Choices = []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		}{{}}
		resp.Choices[0].Message.Content = string(content)
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	e := NewOpenAIExplainer("test-key")
	e.BaseURL = srv.URL

	exp, err := e.Explain(context.Background(), sampleClassification(TypeRealAnomaly))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exp.Reason != "El consumo subió 110% de forma abrupta y sigue así." {
		t.Errorf("unexpected reason: %q", exp.Reason)
	}
	if exp.RecommendedAction != "Investigar el medidor." {
		t.Errorf("unexpected action: %q", exp.RecommendedAction)
	}
}

func TestOpenAIExplainer_APIError_FallsBackToTemplate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{"message": "invalid api key"},
		})
	}))
	defer srv.Close()

	e := NewOpenAIExplainer("bad-key")
	e.BaseURL = srv.URL

	c := sampleClassification(TypeFalsePositive)
	exp, err := e.Explain(context.Background(), c)
	if err != nil {
		t.Fatalf("Explain should fall back, not error: %v", err)
	}
	// Falls back to the template, which for FALSE_POSITIVE cites the event.
	if exp.Reason == "" {
		t.Errorf("expected a fallback reason, got empty")
	}
}

func TestOpenAIExplainer_MalformedContent_FallsBackToTemplate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := chatResponse{}
		resp.Choices = []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		}{{}}
		resp.Choices[0].Message.Content = "not valid json"
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	e := NewOpenAIExplainer("test-key")
	e.BaseURL = srv.URL

	exp, err := e.Explain(context.Background(), sampleClassification(TypeRealAnomaly))
	if err != nil {
		t.Fatalf("Explain should fall back, not error: %v", err)
	}
	if exp.Reason == "" {
		t.Errorf("expected fallback reason, got empty")
	}
}
