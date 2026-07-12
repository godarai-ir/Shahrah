package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type config struct {
	URL               string
	APIKey            string
	Tenant            string
	Model             string
	ReportPath        string
	Machine           string
	WarmupDuration    time.Duration
	WarmupRPS         int
	RampStageDuration time.Duration
	RampBudget        time.Duration
	RampStartRPS      int
	MaxRampRPS        int
	StableDuration    time.Duration
	BurstDuration     time.Duration
	SafeFactor        float64
	BurstFactor       float64
	MaxErrorRate      float64
	MaxP95            time.Duration
	MaxHostCPU        float64
	MaxConcurrency    int
	RequestTimeout    time.Duration
	ResourceInterval  time.Duration
	FixedRPS          int
}

type stageResult struct {
	Name            string           `json:"name"`
	Kind            string           `json:"kind"`
	TargetRPS       int              `json:"target_rps"`
	DurationSeconds float64          `json:"duration_seconds"`
	Attempted       int64            `json:"attempted"`
	Completed       int64            `json:"completed"`
	Succeeded       int64            `json:"succeeded"`
	Failed          int64            `json:"failed"`
	Dropped         int64            `json:"dropped"`
	AchievedRPS     float64          `json:"achieved_rps"`
	ErrorRate       float64          `json:"error_rate"`
	P50Milliseconds float64          `json:"p50_ms"`
	P95Milliseconds float64          `json:"p95_ms"`
	P99Milliseconds float64          `json:"p99_ms"`
	MaxMilliseconds float64          `json:"max_ms"`
	StatusCounts    map[int]int64    `json:"status_counts"`
	Errors          map[string]int64 `json:"errors,omitempty"`
	Resources       resourceSummary  `json:"resources"`
	Pass            bool             `json:"pass"`
	FailureReasons  []string         `json:"failure_reasons,omitempty"`
}

type report struct {
	GeneratedAt       time.Time     `json:"generated_at"`
	Machine           string        `json:"machine"`
	GOOS              string        `json:"goos"`
	LogicalCPUs       int           `json:"logical_cpus"`
	Endpoint          string        `json:"endpoint"`
	Tenant            string        `json:"tenant"`
	Model             string        `json:"model"`
	Thresholds        thresholds    `json:"thresholds"`
	Stages            []stageResult `json:"stages"`
	MaxPassingRPS     int           `json:"max_passing_rps"`
	BreakPointRPS     int           `json:"break_point_rps,omitempty"`
	RecommendedRPS    int           `json:"recommended_rps"`
	StableVerified    bool          `json:"stable_verified"`
	GeneratorSameHost bool          `json:"generator_same_host"`
	Notes             []string      `json:"notes"`
}

type thresholds struct {
	MaxErrorRate float64 `json:"max_error_rate"`
	MaxP95MS     float64 `json:"max_p95_ms"`
	MaxHostCPU   float64 `json:"max_host_cpu_percent"`
	SafeFactor   float64 `json:"safe_factor"`
}

type requestResult struct {
	Latency time.Duration
	Status  int
	Err     string
}

type hostSample struct {
	CPUPercent           float64
	AvailableMemoryBytes uint64
	Valid                bool
}

type hostSampler interface {
	Sample() (hostSample, error)
}

type dockerStatLine struct {
	Name     string `json:"Name"`
	CPUPerc  string `json:"CPUPerc"`
	MemUsage string `json:"MemUsage"`
}

type resourceSample struct {
	Stage                string
	HostCPUPercent       float64
	AvailableMemoryBytes uint64
	HostValid            bool
	Docker               map[string]containerResource
}

type containerResource struct {
	CPUPercent float64 `json:"cpu_percent"`
	MemoryMiB  float64 `json:"memory_mib"`
}

type resourceSummary struct {
	Samples               int                                 `json:"samples"`
	AverageHostCPUPercent float64                             `json:"average_host_cpu_percent"`
	MaxHostCPUPercent     float64                             `json:"max_host_cpu_percent"`
	MinAvailableMemoryMiB float64                             `json:"min_available_memory_mib"`
	Containers            map[string]containerResourceSummary `json:"containers,omitempty"`
}

