import AdminShell from "@/app/components/admin-shell";
import { getTranslations } from "next-intl/server";
import ProfileSettings from "@/app/components/profile-settings";
import { loadSettingsShell } from "@/app/settings/_shared";

export const dynamic = "force-dynamic";

export default async function AccountSettingsPage() {
  const { tenantSlug, currentUser } = await loadSettingsShell("/settings/account");
  const t = await getTranslations("settings");

  return (
    <AdminShell
      tenantSlug={tenantSlug}
      currentUser={currentUser}
      pageHeader={{ title: t("title"), description: t("subtitle") }}
    >
      <ProfileSettings initialUser={currentUser} />
    </AdminShell>
  );
}
