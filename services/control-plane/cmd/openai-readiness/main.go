// Command openai-readiness reconciles the SHA-pinned official OpenAI route
// baseline with routes registered by GodarAI. It performs no live OpenAI call.
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
	"regexp"
	"sort"
	"strings"
	"time"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
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
	ContractSnapshot  string    `json:"contract_snapshot_date"`
	ResponseFields    int       `json:"response_field_count"`
	OutputVariants    int       `json:"response_output_variant_count"`
	ResponseItems     int       `json:"response_item_variant_count"`
	CompactedFields   int       `json:"compacted_response_field_count"`
	InputTokenFields  int       `json:"input_token_response_field_count"`
	InputItemsFields  int       `json:"input_items_page_field_count"`
	StreamingEvents   int       `json:"responses_streaming_event_count"`
	ResponseStatuses  int       `json:"response_status_count"`
	SnapshotCoverage  float64   `json:"snapshot_coverage_percent"`
	OfficialVerified  bool      `json:"official_docs_verified"`
	VerificationError string    `json:"official_docs_verification_error,omitempty"`
	WebSocketReady    bool      `json:"websocket_ready"`
	ProductionReady   bool      `json:"production_ready"`
	LiveValidated     bool      `json:"live_validated"`
	ModelSnapshotPath string    `json:"model_snapshot_path,omitempty"`
}

var routePattern = regexp.MustCompile(`v1\.(Get|Post|Put|Patch|Delete)\("([^"]+)"`)
var parameterPattern = regexp.MustCompile(`(:[^/]+|\{[^/}]+\})`)
var responseFieldPattern = regexp.MustCompile("^ {2}- `([a-z][a-z0-9_]*):")
var responseFieldContractPattern = regexp.MustCompile("^ {2}- `([a-z][a-z0-9_]*): (.+)`$")
var responseOutputMemberPattern = regexp.MustCompile("^ {4}- `([^`]+) object")
var responseOutputTypePattern = regexp.MustCompile("^ {6}- `type: \"([^\"]+)\"")
var responseStatusValuePattern = regexp.MustCompile("^ {4}- `\"([^\"]+)\"`")
var responseEventHeadingPattern = regexp.MustCompile(`^## (response\.[a-z0-9_.]+|error)$`)
var sdkOutputTypePattern = regexp.MustCompile(`^(\s*)- ` + "`" + `type: (?:Optional\[)?Literal\["([^"]+)"\]` + "`" + `$`)
var sdkOperationPattern = regexp.MustCompile("^\\*\\*([a-z]+)\\*\\* `(/responses[^`]*)`$")

