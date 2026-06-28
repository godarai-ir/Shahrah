// Package main provides a small benchmark CLI for metrics ingestion.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"sort"
	"time"
)

type queryCase struct {
	Name     string
	StartTS  string
	EndTS    string
	Interval int
}

func main() {
	baseURL := flag.String("base-url", "http://127.0.0.1:4000", "control-plane base url")
	tenant := flag.String("tenant", "demo", "tenant slug header")
	auth := flag.String("auth", "", "Bearer token")
	out := flag.String("out", "./tmp/model-metrics-benchmark.md", "output markdown path")
	runs := flag.Int("runs", 20, "runs per case")

	flag.Parse()

	now := time.Now().UTC()
	cases := []queryCase{
		{Name: "3h", StartTS: now.Add(-3 * time.Hour).Format(time.RFC3339), EndTS: now.Format(time.RFC3339), Interval: 300},
		{Name: "24h", StartTS: now.Add(-24 * time.Hour).Format(time.RFC3339), EndTS: now.Format(time.RFC3339), Interval: 900},
		{Name: "7d", StartTS: now.Add(-7 * 24 * time.Hour).Format(time.RFC3339), EndTS: now.Format(time.RFC3339), Interval: 3600},
	}

	results := make(map[string][]float64)
	client := &http.Client{Timeout: 15 * time.Second}

	for _, c := range cases {
		for i := 0; i < *runs; i++ {
			durMS, err := doQuery(client, *baseURL, *tenant, *auth, c)
			if err != nil {
				if reportErr := writeReport(*out, results, err); reportErr != nil {
					fmt.Println("write report failed:", reportErr)
				}

				fmt.Println("benchmark failed:", err)
				os.Exit(1)
			}

			results[c.Name] = append(results[c.Name], durMS)
		}
	}

	if err := writeReport(*out, results, nil); err != nil {
		fmt.Println("write report failed:", err)
		os.Exit(1)
	}

	fmt.Println("benchmark report written:", *out)
}

func doQuery(client *http.Client, baseURL, tenant, auth string, c queryCase) (float64, error) {
	reqBody := map[string]interface{}{
		"startTs":    c.StartTS,
		"endTs":      c.EndTS,
		"datasource": "modelMetrics",
		"type":       "timeseries",
		"aggregations": []map[string]string{
			{"type": "count", "column": "modelName"},
			{"type": "p99", "column": "latencyMs"},
		},
		"groupBy":           []string{"modelName"},
		"intervalInSeconds": c.Interval,
	}

	raw, err := json.Marshal(reqBody)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequest(http.MethodPost, baseURL+"/api/v1/metrics/query", bytes.NewReader(raw))
	if err != nil {
		return 0, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-shahrah-tenant", tenant)

	if auth != "" {
		req.Header.Set("Authorization", "Bearer "+auth)
	}

	start := time.Now()

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			return
		}
	}()

	if resp.StatusCode >= 400 {
		return 0, fmt.Errorf("status %d", resp.StatusCode)
	}

	return float64(time.Since(start).Milliseconds()), nil
}

func writeReport(path string, results map[string][]float64, benchErr error) error {
	var md bytes.Buffer
	md.WriteString("# Model Metrics Benchmark Report\n\n")
	fmt.Fprintf(&md, "Generated at: `%s`\n\n", time.Now().UTC().Format(time.RFC3339))

	if benchErr != nil {
		fmt.Fprintf(&md, "Benchmark error: `%v`\n\n", benchErr)
	}

	md.WriteString("| Window | p50 (ms) | p95 (ms) | p99 (ms) | samples |\n")
	md.WriteString("|---|---:|---:|---:|---:|\n")

	for _, k := range []string{"3h", "24h", "7d"} {
		values := append([]float64(nil), results[k]...)
		sort.Float64s(values)
		p50 := percentile(values, 50)
		p95 := percentile(values, 95)
		p99 := percentile(values, 99)
		fmt.Fprintf(&md, "| %s | %.2f | %.2f | %.2f | %d |\n", k, p50, p95, p99, len(values))
	}

	md.WriteString("\n## SLO Check\n")
	md.WriteString("- Target p95 3h < 800ms\n")
	md.WriteString("- Target p95 24h < 1500ms\n")
	md.WriteString("- Target p95 7d < 2000ms\n")

	if err := os.MkdirAll("./tmp", 0o755); err != nil {
		return err
	}

	return os.WriteFile(path, md.Bytes(), 0o644)
}

func percentile(values []float64, p int) float64 {
	if len(values) == 0 {
		return 0
	}

	idx := int(float64(len(values)-1) * float64(p) / 100.0)
	if idx < 0 {
		idx = 0
	}

	if idx >= len(values) {
		idx = len(values) - 1
	}

	return values[idx]
}
