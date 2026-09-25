package guardrails

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/services/control-plane/internal/docextract"
)

const (
	// AttachmentSelector names the check recorded when an inline attachment cannot be inspected.
	AttachmentSelector = "attachment_inspection"

	maxInspectedAttachments = 20

	// maxAttachmentJudgeRunes leaves room for the typed message within maxSemanticContentRunes.
	maxAttachmentJudgeRunes = 36000
	minAttachmentJudgeRunes = 4000
)

// attachment is the extracted text of one inline file part.
type attachment struct {
	MessageIndex int
	Name         string
	Text         string
}

// attachmentFailure is an inline document the gateway could not read, so content guardrails never saw it.
type attachmentFailure struct {
	Name string
	Err  error
}

// inputMessageRange returns the half-open range of messages the scope inspects.
func inputMessageRange(req contractsopenai.ChatCompletionRequest, scope Scope) (int, int) {
	if scope != ScopeLast {
		return 0, len(req.Messages)
	}

	for i := len(req.Messages) - 1; i >= 0; i-- {
		if req.Messages[i].Role == "user" {
			return i, i + 1
		}
	}

	if len(req.Messages) == 0 {
		return 0, 0
	}

	return len(req.Messages) - 1, len(req.Messages)
}

// inputAttachments extracts the text of every inline file part in the scope. File IDs and remote URLs are not
// resolved here, and non-document formats such as images are skipped because other guardrails own them.
func inputAttachments(req contractsopenai.ChatCompletionRequest, scope Scope, limits docextract.Limits, cache *extractionCache) ([]attachment, []attachmentFailure) {
	var (
		attachments []attachment
		failures    []attachmentFailure
	)

	start, end := inputMessageRange(req, scope)
	for i := start; i < end; i++ {
		parts, ok, err := contractsopenai.ChatContentParts(req.Messages[i].Content)
		if !ok || err != nil {
			continue
		}

		for partIndex, part := range parts {
			name, dataURL := inlineFilePart(part, partIndex)
			if dataURL == "" {
				continue
			}

			if len(attachments)+len(failures) >= maxInspectedAttachments {
				failures = append(failures, attachmentFailure{Name: name, Err: fmt.Errorf("%w: more than %d attachments in one request", docextract.ErrTooLarge, maxInspectedAttachments)})

				continue
			}

			text, err := cachedExtract(cache, dataURL, name, limits)

			switch {
			case errors.Is(err, docextract.ErrUnsupported):
				continue
			case err != nil:
				failures = append(failures, attachmentFailure{Name: name, Err: err})
			default:
				attachments = append(attachments, attachment{MessageIndex: i, Name: name, Text: text})
			}
		}
	}

	return attachments, failures
}

// inlineFilePart returns the display name and data URL of a file part that carries its bytes inline.
func inlineFilePart(part contractsopenai.ChatContentPart, partIndex int) (string, string) {
	name := fmt.Sprintf("attachment %d", partIndex+1)

	switch {
	case part.Type == "file" && part.File != nil:
		if filename := strings.TrimSpace(part.File.Filename); filename != "" {
			name = filename
		}

		return name, strings.TrimSpace(part.File.FileData)
	case part.Type == "file_url" && part.FileURL != nil && strings.HasPrefix(strings.ToLower(strings.TrimSpace(part.FileURL.URL)), "data:"):
		return name, strings.TrimSpace(part.FileURL.URL)
	default:
		return name, ""
	}
}

// cachedExtract returns the cached result for identical attachment bytes, so a document resent on every turn of a
// conversation is parsed once.
func cachedExtract(cache *extractionCache, dataURL string, name string, limits docextract.Limits) (string, error) {
	key := extractionCacheKey(dataURL)
	if text, err, ok := cache.get(key); ok {
		return text, err
	}

	text, err := extractDataURL(dataURL, name, limits)
	cache.put(key, text, err)

	return text, err
}

func extractDataURL(dataURL string, name string, limits docextract.Limits) (string, error) {
	mediaType, encoded, err := contractsopenai.ParseFileDataURL(dataURL)
	if err != nil {
		return "", fmt.Errorf("%w: %w", docextract.ErrMalformed, err)
	}

	if base64.StdEncoding.DecodedLen(len(encoded)) > limits.MaxFileBytes {
		return "", fmt.Errorf("%w: file is larger than %d bytes", docextract.ErrTooLarge, limits.MaxFileBytes)
	}

	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("%w: invalid base64 payload: %w", docextract.ErrMalformed, err)
	}

	doc, err := docextract.Extract(data, mediaType, name, limits)
	if err != nil {
		return "", err
	}

	return doc.Text, nil
}

