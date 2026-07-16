import { NextRequest, NextResponse } from "next/server";

import { resolveTenantContextFromRequestHost } from "@/lib/tenant";

export function proxy(request: NextRequest) {
  const pathname = request.nextUrl.pathname;
  if (isBypassedPath(pathname)) {
    return NextResponse.next();
  }

  const canonicalHost = request.headers.get("host") ?? request.nextUrl.host;
  const requestSecurityError = validateStateChangingAPIRequest(
    request,
    canonicalHost,
  );
  if (requestSecurityError) {
    return requestSecurityError;
  }

  const resolved = resolveTenantContextFromRequestHost(canonicalHost);
  if (resolved.ambiguous) {
    return respondTenantError(request, 400, "Ambiguous tenant context.");
  }
  if (!resolved.tenantSlug) {
    return respondTenantError(request, 400, "Tenant host is required.");
  }

  const nextHeaders = new Headers(request.headers);
  nextHeaders.delete("x-godarai-tenant");
  nextHeaders.set("x-forwarded-host", canonicalHost);
  nextHeaders.set("x-godarai-tenant-resolved", resolved.tenantSlug);
  nextHeaders.set("x-shahrah-tenant-source", resolved.source);
  if (resolved.hostTenantSlug) {
    nextHeaders.set("x-shahrah-tenant-host", resolved.hostTenantSlug);
  } else {
    nextHeaders.delete("x-shahrah-tenant-host");
  }

  return NextResponse.next({
    request: {
      headers: nextHeaders,
    },
  });
}

export const config = {
  matcher: ["/((?!_next/static|_next/image|favicon.ico|robots.txt|sitemap.xml).*)"],
};

function isBypassedPath(pathname: string): boolean {
  return pathname.startsWith("/api/v1/auth/captcha/");
}

function validateStateChangingAPIRequest(
  request: NextRequest,
  canonicalHost: string,
): NextResponse | null {
  if (
    !request.nextUrl.pathname.startsWith("/api/") ||
    !isStateChangingMethod(request.method)
  ) {
    return null;
  }

  const fetchSite = request.headers.get("sec-fetch-site")?.trim().toLowerCase();
  if (fetchSite && fetchSite !== "same-origin") {
    return csrfError();
  }

  const origin = request.headers.get("origin")?.trim();
  if (origin) {
    if (!isExpectedOrigin(origin, request.nextUrl.protocol, canonicalHost)) {
      return csrfError();
    }
  } else if (fetchSite !== "same-origin") {
    return csrfError();
  }

  if (!hasSupportedMutationContentType(request)) {
    return NextResponse.json(
      { error: { message: "Unsupported content type." } },
      { status: 415 },
    );
  }

  return null;
}

function isStateChangingMethod(method: string): boolean {
  return method === "POST" || method === "PUT" || method === "PATCH" || method === "DELETE";
}

function isExpectedOrigin(
  origin: string,
  protocol: string,
  canonicalHost: string,
): boolean {
  try {
    return new URL(origin).origin === new URL(`${protocol}//${canonicalHost}`).origin;
  } catch {
    return false;
  }
}

function hasSupportedMutationContentType(request: NextRequest): boolean {
  const rawContentType = request.headers.get("content-type");
  if (!rawContentType) {
    return request.body === null;
  }

  const mediaType = rawContentType.split(";", 1)[0]?.trim().toLowerCase() ?? "";
  if (mediaType === "application/json" || mediaType.endsWith("+json")) {
    return true;
  }

  return (
    request.nextUrl.pathname === "/api/wallet/bank-transfers" &&
    mediaType === "multipart/form-data"
  );
}

function csrfError() {
  return NextResponse.json(
    { error: { message: "Cross-origin request rejected." } },
    { status: 403 },
  );
}

function respondTenantError(request: NextRequest, status: number, message: string) {
  if (request.nextUrl.pathname.startsWith("/api/")) {
    return NextResponse.json({ error: { message } }, { status });
  }
  return new NextResponse(message, { status });
}
