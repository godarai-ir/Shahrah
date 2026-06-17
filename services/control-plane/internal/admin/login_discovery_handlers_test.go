package admin

import "testing"

func TestNormalizeDiscoveryIdentifier(t *testing.T) {
	channel, value, ok := normalizeDiscoveryIdentifier("User@Example.COM")
	if !ok || channel != "email" || value != "user@example.com" {
		t.Fatalf("expected normalized email, got channel=%q value=%q ok=%v", channel, value, ok)
	}

	channel, value, ok = normalizeDiscoveryIdentifier("0098912 345-6789")
	if !ok || channel != "mobile" || value != "+989123456789" {
		t.Fatalf("expected normalized mobile, got channel=%q value=%q ok=%v", channel, value, ok)
	}

	if _, _, ok = normalizeDiscoveryIdentifier("not-a-login"); ok {
		t.Fatal("expected invalid identifier to be rejected")
	}
}
