package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
)

const (
	targetInput  = "input"
	targetOutput = "output"

	langEN = "en"
	langFA = "fa"

	kindPositive     = "positive"
	kindHardNegative = "hard_negative"
	kindBenign       = "benign"
)

var validRegisters = []string{"fa_formal", "fa_colloquial", "finglish", "en"}

// validTags mark the harder samples: injections embedded in longer text, obfuscated spelling, multi-turn context,
// long messages, and indirect paraphrases.
var validTags = []string{"embedded", "obfuscated", "multi_turn", "long", "paraphrase"}

var errInvalidDataset = errors.New("invalid dataset")

// localizedText holds the English and Persian wording of one rule field.
type localizedText struct {
	EN string `json:"en"`
	FA string `json:"fa"`
}

func (t localizedText) in(lang string) string {
	if lang == langFA {
		return t.FA
	}

	return t.EN
}

// rule is one semantic policy rule, evaluated as a single Jev Noul question.
type rule struct {
	Key       string        `json:"key"`
	Target    string        `json:"target"`
	Source    string        `json:"source"`
	Statement localizedText `json:"statement"`
	TrueWhen  localizedText `json:"true_when"`
	FalseWhen localizedText `json:"false_when"`
}

// ruleSet is the versioned list of rules under evaluation.
type ruleSet struct {
	Version string `json:"version"`
	Rules   []rule `json:"rules"`
}

// turn is one earlier message of a multi-turn sample.
type turn struct {
	Role string `json:"role"`
	Text string `json:"text"`
}

// sample is one labeled text. Every rule of the sample's target that is not listed in Positive is a negative label.
// Text is the message under evaluation; Turns are the earlier messages of the same conversation.
type sample struct {
	ID       string   `json:"id"`
	Target   string   `json:"target"`
	Register string   `json:"register"`
	Kind     string   `json:"kind"`
	Positive []string `json:"positive"`
	Tags     []string `json:"tags,omitempty"`
	Turns    []turn   `json:"turns,omitempty"`
	Text     string   `json:"text"`
}

func (s sample) isPositiveFor(ruleKey string) bool {
	return slices.Contains(s.Positive, ruleKey)
}

// dataset is the complete evaluation set loaded from disk.
type dataset struct {
	Rules   ruleSet
	Samples []sample
}

func (d dataset) rulesFor(target string) []rule {
	out := make([]rule, 0, len(d.Rules.Rules))

	for _, r := range d.Rules.Rules {
		if r.Target == target {
			out = append(out, r)
		}
	}

	return out
}

func loadDataset(dir string) (dataset, error) {
	rulesBody, err := os.ReadFile(filepath.Join(dir, "rules.json"))
	if err != nil {
		return dataset{}, fmt.Errorf("read rules: %w", err)
	}

	var rules ruleSet
	if err = decodeStrict(rulesBody, &rules); err != nil {
		return dataset{}, fmt.Errorf("decode rules: %w", err)
	}

	files, err := filepath.Glob(filepath.Join(dir, "samples", "*.jsonl"))
	if err != nil {
		return dataset{}, fmt.Errorf("list sample files: %w", err)
	}

	sort.Strings(files)

	samples := make([]sample, 0, 512)

	for _, file := range files {
		fileSamples, readErr := readSampleFile(file)
		if readErr != nil {
			return dataset{}, readErr
		}

		samples = append(samples, fileSamples...)
	}

	loaded := dataset{Rules: rules, Samples: samples}
	if validateErr := validateDataset(loaded); validateErr != nil {
		return dataset{}, validateErr
	}

	return loaded, nil
}

func readSampleFile(path string) ([]sample, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	var out []sample

	scanner := bufio.NewScanner(bytes.NewReader(body))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	line := 0

	for scanner.Scan() {
		line++

		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}

		var s sample
		if err = decodeStrict([]byte(text), &s); err != nil {
			return nil, fmt.Errorf("%s:%d: %w", filepath.Base(path), line, err)
		}

		out = append(out, s)
	}

	if err = scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan %s: %w", path, err)
	}

	return out, nil
}

func decodeStrict(body []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()

	return decoder.Decode(target)
}

