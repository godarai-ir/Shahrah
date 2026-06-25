import type { GuardrailPayload, TenantGuardrail } from "./types";

type ListResponse = {
  data: TenantGuardrail[];
  total: number;
  limit: number;
  offset: number;
};

export class GuardrailAPIError extends Error {
  status: number;

  constructor(message: string, status: number) {
    super(message);
    this.name = "GuardrailAPIError";
    this.status = status;
  }
}

async function requestJSON<T>(path: string, options?: RequestInit): Promise<T> {
  const headers = new Headers(options?.headers ?? {});
  if (options?.body && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }

  const response = await fetch(path, { ...options, headers, cache: "no-store" });
  const text = await response.text();
  let payload: Record<string, unknown> = {};
  if (text.trim()) {
    try {
      payload = JSON.parse(text) as Record<string, unknown>;
    } catch {
      payload = { message: text };
    }
  }

  if (!response.ok) {
    const rawMessage =
      (payload?.error as { message?: string } | undefined)?.message ??
      (payload?.message as string | undefined) ??
      `Request failed (${response.status})`;
    throw new GuardrailAPIError(rawMessage, response.status);
  }

  return payload as T;
}

export function listGuardrails(query = "") {
  return requestJSON<ListResponse>(`/api/guardrails?query=${encodeURIComponent(query)}`);
}

export function createGuardrail(body: GuardrailPayload) {
  return requestJSON<TenantGuardrail>("/api/guardrails", {
    method: "POST",
    body: JSON.stringify(body),
  });
}

export function updateGuardrail(slug: string, body: GuardrailPayload) {
  return requestJSON<TenantGuardrail>(`/api/guardrails/${encodeURIComponent(slug)}`, {
    method: "PUT",
    body: JSON.stringify(body),
  });
}

export function deleteGuardrail(slug: string) {
  return requestJSON<{ success: true }>(`/api/guardrails/${encodeURIComponent(slug)}`, {
    method: "DELETE",
  });
}
