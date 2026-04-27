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
  if (!slug || slug === "localhost" || slug === "www") {
    return "";
  }

  return slug;
}

export function resolveTenantSlugFromHeaders(
  headers: Headers,
  fallbackTenant = "default",
): string {
  const candidates = [
    headers.get("x-forwarded-host"),
    headers.get("host"),
    headers.get("x-shahrah-tenant"),
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

  return fallbackTenant;
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

function isSlug(value: string): boolean {
  return /^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$/.test(value);
}
