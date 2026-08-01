package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
)

func TestHashSocialHandoffCodeUsesSHA256AndNeverReturnsRawCode(t *testing.T) {
	raw := "sensitive-one-time-handoff-code"
	expectedBytes := sha256.Sum256([]byte(raw))
	expected := hex.EncodeToString(expectedBytes[:])

	actual := hashSocialHandoffCode("  " + raw + "  ")
	if actual != expected {
		t.Fatalf("expected SHA-256 hash %q, got %q", expected, actual)
	}

	if strings.Contains(actual, raw) || len(actual) != 64 {
		t.Fatalf("handoff hash must be a 64-character digest and must not expose the raw code")
	}
}

func TestRandomHexProducesIndependentSessionIdentifiers(t *testing.T) {
	first, err := randomHex(24)
	if err != nil {
		t.Fatal(err)
	}

	second, err := randomHex(24)
	if err != nil {
		t.Fatal(err)
	}

	if len(first) != 48 || len(second) != 48 || first == second {
		t.Fatalf("expected independent 192-bit hexadecimal identifiers")
	}
}
