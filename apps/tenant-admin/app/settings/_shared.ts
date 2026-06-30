import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { fetchControlPlaneJSON } from "@/lib/control-plane";

export type SettingsCurrentUser = {
  id: string;
  email?: string;
  mobile?: string;
  role: string;
};

export type RequestLoggingSettingsData = {
  tenant_id?: string;
  logging_mode: "none" | "metadata-only" | "redacted-content" | "full-content";
  updated_at?: string;
  updated_by_user_id?: string;
};

export type TracePrivacySettingsData = {
  tenant_id?: string;
  capture_policy: "none" | "metadata-only" | "sampled-redacted" | "full-secure";
  updated_at?: string;
  updated_by_user_id?: string;
};

export async function loadSettingsShell(nextPath: string) {
  const auth = await requireAuth(`/login?next=${nextPath}`);
  const tenantSlug = auth.tenantSlug;

  await requirePermissionOr403(tenantSlug, auth.user.id, ["settings:list"]);

  const currentUser: SettingsCurrentUser = {
    id: auth.user.id,
    email: auth.user.email,
    mobile: auth.user.mobile,
    role: auth.user.role,
  };

  return {
    tenantSlug,
    userID: auth.user.id,
    currentUser,
  };
}

export async function loadRequestLoggingSettings(tenantSlug: string, userID: string) {
  return fetchControlPlaneJSON({
    tenantSlug,
    method: "GET",
    path: "/api/v1/settings/request-logging",
    userID,
  });
}

export async function loadTracePrivacySettings(tenantSlug: string, userID: string) {
  return fetchControlPlaneJSON({
    tenantSlug,
    method: "GET",
    path: "/api/v1/settings/trace-privacy",
    userID,
  });
}
