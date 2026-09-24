// Command semantic-guardrail-eval measures how well Jev separates the semantic guardrail rules on a labeled
// Persian, Finglish and English evaluation set. It is the Phase 0 go/no-go check for the semantic_policy guardrail.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	controlconfig "github.com/Delpak/shahrah/services/control-plane/internal/config/environment"
)

const (
	apiKeyEnv = "TYPESAFE_API_KEY"

	// thresholdSpreadWarning flags rules whose best thresholds differ this much between language classes.
	thresholdSpreadWarning = 0.2
)

var errMissingAPIKey = errors.New(apiKeyEnv + " is not set")

type config struct {
	dataDir      string
	outDir       string
	baseURL      string
	model        string
	variants     []string
	concurrency  int
	timeout      time.Duration
	limit        int
	maxFPR       float64
	minRecall    float64
	maxP95MS     float64
	validateOnly bool
}

type job struct {
	variant string
	sample  sample
}

type jobResult struct {
	predictions []prediction
	latency     time.Duration
	inputTokens int
	err         error
	sampleID    string
	variant     string
}

type variantGate struct {
	Variant      string   `json:"variant"`
	Passed       bool     `json:"passed"`
	FailingRules []string `json:"failing_rules,omitempty"`
}

type gateResult struct {
	Passed        bool          `json:"passed"`
	BestVariant   string        `json:"best_variant,omitempty"`
	LatencyPassed bool          `json:"latency_passed"`
	Variants      []variantGate `json:"variants"`
}

type report struct {
	GeneratedAt    time.Time      `json:"generated_at"`
	Model          string         `json:"model"`
	DatasetVersion string         `json:"dataset_version"`
	Samples        int            `json:"samples"`
	Requests       int            `json:"requests"`
	Errors         []string       `json:"errors,omitempty"`
	InputTokens    int            `json:"input_tokens"`
	LatencyP50MS   float64        `json:"latency_p50_ms"`
	LatencyP95MS   float64        `json:"latency_p95_ms"`
	MaxFPR         float64        `json:"max_false_positive_rate"`
	MinRecall      float64        `json:"min_recall"`
	MaxP95MS       float64        `json:"max_latency_p95_ms"`
	Gate           gateResult     `json:"gate"`
	Groups         []groupMetrics `json:"groups"`
}

func main() {
	cfg := parseFlags()

	if err := run(context.Background(), cfg, controlconfig.Environment(apiKeyEnv), os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "semantic-guardrail-eval:", err)
		os.Exit(1)
	}
}

func parseFlags() config {
	var (
		cfg      config
		variants string
	)

	flag.StringVar(&cfg.dataDir, "data", "cmd/semantic-guardrail-eval/data", "directory holding rules.json and samples/*.jsonl")
	flag.StringVar(&cfg.outDir, "out", filepath.Join(os.TempDir(), "semantic-guardrail-eval"), "directory for report.json and predictions.jsonl")
	flag.StringVar(&cfg.baseURL, "base-url", "https://api.typesafe.ai", "TypeSafe API base URL")
	flag.StringVar(&cfg.model, "model", "jev-1.13.0", "pinned Jev model")
	flag.StringVar(&variants, "variants", "en,fa", "comma-separated rule languages to compare: en, fa")
	flag.IntVar(&cfg.concurrency, "concurrency", 4, "parallel requests")
	flag.DurationVar(&cfg.timeout, "timeout", 10*time.Second, "per-request timeout")
	flag.IntVar(&cfg.limit, "limit", 0, "evaluate only the first N samples; 0 means all")
	flag.Float64Var(&cfg.maxFPR, "max-fpr", 0.03, "maximum false positive rate for the go/no-go gate")
	flag.Float64Var(&cfg.minRecall, "min-recall", 0.85, "minimum Persian recall at max-fpr for the go/no-go gate")
	flag.Float64Var(&cfg.maxP95MS, "max-p95-ms", 400, "maximum p95 latency in milliseconds for the go/no-go gate")
	flag.BoolVar(&cfg.validateOnly, "validate-only", false, "validate the dataset and print its statistics without calling Jev")
	flag.Parse()

	cfg.variants = splitList(variants)

	return cfg
}

func run(ctx context.Context, cfg config, apiKey string, out io.Writer) error {
	data, err := loadDataset(cfg.dataDir)
	if err != nil {
		return err
	}

	printStats(out, data)

	if cfg.validateOnly {
		return nil
	}

	if apiKey == "" {
		return errMissingAPIKey
	}

	for _, variant := range cfg.variants {
		if variant != langEN && variant != langFA {
			return fmt.Errorf("unknown variant %q", variant)
		}
	}

	samples := data.Samples
	if cfg.limit > 0 && cfg.limit < len(samples) {
		samples = samples[:cfg.limit]
	}

	client := &jevClient{
		baseURL:     cfg.baseURL,
		apiKey:      apiKey,
		httpClient:  &http.Client{Timeout: cfg.timeout},
		maxAttempts: 4,
		backoff:     500 * time.Millisecond,
	}

	results := evaluateAll(ctx, cfg, client, data, samples)
	result := buildReport(cfg, data, len(samples), results)

	if writeErr := writeOutputs(cfg.outDir, result, results); writeErr != nil {
		return writeErr
	}

	printSummary(out, result)

	return nil
}

