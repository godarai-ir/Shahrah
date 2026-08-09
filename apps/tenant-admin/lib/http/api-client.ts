import type { ZodType } from "zod";

type APIErrorPayload = {
  error?: {
    code?: unknown;
    message?: unknown;
  };
  code?: unknown;
  message?: unknown;
};

export class APIClientError extends Error {
  readonly status: number;
  readonly code?: string;
  readonly payload: unknown;

  constructor(message: string, options: { status: number; code?: string; payload: unknown }) {
    super(message);
    this.name = "APIClientError";
    this.status = options.status;
    this.code = options.code;
    this.payload = options.payload;
  }
}

export class APIContractError extends Error {
  readonly endpoint: string;
  readonly issues: readonly unknown[];

  constructor(endpoint: string, issues: readonly unknown[]) {
    super(`Unexpected response contract from ${endpoint}`);
    this.name = "APIContractError";
    this.endpoint = endpoint;
    this.issues = issues;
  }
}

export async function requestJSON(path: string, options?: RequestInit): Promise<unknown> {
  const headers = new Headers(options?.headers ?? {});
  if (typeof options?.body === "string" && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }

  const response = await fetch(path, { ...options, headers, cache: "no-store" });
  const payload = await readPayload(response);

  if (!response.ok) {
    const errorPayload = asErrorPayload(payload);
    const message = readString(errorPayload?.error?.message)
      ?? readString(errorPayload?.message)
      ?? `Request failed (${response.status})`;
    const code = readString(errorPayload?.error?.code) ?? readString(errorPayload?.code);
    throw new APIClientError(message, { status: response.status, code, payload });
  }

  return payload;
}

export async function requestParsedJSON<T>(
  schema: ZodType<T>,
  path: string,
  options?: RequestInit,
): Promise<T> {
  const payload = await requestJSON(path, options);
  const parsed = schema.safeParse(payload);
  if (!parsed.success) throw new APIContractError(path, parsed.error.issues);
  return parsed.data;
}

export function getAPIErrorMessage(error: unknown, fallback: string): string {
  return error instanceof Error && error.message.trim() ? error.message : fallback;
}

async function readPayload(response: Response): Promise<unknown> {
  const text = await response.text();
  if (!text.trim()) return {};
  try {
    return JSON.parse(text) as unknown;
  } catch {
    return text;
  }
}

function asErrorPayload(payload: unknown): APIErrorPayload | null {
  if (!payload || typeof payload !== "object") return null;
  return payload as APIErrorPayload;
}

function readString(value: unknown): string | undefined {
  return typeof value === "string" && value.trim() ? value.trim() : undefined;
}
