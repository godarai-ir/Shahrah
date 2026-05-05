import { NextResponse } from "next/server";

import { resolveTenantSlugFromHeaders } from "@/lib/tenant";

type ProxyOptions = {
  request: Request;
  method: "GET" | "POST" | "PUT" | "PATCH" | "DELETE";
  path: string;
  body?: unknown;
  userID?: string;
};

type ControlPlaneRequest = {
  tenantSlug: string;
  method: "GET" | "POST" | "PUT" | "PATCH" | "DELETE";
  path: string;
  body?: unknown;
  userID?: string;
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

export async function proxyToControlPlane(options: ProxyOptions) {
  const tenantSlug = resolveTenantSlugFromHeaders(
    options.request.headers,
  );
  if (!tenantSlug) {
    return NextResponse.json(
      { error: { message: "Tenant host is required." } },
      { status: 400 },
    );
  }
  const upstream = await fetchControlPlaneJSON({
    tenantSlug,
    method: options.method,
    path: options.path,
    body: options.body,
    userID: options.userID,
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

  try {
    const upstream = await fetch(url, {
      method: options.method,
      headers: {
        Authorization: `Bearer ${masterKey}`,
        Accept: "application/json",
        "Content-Type": "application/json",
        "x-shahrah-tenant": options.tenantSlug,
        ...(options.userID ? { "x-shahrah-user-id": options.userID } : {}),
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
