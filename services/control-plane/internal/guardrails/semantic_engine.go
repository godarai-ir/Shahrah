package guardrails

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

// JudgeQuestionKind is the TypeSafe question primitive.
type JudgeQuestionKind string

const (
	JudgeQuestionNoul  JudgeQuestionKind = "noul"
	JudgeQuestionScore JudgeQuestionKind = "score"

	severityQuestionKey = "severity"
	checkStatusReview   = "review"
)

var (
	// ErrSemanticJudgeUnavailable means semantic guardrails are selected but no judge is configured.
	ErrSemanticJudgeUnavailable = errors.New("semantic guardrail judge is not configured")

	errSemanticNeedsEngine = errors.New("semantic guardrails are evaluated in a batch by the engine")
	errSemanticAnswer      = errors.New("semantic judge returned no answer")

	// severityLevels describe situations rather than degrees, because the judge scores each level on its own.
	severityLevels = []string{
		"Complying causes no harm to anyone.",
		"Complying is rude, off-policy, or embarrassing, but nothing of value is lost.",
		"Complying exposes personal data, money, or safety to real risk.",
		"Complying is likely to cause major financial, legal, or physical damage.",
	}
)

// JudgeQuestion is one typed question for the semantic judge.
type JudgeQuestion struct {
	Kind          JudgeQuestionKind
	Instructions  string
	TrueWhen      string
	FalseWhen     string
	ExamplesTrue  []string
	ExamplesFalse []string
	Levels        []string
}

// JudgeRequest evaluates every question against one state in a single call.
type JudgeRequest struct {
	State     map[string]any
	Questions map[string]JudgeQuestion
}

// JudgeResponse holds a Noul probability or a Score value per question key.
type JudgeResponse struct {
	Answers map[string]float64
	Model   string
	Cached  bool
}

// SemanticJudge answers typed questions about a text. The Jev client in internal/semanticjudge implements it.
type SemanticJudge interface {
	Judge(ctx context.Context, request JudgeRequest) (JudgeResponse, error)
}

// SemanticCheckDetail is recorded on traces only. It is never serialized to API clients, so an attacker cannot
// learn where a rule's boundary lies.
type SemanticCheckDetail struct {
	Decision  platform.SemanticDecision
	Language  LanguageClass
	Rules     []platform.SemanticRuleScore
	Severity  *float64
	Truncated bool
	Cached    bool
	Model     string
}

type semanticPolicyGuardrail struct {
	name        string
	enforcement Enforcement
	policy      SemanticPolicy
	err         error
}

type semanticItem struct {
	selector string
	prefix   string
	g        *semanticPolicyGuardrail
}

func newSemanticPolicyGuardrail(name string, enforcement Enforcement, config map[string]interface{}) Guardrail {
	policy, err := ParseSemanticPolicyConfig(config)

	return &semanticPolicyGuardrail{name: name, enforcement: enforcement, policy: policy, err: err}
}

func (g *semanticPolicyGuardrail) Name() string             { return g.name }
func (g *semanticPolicyGuardrail) Operation() Operation     { return OperationValidate }
func (g *semanticPolicyGuardrail) Enforcement() Enforcement { return g.enforcement }

func (g *semanticPolicyGuardrail) Evaluate(EvaluationInput) (EvaluationResult, error) {
	return EvaluationResult{}, errSemanticNeedsEngine
}

func (e *Engine) semanticGuardrail(selector string, overrides map[string]Guardrail) *semanticPolicyGuardrail {
	g, ok := e.resolve(selector, overrides).(*semanticPolicyGuardrail)
	if !ok {
		return nil
	}

	return g
}

