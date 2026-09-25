package guardrails

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/docextract"
)

type Hook string

const (
	HookLLMInput  Hook = "llm_input"
	HookLLMOutput Hook = "llm_output"
)

type Operation string

const (
	OperationValidate Operation = "validate"
	OperationMutate   Operation = "mutate"
)

type Enforcement string

const (
	EnforcementAudit              Enforcement = "audit"
	EnforcementEnforce            Enforcement = "enforce"
	EnforcementEnforceIgnoreError Enforcement = "enforce_ignore_error"
)

type Scope string

const (
	ScopeAll  Scope = "all"
	ScopeLast Scope = "last"
)

type RequestSelection struct {
	LLMInput  []string
	LLMOutput []string
	Scope     Scope
}

func (s RequestSelection) Empty() bool {
	return len(s.LLMInput) == 0 && len(s.LLMOutput) == 0
}

type Check struct {
	Selector    string      `json:"selector"`
	Status      string      `json:"status"`
	Operation   Operation   `json:"operation"`
	Enforcement Enforcement `json:"enforcement"`
	Reason      string      `json:"reason,omitempty"`
	LatencyMS   int64       `json:"latency_ms"`
	Mutated     bool        `json:"mutated,omitempty"`
	Error       string      `json:"error,omitempty"`
	// Semantic carries rule probabilities for traces. It is never returned to API clients.
	Semantic *SemanticCheckDetail `json:"-"`
}

type Result struct {
	Request  contractsopenai.ChatCompletionRequest
	Response *contractsopenai.ChatCompletionResponse
	Checks   map[string][]Check
	Blocked  bool
	Reason   string
}

type ViolationError struct {
	Message string
	Checks  map[string][]Check
}

func (e *ViolationError) Error() string {
	if strings.TrimSpace(e.Message) == "" {
		return "guardrail checks failed"
	}

	return e.Message
}

type EvaluationInput struct {
	Hook     Hook
	Selector string
	Texts    []string
}

type EvaluationResult struct {
	Blocked bool
	Reason  string
	Texts   []string
	Mutated bool
}

type Guardrail interface {
	Name() string
	Operation() Operation
	Enforcement() Enforcement
	Evaluate(input EvaluationInput) (EvaluationResult, error)
}

func NewFromTenantGuardrail(item platform.TenantGuardrail) Guardrail {
	operation := Operation(item.Operation)
	if operation == "" {
		operation = OperationValidate
	}

	enforcement := Enforcement(item.Enforcement)
	if enforcement == "" {
		enforcement = EnforcementEnforce
	}

	replacement := stringFromConfig(item.Config, "redaction_text", defaultRedaction)
	switch normalizeSelectorKey(item.Type) {
	case "keyword_blocklist":
		return &keywordBlocklistGuardrail{
			name:        item.Slug,
			terms:       stringSliceFromConfig(item.Config, "terms"),
			enforcement: enforcement,
		}
	case "regex":
		return newPatternGuardrailFromPatterns(item.Slug, operation, enforcement, "Sensitive pattern matched", replacement, stringSliceFromConfig(item.Config, "patterns"))
	case "regex_replace":
		return newRegexReplaceGuardrail(item.Slug, enforcement, stringFromConfig(item.Config, "pattern", ""), stringFromConfig(item.Config, "replacement", replacement), boolFromConfig(item.Config, "case_sensitive", true))
	case "secrets_detection":
		return newSecretsDetectionGuardrail(item.Slug, operation, enforcement, replacement)
	case "basic_pii":
		return newBasicPIIGuardrail(item.Slug, operation, enforcement, replacement)
	case "sentence_count":
		return newCountGuardrail(item.Slug, "sentence", enforcement, intFromConfig(item.Config, "min", -1), intFromConfig(item.Config, "max", -1))
	case "word_count":
		return newCountGuardrail(item.Slug, "word", enforcement, intFromConfig(item.Config, "min", -1), intFromConfig(item.Config, "max", -1))
	case "character_count":
		return newCountGuardrail(item.Slug, "character", enforcement, intFromConfig(item.Config, "min", -1), intFromConfig(item.Config, "max", -1))
	case "json_schema":
		return newJSONSchemaGuardrail(item.Slug, enforcement, mapFromConfig(item.Config, "schema"))
	case "json_keys":
		return newJSONKeysGuardrail(item.Slug, enforcement, stringSliceFromConfig(item.Config, "keys"), stringFromConfig(item.Config, "mode", "all"))
	case "contains_code":
		return newContainsCodeGuardrail(item.Slug, enforcement, stringSliceFromConfig(item.Config, "languages"), containsCodeMustContainFromConfig(item.Config))
	case "not_null":
		return newNotNullGuardrail(item.Slug, enforcement)
	case "jwt_validation":
		return newJWTValidationGuardrail(item.Slug, enforcement, item.Config)
	case "inline_image_urls":
		return newInlineImageURLsGuardrail(item.Slug, enforcement, item.Config)
	case SemanticPolicyType:
		return newSemanticPolicyGuardrail(item.Slug, enforcement, item.Config)
	default:
		return nil
	}
}

