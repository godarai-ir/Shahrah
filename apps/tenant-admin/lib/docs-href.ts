/**
 * The documentation base URL for a tenant request, without a trailing slash. Docs are served from
 * the apex host, so a leading `<tenant>.` label is removed from the forwarded or direct host.
 */
export function buildDocsBaseHref(headersList: Headers, tenantSlug: string): string {
  return buildApexHref(headersList, tenantSlug, "/docs");
}

/**
 * A URL on the public apex site (docs, pricing) for a tenant request: the tenant's `<tenant>.` host
 * label is removed. Without a host it falls back to the bare path.
 */
export function buildApexHref(headersList: Headers, tenantSlug: string, path: string): string {
  const forwardedProto = headersList.get("x-forwarded-proto")?.split(",")[0]?.trim();
  const proto = forwardedProto || (headersList.get("host")?.includes("localhost") ? "http" : "https");
  const host = headersList.get("x-forwarded-host")?.split(",")[0]?.trim() || headersList.get("host") || "";
  const tenantPrefix = `${tenantSlug}.`;
  const apexHost = host.startsWith(tenantPrefix) ? host.slice(tenantPrefix.length) : host;

  return apexHost ? `${proto}://${apexHost}${path}` : path;
}