// applySemantic evaluates every selected semantic guardrail of one hook with a single judge call. It returns the
// checks in selection order and the index of the first check that blocks, or -1.
func (e *Engine) applySemantic(ctx context.Context, hook Hook, items []semanticItem, state semanticState) ([]Check, int) {
	start := time.Now()
	checks := make([]Check, len(items))
	pending := make([]int, 0, len(items))

	for i, item := range items {
		checks[i] = Check{Selector: item.selector, Operation: OperationValidate, Enforcement: item.g.enforcement}

		switch {
		case item.g.err != nil:
			setSemanticError(&checks[i], item.g.err)
		case item.g.policy.Target != hook:
			setSemanticError(&checks[i], fmt.Errorf("%w: policy targets the %s", ErrInvalidSemanticPolicy, targetName(item.g.policy.Target)))
		case e.judge == nil:
			setSemanticError(&checks[i], ErrSemanticJudgeUnavailable)
		default:
			pending = append(pending, i)
		}
	}

	e.judgePending(ctx, items, pending, state, checks)

	blocking := -1

	for i := range checks {
		checks[i].LatencyMS = time.Since(start).Milliseconds()
		if blocking < 0 && semanticCheckBlocks(checks[i], items[i].g) {
			blocking = i
		}
	}

	return checks, blocking
}

func (e *Engine) judgePending(ctx context.Context, items []semanticItem, pending []int, state semanticState, checks []Check) {
	if len(pending) == 0 {
		return
	}

	if strings.TrimSpace(state.Content) == "" {
		for _, i := range pending {
			checks[i].Status = "passed"
		}

		return
	}

	request := JudgeRequest{State: state.judgeState(), Questions: map[string]JudgeQuestion{}}
	for _, i := range pending {
		addSemanticQuestions(request.Questions, items[i])
	}

	response, err := e.judge.Judge(ctx, request)
	language := ClassifyLanguage(state.Content)

	for _, i := range pending {
		if err != nil {
			setSemanticError(&checks[i], err)

			continue
		}

		detail, decideErr := decideSemantic(items[i], response, language)
		if decideErr != nil {
			setSemanticError(&checks[i], decideErr)

			continue
		}

		detail.Truncated = state.Truncated
		applySemanticDecision(&checks[i], items[i].g, detail)
	}
}

func addSemanticQuestions(questions map[string]JudgeQuestion, item semanticItem) {
	for _, rule := range item.g.policy.Rules {
		questions[item.prefix+rule.Key] = JudgeQuestion{
			Kind:          JudgeQuestionNoul,
			Instructions:  rule.Statement,
			TrueWhen:      rule.TrueWhen,
			FalseWhen:     rule.FalseWhen,
			ExamplesTrue:  rule.ExamplesTrue,
			ExamplesFalse: rule.ExamplesFalse,
		}
	}

	if item.g.policy.SeverityBlockAt > 0 {
		questions[severityQuestionKey] = JudgeQuestion{
			Kind:         JudgeQuestionScore,
			Instructions: "How much harm would follow if the assistant fully complied with this content?",
			Levels:       severityLevels,
		}
	}
}

func decideSemantic(item semanticItem, response JudgeResponse, language LanguageClass) (SemanticCheckDetail, error) {
	policy := item.g.policy
	detail := SemanticCheckDetail{Decision: platform.SemanticDecisionPass, Language: language, Cached: response.Cached, Model: response.Model}

	for _, rule := range policy.Rules {
		probability, ok := response.Answers[item.prefix+rule.Key]
		if !ok {
			return SemanticCheckDetail{}, fmt.Errorf("%w for rule %s", errSemanticAnswer, rule.Key)
		}

		thresholds := rule.Thresholds
		if language == LanguageFinglish && policy.Finglish != nil {
			thresholds = *policy.Finglish
		}

		decision := bandDecision(probability, thresholds)
		detail.Rules = append(detail.Rules, platform.SemanticRuleScore{Key: rule.Key, Probability: probability, Decision: decision})
		detail.Decision = strictestDecision(detail.Decision, decision)
	}

	if policy.SeverityBlockAt > 0 {
		severity, ok := response.Answers[severityQuestionKey]
		if !ok {
			return SemanticCheckDetail{}, fmt.Errorf("%w for severity", errSemanticAnswer)
		}

		detail.Severity = &severity
		if detail.Decision == platform.SemanticDecisionReview && severity >= policy.SeverityBlockAt {
			detail.Decision = platform.SemanticDecisionBlock
		}
	}

	return detail, nil
}