type containerResourceSummary struct {
	AverageCPUPercent float64 `json:"average_cpu_percent"`
	MaxCPUPercent     float64 `json:"max_cpu_percent"`
	MaxMemoryMiB      float64 `json:"max_memory_mib"`
}

type resourceMonitor struct {
	interval time.Duration
	stage    atomic.Value
	mu       sync.Mutex
	samples  []resourceSample
}

func main() {
	cfg := parseFlags()
	if err := validateConfig(cfg); err != nil {
		fmt.Fprintln(os.Stderr, "configuration error:", err)
		os.Exit(2)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	monitor := &resourceMonitor{interval: cfg.ResourceInterval}
	monitor.stage.Store("idle")
	go monitor.Run(ctx)

	client := &http.Client{
		Timeout: cfg.RequestTimeout,
		Transport: &http.Transport{
			MaxIdleConns:        cfg.MaxConcurrency,
			MaxIdleConnsPerHost: cfg.MaxConcurrency,
			MaxConnsPerHost:     cfg.MaxConcurrency,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	runner := &loadRunner{cfg: cfg, client: client, monitor: monitor}
	rep, err := runner.Run(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "load test failed:", err)
		os.Exit(1)
	}
	cancel()

	if err := writeReport(cfg.ReportPath, rep); err != nil {
		fmt.Fprintln(os.Stderr, "could not write report:", err)
		os.Exit(1)
	}
	fmt.Printf("\nReport written to %s\n", cfg.ReportPath)
	fmt.Printf("Estimated maximum passing rate: %d RPS\n", rep.MaxPassingRPS)
	fmt.Printf("Recommended rate with safety margin: %d RPS (stable verified: %t)\n", rep.RecommendedRPS, rep.StableVerified)
}

func parseFlags() config {
	var cfg config
	flag.StringVar(&cfg.URL, "url", "http://127.0.0.1:4000/v1/chat/completions", "chat completions endpoint")
	flag.StringVar(&cfg.APIKey, "api-key", os.Getenv("SHAHRAH_LOADTEST_API_KEY"), "benchmark API key")
	flag.StringVar(&cfg.Tenant, "tenant", "loadtest", "tenant header value")
	flag.StringVar(&cfg.Model, "model", "mock:loadtest:mock-model", "model alias")
	flag.StringVar(&cfg.ReportPath, "report", "artifacts/loadtest/report.json", "JSON report path")
	flag.StringVar(&cfg.Machine, "machine", "", "machine description included in the report")
	flag.DurationVar(&cfg.WarmupDuration, "warmup-duration", 2*time.Minute, "warm-up duration")
	flag.IntVar(&cfg.WarmupRPS, "warmup-rps", 20, "warm-up requests per second")
	flag.DurationVar(&cfg.RampStageDuration, "ramp-stage-duration", 45*time.Second, "duration of each ramp/refinement stage")
	flag.DurationVar(&cfg.RampBudget, "ramp-budget", 6*time.Minute, "maximum total ramp duration")
	flag.IntVar(&cfg.RampStartRPS, "ramp-start-rps", 1, "initial ramp rate")
	flag.IntVar(&cfg.MaxRampRPS, "max-ramp-rps", 1600, "maximum tested ramp rate")
	flag.DurationVar(&cfg.StableDuration, "stable-duration", 10*time.Minute, "sustained verification duration")
	flag.DurationVar(&cfg.BurstDuration, "burst-duration", 30*time.Second, "burst duration")
	flag.Float64Var(&cfg.SafeFactor, "safe-factor", 0.8, "recommended fraction of maximum passing RPS")
	flag.Float64Var(&cfg.BurstFactor, "burst-factor", 1.5, "burst multiplier over maximum passing RPS")
	flag.Float64Var(&cfg.MaxErrorRate, "max-error-rate", 0.01, "maximum acceptable error ratio")
	flag.DurationVar(&cfg.MaxP95, "max-p95", 500*time.Millisecond, "maximum acceptable p95 latency")
	flag.Float64Var(&cfg.MaxHostCPU, "max-host-cpu", 85, "maximum acceptable average host CPU percent")
	flag.IntVar(&cfg.MaxConcurrency, "max-concurrency", 4096, "maximum in-flight requests")
	flag.DurationVar(&cfg.RequestTimeout, "request-timeout", 10*time.Second, "per-request timeout")
	flag.DurationVar(&cfg.ResourceInterval, "resource-interval", time.Second, "host resource sampling interval")
	flag.IntVar(&cfg.FixedRPS, "fixed-rps", 0, "skip adaptive ramp and verify this fixed RPS")
	flag.Parse()
	return cfg
}

func validateConfig(cfg config) error {
	if cfg.APIKey == "" {
		return errors.New("--api-key or SHAHRAH_LOADTEST_API_KEY is required")
	}
	if cfg.RampStartRPS <= 0 || cfg.MaxRampRPS < cfg.RampStartRPS || cfg.MaxConcurrency <= 0 {
		return errors.New("RPS and concurrency values must be positive and ordered")
	}
	if cfg.SafeFactor <= 0 || cfg.SafeFactor > 1 || cfg.BurstFactor <= 1 {
		return errors.New("safe-factor must be in (0,1] and burst-factor must be greater than 1")
	}
	return nil
}

type loadRunner struct {
	cfg     config
	client  *http.Client
	monitor *resourceMonitor
	seq     atomic.Uint64
}

func (r *loadRunner) Run(ctx context.Context) (report, error) {
	rep := report{
		GeneratedAt: time.Now().UTC(), Machine: r.cfg.Machine, GOOS: runtime.GOOS,
		LogicalCPUs: runtime.NumCPU(), Endpoint: r.cfg.URL, Tenant: r.cfg.Tenant, Model: r.cfg.Model,
		Thresholds:        thresholds{MaxErrorRate: r.cfg.MaxErrorRate, MaxP95MS: ms(r.cfg.MaxP95), MaxHostCPU: r.cfg.MaxHostCPU, SafeFactor: r.cfg.SafeFactor},
		GeneratorSameHost: true,
		Notes: []string{
			"The load generator ran on the same host, so its CPU and memory overhead are included.",
			"Every request used unique prompt content to avoid response-cache hits.",
			"A stage passes only when errors, p95 latency, achieved rate, and average host CPU all satisfy thresholds.",
		},
	}

	if err := r.probe(ctx); err != nil {
		return rep, fmt.Errorf("preflight request: %w", err)
	}

	warmupRPS := r.cfg.WarmupRPS
	if r.cfg.FixedRPS > 0 {
		warmupRPS = r.cfg.FixedRPS
	}
	warmup := r.runStage(ctx, "warmup", "warmup", warmupRPS, r.cfg.WarmupDuration)
	rep.Stages = append(rep.Stages, warmup)
	if r.cfg.FixedRPS > 0 {
		rep.MaxPassingRPS = r.cfg.FixedRPS
		rep.RecommendedRPS = r.cfg.FixedRPS
		stable := r.runStage(ctx, "stable", "stable", r.cfg.FixedRPS, r.cfg.StableDuration)
		rep.Stages = append(rep.Stages, stable)
		rep.StableVerified = stable.Pass
		if !stable.Pass {
			rep.MaxPassingRPS = 0
			rep.RecommendedRPS = 0
		}
		burstRPS := maxInt(r.cfg.FixedRPS+1, int(math.Ceil(float64(r.cfg.FixedRPS)*r.cfg.BurstFactor)))
		rep.Stages = append(rep.Stages, r.runStage(ctx, "burst", "burst", burstRPS, r.cfg.BurstDuration))
		return rep, nil
	}

	rampDeadline := time.Now().Add(r.cfg.RampBudget)
	lastPass := 0
	firstFail := 0
	for target := r.cfg.RampStartRPS; target <= r.cfg.MaxRampRPS && time.Now().Add(r.cfg.RampStageDuration).Before(rampDeadline); target *= 2 {
		stage := r.runStage(ctx, fmt.Sprintf("ramp-%d", target), "ramp", target, r.cfg.RampStageDuration)
		rep.Stages = append(rep.Stages, stage)
		if stage.Pass {
			lastPass = target
			continue
		}
		firstFail = target
		break
	}

	if firstFail > 0 {
		low := lastPass
		if low == 0 {
			low = maxInt(1, firstFail/2)
		}
		high := firstFail
		for high-low > maxInt(5, low/10) && time.Now().Add(r.cfg.RampStageDuration).Before(rampDeadline) {
			mid := (low + high) / 2
			stage := r.runStage(ctx, fmt.Sprintf("refine-%d", mid), "refine", mid, r.cfg.RampStageDuration)
			rep.Stages = append(rep.Stages, stage)
			if stage.Pass {
				low, lastPass = mid, mid
			} else {
				high, firstFail = mid, mid
			}
		}
	}

	if lastPass == 0 {
		rep.BreakPointRPS = firstFail
		rep.Notes = append(rep.Notes, "No ramp stage met every threshold, so no sustainable rate can be recommended from this run.")
		return rep, nil
	}
	rep.MaxPassingRPS = lastPass
	rep.BreakPointRPS = firstFail
	rep.RecommendedRPS = maxInt(1, int(math.Floor(float64(lastPass)*r.cfg.SafeFactor)))

	stable := r.runStage(ctx, "stable", "stable", rep.RecommendedRPS, r.cfg.StableDuration)
	rep.Stages = append(rep.Stages, stable)
	rep.StableVerified = stable.Pass
	if !stable.Pass {
		rep.Notes = append(rep.Notes, "The 80% safety target did not pass the sustained stage; use a lower rate and rerun before production sizing.")
	}

	burstRPS := maxInt(rep.RecommendedRPS+1, int(math.Ceil(float64(lastPass)*r.cfg.BurstFactor)))
	burst := r.runStage(ctx, "burst", "burst", burstRPS, r.cfg.BurstDuration)
	rep.Stages = append(rep.Stages, burst)
	return rep, nil
}

func (r *loadRunner) probe(ctx context.Context) error {
	result := r.doRequest(ctx)
	if result.Err != "" {
		return errors.New(result.Err)
	}
	if result.Status != http.StatusOK {
		return fmt.Errorf("unexpected HTTP status %d", result.Status)
	}
	return nil
}

func (r *loadRunner) runStage(ctx context.Context, name, kind string, targetRPS int, duration time.Duration) stageResult {
	fmt.Printf("\n[%s] target=%d RPS duration=%s\n", name, targetRPS, duration)
	r.monitor.stage.Store(name)
	time.Sleep(minDuration(r.cfg.ResourceInterval, 500*time.Millisecond))

	start := time.Now()
	deadline := start.Add(duration)
	results := make(chan requestResult, maxInt(1024, targetRPS*2))
	collected := make([]requestResult, 0, int(math.Ceil(duration.Seconds()*float64(targetRPS))))
	collectorDone := make(chan struct{})
	go func() {
		for item := range results {
			collected = append(collected, item)
		}
		close(collectorDone)
	}()
	sem := make(chan struct{}, r.cfg.MaxConcurrency)
	var wg sync.WaitGroup
	var scheduled, dropped int64
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	dispatchUntil := func(expected int64) {
		for scheduled < expected {
			scheduled++
			select {
			case sem <- struct{}{}:
				wg.Add(1)
				go func() {
					defer wg.Done()
					defer func() { <-sem }()
					results <- r.doRequest(ctx)
				}()
			default:
				dropped++
			}
		}
	}

	for now := range ticker.C {
		if !now.Before(deadline) {
			break
		}
		expected := int64(math.Floor(now.Sub(start).Seconds() * float64(targetRPS)))
		dispatchUntil(expected)
	}
	dispatchUntil(int64(math.Floor(duration.Seconds() * float64(targetRPS))))
	go func() { wg.Wait(); close(results) }()
	<-collectorDone

	latencies := make([]time.Duration, 0, scheduled-dropped)
	statusCounts := map[int]int64{}
	errorCounts := map[string]int64{}
	var succeeded, failed int64
	for _, item := range collected {
		latencies = append(latencies, item.Latency)
		if item.Status != 0 {
			statusCounts[item.Status]++
		}
		if item.Err == "" && item.Status == http.StatusOK {
			succeeded++
		} else {
			failed++
			message := item.Err
			if message == "" {
				message = fmt.Sprintf("http_%d", item.Status)
			}
			errorCounts[message]++
		}
	}

	elapsed := time.Since(start)
	result := stageResult{
		Name: name, Kind: kind, TargetRPS: targetRPS, DurationSeconds: elapsed.Seconds(), Attempted: scheduled,
		Completed: scheduled - dropped, Succeeded: succeeded, Failed: failed, Dropped: dropped,
		AchievedRPS: float64(succeeded) / elapsed.Seconds(), StatusCounts: statusCounts, Errors: errorCounts,
	}
	if scheduled > 0 {
		result.ErrorRate = float64(failed+dropped) / float64(scheduled)
	}
	if len(latencies) > 0 {
		sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
		result.P50Milliseconds = percentileMS(latencies, .50)
		result.P95Milliseconds = percentileMS(latencies, .95)
		result.P99Milliseconds = percentileMS(latencies, .99)
		result.MaxMilliseconds = ms(latencies[len(latencies)-1])
	}
	time.Sleep(minDuration(r.cfg.ResourceInterval, 500*time.Millisecond))
	result.Resources = r.monitor.Summary(name)
	result.Pass, result.FailureReasons = evaluate(result, r.cfg)
	fmt.Printf("[%s] achieved=%.1f RPS errors=%.2f%% p95=%.1fms avgCPU=%.1f%% pass=%t\n",
		name, result.AchievedRPS, result.ErrorRate*100, result.P95Milliseconds, result.Resources.AverageHostCPUPercent, result.Pass)
	return result
}

func (r *loadRunner) doRequest(ctx context.Context) requestResult {
	sequence := r.seq.Add(1)
	payload := fmt.Sprintf(`{"model":%q,"messages":[{"role":"user","content":%q}],"stream":false}`,
		r.cfg.Model, fmt.Sprintf("Shahrah load test request %d", sequence))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.cfg.URL, bytes.NewBufferString(payload))
	if err != nil {
		return requestResult{Err: err.Error()}
	}
	req.Header.Set("Authorization", "Bearer "+r.cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Shahrah-Tenant", r.cfg.Tenant)
	start := time.Now()
	resp, err := r.client.Do(req)
	latency := time.Since(start)
	if err != nil {
		return requestResult{Latency: latency, Err: normalizeError(err.Error())}
	}
	defer resp.Body.Close()
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if readErr != nil {
		return requestResult{Latency: latency, Status: resp.StatusCode, Err: "read_response"}
	}
	if resp.StatusCode == http.StatusOK && !bytes.Contains(body, []byte("Mock provider response")) {
		return requestResult{Latency: latency, Status: resp.StatusCode, Err: "unexpected_response_body"}
	}
	return requestResult{Latency: latency, Status: resp.StatusCode}
}

func evaluate(result stageResult, cfg config) (bool, []string) {
	reasons := make([]string, 0, 4)
	if result.ErrorRate >= cfg.MaxErrorRate {
		reasons = append(reasons, fmt.Sprintf("error rate %.2f%% >= %.2f%%", result.ErrorRate*100, cfg.MaxErrorRate*100))
	}
	if result.P95Milliseconds > ms(cfg.MaxP95) {
		reasons = append(reasons, fmt.Sprintf("p95 %.1fms > %.1fms", result.P95Milliseconds, ms(cfg.MaxP95)))
	}
	if result.AchievedRPS < float64(result.TargetRPS)*0.98 {
		reasons = append(reasons, fmt.Sprintf("achieved rate %.1f < 98%% of target", result.AchievedRPS))
	}
	if result.Resources.Samples > 0 && result.Resources.AverageHostCPUPercent > cfg.MaxHostCPU {
		reasons = append(reasons, fmt.Sprintf("average host CPU %.1f%% > %.1f%%", result.Resources.AverageHostCPUPercent, cfg.MaxHostCPU))
	}
	return len(reasons) == 0, reasons
}

func (m *resourceMonitor) Run(ctx context.Context) {
	go m.runHost(ctx)
	go m.runDocker(ctx)
	<-ctx.Done()
}

func (m *resourceMonitor) runHost(ctx context.Context) {
	sampler := newHostSampler()
	_, _ = sampler.Sample()
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			host, _ := sampler.Sample()
			sample := resourceSample{Stage: m.stage.Load().(string)}
			if host.Valid {
				sample.HostCPUPercent = host.CPUPercent
				sample.AvailableMemoryBytes = host.AvailableMemoryBytes
				sample.HostValid = true
			}
			m.mu.Lock()
			m.samples = append(m.samples, sample)
			m.mu.Unlock()
		}
	}
}

