export type ClientIPAddressOptions = {
  ipAddressHeaders: string[];
  trustedProxies?: string[];
};

type ClientIPAddressEnvironment = {
  [key: string]: string | undefined;
  BETTER_AUTH_IP_ADDRESS_HEADERS?: string;
  BETTER_AUTH_TRUSTED_PROXIES?: string;
};

export function resolveClientIPAddressOptions(
  environment: ClientIPAddressEnvironment = process.env,
): ClientIPAddressOptions {
  const configuredHeaders = [...new Set(
    parseCommaSeparatedList(environment.BETTER_AUTH_IP_ADDRESS_HEADERS)
      .map((header) => header.toLowerCase())
      .filter((header) => /^[a-z0-9-]+$/.test(header)),
  )];
  const trustedProxies = parseCommaSeparatedList(
    environment.BETTER_AUTH_TRUSTED_PROXIES,
  );

  // ingress-nginx overwrites X-Real-IP with a single resolved address. Better
  // Auth intentionally rejects an untrusted multi-hop X-Forwarded-For chain;
  // without this setting every production request can share one fallback
  // rate-limit bucket and receive unrelated 429 responses.
  const ipAddressHeaders = configuredHeaders.length > 0
    ? configuredHeaders
    : ["x-real-ip"];

  return trustedProxies.length > 0
    ? { ipAddressHeaders, trustedProxies }
    : { ipAddressHeaders };
}

function parseCommaSeparatedList(value: string | undefined): string[] {
  return [...new Set(
    (value ?? "")
      .split(",")
      .map((entry) => entry.trim())
      .filter(Boolean),
  )];
}
