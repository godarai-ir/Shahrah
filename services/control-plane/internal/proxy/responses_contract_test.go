package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenAIResponsesContractRelaysEveryTopLevelFieldAndOutputVariantLosslessly(t *testing.T) {
	t.Parallel()

	output := responsesOutputContractFixtures()

	response := completeResponsesContractFixture(output)
	upstreamBody, err := json.Marshal(response)
	require.NoError(t, err)

	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       upstreamBody,
	}}
	app, _ := newResponsesTestApp(t, "openai", tunnel)

	req := httptest.NewRequest("POST", "/v1/responses", bytes.NewBufferString(`{"model":"gpt-4o","input":"contract"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	received, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, upstreamBody, received)

	var decoded map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(received, &decoded))

	for _, field := range contractsopenai.ResponsesTopLevelFields() {
		assert.Containsf(t, decoded, field, "official top-level Response field %q was not relayed", field)
	}

	var decodedOutput []map[string]interface{}
	require.NoError(t, json.Unmarshal(decoded["output"], &decodedOutput))
	require.Len(t, decodedOutput, len(contractsopenai.ResponsesOutputVariants()))

	for index, variant := range contractsopenai.ResponsesOutputVariants() {
		assert.Equal(t, variant.Type, decodedOutput[index]["type"])
		assert.Equal(t, variant.Name, decodedOutput[index]["contract_variant"])
		assert.Contains(t, decodedOutput[index], "future_nested_field")
	}
}

func responsesOutputContractFixtures() []map[string]interface{} {
	output := make([]map[string]interface{}, 0, len(contractsopenai.ResponsesOutputVariants()))
	for index, variant := range contractsopenai.ResponsesOutputVariants() {
		item := map[string]interface{}{
			"id":               fmt.Sprintf("item_%d", index),
			"type":             variant.Type,
			"status":           "completed",
			"contract_variant": variant.Name,
			"future_nested_field": map[string]interface{}{
				"array":  []interface{}{nil, true, float64(index), "value"},
				"object": map[string]interface{}{"preserved": true},
			},
		}
		if variant.Type == "message" {
			item["role"] = "assistant"
			item["content"] = []interface{}{
				map[string]interface{}{
					"type":        "output_text",
					"text":        "contract",
					"annotations": []interface{}{},
				},
			}
		}

		output = append(output, item)
	}

	return output
}

func TestOpenAIResponsesContractRelaysEveryStatusLosslessly(t *testing.T) {
	t.Parallel()

	for _, status := range contractsopenai.ResponsesStatuses() {
		t.Run(status, func(t *testing.T) {
			t.Parallel()

			response := completeResponsesContractFixture([]map[string]interface{}{})
			response["id"] = "resp_" + status

			response["status"] = status
			if status == "failed" {
				response["error"] = map[string]interface{}{
					"code":    "server_error",
					"message": "contract failure",
				}
			}

			if status == "incomplete" {
				response["incomplete_details"] = map[string]interface{}{"reason": "max_output_tokens"}
			}

			upstreamBody, err := json.Marshal(response)
			require.NoError(t, err)

			tunnel := &captureTunnel{resp: &platform.TunnelResponse{
				StatusCode: fiber.StatusOK,
				Headers:    map[string]string{"content-type": "application/json"},
				Body:       upstreamBody,
			}}
			app, _ := newResponsesTestApp(t, "openai", tunnel)

			req := httptest.NewRequest("POST", "/v1/responses", bytes.NewBufferString(`{"model":"gpt-4o","input":"contract"}`))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			defer closeResponseBody(t, resp)

			require.NoError(t, err)
			require.Equal(t, fiber.StatusOK, resp.StatusCode)

			received, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			if upstreamBody == nil {
				assert.Empty(t, received)
			} else {
				assert.Equal(t, upstreamBody, received)
			}
		})
	}
}

func TestOpenAIResponsesContractAcceptsEveryOptionalAndNullableFieldMutation(t *testing.T) {
	t.Parallel()

	for _, field := range contractsopenai.ResponsesFieldContracts() {
		if field.Optional {
			t.Run("omit/"+field.Name, func(t *testing.T) {
				t.Parallel()

				response := completeResponsesContractFixture(responsesOutputContractFixtures())
				delete(response, field.Name)
				assertOpenAIResponseRelayedLosslessly(t, response)
			})
		}

		if field.Nullable {
			t.Run("null/"+field.Name, func(t *testing.T) {
				t.Parallel()

				response := completeResponsesContractFixture(responsesOutputContractFixtures())
				response[field.Name] = nil
				assertOpenAIResponseRelayedLosslessly(t, response)
			})
		}
	}
}

func TestOpenAIResponsesOperationContractsRelayEveryResponseShapeLosslessly(t *testing.T) {
	t.Parallel()

	compactedResponse := map[string]interface{}{
		"id":         "resp_compaction_contract",
		"created_at": float64(1_786_147_200),
		"object":     "response.compaction",
		"output":     responsesOutputContractFixtures(),
		"usage": map[string]interface{}{
			"input_tokens":  float64(1),
			"output_tokens": float64(1),
			"total_tokens":  float64(2),
		},
	}
	inputTokenResponse := map[string]interface{}{
		"input_tokens": float64(42),
		"object":       "response.input_tokens",
	}
	inputItemsResponse := map[string]interface{}{
		"object":   "list",
		"data":     responsesItemContractFixtures(),
		"first_id": "item_0",
		"last_id":  "item_28",
		"has_more": false,
	}
	response := completeResponsesContractFixture(responsesOutputContractFixtures())

	tests := []struct {
		name           string
		method         string
		path           string
		requestBody    string
		response       interface{}
		expectedFields []string
	}{
		{
			name:           "compact",
			method:         http.MethodPost,
			path:           "/v1/responses/compact",
			requestBody:    `{"model":"gpt-4o","input":"contract"}`,
			response:       compactedResponse,
			expectedFields: contractsopenai.ResponsesCompactedFields(),
		},
		{
			name:           "input tokens",
			method:         http.MethodPost,
			path:           "/v1/responses/input_tokens",
			requestBody:    `{"model":"gpt-4o","input":"contract"}`,
			response:       inputTokenResponse,
			expectedFields: contractsopenai.ResponsesInputTokenCountFields(),
		},
		{
			name:           "retrieve",
			method:         http.MethodGet,
			path:           "/v1/responses/resp_public",
			response:       response,
			expectedFields: contractsopenai.ResponsesTopLevelFields(),
		},
		{
			name:           "cancel",
			method:         http.MethodPost,
			path:           "/v1/responses/resp_public/cancel",
			response:       response,
			expectedFields: contractsopenai.ResponsesTopLevelFields(),
		},
		{
			name:           "input items",
			method:         http.MethodGet,
			path:           "/v1/responses/resp_public/input_items",
			response:       inputItemsResponse,
			expectedFields: contractsopenai.ResponsesInputItemsPageFields(),
		},
		{
			name:        "delete",
			method:      http.MethodDelete,
			path:        "/v1/responses/resp_public",
			response:    nil,
			requestBody: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var upstreamBody []byte

			if test.response != nil {
				var err error

				upstreamBody, err = json.Marshal(test.response)
				require.NoError(t, err)
			}

			tunnel := &captureTunnel{resp: &platform.TunnelResponse{
				StatusCode: fiber.StatusOK,
				Headers:    map[string]string{"content-type": "application/json"},
				Body:       upstreamBody,
			}}
			resources := &responseOperationResources{mappings: map[string]platform.ProviderNativeResourceMapping{
				"resp_public": {
					PublicID:          "resp_public",
					TenantID:          "tenant_a",
					Provider:          "openai",
					ProviderAccountID: "acct_1",
					ResourceKind:      responsesResourceKind,
					UpstreamID:        "resp_upstream",
					Status:            "completed",
				},
			}}
			app := newResponseOperationsTestApp(t, tunnel, resources)

			var requestBody io.Reader = http.NoBody
			if test.requestBody != "" {
				requestBody = bytes.NewBufferString(test.requestBody)
			}

			req := httptest.NewRequest(test.method, test.path, requestBody)
			if test.requestBody != "" {
				req.Header.Set("Content-Type", "application/json")
			}

			resp, err := app.Test(req)
			defer closeResponseBody(t, resp)

			require.NoError(t, err)
			require.Equal(t, fiber.StatusOK, resp.StatusCode)

			received, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			if upstreamBody == nil {
				assert.Empty(t, received)
			} else {
				assert.Equal(t, upstreamBody, received)
			}

			if len(test.expectedFields) > 0 {
				assertJSONContainsEveryContractField(t, received, test.expectedFields)
			}
		})
	}
}

func TestOpenAIResponsesContractRelaysEveryStreamingEventLosslessly(t *testing.T) {
	t.Parallel()

	eventTypes := contractsopenai.ResponsesStreamingEventTypes()

	ordered := make([]string, 0, len(eventTypes))
	for _, eventType := range eventTypes {
		if eventType != "response.completed" {
			ordered = append(ordered, eventType)
		}
	}

	ordered = append(ordered, "response.completed")

	var upstream bytes.Buffer

	for sequenceNumber, eventType := range ordered {
		payload := responsesStreamingContractFixture(eventType, sequenceNumber)
		encoded, err := json.Marshal(payload)
		require.NoError(t, err)

		_, err = fmt.Fprintf(&upstream, "event: %s\ndata: %s\n\n", eventType, encoded)
		require.NoError(t, err)
	}

	upstreamBody := upstream.Bytes()
	firstBoundary := len(upstreamBody) / 3
	secondBoundary := firstBoundary * 2

	stream := newFakeTunnelStream(map[string]string{"content-type": "text/event-stream"}, 3, 0)
	stream.chunks <- platform.TunnelChunk{Data: append([]byte(nil), upstreamBody[:firstBoundary]...)}

	stream.chunks <- platform.TunnelChunk{Data: append([]byte(nil), upstreamBody[firstBoundary:secondBoundary]...)}

	stream.chunks <- platform.TunnelChunk{Data: append([]byte(nil), upstreamBody[secondBoundary:]...), IsDone: true}

	close(stream.chunks)

	tunnel := &captureTunnel{stream: stream}
	app, _ := newResponsesTestApp(t, "openai", tunnel)

	req := httptest.NewRequest("POST", "/v1/responses", bytes.NewBufferString(`{"model":"gpt-4o","input":"contract","stream":true}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, int((5 * time.Second).Milliseconds()))
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	received, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, upstreamBody, received)
	assert.NotContains(t, string(received), "provider stream closed without a terminal Responses event")

	for _, eventType := range eventTypes {
		assert.Containsf(t, string(received), "event: "+eventType+"\n", "official event %q was not relayed", eventType)
	}
}