// bandDecision maps a probability to a band. A block threshold of 1 turns blocking off, so the rule can only reach review.
func bandDecision(probability float64, thresholds SemanticThresholds) platform.SemanticDecision {
	switch {
	case thresholds.Block < 1 && probability >= thresholds.Block:
		return platform.SemanticDecisionBlock
	case probability >= thresholds.Review:
		return platform.SemanticDecisionReview
	default:
		return platform.SemanticDecisionPass
	}
}

func strictestDecision(a, b platform.SemanticDecision) platform.SemanticDecision {
	rank := map[platform.SemanticDecision]int{platform.SemanticDecisionPass: 0, platform.SemanticDecisionReview: 1, platform.SemanticDecisionBlock: 2}
	if rank[b] > rank[a] {
		return b
	}

	return a
}

func applySemanticDecision(check *Check, g *semanticPolicyGuardrail, detail SemanticCheckDetail) {
	check.Semantic = &detail

	matched := make([]string, 0, len(detail.Rules))
	for _, rule := range detail.Rules {
		if rule.Decision == platform.SemanticDecisionBlock || (detail.Decision == platform.SemanticDecisionBlock && rule.Decision == platform.SemanticDecisionReview) {
			matched = append(matched, rule.Key)
		}
	}

	switch detail.Decision {
	case platform.SemanticDecisionBlock:
		check.Status = "blocked"
		check.Reason = "Semantic policy rule matched: " + strings.Join(matched, ", ")
	case platform.SemanticDecisionReview:
		check.Status = checkStatusReview
		check.Reason = "Semantic policy needs review: " + strings.Join(reviewedRules(detail.Rules), ", ")
	default:
		check.Status = "passed"

		return
	}

	if g.enforcement == EnforcementAudit {
		check.Status = "flagged"
	}
}

func reviewedRules(rules []platform.SemanticRuleScore) []string {
	out := make([]string, 0, len(rules))

	for _, rule := range rules {
		if rule.Decision == platform.SemanticDecisionReview {
			out = append(out, rule.Key)
		}
	}

	return out
}

func setSemanticError(check *Check, err error) {
	check.Status = "error"
	check.Reason = "Guardrail execution failed"
	check.Error = err.Error()
}

func semanticCheckBlocks(check Check, g *semanticPolicyGuardrail) bool {
	switch check.Status {
	case "blocked":
		return g.enforcement != EnforcementAudit
	case "error":
		return shouldBlockOnError(g)
	default:
		return false
	}
}

// TestSemanticPolicy evaluates an unsaved policy config against one sample text for the wizard's live test.
func (e *Engine) TestSemanticPolicy(ctx context.Context, config map[string]interface{}, text string) (platform.SemanticPolicyTestResponse, error) {
	policy, err := ParseSemanticPolicyConfig(config)
	if err != nil {
		return platform.SemanticPolicyTestResponse{}, err
	}

	if e == nil || e.judge == nil {
		return platform.SemanticPolicyTestResponse{}, ErrSemanticJudgeUnavailable
	}

	state := semanticState{Role: "user"}
	if policy.Target == HookLLMOutput {
		state.Role = "assistant"
	}

	state.Content, state.Truncated = truncateTail(NormalizeSemanticText(text))
	item := semanticItem{selector: "test", prefix: "g0__", g: &semanticPolicyGuardrail{name: "test", enforcement: EnforcementEnforce, policy: policy}}
	request := JudgeRequest{State: state.judgeState(), Questions: map[string]JudgeQuestion{}}
	addSemanticQuestions(request.Questions, item)

	start := time.Now()

	response, err := e.judge.Judge(ctx, request)
	if err != nil {
		return platform.SemanticPolicyTestResponse{}, err
	}

	detail, err := decideSemantic(item, response, ClassifyLanguage(state.Content))
	if err != nil {
		return platform.SemanticPolicyTestResponse{}, err
	}

	return platform.SemanticPolicyTestResponse{
		Decision:  detail.Decision,
		Language:  string(detail.Language),
		Rules:     detail.Rules,
		Severity:  detail.Severity,
		LatencyMS: time.Since(start).Milliseconds(),
		Truncated: state.Truncated,
	}, nil
}
