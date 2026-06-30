import AdminShell from "@/app/components/admin-shell";
import ProfileSettings from "@/app/components/profile-settings";
import { loadSettingsShell } from "@/app/settings/_shared";

export const dynamic = "force-dynamic";

export default async function AccountSettingsPage() {
  const { tenantSlug, currentUser } = await loadSettingsShell("/settings/account");

  return (
    <AdminShell tenantSlug={tenantSlug} currentUser={currentUser}>
      <ProfileSettings initialUser={currentUser} />
    </AdminShell>
  );
}
