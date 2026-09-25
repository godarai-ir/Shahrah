package guardrails

import (
	"errors"
	"strings"
	"testing"
)

func TestParseSemanticPolicyConfigFillsDefaults(t *testing.T) {
	t.Parallel()

	policy, err := ParseSemanticPolicyConfig(map[string]interface{}{
		"target": "output",
		"rules": []interface{}{
			map[string]interface{}{
				"key":       "financial_commitment",
				"statement": "The reply commits the company to a refund.",
				"examples":  map[string]interface{}{"true": []interface{}{"We will refund you today.", " "}},
			},
		},
		"severity":            map[string]interface{}{"enabled": true},
		"finglish_thresholds": map[string]interface{}{"review": 0.2, "block": 0.5},
	})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	rule := policy.Rules[0]
	if policy.Target != HookLLMOutput || rule.Thresholds.Review != 0.35 || rule.Thresholds.Block != 0.70 {
		t.Fatalf("defaults not applied: %+v", policy)
	}

	if len(rule.ExamplesTrue) != 1 || policy.SeverityBlockAt != 2.0 || policy.Finglish == nil || policy.Finglish.Block != 0.5 {
		t.Fatalf("unexpected policy: %+v", policy)
	}
}

func TestParseSemanticPolicyConfigUsesPackWordingAndThresholdOverrides(t *testing.T) {
	t.Parallel()

	policy, err := ParseSemanticPolicyConfig(map[string]interface{}{
		"pack":  SecurityBaselinePack,
		"rules": []interface{}{map[string]interface{}{"key": "jailbreak", "block_threshold": 0.9}},
	})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if len(policy.Rules) != 3 || policy.Rules[0].Key != "jailbreak" || policy.Rules[0].Thresholds.Block != 0.9 {
		t.Fatalf("pack rules not loaded: %+v", policy.Rules)
	}

	if !strings.Contains(policy.Rules[0].Statement, "drop its instructions") || policy.Rules[1].Thresholds.Block != 0.70 {
		t.Fatalf("pack wording or defaults changed: %+v", policy.Rules)
	}
}

func TestParseSemanticPolicyConfigRejectsInvalidConfigs(t *testing.T) {
	t.Parallel()

	rule := func(fields map[string]interface{}) map[string]interface{} {
		out := map[string]interface{}{"key": "rule_a", "statement": "A statement."}
		for key, value := range fields {
			out[key] = value
		}

		return out
	}

	cases := map[string]map[string]interface{}{
		"no rules":           {"rules": []interface{}{}},
		"bad target":         {"target": "both", "rules": []interface{}{rule(nil)}},
		"bad key":            {"rules": []interface{}{rule(map[string]interface{}{"key": "Bad Key"})}},
		"duplicate key":      {"rules": []interface{}{rule(nil), rule(nil)}},
		"empty statement":    {"rules": []interface{}{rule(map[string]interface{}{"statement": "  "})}},
		"long statement":     {"rules": []interface{}{rule(map[string]interface{}{"statement": strings.Repeat("x", 501)})}},
		"review over block":  {"rules": []interface{}{rule(map[string]interface{}{"review_threshold": 0.8, "block_threshold": 0.5})}},
		"block over one":     {"rules": []interface{}{rule(map[string]interface{}{"block_threshold": 1.5})}},
		"unknown field":      {"rules": []interface{}{rule(nil)}, "mode": "strict"},
		"unknown pack":       {"pack": "missing@1"},
		"pack on output":     {"pack": SecurityBaselinePack, "target": "output"},
		"unknown pack rule":  {"pack": SecurityBaselinePack, "rules": []interface{}{map[string]interface{}{"key": "nope"}}},
		"rule of other hook": {"pack": BankingPack, "target": "output", "rules": []interface{}{map[string]interface{}{"key": "auth_bypass"}}},
		"bad severity":       {"rules": []interface{}{rule(nil)}, "severity": map[string]interface{}{"enabled": true, "block_at": 4}},
		"bad finglish":       {"rules": []interface{}{rule(nil)}, "finglish_thresholds": map[string]interface{}{"review": 0.9, "block": 0.1}},
		"partial finglish":   {"rules": []interface{}{rule(nil)}, "finglish_thresholds": map[string]interface{}{"review": 0.2}},
		"empty finglish":     {"rules": []interface{}{rule(nil)}, "finglish_thresholds": map[string]interface{}{}},
		"too many examples":  {"rules": []interface{}{rule(map[string]interface{}{"examples": map[string]interface{}{"false": []interface{}{"a", "b", "c", "d", "e", "f"}}})}},
	}

	for name, config := range cases {
		if _, err := ParseSemanticPolicyConfig(config); !errors.Is(err, ErrInvalidSemanticPolicy) {
			t.Errorf("%s: expected ErrInvalidSemanticPolicy, got %v", name, err)
		}
	}
}

