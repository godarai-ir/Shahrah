package guardrails

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/stretchr/testify/require"
)

func TestDetectConfidentialityMarkings(t *testing.T) {
	cases := map[string]struct {
		text string
		want ConfidentialityLevel
	}{
		"persian stamp line":              {text: "گزارش عملکرد\nمحرمانه\nمتن گزارش", want: LevelConfidential},
		"highly confidential wins":        {text: "محرمانه\nخیلی محرمانه", want: LevelHighlyConfidential},
		"top secret with zwnj":            {text: "** به\u200cکلی سری **", want: LevelTopSecret},
		"arabic letter forms":             {text: "خيلي محرمانه", want: LevelHighlyConfidential},
		"english label segment":           {text: "CONFIDENTIAL // INTERNAL USE ONLY", want: LevelConfidential},
		"sensitivity label value":         {text: "Highly Confidential", want: LevelHighlyConfidential},
		"labeled persian secret":          {text: "طبقه\u200cبندی: سری", want: LevelSecret},
		"labeled english restricted":      {text: "Classification: Restricted", want: LevelConfidential},
		"series is not secret":            {text: "سری\nمحصولات سری جدید", want: LevelNone},
		"table cell internal":             {text: "Internal\nExternal", want: LevelNone},
		"email disclaimer prose":          {text: "This email is confidential and intended solely for the addressee.", want: LevelNone},
		"prose mentioning classified":     {text: "لطفا بگو اسناد محرمانه چطور طبقه\u200cبندی می\u200cشوند؟", want: LevelNone},
		"internal use persian":            {text: "فقط برای استفاده داخلی", want: LevelInternal},
		"marking inside long line skips":  {text: strings.Repeat("متن ", 40) + "محرمانه", want: LevelNone},
		"policy level table":              {text: "Level\nPublic\nInternal\nConfidential\nSecret\nDescription", want: LevelNone},
		"persian level table":             {text: "سطح\nعادی\nمحرمانه\nخیلی محرمانه\nسری", want: LevelNone},
		"stamp away from level table":     {text: "Confidential\n" + strings.Repeat("body text\n", 5) + "Public\nInternal\nConfidential\nSecret", want: LevelConfidential},
		"labeled line inside level table": {text: "Public\nInternal\nClassification: Confidential\nSecret", want: LevelConfidential},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, tc.want, DetectConfidentiality(tc.text, nil, LevelConfidential).Level)
		})
	}
}

func TestDetectConfidentialityCustomMarkings(t *testing.T) {
	custom := []string{normalizeMarkingText("Project Falcon")}

	finding := DetectConfidentiality("notes about project  falcon launch", custom, LevelSecret)
	require.Equal(t, LevelSecret, finding.Level)
	require.Equal(t, "project falcon", finding.Marking)
}

func TestDocumentConfidentialityGuardrailBlocksMarkedAttachments(t *testing.T) {
	engine := NewDefaultEngine()
	marked := filePart("budget.txt", "text/plain", []byte("بودجه ۱۴۰۵\nخیلی محرمانه\nجزئیات"))
	internal := filePart("memo.txt", "text/plain", []byte("For internal use only\nmemo"))

	cases := map[string]struct {
		config      map[string]interface{}
		enforcement platform.GuardrailEnforcement
		parts       []contractsopenai.ChatContentPart
		wantStatus  string
		wantBlocked bool
	}{
		"default level blocks highly confidential": {parts: []contractsopenai.ChatContentPart{textPart("summarize"), marked}, wantStatus: "blocked", wantBlocked: true},
		"default level passes internal":            {parts: []contractsopenai.ChatContentPart{textPart("summarize"), internal}, wantStatus: "passed"},
		"internal level blocks internal":           {config: map[string]interface{}{"block_level": "internal"}, parts: []contractsopenai.ChatContentPart{internal}, wantStatus: "blocked", wantBlocked: true},
		"secret level passes highly confidential":  {config: map[string]interface{}{"block_level": "secret"}, parts: []contractsopenai.ChatContentPart{marked}, wantStatus: "passed"},
		"audit flags":                   {enforcement: platform.GuardrailEnforcementAudit, parts: []contractsopenai.ChatContentPart{marked}, wantStatus: "flagged"},
		"typed text ignored by default": {parts: []contractsopenai.ChatContentPart{textPart("محرمانه")}, wantStatus: "passed"},
		"typed text when enabled":       {config: map[string]interface{}{"include_message_text": true}, parts: []contractsopenai.ChatContentPart{textPart("محرمانه")}, wantStatus: "blocked", wantBlocked: true},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			enforcement := tc.enforcement
			if enforcement == "" {
				enforcement = platform.GuardrailEnforcementEnforce
			}

			overrides := map[string]Guardrail{"conf": tenantGuardrail("conf", DocumentConfidentialityType, platform.GuardrailOperationValidate, enforcement, tc.config)}

			result, err := engine.ApplyInputWithGuardrails(context.Background(), requestWithParts(tc.parts...), RequestSelection{LLMInput: []string{"conf"}}, overrides)
			require.Equal(t, tc.wantBlocked, err != nil)
			require.Equal(t, tc.wantStatus, result.Checks["llm_input_guardrails"][0].Status)
		})
	}
}