func completeResponsesContractFixture(output []map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"id":                     "resp_contract",
		"created_at":             float64(1_786_147_200),
		"error":                  nil,
		"incomplete_details":     nil,
		"instructions":           "contract instructions",
		"metadata":               map[string]interface{}{"suite": "responses"},
		"model":                  "gpt-4o",
		"object":                 "response",
		"output":                 output,
		"parallel_tool_calls":    true,
		"temperature":            0.7,
		"tool_choice":            "auto",
		"tools":                  []interface{}{},
		"top_p":                  0.9,
		"background":             false,
		"completed_at":           float64(1_786_147_201),
		"conversation":           map[string]interface{}{"id": "conv_contract"},
		"max_output_tokens":      float64(1024),
		"max_tool_calls":         float64(4),
		"moderation":             map[string]interface{}{"status": "passed"},
		"output_text":            "contract",
		"previous_response_id":   "resp_previous",
		"prompt":                 map[string]interface{}{"id": "pmpt_contract", "version": "1"},
		"prompt_cache_key":       "contract-cache",
		"prompt_cache_options":   map[string]interface{}{"type": "24h"},
		"prompt_cache_retention": "24h",
		"reasoning":              map[string]interface{}{"effort": "medium", "summary": "auto"},
		"safety_identifier":      "contract-user",
		"service_tier":           "default",
		"status":                 "completed",
		"text":                   map[string]interface{}{"format": map[string]interface{}{"type": "text"}},
		"top_logprobs":           float64(3),
		"truncation":             "disabled",
		"usage": map[string]interface{}{
			"input_tokens":  float64(1),
			"output_tokens": float64(1),
			"total_tokens":  float64(2),
		},
		"user": "legacy-contract-user",
	}
}