func (m *resourceMonitor) runDocker(ctx context.Context) {
	interval := 10 * time.Second
	if m.interval > interval {
		interval = m.interval
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sample := resourceSample{Stage: m.stage.Load().(string), Docker: sampleDocker()}
			m.mu.Lock()
			m.samples = append(m.samples, sample)
			m.mu.Unlock()
		}
	}
}

func (m *resourceMonitor) Summary(stage string) resourceSummary {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := resourceSummary{Containers: map[string]containerResourceSummary{}}
	containerCounts := map[string]int{}
	hostSamples := 0
	minAvailable := uint64(math.MaxUint64)
	for _, sample := range m.samples {
		if sample.Stage != stage {
			continue
		}
		if sample.HostValid {
			hostSamples++
			result.AverageHostCPUPercent += sample.HostCPUPercent
			result.MaxHostCPUPercent = math.Max(result.MaxHostCPUPercent, sample.HostCPUPercent)
			if sample.AvailableMemoryBytes > 0 && sample.AvailableMemoryBytes < minAvailable {
				minAvailable = sample.AvailableMemoryBytes
			}
		}
		for name, current := range sample.Docker {
			agg := result.Containers[name]
			agg.AverageCPUPercent += current.CPUPercent
			agg.MaxCPUPercent = math.Max(agg.MaxCPUPercent, current.CPUPercent)
			agg.MaxMemoryMiB = math.Max(agg.MaxMemoryMiB, current.MemoryMiB)
			result.Containers[name] = agg
			containerCounts[name]++
		}
	}
	result.Samples = hostSamples
	if hostSamples > 0 {
		result.AverageHostCPUPercent /= float64(hostSamples)
	}
	if minAvailable != math.MaxUint64 {
		result.MinAvailableMemoryMiB = float64(minAvailable) / 1024 / 1024
	}
	for name, agg := range result.Containers {
		if containerCounts[name] > 0 {
			agg.AverageCPUPercent /= float64(containerCounts[name])
			result.Containers[name] = agg
		}
	}
	return result
}

