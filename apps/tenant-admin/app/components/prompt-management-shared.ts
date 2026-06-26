import type { ProviderCatalogModel, ProviderCatalogProvider } from "@/app/components/provider-account-wizard/types";

import {
  buildPromptUsageSnippet,
  extractPromptVariableNames,
  normalizePromptSlug,
  renderPromptText,
  type PromptEditorMessage,
} from "./prompt-management-utils";

export type GatewayModel = {
  id: string;
  object: string;
  owned_by: string;
  provider?: string;
  aliases?: string[];
};

export type GatewayModelListResponse = {
  data?: GatewayModel[];
};

export type GuardrailItem = {
  id: string;
  slug: string;
  name: string;
  description?: string;
  selector?: string;
  type?: string;
};

export type GuardrailListResponse = {
  data?: GuardrailItem[];
};

export type PromptVariableMap = Record<string, string>;

export type PromptModelSettings = {
  stream?: boolean;
  temperature?: number;
  top_p?: number;
  max_tokens?: number;
};

export type PromptCacheConfig = {
  enabled: boolean;
  type?: "simple" | "smart";
  ttl?: number;
  namespace?: string;
  similarity_threshold?: number;
};

export type PromptGuardrailConfig = {
  input?: string[];
  output?: string[];
  scope?: "all" | "last";
};

export type StructuredOutputConfig = {
  enabled: boolean;
  format?: "json_schema" | "json_object";
  strict?: boolean;
  schema?: Record<string, unknown>;
};

export type PromptManifest = {
  type: string;
  name: string;
  description?: string;
  messages: Array<{ role: string; content?: unknown }>;
  variables?: PromptVariableMap;
  model?: string;
  parameters?: Record<string, unknown>;
  metadata?: Record<string, unknown>;
  version_alias?: string;
  model_settings?: PromptModelSettings;
  cache?: PromptCacheConfig;
  guardrails?: PromptGuardrailConfig;
  structured_output?: StructuredOutputConfig;
  logging?: {
    enabled: boolean;
    store_logs: boolean;
  };
};

export type PromptVersion = {
  id: string;
  prompt_id: string;
  version: number;
  fqn: string;
  version_alias?: string;
  name: string;
  description?: string;
  messages: Array<{ role: string; content?: unknown }>;
  variables?: PromptVariableMap;
  model?: string;
  parameters?: Record<string, unknown>;
  metadata?: Record<string, unknown>;
  manifest?: PromptManifest;
  commit_message?: string;
  created_by?: string;
  created_at?: string;
};

export type Prompt = {
  id: string;
  name: string;
  slug: string;
  description?: string;
  latest_version: number;
  latest_version_detail?: PromptVersion;
  created_by?: string;
  created_at?: string;
  updated_at?: string;
};

export type TenantAdminUserLookup = {
  id: string;
  email?: string;
  mobile?: string;
  display_name?: string;
};

export type TenantAdminUserLookupResponse = {
  data?: TenantAdminUserLookup[];
  total?: number;
};

export type PromptListResponse = {
  data?: Prompt[];
  total?: number;
  limit?: number;
  offset?: number;
  next_offset?: number | null;
};

export type PromptVersionListResponse = {
  data?: PromptVersion[];
};

export type PromptRunResponse = {
  choices?: Array<{
    message?: {
      role?: string;
      content?: unknown;
    };
  }>;
  usage?: {
    prompt_tokens?: number;
    completion_tokens?: number;
    total_tokens?: number;
  };
  model?: string;
  id?: string;
};

export type MetadataRow = {
  id: string;
  key: string;
  value: string;
};

export type Notice = { kind: "success" | "error"; message: string } | null;