func TestDocumentConfidentialityReasonNamesAttachment(t *testing.T) {
	engine := NewDefaultEngine()
	overrides := map[string]Guardrail{"conf": tenantGuardrail("conf", DocumentConfidentialityType, platform.GuardrailOperationValidate, platform.GuardrailEnforcementEnforce, nil)}
	req := requestWithParts(filePart("plan.docx", "application/octet-stream", testDOCX(t, "Top Secret")))

	result, err := engine.ApplyInputWithGuardrails(context.Background(), req, RequestSelection{LLMInput: []string{"conf"}}, overrides)
	require.Error(t, err)
	require.Equal(t, `Attachment "plan.docx" is marked top_secret`, result.Reason)
}

func TestValidateDocumentConfidentialityConfig(t *testing.T) {
	require.NoError(t, ValidateDocumentConfidentialityConfig(nil))
	require.NoError(t, ValidateDocumentConfidentialityConfig(map[string]interface{}{
		"block_level": "secret", "include_message_text": true, "custom_markings": []interface{}{"Project Falcon"},
	}))

	for name, config := range map[string]map[string]interface{}{
		"unknown level":      {"block_level": "none"},
		"level not a string": {"block_level": 2},
		"flag not a bool":    {"include_message_text": "yes"},
		"markings not list":  {"custom_markings": "Falcon"},
		"empty marking":      {"custom_markings": []interface{}{" "}},
		"long marking":       {"custom_markings": []interface{}{strings.Repeat("a", maxCustomMarkingRunes+1)}},
	} {
		t.Run(name, func(t *testing.T) {
			require.Error(t, ValidateDocumentConfidentialityConfig(config))
		})
	}
}

func TestConfidentialityPackJudgesAttachments(t *testing.T) {
	rules := semanticPacks[ConfidentialityPack].rules[HookLLMInput]
	keys := make([]string, 0, len(rules))

	for _, rule := range rules {
		keys = append(keys, rule.Key)
	}

	require.Equal(t, []string{"jailbreak", "system_prompt_extraction", "evaluator_address", "confidential_document", "non_public_business_information", "personal_records", "security_internals"}, keys)

	_, err := ParseSemanticPolicyConfig(map[string]interface{}{"pack": ConfidentialityPack, "target": "input"})
	require.NoError(t, err)
}

func TestAttachmentBlocksKeepHeadAndTailOfLongDocuments(t *testing.T) {
	text := "HEAD " + strings.Repeat("x", maxAttachmentJudgeRunes*2) + " TAIL"

	many := make([]attachment, maxInspectedAttachments)
	for i := range many {
		many[i] = attachment{MessageIndex: 0, Name: fmt.Sprintf("doc%d.txt", i), Text: strings.Repeat("y", maxSemanticContentRunes)}
	}

	manyBlocks, _ := attachmentBlocks(many)
	require.Less(t, len([]rune(manyBlocks[0])), maxSemanticContentRunes, "many long attachments still fit the judge budget")
	require.True(t, strings.HasPrefix(manyBlocks[0], "[Attachment: doc0.txt]"))

	blocks, sampled := attachmentBlocks([]attachment{{MessageIndex: 0, Name: "long.txt", Text: text}})
	require.True(t, sampled)
	require.True(t, strings.HasPrefix(blocks[0], "[Attachment: long.txt]\nHEAD "))
	require.True(t, strings.HasSuffix(blocks[0], " TAIL"))
	require.Contains(t, blocks[0], "characters omitted")
	require.Less(t, len([]rune(blocks[0])), maxAttachmentJudgeRunes+200)
}

var errTestExtraction = errors.New("broken attachment")

func TestExtractionCacheReusesResultsAndEvicts(t *testing.T) {
	cache := newExtractionCache(2*extractionEntryOverhead + 10)
	first, second, third := extractionCacheKey("a"), extractionCacheKey("b"), extractionCacheKey("c")

	cache.put(first, "12345", nil)
	cache.put(second, "", errTestExtraction)

	text, err, ok := cache.get(first)
	require.True(t, ok)
	require.NoError(t, err)
	require.Equal(t, "12345", text)

	_, err, ok = cache.get(second)
	require.True(t, ok)
	require.ErrorIs(t, err, errTestExtraction)

	_, _, ok = cache.get(first)
	require.True(t, ok)

	cache.put(third, "12345", nil)

	_, _, ok = cache.get(second)
	require.False(t, ok, "least recently used entry is evicted")

	_, _, ok = cache.get(first)
	require.True(t, ok)
}
