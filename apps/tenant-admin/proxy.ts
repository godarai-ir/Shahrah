import { NextRequest, NextResponse } from "next/server";

import { resolveTenantContextFromRequestHost } from "@/lib/tenant";

export function proxy(request: NextRequest) {
  const pathname = request.nextUrl.pathname;
  if (isBypassedPath(pathname)) {
    return NextResponse.next();
  }

  const canonicalHost = request.headers.get("host") ?? request.nextUrl.host;
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

function respondTenantError(request: NextRequest, status: number, message: string) {
  if (request.nextUrl.pathname.startsWith("/api/")) {
    return NextResponse.json({ error: { message } }, { status });
  }
  return new NextResponse(message, { status });
}