func validateDataset(d dataset) error {
	problems := validateRules(d.Rules)
	rulesByKey := make(map[string]rule, len(d.Rules.Rules))

	for _, r := range d.Rules.Rules {
		rulesByKey[r.Key] = r
	}

	seenIDs := make(map[string]bool, len(d.Samples))
	seenTexts := make(map[string]string, len(d.Samples))

	for _, s := range d.Samples {
		if seenIDs[s.ID] {
			problems = append(problems, fmt.Sprintf("duplicate sample id %q", s.ID))
		}

		normalized := strings.ToLower(strings.Join(strings.Fields(s.Text), " "))
		if other, ok := seenTexts[normalized]; ok {
			problems = append(problems, fmt.Sprintf("sample %q repeats the text of %q", s.ID, other))
		}

		seenIDs[s.ID] = true
		seenTexts[normalized] = s.ID
		problems = append(problems, validateSample(s, rulesByKey)...)
	}

	if len(problems) == 0 {
		return nil
	}

	return fmt.Errorf("%w:\n  %s", errInvalidDataset, strings.Join(problems, "\n  "))
}

func validateRules(rules ruleSet) []string {
	var problems []string

	if strings.TrimSpace(rules.Version) == "" {
		problems = append(problems, "rules version is empty")
	}

	seen := make(map[string]bool, len(rules.Rules))

	for _, r := range rules.Rules {
		if r.Key == "" || seen[r.Key] {
			problems = append(problems, fmt.Sprintf("rule key %q is empty or duplicated", r.Key))
		}

		seen[r.Key] = true

		if r.Target != targetInput && r.Target != targetOutput {
			problems = append(problems, fmt.Sprintf("rule %q has invalid target %q", r.Key, r.Target))
		}

		for _, field := range []localizedText{r.Statement, r.TrueWhen, r.FalseWhen} {
			if strings.TrimSpace(field.EN) == "" || strings.TrimSpace(field.FA) == "" {
				problems = append(problems, fmt.Sprintf("rule %q is missing an English or Persian text", r.Key))

				break
			}
		}
	}

	return problems
}

func validateSample(s sample, rulesByKey map[string]rule) []string {
	var problems []string

	if s.ID == "" || strings.TrimSpace(s.Text) == "" {
		problems = append(problems, fmt.Sprintf("sample %q has an empty id or text", s.ID))
	}

	if s.Target != targetInput && s.Target != targetOutput {
		problems = append(problems, fmt.Sprintf("sample %q has invalid target %q", s.ID, s.Target))
	}

	if !slices.Contains(validRegisters, s.Register) {
		problems = append(problems, fmt.Sprintf("sample %q has invalid register %q", s.ID, s.Register))
	}

	switch s.Kind {
	case kindPositive:
		if len(s.Positive) == 0 {
			problems = append(problems, fmt.Sprintf("positive sample %q lists no rule", s.ID))
		}
	case kindHardNegative, kindBenign:
		if len(s.Positive) != 0 {
			problems = append(problems, fmt.Sprintf("negative sample %q lists positive rules", s.ID))
		}
	default:
		problems = append(problems, fmt.Sprintf("sample %q has invalid kind %q", s.ID, s.Kind))
	}

	for _, tag := range s.Tags {
		if !slices.Contains(validTags, tag) {
			problems = append(problems, fmt.Sprintf("sample %q has invalid tag %q", s.ID, tag))
		}
	}

	for _, t := range s.Turns {
		if (t.Role != "user" && t.Role != "assistant") || strings.TrimSpace(t.Text) == "" {
			problems = append(problems, fmt.Sprintf("sample %q has an invalid turn", s.ID))
		}
	}

	if len(s.Turns) > 0 && !slices.Contains(s.Tags, "multi_turn") {
		problems = append(problems, fmt.Sprintf("sample %q has turns but no multi_turn tag", s.ID))
	}

	for _, key := range s.Positive {
		r, ok := rulesByKey[key]
		if !ok {
			problems = append(problems, fmt.Sprintf("sample %q references unknown rule %q", s.ID, key))

			continue
		}

		if r.Target != s.Target {
			problems = append(problems, fmt.Sprintf("sample %q targets %s but rule %q targets %s", s.ID, s.Target, key, r.Target))
		}
	}

	return problems
}
