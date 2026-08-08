package openai

import "slices"

const (
	// ResponsesAPIReferenceURL is the official OpenAI Responses API reference
	// used to build the checked-in contract inventory.
	ResponsesAPIReferenceURL = "https://developers.openai.com/api/reference/python/resources/responses"
	// ResponsesCreateReferenceURL is the official create-response reference
	// whose return schema is used for field and output-union reconciliation.
	ResponsesCreateReferenceURL = "https://developers.openai.com/api/reference/resources/responses/methods/create"
	// ResponsesStreamingEventsReferenceURL is the official OpenAI streaming
	// event reference used to build the checked-in contract inventory.
	ResponsesStreamingEventsReferenceURL = "https://developers.openai.com/api/reference/resources/responses/streaming-events"
	// ResponsesInputItemsReferenceURL is the official list-input-items
	// reference used to reconcile the pagination envelope.
	ResponsesInputItemsReferenceURL = "https://developers.openai.com/api/reference/resources/responses/subresources/input_items/methods/list"
	// ResponsesContractSnapshotDate records when the official inventory was
	// reconciled with the references above.
	ResponsesContractSnapshotDate = "2026-08-08"
)

// ResponsesOutputVariant identifies one member of the official Response
// output-item union and its JSON discriminator.
type ResponsesOutputVariant struct {
	Name    string
	SDKName string
	Type    string
}

// ResponsesFieldContract captures the documented response type expression and
// presence semantics for one top-level field.
type ResponsesFieldContract struct {
	Name           string
	TypeExpression string
	Optional       bool
	Nullable       bool
}

var responsesFieldContracts = []ResponsesFieldContract{
	{Name: "id", TypeExpression: "string"},
	{Name: "created_at", TypeExpression: "number"},
	{Name: "error", TypeExpression: "ResponseError or null", Nullable: true},
	{Name: "incomplete_details", TypeExpression: "object { reason }  or null", Nullable: true},
	{Name: "instructions", TypeExpression: "string or array of EasyInputMessage or object { content, role, status, type }  or ResponseOutputMessage or 29 more or null", Nullable: true},
	{Name: "metadata", TypeExpression: "Metadata or null", Nullable: true},
	{Name: "model", TypeExpression: "ResponsesModel"},
	{Name: "object", TypeExpression: "\"response\""},
	{Name: "output", TypeExpression: "array of ResponseOutputItem"},
	{Name: "parallel_tool_calls", TypeExpression: "boolean"},
	{Name: "temperature", TypeExpression: "number or null", Nullable: true},
	{Name: "tool_choice", TypeExpression: "ToolChoiceOptions or ToolChoiceAllowed or ToolChoiceTypes or 6 more"},
	{Name: "tools", TypeExpression: "array of object { name, parameters, strict, 5 more }  or object { type, vector_store_ids, filters, 2 more }  or object { type }  or 13 more"},
	{Name: "top_p", TypeExpression: "number or null", Nullable: true},
	{Name: "background", TypeExpression: "optional boolean or null", Optional: true, Nullable: true},
	{Name: "completed_at", TypeExpression: "optional number or null", Optional: true, Nullable: true},
	{Name: "conversation", TypeExpression: "optional object { id }  or null", Optional: true, Nullable: true},
	{Name: "max_output_tokens", TypeExpression: "optional number or null", Optional: true, Nullable: true},
	{Name: "max_tool_calls", TypeExpression: "optional number or null", Optional: true, Nullable: true},
	{Name: "moderation", TypeExpression: "optional object { input, output }  or null", Optional: true, Nullable: true},
	{Name: "output_text", TypeExpression: "optional string or null", Optional: true, Nullable: true},
	{Name: "previous_response_id", TypeExpression: "optional string or null", Optional: true, Nullable: true},
	{Name: "prompt", TypeExpression: "optional ResponsePrompt or null", Optional: true, Nullable: true},
	{Name: "prompt_cache_key", TypeExpression: "optional string or null", Optional: true, Nullable: true},
	{Name: "prompt_cache_options", TypeExpression: "optional object { mode, ttl }", Optional: true},
	{Name: "prompt_cache_retention", TypeExpression: "optional \"in_memory\" or \"24h\" or null", Optional: true, Nullable: true},
	{Name: "reasoning", TypeExpression: "optional Reasoning or null", Optional: true, Nullable: true},
	{Name: "safety_identifier", TypeExpression: "optional string or null", Optional: true, Nullable: true},
	{Name: "service_tier", TypeExpression: "optional \"auto\" or \"default\" or \"flex\" or 3 more or null", Optional: true, Nullable: true},
	{Name: "status", TypeExpression: "optional ResponseStatus", Optional: true},
	{Name: "text", TypeExpression: "optional ResponseTextConfig", Optional: true},
	{Name: "top_logprobs", TypeExpression: "optional number or null", Optional: true, Nullable: true},
	{Name: "truncation", TypeExpression: "optional \"auto\" or \"disabled\" or null", Optional: true, Nullable: true},
	{Name: "usage", TypeExpression: "optional ResponseUsage", Optional: true},
	{Name: "user", TypeExpression: "optional string", Optional: true},
}