func evaluateAll(ctx context.Context, cfg config, client *jevClient, data dataset, samples []sample) []jobResult {
	jobs := make(chan job)
	results := make([]jobResult, 0, len(samples)*len(cfg.variants))

	var (
		mu sync.Mutex
		wg sync.WaitGroup
	)

	for range max(cfg.concurrency, 1) {
		wg.Go(func() {
			for j := range jobs {
				result := evaluateOne(ctx, client, cfg.model, data, j)

				mu.Lock()

				results = append(results, result)

				mu.Unlock()
			}
		})
	}

	for _, variant := range cfg.variants {
		for _, s := range samples {
			jobs <- job{variant: variant, sample: s}
		}
	}

	close(jobs)
	wg.Wait()

	return results
}

func evaluateOne(ctx context.Context, client *jevClient, model string, data dataset, j job) jobResult {
	rules := data.rulesFor(j.sample.Target)
	request := buildRequest(model, j.variant, rules, j.sample)
	result := jobResult{sampleID: j.sample.ID, variant: j.variant}

	started := time.Now()
	response, err := client.evaluate(ctx, request)
	result.latency = time.Since(started)

	if err != nil {
		result.err = err

		return result
	}

	result.inputTokens = response.Usage.InputTokens

	for _, r := range rules {
		answer, ok := response.Answers[r.Key]
		if !ok || answer.Noul == nil {
			result.err = fmt.Errorf("answer for rule %q is missing", r.Key)

			return result
		}

		result.predictions = append(result.predictions, prediction{
			SampleID: j.sample.ID,
			Variant:  j.variant,
			Rule:     r.Key,
			Register: j.sample.Register,
			Kind:     j.sample.Kind,
			Hard:     len(j.sample.Tags) > 0,
			Label:    j.sample.isPositiveFor(r.Key),
			Score:    *answer.Noul,
		})
	}

	return result
}

func buildReport(cfg config, data dataset, sampleCount int, results []jobResult) report {
	result := report{
		GeneratedAt:    time.Now().UTC(),
		Model:          cfg.model,
		DatasetVersion: data.Rules.Version,
		Samples:        sampleCount,
		Requests:       len(results),
		MaxFPR:         cfg.maxFPR,
		MinRecall:      cfg.minRecall,
		MaxP95MS:       cfg.maxP95MS,
	}

	predictions := make([]prediction, 0, len(results)*4)
	latencies := make([]float64, 0, len(results))

	for _, r := range results {
		if r.err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s/%s: %v", r.variant, r.sampleID, r.err))

			continue
		}

		predictions = append(predictions, r.predictions...)
		latencies = append(latencies, float64(r.latency.Milliseconds()))
		result.InputTokens += r.inputTokens
	}

	sort.Strings(result.Errors)

	ruleKeys := make([]string, 0, len(data.Rules.Rules))
	for _, r := range data.Rules.Rules {
		ruleKeys = append(ruleKeys, r.Key)
	}

	result.LatencyP50MS = percentile(latencies, 50)
	result.LatencyP95MS = percentile(latencies, 95)
	result.Groups = computeGroupMetrics(predictions, cfg.variants, ruleKeys, cfg.maxFPR)
	result.Gate = evaluateGate(result, cfg.variants, ruleKeys)

	return result
}

// evaluateGate passes when no request failed, latency is within budget, and at least one variant reaches the recall target on Persian text for every rule.
func evaluateGate(result report, variants, ruleKeys []string) gateResult {
	gate := gateResult{LatencyPassed: result.LatencyP95MS <= result.MaxP95MS}

	for _, variant := range variants {
		vg := variantGate{Variant: variant, Passed: true}

		for _, ruleKey := range ruleKeys {
			if !persianRecallMet(result.Groups, variant, ruleKey, result.MinRecall) {
				vg.Passed = false
				vg.FailingRules = append(vg.FailingRules, ruleKey)
			}
		}

		if vg.Passed && gate.BestVariant == "" {
			gate.BestVariant = variant
		}

		gate.Variants = append(gate.Variants, vg)
	}

	gate.Passed = gate.LatencyPassed && len(result.Errors) == 0 && gate.BestVariant != ""

	return gate
}

func persianRecallMet(groups []groupMetrics, variant, ruleKey string, minRecall float64) bool {
	for _, g := range groups {
		if g.Variant == variant && g.Rule == ruleKey && g.Group == groupPersian {
			return g.Best != nil && g.Best.Recall >= minRecall
		}
	}

	return false
}

