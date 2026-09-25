package guardrails

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

const (
	// SemanticPolicyType is the tenant guardrail type evaluated by the semantic judge.
	SemanticPolicyType = "semantic_policy"

	// SecurityBaselinePack is the built-in pack of attack-detection rules.
	SecurityBaselinePack = "security-baseline@1"
	// BankingPack adds social engineering, verification bypass, commitments, and personal investment advice.
	BankingPack = "banking@1"
	// HealthcarePack adds self-harm risk, diagnosis, dosage, and self-harm instructions.
	HealthcarePack = "healthcare@1"
	// HRPack adds discrimination in hiring and guessing protected characteristics.
	HRPack = "hr@1"
	// ConfidentialityPack adds classified documents, non-public business data, personal records, and security internals.
	ConfidentialityPack = "confidentiality@1"

	maxSemanticRules          = 20
	maxSemanticStatementRunes = 500
	maxSemanticCriteriaRunes  = 500
	maxSemanticExamples       = 5
	maxSemanticExampleRunes   = 300

	defaultSemanticReviewThreshold = 0.35
	defaultSemanticBlockThreshold  = 0.70
	defaultSemanticSeverityBlockAt = 2.0
	maxSemanticSeverity            = 3.0
)

var (
	// ErrInvalidSemanticPolicy reports a semantic_policy config that cannot be evaluated.
	ErrInvalidSemanticPolicy = errors.New("invalid semantic policy")

	semanticRuleKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{1,47}$`)
)

// SemanticThresholds are the two probability cut points of the three decision bands.
type SemanticThresholds struct {
	Review float64 `json:"review"`
	Block  float64 `json:"block"`
}

// SemanticRule is one yes/no proposition. A higher probability means a violation.
type SemanticRule struct {
	Key           string             `json:"key"`
	Statement     string             `json:"statement"`
	TrueWhen      string             `json:"true_when,omitempty"`
	FalseWhen     string             `json:"false_when,omitempty"`
	ExamplesTrue  []string           `json:"examples_true,omitempty"`
	ExamplesFalse []string           `json:"examples_false,omitempty"`
	Thresholds    SemanticThresholds `json:"thresholds"`
}

// SemanticPolicy is the parsed and validated config of one semantic_policy guardrail.
type SemanticPolicy struct {
	Target Hook
	Pack   string
	Rules  []SemanticRule
	// SeverityBlockAt upgrades a review decision to a block when the severity score reaches it. Zero disables severity.
	SeverityBlockAt float64
	// Finglish replaces every rule's thresholds when the text is Persian written in Latin letters.
	Finglish *SemanticThresholds
}

type semanticRuleConfig struct {
	Key             string                 `json:"key"`
	Statement       string                 `json:"statement"`
	TrueWhen        string                 `json:"true_when"`
	FalseWhen       string                 `json:"false_when"`
	Examples        *semanticExampleConfig `json:"examples"`
	ReviewThreshold *float64               `json:"review_threshold"`
	BlockThreshold  *float64               `json:"block_threshold"`
}

type semanticExampleConfig struct {
	True  []string `json:"true"`
	False []string `json:"false"`
}

type semanticSeverityConfig struct {
	Enabled bool     `json:"enabled"`
	BlockAt *float64 `json:"block_at"`
}

// semanticThresholdsConfig uses pointers so a missing value is rejected instead of silently becoming 0.
type semanticThresholdsConfig struct {
	Review *float64 `json:"review"`
	Block  *float64 `json:"block"`
}

type semanticPolicyConfig struct {
	Target             string                    `json:"target"`
	Pack               string                    `json:"pack"`
	Rules              []semanticRuleConfig      `json:"rules"`
	Severity           *semanticSeverityConfig   `json:"severity"`
	FinglishThresholds *semanticThresholdsConfig `json:"finglish_thresholds"`
}

// ParseSemanticPolicyConfig validates a semantic_policy config and fills defaults. Pack rules take their wording
// from the pack; the config may only override their thresholds.
func ParseSemanticPolicyConfig(config map[string]interface{}) (SemanticPolicy, error) {
	raw, err := json.Marshal(config)
	if err != nil {
		return SemanticPolicy{}, fmt.Errorf("%w: %w", ErrInvalidSemanticPolicy, err)
	}

	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()

	var parsed semanticPolicyConfig
	if err = decoder.Decode(&parsed); err != nil {
		return SemanticPolicy{}, fmt.Errorf("%w: %w", ErrInvalidSemanticPolicy, err)
	}

	policy := SemanticPolicy{Pack: strings.TrimSpace(parsed.Pack)}

	switch strings.TrimSpace(parsed.Target) {
	case "", "input":
		policy.Target = HookLLMInput
	case "output":
		policy.Target = HookLLMOutput
	default:
		return SemanticPolicy{}, invalidSemanticPolicy("target must be input or output")
	}

	if policy.Rules, err = semanticRules(policy, parsed.Rules); err != nil {
		return SemanticPolicy{}, err
	}

	if policy.SeverityBlockAt, err = semanticSeverity(parsed.Severity); err != nil {
		return SemanticPolicy{}, err
	}

	if parsed.FinglishThresholds != nil {
		if parsed.FinglishThresholds.Review == nil || parsed.FinglishThresholds.Block == nil {
			return SemanticPolicy{}, invalidSemanticPolicy("finglish_thresholds requires both review and block")
		}

		finglish := SemanticThresholds{Review: *parsed.FinglishThresholds.Review, Block: *parsed.FinglishThresholds.Block}
		if thresholdErr := validateThresholds("finglish_thresholds", finglish); thresholdErr != nil {
			return SemanticPolicy{}, thresholdErr
		}

		policy.Finglish = &finglish
	}

	return policy, nil
}

func semanticRules(policy SemanticPolicy, configured []semanticRuleConfig) ([]SemanticRule, error) {
	if policy.Pack == "" {
		return customSemanticRules(configured)
	}

	pack, ok := semanticPacks[policy.Pack]
	if !ok {
		return nil, invalidSemanticPolicy("unknown pack " + policy.Pack)
	}

	packRules := pack.rules[policy.Target]
	if len(packRules) == 0 {
		return nil, invalidSemanticPolicy("pack " + policy.Pack + " has no " + targetName(policy.Target) + " rules")
	}

	overrides := make(map[string]semanticRuleConfig, len(configured))
	for _, rule := range configured {
		overrides[rule.Key] = rule
	}

	rules := make([]SemanticRule, 0, len(packRules))

	for _, base := range packRules {
		rule := base

		override, found := overrides[rule.Key]
		if found {
			delete(overrides, rule.Key)

			thresholds, err := ruleThresholds(rule.Key, override)
			if err != nil {
				return nil, err
			}

			rule.Thresholds = thresholds
		}

		rules = append(rules, rule)
	}

	for key := range overrides {
		return nil, invalidSemanticPolicy("pack " + policy.Pack + " has no " + targetName(policy.Target) + " rule " + key)
	}

	return rules, nil
}

func customSemanticRules(configured []semanticRuleConfig) ([]SemanticRule, error) {
	if len(configured) == 0 || len(configured) > maxSemanticRules {
		return nil, invalidSemanticPolicy(fmt.Sprintf("rules must contain between 1 and %d rules", maxSemanticRules))
	}

	rules := make([]SemanticRule, 0, len(configured))
	seen := make(map[string]bool, len(configured))

	for _, item := range configured {
		key := strings.TrimSpace(item.Key)
		if !semanticRuleKeyPattern.MatchString(key) || seen[key] {
			return nil, invalidSemanticPolicy(fmt.Sprintf("rule key %q must be unique, lowercase, and 2-48 characters", key))
		}

		seen[key] = true

		rule, err := customSemanticRule(key, item)
		if err != nil {
			return nil, err
		}

		rules = append(rules, rule)
	}

	return rules, nil
}

func customSemanticRule(key string, item semanticRuleConfig) (SemanticRule, error) {
	rule := SemanticRule{
		Key:       key,
		Statement: strings.TrimSpace(item.Statement),
		TrueWhen:  strings.TrimSpace(item.TrueWhen),
		FalseWhen: strings.TrimSpace(item.FalseWhen),
	}

	if rule.Statement == "" || utf8.RuneCountInString(rule.Statement) > maxSemanticStatementRunes {
		return SemanticRule{}, invalidSemanticPolicy(fmt.Sprintf("rule %s statement must be 1-%d characters", key, maxSemanticStatementRunes))
	}

	if utf8.RuneCountInString(rule.TrueWhen) > maxSemanticCriteriaRunes || utf8.RuneCountInString(rule.FalseWhen) > maxSemanticCriteriaRunes {
		return SemanticRule{}, invalidSemanticPolicy(fmt.Sprintf("rule %s criteria must be at most %d characters", key, maxSemanticCriteriaRunes))
	}

	if item.Examples != nil {
		var err error
		if rule.ExamplesTrue, err = semanticExamples(key, item.Examples.True); err != nil {
			return SemanticRule{}, err
		}

		if rule.ExamplesFalse, err = semanticExamples(key, item.Examples.False); err != nil {
			return SemanticRule{}, err
		}
	}

	thresholds, err := ruleThresholds(key, item)
	if err != nil {
		return SemanticRule{}, err
	}

	rule.Thresholds = thresholds

	return rule, nil
}

func semanticExamples(key string, values []string) ([]string, error) {
	out := make([]string, 0, len(values))

	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}

		if utf8.RuneCountInString(value) > maxSemanticExampleRunes {
			return nil, invalidSemanticPolicy(fmt.Sprintf("rule %s examples must be at most %d characters", key, maxSemanticExampleRunes))
		}

		out = append(out, value)
	}

	if len(out) > maxSemanticExamples {
		return nil, invalidSemanticPolicy(fmt.Sprintf("rule %s may have at most %d examples per side", key, maxSemanticExamples))
	}

	return out, nil
}

func ruleThresholds(key string, item semanticRuleConfig) (SemanticThresholds, error) {
	thresholds := SemanticThresholds{Review: defaultSemanticReviewThreshold, Block: defaultSemanticBlockThreshold}
	if item.ReviewThreshold != nil {
		thresholds.Review = *item.ReviewThreshold
	}

	if item.BlockThreshold != nil {
		thresholds.Block = *item.BlockThreshold
	}

	return thresholds, validateThresholds("rule "+key, thresholds)
}

func validateThresholds(field string, thresholds SemanticThresholds) error {
	if thresholds.Review < 0 || thresholds.Block > 1 || thresholds.Review > thresholds.Block {
		return invalidSemanticPolicy(field + " thresholds must satisfy 0 <= review <= block <= 1")
	}

	return nil
}

func semanticSeverity(config *semanticSeverityConfig) (float64, error) {
	if config == nil || !config.Enabled {
		return 0, nil
	}

	blockAt := defaultSemanticSeverityBlockAt
	if config.BlockAt != nil {
		blockAt = *config.BlockAt
	}

	if blockAt <= 0 || blockAt > maxSemanticSeverity {
		return 0, invalidSemanticPolicy("severity.block_at must be greater than 0 and at most 3")
	}

	return blockAt, nil
}

func invalidSemanticPolicy(message string) error {
	return fmt.Errorf("%w: %s", ErrInvalidSemanticPolicy, message)
}

func targetName(hook Hook) string {
	if hook == HookLLMOutput {
		return "output"
	}

	return "input"
}