export type BuilderState = {
  name: string;
  description: string;
  selectedModel: string;
  selectedProvider: string;
  systemPrompt: string;
  messages: PromptEditorMessage[];
  composerRole: "user" | "assistant";
  composerText: string;
  commitMessage: string;
  variableValues: PromptVariableMap;
  stream: boolean;
  temperature: string;
  topP: string;
  maxTokens: string;
  cacheEnabled: boolean;
  cacheType: "simple" | "smart";
  cacheTTL: string;
  cacheNamespace: string;
  cacheSimilarityThreshold: string;
  inputGuardrails: string[];
  outputGuardrails: string[];
  guardrailScope: "all" | "last";
  structuredEnabled: boolean;
  structuredFormat: "json_schema" | "json_object";
  structuredStrict: boolean;
  structuredSchema: string;
  metadataRows: MetadataRow[];
};

export type PromptTranslationFn = (key: string, values?: Record<string, string | number | boolean>) => string;

export type PromptCatalogState = {
  models: GatewayModel[];
  providerCatalog: ProviderCatalogProvider[];
  providerModels: ProviderCatalogModel[];
  guardrails: GuardrailItem[];
  isModelsLoading: boolean;
  isGuardrailsLoading: boolean;
};

export const DEFAULT_SCHEMA = JSON.stringify(
  {
    type: "object",
    properties: {
      message: {
        type: "string",
      },
    },
    required: ["message"],
  },
  null,
  2,
);

export function createEmptyBuilder(defaultSystemPrompt: string): BuilderState {
  return {
    name: "",
    description: "",
    selectedModel: "",
    selectedProvider: "",
    systemPrompt: defaultSystemPrompt,
    messages: [],
    composerRole: "user",
    composerText: "",
    commitMessage: "",
    variableValues: {},
    stream: true,
    temperature: "0.2",
    topP: "",
    maxTokens: "",
    cacheEnabled: false,
    cacheType: "simple",
    cacheTTL: "600",
    cacheNamespace: "",
    cacheSimilarityThreshold: "0.9",
    inputGuardrails: [],
    outputGuardrails: [],
    guardrailScope: "last",
    structuredEnabled: false,
    structuredFormat: "json_schema",
    structuredStrict: true,
    structuredSchema: DEFAULT_SCHEMA,
    metadataRows: [],
  };
}

export function createMetadataRow(key = "", value = ""): MetadataRow {
  return {
    id: `meta_${Math.random().toString(36).slice(2, 10)}`,
    key,
    value,
  };
}

export function buildFallbackManifest(version: PromptVersion): PromptManifest {
  return {
    type: "chat_prompt",
    name: version.name,
    description: version.description,
    messages: version.messages,
    variables: version.variables,
    model: version.model,
    parameters: version.parameters,
    metadata: version.metadata,
    version_alias: version.version_alias,
  };
}

