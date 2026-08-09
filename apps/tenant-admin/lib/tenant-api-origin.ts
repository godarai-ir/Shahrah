export function getTenantApiBaseURL(): string {
  if (typeof window === "undefined") return "http://tenant1.localhost:4000";

  return getTenantApiBaseURLForOrigin(window.location.origin);
}

export function getTenantApiBaseURLForOrigin(origin: string): string {
  const url = new URL(origin);
  if (isLocalHostname(url.hostname)) {
    url.port = "4000";
    return url.origin;
  }
  if (url.hostname.toLowerCase().startsWith("www.")) {
    url.hostname = url.hostname.slice(4);
  }
  url.protocol = "https:";
  url.port = "";
  return url.origin;
}

function isLocalHostname(hostname: string): boolean {
  const normalized = hostname.toLowerCase();
  return normalized === "localhost"
    || normalized.endsWith(".localhost")
    || normalized === "127.0.0.1"
    || normalized === "[::1]";
}
