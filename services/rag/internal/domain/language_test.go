package domain

import "testing"

func TestResolveAnswerLanguage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name, policy, locale, query, expected string
	}{
		{name: "Persian query", policy: "query", locale: "en-US", query: "شرایط بازگشت چیست؟", expected: "fa"},
		{name: "Arabic keyboard Persian", policy: "query", locale: "fa-IR", query: "شرايط كالا", expected: "fa"},
		{name: "English query", policy: "query", locale: "fa-IR", query: "What is the return policy?", expected: "en"},
		{name: "Code uses Persian default", policy: "query", locale: "fa-IR", query: "SKU-123", expected: "en"},
		{name: "Digits use default", policy: "query", locale: "fa-IR", query: "۱۲۳۴", expected: "fa"},
		{name: "Forced Persian", policy: "fa", locale: "en-US", query: "English", expected: "fa"},
		{name: "Forced English", policy: "en", locale: "fa-IR", query: "فارسی", expected: "en"},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			actual := ResolveAnswerLanguage(test.policy, test.locale, test.query)
			if actual != test.expected {
				t.Fatalf("expected %q, got %q", test.expected, actual)
			}
		})
	}
}