var responsesTopLevelFields = []string{
	"id",
	"created_at",
	"error",
	"incomplete_details",
	"instructions",
	"metadata",
	"model",
	"object",
	"output",
	"parallel_tool_calls",
	"temperature",
	"tool_choice",
	"tools",
	"top_p",
	"background",
	"completed_at",
	"conversation",
	"max_output_tokens",
	"max_tool_calls",
	"moderation",
	"output_text",
	"previous_response_id",
	"prompt",
	"prompt_cache_key",
	"prompt_cache_options",
	"prompt_cache_retention",
	"reasoning",
	"safety_identifier",
	"service_tier",
	"status",
	"text",
	"top_logprobs",
	"truncation",
	"usage",
	"user",
}

var responsesOutputVariants = []ResponsesOutputVariant{
	{Name: "ResponseOutputMessage", SDKName: "ResponseOutputMessage", Type: "message"},
	{Name: "FileSearchCall", SDKName: "ResponseFileSearchToolCall", Type: "file_search_call"},
	{Name: "FunctionCall", SDKName: "ResponseFunctionToolCall", Type: "function_call"},
	{Name: "FunctionCallOutput", SDKName: "ResponseFunctionToolCallOutputItem", Type: "function_call_output"},
	{Name: "WebSearchCall", SDKName: "ResponseFunctionWebSearch", Type: "web_search_call"},
	{Name: "ComputerCall", SDKName: "ResponseComputerToolCall", Type: "computer_call"},
	{Name: "ComputerCallOutput", SDKName: "ResponseComputerToolCallOutputItem", Type: "computer_call_output"},
	{Name: "Reasoning", SDKName: "ResponseReasoningItem", Type: "reasoning"},
	{Name: "Program", SDKName: "Program", Type: "program"},
	{Name: "ProgramOutput", SDKName: "ProgramOutput", Type: "program_output"},
	{Name: "ToolSearchCall", SDKName: "ResponseToolSearchCall", Type: "tool_search_call"},
	{Name: "ToolSearchOutput", SDKName: "ResponseToolSearchOutputItem", Type: "tool_search_output"},
	{Name: "AdditionalTools", SDKName: "AdditionalTools", Type: "additional_tools"},
	{Name: "Compaction", SDKName: "ResponseCompactionItem", Type: "compaction"},
	{Name: "ImageGenerationCall", SDKName: "ImageGenerationCall", Type: "image_generation_call"},
	{Name: "CodeInterpreterCall", SDKName: "ResponseCodeInterpreterToolCall", Type: "code_interpreter_call"},
	{Name: "LocalShellCall", SDKName: "LocalShellCall", Type: "local_shell_call"},
	{Name: "LocalShellCallOutput", SDKName: "LocalShellCallOutput", Type: "local_shell_call_output"},
	{Name: "ShellCall", SDKName: "ResponseFunctionShellToolCall", Type: "shell_call"},
	{Name: "ShellCallOutput", SDKName: "ResponseFunctionShellToolCallOutput", Type: "shell_call_output"},
	{Name: "ApplyPatchCall", SDKName: "ResponseApplyPatchToolCall", Type: "apply_patch_call"},
	{Name: "ApplyPatchCallOutput", SDKName: "ResponseApplyPatchToolCallOutput", Type: "apply_patch_call_output"},
	{Name: "McpCall", SDKName: "McpCall", Type: "mcp_call"},
	{Name: "McpListTools", SDKName: "McpListTools", Type: "mcp_list_tools"},
	{Name: "McpApprovalRequest", SDKName: "McpApprovalRequest", Type: "mcp_approval_request"},
	{Name: "McpApprovalResponse", SDKName: "McpApprovalResponse", Type: "mcp_approval_response"},
	{Name: "CustomToolCall", SDKName: "ResponseCustomToolCall", Type: "custom_tool_call"},
	{Name: "CustomToolCallOutput", SDKName: "ResponseCustomToolCallOutputItem", Type: "custom_tool_call_output"},
}