func TestParseSemanticPolicyConfigLoadsIndustryPackRulesPerHook(t *testing.T) {
	t.Parallel()

	cases := []struct {
		pack   string
		target string
		want   []string
	}{
		{BankingPack, "input", []string{"jailbreak", "system_prompt_extraction", "evaluator_address", "social_engineering", "auth_bypass"}},
		{BankingPack, "output", []string{"financial_commitment", "personal_financial_advice"}},
		{HealthcarePack, "input", []string{"jailbreak", "system_prompt_extraction", "evaluator_address", "self_harm_risk"}},
		{HealthcarePack, "output", []string{"diagnosis", "dosage", "self_harm_instructions"}},
		{HRPack, "input", []string{"jailbreak", "system_prompt_extraction", "evaluator_address", "hiring_discrimination"}},
		{HRPack, "output", []string{"hiring_discrimination_reply", "protected_attribute_inference"}},
	}

	for _, tc := range cases {
		policy, err := ParseSemanticPolicyConfig(map[string]interface{}{"pack": tc.pack, "target": tc.target})
		if err != nil {
			t.Fatalf("%s %s: %v", tc.pack, tc.target, err)
		}

		keys := make([]string, 0, len(policy.Rules))
		for _, rule := range policy.Rules {
			keys = append(keys, rule.Key)
		}

		if strings.Join(keys, ",") != strings.Join(tc.want, ",") {
			t.Errorf("%s %s: rules %v, want %v", tc.pack, tc.target, keys, tc.want)
		}
	}
}

func TestParseSemanticPolicyConfigKeepsReviewOnlyDefaultsAndOverrides(t *testing.T) {
	t.Parallel()

	policy, err := ParseSemanticPolicyConfig(map[string]interface{}{
		"pack":   BankingPack,
		"target": "output",
		"rules":  []interface{}{map[string]interface{}{"key": "financial_commitment", "review_threshold": 0.5, "block_threshold": 0.8}},
	})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if got := policy.Rules[0].Thresholds; got.Review != 0.5 || got.Block != 0.8 {
		t.Fatalf("override not applied: %+v", got)
	}

	if got := policy.Rules[1].Thresholds; got.Block != 1 {
		t.Fatalf("personal_financial_advice must default to review only: %+v", got)
	}
}

// Pack rules skip the validation custom rules get, so this keeps every built-in rule within the same limits.
func TestSemanticPacksAreValid(t *testing.T) {
	t.Parallel()

	for id, pack := range semanticPacks {
		if len(pack.rules) == 0 {
			t.Errorf("%s has no rules", id)
		}

		for hook, rules := range pack.rules {
			if len(rules) == 0 || len(rules) > maxSemanticRules {
				t.Errorf("%s %s: %d rules", id, hook, len(rules))
			}

			seen := map[string]bool{}

			for _, rule := range rules {
				if !semanticRuleKeyPattern.MatchString(rule.Key) || seen[rule.Key] {
					t.Errorf("%s %s: bad or duplicate key %q", id, hook, rule.Key)
				}

				seen[rule.Key] = true

				config := semanticRuleConfig{
					Key: rule.Key, Statement: rule.Statement, TrueWhen: rule.TrueWhen, FalseWhen: rule.FalseWhen,
					ReviewThreshold: &rule.Thresholds.Review, BlockThreshold: &rule.Thresholds.Block,
				}
				if _, ruleErr := customSemanticRule(rule.Key, config); ruleErr != nil {
					t.Errorf("%s %s: %v", id, hook, ruleErr)
				}
			}
		}
	}
}
