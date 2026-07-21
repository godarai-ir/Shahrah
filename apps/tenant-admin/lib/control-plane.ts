import { NextRequest, NextResponse } from "next/server";

import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { assertTenantMatch, ensureTenantFromHeaders } from "@/lib/tenant";

type ProxyOptions = {
  request: NextRequest;
  method: "GET" | "POST" | "PUT" | "PATCH" | "DELETE";
  path: string;
  body?: unknown;
  userID?: string;
  tenantSlug?: string;
  upstreamHeaders?: HeadersInit;
  exposeResponseHeaders?: string[];
};

type ControlPlaneRequest = {
  tenantSlug: string;
  method: "GET" | "POST" | "PUT" | "PATCH" | "DELETE";
  path: string;
  body?: unknown;
  userID?: string;
  requestHeaders?: Headers;
  upstreamHeaders?: HeadersInit;
  signal?: AbortSignal;
};

type APIErrorPayload = {
  error?: {
    message?: string;
  };
  message?: string;
};

const CONTROL_PLANE_USER_ID_HEADER = "x-godarai-user-id";
const LEGACY_CONTROL_PLANE_USER_ID_HEADER = "x-shahrah-user-id";

export type ControlPlaneResponse = {
  ok: boolean;
  status: number;
  statusCode?: number;
  payload: unknown;
  errorMessage?: string;
  responseHeaders: Record<string, string>;
};

const SAFE_OPERATIONAL_RESPONSE_HEADERS = new Set([
  "x-godarai-trace-id",
  "x-godarai-cache-status",
  "x-godarai-cached-trace-id",
  "x-godarai-cache-similarity-score",
  "x-godarai-cost-usd",
  "x-godarai-cost-irr",
  "x-request-id",
]);

type ControlPlaneBinaryResponse = {
  ok: boolean;
  status: number;
  body: ArrayBuffer;
  contentType?: string;
  cacheControl?: string;
  contentDisposition?: string;
  errorMessage?: string;
};

export async function proxyToControlPlane(options: ProxyOptions) {
  const tenantResolution = ensureTenantFromHeaders(options.request.headers, options.tenantSlug);
  if (!tenantResolution.ok) {
    return NextResponse.json(
      { error: { message: tenantResolution.message } },
      { status: tenantResolution.status },
    );
  }

  const auth = await readAuthFromRequest(options.request);
  if (!auth.ok) {
    return NextResponse.json(
      { error: { message: auth.message } },
      { status: auth.status },
    );
  }

  const tenantSlug = tenantResolution.tenantSlug;
  if (!assertTenantMatch(tenantSlug, auth.tenantSlug)) {
    return NextResponse.json(
      { error: { message: "Tenant mismatch between host and authenticated session." } },
      { status: 403 },
    );
  }

  const requestedUserID = options.userID?.trim() ?? "";
  if (requestedUserID && requestedUserID !== auth.user.id) {
    return NextResponse.json(
      { error: { message: "Authenticated user does not match the requested user context." } },
      { status: 403 },
    );
  }

  const upstream = await fetchControlPlaneJSON({
    tenantSlug,
    method: options.method,
    path: options.path,
    body: options.body,
    userID: auth.user.id,
    requestHeaders: options.request.headers,
    upstreamHeaders: options.upstreamHeaders,
  });

  return controlPlaneJSONResponse(upstream, options.exposeResponseHeaders);
}

/**
 * Relays a streaming control-plane response without buffering it. Only the
 * explicitly requested operational headers can be exposed to the browser.
 */
