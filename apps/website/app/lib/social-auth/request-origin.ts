export function resolveRequestOrigin(
  request: Request,
  requestURL = new URL(request.url),
) {
  const forwardedHost = firstHeaderValue(request.headers.get("x-forwarded-host"));
  const host = forwardedHost || firstHeaderValue(request.headers.get("host"));
  const forwardedProto = firstHeaderValue(
    request.headers.get("x-forwarded-proto"),
  ).toLowerCase();
  const protocol =
    forwardedProto === "http" || forwardedProto === "https"
      ? `${forwardedProto}:`
      : requestURL.protocol;

  return host ? `${protocol}//${host.toLowerCase()}` : requestURL.origin;
}

export function hasExpectedRequestOrigin(request: Request, expectedOrigin: string) {
  const expected = new URL(expectedOrigin).origin;
  if (resolveRequestOrigin(request) !== expected) {
    return false;
  }

  const originHeader = request.headers.get("origin");
  if (!originHeader) {
    return true;
  }

  try {
    return new URL(originHeader).origin === expected;
  } catch {
    return false;
  }
}

function firstHeaderValue(value: string | null) {
  return value?.split(",")[0]?.trim() ?? "";
}