func main() {
	var (
		baselinePath, routesPath, reportPath, snapshotPath string
		allowIncomplete                                    bool
		verifyOfficialDocs                                 bool
	)

	flag.StringVar(&baselinePath, "baseline", "../../packages/contracts/openapi/openai-scoped-routes-db14b6e.json", "SHA-pinned scoped OpenAI route baseline")
	flag.StringVar(&routesPath, "routes", "internal/server/routes_proxy.go", "GodarAI route registration source")
	flag.StringVar(&reportPath, "report", "openai-readiness-report.json", "JSON report output")
	flag.StringVar(&snapshotPath, "model-snapshot", "", "optional exported Model Discovery OpenAI snapshot")
	flag.BoolVar(&allowIncomplete, "allow-incomplete", false, "write an incomplete report without a failing exit status")
	flag.BoolVar(&verifyOfficialDocs, "verify-official-docs", true, "reconcile the checked-in Responses contract inventory with official OpenAI documentation")
	flag.Parse()

	report, err := auditFiles(baselinePath, routesPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	report.ModelSnapshotPath = snapshotPath

	if verifyOfficialDocs {
		verifyCtx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		verifyErr := verifyOfficialResponsesContract(verifyCtx, &http.Client{Timeout: 2 * time.Minute})

		cancel()

		if verifyErr != nil {
			report.VerificationError = verifyErr.Error()
		} else {
			report.OfficialVerified = true
			report.SnapshotCoverage = 100
		}
	}

	report.ContractReady = report.RouteReady && report.OfficialVerified && report.SnapshotCoverage == 100

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

	if err := verifyPinnedResponsesOperations(source.Operations); err != nil {
		return out, err
	}

	routes, err := os.ReadFile(routesPath)
	if err != nil {
		return out, err
	}

	return audit(source.Operations, string(routes)), nil
}

func verifyPinnedResponsesOperations(official []string) error {
	pinned := stringSetValues(responsesOperationSet(official))
	contract := stringSetValues(operationSet(contractsopenai.ResponsesOperations()))

	return requireExactInventory("pinned operation baseline", contract, pinned)
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
		ContractReady:    false,
		ContractSnapshot: contractsopenai.ResponsesContractSnapshotDate,
		ResponseFields:   len(contractsopenai.ResponsesTopLevelFields()),
		OutputVariants:   len(contractsopenai.ResponsesOutputVariants()),
		ResponseItems:    len(contractsopenai.ResponsesItemVariants()),
		CompactedFields:  len(contractsopenai.ResponsesCompactedFields()),
		InputTokenFields: len(contractsopenai.ResponsesInputTokenCountFields()),
		InputItemsFields: len(contractsopenai.ResponsesInputItemsPageFields()),
		StreamingEvents:  len(contractsopenai.ResponsesStreamingEventTypes()),
		ResponseStatuses: len(contractsopenai.ResponsesStatuses()),
		SnapshotCoverage: 0,
		WebSocketReady:   false,
		ProductionReady:  false,
		LiveValidated:    false,
	}
}

func verifyOfficialResponsesContract(ctx context.Context, client *http.Client) error {
	createMarkdown, err := fetchOfficialMarkdown(ctx, client, contractsopenai.ResponsesCreateReferenceURL+".md")
	if err != nil {
		return fmt.Errorf("fetch official Responses create reference: %w", err)
	}

	streamingMarkdown, err := fetchOfficialMarkdown(ctx, client, contractsopenai.ResponsesStreamingEventsReferenceURL+".md")
	if err != nil {
		return fmt.Errorf("fetch official Responses streaming reference: %w", err)
	}

	sdkMarkdown, err := fetchOfficialMarkdown(ctx, client, contractsopenai.ResponsesAPIReferenceURL+".md")
	if err != nil {
		return fmt.Errorf("fetch official Responses SDK reference: %w", err)
	}

	inputItemsMarkdown, err := fetchOfficialMarkdown(ctx, client, contractsopenai.ResponsesInputItemsReferenceURL+".md")
	if err != nil {
		return fmt.Errorf("fetch official Responses input-items reference: %w", err)
	}

	parsed, err := parseOfficialResponsesContract(createMarkdown, streamingMarkdown)
	if err != nil {
		return err
	}

	checks := []struct {
		label    string
		expected []string
		actual   []string
	}{
		{
			label:    "operations",
			expected: contractsopenai.ResponsesOperations(),
			actual:   parseSDKOperations(sdkMarkdown),
		},
		{label: "top-level fields", expected: contractsopenai.ResponsesTopLevelFields(), actual: parsed.fields},
		{label: "output variants", expected: responseOutputVariantNames(), actual: parsed.outputVariants},
		{label: "streaming events", expected: contractsopenai.ResponsesStreamingEventTypes(), actual: parsed.streamingEvents},
		{label: "statuses", expected: contractsopenai.ResponsesStatuses(), actual: parsed.statuses},
		{
			label:    "compacted response fields",
			expected: contractsopenai.ResponsesCompactedFields(),
			actual:   parseSDKReturnFields(sdkMarkdown, "## Compact a response", "CompactedResponse"),
		},
		{
			label:    "input token response fields",
			expected: contractsopenai.ResponsesInputTokenCountFields(),
			actual:   parseSDKReturnFields(sdkMarkdown, "## Get input token counts", "InputTokenCountResponse"),
		},
		{
			label:    "response item variants",
			expected: responseItemVariantSDKNames(),
			actual:   parseSDKResponseItemVariants(sdkMarkdown),
		},
		{
			label:    "input items page fields",
			expected: contractsopenai.ResponsesInputItemsPageFields(),
			actual:   parseResourceReturnFields(inputItemsMarkdown),
		},
	}
	for _, check := range checks {
		if err := requireExactInventory(check.label, check.expected, check.actual); err != nil {
			return err
		}
	}

	if err := verifyOutputVariantDiscriminators(createMarkdown, sdkMarkdown); err != nil {
		return err
	}

	if err := verifyResponseFieldContracts(parsed.fieldContracts); err != nil {
		return err
	}

	return verifyResponseItemVariants(inputItemsMarkdown)
}