export async function proxyToControlPlaneStream(options: ProxyOptions) {
  const tenantResolution = ensureTenantFromHeaders(options.request.headers, options.tenantSlug);
  if (!tenantResolution.ok) {
    return NextResponse.json({ error: { message: tenantResolution.message } }, { status: tenantResolution.status });
  }

  const auth = await readAuthFromRequest(options.request);
  if (!auth.ok) {
    return NextResponse.json({ error: { message: auth.message } }, { status: auth.status });
  }

  if (!assertTenantMatch(tenantResolution.tenantSlug, auth.tenantSlug)) {
    return NextResponse.json(
      { error: { message: "Tenant mismatch between host and authenticated session." } },
      { status: 403 },
    );
  }

  const requestedUserID = options.userID?.trim() ?? "";
  if (requestedUserID && requestedUserID !== auth.user.id) {
    return NextResponse.json(
      { error: { message: "Authenticated user does not match the requested user context." } },
      { status: 403 },
    );
  }

  const baseURL = getControlPlaneBaseURL();
  const masterKey = process.env.SHAHRAH_MASTER_KEY ?? process.env.MASTER_KEY;
  if (!masterKey) {
    return NextResponse.json(
      { error: { message: "برای routeهای tenant-admin باید SHAHRAH_MASTER_KEY یا MASTER_KEY تنظیم شود." } },
      { status: 500 },
    );
  }

  const upstreamHeaders = buildControlPlaneHeaders({
    tenantSlug: tenantResolution.tenantSlug,
    userID: auth.user.id,
    requestHeaders: options.request.headers,
    upstreamHeaders: options.upstreamHeaders,
    masterKey,
    accept: "text/event-stream",
    hasBody: options.body !== undefined,
  });

  try {
    const upstream = await fetch(new URL(normalizePath(options.path), baseURL), {
      method: options.method,
      headers: upstreamHeaders,
      body: options.body !== undefined ? JSON.stringify(options.body) : undefined,
      cache: "no-store",
      signal: options.request.signal,
    });

    const responseHeaders = new Headers({
      "Content-Type": upstream.headers.get("content-type") || "text/event-stream; charset=utf-8",
      "Cache-Control": "no-cache, no-transform",
      "X-Accel-Buffering": "no",
    });
    copySafeOperationalHeaders(upstream.headers, responseHeaders, options.exposeResponseHeaders);

    return new NextResponse(upstream.body, { status: upstream.status, headers: responseHeaders });
  } catch (error) {
    if (options.request.signal.aborted) {
      return new NextResponse(null, { status: 499 });
    }
    const message = error instanceof Error ? error.message : "اتصال به API مدیریت control-plane برقرار نشد.";
    return NextResponse.json({ error: { message } }, { status: 502 });
  }
}

export async function proxyToControlPlaneBinary(options: ProxyOptions) {
  const tenantResolution = ensureTenantFromHeaders(options.request.headers, options.tenantSlug);
  if (!tenantResolution.ok) {
    return NextResponse.json(
      { error: { message: tenantResolution.message } },
      { status: tenantResolution.status },
    );
  }

  const auth = await readAuthFromRequest(options.request);
  if (!auth.ok) {
    return NextResponse.json(
      { error: { message: auth.message } },
      { status: auth.status },
    );
  }

  const tenantSlug = tenantResolution.tenantSlug;
  if (!assertTenantMatch(tenantSlug, auth.tenantSlug)) {
    return NextResponse.json(
      { error: { message: "Tenant mismatch between host and authenticated session." } },
      { status: 403 },
    );
  }

  const requestedUserID = options.userID?.trim() ?? "";
  if (requestedUserID && requestedUserID !== auth.user.id) {
    return NextResponse.json(
      { error: { message: "Authenticated user does not match the requested user context." } },
      { status: 403 },
    );
  }

  const upstream = await fetchControlPlaneRaw({
    tenantSlug,
    method: options.method,
    path: options.path,
    body: options.body,
    userID: auth.user.id,
    requestHeaders: options.request.headers,
    upstreamHeaders: options.upstreamHeaders,
  });

  if (!upstream.ok) {
    return new NextResponse(
      upstream.errorMessage ?? `Upstream asset request failed (${upstream.status})`,
      {
        status: upstream.status,
        headers: {
          "Content-Type": "text/plain; charset=utf-8",
        },
      },
    );
  }

  return new NextResponse(upstream.body, {
    status: upstream.status,
    headers: {
      ...(upstream.contentType ? { "Content-Type": upstream.contentType } : {}),
      ...(upstream.cacheControl ? { "Cache-Control": upstream.cacheControl } : {}),
      ...(upstream.contentDisposition ? { "Content-Disposition": upstream.contentDisposition } : {}),
    },
  });
}