var responsesItemVariants = []ResponsesOutputVariant{
	{Name: "ResponseInputMessageItem", SDKName: "ResponseInputMessageItem", Type: "message"},
	{Name: "ResponseOutputMessage", SDKName: "ResponseOutputMessage", Type: "message"},
	{Name: "FileSearchCall", SDKName: "ResponseFileSearchToolCall", Type: "file_search_call"},
	{Name: "ComputerCall", SDKName: "ResponseComputerToolCall", Type: "computer_call"},
	{Name: "ComputerCallOutput", SDKName: "ResponseComputerToolCallOutputItem", Type: "computer_call_output"},
	{Name: "WebSearchCall", SDKName: "ResponseFunctionWebSearch", Type: "web_search_call"},
	{Name: "FunctionCall", SDKName: "ResponseFunctionToolCallItem", Type: "function_call"},
	{Name: "FunctionCallOutput", SDKName: "ResponseFunctionToolCallOutputItem", Type: "function_call_output"},
	{Name: "ToolSearchCall", SDKName: "ResponseToolSearchCall", Type: "tool_search_call"},
	{Name: "ToolSearchOutput", SDKName: "ResponseToolSearchOutputItem", Type: "tool_search_output"},
	{Name: "AdditionalTools", SDKName: "AdditionalTools", Type: "additional_tools"},
	{Name: "Reasoning", SDKName: "ResponseReasoningItem", Type: "reasoning"},
	{Name: "Program", SDKName: "Program", Type: "program"},
	{Name: "ProgramOutput", SDKName: "ProgramOutput", Type: "program_output"},
	{Name: "Compaction", SDKName: "ResponseCompactionItem", Type: "compaction"},
	{Name: "ImageGenerationCall", SDKName: "ImageGenerationCall", Type: "image_generation_call"},
	{Name: "CodeInterpreterCall", SDKName: "ResponseCodeInterpreterToolCall", Type: "code_interpreter_call"},
	{Name: "LocalShellCall", SDKName: "LocalShellCall", Type: "local_shell_call"},
	{Name: "LocalShellCallOutput", SDKName: "LocalShellCallOutput", Type: "local_shell_call_output"},
	{Name: "ShellCall", SDKName: "ResponseFunctionShellToolCall", Type: "shell_call"},
	{Name: "ShellCallOutput", SDKName: "ResponseFunctionShellToolCallOutput", Type: "shell_call_output"},
	{Name: "ApplyPatchCall", SDKName: "ResponseApplyPatchToolCall", Type: "apply_patch_call"},
	{Name: "ApplyPatchCallOutput", SDKName: "ResponseApplyPatchToolCallOutput", Type: "apply_patch_call_output"},
	{Name: "McpListTools", SDKName: "McpListTools", Type: "mcp_list_tools"},
	{Name: "McpApprovalRequest", SDKName: "McpApprovalRequest", Type: "mcp_approval_request"},
	{Name: "McpApprovalResponse", SDKName: "McpApprovalResponse", Type: "mcp_approval_response"},
	{Name: "McpCall", SDKName: "McpCall", Type: "mcp_call"},
	{Name: "CustomToolCall", SDKName: "ResponseCustomToolCallItem", Type: "custom_tool_call"},
	{Name: "CustomToolCallOutput", SDKName: "ResponseCustomToolCallOutputItem", Type: "custom_tool_call_output"},
}

var responsesCompactedFields = []string{
	"id",
	"created_at",
	"object",
	"output",
	"usage",
}

var responsesInputTokenCountFields = []string{
	"input_tokens",
	"object",
}

var responsesInputItemsPageFields = []string{
	"object",
	"data",
	"first_id",
	"last_id",
	"has_more",
}

var responsesOperations = []string{
	"POST /responses",
	"POST /responses/compact",
	"POST /responses/input_tokens",
	"GET /responses/{response_id}",
	"DELETE /responses/{response_id}",
	"POST /responses/{response_id}/cancel",
	"GET /responses/{response_id}/input_items",
}