export function buildManifest(
  builder: BuilderState,
  t: PromptTranslationFn,
): { ok: true; manifest: PromptManifest } | { ok: false; message: string } {
  const name = normalizePromptSlug(builder.name);
  const detectedVariables = extractPromptVariableNames([
    builder.systemPrompt,
    ...builder.messages.map((message) => message.content),
  ]);
  if (!/^[A-Za-z0-9_-]{1,256}$/.test(name)) {
    return { ok: false, message: t("validation.invalidName") };
  }
  if (!builder.selectedModel.trim()) {
    return { ok: false, message: t("validation.modelRequired") };
  }
  const temperature = parseOptionalNumber(builder.temperature);
  const topP = parseOptionalNumber(builder.topP);
  const maxTokens = parseOptionalInteger(builder.maxTokens);
  if (temperature !== undefined && (temperature < 0 || temperature > 2)) {
    return { ok: false, message: t("validation.temperatureRange") };
  }
  if (topP !== undefined && (topP < 0 || topP > 1)) {
    return { ok: false, message: t("validation.topPRange") };
  }
  if (maxTokens !== undefined && maxTokens < 1) {
    return { ok: false, message: t("validation.maxTokensMin") };
  }

  const metadata = buildMetadataObject(builder.metadataRows, t);
  if (!metadata.ok) return metadata;

  let structuredOutput: StructuredOutputConfig | undefined;
  if (builder.structuredEnabled) {
    if (builder.structuredFormat === "json_schema") {
      try {
        const schema = JSON.parse(builder.structuredSchema) as Record<string, unknown>;
        if (!schema || typeof schema !== "object" || Array.isArray(schema)) {
          return { ok: false, message: t("validation.structuredSchemaObject") };
        }
        structuredOutput = {
          enabled: true,
          format: "json_schema",
          strict: builder.structuredStrict,
          schema,
        };
      } catch {
        return { ok: false, message: t("validation.structuredSchemaInvalidJson") };
      }
    } else {
      structuredOutput = {
        enabled: true,
        format: "json_object",
        strict: builder.structuredStrict,
      };
    }
  }

  const cacheConfig = buildCacheConfig(builder, t);
  if (!cacheConfig.ok) return cacheConfig;

  if (builder.stream && builder.outputGuardrails.length > 0) {
    return { ok: false, message: t("validation.outputGuardrailsRequireStreamingOff") };
  }

  const promptMessages: Array<{ role: string; content?: unknown }> = [];
  if (builder.systemPrompt.trim()) {
    promptMessages.push({ role: "system", content: builder.systemPrompt.trim() });
  }
  for (const message of builder.messages) {
    if (message.content.trim()) {
      promptMessages.push({ role: message.role, content: message.content.trim() });
    }
  }
  if (promptMessages.length === 0) {
    return { ok: false, message: t("validation.messagesRequired") };
  }

  const modelSettings: PromptModelSettings = {
    stream: builder.stream,
    ...(temperature !== undefined ? { temperature } : {}),
    ...(topP !== undefined ? { top_p: topP } : {}),
    ...(maxTokens !== undefined ? { max_tokens: maxTokens } : {}),
  };
  const parameters: Record<string, unknown> = {
    stream: builder.stream,
    ...(temperature !== undefined ? { temperature } : {}),
    ...(topP !== undefined ? { top_p: topP } : {}),
    ...(maxTokens !== undefined ? { max_tokens: maxTokens } : {}),
  };

  return {
    ok: true,
    manifest: {
      type: "chat_prompt",
      name,
      description: builder.description.trim(),
      messages: promptMessages,
      variables: Object.fromEntries(detectedVariables.map((variable) => [variable, builder.variableValues[variable] ?? ""])),
      model: builder.selectedModel.trim(),
      parameters,
      metadata: metadata.value,
      model_settings: modelSettings,
      cache: cacheConfig.value,
      guardrails: {
        input: builder.inputGuardrails,
        output: builder.outputGuardrails,
        scope: builder.guardrailScope,
      },
      structured_output: structuredOutput,
    },
  };
}