func parseSDKOperations(markdown string) []string {
	lines := strings.Split(strings.ReplaceAll(markdown, "\r\n", "\n"), "\n")
	operations := make([]string, 0, len(contractsopenai.ResponsesOperations()))

	for _, line := range lines {
		match := sdkOperationPattern.FindStringSubmatch(line)
		if len(match) == 3 {
			operations = append(operations, strings.ToUpper(match[1])+" "+match[2])
		}
	}

	return operations
}

func verifyResponseFieldContracts(actual map[string]string) error {
	for _, expected := range contractsopenai.ResponsesFieldContracts() {
		actualExpression, ok := actual[expected.Name]
		if !ok {
			return fmt.Errorf("official Response field contract is missing for %s", expected.Name)
		}

		if actualExpression != expected.TypeExpression {
			return fmt.Errorf("official Response field type drifted for %s: expected=%q actual=%q", expected.Name, expected.TypeExpression, actualExpression)
		}

		actualOptional := strings.HasPrefix(actualExpression, "optional ")

		actualNullable := strings.Contains(actualExpression, "null")
		if actualOptional != expected.Optional || actualNullable != expected.Nullable {
			return fmt.Errorf(
				"official Response field presence drifted for %s: expected_optional=%t actual_optional=%t expected_nullable=%t actual_nullable=%t",
				expected.Name,
				expected.Optional,
				actualOptional,
				expected.Nullable,
				actualNullable,
			)
		}
	}

	return nil
}

func verifyResponseItemVariants(inputItemsMarkdown string) error {
	actual := parseResourceItemVariants(inputItemsMarkdown)
	if err := requireExactInventory("response item resource variants", responseItemVariantNames(), mapKeys(actual)); err != nil {
		return err
	}

	for _, variant := range contractsopenai.ResponsesItemVariants() {
		if actual[variant.Name] != variant.Type {
			return fmt.Errorf("official ResponseItem discriminator drifted for %s: expected=%s actual=%s", variant.Name, variant.Type, actual[variant.Name])
		}
	}

	return nil
}

func parseResourceItemVariants(markdown string) map[string]string {
	lines := strings.Split(strings.ReplaceAll(markdown, "\r\n", "\n"), "\n")

	returnsIndex := indexOfExactLine(lines, "### Returns")
	if returnsIndex < 0 {
		return nil
	}

	out := make(map[string]string)
	currentVariant := ""

	for _, line := range lines[returnsIndex+1:] {
		if match := responseOutputMemberPattern.FindStringSubmatch(line); len(match) == 2 {
			currentVariant = strings.Fields(match[1])[0]
			continue
		}

		if currentVariant == "" {
			continue
		}

		if match := responseOutputTypePattern.FindStringSubmatch(line); len(match) == 2 {
			out[currentVariant] = match[1]
		}
	}

	return out
}

