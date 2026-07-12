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
};

type ControlPlaneRequest = {
  tenantSlug: string;
  method: "GET" | "POST" | "PUT" | "PATCH" | "DELETE";
  path: string;
  body?: unknown;
  userID?: string;
  requestHeaders?: Headers;
  upstreamHeaders?: HeadersInit;
};

type APIErrorPayload = {
  error?: {
    message?: string;
  };
  message?: string;
};

export type ControlPlaneResponse = {
  ok: boolean;
  status: number;
  statusCode?: number;
  payload: unknown;
  errorMessage?: string;
};

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

  const tenantSlug = tenantResolution.tenantSlug;
  let resolvedUserID = options.userID?.trim() ?? "";
  const auth = await readAuthFromRequest(options.request);
  if (auth.ok) {
    if (!assertTenantMatch(tenantSlug, auth.tenantSlug)) {
      return NextResponse.json(
        { error: { message: "Tenant mismatch between host and authenticated session." } },
        { status: 403 },
      );
    }
    if (!resolvedUserID) {
      resolvedUserID = auth.user.id;
    }
  } else if (!resolvedUserID && auth.status >= 500) {
    return NextResponse.json(
      { error: { message: auth.message } },
      { status: auth.status },
    );
  }

  const upstream = await fetchControlPlaneJSON({
    tenantSlug,
    method: options.method,
    path: options.path,
    body: options.body,
    userID: resolvedUserID || undefined,
    requestHeaders: options.request.headers,
    upstreamHeaders: options.upstreamHeaders,
  });

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
      { status: upstream.status },
    );
  }

  if (upstream.status === 204 || upstream.status === 205 || upstream.status === 304) {
    return new NextResponse(null, { status: upstream.status });
  }

  return NextResponse.json(upstream.payload, { status: upstream.status });
}

export async function proxyToControlPlaneBinary(options: ProxyOptions) {
  const tenantResolution = ensureTenantFromHeaders(options.request.headers, options.tenantSlug);
  if (!tenantResolution.ok) {
    return NextResponse.json(
      { error: { message: tenantResolution.message } },
      { status: tenantResolution.status },
    );
  }

  const tenantSlug = tenantResolution.tenantSlug;
  let resolvedUserID = options.userID?.trim() ?? "";
  const auth = await readAuthFromRequest(options.request);
  if (auth.ok) {
    if (!assertTenantMatch(tenantSlug, auth.tenantSlug)) {
      return NextResponse.json(
        { error: { message: "Tenant mismatch between host and authenticated session." } },
        { status: 403 },
      );
    }
    if (!resolvedUserID) {
      resolvedUserID = auth.user.id;
    }
  } else if (!resolvedUserID && auth.status >= 500) {
    return NextResponse.json(
      { error: { message: auth.message } },
      { status: auth.status },
    );
  }

  const upstream = await fetchControlPlaneRaw({
    tenantSlug,
    method: options.method,
    path: options.path,
    body: options.body,
    userID: resolvedUserID || undefined,
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
    };
  }

  const url = new URL(normalizePath(options.path), baseURL).toString();
  const forwardedHost = resolveForwardedHost(options.requestHeaders);
  const forwardedProto = resolveForwardedProto(options.requestHeaders);
  const upstreamHeaders = new Headers(options.upstreamHeaders ?? {});

  try {
    const upstream = await fetch(url, {
      method: options.method,
      headers: {
        Authorization: `Bearer ${masterKey}`,
        Accept: "application/json",
        "Content-Type": "application/json",
        "x-godarai-tenant": options.tenantSlug,
        ...(options.userID ? { "x-shahrah-user-id": options.userID } : {}),
        ...(forwardedHost ? { "x-forwarded-host": forwardedHost } : {}),
        ...(forwardedProto ? { "x-forwarded-proto": forwardedProto } : {}),
        ...Object.fromEntries(upstreamHeaders.entries()),
      },
      body: options.body ? JSON.stringify(options.body) : undefined,
      cache: "no-store",
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

  try {
    const upstream = await fetch(url, {
      method: options.method,
      headers: {
        Authorization: `Bearer ${masterKey}`,
        Accept: "*/*",
        ...(options.body ? { "Content-Type": "application/json" } : {}),
        "x-godarai-tenant": options.tenantSlug,
        ...(options.userID ? { "x-shahrah-user-id": options.userID } : {}),
        ...(forwardedHost ? { "x-forwarded-host": forwardedHost } : {}),
        ...(forwardedProto ? { "x-forwarded-proto": forwardedProto } : {}),
        ...Object.fromEntries(upstreamHeaders.entries()),
      },
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
