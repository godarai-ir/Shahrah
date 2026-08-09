package proxy

import (
	"bytes"
	"encoding/json"
	"testing"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
)

func TestBuildChatCompletionPayloadPreservesModelDecodingControls(t *testing.T) {
	originalBody := []byte(`{
		"model":"public-model",
		"messages":[{"role":"user","content":"hello"}],
		"temperature":7.25,
		"top_p":-0.5,
		"top_k":"provider-native",
		"presence_penalty":3,
		"frequency_penalty":-4,
		"seed":"stable-seed",
		"random_seed":42,
		"stop":{"provider":"native"},
		"logit_bias":{"42":100},
		"logprobs":"provider-default",
		"top_logprobs":99
	}`)

	var request contractsopenai.ChatCompletionRequest
	if err := json.Unmarshal(originalBody, &request); err != nil {
		t.Fatalf("unmarshal request: %v", err)
	}

	outboundBody, err := buildChatCompletionPayload(request, "upstream-model", "fallback-model")
	if err != nil {
		t.Fatalf("build payload: %v", err)
	}

	var original map[string]json.RawMessage
	if err := json.Unmarshal(originalBody, &original); err != nil {
		t.Fatalf("unmarshal original fields: %v", err)
	}

	var outbound map[string]json.RawMessage
	if err := json.Unmarshal(outboundBody, &outbound); err != nil {
		t.Fatalf("unmarshal outbound fields: %v", err)
	}

	for name := range original {
		if !contractsopenai.IsModelDecodingPassThroughParameter(name) {
			continue
		}

		if !bytes.Equal(original[name], outbound[name]) {
			t.Fatalf("%s changed: got %s, want %s", name, outbound[name], original[name])
		}
	}

	var outboundModel string
	if err := json.Unmarshal(outbound["model"], &outboundModel); err != nil {
		t.Fatalf("decode outbound model: %v", err)
	}

	if outboundModel != "upstream-model" {
		t.Fatalf("model = %q, want upstream-model", outboundModel)
	}
}

func TestBuildChatCompletionPayloadDoesNotInjectModelDecodingControls(t *testing.T) {
	request := contractsopenai.ChatCompletionRequest{
		Model:    "public-model",
		Messages: []contractsopenai.ChatMessage{{Role: "user", Content: "hello"}},
	}

	outboundBody, err := buildChatCompletionPayload(request, "upstream-model", "fallback-model")
	if err != nil {
		t.Fatalf("build payload: %v", err)
	}

	var outbound map[string]json.RawMessage
	if err := json.Unmarshal(outboundBody, &outbound); err != nil {
		t.Fatalf("unmarshal outbound fields: %v", err)
	}

	for _, name := range []string{
		"temperature", "top_p", "top_k", "presence_penalty",
		"frequency_penalty", "seed", "random_seed", "stop",
		"logit_bias", "logprobs", "top_logprobs",
	} {
		if _, present := outbound[name]; present {
			t.Fatalf("%s was injected: %s", name, outboundBody)
		}
	}
}
