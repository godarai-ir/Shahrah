import { NextRequest } from "next/server";

import { TENANT_ADMIN_SESSION_COOKIE } from "@/lib/auth/session";
import { getTenantSession, getTenantUser } from "@/lib/auth/store";
import { resolveTenantSlugFromHeaders } from "@/lib/tenant";

export async function readAuthFromRequest(request: NextRequest) {
  const sessionID = request.cookies.get(TENANT_ADMIN_SESSION_COOKIE)?.value;
  if (!sessionID) {
    return {
      ok: false as const,
      status: 401,
      message: "برای ادامه باید وارد حساب کاربری شوید.",
    };
  }

  const session = await getTenantSession(sessionID);
  if (!session) {
    return {
      ok: false as const,
      status: 401,
      message: "نشست کاربری معتبر نیست یا منقضی شده است.",
    };
  }

  // In real deployments host/header tenant hints can differ from the authenticated
  // tenant context. Session tenant is the source of truth for protected routes.
  const hintedTenantSlug = resolveTenantSlugFromHeaders(
    request.headers,
    process.env.SHAHRAH_DEFAULT_TENANT ?? "default",
  );
  const tenantSlug =
    hintedTenantSlug === session.tenantSlug
      ? hintedTenantSlug
      : session.tenantSlug;

  const user = await getTenantUser(tenantSlug, session.userID);
  if (!user) {
    return {
      ok: false as const,
      status: 401,
      message: "کاربر مربوط به این نشست یافت نشد.",
    };
  }

  return {
    ok: true as const,
    tenantSlug,
    session,
    user,
  };
}