func sampleDocker() map[string]containerResource {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	output, err := exec.CommandContext(ctx, "docker", "stats", "--no-stream", "--format", "{{json .}}").Output()
	if err != nil {
		return nil
	}
	result := map[string]containerResource{}
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		var item dockerStatLine
		if json.Unmarshal([]byte(line), &item) != nil || !strings.HasPrefix(item.Name, "compose-") {
			continue
		}
		name := strings.TrimSuffix(strings.TrimPrefix(item.Name, "compose-"), "-1")
		result[name] = containerResource{CPUPercent: parsePercent(item.CPUPerc), MemoryMiB: parseMemoryMiB(strings.Split(item.MemUsage, "/")[0])}
	}
	return result
}

func parsePercent(value string) float64 {
	parsed, _ := strconv.ParseFloat(strings.TrimSpace(strings.TrimSuffix(value, "%")), 64)
	return parsed
}

func parseMemoryMiB(value string) float64 {
	value = strings.TrimSpace(value)
	units := map[string]float64{"B": 1.0 / 1024 / 1024, "KiB": 1.0 / 1024, "MiB": 1, "GiB": 1024, "KB": 1.0 / 1000, "MB": 1.0 / 1000 * 1024, "GB": 1000}
	for unit, factor := range units {
		if strings.HasSuffix(value, unit) {
			number, _ := strconv.ParseFloat(strings.TrimSpace(strings.TrimSuffix(value, unit)), 64)
			return number * factor
		}
	}
	return 0
}