func assertOpenAIResponseRelayedLosslessly(t *testing.T, response map[string]interface{}) {
	t.Helper()

	upstreamBody, err := json.Marshal(response)
	require.NoError(t, err)

	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       upstreamBody,
	}}
	app, _ := newResponsesTestApp(t, "openai", tunnel)
	req := httptest.NewRequest("POST", "/v1/responses", bytes.NewBufferString(`{"model":"gpt-4o","input":"contract"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	received, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, upstreamBody, received)
}

func responsesItemContractFixtures() []map[string]interface{} {
	items := make([]map[string]interface{}, 0, len(contractsopenai.ResponsesItemVariants()))
	for index, variant := range contractsopenai.ResponsesItemVariants() {
		item := map[string]interface{}{
			"id":               fmt.Sprintf("item_%d", index),
			"type":             variant.Type,
			"status":           "completed",
			"contract_variant": variant.Name,
			"future_nested_field": map[string]interface{}{
				"preserved": true,
			},
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
				map[string]interface{}{"type": "output_text", "text": "contract", "annotations": []interface{}{}},
			}
		}

		items = append(items, item)
	}

	return items
}

func assertJSONContainsEveryContractField(t *testing.T, body []byte, fields []string) {
	t.Helper()

	var decoded map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(body, &decoded))

	for _, field := range fields {
		assert.Containsf(t, decoded, field, "official response field %q was not relayed", field)
	}
}

func responsesStreamingContractFixture(eventType string, sequenceNumber int) map[string]interface{} {
	payload := map[string]interface{}{
		"type":            eventType,
		"sequence_number": sequenceNumber,
		"contract_values": map[string]interface{}{
			"null":   nil,
			"bool":   true,
			"number": sequenceNumber,
			"string": eventType,
			"array":  []interface{}{eventType},
			"object": map[string]interface{}{"preserved": true},
		},
	}

	switch eventType {
	case "response.created", "response.in_progress", "response.queued":
		payload["response"] = map[string]interface{}{
			"id": "resp_contract", "object": "response", "model": "gpt-4o", "status": "in_progress",
		}
	case "response.completed":
		payload["response"] = completeResponsesContractFixture([]map[string]interface{}{})
	case "response.failed":
		response := completeResponsesContractFixture([]map[string]interface{}{})
		response["status"] = "failed"
		response["error"] = map[string]interface{}{"code": "server_error", "message": "contract failure"}
		payload["response"] = response
	case "response.incomplete":
		response := completeResponsesContractFixture([]map[string]interface{}{})
		response["status"] = "incomplete"
		response["incomplete_details"] = map[string]interface{}{"reason": "max_output_tokens"}
		payload["response"] = response
	case "error":
		payload["code"] = nil
		payload["message"] = "contract streaming error"
		payload["param"] = nil
	}

	return payload
}
