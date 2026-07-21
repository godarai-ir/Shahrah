import { extractPromptVariableNames, renderPromptText } from "./prompt-management-utils";
import type { PromptManifest, PromptVersion } from "./prompt-management-shared";

export type PlaygroundRole = "system" | "user" | "assistant" | "tool";

export type PlaygroundMessage = {
  id: string;
  role: PlaygroundRole;
  content: string;
  toolCallId?: string;
};

export type PlaygroundTool = {
  id: string;
  name: string;
  description: string;
  schema: string;
};

export type PlaygroundMetadataRow = { id: string; key: string; value: string };

export type PlaygroundDomainFailure = {
  ok: false;
  errorKey: string;
  errorValues?: Record<string, string | number>;
};

export type PlaygroundConfig = {
  stream: boolean;
  temperature: string;
  topP: string;
  maxTokens: string;
  stop: string;
  reasoningEffort: string;
  parallelToolCalls: boolean;
  toolChoice: string;
  cacheEnabled: boolean;
  cacheType: "simple" | "smart";
  cacheTTL: string;
  cacheNamespace: string;
  cacheSimilarity: string;
  inputGuardrails: string[];
  outputGuardrails: string[];
  guardrailScope: "all" | "last";
  structuredEnabled: boolean;
  structuredFormat: "json_object" | "json_schema";
  structuredName: string;
  structuredStrict: boolean;
  structuredSchema: string;
  metadata: PlaygroundMetadataRow[];
};

export type PlaygroundColumnDraft = {
  id: string;
  label: string;
  selectedModel: string;
  selectedProvider: string;
  messages: PlaygroundMessage[];
  tools: PlaygroundTool[];
  config: PlaygroundConfig;
  variableValues: Record<string, string>;
  editorMode: "visual" | "json";
  jsonDraft: string;
  jsonError: string;
  sourcePromptID?: string;
  sourceVersionID?: string;
  sourceFQN?: string;
  sourcePromptName?: string;
  sourcePromptDescription?: string;
  dirty: boolean;
};

export type PlaygroundTelemetry = {
  requestId?: string;
  traceId?: string;
  requestedModel?: string;
  effectiveModel?: string;
  provider?: string;
  durationMs?: number;
  ttftMs?: number;
  inputTokens?: number;
  outputTokens?: number;
  totalTokens?: number;
  costUSD?: number;
  costIRR?: number;
  cacheStatus?: string;
  cacheSimilarity?: number;
  retryCount?: number;
  fallbackCount?: number;
  guardrailStatus?: string;
};

export type PlaygroundRunResult = {
  status: "idle" | "running" | "success" | "stopped" | "error";
  responseText: string;
  rawResponse: unknown;
  toolCalls: unknown[];
  finishReason?: string;
  error?: string;
  responseIsPrevious?: boolean;
  telemetry?: PlaygroundTelemetry;
  startedAt?: number;
};

const DEFAULT_SCHEMA = JSON.stringify({
  type: "object",
  properties: { result: { type: "string" } },
  required: ["result"],
}, null, 2);

const RAW_PARAMETER_KEYS = new Set([
  "stream", "temperature", "top_p", "max_tokens", "stop", "reasoning_effort",
  "parallel_tool_calls", "tool_choice", "response_format",
]);

export function createPlaygroundColumn(label: string, defaultSystemPrompt: string): PlaygroundColumnDraft {
  const messages: PlaygroundMessage[] = [
    createPlaygroundMessage("system", defaultSystemPrompt),
    createPlaygroundMessage("user", ""),
  ];
  const column: PlaygroundColumnDraft = {
    id: createID("column"),
    label,
    selectedModel: "",
    selectedProvider: "",
    messages,
    tools: [],
    config: createDefaultConfig(),
    variableValues: {},
    editorMode: "visual",
    jsonDraft: "",
    jsonError: "",
    dirty: false,
  };
  column.jsonDraft = serializeColumnJSON(column);
  return column;
}

