package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

// mistralAdapter keeps Mistral's fixed upstream routes while applying only
// documented OpenAI-compatibility aliases. Native fields remain untouched.
type mistralAdapter struct {
	upstream providerAdapter
}

func newMistralAdapter() providerAdapter {
	return &mistralAdapter{
		upstream: newOpenAICompatibleAdapter("https://api.mistral.ai/v1"),
	}
}

func (a *mistralAdapter) Dispatch(
	ctx context.Context,
	client *http.Client,
	payload platform.TunnelRequest,
) (*adapterResponse, error) {
	var err error

	payload.Body, err = normalizeMistralJSONRequest(payload.Operation, payload.Body)
	if err != nil {
		return nil, err
	}

	return a.upstream.Dispatch(ctx, client, payload)
}

func normalizeMistralJSONRequest(
	operation platform.ProviderOperation,
	body []byte,
) ([]byte, error) {
	switch operation {
	case platform.ProviderOperationChatCompletions,
		platform.ProviderOperationMistralFIM,
		platform.ProviderOperationOpenAIEmbeddings:
	default:
		return body, nil
	}

	var object map[string]json.RawMessage
	if err := json.Unmarshal(body, &object); err != nil {
		return nil, fmt.Errorf("invalid Mistral JSON request: %w", err)
	}

	aliases := make([][2]string, 0, 2)
	if operation != platform.ProviderOperationChatCompletions {
		aliases = append(aliases, [2]string{"seed", "random_seed"})
	}

	if operation == platform.ProviderOperationOpenAIEmbeddings {
		aliases = append(aliases, [2]string{"dimensions", "output_dimension"})
	}

	changed := false

	for _, alias := range aliases {
		from, hasFrom := object[alias[0]]

		_, hasTo := object[alias[1]]
		if hasFrom && hasTo {
			return nil, fmt.Errorf("mistral request cannot contain both %s and %s", alias[0], alias[1])
		}

		if hasFrom {
			object[alias[1]] = from
			delete(object, alias[0])

			changed = true
		}
	}

	if !changed {
		return body, nil
	}

	return json.Marshal(object)
}
