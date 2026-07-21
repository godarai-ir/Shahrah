import { NextRequest, NextResponse } from "next/server";

import { getCurrentMePermissions, getNormalizedPermissionSet } from "@/lib/auth/me-permissions";
import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { proxyToControlPlane, proxyToControlPlaneStream } from "@/lib/control-plane";

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

const REQUEST_FIELDS = new Set([
  "model",
  "messages",
  "stream",
  "temperature",
  "top_p",
  "max_tokens",
  "max_output_tokens",
  "stop",
  "reasoning_effort",
  "thinking",
  "response_format",
  "tools",
  "tool_choice",
  "parallel_tool_calls",
  "metadata",
]);
const ENVELOPE_FIELDS = new Set(["client_run_id", "comparison_id", "column_id", "request", "godarai"]);
const MESSAGE_FIELDS = new Set(["role", "content", "tool_call_id"]);
const CACHE_FIELDS = new Set(["type", "ttl", "namespace", "similarity_threshold"]);
const GODARAI_FIELDS = new Set(["cache_config", "guardrails"]);
const GUARDRAIL_FIELDS = new Set(["input", "output", "scope"]);
const RESERVED_METADATA_KEYS = new Set(["source", "client_run_id", "comparison_id", "column_id", "__proto__", "prototype", "constructor"]);

type PlaygroundRunPayload = {
  client_run_id?: string;
  comparison_id?: string;
  column_id?: string;
  request?: Record<string, unknown>;
  godarai?: {
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

  const upstreamHeaders = new Headers();
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

  const upstreamRequest: Record<string, unknown> = {
    ...(body!.request as Record<string, unknown>),
    metadata: sanitizeMetadata(body!.request!.metadata, body!),
  };
  const proxyOptions = {
    request,
    method: "POST" as const,
    path: "/v1/chat/completions",
    body: upstreamRequest,
    userID: auth.user.id,
    upstreamHeaders,
    exposeResponseHeaders: OPERATIONAL_HEADERS,
  };

  return upstreamRequest["stream"] === true
    ? proxyToControlPlaneStream(proxyOptions)
    : proxyToControlPlane(proxyOptions);
}

function validatePayload(body: PlaygroundRunPayload | null): { ok: true } | { ok: false; message: string } {
  if (!body || typeof body !== "object" || !body.request || typeof body.request !== "object" || Array.isArray(body.request)) {
    return { ok: false, message: "Invalid playground payload." };
  }
  const unknownEnvelopeFields = Object.keys(body).filter((key) => !ENVELOPE_FIELDS.has(key));
  if (unknownEnvelopeFields.length > 0) return { ok: false, message: `Unsupported playground fields: ${unknownEnvelopeFields.join(", ")}` };
  for (const [name, value] of [["client_run_id", body.client_run_id], ["comparison_id", body.comparison_id], ["column_id", body.column_id]] as const) {
    if (value !== undefined && (typeof value !== "string" || !/^[A-Za-z0-9_-]{1,100}$/.test(value))) return { ok: false, message: `${name} is invalid.` };
  }
  const unknownFields = Object.keys(body.request).filter((key) => !REQUEST_FIELDS.has(key));
  if (unknownFields.length > 0) {
    return { ok: false, message: `Unsupported request fields: ${unknownFields.join(", ")}` };
  }
  if (typeof body.request.model !== "string" || !body.request.model.trim() || body.request.model.length > 512) {
    return { ok: false, message: "A model is required." };
  }
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
  const metadataError = validateMetadata(body.request.metadata);
  if (metadataError) return { ok: false, message: metadataError };
  const godaraiError = validateGodaraiOptions(body.godarai);
  if (godaraiError) return { ok: false, message: godaraiError };
  return { ok: true };
}

function validateRequestParameters(request: Record<string, unknown>): string | null {
  if (request.stream !== undefined && typeof request.stream !== "boolean") return "stream must be a boolean.";
  if (!optionalFiniteNumber(request.temperature, 0, 2)) return "temperature must be between 0 and 2.";
  if (!optionalFiniteNumber(request.top_p, 0, 1)) return "top_p must be between 0 and 1.";
  for (const name of ["max_tokens", "max_output_tokens"] as const) {
    const value = request[name];
    if (value !== undefined && (typeof value !== "number" || !Number.isInteger(value) || value < 1)) return `${name} must be a positive integer.`;
  }
  if (request.parallel_tool_calls !== undefined && typeof request.parallel_tool_calls !== "boolean") return "parallel_tool_calls must be a boolean.";
  if (request.reasoning_effort !== undefined && (typeof request.reasoning_effort !== "string" || request.reasoning_effort.length > 32)) return "reasoning_effort is invalid.";
  if (request.stop !== undefined && ((typeof request.stop === "string" && request.stop.length > 1_000) || (typeof request.stop !== "string" && (!Array.isArray(request.stop) || request.stop.length > 20 || request.stop.some((item) => typeof item !== "string" || item.length > 1_000))))) return "stop must be a string or a bounded string array.";
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

function sanitizeMetadata(value: unknown, envelope: PlaygroundRunPayload): Record<string, string | number | boolean> {
  const output: Record<string, string | number | boolean> = { source: "tenant_admin_playground" };
  if (value && typeof value === "object" && !Array.isArray(value)) {
    for (const [key, entry] of Object.entries(value).slice(0, 20)) {
      const normalizedKey = key.toLowerCase();
      if (!/^[A-Za-z0-9_.-]{1,64}$/.test(key) || normalizedKey.includes("authorization") || RESERVED_METADATA_KEYS.has(normalizedKey)) continue;
      if (typeof entry === "string") output[key] = entry.slice(0, 256);
      if (typeof entry === "number" && Number.isFinite(entry)) output[key] = entry;
      if (typeof entry === "boolean") output[key] = entry;
    }
  }
  if (typeof envelope.client_run_id === "string") output.client_run_id = envelope.client_run_id.slice(0, 100);
  if (typeof envelope.comparison_id === "string") output.comparison_id = envelope.comparison_id.slice(0, 100);
  if (typeof envelope.column_id === "string") output.column_id = envelope.column_id.slice(0, 100);
  return output;
}