export function createDefaultConfig(): PlaygroundConfig {
  return {
    stream: true,
    temperature: "0.2",
    topP: "",
    maxTokens: "",
    stop: "",
    reasoningEffort: "",
    parallelToolCalls: true,
    toolChoice: "auto",
    cacheEnabled: false,
    cacheType: "simple",
    cacheTTL: "600",
    cacheNamespace: "",
    cacheSimilarity: "0.9",
    inputGuardrails: [],
    outputGuardrails: [],
    guardrailScope: "last",
    structuredEnabled: false,
    structuredFormat: "json_schema",
    structuredName: "playground_output",
    structuredStrict: true,
    structuredSchema: DEFAULT_SCHEMA,
    metadata: [],
  };
}

export function createPlaygroundMessage(role: PlaygroundRole = "user", content = ""): PlaygroundMessage {
  return { id: createID("message"), role, content };
}

export function createPlaygroundTool(): PlaygroundTool {
  return {
    id: createID("tool"),
    name: "",
    description: "",
    schema: JSON.stringify({ type: "object", properties: {} }, null, 2),
  };
}

export function createPlaygroundMetadataRow(): PlaygroundMetadataRow {
  return { id: createID("metadata"), key: "", value: "" };
}

export function clonePlaygroundColumn(column: PlaygroundColumnDraft, label: string): PlaygroundColumnDraft {
  const clone = structuredClone(column);
  clone.id = createID("column");
  clone.label = label;
  clone.messages = clone.messages.map((message) => ({ ...message, id: createID("message") }));
  clone.tools = clone.tools.map((tool) => ({ ...tool, id: createID("tool") }));
  clone.config.metadata = clone.config.metadata.map((row) => ({ ...row, id: createID("metadata") }));
  clone.sourcePromptID = undefined;
  clone.sourceVersionID = undefined;
  clone.sourceFQN = undefined;
  clone.sourcePromptName = undefined;
  clone.sourcePromptDescription = undefined;
  clone.dirty = true;
  clone.jsonDraft = serializeColumnJSON(clone);
  clone.jsonError = "";
  return clone;
}

export function getColumnVariables(column: PlaygroundColumnDraft): string[] {
  return extractPromptVariableNames(column.messages.map((message) => message.content));
}

export function serializeColumnJSON(column: PlaygroundColumnDraft): string {
  const toolsResult = buildTools(column.tools);
  return JSON.stringify({
    messages: column.messages.map((message) => ({
      role: message.role,
      content: message.content,
      ...(message.role === "tool" && message.toolCallId ? { tool_call_id: message.toolCallId } : {}),
    })),
    parameters: buildParameters(column, true),
    tools: toolsResult.ok ? toolsResult.tools : column.tools.map((tool) => ({
      type: "function",
      function: { name: tool.name, description: tool.description, parameters: safeParseObject(tool.schema) },
    })),
  }, null, 2);
}

export function applyColumnJSON(column: PlaygroundColumnDraft): { ok: true; column: PlaygroundColumnDraft } | PlaygroundDomainFailure {
  let parsed: unknown;
  try {
    parsed = JSON.parse(column.jsonDraft);
  } catch (error) {
    return failure("validation.jsonInvalid", { details: error instanceof SyntaxError ? error.message : "Invalid JSON" });
  }
  if (!isRecord(parsed)) return failure("validation.jsonRootObject");
  const unsupported = Object.keys(parsed).filter((key) => !["messages", "parameters", "tools"].includes(key));
  if (unsupported.length) return failure("validation.jsonUnsupportedFields", { fields: unsupported.join(", ") });
  if (!Array.isArray(parsed.messages)) return failure("validation.jsonMessagesArray");

  const messages: PlaygroundMessage[] = [];
  for (let index = 0; index < parsed.messages.length; index += 1) {
    const message = parsed.messages[index];
    const displayIndex = index + 1;
    if (!isRecord(message)) return failure("validation.jsonMessageObject", { index: displayIndex, path: `messages[${index}]` });
    const unsupportedMessageFields = unsupportedKeys(message, ["role", "content", "tool_call_id"]);
    if (unsupportedMessageFields.length) return failure("validation.jsonMessageUnsupportedFields", { index: displayIndex, path: `messages[${index}]`, fields: unsupportedMessageFields.join(", ") });
    if (!isRole(message.role)) return failure("validation.jsonMessageRole", { index: displayIndex, path: `messages[${index}].role` });
    if (typeof message.content !== "string") return failure("validation.jsonMessageContent", { index: displayIndex, path: `messages[${index}].content` });
    if (message.role === "tool" && (typeof message.tool_call_id !== "string" || !message.tool_call_id.trim())) {
      return failure("validation.jsonMessageToolCallId", { index: displayIndex, path: `messages[${index}].tool_call_id` });
    }
    if (message.role !== "tool" && message.tool_call_id !== undefined) return failure("validation.jsonMessageUnexpectedToolCallId", { index: displayIndex, path: `messages[${index}].tool_call_id` });
    messages.push({ id: createID("message"), role: message.role, content: message.content, toolCallId: typeof message.tool_call_id === "string" ? message.tool_call_id : undefined });
  }

  const parameters = parsed.parameters ?? {};
  if (!isRecord(parameters)) return failure("validation.jsonParametersObject");
  const unsupportedParameters = Object.keys(parameters).filter((key) => !RAW_PARAMETER_KEYS.has(key));
  if (unsupportedParameters.length) return failure("validation.jsonUnsupportedParameters", { fields: unsupportedParameters.join(", ") });
  const parameterValidation = validateRawParameters(parameters);
  if (!parameterValidation.ok) return parameterValidation;

  const toolsResult = parseRawTools(parsed.tools ?? []);
  if (!toolsResult.ok) return toolsResult;
  const next = structuredClone(column);
  next.messages = messages;
  next.tools = toolsResult.tools;
  next.config = applyRawParameters(next.config, parameters);
  next.jsonError = "";
  next.dirty = true;
  return { ok: true, column: next };
}

