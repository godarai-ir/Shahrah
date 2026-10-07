import { NextRequest, NextResponse } from "next/server";

import { TENANT_ADMIN_SESSION_COOKIE } from "@/lib/auth/session";
import {
  OIDC_REQUEST_PARAM,
  acceptOIDCLogin,
  normalizeOIDCRequestID,
} from "@/lib/auth/oidc-bridge";
import { resolveRequestOriginFromHeaders, resolveTenantSlugFromHeaders } from "@/lib/tenant";

export const dynamic = "force-dynamic";

function loginRedirect(origin: string, params: Record<string, string>) {
  const response = NextResponse.redirect(new URL(`/login?${new URLSearchParams(params)}`, origin), 303);
  response.headers.set("Cache-Control", "no-store");
  return response;
}

/**
 * Completes a pending OIDC sign-in for the current tenant-admin session and sends the
 * browser back to the relying party. The control plane validates the session, tenant, and
 * redirect URI; this handler only relays its decision.
 */
export async function GET(request: NextRequest) {
  const origin = resolveRequestOriginFromHeaders(request.headers, request.url);
  const tenantSlug = resolveTenantSlugFromHeaders(request.headers);
  const requestID = normalizeOIDCRequestID(request.nextUrl.searchParams.get(OIDC_REQUEST_PARAM));
  if (!tenantSlug || !requestID) {
    return loginRedirect(origin, { oidc_error: "expired" });
  }

  const sessionID = request.cookies.get(TENANT_ADMIN_SESSION_COOKIE)?.value;
  if (!sessionID) {
    return loginRedirect(origin, { [OIDC_REQUEST_PARAM]: requestID });
  }

  const decision = await acceptOIDCLogin(tenantSlug, requestID, sessionID);
  if (!decision.ok) {
    if (decision.code === "session_invalid") {
      return loginRedirect(origin, { [OIDC_REQUEST_PARAM]: requestID, oidc_reauth: "1" });
    }
    return loginRedirect(origin, { oidc_error: decision.code });
  }

  const response = NextResponse.redirect(decision.value, 303);
  response.headers.set("Cache-Control", "no-store");
  return response;
}
