import ProfileSettings from "@/app/components/profile-settings";
import RequestLoggingSettingsPanel from "@/app/components/request-logging-settings";
import TracePrivacySettingsPanel from "@/app/components/trace-privacy-settings";
import AdminShell from "@/app/components/admin-shell";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { fetchControlPlaneJSON } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

export default async function SettingsPage() {
  const auth = await requireAuth("/login?next=/settings");
  const tenantSlug = auth.tenantSlug;
  await requirePermissionOr403(tenantSlug, auth.user.id, ["settings:list"]);
  const tracePrivacyResult = await fetchControlPlaneJSON({
    tenantSlug,
    method: "GET",
    path: "/api/v1/settings/trace-privacy",
    userID: auth.user.id,
  });
  const requestLoggingResult = await fetchControlPlaneJSON({
    tenantSlug,
    method: "GET",
    path: "/api/v1/settings/request-logging",
    userID: auth.user.id,
  });

  const currentUser = {
    id: auth.user.id,
    email: auth.user.email,
    mobile: auth.user.mobile,
    role: auth.user.role,
  };

  return (
    <AdminShell tenantSlug={tenantSlug} currentUser={currentUser}>
      <div className="space-y-6">
        <ProfileSettings initialUser={currentUser} />
        <RequestLoggingSettingsPanel
          initialSettings={
            requestLoggingResult.ok
              ? (requestLoggingResult.payload as {
                  tenant_id?: string;
                  logging_mode: "none" | "metadata-only" | "redacted-content" | "full-content";
                  updated_at?: string;
                  updated_by_user_id?: string;
                })
              : null
          }
          initialError={requestLoggingResult.ok ? null : requestLoggingResult.errorMessage ?? "Failed to load request logging settings."}
        />
        <TracePrivacySettingsPanel
          initialSettings={
            tracePrivacyResult.ok
              ? (tracePrivacyResult.payload as {
                  tenant_id?: string;
                  capture_policy: "none" | "metadata-only" | "sampled-redacted" | "full-secure";
                  updated_at?: string;
                  updated_by_user_id?: string;
                })
              : null
          }
          initialError={tracePrivacyResult.ok ? null : tracePrivacyResult.errorMessage ?? "Failed to load tracing privacy settings."}
        />
      </div>
    </AdminShell>
  );
}
