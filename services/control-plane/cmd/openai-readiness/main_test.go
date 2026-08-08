package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
)

func TestFetchOfficialMarkdownRetriesTransientServerFailures(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		if attempts.Add(1) < officialFetchAttempts {
			writer.WriteHeader(http.StatusServiceUnavailable)

			return
		}

		_, err := writer.Write([]byte("official contract"))
		if err != nil {
			t.Errorf("write test response: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	body, err := fetchOfficialMarkdown(ctx, server.Client(), server.URL)
	if err != nil {
		t.Fatalf("fetch after transient failures: %v", err)
	}

	if body != "official contract" || attempts.Load() != officialFetchAttempts {
		t.Fatalf("unexpected retry result: body=%q attempts=%d", body, attempts.Load())
	}
}

func TestFetchOfficialMarkdownDoesNotRetryPermanentClientFailure(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		attempts.Add(1)
		writer.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(server.Close)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := fetchOfficialMarkdown(ctx, server.Client(), server.URL)
	if err == nil || attempts.Load() != 1 {
		t.Fatalf("expected one permanent-failure attempt, err=%v attempts=%d", err, attempts.Load())
	}
}

func TestAuditNormalizesRouteParametersAndReportsMissing(t *testing.T) {
	official := []string{
		"POST /responses",
		"GET /responses/{response_id}",
		"POST /responses/{response_id}/cancel?beta=true",
	}
	routes := `
v1.Post("/responses", handler)
v1.Get("/responses/:id", handler)
`

	report := audit(official, routes)
	if report.RouteReady || report.ContractReady {
		t.Fatal("expected incomplete route matrix")
	}

	if report.ImplementedCount != 2 || len(report.Missing) != 1 ||
		report.Missing[0] != "POST /responses/{}/cancel" {
		t.Fatalf("unexpected report: %+v", report)
	}

	if report.LiveValidated {
		t.Fatal("contract audit must never claim live validation")
	}
}

func TestAuditScopesTheOfficialBaselineToResponsesOperations(t *testing.T) {
	official := []string{
		"POST /embeddings",
		"GET /models",
		"POST /responses",
		"GET /responses/{response_id}",
	}
	routes := `v1.Post("/responses", handler)
v1.Get("/responses/:id", handler)`

	report := audit(official, routes)
	if !report.RouteReady || report.ContractReady || len(report.Missing) != 0 {
		t.Fatalf("expected route-ready but not live-verified report: %+v", report)
	}

	if report.OfficialCount != 2 || report.ImplementedCount != 2 {
		t.Fatalf("expected responses-only counts: %+v", report)
	}

	if report.ProductionReady || report.WebSocketReady {
		t.Fatalf("static REST audit must not claim websocket or production readiness: %+v", report)
	}

	if report.SnapshotCoverage != 0 {
		t.Fatalf("unverified audit must not claim snapshot coverage: %+v", report)
	}
}

func TestPinnedResponsesOperationsMustMatchLiveContractInventory(t *testing.T) {
	t.Parallel()

	if err := verifyPinnedResponsesOperations(contractsopenai.ResponsesOperations()); err != nil {
		t.Fatalf("current operation inventory must match itself: %v", err)
	}

	mutant := contractsopenai.ResponsesOperations()[1:]
	if err := verifyPinnedResponsesOperations(mutant); err == nil {
		t.Fatal("expected a stale pinned operation baseline to fail")
	}
}

func TestParseSDKOperationsExtractsExactMethodsAndPaths(t *testing.T) {
	t.Parallel()

	markdown := "## Cancel a response\n\n" +
		"**post** `/responses/{response_id}/cancel`\n\n" +
		"## Domain Types\n\n" +
		"**get** `/responses/{response_id}`\n\n" +
		"**get** `/unrelated`\n"

	assertStringsEqual(t, []string{"POST /responses/{response_id}/cancel", "GET /responses/{response_id}"}, parseSDKOperations(markdown))
}

func TestParseOfficialResponsesContractExtractsEveryInventoryDimension(t *testing.T) {
	createMarkdown := "# Create a model response\n\n" +
		"### Returns\n\n" +
		"- Response\n\n" +
		"  - `id: string`\n\n" +
		"  - `output: array of ResponseOutputItem`\n\n" +
		"    - `Message object { id, type }`\n\n" +
		"      - `type: \"message\"`\n\n" +
		"    - `FunctionCall object { id, type }`\n\n" +
		"      - `type: \"function_call\"`\n\n" +
		"  - `parallel_tool_calls: bool`\n\n" +
		"  - `status: optional ResponseStatus`\n\n" +
		"    - `\"completed\"`\n\n" +
		"    - `\"failed\"`\n\n" +
		"  - `text: object`\n"
	streamingMarkdown := `# Events

## response.created

## response.completed

## error
`

	parsed, err := parseOfficialResponsesContract(createMarkdown, streamingMarkdown)
	if err != nil {
		t.Fatalf("parse contract: %v", err)
	}

	assertStringsEqual(t, []string{"id", "output", "parallel_tool_calls", "status", "text"}, parsed.fields)

	if parsed.fieldContracts["id"] != "string" || parsed.fieldContracts["status"] != "optional ResponseStatus" {
		t.Fatalf("unexpected field contracts: %v", parsed.fieldContracts)
	}

	assertStringsEqual(t, []string{"Message", "FunctionCall"}, parsed.outputVariants)
	assertStringsEqual(t, []string{"completed", "failed"}, parsed.statuses)
	assertStringsEqual(t, []string{"response.created", "response.completed", "error"}, parsed.streamingEvents)
}

func TestOutputDiscriminatorParsersReadResourceAndSDKReferences(t *testing.T) {
	createMarkdown := "### Returns\n\n" +
		"  - `output: array of ResponseOutputItem`\n\n" +
		"    - `Message object { type }`\n\n" +
		"      - `type: \"message\"`\n\n" +
		"  - `parallel_tool_calls: bool`\n"
	sdkMarkdown := "- `class ResponseFunctionToolCall: …`\n\n" +
		"  - `type: Literal[\"function_call\"]`\n"

	createTypes := parseCreateOutputDiscriminators(createMarkdown)
	if createTypes["Message"] != "message" {
		t.Fatalf("expected message discriminator, got %q", createTypes["Message"])
	}

	if got := parseSDKOutputDiscriminator(sdkMarkdown, "ResponseFunctionToolCall"); got != "function_call" {
		t.Fatalf("expected function_call discriminator, got %q", got)
	}
}

func TestRequireExactInventoryRejectsDrift(t *testing.T) {
	err := requireExactInventory("events", []string{"response.created", "response.completed"}, []string{"response.created", "response.failed"})
	if err == nil {
		t.Fatal("expected inventory drift to fail verification")
	}
}

func TestContractGateKillsEverySingleMemberDeletionMutation(t *testing.T) {
	t.Parallel()

	inventories := []struct {
		name   string
		values []string
	}{
		{name: "operations", values: contractsopenai.ResponsesOperations()},
		{name: "response fields", values: contractsopenai.ResponsesTopLevelFields()},
		{name: "output variants", values: responseOutputVariantNames()},
		{name: "response item variants", values: responseItemVariantSDKNames()},
		{name: "compacted fields", values: contractsopenai.ResponsesCompactedFields()},
		{name: "input token fields", values: contractsopenai.ResponsesInputTokenCountFields()},
		{name: "input items page fields", values: contractsopenai.ResponsesInputItemsPageFields()},
		{name: "streaming events", values: contractsopenai.ResponsesStreamingEventTypes()},
		{name: "response statuses", values: contractsopenai.ResponsesStatuses()},
	}

	for _, inventory := range inventories {
		for removedIndex, removedValue := range inventory.values {
			name := inventory.name + "/remove " + removedValue
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				mutant := append([]string(nil), inventory.values[:removedIndex]...)

				mutant = append(mutant, inventory.values[removedIndex+1:]...)
				if err := requireExactInventory(inventory.name, inventory.values, mutant); err == nil {
					t.Fatalf("deletion mutation for %q survived the contract gate", removedValue)
				}
			})
		}
	}
}

func assertStringsEqual(t *testing.T, expected, actual []string) {
	t.Helper()

	if len(expected) != len(actual) {
		t.Fatalf("expected %v, got %v", expected, actual)
	}

	for index := range expected {
		if expected[index] != actual[index] {
			t.Fatalf("expected %v, got %v", expected, actual)
		}
	}
}
