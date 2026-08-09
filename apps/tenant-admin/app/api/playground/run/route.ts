import { NextRequest, NextResponse } from "next/server";

import { modelModesSupportPlaygroundMode } from "@/app/components/playground-mode";
import { getCurrentMePermissions, getNormalizedPermissionSet } from "@/lib/auth/me-permissions";
import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { fetchControlPlaneJSON, proxyToControlPlane, proxyToControlPlaneStream } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

const OPERATIONAL_HEADERS = [
  "x-godarai-trace-id",
  "x-godarai-cache-status",
  "x-godarai-cached-trace-id",
  "x-godarai-cache-similarity-score",
  "x-godarai-cost-usd",
  "x-godarai-cost-irr",
  "x-request-id",
];

const CHAT_REQUEST_FIELDS = new Set([
  "model",
  "messages",
  "stream",
  "stream_options",
  "temperature",
  "top_p",
  "top_k",
  "presence_penalty",
  "frequency_penalty",
  "seed",
  "random_seed",
  "max_tokens",
  "max_output_tokens",
  "stop",
  "logit_bias",
  "logprobs",
  "top_logprobs",
  "reasoning_effort",
  "thinking",
  "response_format",
  "tools",
  "tool_choice",
  "parallel_tool_calls",
]);
const IMAGE_REQUEST_FIELDS = new Set([
  "model",
  "prompt",
  "n",
  "size",
  "quality",
  "response_format",
  "output_format",
  "output_compression",
  "background",
  "moderation",
]);
const ENVELOPE_FIELDS = new Set(["mode", "client_run_id", "comparison_id", "column_id", "request", "godarai"]);
const MESSAGE_FIELDS = new Set(["role", "content", "tool_call_id"]);
const CACHE_FIELDS = new Set(["type", "ttl", "namespace", "similarity_threshold"]);
const GODARAI_FIELDS = new Set(["cache_config", "guardrails", "observability"]);
const OBSERVABILITY_FIELDS = new Set(["tags"]);
const GUARDRAIL_FIELDS = new Set(["input", "output", "scope"]);
const RESERVED_METADATA_KEYS = new Set(["source", "client_run_id", "comparison_id", "column_id", "__proto__", "prototype", "constructor"]);

type PlaygroundRunPayload = {
  mode?: "chat" | "image_generation";
  client_run_id?: string;
  comparison_id?: string;
  column_id?: string;
  request?: Record<string, unknown>;
  godarai?: {
    observability?: {
      tags?: Record<string, unknown>;
    } | null;
    cache_config?: Record<string, unknown> | null;
    guardrails?: {
      input?: string[];
      output?: string[];
      scope?: "all" | "last" | "";
    } | null;
  };
};

