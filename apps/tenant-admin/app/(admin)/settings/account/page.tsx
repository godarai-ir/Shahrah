import AdminShell from "@/features/admin-shell/AdminShell";
import { getTranslations } from "next-intl/server";
import ProfileSettings from "@/features/profile/ProfileSettings";
import { loadSettingsShell } from "@/features/settings/server";

export const dynamic = "force-dynamic";

export default async function AccountSettingsPage() {
  const { tenantSlug, currentUser } = await loadSettingsShell("/settings/account");
  const t = await getTranslations("settings");

  return (
    <AdminShell messageNamespaces={["settings"]}
      tenantSlug={tenantSlug}
      currentUser={currentUser}
      pageHeader={{ title: t("title"), description: t("subtitle") }}
    >
      <ProfileSettings initialUser={currentUser} />
    </AdminShell>
  );
}
