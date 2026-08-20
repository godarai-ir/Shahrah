package domain

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

var (
	currencyAfterAmount  = regexp.MustCompile(`(?i)[0-9][0-9,]*(?:\.[0-9]+)?\s*(?:ریال|تومان|irr|usd|eur)`)
	currencyBeforeAmount = regexp.MustCompile(`(?i)(?:ریال|تومان|irr|usd|eur)\s*[0-9][0-9,]*(?:\.[0-9]+)?`)
	numericDate          = regexp.MustCompile(`\d{4}[/-]\d{1,2}[/-]\d{1,2}`)
)

// ValidateCulturalFacts prevents unsupported currency and calendar conversions.
func ValidateCulturalFacts(answer string, chunks []platform.RetrievalChunk) error {
	normalizedAnswer := NormalizePersian(answer)

	var evidence strings.Builder
	for _, chunk := range chunks {
		evidence.WriteString(NormalizePersian(chunk.Text))
		evidence.WriteRune('\n')
	}

	normalizedEvidence := evidence.String()

	currencyEvidence := culturalFactSet(normalizedEvidence, []*regexp.Regexp{currencyAfterAmount, currencyBeforeAmount}, canonicalCurrencyFact)
	for _, pattern := range []*regexp.Regexp{currencyAfterAmount, currencyBeforeAmount} {
		for _, fact := range pattern.FindAllString(normalizedAnswer, -1) {
			if _, exists := currencyEvidence[canonicalCurrencyFact(fact)]; !exists {
				return errors.New("answer contains a currency amount or date not present in context")
			}
		}
	}

	dateEvidence := culturalFactSet(normalizedEvidence, []*regexp.Regexp{numericDate}, canonicalDateFact)
	for _, fact := range numericDate.FindAllString(normalizedAnswer, -1) {
		if _, exists := dateEvidence[canonicalDateFact(fact)]; !exists {
			return errors.New("answer contains a currency amount or date not present in context")
		}
	}

	return nil
}

func culturalFactSet(input string, patterns []*regexp.Regexp, canonical func(string) string) map[string]struct{} {
	result := make(map[string]struct{})

	for _, pattern := range patterns {
		for _, fact := range pattern.FindAllString(input, -1) {
			result[canonical(fact)] = struct{}{}
		}
	}

	return result
}

func canonicalCurrencyFact(input string) string {
	input = strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(input, ",", ""), " ", ""))

	units := map[string]string{"ریال": "irr", "irr": "irr", "تومان": "toman", "usd": "usd", "eur": "eur"}
	for unit, canonical := range units {
		if strings.Contains(input, unit) {
			return canonical + "|" + strings.Replace(input, unit, "", 1)
		}
	}

	return input
}

func canonicalDateFact(input string) string {
	parts := strings.FieldsFunc(input, func(value rune) bool { return value == '/' || value == '-' })
	if len(parts) != 3 {
		return input
	}

	values := make([]int, 3)

	for index, part := range parts {
		value, err := strconv.Atoi(part)
		if err != nil {
			return input
		}

		values[index] = value
	}

	return fmt.Sprintf("%d/%d/%d", values[0], values[1], values[2])
}
