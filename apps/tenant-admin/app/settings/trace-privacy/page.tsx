import AdminShell from "@/app/components/admin-shell";
import TracePrivacySettingsPanel from "@/app/components/trace-privacy-settings";
import {
  loadSettingsShell,
  loadTracePrivacySettings,
  type TracePrivacySettingsData,
} from "@/app/settings/_shared";

export const dynamic = "force-dynamic";

export default async function TracePrivacySettingsPage() {
  const { tenantSlug, userID, currentUser } = await loadSettingsShell("/settings/trace-privacy");
  const tracePrivacyResult = await loadTracePrivacySettings(tenantSlug, userID);

  return (
    <AdminShell tenantSlug={tenantSlug} currentUser={currentUser}>
      <TracePrivacySettingsPanel
        initialSettings={tracePrivacyResult.ok ? (tracePrivacyResult.payload as TracePrivacySettingsData) : null}
        initialError={tracePrivacyResult.ok ? null : tracePrivacyResult.errorMessage ?? "Failed to load tracing privacy settings."}
      />
    </AdminShell>
  );
}