export function buildRunPayload(
  builder: BuilderState,
  manifest: PromptManifest,
  t: PromptTranslationFn,
) {
  const requiredVariables = extractPromptVariableNames([
    builder.systemPrompt,
    ...builder.messages.map((message) => message.content),
  ]);
  const missingVariables = requiredVariables.filter((variable) => !(builder.variableValues[variable] ?? "").trim());
  if (missingVariables.length > 0) {
    return { ok: false as const, message: t("validation.runtimeVariablesMissing", { names: missingVariables.join(", ") }) };
  }

  const renderedMessages: Array<{ role: string; content: string }> = [];
  const allMissing = new Set<string>();
  if (builder.systemPrompt.trim()) {
    const rendered = renderPromptText(builder.systemPrompt.trim(), builder.variableValues);
    rendered.missing.forEach((item) => allMissing.add(item));
    renderedMessages.push({ role: "system", content: rendered.rendered });
  }
  for (const message of builder.messages) {
    const rendered = renderPromptText(message.content.trim(), builder.variableValues);
    rendered.missing.forEach((item) => allMissing.add(item));
    renderedMessages.push({ role: message.role, content: rendered.rendered });
  }
  if (allMissing.size > 0) {
    return { ok: false as const, message: t("validation.renderVariablesMissing", { names: [...allMissing].join(", ") }) };
  }

  const request: Record<string, unknown> = {
    model: manifest.model,
    messages: renderedMessages,
    stream: builder.stream,
    metadata: manifest.metadata ?? {},
  };
  const temperature = parseOptionalNumber(builder.temperature);
  const topP = parseOptionalNumber(builder.topP);
  const maxTokens = parseOptionalInteger(builder.maxTokens);
  if (temperature !== undefined) request.temperature = temperature;
  if (topP !== undefined) request.top_p = topP;
  if (maxTokens !== undefined) request.max_tokens = maxTokens;

  if (builder.structuredEnabled) {
    if (builder.structuredFormat === "json_schema") {
      const schema = JSON.parse(builder.structuredSchema) as Record<string, unknown>;
      request.response_format = {
        type: "json_schema",
        json_schema: {
          name: normalizePromptSlug(manifest.name || "prompt_output"),
          schema,
          strict: builder.structuredStrict,
        },
      };
    } else {
      request.response_format = { type: "json_object" };
    }
  }

  const cacheConfig = builder.cacheEnabled
    ? builder.cacheType === "smart"
      ? {
          type: "smart",
          ttl: parseOptionalInteger(builder.cacheTTL),
          namespace: builder.cacheNamespace.trim() || undefined,
          similarity_threshold: parseOptionalNumber(builder.cacheSimilarityThreshold),
        }
      : {
          type: "simple",
          ttl: parseOptionalInteger(builder.cacheTTL),
          namespace: builder.cacheNamespace.trim() || undefined,
        }
    : null;

  return {
    ok: true as const,
    payload: {
      request,
      cache_config: cacheConfig,
      guardrails: {
        input: builder.inputGuardrails,
        output: builder.outputGuardrails,
        scope: builder.guardrailScope,
      },
    },
  };
}

export function previewRunResponse(response: PromptRunResponse): string {
  const content = response.choices?.[0]?.message?.content;
  if (typeof content === "string") {
    return content;
  }
  return JSON.stringify(response, null, 2);
}

export function findProviderForModel(modelID: string, models: GatewayModel[]): string {
  const found = models.find((item) => item.id === modelID);
  return found?.provider?.trim().toLowerCase() || "";
}

export function formatNumber(value: number, locale: string) {
  return new Intl.NumberFormat(locale).format(value);
}

