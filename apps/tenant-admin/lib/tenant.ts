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
  const envFallback =
    process.env.SHAHRAH_TENANT_SLUG ??
    process.env.TENANT_SLUG ??
    process.env.NEXT_PUBLIC_TENANT_SLUG ??
    "";
  const effectiveFallback = fallbackTenant || envFallback;
  const candidates = [
    headers.get("x-shahrah-tenant"),
    headers.get("x-forwarded-host"),
    headers.get("host"),
  ];

  for (const candidate of candidates) {
    if (!candidate) {
      continue;
    }
    const slug = extractTenantSlugFromHost(candidate);
    if (slug) {
      return slug;
    }

    const rawHint = candidate.split(",")[0]?.trim().toLowerCase();
    if (rawHint && isSlug(rawHint)) {
      return rawHint;
    }
  }

  return effectiveFallback;
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
  return value === "tenant-admin";
}
