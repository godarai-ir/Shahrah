package guardrails

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// DocumentConfidentialityType is the tenant guardrail type that blocks attachments carrying a classification marking.
const DocumentConfidentialityType = "document_confidentiality"

// ConfidentialityLevel orders classification markings from least to most sensitive.
type ConfidentialityLevel int

const (
	LevelNone ConfidentialityLevel = iota
	LevelInternal
	LevelConfidential
	LevelHighlyConfidential
	LevelSecret
	LevelTopSecret
)

var confidentialityLevelNames = map[ConfidentialityLevel]string{
	LevelNone:               "none",
	LevelInternal:           "internal",
	LevelConfidential:       "confidential",
	LevelHighlyConfidential: "highly_confidential",
	LevelSecret:             "secret",
	LevelTopSecret:          "top_secret",
}

// String returns the config name of the level.
func (l ConfidentialityLevel) String() string {
	return confidentialityLevelNames[l]
}

// ParseConfidentialityLevel reads a config level name. The none level is not a valid block level.
func ParseConfidentialityLevel(value string) (ConfidentialityLevel, bool) {
	value = strings.ToLower(strings.TrimSpace(value))
	for level, name := range confidentialityLevelNames {
		if level != LevelNone && name == value {
			return level, true
		}
	}

	return LevelNone, false
}

const (
	defaultConfidentialityBlockLevel = LevelConfidential
	maxCustomMarkings                = 50
	maxCustomMarkingRunes            = 100
)

// ValidateDocumentConfidentialityConfig reports a document_confidentiality config that cannot be evaluated as intended.
func ValidateDocumentConfidentialityConfig(config map[string]interface{}) error {
	if raw, ok := config["block_level"]; ok {
		value, isString := raw.(string)
		if _, valid := ParseConfidentialityLevel(value); !isString || !valid {
			return errors.New("config.block_level must be internal, confidential, highly_confidential, secret, or top_secret")
		}
	}

	if raw, ok := config["include_message_text"]; ok {
		if _, isBool := raw.(bool); !isBool {
			return errors.New("config.include_message_text must be a boolean")
		}
	}

	raw, ok := config["custom_markings"]
	if !ok {
		return nil
	}

	values, isList := raw.([]interface{})
	if !isList {
		return errors.New("config.custom_markings must be a list of strings")
	}

	if len(values) > maxCustomMarkings {
		return fmt.Errorf("config.custom_markings allows at most %d entries", maxCustomMarkings)
	}

	for _, value := range values {
		text, isString := value.(string)
		if !isString || strings.TrimSpace(text) == "" || utf8.RuneCountInString(text) > maxCustomMarkingRunes {
			return fmt.Errorf("config.custom_markings entries must be 1-%d characters", maxCustomMarkingRunes)
		}
	}

	return nil
}

// confidentialityMarkings maps normalized marking phrases to levels. A marking only counts when it fills a whole line
// or a separated segment of a line, the way stamps, headers, footers, and sensitivity labels are laid out, so running
// prose such as "this email is confidential" does not match.
var confidentialityMarkings = map[string]ConfidentialityLevel{
	"به کلی سری":               LevelTopSecret,
	"بکلی سری":                 LevelTopSecret,
	"top secret":               LevelTopSecret,
	"خیلی محرمانه":             LevelHighlyConfidential,
	"highly confidential":      LevelHighlyConfidential,
	"strictly confidential":    LevelHighlyConfidential,
	"محرمانه":                  LevelConfidential,
	"confidential":             LevelConfidential,
	"company confidential":     LevelConfidential,
	"private and confidential": LevelConfidential,
	"فقط برای استفاده داخلی":   LevelInternal,
	"برای استفاده داخلی":       LevelInternal,
	"استفاده داخلی":            LevelInternal,
	"internal use only":        LevelInternal,
	"for internal use only":    LevelInternal,
	"internal only":            LevelInternal,
}

// labeledOnlyMarkings are ordinary words elsewhere: سری also means series, and a table cell may read Internal or
// Secret. They only count after a classification label such as "طبقه‌بندی:" or "Classification:".
var labeledOnlyMarkings = map[string]ConfidentialityLevel{
	"سری":        LevelSecret,
	"secret":     LevelSecret,
	"restricted": LevelConfidential,
	"داخلی":      LevelInternal,
	"internal":   LevelInternal,
}

// schemeOnlyTerms name levels that are never a stamp on their own. They only help recognize a listing of the
// classification scheme, such as the level table of a security policy.
var schemeOnlyTerms = map[string]bool{
	"public": true, "general": true, "unclassified": true, "عمومی": true, "عادی": true, "طبقه بندی نشده": true,
}

const (
	maxMarkingLineRunes = 120
	// schemeWindow is how many nearby lines on each side are searched for other level names.
	schemeWindow = 3
	// schemeNeighbors is how many other distinct level names nearby turn a bare line into a scheme listing.
	schemeNeighbors = 2
)

var (
	classificationLabel = regexp.MustCompile(`^(?:طبقه ?بندی(?: حفاظتی)?|درجه(?: حفاظتی| طبقه ?بندی)?|سطح(?: محرمانگی| دسترسی)|` +
		`classification|security classification|sensitivity(?: label)?|data classification)\s*[:：\-]\s*`)
	markingSeparators = regexp.MustCompile(`\s*(?://|[|/،,؛;–—]|\s-\s)\s*`)
	markingSpaces     = regexp.MustCompile(`[\s\x{200c}]+`)
)

// ConfidentialityFinding is the strongest marking found in one text.
type ConfidentialityFinding struct {
	Level   ConfidentialityLevel
	Marking string
}