func writeReport(path string, rep report) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(rep, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, append(raw, '\n'), 0o644); err != nil {
		return err
	}
	return os.WriteFile(strings.TrimSuffix(path, filepath.Ext(path))+".md", []byte(markdownReport(rep)), 0o644)
}

func markdownReport(rep report) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Shahrah gateway load-test report\n\nGenerated: %s\n\n", rep.GeneratedAt.Format(time.RFC3339))
	fmt.Fprintf(&b, "- Machine: %s\n- Logical CPUs: %d\n- Endpoint: `%s`\n- Generator on same host: yes\n\n", rep.Machine, rep.LogicalCPUs, rep.Endpoint)
	fmt.Fprintf(&b, "## Capacity estimate\n\n- Maximum passing ramp rate: **%d RPS**\n", rep.MaxPassingRPS)
	if rep.BreakPointRPS > 0 {
		fmt.Fprintf(&b, "- First/refined failing rate: **%d RPS**\n", rep.BreakPointRPS)
	} else {
		fmt.Fprintf(&b, "- Break point: not reached within the configured maximum\n")
	}
	fmt.Fprintf(&b, "- Recommended rate with %.0f%% safety factor: **%d RPS**\n- Sustained verification passed: **%t**\n\n", rep.Thresholds.SafeFactor*100, rep.RecommendedRPS, rep.StableVerified)
	b.WriteString("## Stages\n\n| Stage | Target RPS | Achieved RPS | Errors | p95 | Avg host CPU | Result |\n|---|---:|---:|---:|---:|---:|---|\n")
	for _, stage := range rep.Stages {
		fmt.Fprintf(&b, "| %s | %d | %.1f | %.2f%% | %.1f ms | %.1f%% | %s |\n", stage.Name, stage.TargetRPS, stage.AchievedRPS, stage.ErrorRate*100, stage.P95Milliseconds, stage.Resources.AverageHostCPUPercent, map[bool]string{true: "pass", false: "fail"}[stage.Pass])
	}
	b.WriteString("\n## Notes\n\n")
	for _, note := range rep.Notes {
		fmt.Fprintf(&b, "- %s\n", note)
	}
	return b.String()
}

func percentileMS(values []time.Duration, percentile float64) float64 {
	index := int(math.Ceil(percentile*float64(len(values)))) - 1
	index = maxInt(0, minInt(index, len(values)-1))
	return ms(values[index])
}

func normalizeError(value string) string {
	lower := strings.ToLower(value)
	for _, item := range []struct{ match, label string }{{"timeout", "timeout"}, {"connection refused", "connection_refused"}, {"reset by peer", "connection_reset"}} {
		if strings.Contains(lower, item.match) {
			return item.label
		}
	}
	if len(value) > 160 {
		return value[:160]
	}
	return value
}

func ms(value time.Duration) float64 { return float64(value) / float64(time.Millisecond) }
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