type Engine struct {
	registry map[string]Guardrail
	judge    SemanticJudge
}

// EngineOption configures optional engine collaborators.
type EngineOption func(*Engine)

// WithSemanticJudge enables semantic_policy guardrails. Without a judge they fail as execution errors.
func WithSemanticJudge(judge SemanticJudge) EngineOption {
	return func(e *Engine) {
		e.judge = judge
	}
}

func NewDefaultEngine(options ...EngineOption) *Engine {
	e := &Engine{registry: map[string]Guardrail{}}
	for _, option := range options {
		option(e)
	}

	for _, g := range []Guardrail{
		NewKeywordBlocklistGuardrail(),
		NewRegexGuardrail(),
		NewSecretsDetectionGuardrail(),
		NewBasicPIIGuardrail(),
	} {
		e.Register(g)
	}

	e.registry["pii_redaction"] = e.registry["basic_pii"]
	e.registry["pii_detection"] = e.registry["basic_pii"]

	return e
}

func (e *Engine) Register(g Guardrail) {
	if e == nil || g == nil {
		return
	}

	e.registry[normalizeSelectorKey(g.Name())] = g
}

func ParseSelection(rawGuardrails string, rawScope string) (RequestSelection, error) {
	var sel RequestSelection

	scope := strings.ToLower(strings.TrimSpace(rawScope))
	switch Scope(scope) {
	case "", ScopeAll:
		sel.Scope = ScopeAll
	case ScopeLast:
		sel.Scope = ScopeLast
	default:
		return sel, fmt.Errorf("guardrails scope must be all or last")
	}

	rawGuardrails = strings.TrimSpace(rawGuardrails)
	if rawGuardrails == "" {
		return sel, nil
	}

	if !strings.HasPrefix(rawGuardrails, "{") && !strings.HasPrefix(rawGuardrails, "[") {
		sel.LLMInput = parseSelectorList(rawGuardrails)
		return sel, nil
	}

	var payload struct {
		LLMInput          []string `json:"llm_input_guardrails"`
		LLMOutput         []string `json:"llm_output_guardrails"`
		InputAlias        []string `json:"input_guardrails"`
		OutputAlias       []string `json:"output_guardrails"`
		MCPToolPreInvoke  []string `json:"mcp_tool_pre_invoke_guardrails"`
		MCPToolPostInvoke []string `json:"mcp_tool_post_invoke_guardrails"`
	}
	if err := json.Unmarshal([]byte(rawGuardrails), &payload); err != nil {
		return sel, fmt.Errorf("invalid guardrails header: %w", err)
	}

	sel.LLMInput = uniqueSelectors(append(payload.LLMInput, payload.InputAlias...))
	sel.LLMOutput = uniqueSelectors(append(payload.LLMOutput, payload.OutputAlias...))

	return sel, nil
}

func (e *Engine) ApplyInput(req contractsopenai.ChatCompletionRequest, sel RequestSelection) (Result, error) {
	return e.ApplyInputWithGuardrails(context.Background(), req, sel, nil)
}