var responsesStreamingEventTypes = []string{
	"response.created",
	"response.in_progress",
	"response.completed",
	"response.failed",
	"response.incomplete",
	"response.output_item.added",
	"response.output_item.done",
	"response.content_part.added",
	"response.content_part.done",
	"response.output_text.delta",
	"response.output_text.done",
	"response.refusal.delta",
	"response.refusal.done",
	"response.function_call_arguments.delta",
	"response.function_call_arguments.done",
	"response.file_search_call.in_progress",
	"response.file_search_call.searching",
	"response.file_search_call.completed",
	"response.web_search_call.in_progress",
	"response.web_search_call.searching",
	"response.web_search_call.completed",
	"response.reasoning_summary_part.added",
	"response.reasoning_summary_part.done",
	"response.reasoning_summary_text.delta",
	"response.reasoning_summary_text.done",
	"response.reasoning_text.delta",
	"response.reasoning_text.done",
	"response.image_generation_call.completed",
	"response.image_generation_call.generating",
	"response.image_generation_call.in_progress",
	"response.image_generation_call.partial_image",
	"response.mcp_call_arguments.delta",
	"response.mcp_call_arguments.done",
	"response.mcp_call.completed",
	"response.mcp_call.failed",
	"response.mcp_call.in_progress",
	"response.mcp_list_tools.completed",
	"response.mcp_list_tools.failed",
	"response.mcp_list_tools.in_progress",
	"response.code_interpreter_call.in_progress",
	"response.code_interpreter_call.interpreting",
	"response.code_interpreter_call.completed",
	"response.code_interpreter_call_code.delta",
	"response.code_interpreter_call_code.done",
	"response.output_text.annotation.added",
	"response.queued",
	"response.custom_tool_call_input.delta",
	"response.custom_tool_call_input.done",
	"error",
	"response.audio.delta",
	"response.audio.done",
	"response.audio.transcript.delta",
	"response.audio.transcript.done",
}

var responsesStatuses = []string{
	"cancelled", //nolint:misspell // OpenAI's ResponseStatus contract uses this spelling.
	"completed",
	"failed",
	"in_progress",
	"incomplete",
	"queued",
}

var responsesTerminalStreamingEventTypes = []string{
	"response.completed",
	"response.failed",
	"response.incomplete",
	"error",
}

// ResponsesTopLevelFields returns every documented top-level Response field
// in the pinned official inventory.
func ResponsesTopLevelFields() []string {
	return slices.Clone(responsesTopLevelFields)
}

// ResponsesFieldContracts returns every documented top-level Response field
// with its type expression, optionality, and nullability.
func ResponsesFieldContracts() []ResponsesFieldContract {
	return slices.Clone(responsesFieldContracts)
}

// ResponsesOutputVariants returns every documented member of the Response
// output-item union in the pinned official inventory.
func ResponsesOutputVariants() []ResponsesOutputVariant {
	return slices.Clone(responsesOutputVariants)
}

// ResponsesItemVariants returns every documented member of the ResponseItem
// union returned by the input-items operation.
func ResponsesItemVariants() []ResponsesOutputVariant {
	return slices.Clone(responsesItemVariants)
}

// ResponsesCompactedFields returns every top-level CompactedResponse field.
func ResponsesCompactedFields() []string {
	return slices.Clone(responsesCompactedFields)
}

// ResponsesInputTokenCountFields returns every top-level field returned by
// the input-token-count operation.
func ResponsesInputTokenCountFields() []string {
	return slices.Clone(responsesInputTokenCountFields)
}

// ResponsesInputItemsPageFields returns every pagination-envelope field
// returned by the input-items operation.
func ResponsesInputItemsPageFields() []string {
	return slices.Clone(responsesInputItemsPageFields)
}

// ResponsesOperations returns every method and path documented by the
// official Responses SDK resource reference.
func ResponsesOperations() []string {
	return slices.Clone(responsesOperations)
}

// ResponsesStreamingEventTypes returns every documented Responses SSE event
// type in the pinned official inventory.
func ResponsesStreamingEventTypes() []string {
	return slices.Clone(responsesStreamingEventTypes)
}

// ResponsesStatuses returns every documented top-level Response status.
func ResponsesStatuses() []string {
	return slices.Clone(responsesStatuses)
}

// ResponsesTerminalStreamingEventTypes returns the lifecycle events that
// terminate a Responses SSE stream.
func ResponsesTerminalStreamingEventTypes() []string {
	return slices.Clone(responsesTerminalStreamingEventTypes)
}
