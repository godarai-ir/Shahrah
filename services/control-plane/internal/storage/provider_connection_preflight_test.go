package storage

import (
	"context"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

func TestProviderConnectionPreflightDoesNotRequirePersistence(t *testing.T) {
	store := &Store{}

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