export async function proxyPublicInvitationAcceptance(
  request: NextRequest,
  token: string,
) {
  const tenantResolution = ensureTenantFromHeaders(request.headers);
  if (!tenantResolution.ok) {
    return NextResponse.json(
      { error: { message: tenantResolution.message } },
      { status: tenantResolution.status },
    );
  }

  const upstream = await fetchControlPlaneJSON({
    tenantSlug: tenantResolution.tenantSlug,
    method: "POST",
    path: "/api/v1/users/invitations/accept",
    body: { token },
    requestHeaders: request.headers,
  });

  return controlPlaneJSONResponse(upstream);
}

export async function fetchControlPlaneJSON(
  options: ControlPlaneRequest,
): Promise<ControlPlaneResponse> {
  const baseURL = getControlPlaneBaseURL();
  const masterKey = process.env.SHAHRAH_MASTER_KEY ?? process.env.MASTER_KEY;

  if (!masterKey) {
    return {
      ok: false,
      status: 500,
      payload: {
        error: {
          message:
            "برای routeهای tenant-admin باید SHAHRAH_MASTER_KEY یا MASTER_KEY تنظیم شود.",
        },
      },
      errorMessage:
        "برای routeهای tenant-admin باید SHAHRAH_MASTER_KEY یا MASTER_KEY تنظیم شود.",
      responseHeaders: {},
    };
  }

  const url = new URL(normalizePath(options.path), baseURL).toString();
  const forwardedHost = resolveForwardedHost(options.requestHeaders);
  const forwardedProto = resolveForwardedProto(options.requestHeaders);
  const upstreamHeaders = new Headers(options.upstreamHeaders ?? {});
  upstreamHeaders.set("Authorization", `Bearer ${masterKey}`);
  upstreamHeaders.set("Accept", "application/json");
  upstreamHeaders.set("Content-Type", "application/json");
  upstreamHeaders.set("x-godarai-tenant", options.tenantSlug);
  upstreamHeaders.delete(LEGACY_CONTROL_PLANE_USER_ID_HEADER);
  if (options.userID) {
    upstreamHeaders.set(CONTROL_PLANE_USER_ID_HEADER, options.userID);
  } else {
    upstreamHeaders.delete(CONTROL_PLANE_USER_ID_HEADER);
  }
  if (forwardedHost) {
    upstreamHeaders.set("x-forwarded-host", forwardedHost);
  }
  if (forwardedProto) {
    upstreamHeaders.set("x-forwarded-proto", forwardedProto);
  }

  try {
    const upstream = await fetch(url, {
      method: options.method,
      headers: upstreamHeaders,
      body: options.body ? JSON.stringify(options.body) : undefined,
      cache: "no-store",
      signal: options.signal,
    });

    const responseText = await upstream.text();
    const contentType = upstream.headers.get("content-type") ?? "";
    const payload = parsePayload(responseText, contentType);

    return {
      ok: upstream.ok,
      status: upstream.status,
      statusCode: upstream.status,
      payload,
      errorMessage: upstream.ok
        ? undefined
        : extractErrorMessage(payload) ?? `خطای سرویس بالادستی: ${upstream.status}`,
      responseHeaders: collectSafeOperationalHeaders(upstream.headers),
    };
  } catch (error) {
    const message =
      error instanceof Error
        ? error.message
        : "اتصال به API مدیریت control-plane برقرار نشد.";
    return {
      ok: false,
      status: 502,
      payload: {
        error: {
          message,
        },
      },
      errorMessage: message,
      responseHeaders: {},
    };
  }
}

