package ai

import "context"

// Explanation is the narrative half of a classification: why it matters
// and what to do about it, in plain language grounded in the evidence
// Classify already computed.
type Explanation struct {
	Reason            string
	RecommendedAction string
}

// Explainer turns a Classification's evidence into prose. Implementations
// must never change type/severity/confidence - only narrate them.
type Explainer interface {
	Explain(ctx context.Context, c Classification) (Explanation, error)
}
