package responsecache

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
)

const semanticFinalUserText = "__semantic_final_user_text__"

// LookupResponse finds an exact or semantic cache entry for a Responses API request.
func (s *Service) LookupResponse(
	ctx context.Context,
	scope Scope,
	req contractsopenai.ResponseRequest,
	rawRequest []byte,
	cfg RequestConfig,
) (Result, *Pending, error) {
	exact, err := decodeResponseRequest(rawRequest)
	if err != nil {
		return Result{Status: "error", CacheType: cfg.Type}, nil, err
	}

	usesProviderState, stateErr := responseRequestUsesProviderState(exact, req)
	if stateErr != nil {
		return Result{Status: "error", CacheType: cfg.Type}, nil, stateErr
	}

	if usesProviderState {
		return Result{Status: "bypass", CacheType: cfg.Type}, nil, nil
	}

	fingerprint := requestFingerprint{exact: exact}

	if cfg.Type == "smart" {
		if req.Stream || len(req.Tools) > 0 || req.ToolChoice != nil || s.embedder == nil {
			return Result{Status: "bypass", CacheType: cfg.Type}, nil, nil
		}

		text, maskedInput, eligible, inputErr := semanticResponseInput(exact["input"])
		if inputErr != nil {
			return Result{Status: "error", CacheType: cfg.Type}, nil, inputErr
		}

		if !eligible {
			return Result{Status: "bypass", CacheType: cfg.Type}, nil, nil
		}

		semantic := cloneRawObject(exact)
		semantic["input"] = maskedInput
		fingerprint.semantic = semantic
		fingerprint.semanticText = text
	}

	scope.Operation = "responses"

	return s.lookup(ctx, scope, fingerprint, cfg)
}

func responseRequestUsesProviderState(
	request map[string]json.RawMessage,
	req contractsopenai.ResponseRequest,
) (bool, error) {
	if req.Stream || strings.TrimSpace(req.PreviousResponseID) != "" {
		return true, nil
	}

	conversation, hasConversation := request["conversation"]
	if hasConversation && !isNullJSON(conversation) {
		return true, nil
	}

	background, hasBackground := request["background"]
	if !hasBackground || isNullJSON(background) {
		return false, nil
	}

	var enabled bool
	if err := json.Unmarshal(background, &enabled); err != nil {
		return false, err
	}

	return enabled, nil
}

func isNullJSON(raw json.RawMessage) bool {
	trimmed := bytes.TrimSpace(raw)

	return len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null"))
}

func decodeResponseRequest(rawRequest []byte) (map[string]json.RawMessage, error) {
	decoder := json.NewDecoder(strings.NewReader(string(rawRequest)))

	var request map[string]json.RawMessage
	if err := decoder.Decode(&request); err != nil {
		return nil, err
	}

	var trailing interface{}
	if err := decoder.Decode(&trailing); err != io.EOF {
		return nil, errors.New("response cache request contains multiple JSON values")
	}

	if request == nil {
		return nil, errors.New("response cache request must be a JSON object")
	}

	return request, nil
}

func semanticResponseInput(raw json.RawMessage) (string, json.RawMessage, bool, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return "", nil, false, nil
	}

	if trimmed[0] == '"' {
		var text string
		if err := json.Unmarshal(trimmed, &text); err != nil {
			return "", nil, false, err
		}

		text = strings.TrimSpace(text)
		if text == "" {
			return "", nil, false, nil
		}

		masked, err := json.Marshal(semanticFinalUserText)

		return text, masked, true, err
	}

	if trimmed[0] != '[' {
		return "", nil, false, nil
	}

	var items []json.RawMessage
	if err := json.Unmarshal(trimmed, &items); err != nil {
		return "", nil, false, err
	}

	if len(items) == 0 {
		return "", nil, false, nil
	}

	lastIndex := len(items) - 1

	var message map[string]json.RawMessage
	if err := json.Unmarshal(items[lastIndex], &message); err != nil {
		return "", nil, false, err
	}

	roleRaw := message["role"]
	if len(roleRaw) == 0 {
		return "", nil, false, nil
	}

	var role string
	if err := json.Unmarshal(roleRaw, &role); err != nil {
		return "", nil, false, err
	}

	if !strings.EqualFold(strings.TrimSpace(role), "user") {
		return "", nil, false, nil
	}

	text, maskedContent, eligible, err := semanticResponseContent(message["content"])
	if err != nil || !eligible {
		return "", nil, eligible, err
	}

	maskedMessage := cloneRawObject(message)
	maskedMessage["content"] = maskedContent

	items[lastIndex], err = json.Marshal(maskedMessage)
	if err != nil {
		return "", nil, false, err
	}

	maskedInput, err := json.Marshal(items)

	return text, maskedInput, true, err
}

func semanticResponseContent(raw json.RawMessage) (string, json.RawMessage, bool, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return "", nil, false, nil
	}

	if trimmed[0] == '"' {
		var text string
		if err := json.Unmarshal(trimmed, &text); err != nil {
			return "", nil, false, err
		}

		text = strings.TrimSpace(text)
		if text == "" {
			return "", nil, false, nil
		}

		masked, err := json.Marshal(semanticFinalUserText)

		return text, masked, true, err
	}

	if trimmed[0] != '[' {
		return "", nil, false, nil
	}

	var parts []map[string]json.RawMessage
	if err := json.Unmarshal(trimmed, &parts); err != nil {
		return "", nil, false, err
	}

	if len(parts) == 0 {
		return "", nil, false, nil
	}

	texts := make([]string, 0, len(parts))
	for index := range parts {
		partTypeRaw := parts[index]["type"]
		if len(partTypeRaw) == 0 {
			return "", nil, false, nil
		}

		var partType string
		if err := json.Unmarshal(partTypeRaw, &partType); err != nil {
			return "", nil, false, err
		}

		if !strings.EqualFold(strings.TrimSpace(partType), "input_text") {
			return "", nil, false, nil
		}

		partTextRaw := parts[index]["text"]
		if len(partTextRaw) == 0 {
			return "", nil, false, nil
		}

		var partText string
		if err := json.Unmarshal(partTextRaw, &partText); err != nil {
			return "", nil, false, err
		}

		if strings.TrimSpace(partText) == "" {
			return "", nil, false, nil
		}

		texts = append(texts, strings.TrimSpace(partText))
		parts[index] = cloneRawObject(parts[index])

		maskedText, err := json.Marshal(semanticFinalUserText)
		if err != nil {
			return "", nil, false, err
		}

		parts[index]["text"] = maskedText
	}

	masked, err := json.Marshal(parts)

	return strings.Join(texts, "\n"), masked, true, err
}

func cloneRawObject(source map[string]json.RawMessage) map[string]json.RawMessage {
	clone := make(map[string]json.RawMessage, len(source))
	for key, value := range source {
		clone[key] = append(json.RawMessage(nil), value...)
	}

	return clone
}
