//go:build live

package ai

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
)

func loadEnvLocal(t *testing.T) {
	f, err := os.Open("../../.env.local")
	if err != nil {
		t.Fatalf("open .env.local: %v", err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			os.Setenv(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
		}
	}
}

func TestLive_OpenAIExplainer_RealCall(t *testing.T) {
	loadEnvLocal(t)
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Fatal("OPENAI_API_KEY not set in .env.local")
	}

	e := NewOpenAIExplainer(apiKey)
	c := sampleClassification(TypeRealAnomaly)

	exp, err := e.Explain(context.Background(), c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fmt.Println("=== REASON ===")
	fmt.Println(exp.Reason)
	fmt.Println("=== RECOMMENDED ACTION ===")
	fmt.Println(exp.RecommendedAction)

	if exp.Reason == "" || exp.RecommendedAction == "" {
		t.Fatal("got empty reason/action - fell back to template, check the error path")
	}
}