export function validateAndBuildRunPayload(column: PlaygroundColumnDraft, comparisonID: string) {
  if (!column.selectedModel.trim()) return failure("validation.modelRequired");
  if (column.messages.length === 0 || column.messages.every((message) => !message.content.trim())) return failure("validation.messagesRequired");
  for (let index = 0; index < column.messages.length; index += 1) {
    const message = column.messages[index];
    if (!message.content.trim()) return failure("validation.messageEmpty", { index: index + 1 });
    if (message.role === "tool" && !message.toolCallId?.trim()) return failure("validation.messageToolCallId", { index: index + 1 });
  }
  const variables = getColumnVariables(column);
  const missing = variables.filter((name) => !(column.variableValues[name] ?? "").trim());
  if (missing.length) return failure("validation.variablesMissing", { names: missing.join(", ") });
  if (column.config.stream && column.config.outputGuardrails.length) return failure("validation.outputGuardrailsRequireStreamingOff");

  const numeric = validateNumericParameters(column.config);
  if (!numeric.ok) return numeric;
  const toolsResult = buildTools(column.tools);
  if (!toolsResult.ok) return toolsResult;
  const metadata = buildMetadata(column.config.metadata);
  if (!metadata.ok) return metadata;

  const renderedMessages = column.messages.map((message) => {
    const rendered = renderPromptText(message.content, column.variableValues).rendered;
    return { role: message.role, content: rendered, ...(message.role === "tool" ? { tool_call_id: message.toolCallId } : {}) };
  });
  const request: Record<string, unknown> = {
    model: column.selectedModel.trim(),
    messages: renderedMessages,
    ...buildParameters(column, true),
    metadata: metadata.value,
  };
  if (toolsResult.tools.length) request.tools = toolsResult.tools;

  return {
    ok: true as const,
    payload: {
      client_run_id: createID("run"),
      comparison_id: comparisonID,
      column_id: column.id,
      request,
      godarai: {
        cache_config: column.config.cacheEnabled ? {
          type: column.config.cacheType,
          ttl: Number(column.config.cacheTTL),
          ...(column.config.cacheNamespace.trim() ? { namespace: column.config.cacheNamespace.trim() } : {}),
          ...(column.config.cacheType === "smart" ? { similarity_threshold: Number(column.config.cacheSimilarity) } : {}),
        } : null,
        guardrails: { input: column.config.inputGuardrails, output: column.config.outputGuardrails, scope: column.config.guardrailScope },
      },
    },
  };
}