async function fetchControlPlaneRaw(
  options: ControlPlaneRequest,
): Promise<ControlPlaneBinaryResponse> {
  const baseURL = getControlPlaneBaseURL();
  const masterKey = process.env.SHAHRAH_MASTER_KEY ?? process.env.MASTER_KEY;

  if (!masterKey) {
    const message =
      "Ø¨Ø±Ø§ÛŒ routeÙ‡Ø§ÛŒ tenant-admin Ø¨Ø§ÛŒØ¯ SHAHRAH_MASTER_KEY ÛŒØ§ MASTER_KEY ØªÙ†Ø¸ÛŒÙ… Ø´ÙˆØ¯.";
    return {
      ok: false,
      status: 500,
      body: new TextEncoder().encode(message).buffer as ArrayBuffer,
      contentType: "text/plain; charset=utf-8",
      errorMessage: message,
    };
  }

  const url = new URL(normalizePath(options.path), baseURL).toString();
  const forwardedHost = resolveForwardedHost(options.requestHeaders);
  const forwardedProto = resolveForwardedProto(options.requestHeaders);
  const upstreamHeaders = new Headers(options.upstreamHeaders ?? {});
  upstreamHeaders.set("Authorization", `Bearer ${masterKey}`);
  upstreamHeaders.set("Accept", "*/*");
  upstreamHeaders.set("x-godarai-tenant", options.tenantSlug);
  if (options.body) {
    upstreamHeaders.set("Content-Type", "application/json");
  } else {
    upstreamHeaders.delete("Content-Type");
  }
  upstreamHeaders.delete(LEGACY_CONTROL_PLANE_USER_ID_HEADER);
  if (options.userID) {
    upstreamHeaders.set(CONTROL_PLANE_USER_ID_HEADER, options.userID);
  } else {
    upstreamHeaders.delete(CONTROL_PLANE_USER_ID_HEADER);
  }
  if (forwardedHost) {
    upstreamHeaders.set("x-forwarded-host", forwardedHost);
  }
  if (forwardedProto) {
    upstreamHeaders.set("x-forwarded-proto", forwardedProto);
  }

  try {
    const upstream = await fetch(url, {
      method: options.method,
      headers: upstreamHeaders,
      body: options.body ? JSON.stringify(options.body) : undefined,
      cache: "no-store",
    });

    const body = await upstream.arrayBuffer();
    const contentType = upstream.headers.get("content-type") ?? "";
    const cacheControl = upstream.headers.get("cache-control") ?? "";
    const contentDisposition = upstream.headers.get("content-disposition") ?? "";
    let errorMessage: string | undefined;

    if (!upstream.ok) {
      const text = new TextDecoder().decode(body);
      const payload = parsePayload(text, contentType);
      errorMessage =
        extractErrorMessage(payload) ?? `Ø®Ø·Ø§ÛŒ Ø³Ø±ÙˆÛŒØ³ Ø¨Ø§Ù„Ø§Ø¯Ø³ØªÛŒ: ${upstream.status}`;
    }

    return {
      ok: upstream.ok,
      status: upstream.status,
      body,
      contentType,
      cacheControl,
      contentDisposition,
      errorMessage,
    };
  } catch (error) {
    const message =
      error instanceof Error
        ? error.message
        : "Ø§ØªØµØ§Ù„ Ø¨Ù‡ API Ù…Ø¯ÛŒØ±ÛŒØª control-plane Ø¨Ø±Ù‚Ø±Ø§Ø± Ù†Ø´Ø¯.";
    return {
      ok: false,
      status: 502,
      body: new TextEncoder().encode(message).buffer as ArrayBuffer,
      contentType: "text/plain; charset=utf-8",
      errorMessage: message,
    };
  }
}

function getControlPlaneBaseURL(): string {
  const raw = process.env.SHAHRAH_CONTROL_PLANE_URL ?? "http://127.0.0.1:4000";
  return raw.endsWith("/") ? raw : `${raw}/`;
}

function normalizePath(path: string): string {
  return path.startsWith("/") ? path.slice(1) : path;
}