// DetectConfidentiality returns the strongest classification marking in the text, or LevelNone. Custom markings are
// normalized phrases that count anywhere in the text at the given level.
func DetectConfidentiality(text string, custom []string, customLevel ConfidentialityLevel) ConfidentialityFinding {
	var best ConfidentialityFinding

	normalized := normalizeMarkingText(text)
	for _, marking := range custom {
		if marking != "" && strings.Contains(normalized, marking) && customLevel > best.Level {
			best = ConfidentialityFinding{Level: customLevel, Marking: marking}
		}
	}

	lines := markingLines(normalized)
	terms := make([]string, len(lines))

	for i, line := range lines {
		if isSchemeTerm(line) {
			terms[i] = line
		}
	}

	for i, line := range lines {
		if len([]rune(line)) > maxMarkingLineRunes {
			continue
		}

		labeled := false
		if loc := classificationLabel.FindStringIndex(line); loc != nil {
			labeled = true
			line = trimMarkingDecoration(line[loc[1]:])
		}

		// A bare level name next to other level names lists the scheme, like a policy's level table, rather than
		// stamping this document. Labeled lines such as "Classification: Confidential" always count.
		if !labeled && terms[i] != "" && inSchemeListing(terms, i) {
			continue
		}

		for _, segment := range markingSeparators.Split(line, -1) {
			segment = trimMarkingDecoration(segment)

			level, ok := confidentialityMarkings[segment]
			if !ok && labeled {
				level, ok = labeledOnlyMarkings[segment]
			}

			if ok && level > best.Level {
				best = ConfidentialityFinding{Level: level, Marking: segment}
			}
		}
	}

	return best
}

// markingLines returns the non-empty lines without decoration, so blank lines do not separate neighbors.
func markingLines(normalized string) []string {
	var lines []string

	for _, line := range strings.Split(normalized, "\n") {
		if line = trimMarkingDecoration(line); line != "" {
			lines = append(lines, line)
		}
	}

	return lines
}

func isSchemeTerm(line string) bool {
	_, marking := confidentialityMarkings[line]
	_, labeledOnly := labeledOnlyMarkings[line]

	return marking || labeledOnly || schemeOnlyTerms[line]
}

// inSchemeListing reports whether enough other distinct level names stand on nearby lines.
func inSchemeListing(terms []string, index int) bool {
	others := map[string]bool{}

	for j := max(0, index-schemeWindow); j <= min(len(terms)-1, index+schemeWindow); j++ {
		if j != index && terms[j] != "" && terms[j] != terms[index] {
			others[terms[j]] = true
		}
	}

	return len(others) >= schemeNeighbors
}

// normalizeMarkingText lowercases, maps Arabic letter forms to Persian, and collapses spaces and zero-width non-joiners
// inside each line, so "به‌کلی سری" and "به کلی  سری" read the same.
func normalizeMarkingText(text string) string {
	lines := strings.Split(NormalizeSemanticText(strings.ToLower(text)), "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(markingSpaces.ReplaceAllString(line, " "))
	}

	return strings.Join(lines, "\n")
}

func trimMarkingDecoration(value string) string {
	return strings.TrimFunc(value, func(r rune) bool {
		return unicode.IsSpace(r) || unicode.IsPunct(r) || unicode.IsSymbol(r) || r == '‌'
	})
}

type documentConfidentialityGuardrail struct {
	name               string
	enforcement        Enforcement
	blockLevel         ConfidentialityLevel
	customMarkings     []string
	includeMessageText bool
}

func newDocumentConfidentialityGuardrail(name string, enforcement Enforcement, config map[string]interface{}) Guardrail {
	blockLevel, ok := ParseConfidentialityLevel(stringFromConfig(config, "block_level", ""))
	if !ok {
		blockLevel = defaultConfidentialityBlockLevel
	}

	custom := stringSliceFromConfig(config, "custom_markings")
	normalized := make([]string, 0, len(custom))

	for _, marking := range custom {
		if value := normalizeMarkingText(marking); value != "" {
			normalized = append(normalized, value)
		}
	}

	return &documentConfidentialityGuardrail{
		name:               name,
		enforcement:        enforcement,
		blockLevel:         blockLevel,
		customMarkings:     normalized,
		includeMessageText: boolFromConfig(config, "include_message_text", false),
	}
}

func (g *documentConfidentialityGuardrail) Name() string             { return g.name }
func (g *documentConfidentialityGuardrail) Operation() Operation     { return OperationValidate }
func (g *documentConfidentialityGuardrail) Enforcement() Enforcement { return g.enforcement }

// Evaluate checks attachments, and the typed text when configured, against the block level. Typed text is off by
// default, because users often mention a document's classification while asking about it.
func (g *documentConfidentialityGuardrail) Evaluate(input EvaluationInput) (EvaluationResult, error) {
	result := EvaluationResult{Texts: input.Texts}

	for _, item := range input.Attachments {
		finding := DetectConfidentiality(item.Text, g.customMarkings, g.blockLevel)
		if finding.Level >= g.blockLevel {
			result.Blocked = true
			result.Reason = fmt.Sprintf("Attachment %q is marked %s", item.Name, finding.Level)

			return result, nil
		}
	}

	if !g.includeMessageText {
		return result, nil
	}

	for _, text := range input.Texts {
		finding := DetectConfidentiality(text, g.customMarkings, g.blockLevel)
		if finding.Level >= g.blockLevel {
			result.Blocked = true
			result.Reason = fmt.Sprintf("Message content is marked %s", finding.Level)

			return result, nil
		}
	}

	return result, nil
}
