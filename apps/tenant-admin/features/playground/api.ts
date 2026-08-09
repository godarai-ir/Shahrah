import { z } from "zod";

import {
  promptListResponseSchema,
  promptVersionEnvelopeSchema,
  promptVersionListResponseSchema,
  type PromptManifest,
} from "@/lib/contracts/prompts";
import { requestParsedJSON } from "@/lib/http/api-client";
import { APIContractError } from "@/lib/http/api-client";

const gatewayModelSchema = z.object({
  id: z.string(),
  object: z.string(),
  owned_by: z.string(),
  provider: z.string().optional(),
  aliases: z.array(z.string()).optional(),
  supported_modes: z.array(z.string()).optional(),
  types: z.array(z.string()).optional(),
});

const gatewayModelListSchema = z.object({ data: z.array(gatewayModelSchema).optional() });

const guardrailSchema = z.object({
  id: z.string(),
  slug: z.string(),
  name: z.string(),
  description: z.string().optional(),
  selector: z.string().optional(),
  type: z.string().optional(),
});

const guardrailListSchema = z.object({ data: z.array(guardrailSchema).optional() });
const playgroundEntitlementsSchema = z.object({
  features: z.record(z.string(), z.unknown()).optional(),
}).nullable();
const playgroundResponseSchema = z.record(z.string(), z.unknown());
const playgroundErrorSchema = z.object({
  error: z.object({ message: z.string().optional() }).optional(),
  message: z.string().optional(),
});

export function listGatewayModels() {
  return requestParsedJSON(gatewayModelListSchema, "/api/models");
}

export function listGuardrails() {
  return requestParsedJSON(guardrailListSchema, "/api/guardrails?limit=200");
}

export function getPlaygroundEntitlements() {
  return requestParsedJSON(playgroundEntitlementsSchema, "/api/entitlements");
}

export async function getPromptVersion(id: string) {
  const payload = await requestParsedJSON(
    promptVersionEnvelopeSchema,
    `/api/prompt-versions/${encodeURIComponent(id)}`,
  );
  return "data" in payload ? payload.data : payload;
}

export function listPlaygroundPrompts(limit: number, signal?: AbortSignal) {
  return requestParsedJSON(promptListResponseSchema, `/api/prompts?limit=${limit}`, { signal });
}

export function listPlaygroundPromptVersions(id: string, signal?: AbortSignal) {
  return requestParsedJSON(
    promptVersionListResponseSchema,
    `/api/prompts/${encodeURIComponent(id)}/versions?limit=100`,
    { signal },
  );
}

export async function savePlaygroundPrompt(input: {
  promptID?: string;
  manifest: PromptManifest;
  commitMessage: string;
}) {
  const path = input.promptID
    ? `/api/prompts/${encodeURIComponent(input.promptID)}/versions`
    : "/api/prompts";
  const payload = await requestParsedJSON(promptVersionEnvelopeSchema, path, {
    method: "POST",
    body: JSON.stringify({ manifest: input.manifest, commit_message: input.commitMessage }),
  });
  return "data" in payload ? payload.data : payload;
}

export async function runPlaygroundRequest(
  payload: unknown,
  signal: AbortSignal,
  fallbackError: string,
): Promise<Response> {
  const response = await fetch("/api/playground/run", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
    signal,
  });
  if (response.ok) return response;

  const errorPayload = playgroundErrorSchema.safeParse(await response.json().catch(() => null));
  const message = errorPayload.success
    ? errorPayload.data.error?.message ?? errorPayload.data.message
    : undefined;
  throw new Error(message || `${fallbackError} (${response.status})`);
}

export async function readPlaygroundJSON(response: Response) {
  const payload = await response.json() as unknown;
  const parsed = playgroundResponseSchema.safeParse(payload);
  if (!parsed.success) {
    throw new APIContractError("/api/playground/run", parsed.error.issues);
  }
  return parsed.data;
}