func writeOutputs(dir string, result report, results []jobResult) error {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	reportBody, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("encode report: %w", err)
	}

	if err = os.WriteFile(filepath.Join(dir, "report.json"), append(reportBody, '\n'), 0o600); err != nil {
		return fmt.Errorf("write report: %w", err)
	}

	var lines strings.Builder

	for _, r := range results {
		for _, p := range r.predictions {
			line, marshalErr := json.Marshal(p)
			if marshalErr != nil {
				return fmt.Errorf("encode prediction: %w", marshalErr)
			}

			lines.Write(line)
			lines.WriteByte('\n')
		}
	}

	if err = os.WriteFile(filepath.Join(dir, "predictions.jsonl"), []byte(lines.String()), 0o600); err != nil {
		return fmt.Errorf("write predictions: %w", err)
	}

	return nil
}

func printStats(out io.Writer, data dataset) {
	counts := make(map[string]int)
	positives := make(map[string]map[string]int)

	for _, s := range data.Samples {
		counts[s.Target+"/"+s.Kind]++

		for _, key := range s.Positive {
			if positives[key] == nil {
				positives[key] = make(map[string]int)
			}

			positives[key][s.Register]++
		}
	}

	writef(out, "dataset %s: %d samples, %d rules\n", data.Rules.Version, len(data.Samples), len(data.Rules.Rules))

	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	for _, key := range keys {
		writef(out, "  %-24s %4d\n", key, counts[key])
	}

	writef(out, "positives per rule (%s):\n", strings.Join(validRegisters, " / "))

	for _, r := range data.Rules.Rules {
		cells := make([]string, 0, len(validRegisters))
		for _, register := range validRegisters {
			cells = append(cells, fmt.Sprint(positives[r.Key][register]))
		}

		writef(out, "  %-26s %s\n", r.Key, strings.Join(cells, " / "))
	}
}

func printSummary(out io.Writer, result report) {
	writef(out, "\nrequests %d, errors %d, input tokens %d, latency p50 %.0f ms, p95 %.0f ms\n",
		result.Requests, len(result.Errors), result.InputTokens, result.LatencyP50MS, result.LatencyP95MS)
	writef(out, "\n| variant | rule | group | AUC | recall @ FPR<=%.0f%% | threshold | recall @0.70 | FPR @0.70 |\n", result.MaxFPR*100)
	writef(out, "| --- | --- | --- | --- | --- | --- | --- | --- |\n")

	for _, g := range result.Groups {
		if g.Group != groupPersian && g.Group != groupHard && g.Group != "finglish" && g.Group != langEN {
			continue
		}

		auc, recall, threshold := "-", "-", "-"
		if g.AUC != nil {
			auc = fmt.Sprintf("%.3f", *g.AUC)
		}

		if g.Best != nil {
			recall = fmt.Sprintf("%.2f", g.Best.Recall)
			threshold = fmt.Sprintf("%.2f", g.Best.Threshold)
		}

		writef(out, "| %s | %s | %s | %s | %s | %s | %.2f | %.3f |\n",
			g.Variant, g.Rule, g.Group, auc, recall, threshold, g.AtBlock.Recall, g.AtBlock.FPR)
	}

	printThresholds(out, result.Groups)

	verdict := "NO-GO"
	if result.Gate.Passed {
		verdict = "GO with variant " + result.Gate.BestVariant
	}

	writef(out, "\ngate: %s (latency passed: %t)\n", verdict, result.Gate.LatencyPassed)

	for _, vg := range result.Gate.Variants {
		if !vg.Passed {
			writef(out, "  variant %s misses Persian recall on: %s\n", vg.Variant, strings.Join(vg.FailingRules, ", "))
		}
	}
}

// printThresholds lists the best threshold per language class, because one threshold for every language can
// sacrifice recall on the class whose scores run lower.
func printThresholds(out io.Writer, groups []groupMetrics) {
	classes := []string{groupPersian, "finglish", langEN}
	best := make(map[string]*operatingPoint, len(groups))

	var keys []string

	for i := range groups {
		g := groups[i]
		if !slices.Contains(classes, g.Group) {
			continue
		}

		key := g.Variant + "|" + g.Rule
		if !slices.Contains(keys, key) {
			keys = append(keys, key)
		}

		best[key+"|"+g.Group] = g.Best
	}

	writef(out, "\nthresholds at max FPR (persian / finglish / en):\n")

	for _, key := range keys {
		cells := make([]string, 0, len(classes))
		low, high := 1.0, 0.0

		for _, class := range classes {
			point := best[key+"|"+class]
			if point == nil {
				cells = append(cells, "-")

				continue
			}

			cells = append(cells, fmt.Sprintf("%.2f", point.Threshold))
			low, high = min(low, point.Threshold), max(high, point.Threshold)
		}

		note := ""
		if high-low > thresholdSpreadWarning {
			note = "  <- needs per-language thresholds"
		}

		writef(out, "  %-34s %s%s\n", strings.ReplaceAll(key, "|", " "), strings.Join(cells, " / "), note)
	}
}

func writef(out io.Writer, format string, args ...any) {
	if _, err := fmt.Fprintf(out, format, args...); err != nil {
		fmt.Fprintln(os.Stderr, "semantic-guardrail-eval: write output:", err)
	}
}

func splitList(value string) []string {
	var out []string

	for part := range strings.SplitSeq(value, ",") {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}

	return out
}
