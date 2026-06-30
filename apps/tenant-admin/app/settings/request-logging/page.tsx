import AdminShell from "@/app/components/admin-shell";
import RequestLoggingSettingsPanel from "@/app/components/request-logging-settings";
import {
  loadRequestLoggingSettings,
  loadSettingsShell,
  type RequestLoggingSettingsData,
} from "@/app/settings/_shared";

export const dynamic = "force-dynamic";

export default async function RequestLoggingSettingsPage() {
  const { tenantSlug, userID, currentUser } = await loadSettingsShell("/settings/request-logging");
  const requestLoggingResult = await loadRequestLoggingSettings(tenantSlug, userID);

  return (
    <AdminShell tenantSlug={tenantSlug} currentUser={currentUser}>
      <RequestLoggingSettingsPanel
        initialSettings={requestLoggingResult.ok ? (requestLoggingResult.payload as RequestLoggingSettingsData) : null}
        initialError={requestLoggingResult.ok ? null : requestLoggingResult.errorMessage ?? "Failed to load request logging settings."}
      />
    </AdminShell>
  );
}