// inspectsAttachments reports whether a guardrail looks at content, as opposed to the shape of the user's message.
// Count, JSON, code, and token checks describe the typed message and would misfire on document text.
func inspectsAttachments(g Guardrail) bool {
	switch g.(type) {
	case *patternGuardrail, *keywordBlocklistGuardrail, *regexReplaceGuardrail, *semanticPolicyGuardrail, *documentConfidentialityGuardrail:
		return true
	default:
		return false
	}
}

// attachmentEnforcement returns the strictest enforcement among the selected guardrails that inspect attachments, and
// false when none of them do, in which case attachments are not extracted at all.
func (e *Engine) attachmentEnforcement(selectors []string, overrides map[string]Guardrail) (Enforcement, bool) {
	strictest, found := EnforcementAudit, false

	for _, selector := range selectors {
		g := e.resolve(selector, overrides)
		if g == nil || !inspectsAttachments(g) {
			continue
		}

		found = true

		switch g.Enforcement() {
		case EnforcementEnforce:
			return EnforcementEnforce, true
		case EnforcementEnforceIgnoreError:
			strictest = EnforcementEnforceIgnoreError
		}
	}

	return strictest, found
}

// attachmentFailureCheck records unreadable attachments. Only enforce fails closed; the other modes let the request
// through with the attachment uninspected.
func attachmentFailureCheck(failures []attachmentFailure, enforcement Enforcement) Check {
	details := make([]string, 0, len(failures))
	for _, failure := range failures {
		details = append(details, fmt.Sprintf("%s: %v", failure.Name, failure.Err))
	}

	return Check{
		Selector:    AttachmentSelector,
		Status:      "error",
		Operation:   OperationValidate,
		Enforcement: enforcement,
		Reason:      "Attachment could not be inspected",
		Error:       strings.Join(details, "; "),
	}
}

// attachmentInputs returns the attachments in order, for deterministic guardrails.
func attachmentInputs(attachments []attachment) []AttachmentText {
	files := make([]AttachmentText, 0, len(attachments))
	for _, item := range attachments {
		files = append(files, AttachmentText{Name: item.Name, Text: item.Text})
	}

	return files
}

// attachmentsChanged reports whether a mutating guardrail rewrote attachment text. Attachments are forwarded as the
// original files, so a redaction there cannot be applied and must be treated as a violation instead.
func attachmentsChanged(before []string, after []string) bool {
	if len(after) != len(before) {
		return true
	}

	for i := range before {
		if before[i] != after[i] {
			return true
		}
	}

	return false
}

// attachmentBlocks groups the attachment texts of each message under a label, for the semantic judge. A long attachment
// keeps its head and tail within a per-attachment share of the judge budget, because classification stamps, titles,
// and signatures sit at the ends of a document. It reports whether any attachment was shortened.
func attachmentBlocks(attachments []attachment) (map[int]string, bool) {
	blocks := map[int]string{}
	sampled := false

	budget := maxAttachmentJudgeRunes
	if len(attachments) > 0 {
		budget = max(minAttachmentJudgeRunes, maxAttachmentJudgeRunes/len(attachments))
	}

	for _, item := range attachments {
		text := strings.TrimSpace(item.Text)
		if text == "" {
			continue
		}

		text, shortened := sampleHeadTail(text, budget)
		sampled = sampled || shortened

		block := fmt.Sprintf("[Attachment: %s]\n%s", item.Name, text)
		if existing := blocks[item.MessageIndex]; existing != "" {
			block = existing + "\n\n" + block
		}

		blocks[item.MessageIndex] = block
	}

	return blocks, sampled
}

// sampleHeadTail keeps two thirds of the budget from the start of the text and one third from the end.
func sampleHeadTail(text string, budget int) (string, bool) {
	runes := []rune(text)
	if len(runes) <= budget {
		return text, false
	}

	head := budget * 2 / 3
	tail := budget - head
	omitted := len(runes) - head - tail

	return fmt.Sprintf("%s\n[... %d characters omitted ...]\n%s", string(runes[:head]), omitted, string(runes[len(runes)-tail:])), true
}