export function formatDate(value: string | undefined, locale: string) {
  if (!value) return "-";
  return new Intl.DateTimeFormat(locale, {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(value));
}

export function buildPromptSnippet(fqn: string, model?: string) {
  return buildPromptUsageSnippet(fqn, model);
}

function buildCacheConfig(
  builder: BuilderState,
  t: PromptTranslationFn,
): { ok: true; value?: PromptCacheConfig } | { ok: false; message: string } {
  if (!builder.cacheEnabled) {
    return { ok: true, value: { enabled: false } };
  }
  const ttl = parseOptionalInteger(builder.cacheTTL);
  if (ttl === undefined || ttl < 1) {
    return { ok: false, message: t("validation.cacheTtl") };
  }
  if (builder.cacheType === "smart") {
    const similarity = parseOptionalNumber(builder.cacheSimilarityThreshold);
    if (similarity === undefined || similarity < 0.8 || similarity > 1) {
      return { ok: false, message: t("validation.smartCacheSimilarity") };
    }
    return {
      ok: true,
      value: {
        enabled: true,
        type: "smart",
        ttl,
        namespace: builder.cacheNamespace.trim(),
        similarity_threshold: similarity,
      },
    };
  }
  return {
    ok: true,
    value: {
      enabled: true,
      type: "simple",
      ttl,
      namespace: builder.cacheNamespace.trim(),
    },
  };
}

function buildMetadataObject(
  rows: MetadataRow[],
  t: PromptTranslationFn,
): { ok: true; value: Record<string, unknown> } | { ok: false; message: string } {
  const out: Record<string, unknown> = {};
  for (const row of rows) {
    const key = row.key.trim();
    const value = row.value.trim();
    if (!key && !value) continue;
    if (!key) {
      return { ok: false, message: t("validation.metadataKeyRequired") };
    }
    if (key in out) {
      return { ok: false, message: t("validation.metadataKeyDuplicate", { key }) };
    }
    out[key] = value;
  }
  return { ok: true, value: out };
}

function parseOptionalNumber(value: string): number | undefined {
  if (!value.trim()) return undefined;
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : undefined;
}

function parseOptionalInteger(value: string): number | undefined {
  if (!value.trim()) return undefined;
  const parsed = Number.parseInt(value, 10);
  return Number.isFinite(parsed) ? parsed : undefined;
}

function numberToInput(value?: number): string {
  return typeof value === "number" && Number.isFinite(value) ? `${value}` : "";
}

function numberFromUnknown(value: unknown): number | undefined {
  if (typeof value === "number" && Number.isFinite(value)) return value;
  if (typeof value === "string" && value.trim()) {
    const parsed = Number(value);
    return Number.isFinite(parsed) ? parsed : undefined;
  }
  return undefined;
}

function booleanFromUnknown(value: unknown, fallback = false): boolean {
  if (typeof value === "boolean") return value;
  if (typeof value === "string") return value.trim().toLowerCase() === "true";
  return fallback;
}

function stringifyScalar(value: unknown): string {
  if (value == null) return "";
  if (typeof value === "string") return value;
  if (typeof value === "number" || typeof value === "boolean") return `${value}`;
  return JSON.stringify(value);
}

export function hydrateBuilderFromVersion(
  version: PromptVersion,
  defaultSystemPrompt: string,
  models: GatewayModel[],
): BuilderState {
  const manifest = version.manifest ?? buildFallbackManifest(version);
  const systemMessage = manifest.messages.find((message) => message.role === "system");
  const threadMessages = manifest.messages
    .filter((message) => message.role === "user" || message.role === "assistant")
    .map((message) => ({
      role: message.role as "user" | "assistant",
      content: typeof message.content === "string" ? message.content : "",
    }));
  const metadataRows = Object.entries(manifest.metadata ?? {}).map(([key, value]) => createMetadataRow(key, stringifyScalar(value)));

  return {
    name: manifest.name || version.name,
    description: manifest.description || version.description || "",
    selectedModel: manifest.model || version.model || "",
    selectedProvider: findProviderForModel(manifest.model || version.model || "", models),
    systemPrompt: typeof systemMessage?.content === "string" ? systemMessage.content : defaultSystemPrompt,
    messages: threadMessages,
    composerRole: "user",
    composerText: "",
    commitMessage: "",
    variableValues: { ...(manifest.variables ?? version.variables ?? {}) },
    stream: manifest.model_settings?.stream ?? booleanFromUnknown(manifest.parameters?.stream, true),
    temperature: numberToInput(manifest.model_settings?.temperature ?? numberFromUnknown(manifest.parameters?.temperature)),
    topP: numberToInput(manifest.model_settings?.top_p ?? numberFromUnknown(manifest.parameters?.top_p)),
    maxTokens: numberToInput(manifest.model_settings?.max_tokens ?? numberFromUnknown(manifest.parameters?.max_tokens)),
    cacheEnabled: manifest.cache?.enabled ?? false,
    cacheType: manifest.cache?.type ?? "simple",
    cacheTTL: numberToInput(manifest.cache?.ttl ?? 600),
    cacheNamespace: manifest.cache?.namespace ?? "",
    cacheSimilarityThreshold: numberToInput(manifest.cache?.similarity_threshold ?? 0.9),
    inputGuardrails: manifest.guardrails?.input ?? [],
    outputGuardrails: manifest.guardrails?.output ?? [],
    guardrailScope: manifest.guardrails?.scope ?? "last",
    structuredEnabled: manifest.structured_output?.enabled ?? false,
    structuredFormat: manifest.structured_output?.format ?? "json_schema",
    structuredStrict: manifest.structured_output?.strict ?? true,
    structuredSchema: manifest.structured_output?.schema ? JSON.stringify(manifest.structured_output.schema, null, 2) : DEFAULT_SCHEMA,
    metadataRows,
  };
}