// ApplyInputWithGuardrails runs deterministic guardrails in selection order, then every semantic guardrail in one
// judge call on the already-mutated request, so redacted values never reach the judge. Inline document attachments are
// extracted to text first and inspected by the content guardrails alongside the typed text.
func (e *Engine) ApplyInputWithGuardrails(ctx context.Context, req contractsopenai.ChatCompletionRequest, sel RequestSelection, overrides map[string]Guardrail) (Result, error) {
	result := Result{Request: req, Checks: map[string][]Check{}}
	if e == nil || len(sel.LLMInput) == 0 {
		return result, nil
	}

	texts, refs := inputTexts(req, sel.Scope)

	var attachments []attachment

	if enforcement, ok := e.attachmentEnforcement(sel.LLMInput, overrides); ok {
		var failures []attachmentFailure

		attachments, failures = inputAttachments(req, sel.Scope, docextract.DefaultLimits())
		if len(failures) > 0 {
			check := attachmentFailureCheck(failures, enforcement)
			result.Checks["llm_input_guardrails"] = append(result.Checks["llm_input_guardrails"], check)

			if enforcement == EnforcementEnforce {
				result.Blocked = true
				result.Reason = check.Reason

				return result, &ViolationError{Message: "Request blocked by GodarAI guardrail", Checks: result.Checks}
			}
		}
	}

	fileTexts := attachmentTexts(attachments)

	var semantic []semanticItem

	outTexts := append([]string(nil), texts...)

	for _, selector := range sel.LLMInput {
		if sg := e.semanticGuardrail(selector, overrides); sg != nil {
			semantic = append(semantic, semanticItem{selector: selector, prefix: fmt.Sprintf("g%d__", len(semantic)), g: sg})

			continue
		}

		g, check, eval, redactedFiles, err := e.evaluateInput(selector, outTexts, fileTexts, overrides)
		result.Checks["llm_input_guardrails"] = append(result.Checks["llm_input_guardrails"], check)

		if err != nil {
			if shouldBlockOnError(g) {
				result.Blocked = true
				result.Reason = check.Reason

				return result, &ViolationError{Message: "Request blocked by GodarAI guardrail", Checks: result.Checks}
			}

			continue
		}

		if shouldBlockOnViolation(g, eval) {
			result.Blocked = true
			result.Reason = check.Reason

			return result, &ViolationError{Message: "Request blocked by GodarAI guardrail", Checks: result.Checks}
		}

		if eval.Mutated && g.Operation() == OperationMutate {
			outTexts = eval.Texts
			applyInputTexts(&result.Request, refs, outTexts)

			// The forwarded file cannot change, but later guardrails and the judge see the redacted copy.
			fileTexts = redactedFiles
		}
	}

	if len(semantic) == 0 {
		return result, nil
	}

	for i := range attachments {
		attachments[i].Text = fileTexts[i]
	}

	checks, blocking := e.applySemantic(ctx, HookLLMInput, semantic, semanticInputState(result.Request, sel.Scope, attachmentBlocks(attachments)))
	result.Checks["llm_input_guardrails"] = append(result.Checks["llm_input_guardrails"], checks...)

	if blocking >= 0 {
		result.Blocked = true
		result.Reason = checks[blocking].Reason

		return result, &ViolationError{Message: "Request blocked by GodarAI guardrail", Checks: result.Checks}
	}

	return result, nil
}

// evaluateInput runs one deterministic guardrail on the typed texts and, for content guardrails, on the attachment
// texts too. The returned evaluation only carries the typed texts, and the attachment texts are returned separately with
// any redaction applied. A redaction that would touch an attachment cannot be applied to the forwarded file, so it is
// reported as a block.
func (e *Engine) evaluateInput(selector string, texts []string, fileTexts []string, overrides map[string]Guardrail) (Guardrail, Check, EvaluationResult, []string, error) {
	g := e.resolve(selector, overrides)
	if len(fileTexts) == 0 || g == nil || !inspectsAttachments(g) {
		resolved, check, eval, err := e.evaluate(selector, HookLLMInput, texts, overrides)

		return resolved, check, eval, fileTexts, err
	}

	combined := make([]string, 0, len(texts)+len(fileTexts))
	combined = append(combined, texts...)
	combined = append(combined, fileTexts...)

	g, check, eval, err := e.evaluate(selector, HookLLMInput, combined, overrides)
	if err != nil {
		return g, check, eval, fileTexts, err
	}

	redactedFiles := fileTexts

	if len(eval.Texts) == len(combined) {
		redactedFiles = eval.Texts[len(texts):]

		if eval.Mutated && attachmentsChanged(fileTexts, eval.Texts[len(texts):]) {
			eval.Blocked = true
			eval.Reason = "Attachment contains content that must be redacted; attachments cannot be modified"
			check.Status = "blocked"
			check.Reason = eval.Reason

			if g.Enforcement() == EnforcementAudit {
				check.Status = "flagged"
			}
		}

		eval.Texts = eval.Texts[:len(texts)]
	}

	return g, check, eval, redactedFiles, nil
}

