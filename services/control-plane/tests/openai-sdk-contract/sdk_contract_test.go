package openaisdkcontract

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
)

func TestOfficialOpenAIGoSDKDecodesEveryResponseOutputVariant(t *testing.T) {
	t.Parallel()

	responseBody := officialSDKResponseFixture(t)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer closeRequestBody(t, request)

		if request.Method != http.MethodPost || request.URL.Path != "/v1/responses" {
			t.Errorf("unexpected SDK request %s %s", request.Method, request.URL.Path)
			writer.WriteHeader(http.StatusNotFound)

			return
		}

		if request.Header.Get("Authorization") != "Bearer contract-key" {
			t.Errorf("unexpected Authorization header %q", request.Header.Get("Authorization"))
		}

		writer.Header().Set("Content-Type", "application/json")

		_, err := writer.Write(responseBody)
		if err != nil {
			t.Errorf("write SDK response fixture: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	client := openai.NewClient(
		option.WithAPIKey("contract-key"),
		option.WithBaseURL(server.URL+"/v1"),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := client.Responses.New(ctx, responses.ResponseNewParams{
		Model: openai.ChatModelGPT4o,
		Input: responses.ResponseNewParamsInputUnion{OfString: openai.String("contract")},
	})
	if err != nil {
		t.Fatalf("official OpenAI SDK decode response: %v", err)
	}

	if response.ID != "resp_sdk_contract" {
		t.Fatalf("expected response ID resp_sdk_contract, got %q", response.ID)
	}

	if len(response.Output) != len(contractsopenai.ResponsesOutputVariants()) {
		t.Fatalf("expected %d output variants, got %d", len(contractsopenai.ResponsesOutputVariants()), len(response.Output))
	}

	for index, variant := range contractsopenai.ResponsesOutputVariants() {
		if response.Output[index].Type != variant.Type {
			t.Errorf("output variant %s decoded as %q, expected %q", variant.Name, response.Output[index].Type, variant.Type)
		}
	}

	if response.OutputText() != "contract output" {
		t.Fatalf("expected SDK OutputText helper to return contract output, got %q", response.OutputText())
	}
}

func TestOfficialOpenAIGoSDKDecodesEveryResponsesStreamingEvent(t *testing.T) {
	t.Parallel()

	orderedEvents := orderedStreamingEvents()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer closeRequestBody(t, request)

		writer.Header().Set("Content-Type", "text/event-stream")

		for sequenceNumber, eventType := range orderedEvents {
			payload := map[string]interface{}{
				"type":            eventType,
				"sequence_number": sequenceNumber,
			}
			if eventType == "response.completed" {
				var response map[string]interface{}
				if err := json.Unmarshal(officialSDKResponseFixture(t), &response); err != nil {
					t.Errorf("decode completed response fixture: %v", err)

					return
				}

				payload["response"] = response
			}

			if eventType == "error" {
				payload["code"] = nil
				payload["message"] = "contract error event"
				payload["param"] = nil
			}

			encoded, err := json.Marshal(payload)
			if err != nil {
				t.Errorf("encode %s fixture: %v", eventType, err)

				return
			}

			if _, err := fmt.Fprintf(writer, "event: %s\ndata: %s\n\n", eventType, encoded); err != nil {
				t.Errorf("write %s fixture: %v", eventType, err)

				return
			}
		}
	}))
	t.Cleanup(server.Close)

	client := openai.NewClient(
		option.WithAPIKey("contract-key"),
		option.WithBaseURL(server.URL+"/v1"),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream := client.Responses.NewStreaming(ctx, responses.ResponseNewParams{
		Model: openai.ChatModelGPT4o,
		Input: responses.ResponseNewParamsInputUnion{OfString: openai.String("contract")},
	})
	defer stream.Close()

	decodedTypes := make([]string, 0, len(orderedEvents))
	for stream.Next() {
		decodedTypes = append(decodedTypes, stream.Current().Type)
	}

	if err := stream.Err(); err != nil {
		t.Fatalf("official OpenAI SDK decode stream: %v", err)
	}

	if len(decodedTypes) != len(orderedEvents) {
		t.Fatalf("expected %d decoded events, got %d: %v", len(orderedEvents), len(decodedTypes), decodedTypes)
	}

	for index, expectedType := range orderedEvents {
		if decodedTypes[index] != expectedType {
			t.Errorf("stream event %d decoded as %q, expected %q", index, decodedTypes[index], expectedType)
		}
	}
}

