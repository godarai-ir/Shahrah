package modelid

import "testing"

func TestSplitModelFQN_StripsRedundantProviderPrefix(t *testing.T) {
	provider, accountAlias, officialModel, ok := SplitModelFQN("openai:openai-daring-lotus:openai/gpt-4o-mini")
	if !ok {
		t.Fatalf("expected FQN to parse")
	}

	if provider != "openai" {
		t.Fatalf("expected provider openai, got %q", provider)
	}

	if accountAlias != "openai-daring-lotus" {
		t.Fatalf("expected account alias, got %q", accountAlias)
	}

	if officialModel != "gpt-4o-mini" {
		t.Fatalf("expected canonical model id, got %q", officialModel)
	}
}

func TestBuildModelFQN_CanonicalizesOfficialModel(t *testing.T) {
	got := BuildModelFQN("openai", "openai-daring-lotus", "openai/gpt-4o-mini")

	want := "openai:openai-daring-lotus:gpt-4o-mini"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