func (e *Engine) ApplyOutput(resp contractsopenai.ChatCompletionResponse, sel RequestSelection) (Result, error) {
	return e.ApplyOutputWithGuardrails(context.Background(), resp, sel, nil)
}

// ApplyOutputWithGuardrails mirrors ApplyInputWithGuardrails for the model's reply.
func (e *Engine) ApplyOutputWithGuardrails(ctx context.Context, resp contractsopenai.ChatCompletionResponse, sel RequestSelection, overrides map[string]Guardrail) (Result, error) {
	result := Result{Response: &resp, Checks: map[string][]Check{}}
	if e == nil || len(sel.LLMOutput) == 0 {
		return result, nil
	}

	texts, refs := outputTexts(resp)

	var semantic []semanticItem

	outTexts := append([]string(nil), texts...)

	for _, selector := range sel.LLMOutput {
		if sg := e.semanticGuardrail(selector, overrides); sg != nil {
			semantic = append(semantic, semanticItem{selector: selector, prefix: fmt.Sprintf("g%d__", len(semantic)), g: sg})

			continue
		}

		g, check, eval, err := e.evaluate(selector, HookLLMOutput, outTexts, overrides)
		if err != nil {
			result.Checks["llm_output_guardrails"] = append(result.Checks["llm_output_guardrails"], check)
			if shouldBlockOnError(g) {
				result.Blocked = true
				result.Reason = check.Reason

				return result, &ViolationError{Message: "Response blocked by GodarAI guardrail", Checks: result.Checks}
			}

			continue
		}

		result.Checks["llm_output_guardrails"] = append(result.Checks["llm_output_guardrails"], check)
		if shouldBlockOnViolation(g, eval) {
			result.Blocked = true
			result.Reason = check.Reason

			return result, &ViolationError{Message: "Response blocked by GodarAI guardrail", Checks: result.Checks}
		}

		if eval.Mutated && g.Operation() == OperationMutate {
			outTexts = eval.Texts
			applyOutputTexts(result.Response, refs, outTexts)
		}
	}

	if len(semantic) == 0 {
		return result, nil
	}

	checks, blocking := e.applySemantic(ctx, HookLLMOutput, semantic, semanticOutputState(result.Response))
	result.Checks["llm_output_guardrails"] = append(result.Checks["llm_output_guardrails"], checks...)

	if blocking >= 0 {
		result.Blocked = true
		result.Reason = checks[blocking].Reason

		return result, &ViolationError{Message: "Response blocked by GodarAI guardrail", Checks: result.Checks}
	}

	return result, nil
}