function parsePayload(body: string, contentType: string): unknown {
  if (!body) {
    return {};
  }

  if (contentType.includes("application/json")) {
    try {
      return JSON.parse(body) as unknown;
    } catch {
      return {
        error: {
          message: body,
        },
      };
    }
  }

  return {
    message: body,
  };
}

function extractErrorMessage(payload: unknown): string | undefined {
  if (!payload) {
    return undefined;
  }
  if (typeof payload === "string") {
    return payload;
  }
  if (typeof payload === "object") {
    const value = payload as APIErrorPayload;
    return value.error?.message ?? value.message;
  }
  return undefined;
}

function controlPlaneJSONResponse(upstream: ControlPlaneResponse, exposedHeaders?: string[]) {
  const headers = new Headers();
  copySafeOperationalHeaders(new Headers(upstream.responseHeaders), headers, exposedHeaders);
  if (!upstream.ok) {
    return NextResponse.json(
      {
        error: {
          message:
            upstream.errorMessage ??
            `خطا در ارتباط با کنترل‌پلین (${upstream.status})`,
        },
        details: upstream.payload,
      },
      { status: upstream.status, headers },
    );
  }

  if (upstream.status === 204 || upstream.status === 205 || upstream.status === 304) {
    return new NextResponse(null, { status: upstream.status, headers });
  }

  return NextResponse.json(upstream.payload, { status: upstream.status, headers });
}

function collectSafeOperationalHeaders(source: Headers): Record<string, string> {
  const result: Record<string, string> = {};
  for (const name of SAFE_OPERATIONAL_RESPONSE_HEADERS) {
    const value = source.get(name);
    if (value) result[name] = value;
  }
  return result;
}

function copySafeOperationalHeaders(source: Headers, target: Headers, requested?: string[]) {
  for (const rawName of requested ?? []) {
    const name = rawName.trim().toLowerCase();
    if (!SAFE_OPERATIONAL_RESPONSE_HEADERS.has(name)) continue;
    const value = source.get(name);
    if (value) target.set(name, value);
  }
}

function buildControlPlaneHeaders(options: {
  tenantSlug: string;
  userID?: string;
  requestHeaders?: Headers;
  upstreamHeaders?: HeadersInit;
  masterKey: string;
  accept: string;
  hasBody: boolean;
}) {
  const upstreamHeaders = new Headers(options.upstreamHeaders ?? {});
  upstreamHeaders.set("Authorization", `Bearer ${options.masterKey}`);
  upstreamHeaders.set("Accept", options.accept);
  upstreamHeaders.set("x-godarai-tenant", options.tenantSlug);
  upstreamHeaders.delete(LEGACY_CONTROL_PLANE_USER_ID_HEADER);
  if (options.hasBody) upstreamHeaders.set("Content-Type", "application/json");
  else upstreamHeaders.delete("Content-Type");
  if (options.userID) upstreamHeaders.set(CONTROL_PLANE_USER_ID_HEADER, options.userID);
  else upstreamHeaders.delete(CONTROL_PLANE_USER_ID_HEADER);
  const forwardedHost = resolveForwardedHost(options.requestHeaders);
  const forwardedProto = resolveForwardedProto(options.requestHeaders);
  if (forwardedHost) upstreamHeaders.set("x-forwarded-host", forwardedHost);
  if (forwardedProto) upstreamHeaders.set("x-forwarded-proto", forwardedProto);
  return upstreamHeaders;
}

function resolveForwardedHost(headers?: Headers): string {
  if (!headers) {
    return "";
  }
  const forwarded = firstHeaderValue(headers.get("x-forwarded-host"));
  if (forwarded) {
    return forwarded;
  }
  return firstHeaderValue(headers.get("host"));
}

function resolveForwardedProto(headers?: Headers): string {
  if (!headers) {
    return "";
  }
  return firstHeaderValue(headers.get("x-forwarded-proto")).toLowerCase();
}

function firstHeaderValue(value: string | null): string {
  if (!value) {
    return "";
  }
  return value.split(",")[0]?.trim() ?? "";
}
