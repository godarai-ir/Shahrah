package storage

import (
	"context"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/platformcredentials"
)

type recordingTenantCredentialProber struct {
	record   platformcredentials.EncryptedCredential
	secrets  map[string]string
	success  bool
	category string
	summary  string
}

func (p *recordingTenantCredentialProber) Probe(
	_ context.Context,
	record platformcredentials.EncryptedCredential,
	secrets map[string]string,
) (bool, string, string) {
	p.record = record

	p.secrets = make(map[string]string, len(secrets))
	for key, value := range secrets {
		p.secrets[key] = value
	}

	return p.success, p.category, p.summary
}

func TestProviderConnectionPreflightDoesNotRequirePersistence(t *testing.T) {
	store := &ProviderAccountRepository{}

	managed, err := store.TestProviderConnection(context.Background(), platform.TestProviderConnectionRequest{
		Provider:                      "openai",
		AuthMode:                      platform.ProviderAuthModeAPIKey,
		UsePlatformManagedCredentials: true,
	})
	if err != nil || managed == nil || !managed.Success {
		t.Fatalf("expected managed preflight success without a database: response=%#v err=%v", managed, err)
	}

	unsupported, err := store.TestProviderConnection(context.Background(), platform.TestProviderConnectionRequest{
		Provider: "aws-bedrock",
		AuthMode: platform.ProviderAuthModeAWSSigV4,
	})
	if err != nil || unsupported == nil || !unsupported.Unsupported {
		t.Fatalf("expected unsupported auth mode without a database: response=%#v err=%v", unsupported, err)
	}

	missing, err := store.TestProviderConnection(context.Background(), platform.TestProviderConnectionRequest{
		Provider: "openai",
		AuthMode: platform.ProviderAuthModeAPIKey,
		Credentials: &platform.ProviderAccountCredentials{
			APIKey: "********",
		},
	})
	if err != nil || missing == nil || missing.Success {
		t.Fatalf("expected masked credential rejection without a database: response=%#v err=%v", missing, err)
	}
}

func TestProviderConnectionPreflightUsesSharedProberForGoogleGemini(t *testing.T) {
	prober := &recordingTenantCredentialProber{
		success: true, category: "ok", summary: "Provider accepted the credential.",
	}
	store := &ProviderAccountRepository{tenantProviderProber: prober}

	result, err := store.TestProviderConnection(context.Background(), platform.TestProviderConnectionRequest{
		Provider: "google-gemini",
		AuthMode: platform.ProviderAuthModeAPIKey,
		ProviderConfig: map[string]interface{}{
			"native_base_url": "https://generativelanguage.googleapis.com",
		},
		Credentials: &platform.ProviderAccountCredentials{APIKey: "gemini-secret"},
	})
	if err != nil {
		t.Fatalf("preflight returned an error: %v", err)
	}

	if result == nil || !result.Success || result.Unsupported {
		t.Fatalf("expected successful shared probe result, got %#v", result)
	}

	if prober.record.ProviderSlug != "google-gemini" ||
		prober.record.BaseURL != "https://generativelanguage.googleapis.com" ||
		prober.record.AuthMode != "api_key" {
		t.Fatalf("unexpected probe record: %#v", prober.record)
	}

	if prober.secrets["api_key"] != "gemini-secret" {
		t.Fatalf("Gemini API key was not passed to the shared prober")
	}
}

func TestProviderConnectionPreflightPreservesUnsupportedCategory(t *testing.T) {
	prober := &recordingTenantCredentialProber{
		category: "unsupported", summary: "This provider authentication mode has no implemented non-billable validation probe.",
	}
	store := &ProviderAccountRepository{tenantProviderProber: prober}

	result, err := store.TestProviderConnection(context.Background(), platform.TestProviderConnectionRequest{
		Provider:    "deepgram",
		AuthMode:    platform.ProviderAuthModeAPIKey,
		Credentials: &platform.ProviderAccountCredentials{APIKey: "deepgram-secret"},
	})
	if err != nil {
		t.Fatalf("preflight returned an error: %v", err)
	}

	if result == nil || result.Success || !result.Unsupported {
		t.Fatalf("expected unsupported shared probe result, got %#v", result)
	}
}

func TestProviderConnectionPreflightMapsPersonalAccessTokenForSharedProber(t *testing.T) {
	prober := &recordingTenantCredentialProber{success: true, category: "ok", summary: "accepted"}
	store := &ProviderAccountRepository{tenantProviderProber: prober}

	result, err := store.TestProviderConnection(context.Background(), platform.TestProviderConnectionRequest{
		Provider: "databricks",
		AuthMode: platform.ProviderAuthMode("personal_access_token"),
		ProviderConfig: map[string]interface{}{
			"serving_endpoint_url": "https://workspace.cloud.databricks.com/serving-endpoints/example",
		},
		Credentials: &platform.ProviderAccountCredentials{
			Raw: map[string]interface{}{"access_token": "databricks-token"},
		},
	})
	if err != nil || result == nil || !result.Success {
		t.Fatalf("expected successful Databricks preflight: response=%#v err=%v", result, err)
	}

	if prober.record.BaseURL != "https://workspace.cloud.databricks.com/serving-endpoints/example" ||
		prober.secrets["access_token"] != "databricks-token" {
		t.Fatalf("unexpected Databricks probe input: record=%#v secrets=%#v", prober.record, prober.secrets)
	}
}
