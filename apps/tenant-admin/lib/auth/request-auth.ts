import { NextRequest } from "next/server";

import { TENANT_ADMIN_SESSION_COOKIE } from "@/lib/auth/session";
import { getTenantSession, getTenantUser } from "@/lib/auth/store";
import { resolveTenantSlugFromHeaders } from "@/lib/tenant";

export async function readAuthFromRequest(request: NextRequest) {
  const tenantSlug = resolveTenantSlugFromHeaders(
    request.headers,
    process.env.SHAHRAH_DEFAULT_TENANT ?? "default",
  );
  const sessionID = request.cookies.get(TENANT_ADMIN_SESSION_COOKIE)?.value;
  if (!sessionID) {
    return {
      ok: false as const,
      status: 401,
      message: "برای ادامه باید وارد حساب کاربری شوید.",
    };
  }

  const session = await getTenantSession(sessionID);
  if (!session || session.tenantSlug !== tenantSlug) {
    return {
      ok: false as const,
      status: 401,
      message: "نشست کاربری معتبر نیست یا منقضی شده است.",
    };
  }

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
