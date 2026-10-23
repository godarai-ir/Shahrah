package credits

// Side-usage prices, in millicredits per call (product decision, 2026-10-17).
// The judge and the semantic-cache embedding report no provider cost, so they
// are priced per call instead of from cost components. A change here is a
// price change and belongs in the public pricing changelog.
const (
	// GuardrailJudgeCallMC is charged for each semantic-guardrail judge call
	// that reaches the judge; cached answers are free.
	GuardrailJudgeCallMC int64 = 1 * MillicreditsPerCredit
	// SemanticCacheEmbeddingMC is charged for each semantic-cache embedding.
	SemanticCacheEmbeddingMC int64 = 0
)

// SourceFixedPrice means the charge is a fixed price per call.
const SourceFixedPrice Source = "fixed_price"
