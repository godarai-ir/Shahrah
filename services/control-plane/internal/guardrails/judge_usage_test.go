package guardrails

import (
	"context"
	"errors"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

type cachedJudge struct{}

func (cachedJudge) Judge(_ context.Context, request JudgeRequest) (JudgeResponse, error) {
	answers := make(map[string]float64, len(request.Questions))
	for key := range request.Questions {
		answers[key] = 0.05
	}

	return JudgeResponse{Answers: answers, Model: "jev-test", Cached: true}, nil
}

func judgeUsageFor(t *testing.T, judge SemanticJudge) *JudgeUsage {
	t.Helper()

	engine := NewDefaultEngine(WithSemanticJudge(judge))
	overrides := map[string]Guardrail{
		"topic": semanticTenantGuardrail("topic", platform.GuardrailEnforcementEnforce, customRuleConfig("input", "off_topic")),
	}
	selection := RequestSelection{LLMInput: []string{"guardrails/topic"}, Scope: ScopeLast}

	usage := &JudgeUsage{}
	ctx := WithJudgeUsage(context.Background(), usage)

	// Only the counted calls matter here; a failed judge may block the request.
	_, _ = engine.ApplyInputWithGuardrails(ctx, userRequest("hello"), selection, overrides)

	return usage
}

func TestJudgeUsageCountsCallsThatReachTheJudge(t *testing.T) {
	t.Parallel()

	usage := judgeUsageFor(t, &fakeJudge{})
	if usage.Calls() != 1 || usage.Model() != "jev-test" {
		t.Fatalf("expected one counted jev-test call, got %d %q", usage.Calls(), usage.Model())
	}

	usage.Reset()

	if usage.Calls() != 0 {
		t.Fatalf("reset kept %d calls", usage.Calls())
	}
}

func TestJudgeUsageSkipsCachedAndFailedCalls(t *testing.T) {
	t.Parallel()

	if calls := judgeUsageFor(t, cachedJudge{}).Calls(); calls != 0 {
		t.Fatalf("a cached answer is free, counted %d", calls)
	}

	if calls := judgeUsageFor(t, &fakeJudge{err: errors.New("judge down")}).Calls(); calls != 0 {
		t.Fatalf("a failed call is free, counted %d", calls)
	}
}

func TestJudgeUsageIsOptional(t *testing.T) {
	t.Parallel()

	engine := NewDefaultEngine(WithSemanticJudge(&fakeJudge{}))
	overrides := map[string]Guardrail{
		"topic": semanticTenantGuardrail("topic", platform.GuardrailEnforcementEnforce, customRuleConfig("input", "off_topic")),
	}

	if _, err := engine.ApplyInputWithGuardrails(context.Background(), userRequest("hello"), RequestSelection{LLMInput: []string{"guardrails/topic"}, Scope: ScopeLast}, overrides); err != nil {
		t.Fatalf("a context without usage still evaluates: %v", err)
	}
}
