import "server-only";

import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { fetchControlPlaneJSON } from "@/lib/control-plane";

export type SettingsCurrentUser = {
  id: string;
  email?: string;
  mobile?: string;
  socialProviders?: string[];
  role: string;
};

export async function loadSettingsShell(nextPath: string) {
  const auth = await requireAuth(`/login?next=${nextPath}`);
  const tenantSlug = auth.tenantSlug;
  await requirePermissionOr403(tenantSlug, auth.user.id, ["settings:list"]);

  const currentUser: SettingsCurrentUser = {
    id: auth.user.id,
    email: auth.user.email,
    mobile: auth.user.mobile,
    socialProviders: auth.user.socialProviders,
    role: auth.user.role,
  };
  return { tenantSlug, userID: auth.user.id, currentUser };
}

export function loadTracePrivacySettings(tenantSlug: string, userID: string) {
  return fetchControlPlaneJSON({
    tenantSlug,
    method: "GET",
    path: "/api/v1/settings/trace-privacy",
    userID,
  });
}