func (e *Engine) evaluate(selector string, hook Hook, texts []string, overrides map[string]Guardrail) (Guardrail, Check, EvaluationResult, error) {
	start := time.Now()

	g := e.resolve(selector, overrides)
	if g == nil {
		check := Check{Selector: selector, Status: "error", Operation: OperationValidate, Enforcement: EnforcementEnforce, Reason: "Unknown guardrail selector", Error: "unknown_selector"}
		check.LatencyMS = time.Since(start).Milliseconds()

		return nil, check, EvaluationResult{}, fmt.Errorf("unknown guardrail selector %q", selector)
	}

	eval, err := g.Evaluate(EvaluationInput{Hook: hook, Selector: selector, Texts: texts})

	check := Check{Selector: selector, Operation: g.Operation(), Enforcement: g.Enforcement(), LatencyMS: time.Since(start).Milliseconds(), Mutated: eval.Mutated}
	if err != nil {
		check.Status = "error"
		check.Reason = "Guardrail execution failed"
		check.Error = err.Error()

		return g, check, eval, err
	}

	if eval.Blocked {
		check.Status = "blocked"
		check.Reason = eval.Reason
	} else if eval.Mutated {
		check.Status = "mutated"
		check.Reason = eval.Reason
	} else {
		check.Status = "passed"
	}

	if g.Enforcement() == EnforcementAudit && eval.Blocked {
		check.Status = "flagged"
	}

	return g, check, eval, nil
}

func (e *Engine) resolve(selector string, overrides map[string]Guardrail) Guardrail {
	if e == nil {
		return nil
	}

	selector = strings.TrimSpace(selector)
	if selector == "" {
		return nil
	}

	parts := strings.Split(selector, "/")

	candidate := parts[len(parts)-1]
	if g := overrides[normalizeSelectorKey(candidate)]; g != nil {
		return g
	}

	if g := overrides[normalizeSelectorKey(selector)]; g != nil {
		return g
	}

	if g := e.registry[normalizeSelectorKey(candidate)]; g != nil {
		return g
	}

	return e.registry[normalizeSelectorKey(selector)]
}

func stringSliceFromConfig(config map[string]interface{}, key string) []string {
	raw := config[key]
	switch values := raw.(type) {
	case []string:
		return uniqueSelectors(values)
	case []interface{}:
		out := make([]string, 0, len(values))
		for _, value := range values {
			if text, ok := value.(string); ok {
				out = append(out, text)
			}
		}

		return uniqueSelectors(out)
	default:
		return nil
	}
}

func stringFromConfig(config map[string]interface{}, key string, fallback string) string {
	if value, ok := config[key].(string); ok && strings.TrimSpace(value) != "" {
		return value
	}

	return fallback
}

func boolFromConfig(config map[string]interface{}, key string, fallback bool) bool {
	if value, ok := config[key].(bool); ok {
		return value
	}

	return fallback
}

func containsCodeMustContainFromConfig(config map[string]interface{}) bool {
	if value, ok := config["must_contain"].(bool); ok {
		return value
	}

	return !strings.EqualFold(strings.TrimSpace(stringFromConfig(config, "mode", "must_contain")), "must_not_contain")
}

func intFromConfig(config map[string]interface{}, key string, fallback int) int {
	switch value := config[key].(type) {
	case int:
		return value
	case int64:
		return int(value)
	case float64:
		return int(value)
	case json.Number:
		parsed, err := value.Int64()
		if err == nil {
			return int(parsed)
		}
	}

	return fallback
}

func mapFromConfig(config map[string]interface{}, key string) map[string]interface{} {
	if value, ok := config[key].(map[string]interface{}); ok {
		return value
	}

	return nil
}

func regexpFromConfig(pattern string, caseSensitive bool) (*regexp.Regexp, error) {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return nil, fmt.Errorf("regex pattern is required")
	}

	if !caseSensitive && !strings.HasPrefix(pattern, "(?i)") {
		pattern = "(?i)" + pattern
	}

	return regexp.Compile(pattern)
}

func shouldBlockOnViolation(g Guardrail, eval EvaluationResult) bool {
	return g != nil && eval.Blocked && g.Enforcement() != EnforcementAudit
}

func shouldBlockOnError(g Guardrail) bool {
	return g == nil || g.Enforcement() == EnforcementEnforce
}

func normalizeSelectorKey(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "-", "_")

	return value
}

func uniqueSelectors(in []string) []string {
	out := make([]string, 0, len(in))
	seen := map[string]bool{}

	for _, value := range in {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}

		seen[value] = true
		out = append(out, value)
	}

	return out
}

func parseSelectorList(raw string) []string {
	raw = strings.ReplaceAll(raw, "\n", ",")
	raw = strings.ReplaceAll(raw, ";", ",")

	return uniqueSelectors(strings.Split(raw, ","))
}