func parseSDKReturnFields(markdown, heading, className string) []string {
	lines := strings.Split(strings.ReplaceAll(markdown, "\r\n", "\n"), "\n")
	headingIndex := indexOfExactLine(lines, heading)
	returnsIndex := indexOfExactLineAfter(lines, "### Returns", headingIndex+1)

	classIndex := indexOfExactLineAfter(lines, "- `class "+className+": …`", returnsIndex+1)
	if headingIndex < 0 || returnsIndex < 0 || classIndex < 0 {
		return nil
	}

	fields := make([]string, 0)

	for index := classIndex + 1; index < len(lines); index++ {
		line := lines[index]
		if strings.HasPrefix(line, "## ") {
			break
		}

		if match := responseFieldPattern.FindStringSubmatch(line); len(match) == 2 {
			fields = append(fields, match[1])
		}
	}

	return fields
}

func parseSDKResponseItemVariants(markdown string) []string {
	lines := strings.Split(strings.ReplaceAll(markdown, "\r\n", "\n"), "\n")
	headingIndex := indexOfExactLine(lines, "## List input items")

	returnsIndex := indexOfExactLineAfter(lines, "### Returns", headingIndex+1)
	if headingIndex < 0 || returnsIndex < 0 {
		return nil
	}

	classPattern := regexp.MustCompile(`^ {2}- ` + "`" + `class ([^:]+): …` + "`" + `$`)
	variants := make([]string, 0)

	for index := returnsIndex + 1; index < len(lines); index++ {
		line := lines[index]
		if strings.HasPrefix(line, "## ") {
			break
		}

		if match := classPattern.FindStringSubmatch(line); len(match) == 2 {
			variants = append(variants, match[1])
		}
	}

	return variants
}

func parseResourceReturnFields(markdown string) []string {
	lines := strings.Split(strings.ReplaceAll(markdown, "\r\n", "\n"), "\n")

	returnsIndex := indexOfExactLine(lines, "### Returns")
	if returnsIndex < 0 {
		return nil
	}

	fields := make([]string, 0)

	for _, line := range lines[returnsIndex+1:] {
		if match := responseFieldPattern.FindStringSubmatch(line); len(match) == 2 {
			fields = append(fields, match[1])
		}
	}

	return fields
}

func verifyOutputVariantDiscriminators(createMarkdown, sdkMarkdown string) error {
	createTypes := parseCreateOutputDiscriminators(createMarkdown)

	for _, variant := range contractsopenai.ResponsesOutputVariants() {
		verified := false
		if createType := createTypes[variant.Name]; createType != "" {
			verified = true

			if createType != variant.Type {
				return fmt.Errorf("official Responses output discriminator drifted for %s: expected=%s actual=%s", variant.Name, variant.Type, createType)
			}
		}

		if sdkType := parseSDKOutputDiscriminator(sdkMarkdown, variant.SDKName); sdkType != "" {
			verified = true

			if sdkType != variant.Type {
				return fmt.Errorf("official Responses SDK output discriminator drifted for %s: expected=%s actual=%s", variant.SDKName, variant.Type, sdkType)
			}
		}

		if !verified {
			return fmt.Errorf("official Responses output discriminator is undocumented for %s", variant.Name)
		}
	}

	return nil
}

func parseCreateOutputDiscriminators(markdown string) map[string]string {
	lines := strings.Split(strings.ReplaceAll(markdown, "\r\n", "\n"), "\n")
	outputIndex := indexOfPrefixLine(lines, "  - `output:")

	parallelToolsIndex := indexOfPrefixLineAfter(lines, "  - `parallel_tool_calls:", outputIndex+1)
	if outputIndex < 0 || parallelToolsIndex <= outputIndex {
		return nil
	}

	out := make(map[string]string)
	currentVariant := ""

	for _, line := range lines[outputIndex+1 : parallelToolsIndex] {
		if match := responseOutputMemberPattern.FindStringSubmatch(line); len(match) == 2 {
			currentVariant = match[1]
			continue
		}

		if currentVariant == "" {
			continue
		}

		if match := responseOutputTypePattern.FindStringSubmatch(line); len(match) == 2 {
			out[currentVariant] = match[1]
		}
	}

	return out
}

