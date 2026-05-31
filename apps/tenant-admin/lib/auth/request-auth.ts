import { NextRequest } from "next/server";

import { TENANT_ADMIN_SESSION_COOKIE } from "@/lib/auth/session";
import { ensureTenantFromHeaders } from "@/lib/tenant";
import { getAuthService } from "./factory";

export async function readAuthFromRequest(request: NextRequest) {
  const tenantResolution = ensureTenantFromHeaders(request.headers);
  if (!tenantResolution.ok) {
    return {
      ok: false as const,
      status: tenantResolution.status,
      message: tenantResolution.message,
    };
  }

  const sessionID = request.cookies.get(TENANT_ADMIN_SESSION_COOKIE)?.value;
  if (!sessionID) {
    return {
      ok: false as const,
      status: 401,
      message: "برای ادامه باید وارد حساب کاربری شوید.",
    };
  }

  const hintedTenantSlug = tenantResolution.tenantSlug;
  const authService = getAuthService();
  const ctx = await authService.getAuthContext(sessionID, hintedTenantSlug);

  if (!ctx) {
    return {
      ok: false as const,
      status: 403,
      message: "نشست کاربری با tenant جاری همخوانی ندارد یا منقضی شده است.",
    };
  }

  return {
    ok: true as const,
    tenantSlug: ctx.tenantSlug,
    session: ctx.session,
    user: ctx.user,
  };
}
