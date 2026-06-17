import { NextResponse } from "next/server";

type DiscoveryAction = "identifier" | "tenant";

type DiscoveryBody = {
  action?: DiscoveryAction;
  identifier?: string;
  tenantName?: string;
};

type Tenant = {
  id: string;
  name?: string;
  slug: string;
  domain?: string;
  loginUrl?: string;
  login_url?: string;
};

const tenantSlugPattern = /^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$/;

export async function POST(request: Request) {
  const body = (await request.json().catch(() => null)) as DiscoveryBody | null;
  if (!body?.action) {
    return NextResponse.json({ error: "Invalid login discovery request." }, { status: 400 });
  }

  if (body.action === "identifier") {
    const identifier = normalizeIdentifier(body.identifier);
    if (!identifier) {
      return NextResponse.json(
        { error: "Please enter a valid email address or mobile number." },
        { status: 400 },
      );
    }
    const upstream = await callControlPlane(
      `/api/v1/login-discovery/tenants?identifier=${encodeURIComponent(identifier)}`,
    );
    if (!isSuccessfulStatus(upstream.status)) {
      return upstreamErrorResponse(upstream);
    }
    const payload = normalizeIdentifierPayload(upstream.payload, request);
    return NextResponse.json(payload, { status: upstream.status });
  }

  const tenantName = normalizeTenantName(body.tenantName);
  if (!tenantName || !tenantSlugPattern.test(tenantName)) {
    return NextResponse.json(
      { error: "Tenant name can only include lowercase letters, numbers, and hyphens." },
      { status: 400 },
    );
  }
  const upstream = await callControlPlane(
    `/api/v1/login-discovery/tenant/${encodeURIComponent(tenantName)}`,
  );
  if (!isSuccessfulStatus(upstream.status)) {
    return upstreamErrorResponse(upstream);
  }
  const payload = normalizeTenantPayload(upstream.payload, request);
  return NextResponse.json(payload, { status: upstream.status });
}

async function callControlPlane(path: string) {
  const baseURL = process.env.SHAHRAH_CONTROL_PLANE_URL ?? "http://127.0.0.1:4000";
  try {
    const response = await fetch(new URL(path, ensureTrailingSlash(baseURL)), {
      method: "GET",
      cache: "no-store",
    });
    const text = await response.text();
    return {
      status: response.status,
      payload: safeParsePayload(text),
    };
  } catch {
    return {
      status: 502,
      payload: { error: "We could not check workspaces right now. Please try again." },
    };
  }
}

function normalizeIdentifierPayload(payload: unknown, request: Request) {
  if (!payload || typeof payload !== "object") {
    return { status: "none", tenants: [] };
  }
  const candidate = payload as { status?: string; tenants?: Tenant[] };
  const tenants = Array.isArray(candidate.tenants)
    ? candidate.tenants.map((tenant) => withLoginURL(tenant, request)).filter(Boolean)
    : [];
  const status = tenants.length === 0 ? "none" : tenants.length === 1 ? "single" : "multiple";
  return { status, tenants };
}

function normalizeTenantPayload(payload: unknown, request: Request) {
  if (!payload || typeof payload !== "object") {
    return { exists: false };
  }
  const candidate = payload as { exists?: boolean; tenant?: Tenant };
  if (!candidate.exists || !candidate.tenant) {
    return { exists: false };
  }
  const tenant = withLoginURL(candidate.tenant, request);
  return tenant ? { exists: true, tenant } : { exists: false };
}

function withLoginURL(tenant: Tenant, request: Request): Tenant | null {
  const slug = normalizeTenantName(tenant.slug);
  if (!slug || !tenantSlugPattern.test(slug)) {
    return null;
  }
  return {
    ...tenant,
    name: tenant.name || slug,
    slug,
    loginUrl: buildTenantLoginURL(slug, request),
  };
}

function buildTenantLoginURL(slug: string, request: Request) {
  const requestURL = new URL(request.url);
  const protocol = process.env.NEXT_PUBLIC_TENANT_LOGIN_PROTOCOL || requestURL.protocol.replace(":", "");
  const baseHost =
    normalizeHost(process.env.TENANT_ADMIN_BASE_DOMAIN) ||
    normalizeHost(process.env.NEXT_PUBLIC_TENANT_ADMIN_BASE_DOMAIN) ||
    normalizeHost(process.env.NEXT_PUBLIC_TENANT_BASE_DOMAIN) ||
    normalizeHost(process.env.SHAHRAH_TENANT_BASE_DOMAIN) ||
    normalizeHost(request.headers.get("x-forwarded-host")) ||
    normalizeHost(request.headers.get("host")) ||
    requestURL.host;

  return `${protocol}://${slug}.${stripWWW(baseHost)}/login`;
}

function normalizeIdentifier(value: string | undefined) {
  const trimmed = (value ?? "").trim();
  if (!trimmed) {
    return "";
  }
  if (/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(trimmed.toLowerCase())) {
    return trimmed.toLowerCase();
  }
  const mobile = trimmed.replace(/[\s-]/g, "").replace(/^00/, "+");
  if (/^\+?[0-9]{10,15}$/.test(mobile)) {
    return mobile;
  }
  return "";
}

function normalizeTenantName(value: string | undefined) {
  return (value ?? "").trim().toLowerCase();
}

function normalizeHost(value: string | null | undefined) {
  const first = (value ?? "").split(",")[0]?.trim().toLowerCase() ?? "";
  if (!first) {
    return "";
  }
  const withoutProtocol = first.replace(/^https?:\/\//, "");
  return withoutProtocol.split("/")[0]?.replace(/\.$/, "") ?? "";
}

function stripWWW(host: string) {
  return host.startsWith("www.") ? host.slice(4) : host;
}

function ensureTrailingSlash(value: string) {
  return value.endsWith("/") ? value : `${value}/`;
}

function safeParsePayload(text: string) {
  if (!text.trim()) {
    return {};
  }
  try {
    return JSON.parse(text) as unknown;
  } catch {
    return { message: text };
  }
}

function isSuccessfulStatus(status: number) {
  return status >= 200 && status < 300;
}

function upstreamErrorResponse(upstream: { status: number; payload: unknown }) {
  const error = extractUpstreamError(upstream.payload);
  const status = upstream.status >= 400 && upstream.status < 500 ? 502 : upstream.status;
  return NextResponse.json({ error }, { status });
}

function extractUpstreamError(payload: unknown) {
  if (payload && typeof payload === "object") {
    const candidate = payload as { error?: unknown; message?: unknown };
    if (typeof candidate.error === "string" && candidate.error.trim()) {
      return candidate.error;
    }
    if (typeof candidate.message === "string" && candidate.message.trim()) {
      return candidate.message;
    }
  }
  return "We could not check workspaces right now. Please try again.";
}