func parseSDKOutputDiscriminator(markdown, sdkName string) string {
	lines := strings.Split(strings.ReplaceAll(markdown, "\r\n", "\n"), "\n")

	target := "- `class " + sdkName + ": …`"
	for index, line := range lines {
		if strings.TrimSpace(line) != target {
			continue
		}

		indent := len(line) - len(strings.TrimLeft(line, " "))
		for childIndex := index + 1; childIndex < len(lines); childIndex++ {
			childLine := lines[childIndex]

			trimmed := strings.TrimSpace(childLine)
			if strings.HasPrefix(trimmed, "- ") {
				childIndent := len(childLine) - len(strings.TrimLeft(childLine, " "))
				if childIndent <= indent {
					break
				}
			}

			match := sdkOutputTypePattern.FindStringSubmatch(childLine)
			if len(match) == 3 && len(match[1]) == indent+2 {
				return match[2]
			}
		}
	}

	return ""
}

const maxOfficialReferenceBytes = 32 << 20

const (
	officialFetchAttempts = 3
	officialFetchBackoff  = 250 * time.Millisecond
)

func fetchOfficialMarkdown(ctx context.Context, client *http.Client, url string) (bodyText string, err error) {
	attemptErrors := make([]error, 0, officialFetchAttempts)

	for attempt := 0; attempt < officialFetchAttempts; attempt++ {
		body, retryable, fetchErr := fetchOfficialMarkdownOnce(ctx, client, url)
		if fetchErr == nil {
			return body, nil
		}

		attemptErrors = append(attemptErrors, fmt.Errorf("attempt %d: %w", attempt+1, fetchErr))
		if !retryable || attempt == officialFetchAttempts-1 {
			return "", errors.Join(attemptErrors...)
		}

		timer := time.NewTimer(officialFetchBackoff * time.Duration(1<<attempt))
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}

			return "", errors.Join(append(attemptErrors, ctx.Err())...)
		case <-timer.C:
		}
	}

	return "", errors.Join(attemptErrors...)
}

func fetchOfficialMarkdownOnce(ctx context.Context, client *http.Client, url string) (bodyText string, retryable bool, err error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return "", false, err
	}

	response, err := client.Do(request)
	if err != nil {
		return "", true, err
	}
	defer func() {
		if closeErr := response.Body.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close official reference response: %w", closeErr))
		}
	}()

	if response.StatusCode != http.StatusOK {
		retryableStatus := response.StatusCode == http.StatusTooManyRequests || response.StatusCode >= http.StatusInternalServerError

		return "", retryableStatus, fmt.Errorf("official reference returned HTTP %d", response.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(response.Body, maxOfficialReferenceBytes+1))
	if err != nil {
		return "", true, err
	}

	if len(body) > maxOfficialReferenceBytes {
		return "", false, errors.New("official reference exceeds the maximum supported size")
	}

	return string(body), true, nil
}

type parsedResponsesContract struct {
	fields          []string
	fieldContracts  map[string]string
	outputVariants  []string
	streamingEvents []string
	statuses        []string
}

