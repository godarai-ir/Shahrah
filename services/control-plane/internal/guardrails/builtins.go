package guardrails

import (
	"regexp"
	"strings"
)

const defaultRedaction = "[REDACTED]"

type patternGuardrail struct {
	name        string
	operation   Operation
	enforcement Enforcement
	reason      string
	patterns    []*regexp.Regexp
	replacement string
}

func (g *patternGuardrail) Name() string             { return g.name }
func (g *patternGuardrail) Operation() Operation     { return g.operation }
func (g *patternGuardrail) Enforcement() Enforcement { return g.enforcement }

func (g *patternGuardrail) Evaluate(input EvaluationInput) (EvaluationResult, error) {
	out := append([]string(nil), input.Texts...)
	matched := false

	for i, text := range out {
		for _, pattern := range g.patterns {
			if !pattern.MatchString(text) {
				continue
			}

			matched = true

			if g.operation == OperationMutate {
				out[i] = pattern.ReplaceAllString(out[i], g.replacement)
			}
		}
	}

	if !matched {
		return EvaluationResult{Texts: out}, nil
	}

	if g.operation == OperationMutate {
		return EvaluationResult{Texts: out, Mutated: true, Reason: g.reason}, nil
	}

	return EvaluationResult{Texts: out, Blocked: true, Reason: g.reason}, nil
}

type keywordBlocklistGuardrail struct {
	name        string
	terms       []string
	enforcement Enforcement
}

func NewKeywordBlocklistGuardrail() Guardrail {
	return &keywordBlocklistGuardrail{name: "keyword_blocklist", terms: []string{
		"ignore previous instructions",
		"jailbreak",
		"rm -rf",
		"drop table",
	}, enforcement: EnforcementEnforce}
}

func (g *keywordBlocklistGuardrail) Name() string             { return g.name }
func (g *keywordBlocklistGuardrail) Operation() Operation     { return OperationValidate }
func (g *keywordBlocklistGuardrail) Enforcement() Enforcement { return g.enforcement }

func (g *keywordBlocklistGuardrail) Evaluate(input EvaluationInput) (EvaluationResult, error) {
	for _, text := range input.Texts {
		lower := strings.ToLower(text)
		for _, term := range g.terms {
			if strings.Contains(lower, strings.ToLower(term)) {
				return EvaluationResult{Texts: input.Texts, Blocked: true, Reason: "Keyword blocklist matched"}, nil
			}
		}
	}

	return EvaluationResult{Texts: input.Texts}, nil
}

func NewRegexGuardrail() Guardrail {
	return newPatternGuardrailFromPatterns("regex", OperationMutate, EnforcementEnforce, "Sensitive pattern redacted", defaultRedaction, []string{
		`(?i)\b(?:secret|token|password)\s*[:=]\s*['"]?[^'"\s,;]+`,
	})
}

func NewSecretsDetectionGuardrail() Guardrail {
	return newSecretsDetectionGuardrail("secrets_detection", OperationMutate, EnforcementEnforce, defaultRedaction)
}

func newSecretsDetectionGuardrail(name string, operation Operation, enforcement Enforcement, replacement string) Guardrail {
	return &patternGuardrail{
		name:        name,
		operation:   operation,
		enforcement: enforcement,
		reason:      "Secret redacted",
		replacement: replacement,
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)\bBearer\s+[A-Za-z0-9._~+/=-]{16,}`),
			regexp.MustCompile(`\b[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\b`),
			regexp.MustCompile(`(?s)-----BEGIN [A-Z ]*PRIVATE KEY-----.*?-----END [A-Z ]*PRIVATE KEY-----`),
			regexp.MustCompile(`\bAKIA[0-9A-Z]{16}\b`),
			regexp.MustCompile(`\bsk-[A-Za-z0-9]{20,}\b`),
		},
	}
}

func NewBasicPIIGuardrail() Guardrail {
	return newBasicPIIGuardrail("basic_pii", OperationMutate, EnforcementEnforce, defaultRedaction)
}

func newBasicPIIGuardrail(name string, operation Operation, enforcement Enforcement, replacement string) Guardrail {
	return &patternGuardrail{
		name:        name,
		operation:   operation,
		enforcement: enforcement,
		reason:      "PII redacted",
		replacement: replacement,
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)\b[A-Z0-9._%+\-]+@[A-Z0-9.\-]+\.[A-Z]{2,}\b`),
			regexp.MustCompile(`(?:\+98|0)?9\d{9}\b`),
			regexp.MustCompile(`\b\d{10}\b`),
			regexp.MustCompile(`\b(?:\d[ -]?){16}\b`),
		},
	}
}

func newPatternGuardrailFromPatterns(name string, operation Operation, enforcement Enforcement, reason string, replacement string, rawPatterns []string) Guardrail {
	patterns := make([]*regexp.Regexp, 0, len(rawPatterns))
	for _, raw := range rawPatterns {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}

		pattern, err := regexp.Compile(raw)
		if err == nil {
			patterns = append(patterns, pattern)
		}
	}

	return &patternGuardrail{
		name:        name,
		operation:   operation,
		enforcement: enforcement,
		reason:      reason,
		replacement: replacement,
		patterns:    patterns,
	}
}