export async function POST(request: NextRequest) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json({ error: { message: auth.message } }, { status: auth.status });
  }

  const permissionPayload = await getCurrentMePermissions(auth.tenantSlug, auth.user.id).catch(() => null);
  const permissions = permissionPayload ? getNormalizedPermissionSet(permissionPayload) : new Set<string>();
  if (!permissions.has("playground:run") && !permissions.has("policy:manage")) {
    return NextResponse.json({ error: { message: "You do not have permission to run billable playground requests." } }, { status: 403 });
  }

  const body = (await request.json().catch(() => null)) as PlaygroundRunPayload | null;
  const validation = validatePayload(body);
  if (!validation.ok) {
    return NextResponse.json({ error: { message: validation.message } }, { status: 400 });
  }

  const mode = body?.mode ?? "chat";
  if (mode === "image_generation") {
    const modelValidation = await validateImageModelCapability(request, auth.tenantSlug, auth.user.id, body!.request!.model as string);
    if (!modelValidation.ok) {
      return NextResponse.json({ error: { message: modelValidation.message } }, { status: modelValidation.status });
    }
  }
  const upstreamHeaders = new Headers();
  if (mode === "chat") {
    upstreamHeaders.set("x-godarai-playground-observability", JSON.stringify(buildObservabilityContext(body!)));
  }
  if (mode === "chat" && body?.request?.stream === true) {
    upstreamHeaders.set("x-godarai-include-stream-telemetry", "true");
  }
  const cacheConfig = body?.godarai?.cache_config;
  if (cacheConfig && Object.keys(cacheConfig).length > 0) {
    upstreamHeaders.set("x-godarai-cache-config", JSON.stringify(cacheConfig));
  }

  const input = normalizeStringList(body?.godarai?.guardrails?.input);
  const output = normalizeStringList(body?.godarai?.guardrails?.output);
  if (input.length > 0 || output.length > 0) {
    upstreamHeaders.set("x-godarai-guardrails", JSON.stringify({ llm_input_guardrails: input, llm_output_guardrails: output }));
    const scope = body?.godarai?.guardrails?.scope;
    if (scope) upstreamHeaders.set("x-godarai-guardrails-scope", scope);
  }

  const upstreamRequest: Record<string, unknown> = mode === "image_generation"
    ? { ...(body!.request as Record<string, unknown>) }
    : { ...(body!.request as Record<string, unknown>) };
  const proxyOptions = {
    request,
    method: "POST" as const,
    path: mode === "image_generation" ? "/v1/images/generations" : "/v1/chat/completions",
    body: upstreamRequest,
    userID: auth.user.id,
    upstreamHeaders,
    exposeResponseHeaders: OPERATIONAL_HEADERS,
  };

  return mode === "chat" && upstreamRequest["stream"] === true
    ? proxyToControlPlaneStream(proxyOptions)
    : proxyToControlPlane(proxyOptions);
}

function validatePayload(body: PlaygroundRunPayload | null): { ok: true } | { ok: false; message: string } {
  if (!body || typeof body !== "object" || !body.request || typeof body.request !== "object" || Array.isArray(body.request)) {
    return { ok: false, message: "Invalid playground payload." };
  }
  const unknownEnvelopeFields = Object.keys(body).filter((key) => !ENVELOPE_FIELDS.has(key));
  if (unknownEnvelopeFields.length > 0) return { ok: false, message: `Unsupported playground fields: ${unknownEnvelopeFields.join(", ")}` };
  const mode = body.mode ?? "chat";
  if (mode !== "chat" && mode !== "image_generation") return { ok: false, message: "Playground mode is invalid." };
  for (const [name, value] of [["client_run_id", body.client_run_id], ["comparison_id", body.comparison_id], ["column_id", body.column_id]] as const) {
    if (value !== undefined && (typeof value !== "string" || !/^[A-Za-z0-9_-]{1,100}$/.test(value))) return { ok: false, message: `${name} is invalid.` };
  }
  const requestFields = mode === "image_generation" ? IMAGE_REQUEST_FIELDS : CHAT_REQUEST_FIELDS;
  const unknownFields = Object.keys(body.request).filter((key) => !requestFields.has(key));
  if (unknownFields.length > 0) {
    return { ok: false, message: `Unsupported request fields: ${unknownFields.join(", ")}` };
  }
  if (typeof body.request.model !== "string" || !body.request.model.trim() || body.request.model.length > 512) {
    return { ok: false, message: "A model is required." };
  }
  if (mode === "image_generation") return validateImagePayload(body);
  if (!Array.isArray(body.request.messages) || body.request.messages.length === 0) {
    return { ok: false, message: "At least one message is required." };
  }
  if (body.request.messages.length > 200) {
    return { ok: false, message: "Too many messages." };
  }
  let totalMessageLength = 0;
  for (let index = 0; index < body.request.messages.length; index += 1) {
    const message = body.request.messages[index];
    if (!isRecord(message)) return { ok: false, message: `messages[${index}] must be an object.` };
    const unsupported = Object.keys(message).filter((key) => !MESSAGE_FIELDS.has(key));
    if (unsupported.length > 0) return { ok: false, message: `Unsupported messages[${index}] fields: ${unsupported.join(", ")}` };
    if (!isMessageRole(message.role) || typeof message.content !== "string" || message.content.length === 0) return { ok: false, message: `messages[${index}] is invalid.` };
    totalMessageLength += message.content.length;
    if (message.role === "tool" && (typeof message.tool_call_id !== "string" || !message.tool_call_id.trim() || message.tool_call_id.length > 256)) return { ok: false, message: `messages[${index}].tool_call_id is required.` };
    if (message.role !== "tool" && message.tool_call_id !== undefined) return { ok: false, message: `messages[${index}].tool_call_id is not supported for this role.` };
  }
  if (totalMessageLength > 2_000_000) return { ok: false, message: "Message content is too large." };
  const parameterError = validateRequestParameters(body.request);
  if (parameterError) return { ok: false, message: parameterError };
  if (body.request.tools !== undefined && !Array.isArray(body.request.tools)) {
    return { ok: false, message: "Tools must be an array." };
  }
  if (body.request.tools && body.request.tools.length > 50) {
    return { ok: false, message: "Too many tools." };
  }
  if (Array.isArray(body.request.tools)) {
    for (let index = 0; index < body.request.tools.length; index += 1) {
      const toolError = validateTool(body.request.tools[index], index);
      if (toolError) return { ok: false, message: toolError };
    }
  }
  const godaraiError = validateGodaraiOptions(body.godarai);
  if (godaraiError) return { ok: false, message: godaraiError };
  return { ok: true };
}

