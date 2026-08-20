package domain

import (
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

func TestValidateCulturalFactsAcceptsNormalizedSourceValues(t *testing.T) {
	t.Parallel()

	chunks := []platform.RetrievalChunk{{Text: "مبلغ ۱۰۰٬۰۰۰ ریال در تاریخ ۱۴۰۵/۰۱/۱۵ ثبت شد."}}
	if err := ValidateCulturalFacts("مبلغ 100000 IRR در تاریخ ۱۴۰۵/۱/۱۵ ثبت شده است [1].", chunks); err != nil {
		t.Fatalf("validate supported cultural facts: %v", err)
	}
}

func TestValidateCulturalFactsRejectsCurrencyConversion(t *testing.T) {
	t.Parallel()

	chunks := []platform.RetrievalChunk{{Text: "مبلغ ۱۰۰٬۰۰۰ ریال است."}}
	if err := ValidateCulturalFacts("مبلغ ۱۰٬۰۰۰ تومان است [1].", chunks); err == nil {
		t.Fatal("expected an unsupported rial-to-toman conversion to fail")
	}
}

func TestValidateCulturalFactsRejectsCalendarConversion(t *testing.T) {
	t.Parallel()

	chunks := []platform.RetrievalChunk{{Text: "تاریخ قرارداد ۱۴۰۵/۱/۱۵ شمسی است."}}
	if err := ValidateCulturalFacts("تاریخ قرارداد 2026/4/4 است [1].", chunks); err == nil {
		t.Fatal("expected an unsupported calendar conversion to fail")
	}
}
