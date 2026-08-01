export type TenantResolutionSource =
  | "resolved-header"
  | "x-godarai-tenant"
  | "x-forwarded-host"
  | "host"
  | "fallback"
  | "none";

export type TenantResolutionContext = {
  tenantSlug: string;
  source: TenantResolutionSource;
  hostTenantSlug: string;
  ambiguous: boolean;
};

export function extractTenantSlugFromHost(host: string): string {
  const normalizedHost = normalizeHost(host);
  if (!normalizedHost) {
    return "";
  }

  const parts = normalizedHost.split(".");
  if (parts.length < 2) {
    return "";
  }

  const slug = parts[0]?.trim().toLowerCase();
  if (!slug || slug === "localhost" || slug === "www" || isReservedTenantLabel(slug)) {
    return "";
  }

  return slug;
}

export function resolveTenantSlugFromHeaders(
  headers: Headers,
  fallbackTenant = "",
): string {
  return resolveTenantContextFromHeaders(headers, fallbackTenant).tenantSlug;
}

export function resolveTenantContextFromRequestHost(
  host: string,
  fallbackTenant = "",
): TenantResolutionContext {
  const hostTenantSlug = resolveFromHostHeader(host);
  if (hostTenantSlug) {
    return {
      tenantSlug: hostTenantSlug,
      source: "host",
      hostTenantSlug,
      ambiguous: false,
    };
  }

  const effectiveFallback = resolveTenantFallback(fallbackTenant);
  if (effectiveFallback) {
    return {
      tenantSlug: effectiveFallback,
      source: "fallback",
      hostTenantSlug: "",
      ambiguous: false,
    };
  }

  return {
    tenantSlug: "",
    source: "none",
    hostTenantSlug: "",
    ambiguous: false,
  };
}

export function resolveTenantContextFromHeaders(
  headers: Headers,
  fallbackTenant = "",
): TenantResolutionContext {
  const effectiveFallback = resolveTenantFallback(fallbackTenant);

  const resolvedHeaderTenant = normalizeTenantSlug(headers.get("x-godarai-tenant-resolved"));
  if (resolvedHeaderTenant) {
    return {
      tenantSlug: resolvedHeaderTenant,
      source: "resolved-header",
      hostTenantSlug: resolvedHeaderTenant,
      ambiguous: false,
    };
  }

  const hostTenantSlug = resolveFromHostHeader(headers.get("x-forwarded-host")) ||
    resolveFromHostHeader(headers.get("host"));
  const hintedTenantSlug = normalizeTenantSlug(headers.get("x-godarai-tenant"));

  if (hostTenantSlug && hintedTenantSlug && hostTenantSlug !== hintedTenantSlug) {
    return {
      tenantSlug: "",
      source: "none",
      hostTenantSlug,
      ambiguous: true,
    };
  }

  if (hintedTenantSlug) {
    return {
      tenantSlug: hintedTenantSlug,
      source: "x-godarai-tenant",
      hostTenantSlug,
      ambiguous: false,
    };
  }

  const forwardedHostTenant = resolveFromHostHeader(headers.get("x-forwarded-host"));
  if (forwardedHostTenant) {
    return {
      tenantSlug: forwardedHostTenant,
      source: "x-forwarded-host",
      hostTenantSlug: forwardedHostTenant,
      ambiguous: false,
    };
  }

  const hostTenant = resolveFromHostHeader(headers.get("host"));
  if (hostTenant) {
    return {
      tenantSlug: hostTenant,
      source: "host",
      hostTenantSlug: hostTenant,
      ambiguous: false,
    };
  }

  if (effectiveFallback) {
    return {
      tenantSlug: effectiveFallback,
      source: "fallback",
      hostTenantSlug: "",
      ambiguous: false,
    };
  }

  return {
    tenantSlug: "",
    source: "none",
    hostTenantSlug: "",
    ambiguous: false,
  };
}

export function ensureTenantFromHeaders(
  headers: Headers,
  fallbackTenant = "",
): { ok: true; tenantSlug: string } | { ok: false; status: number; message: string } {
  const resolved = resolveTenantContextFromHeaders(headers, fallbackTenant);
  if (resolved.ambiguous) {
    return {
      ok: false,
      status: 400,
      message: "Ambiguous tenant context: host and tenant hint do not match.",
    };
  }
  if (!resolved.tenantSlug) {
    return {
      ok: false,
      status: 400,
      message: "Tenant host is required.",
    };
  }
  if (resolved.hostTenantSlug && resolved.hostTenantSlug !== resolved.tenantSlug) {
    return {
      ok: false,
      status: 403,
      message: "Tenant mismatch between host and request context.",
    };
  }
  return { ok: true, tenantSlug: resolved.tenantSlug };
}

export function assertTenantMatch(expected: string, actual: string): boolean {
  return normalizeTenantSlug(expected) !== "" && normalizeTenantSlug(expected) === normalizeTenantSlug(actual);
}

export function resolveRequestOriginFromHeaders(
  headers: Headers,
  fallbackOrigin: string,
): string {
  const host = normalizeHostWithPort(
    headers.get("x-forwarded-host") ?? headers.get("host") ?? "",
  );
  if (!host) {
    return fallbackOrigin;
  }

  const protocol =
    normalizeProtocol(headers.get("x-forwarded-proto") ?? "") ??
    normalizeProtocolFromOrigin(fallbackOrigin) ??
    "http";

  return `${protocol}://${host}`;
}

function resolveFromHostHeader(value: string | null): string {
  if (!value) {
    return "";
  }
  const fromHost = extractTenantSlugFromHost(value);
  if (fromHost) {
    return fromHost;
  }
  const rawHint = value.split(",")[0]?.trim().toLowerCase();
  if (rawHint && isSlug(rawHint)) {
    return rawHint;
  }
  return "";
}

function resolveTenantFallback(fallbackTenant: string): string {
  const envFallback =
    process.env.SHAHRAH_TENANT_SLUG ??
    process.env.TENANT_SLUG ??
    process.env.NEXT_PUBLIC_TENANT_SLUG ??
    "";
  return normalizeTenantSlug(fallbackTenant || envFallback);
}

function normalizeTenantSlug(value: string | null | undefined): string {
  const normalized = value?.trim().toLowerCase() ?? "";
  if (!normalized || !isSlug(normalized)) {
    return "";
  }
  return normalized;
}

function normalizeHost(value: string): string {
  const first = value.split(",")[0]?.trim().toLowerCase() ?? "";
  if (!first) {
    return "";
  }

  const withoutProtocol = first.replace(/^https?:\/\//, "");
  const withoutPath = withoutProtocol.split("/")[0] ?? "";

  return withoutPath.replace(/:\d+$/, "");
}

function normalizeHostWithPort(value: string): string {
  const first = value.split(",")[0]?.trim() ?? "";
  if (!first) {
    return "";
  }

  const withoutProtocol = first.replace(/^[a-z]+:\/\//i, "");
  return withoutProtocol.split("/")[0]?.trim() ?? "";
}

function normalizeProtocol(value: string): "http" | "https" | null {
  const first = value.split(",")[0]?.trim().toLowerCase() ?? "";
  if (first === "http" || first === "https") {
    return first;
  }
  return null;
}

function normalizeProtocolFromOrigin(origin: string): "http" | "https" | null {
  try {
    return normalizeProtocol(new URL(origin).protocol.replace(":", ""));
  } catch {
    return null;
  }
}

function isSlug(value: string): boolean {
  return /^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$/.test(value);
}

function isReservedTenantLabel(value: string): boolean {
  return value === "admin" || value === "auth" || value === "tenant-admin";
}