async function validateImageModelCapability(request: NextRequest, tenantSlug: string, userID: string, modelID: string) {
  const upstream = await fetchControlPlaneJSON({
    tenantSlug,
    method: "GET",
    path: "/api/v1/models",
    userID,
    requestHeaders: request.headers,
  });
  if (!upstream.ok) {
    return { ok: false as const, status: 502, message: "The model capability list could not be verified." };
  }
  const payload = isRecord(upstream.payload) ? upstream.payload : {};
  const models = Array.isArray(payload.data) ? payload.data : [];
  const model = models.find((entry) => isRecord(entry) && entry.id === modelID);
  if (!isRecord(model)) {
    return { ok: false as const, status: 400, message: "This model is not available in the current workspace." };
  }
  const modes = Array.isArray(model.supported_modes) ? model.supported_modes : Array.isArray(model.types) ? model.types : [];
  if (!modelModesSupportPlaygroundMode(modes, "image_generation")) {
    return { ok: false as const, status: 400, message: "This model is not enabled for image generation in the current workspace." };
  }
  return { ok: true as const };
}

function validateImagePayload(body: PlaygroundRunPayload): { ok: true } | { ok: false; message: string } {
  const request = body.request!;
  if (typeof request.prompt !== "string" || !request.prompt.trim()) return { ok: false, message: "An image prompt is required." };
  if (request.prompt.length > 100_000) return { ok: false, message: "The image prompt is too large." };
  if (request.n !== undefined && request.n !== 1) return { ok: false, message: "The playground currently supports exactly one image per request." };
  if (request.response_format !== undefined && request.response_format !== "url" && request.response_format !== "b64_json") {
    return { ok: false, message: "response_format must be url or b64_json." };
  }
  if (request.size !== undefined && (typeof request.size !== "string" || !/^(auto|[1-9][0-9]{1,3}x[1-9][0-9]{1,3})$/.test(request.size))) {
    return { ok: false, message: "size is invalid." };
  }
  if (request.quality !== undefined && (typeof request.quality !== "string" || !["auto", "low", "medium", "high", "standard", "hd"].includes(request.quality))) {
    return { ok: false, message: "quality is invalid." };
  }
  if (request.output_format !== undefined && (typeof request.output_format !== "string" || !["png", "jpeg", "webp"].includes(request.output_format))) {
    return { ok: false, message: "output_format is invalid." };
  }
  if (request.output_compression !== undefined && (typeof request.output_compression !== "number" || !Number.isInteger(request.output_compression) || request.output_compression < 0 || request.output_compression > 100)) {
    return { ok: false, message: "output_compression must be an integer between 0 and 100." };
  }
  if (request.background !== undefined && (typeof request.background !== "string" || !["auto", "transparent", "opaque"].includes(request.background))) {
    return { ok: false, message: "background is invalid." };
  }
  if (request.moderation !== undefined && (typeof request.moderation !== "string" || !["auto", "low"].includes(request.moderation))) {
    return { ok: false, message: "moderation is invalid." };
  }
  if (body.godarai !== undefined) return { ok: false, message: "GodarAI chat cache and guardrails are not supported in image generation mode." };
  return { ok: true };
}

