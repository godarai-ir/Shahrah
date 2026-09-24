package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestShippedDatasetIsValidAndBalanced(t *testing.T) {
	t.Parallel()

	data, err := loadDataset("data")
	if err != nil {
		t.Fatalf("load dataset: %v", err)
	}

	if len(data.Samples) < 400 {
		t.Fatalf("expected at least 400 samples, got %d", len(data.Samples))
	}

	for _, r := range data.Rules.Rules {
		perRegister := make(map[string]int)
		negatives := 0

		for _, s := range data.Samples {
			if s.Target != r.Target {
				continue
			}

			if s.isPositiveFor(r.Key) {
				perRegister[s.Register]++
			} else {
				negatives++
			}
		}

		for _, register := range validRegisters {
			if perRegister[register] < 6 {
				t.Errorf("rule %s has %d positives in %s, want at least 6", r.Key, perRegister[register], register)
			}
		}

		if negatives < 30 {
			t.Errorf("rule %s has %d negatives, want at least 30", r.Key, negatives)
		}
	}
}

func TestLoadDatasetRejectsInvalidSamples(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rules := `{"version":"t","rules":[{"key":"a","target":"input","source":"s",` +
		`"statement":{"en":"x","fa":"x"},"true_when":{"en":"x","fa":"x"},"false_when":{"en":"x","fa":"x"}}]}`
	samples := `{"id":"1","target":"input","register":"fa_formal","kind":"positive","positive":["missing"],"text":"x"}` + "\n" +
		`{"id":"1","target":"output","register":"klingon","kind":"benign","positive":["a"],"text":"x"}`

	writeTestFile(t, filepath.Join(dir, "rules.json"), rules)
	writeTestFile(t, filepath.Join(dir, "samples", "bad.jsonl"), samples)

	_, err := loadDataset(dir)
	if !errors.Is(err, errInvalidDataset) {
		t.Fatalf("expected errInvalidDataset, got %v", err)
	}

	for _, want := range []string{"unknown rule", "duplicate sample id", "invalid register", "lists positive rules", "targets output"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error does not mention %q:\n%v", want, err)
		}
	}
}

func TestBuildRequestUsesVariantLanguageAndWrapsState(t *testing.T) {
	t.Parallel()

	rules := []rule{{
		Key:       "k",
		Statement: localizedText{EN: "english", FA: "فارسی"},
		TrueWhen:  localizedText{EN: "t-en", FA: "t-fa"},
		FalseWhen: localizedText{EN: "f-en", FA: "f-fa"},
	}}

	request := buildRequest("jev-1.13.0", langFA, rules, sample{Target: targetOutput, Text: "hello"})

	question := request.Questions["k"]
	if question.Type != "noul" || question.Instructions != "فارسی" || question.Criteria["true"] != "t-fa" || question.Criteria["false"] != "f-fa" {
		t.Fatalf("unexpected question: %+v", question)
	}

	if request.State["untrusted_content"] != "hello" || request.State["role"] != "assistant" {
		t.Fatalf("unexpected state: %+v", request.State)
	}

	if _, ok := request.State["prior_turns"]; ok {
		t.Fatal("single-turn sample must not send prior_turns")
	}

	turns := []turn{{Role: "user", Text: "earlier"}}
	request = buildRequest("jev-1.13.0", langEN, rules, sample{Target: targetInput, Text: "now", Turns: turns})

	if got, ok := request.State["prior_turns"].([]turn); !ok || len(got) != 1 || request.State["role"] != "user" {
		t.Fatalf("multi-turn state not sent: %+v", request.State)
	}
}

func TestMetricsIgnoreSamplesPositiveOnlyForAnotherRule(t *testing.T) {
	t.Parallel()

	predictions := []prediction{
		{Variant: langEN, Rule: "a", Register: "en", Kind: kindPositive, Label: true, Score: 0.9},
		{Variant: langEN, Rule: "a", Register: "en", Kind: kindPositive, Label: false, Score: 0.95},
		{Variant: langEN, Rule: "a", Register: "en", Kind: kindHardNegative, Label: false, Score: 0.1},
		{Variant: langEN, Rule: "a", Register: "en", Kind: kindBenign, Hard: true, Label: false, Score: 0.2},
	}

	groups := computeGroupMetrics(predictions, []string{langEN}, []string{"a"}, 0)

	var all, hard *groupMetrics

	for i := range groups {
		switch groups[i].Group {
		case groupAll:
			all = &groups[i]
		case groupHard:
			hard = &groups[i]
		}
	}

	if all == nil || all.Positives != 1 || all.Negatives != 2 || all.Best == nil || all.Best.Recall != 1 {
		t.Fatalf("other-rule positive must not count as a negative: %+v", all)
	}

	if hard == nil || hard.Negatives != 1 || hard.Positives != 0 {
		t.Fatalf("hard group must hold only tagged samples: %+v", hard)
	}
}

func TestValidateRejectsTurnsWithoutMultiTurnTag(t *testing.T) {
	t.Parallel()

	rules := map[string]rule{}
	s := sample{ID: "x", Target: targetInput, Register: "en", Kind: kindBenign, Text: "t", Turns: []turn{{Role: "user", Text: "u"}}}

	problems := validateSample(s, rules)
	if len(problems) != 1 || !strings.Contains(problems[0], "multi_turn") {
		t.Fatalf("expected a missing multi_turn tag problem, got %v", problems)
	}

	s.Tags = []string{"multi_turn"}
	s.Turns = []turn{{Role: "system", Text: "u"}}

	if problems = validateSample(s, rules); len(problems) != 1 || !strings.Contains(problems[0], "invalid turn") {
		t.Fatalf("expected an invalid turn problem, got %v", problems)
	}
}

