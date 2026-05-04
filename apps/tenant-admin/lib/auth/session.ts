import { cookies, headers } from "next/headers";
import { redirect } from "next/navigation";

import { resolveTenantSlugFromHeaders } from "@/lib/tenant";
import {
  getTenantSession,
  getTenantUser,
  revokeTenantSession,
  TenantSession,
  TenantUser,
} from "@/lib/auth/store";

export const TENANT_ADMIN_SESSION_COOKIE = "tenant_admin_session";

export type AuthContext = {
  tenantSlug: string;
  session: TenantSession;
  user: TenantUser;
};

export async function getAuthContext(): Promise<AuthContext | null> {
  const cookieStore = await cookies();
  const sessionID = cookieStore.get(TENANT_ADMIN_SESSION_COOKIE)?.value;
  if (!sessionID) {
    return null;
  }

  const session = await getTenantSession(sessionID);
  if (!session) {
    return null;
  }

  const incomingHeaders = await headers();
  const hintedTenantSlug = resolveTenantSlugFromHeaders(
    incomingHeaders,
  );
  const tenantSlug =
    hintedTenantSlug === session.tenantSlug
      ? hintedTenantSlug
      : session.tenantSlug;

  const user = await getTenantUser(tenantSlug, session.userID);
  if (!user) {
    await revokeTenantSession(session.id);
    return null;
  }

  return {
    tenantSlug,
    session,
    user,
  };
}

export async function requireAuth(redirectTo = "/login") {
  const auth = await getAuthContext();
  if (!auth) {
    redirect(redirectTo);
  }
  return auth;
}