function validateRequestParameters(request: Record<string, unknown>): string | null {
  if (request.stream !== undefined && typeof request.stream !== "boolean") return "stream must be a boolean.";
  if (request.stream_options !== undefined) {
    if (!isRecord(request.stream_options)) return "stream_options must be an object.";
    const unsupported = Object.keys(request.stream_options).filter((key) => key !== "include_usage");
    if (unsupported.length) return `Unsupported stream_options fields: ${unsupported.join(", ")}`;
    if (request.stream_options.include_usage !== undefined && typeof request.stream_options.include_usage !== "boolean") return "stream_options.include_usage must be a boolean.";
  }
  for (const name of ["max_tokens", "max_output_tokens"] as const) {
    const value = request[name];
    if (value !== undefined && (typeof value !== "number" || !Number.isInteger(value) || value < 1)) return `${name} must be a positive integer.`;
  }
  if (request.parallel_tool_calls !== undefined && typeof request.parallel_tool_calls !== "boolean") return "parallel_tool_calls must be a boolean.";
  if (request.reasoning_effort !== undefined && (typeof request.reasoning_effort !== "string" || request.reasoning_effort.length > 32)) return "reasoning_effort is invalid.";
  if (request.response_format !== undefined && !isRecord(request.response_format)) return "response_format must be an object.";
  if (request.thinking !== undefined && !isRecord(request.thinking)) return "thinking must be an object.";
  if (request.tool_choice !== undefined && typeof request.tool_choice !== "string" && !isRecord(request.tool_choice)) return "tool_choice is invalid.";
  return null;
}

function validateTool(value: unknown, index: number): string | null {
  if (!isRecord(value) || value.type !== "function" || !isRecord(value.function)) return `tools[${index}] is invalid.`;
  if (Object.keys(value).some((key) => key !== "type" && key !== "function")) return `tools[${index}] contains unsupported fields.`;
  if (Object.keys(value.function).some((key) => !["name", "description", "parameters"].includes(key))) return `tools[${index}].function contains unsupported fields.`;
  if (typeof value.function.name !== "string" || !/^[A-Za-z0-9_-]{1,64}$/.test(value.function.name)) return `tools[${index}].function.name is invalid.`;
  if (value.function.description !== undefined && (typeof value.function.description !== "string" || value.function.description.length > 4_000)) return `tools[${index}].function.description is invalid.`;
  if (!isRecord(value.function.parameters)) return `tools[${index}].function.parameters must be an object.`;
  return null;
}

