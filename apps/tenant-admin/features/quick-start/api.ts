import { z } from "zod";

import { providersResponseSchema } from "@/lib/contracts/provider-catalog";
import {
  APIClientError,
  APIContractError,
  requestParsedJSON,
} from "@/lib/http/api-client";

const personalAccessTokenSchema = z.object({ id: z.string(), key: z.string() });
const sampleCompletionSchema = z.object({
  choices: z.array(z.object({
    message: z.object({ content: z.unknown().optional() }).optional(),
  })).optional(),
}).passthrough();
const quickStartSubscriptionSchema = z.object({
  plan_slug: z.string().optional(), plan_display_name: z.string().optional(), status: z.string().optional(),
  billing_interval: z.string().optional(),
}).nullable();
const quickStartEntitlementsSchema = z.object({ features: z.record(z.string(), z.unknown()).optional() }).nullable();
const resourceListSchema = z.object({ data: z.array(z.unknown()).optional(), total: z.number().optional() });
const allowedModelsSchema = z.object({ data: z.array(z.object({ id: z.string() })).optional() });

export function listQuickStartProviders() {
  return requestParsedJSON(providersResponseSchema, "/api/provider-catalog/providers");
}

export function createPersonalAccessToken(body: { name: string; expiresIn?: string }) {
  return requestParsedJSON(personalAccessTokenSchema, "/api/personal-access-tokens", {
    method: "POST",
    body: JSON.stringify({ name: body.name, expires_in: body.expiresIn }),
  });
}

export async function runQuickStartSample(input: {
  baseUrl: string;
  token: string;
  model: string;
  prompt: string;
}) {
  const endpoint = `${input.baseUrl}/v1/chat/completions`;
  const response = await fetch(endpoint, {
    method: "POST",
    headers: {
      Authorization: `Bearer ${input.token}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      model: input.model,
      messages: [{ role: "user", content: input.prompt }],
    }),
  });
  const payload = parseResponsePayload(await response.text());
  if (!response.ok) {
    throw new APIClientError(readErrorMessage(payload) ?? `Request failed (${response.status})`, {
      status: response.status,
      payload,
    });
  }
  const parsed = sampleCompletionSchema.safeParse(payload);
  if (!parsed.success) throw new APIContractError(endpoint, parsed.error.issues);

  return { payload: parsed.data, status: response.status };
}

function parseResponsePayload(text: string): unknown {
  if (!text.trim()) return {};
  try {
    return JSON.parse(text) as unknown;
  } catch {
    return text;
  }
}

function readErrorMessage(payload: unknown): string | null {
  if (typeof payload === "string") return payload;
  if (!payload || typeof payload !== "object") return null;
  const candidate = payload as { error?: { message?: unknown }; message?: unknown };
  if (typeof candidate.error?.message === "string") return candidate.error.message;
  return typeof candidate.message === "string" ? candidate.message : null;
}

export function parseQuickStartSubscription(payload: unknown) {
  const parsed = quickStartSubscriptionSchema.safeParse(payload);
  return parsed.success ? parsed.data : null;
}

export function parseQuickStartEntitlements(payload: unknown) {
  const parsed = quickStartEntitlementsSchema.safeParse(payload);
  return parsed.success ? parsed.data : null;
}

export function countQuickStartResources(payload: unknown): number {
  const parsed = resourceListSchema.safeParse(payload);
  if (!parsed.success) return 0;
  return parsed.data.total ?? parsed.data.data?.length ?? 0;
}

export function parseAllowedModelIDs(payload: unknown): string[] {
  const parsed = allowedModelsSchema.safeParse(payload);
  if (!parsed.success) return [];
  return (parsed.data.data ?? []).map((item) => item.id.trim()).filter(Boolean);
}