func TestOfficialOpenAIGoSDKDecodesEveryResponsesOperation(t *testing.T) {
	t.Parallel()

	responseBody := officialSDKResponseFixture(t)

	var responseObject map[string]interface{}
	if err := json.Unmarshal(responseBody, &responseObject); err != nil {
		t.Fatalf("decode response fixture: %v", err)
	}

	compactedBody, err := json.Marshal(map[string]interface{}{
		"id":         "resp_compaction_sdk",
		"created_at": int64(1_786_147_200),
		"object":     "response.compaction",
		"output":     responseObject["output"],
		"usage":      responseObject["usage"],
	})
	if err != nil {
		t.Fatalf("encode compacted response fixture: %v", err)
	}

	inputTokenBody := []byte(`{"object":"response.input_tokens","input_tokens":42}`)
	inputItemsBody := officialSDKInputItemsFixture(t)

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer closeRequestBody(t, request)

		writer.Header().Set("Content-Type", "application/json")

		var body []byte

		switch {
		case request.Method == http.MethodPost && request.URL.Path == "/v1/responses/compact":
			body = compactedBody
		case request.Method == http.MethodPost && request.URL.Path == "/v1/responses/input_tokens":
			body = inputTokenBody
		case request.Method == http.MethodGet && request.URL.Path == "/v1/responses/resp_sdk/input_items":
			body = inputItemsBody
		case request.Method == http.MethodGet && request.URL.Path == "/v1/responses/resp_sdk":
			body = responseBody
		case request.Method == http.MethodPost && request.URL.Path == "/v1/responses/resp_sdk/cancel":
			body = responseBody
		case request.Method == http.MethodDelete && request.URL.Path == "/v1/responses/resp_sdk":
			writer.WriteHeader(http.StatusOK)

			return
		default:
			t.Errorf("unexpected SDK operation %s %s", request.Method, request.URL.Path)
			writer.WriteHeader(http.StatusNotFound)

			return
		}

		if _, writeErr := writer.Write(body); writeErr != nil {
			t.Errorf("write operation fixture: %v", writeErr)
		}
	}))
	t.Cleanup(server.Close)

	client := openai.NewClient(
		option.WithAPIKey("contract-key"),
		option.WithBaseURL(server.URL+"/v1"),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	gotResponse, err := client.Responses.Get(ctx, "resp_sdk", responses.ResponseGetParams{})
	if err != nil || gotResponse.ID != "resp_sdk_contract" {
		t.Fatalf("official SDK retrieve: response=%v err=%v", gotResponse, err)
	}

	cancelledResponse, err := client.Responses.Cancel(ctx, "resp_sdk")
	if err != nil || cancelledResponse.ID != "resp_sdk_contract" {
		t.Fatalf("official SDK cancel: response=%v err=%v", cancelledResponse, err)
	}

	if deleteErr := client.Responses.Delete(ctx, "resp_sdk"); deleteErr != nil {
		t.Fatalf("official SDK delete: %v", deleteErr)
	}

	compactedResponse, err := client.Responses.Compact(ctx, responses.ResponseCompactParams{
		Model: responses.ResponseCompactParamsModel("gpt-4o"),
	})
	if err != nil {
		t.Fatalf("official SDK compact: %v", err)
	}

	if len(compactedResponse.Output) != len(contractsopenai.ResponsesOutputVariants()) {
		t.Fatalf("expected %d compacted output variants, got %d", len(contractsopenai.ResponsesOutputVariants()), len(compactedResponse.Output))
	}

	inputTokenResponse, err := client.Responses.InputTokens.Count(ctx, responses.InputTokenCountParams{
		Model: openai.String("gpt-4o"),
	})
	if err != nil || inputTokenResponse.InputTokens != 42 {
		t.Fatalf("official SDK input tokens: response=%v err=%v", inputTokenResponse, err)
	}

	inputItems, err := client.Responses.InputItems.List(ctx, "resp_sdk", responses.InputItemListParams{})
	if err != nil {
		t.Fatalf("official SDK input items: %v", err)
	}

	if len(inputItems.Data) != len(contractsopenai.ResponsesItemVariants()) {
		t.Fatalf("expected %d response item variants, got %d", len(contractsopenai.ResponsesItemVariants()), len(inputItems.Data))
	}

	for index, variant := range contractsopenai.ResponsesItemVariants() {
		if inputItems.Data[index].Type != variant.Type {
			t.Errorf("response item %s decoded as %q, expected %q", variant.Name, inputItems.Data[index].Type, variant.Type)
		}
	}
}