func parseOfficialResponsesContract(createMarkdown, streamingMarkdown string) (parsedResponsesContract, error) {
	parsed := parsedResponsesContract{fieldContracts: make(map[string]string)}

	createLines := strings.Split(strings.ReplaceAll(createMarkdown, "\r\n", "\n"), "\n")

	returnsIndex := indexOfExactLine(createLines, "### Returns")
	if returnsIndex < 0 {
		return parsed, errors.New("official Responses create reference has no Returns section")
	}

	outputIndex := -1
	parallelToolsIndex := -1
	statusIndex := -1

	for index := returnsIndex + 1; index < len(createLines); index++ {
		line := createLines[index]
		if match := responseFieldPattern.FindStringSubmatch(line); len(match) == 2 {
			parsed.fields = append(parsed.fields, match[1])
		}

		if match := responseFieldContractPattern.FindStringSubmatch(line); len(match) == 3 {
			parsed.fieldContracts[match[1]] = match[2]
		}

		switch {
		case strings.HasPrefix(line, "  - `output:"):
			outputIndex = index
		case strings.HasPrefix(line, "  - `parallel_tool_calls:"):
			parallelToolsIndex = index
		case strings.HasPrefix(line, "  - `status:"):
			statusIndex = index
		}
	}

	if outputIndex < 0 || parallelToolsIndex <= outputIndex {
		return parsed, errors.New("official Responses output union was not found")
	}

	for _, line := range createLines[outputIndex+1 : parallelToolsIndex] {
		if match := responseOutputMemberPattern.FindStringSubmatch(line); len(match) == 2 {
			parsed.outputVariants = append(parsed.outputVariants, match[1])
		}
	}

	if statusIndex < 0 {
		return parsed, errors.New("official Response status field was not found")
	}

	for _, line := range createLines[statusIndex+1:] {
		if strings.HasPrefix(line, "  - `") {
			break
		}

		if match := responseStatusValuePattern.FindStringSubmatch(line); len(match) == 2 {
			parsed.statuses = append(parsed.statuses, match[1])
		}
	}

	streamingLines := strings.Split(strings.ReplaceAll(streamingMarkdown, "\r\n", "\n"), "\n")
	for _, line := range streamingLines {
		if match := responseEventHeadingPattern.FindStringSubmatch(line); len(match) == 2 {
			parsed.streamingEvents = append(parsed.streamingEvents, match[1])
		}
	}

	return parsed, nil
}

func indexOfExactLine(lines []string, target string) int {
	return indexOfExactLineAfter(lines, target, 0)
}

func indexOfExactLineAfter(lines []string, target string, start int) int {
	if start < 0 {
		return -1
	}

	for index := start; index < len(lines); index++ {
		if lines[index] == target {
			return index
		}
	}

	return -1
}

func indexOfPrefixLine(lines []string, prefix string) int {
	return indexOfPrefixLineAfter(lines, prefix, 0)
}

func indexOfPrefixLineAfter(lines []string, prefix string, start int) int {
	if start < 0 {
		return -1
	}

	for index := start; index < len(lines); index++ {
		if strings.HasPrefix(lines[index], prefix) {
			return index
		}
	}

	return -1
}

func responseOutputVariantNames() []string {
	variants := contractsopenai.ResponsesOutputVariants()

	names := make([]string, 0, len(variants))
	for _, variant := range variants {
		names = append(names, variant.Name)
	}

	return names
}

func responseItemVariantNames() []string {
	variants := contractsopenai.ResponsesItemVariants()

	names := make([]string, 0, len(variants))
	for _, variant := range variants {
		names = append(names, variant.Name)
	}

	return names
}

func responseItemVariantSDKNames() []string {
	variants := contractsopenai.ResponsesItemVariants()

	names := make([]string, 0, len(variants))
	for _, variant := range variants {
		names = append(names, variant.SDKName)
	}

	return names
}

func mapKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}

	return keys
}

func stringSetValues(values map[string]struct{}) []string {
	items := make([]string, 0, len(values))
	for value := range values {
		items = append(items, value)
	}

	return items
}

func requireExactInventory(label string, expected, actual []string) error {
	expectedSet := exactStringSet(expected)
	actualSet := exactStringSet(actual)
	missing := difference(expectedSet, actualSet)

	extra := difference(actualSet, expectedSet)
	if len(missing) > 0 || len(extra) > 0 || len(expected) != len(actual) {
		return fmt.Errorf("official Responses %s drifted: expected=%d actual=%d missing=%v extra=%v", label, len(expected), len(actual), missing, extra)
	}

	return nil
}

func exactStringSet(values []string) map[string]struct{} {
	out := make(map[string]struct{}, len(values))
	for _, value := range values {
		out[value] = struct{}{}
	}

	return out
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
