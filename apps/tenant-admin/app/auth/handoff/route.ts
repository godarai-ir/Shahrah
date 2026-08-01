import { NextRequest, NextResponse } from "next/server";

import { fetchControlPlaneJSON } from "@/lib/control-plane";
import { TENANT_ADMIN_SESSION_COOKIE } from "@/lib/auth/session";
import { resolveTenantSlugFromHeaders } from "@/lib/tenant";

type RedeemedHandoff = {
  session: { id: string; expires_at_iso: string };
  user: { id: string; tenant_slug: string };
};

export const dynamic = "force-dynamic";

export async function GET(request: NextRequest) {
  const tenantSlug = resolveTenantSlugFromHeaders(request.headers);
  const code = request.nextUrl.searchParams.get("code")?.trim() ?? "";
  const nextPath = sanitizeNextPath(request.nextUrl.searchParams.get("next"));
  if (!tenantSlug || !/^[A-Za-z0-9_-]{32,256}$/.test(code)) {
    return NextResponse.redirect(new URL("/login?social_error=invalid", request.nextUrl.origin), 303);
  }

  const result = await fetchControlPlaneJSON({
    tenantSlug,
    method: "POST",
    path: "/api/v1/auth/social/handoff/redeem",
    body: { tenant_slug: tenantSlug, code },
    requestHeaders: request.headers,
  });
  if (!result.ok) {
    return NextResponse.redirect(new URL("/login?social_error=expired", request.nextUrl.origin), 303);
  }
  const payload = result.payload as RedeemedHandoff;
  const response = NextResponse.redirect(new URL(nextPath, request.nextUrl.origin), 303);
  response.cookies.set({
    name: TENANT_ADMIN_SESSION_COOKIE,
    value: payload.session.id,
    httpOnly: true,
    secure: process.env.NODE_ENV === "production",
    sameSite: "lax",
    path: "/",
    expires: new Date(payload.session.expires_at_iso),
  });
  response.headers.set("Cache-Control", "no-store");
  return response;
}

function sanitizeNextPath(value: string | null) {
  const normalized = value?.trim() ?? "";
  return normalized.startsWith("/") && !normalized.startsWith("//") ? normalized.slice(0, 1024) : "/";
}