func TestEvaluateRetriesRateLimitAndSendsAuth(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.WriteHeader(http.StatusTooManyRequests)

			return
		}

		if r.URL.Path != systemOnePath || r.Header.Get("Authorization") != "Bearer secret" {
			w.WriteHeader(http.StatusUnauthorized)

			return
		}

		var request systemOneRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil || request.Model != "m" {
			w.WriteHeader(http.StatusUnprocessableEntity)

			return
		}

		_, _ = w.Write([]byte(`{"model":"m","answers":{"k":{"type":"noul","noul":0.9}},"usage":{"input_tokens":12}}`))
	}))
	defer server.Close()

	client := &jevClient{baseURL: server.URL, apiKey: "secret", httpClient: server.Client(), maxAttempts: 3, backoff: time.Millisecond}

	response, err := client.evaluate(context.Background(), systemOneRequest{Model: "m"})
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}

	if calls.Load() != 2 || response.Usage.InputTokens != 12 || *response.Answers["k"].Noul != 0.9 {
		t.Fatalf("unexpected result: calls=%d response=%+v", calls.Load(), response)
	}
}

func TestEvaluateDoesNotRetryClientErrors(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusUnprocessableEntity)
	}))
	defer server.Close()

	client := &jevClient{baseURL: server.URL, apiKey: "k", httpClient: server.Client(), maxAttempts: 3, backoff: time.Millisecond}

	_, err := client.evaluate(context.Background(), systemOneRequest{})
	if !errors.Is(err, errJevStatus) || calls.Load() != 1 {
		t.Fatalf("expected one non-retried call with errJevStatus, got calls=%d err=%v", calls.Load(), err)
	}
}

func TestBestAtMaxFPRPicksHighestRecallWithinBudget(t *testing.T) {
	t.Parallel()

	predictions := []prediction{
		{Label: true, Score: 0.95},
		{Label: true, Score: 0.80},
		{Label: true, Score: 0.40},
		{Label: false, Score: 0.60},
		{Label: false, Score: 0.20},
		{Label: false, Score: 0.10},
	}

	best, ok := bestAtMaxFPR(predictions, 0)
	if !ok || best.Recall != 2.0/3 || best.FP != 0 || best.Threshold <= 0.60 || best.Threshold > 0.80 {
		t.Fatalf("unexpected best point: %+v ok=%t", best, ok)
	}

	auc, ok := rocAUC(predictions)
	if !ok || auc != 8.0/9 {
		t.Fatalf("expected AUC 8/9, got %v ok=%t", auc, ok)
	}

	if _, ok = rocAUC(predictions[:3]); ok {
		t.Fatal("AUC must be undefined without negatives")
	}
}

func TestPercentileNearestRank(t *testing.T) {
	t.Parallel()

	values := []float64{50, 10, 40, 20, 30}
	if got := percentile(values, 50); got != 30 {
		t.Fatalf("p50 = %v, want 30", got)
	}

	if got := percentile(values, 95); got != 50 {
		t.Fatalf("p95 = %v, want 50", got)
	}

	if values[0] != 50 {
		t.Fatal("percentile must not reorder its input")
	}
}

func TestRunEndToEndAgainstFakeJev(t *testing.T) {
	t.Parallel()

	data, err := loadDataset("data")
	if err != nil {
		t.Fatalf("load dataset: %v", err)
	}

	labels := make(map[string]sample, len(data.Samples))
	for _, s := range data.Samples {
		labels[s.Text] = s
	}

	// The fake scores each rule from the label, so the gate must pass on every variant.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request systemOneRequest
		if decodeErr := json.NewDecoder(r.Body).Decode(&request); decodeErr != nil {
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		text, ok := request.State["untrusted_content"].(string)
		if !ok {
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		s := labels[text]
		answers := make(map[string]systemOneAnswer, len(request.Questions))

		for key := range request.Questions {
			score := 0.05
			if s.isPositiveFor(key) {
				score = 0.95
			}

			answers[key] = systemOneAnswer{Type: "noul", Noul: &score}
		}

		_ = json.NewEncoder(w).Encode(systemOneResponse{Model: request.Model, Answers: answers})
	}))
	defer server.Close()

	cfg := config{
		dataDir:     "data",
		outDir:      t.TempDir(),
		baseURL:     server.URL,
		model:       "jev-1.13.0",
		variants:    []string{langEN, langFA},
		concurrency: 4,
		timeout:     5 * time.Second,
		maxFPR:      0.03,
		minRecall:   0.85,
		maxP95MS:    5000,
	}

	var out bytes.Buffer
	if err = run(context.Background(), cfg, "key", &out); err != nil {
		t.Fatalf("run: %v", err)
	}

	if !strings.Contains(out.String(), "gate: GO with variant en") {
		t.Fatalf("expected a passing gate, got:\n%s", out.String())
	}

	if _, err = os.Stat(filepath.Join(cfg.outDir, "predictions.jsonl")); err != nil {
		t.Fatalf("predictions file missing: %v", err)
	}
}

func TestRunRequiresAPIKey(t *testing.T) {
	t.Parallel()

	err := run(context.Background(), config{dataDir: "data"}, "", &bytes.Buffer{})
	if !errors.Is(err, errMissingAPIKey) {
		t.Fatalf("expected errMissingAPIKey, got %v", err)
	}
}

func writeTestFile(t *testing.T, path, body string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
}