export function buildPlaygroundManifest(column: PlaygroundColumnDraft, name: string, description: string): PromptManifest {
  const parameters = buildParameters(column, false);
  const tools = buildTools(column.tools);
  if (tools.ok && tools.tools.length) parameters.tools = tools.tools;
  return {
    type: "chat_prompt",
    name,
    description,
    messages: column.messages.map((message) => ({ role: message.role, content: message.content, ...(message.role === "tool" ? { tool_call_id: message.toolCallId } : {}) })),
    variables: Object.fromEntries(getColumnVariables(column).map((name) => [name, column.variableValues[name] ?? ""])),
    model: column.selectedModel,
    parameters,
    metadata: Object.fromEntries(column.config.metadata.filter((row) => row.key.trim()).map((row) => [row.key.trim(), row.value])),
    model_settings: {
      stream: column.config.stream,
      ...(column.config.temperature ? { temperature: Number(column.config.temperature) } : {}),
      ...(column.config.topP ? { top_p: Number(column.config.topP) } : {}),
      ...(column.config.maxTokens ? { max_tokens: Number(column.config.maxTokens) } : {}),
    },
    cache: column.config.cacheEnabled ? {
      enabled: true, type: column.config.cacheType, ttl: Number(column.config.cacheTTL), namespace: column.config.cacheNamespace,
      ...(column.config.cacheType === "smart" ? { similarity_threshold: Number(column.config.cacheSimilarity) } : {}),
    } : { enabled: false },
    guardrails: { input: column.config.inputGuardrails, output: column.config.outputGuardrails, scope: column.config.guardrailScope },
    structured_output: column.config.structuredEnabled ? {
      enabled: true,
      format: column.config.structuredFormat,
      strict: column.config.structuredStrict,
      ...(column.config.structuredFormat === "json_schema" ? { schema: JSON.parse(column.config.structuredSchema) as Record<string, unknown> } : {}),
    } : undefined,
  };
}

export function hydratePlaygroundColumn(version: PromptVersion, label: string): PlaygroundColumnDraft {
  const manifest = version.manifest ?? {
    type: "chat_prompt", name: version.name, description: version.description, messages: version.messages,
    variables: version.variables, model: version.model, parameters: version.parameters, metadata: version.metadata,
  };
  const config = createDefaultConfig();
  const parameters = manifest.parameters ?? {};
  const hydratedConfig = applyRawParameters(config, parameters);
  hydratedConfig.stream = manifest.model_settings?.stream ?? hydratedConfig.stream;
  hydratedConfig.temperature = numberInput(manifest.model_settings?.temperature ?? parameters.temperature);
  hydratedConfig.topP = numberInput(manifest.model_settings?.top_p ?? parameters.top_p);
  hydratedConfig.maxTokens = numberInput(manifest.model_settings?.max_tokens ?? parameters.max_tokens);
  hydratedConfig.cacheEnabled = manifest.cache?.enabled ?? false;
  hydratedConfig.cacheType = manifest.cache?.type ?? "simple";
  hydratedConfig.cacheTTL = numberInput(manifest.cache?.ttl ?? 600);
  hydratedConfig.cacheNamespace = manifest.cache?.namespace ?? "";
  hydratedConfig.cacheSimilarity = numberInput(manifest.cache?.similarity_threshold ?? 0.9);
  hydratedConfig.inputGuardrails = manifest.guardrails?.input ?? [];
  hydratedConfig.outputGuardrails = manifest.guardrails?.output ?? [];
  hydratedConfig.guardrailScope = manifest.guardrails?.scope ?? "last";
  hydratedConfig.structuredEnabled = manifest.structured_output?.enabled ?? false;
  hydratedConfig.structuredFormat = manifest.structured_output?.format ?? "json_schema";
  hydratedConfig.structuredStrict = manifest.structured_output?.strict ?? true;
  if (manifest.structured_output?.schema) hydratedConfig.structuredSchema = JSON.stringify(manifest.structured_output.schema, null, 2);
  hydratedConfig.metadata = Object.entries(manifest.metadata ?? {}).map(([key, value]) => ({ id: createID("metadata"), key, value: typeof value === "string" ? value : JSON.stringify(value) }));

  const rawTools = Array.isArray(parameters.tools) ? parameters.tools : [];
  const tools = parseRawTools(rawTools);
  const column: PlaygroundColumnDraft = {
    id: createID("column"), label, selectedModel: manifest.model ?? "", selectedProvider: "",
    messages: manifest.messages.map((message) => ({ id: createID("message"), role: isRole(message.role) ? message.role : "user", content: typeof message.content === "string" ? message.content : JSON.stringify(message.content ?? ""), toolCallId: message.tool_call_id })),
    tools: tools.ok ? tools.tools : [], config: hydratedConfig, variableValues: { ...(manifest.variables ?? {}) },
    editorMode: "visual", jsonDraft: "", jsonError: "", sourcePromptID: version.prompt_id,
    sourceVersionID: version.id, sourceFQN: version.fqn, sourcePromptName: version.name,
    sourcePromptDescription: version.description ?? manifest.description ?? "", dirty: false,
  };
  column.jsonDraft = serializeColumnJSON(column);
  return column;
}