func officialSDKResponseFixture(t *testing.T) []byte {
	t.Helper()

	output := make([]map[string]interface{}, 0, len(contractsopenai.ResponsesOutputVariants()))
	for index, variant := range contractsopenai.ResponsesOutputVariants() {
		item := map[string]interface{}{
			"id":     fmt.Sprintf("item_%d", index),
			"type":   variant.Type,
			"status": "completed",
		}
		if variant.Type == "message" {
			item["role"] = "assistant"
			item["content"] = []interface{}{
				map[string]interface{}{
					"type":        "output_text",
					"text":        "contract output",
					"annotations": []interface{}{},
					"logprobs":    []interface{}{},
				},
			}
		}

		output = append(output, item)
	}

	response := map[string]interface{}{
		"id":                  "resp_sdk_contract",
		"created_at":          float64(1_786_147_200),
		"error":               nil,
		"incomplete_details":  nil,
		"instructions":        nil,
		"metadata":            map[string]interface{}{},
		"model":               "gpt-4o",
		"object":              "response",
		"output":              output,
		"parallel_tool_calls": true,
		"temperature":         0.7,
		"tool_choice":         "auto",
		"tools":               []interface{}{},
		"top_p":               0.9,
		"status":              "completed",
		"text":                map[string]interface{}{"format": map[string]interface{}{"type": "text"}},
		"truncation":          "disabled",
		"usage": map[string]interface{}{
			"input_tokens":          float64(1),
			"input_tokens_details":  map[string]interface{}{"cached_tokens": float64(0)},
			"output_tokens":         float64(1),
			"output_tokens_details": map[string]interface{}{"reasoning_tokens": float64(0)},
			"total_tokens":          float64(2),
		},
	}

	encoded, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("encode SDK response fixture: %v", err)
	}

	return encoded
}

func orderedStreamingEvents() []string {
	all := contractsopenai.ResponsesStreamingEventTypes()

	ordered := make([]string, 0, len(all))
	for _, eventType := range all {
		if eventType != "response.completed" {
			ordered = append(ordered, eventType)
		}
	}

	return append(ordered, "response.completed")
}

func officialSDKInputItemsFixture(t *testing.T) []byte {
	t.Helper()

	items := make([]map[string]interface{}, 0, len(contractsopenai.ResponsesItemVariants()))
	for index, variant := range contractsopenai.ResponsesItemVariants() {
		item := map[string]interface{}{
			"id":     fmt.Sprintf("item_%d", index),
			"type":   variant.Type,
			"status": "completed",
		}
		switch variant.Name {
		case "ResponseInputMessageItem":
			item["role"] = "user"
			item["content"] = []interface{}{
				map[string]interface{}{"type": "input_text", "text": "contract"},
			}
		case "ResponseOutputMessage":
			item["role"] = "assistant"
			item["content"] = []interface{}{
				map[string]interface{}{
					"type":        "output_text",
					"text":        "contract output",
					"annotations": []interface{}{},
					"logprobs":    []interface{}{},
				},
			}
		}

		items = append(items, item)
	}

	encoded, err := json.Marshal(map[string]interface{}{
		"object":   "list",
		"data":     items,
		"first_id": "item_0",
		"last_id":  fmt.Sprintf("item_%d", len(items)-1),
		"has_more": false,
	})
	if err != nil {
		t.Fatalf("encode SDK input-items fixture: %v", err)
	}

	return encoded
}

func closeRequestBody(t *testing.T, request *http.Request) {
	t.Helper()

	if request.Body == nil {
		return
	}

	if err := request.Body.Close(); err != nil {
		t.Errorf("close request body: %v", err)
	}
}

func TestOfficialSDKFixtureIsStableJSON(t *testing.T) {
	t.Parallel()

	fixture := officialSDKResponseFixture(t)
	if !json.Valid(fixture) || !bytes.Contains(fixture, []byte(`"object":"response"`)) {
		t.Fatalf("invalid SDK fixture: %s", fixture)
	}
}
