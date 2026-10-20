package proxy

import (
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/stretchr/testify/require"
)

func TestAnthropicNativeAccountAllowlist(t *testing.T) {
	handler := &AnthropicNativeHandler{
		accountAllowlist: map[string]struct{}{"account-enabled": {}},
	}
	if !handler.accountEnabled("account-enabled") {
		t.Fatal("allowlisted account should be enabled")
	}

	if handler.accountEnabled("account-disabled") {
		t.Fatal("non-allowlisted account should be disabled")
	}
}

func TestAnthropicCodeExecutionPricingPolicyRejectsStandaloneExecution(t *testing.T) {
	body := []byte(`{
		"model":"claude-opus-4-7",
		"max_tokens":1024,
		"tools":[
			{"type":"code_execution_20260120","name":"code_execution"}
		]
	}`)

	conditions, err := anthropicCodeExecutionPricingPolicy(
		body,
		platform.ProviderOperationAnthropicMessages,
	)

	require.Error(t, err)
	require.Nil(t, conditions)
	require.Contains(t, err.Error(), "web_search_20260209")
}

func TestAnthropicCodeExecutionPricingPolicyAllowsOfficialFreeCombination(t *testing.T) {
	body := []byte(`{
		"model":"claude-opus-4-7",
		"max_tokens":1024,
		"tools":[
			{"type":"code_execution_20260120","name":"code_execution"},
			{"type":"web_search_20260209","name":"web_search"}
		]
	}`)

	conditions, err := anthropicCodeExecutionPricingPolicy(
		body,
		platform.ProviderOperationAnthropicMessages,
	)

	require.NoError(t, err)
	require.Equal(
		t,
		platform.AnthropicCodeExecutionFreeWithCurrentWebTools,
		conditions[platform.AnthropicCodeExecutionPricingCondition],
	)
}

func TestAnthropicCodeExecutionPricingPolicyPinsImplicitWebToolExecution(t *testing.T) {
	body := []byte(`{
		"model":"claude-opus-4-7",
		"tools":[
			{"type":"web_search_20260210","name":"web_search"}
		]
	}`)

	conditions, err := anthropicCodeExecutionPricingPolicy(
		body,
		platform.ProviderOperationAnthropicMessages,
	)

	require.NoError(t, err)
	require.Equal(
		t,
		platform.AnthropicCodeExecutionFreeWithCurrentWebTools,
		conditions[platform.AnthropicCodeExecutionPricingCondition],
	)
}

func TestAnthropicCodeExecutionPricingPolicyValidatesEveryBatchItem(t *testing.T) {
	body := []byte(`{
		"requests":[
			{
				"custom_id":"free",
				"params":{
					"model":"claude-opus-4-7",
					"tools":[
						{"type":"code_execution_20260120","name":"code_execution"},
						{"type":"web_fetch_20260209","name":"web_fetch"}
					]
				}
			},
			{
				"custom_id":"billable",
				"params":{
					"model":"claude-opus-4-7",
					"tools":[
						{"type":"code_execution_20260120","name":"code_execution"}
					]
				}
			}
		]
	}`)

	conditions, err := anthropicCodeExecutionPricingPolicy(
		body,
		platform.ProviderOperationAnthropicBatchesCreate,
	)

	require.Error(t, err)
	require.Nil(t, conditions)
	require.Contains(t, err.Error(), "batch item 1")
}

func TestAnthropicCodeExecutionPricingPolicyDoesNotTreatLegacyWebToolAsFree(
	t *testing.T,
) {
	body := []byte(`{
		"model":"claude-opus-4-7",
		"tools":[
			{"type":"code_execution_20250825","name":"code_execution"},
			{"type":"web_search_20250305","name":"web_search"}
		]
	}`)

	conditions, err := anthropicCodeExecutionPricingPolicy(
		body,
		platform.ProviderOperationAnthropicMessages,
	)

	require.Error(t, err)
	require.Nil(t, conditions)
}
