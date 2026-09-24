package guardrails

import (
	"strings"
	"unicode"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
)

// LanguageClass selects which thresholds apply to a text.
type LanguageClass string

const (
	LanguagePersian  LanguageClass = "persian"
	LanguageFinglish LanguageClass = "finglish"
	LanguageEnglish  LanguageClass = "english"

	// maxSemanticContentRunes keeps state plus the longest question under Jev's 32k-token limit with room to spare.
	maxSemanticContentRunes = 48000
	maxSemanticPriorTurns   = 10

	persianScriptShare    = 0.3
	finglishTokenShare    = 0.15
	finglishMinTokenCount = 2
)

var persianNormalizer = strings.NewReplacer(
	"\u064a", "\u06cc", // Arabic yeh to Persian yeh.
	"\u0649", "\u06cc", // Alef maksura to Persian yeh.
	"\u0643", "\u06a9", // Arabic kaf to Persian keheh.
	"\u200b", "", // Zero-width space.
	"\u200d", "", // Zero-width joiner.
	"\u200e", "", // Left-to-right mark.
	"\u200f", "", // Right-to-left mark.
	"\u2060", "", // Word joiner.
	"\ufeff", "", // Byte order mark.
)

// finglishTokens are frequent Persian words in Latin script that are not common English words.
var finglishTokens = map[string]bool{
	"ro": true, "va": true, "ke": true, "az": true, "baraye": true, "bara": true, "chi": true, "chie": true, "chiye": true,
	"chetori": true, "chera": true, "kon": true, "bokon": true, "bede": true, "begoo": true, "begu": true, "bego": true,
	"mikham": true, "mikhay": true, "mikhad": true, "nist": true, "hast": true, "hastam": true, "hasti": true, "dare": true,
	"daram": true, "nadare": true, "nadaram": true, "khob": true, "lotfan": true, "salam": true, "mishe": true, "mitoonam": true,
	"mitooni": true, "kheili": true, "kheyli": true, "ye": true, "inja": true, "alan": true, "faghat": true, "hame": true,
	"chizi": true, "midam": true, "midi": true, "konam": true, "koni": true, "kardam": true, "kardi": true, "shode": true,
	"bashe": true, "bashi": true, "ghabli": true, "dige": true, "age": true, "vali": true, "hamin": true, "hamoon": true,
	"javab": true, "soal": true, "dastoor": true, "dastoorat": true, "ghanoon": true, "ghavanin": true, "pool": true,
	"poolet": true, "hesab": true, "ramz": true, "chand": true, "chande": true, "koja": true, "nemitoonam": true, "behet": true,
	"befrest": true, "bekhoon": true, "benevis": true, "bikhial": true, "ba": true,
}

// semanticState is the text a semantic batch evaluates, before it is wrapped for the judge.
type semanticState struct {
	Role       string
	Content    string
	PriorTurns []semanticTurn
	Truncated  bool
}

type semanticTurn struct {
	Role string `json:"role"`
	Text string `json:"text"`
}

// NormalizeSemanticText maps Arabic letter forms to Persian and removes invisible characters, so spelling tricks do not
// change what the judge sees. The zero-width non-joiner stays because Persian orthography needs it.
func NormalizeSemanticText(text string) string {
	return persianNormalizer.Replace(text)
}

// ClassifyLanguage separates Persian script, Persian written in Latin letters, and English.
func ClassifyLanguage(text string) LanguageClass {
	var arabicLetters, latinLetters int

	for _, r := range text {
		switch {
		case unicode.In(r, unicode.Arabic):
			arabicLetters++
		case unicode.In(r, unicode.Latin):
			latinLetters++
		}
	}

	letters := arabicLetters + latinLetters
	if letters == 0 {
		return LanguagePersian
	}

	if float64(arabicLetters) >= persianScriptShare*float64(letters) {
		return LanguagePersian
	}

	words := strings.FieldsFunc(strings.ToLower(text), func(r rune) bool { return !unicode.IsLetter(r) })
	hits := 0

	for _, word := range words {
		if finglishTokens[word] {
			hits++
		}
	}

	if hits >= finglishMinTokenCount && float64(hits) >= finglishTokenShare*float64(len(words)) {
		return LanguageFinglish
	}

	return LanguageEnglish
}

// semanticInputState evaluates the last user message. With the all scope, earlier user and assistant turns go along as
// context; system and tool messages never leave the gateway.
func semanticInputState(request contractsopenai.ChatCompletionRequest, scope Scope) semanticState {
	last := -1

	for i := len(request.Messages) - 1; i >= 0; i-- {
		if request.Messages[i].Role == "user" {
			last = i

			break
		}
	}

	if last < 0 {
		return semanticState{Role: "user"}
	}

	state := semanticState{Role: "user"}
	state.Content, state.Truncated = truncateTail(NormalizeSemanticText(contractsopenai.ChatContentText(request.Messages[last].Content)))

	if scope != ScopeAll {
		return state
	}

	for i := last - 1; i >= 0 && len(state.PriorTurns) < maxSemanticPriorTurns; i-- {
		message := request.Messages[i]
		if message.Role != "user" && message.Role != "assistant" {
			continue
		}

		text, truncated := truncateTail(NormalizeSemanticText(contractsopenai.ChatContentText(message.Content)))
		if strings.TrimSpace(text) == "" {
			continue
		}

		state.Truncated = state.Truncated || truncated
		state.PriorTurns = append([]semanticTurn{{Role: message.Role, Text: text}}, state.PriorTurns...)
	}

	return state
}

func semanticOutputState(response *contractsopenai.ChatCompletionResponse) semanticState {
	state := semanticState{Role: "assistant"}
	if response == nil {
		return state
	}

	texts, _ := outputTexts(*response)
	state.Content, state.Truncated = truncateTail(NormalizeSemanticText(strings.Join(texts, "\n\n")))

	return state
}

// judgeState wraps the text under an explicit untrusted label; the judge must never see it as bare instructions.
func (s semanticState) judgeState() map[string]any {
	state := map[string]any{"role": s.Role, "untrusted_content": s.Content}
	if len(s.PriorTurns) > 0 {
		state["prior_turns"] = s.PriorTurns
	}

	return state
}

// truncateTail keeps the end of the text, where the newest instructions are.
func truncateTail(text string) (string, bool) {
	runes := []rune(text)
	if len(runes) <= maxSemanticContentRunes {
		return text, false
	}

	return string(runes[len(runes)-maxSemanticContentRunes:]), true
}
