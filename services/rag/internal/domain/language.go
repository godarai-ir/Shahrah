package domain

import (
	"strings"
	"unicode"
)

// ResolveAnswerLanguage applies the published chatbot policy to the current query.
func ResolveAnswerLanguage(policy, defaultLocale, query string) string {
	switch strings.ToLower(strings.TrimSpace(policy)) {
	case "fa":
		return "fa"
	case "en":
		return "en"
	}

	persianLetters, latinLetters := 0, 0

	for _, value := range query {
		switch {
		case isArabicScript(value) && unicode.IsLetter(value):
			persianLetters++
		case unicode.In(value, unicode.Latin) && unicode.IsLetter(value):
			latinLetters++
		}
	}

	if persianLetters > latinLetters {
		return "fa"
	}

	if latinLetters > persianLetters {
		return "en"
	}

	if strings.EqualFold(strings.TrimSpace(defaultLocale), "en-US") {
		return "en"
	}

	return "fa"
}

func isArabicScript(value rune) bool {
	return value >= '\u0600' && value <= '\u06ff' ||
		value >= '\u0750' && value <= '\u077f' ||
		value >= '\u08a0' && value <= '\u08ff' ||
		value >= '\ufb50' && value <= '\ufdff' ||
		value >= '\ufe70' && value <= '\ufeff'
}
