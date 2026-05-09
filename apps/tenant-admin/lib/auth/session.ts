import { cookies, headers } from "next/headers";
import { redirect } from "next/navigation";

import { resolveTenantSlugFromHeaders } from "@/lib/tenant";
import { getAuthService } from "./factory";
import type { TenantSession, TenantUser } from "./types";

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

  const incomingHeaders = await headers();
  const hintedTenantSlug = resolveTenantSlugFromHeaders(incomingHeaders);

  const authService = getAuthService();
  return authService.getAuthContext(sessionID, hintedTenantSlug);
}

export async function requireAuth(redirectTo = "/login") {
  const auth = await getAuthContext();
  if (!auth) {
    redirect(redirectTo);
  }
  return auth;
}
