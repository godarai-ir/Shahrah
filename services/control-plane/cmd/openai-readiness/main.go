// Command openai-readiness reconciles the SHA-pinned official OpenAI route
// baseline with routes registered by GodarAI. It performs no live OpenAI call.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	officialCommit = "db14b6e1712aaf5265cf5a6871adff7a9c61d31c"
	officialDate   = "2026-07-28T20:57:12Z"
)

type baseline struct {
	Baseline struct {
		Repository  string `json:"repository"`
		Commit      string `json:"commit"`
		CommittedAt string `json:"committed_at"`
	} `json:"baseline"`
	Operations []string `json:"operations"`
}

type readinessReport struct {
	GeneratedAt       time.Time `json:"generated_at"`
	OfficialCommit    string    `json:"official_commit"`
	OfficialDate      string    `json:"official_date"`
	OfficialCount     int       `json:"official_operation_count"`
	ImplementedCount  int       `json:"implemented_operation_count"`
	Missing           []string  `json:"missing_operations"`
	Extra             []string  `json:"extra_operations"`
	RouteReady        bool      `json:"route_ready"`
	ContractReady     bool      `json:"contract_ready"`
	WebSocketReady    bool      `json:"websocket_ready"`
	ProductionReady   bool      `json:"production_ready"`
	LiveValidated     bool      `json:"live_validated"`
	ModelSnapshotPath string    `json:"model_snapshot_path,omitempty"`
}

var routePattern = regexp.MustCompile(`v1\.(Get|Post|Put|Patch|Delete)\("([^"]+)"`)
var parameterPattern = regexp.MustCompile(`(:[^/]+|\{[^/}]+\})`)

func main() {
	var (
		baselinePath, routesPath, reportPath, snapshotPath string
		allowIncomplete                                    bool
	)

	flag.StringVar(&baselinePath, "baseline", "../../packages/contracts/openapi/openai-scoped-routes-db14b6e.json", "SHA-pinned scoped OpenAI route baseline")
	flag.StringVar(&routesPath, "routes", "internal/server/routes_proxy.go", "GodarAI route registration source")
	flag.StringVar(&reportPath, "report", "openai-readiness-report.json", "JSON report output")
	flag.StringVar(&snapshotPath, "model-snapshot", "", "optional exported Model Discovery OpenAI snapshot")
	flag.BoolVar(&allowIncomplete, "allow-incomplete", false, "write an incomplete report without a failing exit status")
	flag.Parse()

	report, err := auditFiles(baselinePath, routesPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	report.ModelSnapshotPath = snapshotPath

	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	if err := os.WriteFile(reportPath, encoded, 0o600); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	fmt.Printf("OpenAI contract readiness: %t (%d/%d operations), report=%s\n",
		report.ContractReady, report.ImplementedCount, report.OfficialCount, reportPath)

	if !report.ContractReady && !allowIncomplete {
		os.Exit(1)
	}
}

//nolint:govet // Short-lived branch errors are checked immediately and intentionally scoped with short declarations.
func auditFiles(baselinePath, routesPath string) (readinessReport, error) {
	var out readinessReport

	baselineBytes, err := os.ReadFile(baselinePath)
	if err != nil {
		return out, err
	}

	var source baseline
	if err := json.Unmarshal(baselineBytes, &source); err != nil {
		return out, err
	}

	if source.Baseline.Commit != officialCommit ||
		source.Baseline.CommittedAt != officialDate {
		return out, errors.New("OpenAI baseline commit/date does not match the compiled release gate")
	}

	routes, err := os.ReadFile(routesPath)
	if err != nil {
		return out, err
	}

	return audit(source.Operations, string(routes)), nil
}

func audit(official []string, routeSource string) readinessReport {
	expected := responsesOperationSet(official)
	implementedRoutes := make([]string, 0, 7)
	// Anthropic-native fall-through routes share several public path strings
	// but are not OpenAI implementations. Only count the OpenAI-compatible
	// registry segment; a future OpenAINativeHandler must be registered there.
	openAISource := strings.SplitN(routeSource, "if deps.AnthropicNativeHandler", 2)[0]
	for _, match := range routePattern.FindAllStringSubmatch(openAISource, -1) {
		implementedRoutes = append(implementedRoutes, strings.ToUpper(match[1])+" "+match[2])
	}

	implemented := responsesOperationSet(implementedRoutes)
	missing := difference(expected, implemented)
	extra := difference(implemented, expected)

	return readinessReport{
		GeneratedAt:      time.Now().UTC(),
		OfficialCommit:   officialCommit,
		OfficialDate:     officialDate,
		OfficialCount:    len(expected),
		ImplementedCount: len(expected) - len(missing),
		Missing:          missing,
		Extra:            extra,
		RouteReady:       len(missing) == 0,
		ContractReady:    len(missing) == 0,
		WebSocketReady:   false,
		ProductionReady:  false,
		LiveValidated:    false,
	}
}

func responsesOperationSet(values []string) map[string]struct{} {
	all := operationSet(values)
	out := make(map[string]struct{}, 7)

	for operation := range all {
		parts := strings.SplitN(operation, " ", 2)
		if len(parts) == 2 &&
			(parts[1] == "/responses" || strings.HasPrefix(parts[1], "/responses/")) {
			out[operation] = struct{}{}
		}
	}

	return out
}

func operationSet(values []string) map[string]struct{} {
	out := make(map[string]struct{}, len(values))
	for _, value := range values {
		if normalized := normalizeOperation(value); normalized != "" {
			out[normalized] = struct{}{}
		}
	}

	return out
}

func normalizeOperation(value string) string {
	parts := strings.SplitN(strings.TrimSpace(value), " ", 2)
	if len(parts) != 2 {
		return ""
	}

	path := strings.SplitN(strings.TrimSpace(parts[1]), "?", 2)[0]
	path = parameterPattern.ReplaceAllString(path, "{}")

	return strings.ToUpper(parts[0]) + " " + path
}

func difference(left, right map[string]struct{}) []string {
	out := make([]string, 0)

	for value := range left {
		if _, ok := right[value]; !ok {
			out = append(out, value)
		}
	}

	sort.Strings(out)

	return out
}