function buildParameters(column: PlaygroundColumnDraft, includeStructured: boolean): Record<string, unknown> {
  const config = column.config;
  const parameters: Record<string, unknown> = { stream: config.stream };
  if (config.temperature.trim()) parameters.temperature = Number(config.temperature);
  if (config.topP.trim()) parameters.top_p = Number(config.topP);
  if (config.maxTokens.trim()) parameters.max_tokens = Number(config.maxTokens);
  if (config.stop.trim()) parameters.stop = config.stop.split("\n").filter(Boolean);
  if (config.reasoningEffort) parameters.reasoning_effort = config.reasoningEffort;
  if (column.tools.length) {
    parameters.tool_choice = config.toolChoice || "auto";
    parameters.parallel_tool_calls = config.parallelToolCalls;
  }
  if (includeStructured && config.structuredEnabled) {
    if (config.structuredFormat === "json_object") {
      parameters.response_format = { type: "json_object" };
    } else {
      const schema = tryParseObject(config.structuredSchema);
      if (schema) parameters.response_format = { type: "json_schema", json_schema: { name: config.structuredName || "playground_output", strict: config.structuredStrict, schema } };
    }
  }
  return parameters;
}

function validateRawParameters(parameters: Record<string, unknown>): { ok: true } | PlaygroundDomainFailure {
  if (parameters.stream !== undefined && typeof parameters.stream !== "boolean") return failure("validation.jsonParameterType", { name: "stream" });
  if (parameters.temperature !== undefined && (typeof parameters.temperature !== "number" || !Number.isFinite(parameters.temperature) || parameters.temperature < 0 || parameters.temperature > 2)) return failure("validation.temperatureRange");
  if (parameters.top_p !== undefined && (typeof parameters.top_p !== "number" || !Number.isFinite(parameters.top_p) || parameters.top_p < 0 || parameters.top_p > 1)) return failure("validation.topPRange");
  if (parameters.max_tokens !== undefined && (typeof parameters.max_tokens !== "number" || !Number.isInteger(parameters.max_tokens) || parameters.max_tokens < 1)) return failure("validation.maxTokensPositive");
  if (parameters.reasoning_effort !== undefined && typeof parameters.reasoning_effort !== "string") return failure("validation.jsonParameterType", { name: "reasoning_effort" });
  if (parameters.parallel_tool_calls !== undefined && typeof parameters.parallel_tool_calls !== "boolean") return failure("validation.jsonParameterType", { name: "parallel_tool_calls" });
  if (parameters.tool_choice !== undefined && typeof parameters.tool_choice !== "string") return failure("validation.jsonParameterType", { name: "tool_choice" });
  if (parameters.stop !== undefined && typeof parameters.stop !== "string" && (!Array.isArray(parameters.stop) || parameters.stop.some((item) => typeof item !== "string"))) return failure("validation.jsonParameterType", { name: "stop" });

  if (parameters.response_format !== undefined) {
    if (!isRecord(parameters.response_format)) return failure("validation.jsonResponseFormat");
    const unsupportedResponseFields = unsupportedKeys(parameters.response_format, ["type", "json_schema"]);
    if (unsupportedResponseFields.length) return failure("validation.jsonResponseFormat");
    if (parameters.response_format.type === "json_object") {
      if (parameters.response_format.json_schema !== undefined) return failure("validation.jsonResponseFormat");
    } else if (parameters.response_format.type === "json_schema") {
      const jsonSchema = parameters.response_format.json_schema;
      if (!isRecord(jsonSchema) || typeof jsonSchema.name !== "string" || !jsonSchema.name.trim() || !isRecord(jsonSchema.schema)) return failure("validation.jsonResponseFormat");
      if (jsonSchema.strict !== undefined && typeof jsonSchema.strict !== "boolean") return failure("validation.jsonResponseFormat");
      if (unsupportedKeys(jsonSchema, ["name", "strict", "schema"]).length) return failure("validation.jsonResponseFormat");
    } else {
      return failure("validation.jsonResponseFormat");
    }
  }

  return { ok: true };
}