function validateGodaraiOptions(value: unknown): string | null {
  if (value === undefined) return null;
  if (!isRecord(value)) return "godarai must be an object.";
  const unsupported = Object.keys(value).filter((key) => !GODARAI_FIELDS.has(key));
  if (unsupported.length) return `Unsupported godarai fields: ${unsupported.join(", ")}`;
  if (value.observability !== undefined && value.observability !== null) {
    if (!isRecord(value.observability)) return "observability must be an object.";
    const observabilityUnsupported = Object.keys(value.observability).filter((key) => !OBSERVABILITY_FIELDS.has(key));
    if (observabilityUnsupported.length) return `Unsupported observability fields: ${observabilityUnsupported.join(", ")}`;
    const metadataError = validateMetadata(value.observability.tags);
    if (metadataError) return metadataError;
  }
  if (value.cache_config !== undefined && value.cache_config !== null) {
    if (!isRecord(value.cache_config)) return "cache_config must be an object.";
    const cacheUnsupported = Object.keys(value.cache_config).filter((key) => !CACHE_FIELDS.has(key));
    if (cacheUnsupported.length) return `Unsupported cache_config fields: ${cacheUnsupported.join(", ")}`;
    if (value.cache_config.type !== "simple" && value.cache_config.type !== "smart") return "cache_config.type is invalid.";
    if (typeof value.cache_config.ttl !== "number" || !Number.isInteger(value.cache_config.ttl) || value.cache_config.ttl < 1 || value.cache_config.ttl > 2_592_000) return "cache_config.ttl is invalid.";
    if (value.cache_config.namespace !== undefined && (typeof value.cache_config.namespace !== "string" || value.cache_config.namespace.length > 128)) return "cache_config.namespace is invalid.";
    if (value.cache_config.similarity_threshold !== undefined && !optionalFiniteNumber(value.cache_config.similarity_threshold, 0.8, 1)) return "cache_config.similarity_threshold is invalid.";
  }
  if (value.guardrails !== undefined && value.guardrails !== null) {
    if (!isRecord(value.guardrails)) return "guardrails must be an object.";
    const guardrailUnsupported = Object.keys(value.guardrails).filter((key) => !GUARDRAIL_FIELDS.has(key));
    if (guardrailUnsupported.length) return `Unsupported guardrail fields: ${guardrailUnsupported.join(", ")}`;
    for (const name of ["input", "output"] as const) {
      const entries = value.guardrails[name];
      if (entries !== undefined && (!Array.isArray(entries) || entries.length > 100 || entries.some((item) => typeof item !== "string" || !item.trim() || item.length > 256))) return `guardrails.${name} is invalid.`;
    }
    if (value.guardrails.scope !== undefined && value.guardrails.scope !== "all" && value.guardrails.scope !== "last" && value.guardrails.scope !== "") return "guardrails.scope is invalid.";
  }
  return null;
}

function validateMetadata(value: unknown): string | null {
  if (value === undefined) return null;
  if (!isRecord(value)) return "metadata must be an object.";
  const entries = Object.entries(value);
  if (entries.length > 20) return "metadata contains too many entries.";
  for (const [key, entry] of entries) {
    const normalizedKey = key.toLowerCase();
    if (!/^[A-Za-z0-9_.-]{1,64}$/.test(key) || normalizedKey.includes("authorization") || RESERVED_METADATA_KEYS.has(normalizedKey)) return `metadata key ${key} is not allowed.`;
    if (typeof entry === "string" && entry.length <= 256) continue;
    if (typeof entry === "number" && Number.isFinite(entry)) continue;
    if (typeof entry === "boolean") continue;
    return `metadata value for ${key} is invalid.`;
  }
  return null;
}

function optionalFiniteNumber(value: unknown, minimum: number, maximum: number): boolean {
  return value === undefined || (typeof value === "number" && Number.isFinite(value) && value >= minimum && value <= maximum);
}

function isMessageRole(value: unknown): value is "system" | "user" | "assistant" | "tool" {
  return value === "system" || value === "user" || value === "assistant" || value === "tool";
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return Boolean(value) && typeof value === "object" && !Array.isArray(value);
}

function normalizeStringList(value: unknown): string[] {
  return Array.isArray(value) ? value.filter((item): item is string => typeof item === "string" && item.trim() !== "").slice(0, 100) : [];
}

function buildObservabilityContext(envelope: PlaygroundRunPayload) {
  return {
    source: "tenant_admin_playground",
    ...(typeof envelope.client_run_id === "string" ? { client_run_id: envelope.client_run_id } : {}),
    ...(typeof envelope.comparison_id === "string" ? { comparison_id: envelope.comparison_id } : {}),
    ...(typeof envelope.column_id === "string" ? { column_id: envelope.column_id } : {}),
    tags: envelope.godarai?.observability?.tags ?? {},
  };
}
