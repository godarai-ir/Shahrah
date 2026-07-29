package proxy

import (
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

func TestAnthropicNativeFeatureFlagsAreIndependent(t *testing.T) {
	handler := &AnthropicNativeHandler{
		nativeEnabled:  true,
		filesEnabled:   false,
		batchesEnabled: true,
	}

	if !handler.operationEnabled(platform.ProviderOperationAnthropicMessages) {
		t.Fatal("messages should remain enabled")
	}
	if handler.operationEnabled(platform.ProviderOperationAnthropicFilesCreate) {
		t.Fatal("files should be disabled independently")
	}
	if !handler.operationEnabled(platform.ProviderOperationAnthropicBatchesCreate) {
		t.Fatal("batches should remain enabled")
	}
}

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