function buildTools(tools: PlaygroundTool[]): { ok: true; tools: unknown[] } | PlaygroundDomainFailure {
  const result: unknown[] = [];
  for (let index = 0; index < tools.length; index += 1) {
    const tool = tools[index];
    if (!/^[A-Za-z0-9_-]{1,64}$/.test(tool.name.trim())) return failure("validation.toolName", { index: index + 1 });
    try {
      const schema = JSON.parse(tool.schema);
      if (!isRecord(schema)) return failure("validation.toolSchemaObject", { name: tool.name || index + 1 });
      result.push({ type: "function", function: { name: tool.name.trim(), description: tool.description.trim(), parameters: schema } });
    } catch {
      return failure("validation.toolSchemaInvalid", { name: tool.name || index + 1 });
    }
  }
  return { ok: true, tools: result };
}

function parseRawTools(value: unknown): { ok: true; tools: PlaygroundTool[] } | PlaygroundDomainFailure {
  if (!Array.isArray(value)) return failure("validation.jsonToolsArray");
  const tools: PlaygroundTool[] = [];
  for (let index = 0; index < value.length; index += 1) {
    const entry = value[index];
    const displayIndex = index + 1;
    if (!isRecord(entry) || entry.type !== "function" || !isRecord(entry.function) || typeof entry.function.name !== "string" || !isRecord(entry.function.parameters)) {
      return failure("validation.jsonToolStructure", { index: displayIndex, path: `tools[${index}]` });
    }
    const unsupportedToolFields = unsupportedKeys(entry, ["type", "function"]);
    const unsupportedFunctionFields = unsupportedKeys(entry.function, ["name", "description", "parameters"]);
    if (unsupportedToolFields.length || unsupportedFunctionFields.length) {
      return failure("validation.jsonToolUnsupportedFields", {
        index: displayIndex,
        path: `tools[${index}]`,
        fields: [...unsupportedToolFields, ...unsupportedFunctionFields.map((field) => `function.${field}`)].join(", "),
      });
    }
    if (!/^[A-Za-z0-9_-]{1,64}$/.test(entry.function.name.trim())) return failure("validation.toolName", { index: index + 1 });
    if (entry.function.description !== undefined && typeof entry.function.description !== "string") return failure("validation.jsonToolDescription", { index: displayIndex, path: `tools[${index}].function.description` });
    tools.push({ id: createID("tool"), name: entry.function.name, description: typeof entry.function.description === "string" ? entry.function.description : "", schema: JSON.stringify(entry.function.parameters, null, 2) });
  }
  return { ok: true, tools };
}

function applyRawParameters(config: PlaygroundConfig, parameters: Record<string, unknown>): PlaygroundConfig {
  const next = structuredClone(config);
  if (typeof parameters.stream === "boolean") next.stream = parameters.stream;
  next.temperature = numberInput(parameters.temperature);
  next.topP = numberInput(parameters.top_p);
  next.maxTokens = numberInput(parameters.max_tokens);
  next.stop = Array.isArray(parameters.stop) ? parameters.stop.filter((item): item is string => typeof item === "string").join("\n") : typeof parameters.stop === "string" ? parameters.stop : "";
  next.reasoningEffort = typeof parameters.reasoning_effort === "string" ? parameters.reasoning_effort : "";
  next.parallelToolCalls = typeof parameters.parallel_tool_calls === "boolean" ? parameters.parallel_tool_calls : true;
  next.toolChoice = typeof parameters.tool_choice === "string" ? parameters.tool_choice : "auto";
  next.structuredEnabled = false;
  if (isRecord(parameters.response_format)) {
    next.structuredEnabled = true;
    next.structuredFormat = parameters.response_format.type === "json_object" ? "json_object" : "json_schema";
    if (isRecord(parameters.response_format.json_schema)) {
      next.structuredName = typeof parameters.response_format.json_schema.name === "string" ? parameters.response_format.json_schema.name : "playground_output";
      next.structuredStrict = parameters.response_format.json_schema.strict !== false;
      if (isRecord(parameters.response_format.json_schema.schema)) next.structuredSchema = JSON.stringify(parameters.response_format.json_schema.schema, null, 2);
    }
  }
  return next;
}

function validateNumericParameters(config: PlaygroundConfig): { ok: true } | PlaygroundDomainFailure {
  const temperature = optionalNumber(config.temperature);
  const topP = optionalNumber(config.topP);
  const maxTokens = optionalNumber(config.maxTokens);
  if (temperature.invalid || (temperature.value !== undefined && (temperature.value < 0 || temperature.value > 2))) return failure("validation.temperatureRange");
  if (topP.invalid || (topP.value !== undefined && (topP.value < 0 || topP.value > 1))) return failure("validation.topPRange");
  if (maxTokens.invalid || (maxTokens.value !== undefined && (!Number.isInteger(maxTokens.value) || maxTokens.value < 1))) return failure("validation.maxTokensPositive");
  if (config.cacheEnabled && (!Number.isInteger(Number(config.cacheTTL)) || Number(config.cacheTTL) < 1)) return failure("validation.cacheTTLPositive");
  if (config.cacheEnabled && config.cacheType === "smart" && (Number(config.cacheSimilarity) < 0.8 || Number(config.cacheSimilarity) > 1)) return failure("validation.cacheSimilarityRange");
  if (config.structuredEnabled && config.structuredFormat === "json_schema") {
    try { if (!isRecord(JSON.parse(config.structuredSchema))) return failure("validation.structuredSchemaObject"); }
    catch { return failure("validation.structuredSchemaInvalid"); }
  }
  return { ok: true };
}

function buildMetadata(rows: PlaygroundMetadataRow[]): { ok: true; value: Record<string, string> } | PlaygroundDomainFailure {
  const value: Record<string, string> = {};
  for (const row of rows) {
    const key = row.key.trim();
    if (!key && !row.value.trim()) continue;
    if (!/^[A-Za-z0-9_.-]{1,64}$/.test(key)) return failure("validation.metadataKey");
    const normalizedKey = key.toLowerCase();
    if (["source", "client_run_id", "comparison_id", "column_id", "__proto__", "prototype", "constructor"].includes(normalizedKey) || normalizedKey.includes("authorization")) {
      return failure("validation.metadataReserved", { key });
    }
    if (row.value.length > 256) return failure("validation.metadataValueLength", { key });
    if (Object.hasOwn(value, key)) return failure("validation.metadataDuplicate", { key });
    value[key] = row.value;
  }
  if (Object.keys(value).length > 20) return failure("validation.metadataTooMany");
  return { ok: true, value };
}

function optionalNumber(value: string): { value?: number; invalid: boolean } {
  if (!value.trim()) return { invalid: false };
  const parsed = Number(value);
  return { value: parsed, invalid: !Number.isFinite(parsed) };
}

function numberInput(value: unknown): string {
  return typeof value === "number" && Number.isFinite(value) ? String(value) : typeof value === "string" ? value : "";
}

function safeParseObject(value: string): Record<string, unknown> {
  try {
    const parsed = JSON.parse(value);
    return isRecord(parsed) ? parsed : {};
  } catch {
    return {};
  }
}

function tryParseObject(value: string): Record<string, unknown> | undefined {
  try {
    const parsed = JSON.parse(value);
    return isRecord(parsed) ? parsed : undefined;
  } catch {
    return undefined;
  }
}

function unsupportedKeys(value: Record<string, unknown>, allowed: string[]): string[] {
  const accepted = new Set(allowed);
  return Object.keys(value).filter((key) => !accepted.has(key));
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return Boolean(value) && typeof value === "object" && !Array.isArray(value);
}

function isRole(value: unknown): value is PlaygroundRole {
  return value === "system" || value === "user" || value === "assistant" || value === "tool";
}

function failure(errorKey: string, errorValues?: Record<string, string | number>): PlaygroundDomainFailure {
  return { ok: false, errorKey, ...(errorValues ? { errorValues } : {}) };
}

function createID(prefix: string): string {
  const suffix = globalThis.crypto?.randomUUID?.() ?? Math.random().toString(36).slice(2);
  return `${prefix}_${suffix}`;
}
